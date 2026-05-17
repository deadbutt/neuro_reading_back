package comment

import (
	"neuro-reading/db"
	"neuro-reading/model"
	"neuro-reading/notification"
	"neuro-reading/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

type CreateCommentRequest struct {
	Content  string  `json:"content" binding:"required,min=1,max=1000"`
	ParentID *string `json:"parentId,omitempty"`
}

func (h *Handler) GetComments(c *gin.Context) {
	bookID := c.Param("bookId")
	if bookID == "" {
		bookID = c.Param("articleId")
	}
	sort := c.DefaultQuery("sort", "hot")

	var req model.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req.Page = 1
		req.PageSize = 20
	}

	var comments []model.Comment
	var total int64

	query := db.DB.Model(&model.Comment{}).Where("book_id = ? AND parent_id IS NULL", bookID)
	query.Count(&total)

	switch sort {
	case "new":
		query = query.Order("create_time DESC")
	case "author":
		query = query.Order("reply_count DESC")
	default:
		query = query.Order("like_count DESC")
	}

	query.Offset(req.GetOffset()).Limit(req.GetLimit()).Find(&comments)

	userID := c.GetString("userId")
	list := make([]model.CommentResponse, 0)
	for _, comment := range comments {
		var user model.User
		db.DB.Where("user_id = ?", comment.UserID).First(&user)

		var isLiked bool
		if userID != "" {
			var like model.Like
			if err := db.DB.Where("user_id = ? AND target_id = ? AND type = ?", userID, comment.CommentID, "comment").First(&like).Error; err == nil {
				isLiked = true
			}
		}

		var replies []model.Comment
		db.DB.Where("parent_id = ?", comment.CommentID).Order("create_time ASC").Limit(3).Find(&replies)

		var replyResponses []model.CommentReplyResponse
		for _, reply := range replies {
			var replyUser model.User
			db.DB.Where("user_id = ?", reply.UserID).First(&replyUser)

			replyResponses = append(replyResponses, model.CommentReplyResponse{
				ReplyID:    reply.CommentID,
				UserID:     reply.UserID,
				UserName:   replyUser.Nickname,
				UserAvatar: replyUser.Avatar,
				Content:    reply.Content,
				CreateTime: reply.CreateTime,
				ToUserName: user.Nickname,
			})
		}

		list = append(list, model.CommentResponse{
			CommentID:  comment.CommentID,
			BookID:     comment.BookID,
			UserID:     comment.UserID,
			UserName:   user.Nickname,
			UserAvatar: user.Avatar,
			Content:    comment.Content,
			CreateTime: comment.CreateTime,
			LikeCount:  comment.LikeCount,
			IsLiked:    isLiked,
			ReplyCount: comment.ReplyCount,
			Replies:    replyResponses,
		})
	}

	c.JSON(200, model.PageSuccess(list, total, req.Page, req.PageSize))
}

func (h *Handler) CreateComment(c *gin.Context) {
	userID := c.GetString("userId")
	bookID := c.Param("bookId")
	if bookID == "" {
		bookID = c.Param("articleId")
	}

	var req CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, model.Error(1001, "参数错误"))
		return
	}

	var currentUser model.User
	db.DB.Where("user_id = ?", userID).First(&currentUser)

	comment := model.Comment{
		CommentID:  utils.GenerateCommentID(),
		BookID:     bookID,
		UserID:     userID,
		Content:    req.Content,
		ParentID:   req.ParentID,
		CreateTime: utils.CurrentTime(),
	}

	if err := db.DB.Create(&comment).Error; err != nil {
		c.JSON(500, model.Error(1005, "服务器内部错误"))
		return
	}

	if req.ParentID != nil {
		db.DB.Model(&model.Comment{}).Where("comment_id = ?", *req.ParentID).UpdateColumn("reply_count", gorm.Expr("reply_count + 1"))

		var parentComment model.Comment
		if err := db.DB.Where("comment_id = ?", *req.ParentID).First(&parentComment).Error; err == nil {
			if parentComment.UserID != userID {
				notification.CreateNotification(
					parentComment.UserID,
					"comment_reply",
					"评论回复",
					currentUser.Nickname+" 回复了你的评论："+req.Content,
					bookID,
					userID,
					currentUser.Nickname,
					currentUser.Avatar,
				)
			}
		}
	}

	db.DB.Model(&model.Book{}).Where("book_id = ?", bookID).UpdateColumn("comment_count", gorm.Expr("comment_count + 1"))

	c.JSON(200, model.Success(model.CommentResponse{
		CommentID:  comment.CommentID,
		BookID:     comment.BookID,
		UserID:     comment.UserID,
		UserName:   currentUser.Nickname,
		UserAvatar: currentUser.Avatar,
		Content:    comment.Content,
		CreateTime: comment.CreateTime,
		LikeCount:  0,
		IsLiked:    false,
		ReplyCount: 0,
		Replies:    []model.CommentReplyResponse{},
	}))
}

func (h *Handler) LikeComment(c *gin.Context) {
	userID := c.GetString("userId")
	commentID := c.Param("commentId")

	var existing model.Like
	if err := db.DB.Where("user_id = ? AND target_id = ? AND type = ?", userID, commentID, "comment").First(&existing).Error; err == nil {
		c.JSON(200, model.Success(nil))
		return
	}

	like := model.Like{
		UserID:   userID,
		TargetID: commentID,
		Type:     "comment",
	}

	if err := db.DB.Create(&like).Error; err != nil {
		c.JSON(500, model.Error(1005, "服务器内部错误"))
		return
	}

	db.DB.Model(&model.Comment{}).Where("comment_id = ?", commentID).UpdateColumn("like_count", gorm.Expr("like_count + 1"))

	c.JSON(200, model.Success(nil))
}

func (h *Handler) UnlikeComment(c *gin.Context) {
	userID := c.GetString("userId")
	commentID := c.Param("commentId")

	if err := db.DB.Where("user_id = ? AND target_id = ? AND type = ?", userID, commentID, "comment").Delete(&model.Like{}).Error; err != nil {
		c.JSON(500, model.Error(1005, "服务器内部错误"))
		return
	}

	db.DB.Model(&model.Comment{}).Where("comment_id = ?", commentID).UpdateColumn("like_count", gorm.Expr("like_count - 1"))

	c.JSON(200, model.Success(nil))
}