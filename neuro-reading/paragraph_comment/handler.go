package paragraph_comment

import (
	"neuro-reading/db"
	"neuro-reading/model"
	"neuro-reading/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

type CreateParagraphCommentRequest struct {
	ParagraphIndex int    `json:"paragraphIndex" binding:"min=0"`
	ParagraphIndex2 int   `json:"paragraph_index" binding:"min=0"`
	Content        string `json:"content" binding:"required,min=1,max=500"`
}

func (h *Handler) GetList(c *gin.Context) {
	chapterID := c.Param("chapterId")
	paragraphIndex := c.Query("paragraphIndex")

	var comments []model.ParagraphComment
	query := db.DB.Where("chapter_id = ?", chapterID)
	if paragraphIndex != "" {
		query = query.Where("paragraph_index = ?", paragraphIndex)
	}
	query.Order("create_time DESC").Find(&comments)

	userID := c.GetString("userId")
	var list []model.ParagraphCommentResponse
	for _, comment := range comments {
		var user model.User
		db.DB.Where("user_id = ?", comment.UserID).First(&user)

		var isLiked bool
		if userID != "" {
			var like model.Like
			if err := db.DB.Where("user_id = ? AND target_id = ? AND type = ?", userID, comment.CommentID, "paragraph_comment").First(&like).Error; err == nil {
				isLiked = true
			}
		}

		list = append(list, model.ParagraphCommentResponse{
			CommentID:      comment.CommentID,
			ChapterID:      comment.ChapterID,
			ParagraphIndex: comment.ParagraphIndex,
			UserID:         comment.UserID,
			UserName:       user.Nickname,
			UserAvatar:     user.Avatar,
			Content:        comment.Content,
			CreateTime:     comment.CreateTime,
			LikeCount:      comment.LikeCount,
			IsLiked:        isLiked,
		})
	}

	c.JSON(200, model.Success(list))
}

func (h *Handler) Create(c *gin.Context) {
	userID := c.GetString("userId")
	chapterID := c.Param("chapterId")

	var req CreateParagraphCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, model.Error(1001, "参数错误"))
		return
	}

	// 兼容 paragraphIndex 和 paragraph_index
	paragraphIndex := req.ParagraphIndex
	if paragraphIndex == 0 && req.ParagraphIndex2 > 0 {
		paragraphIndex = req.ParagraphIndex2
	}

	comment := model.ParagraphComment{
		CommentID:      utils.GenerateParagraphCommentID(),
		ChapterID:      chapterID,
		ParagraphIndex: paragraphIndex,
		UserID:         userID,
		Content:        req.Content,
		CreateTime:     utils.CurrentTime(),
	}

	if err := db.DB.Create(&comment).Error; err != nil {
		c.JSON(500, model.Error(1005, "服务器内部错误"))
		return
	}

	var user model.User
	db.DB.Where("user_id = ?", userID).First(&user)

	c.JSON(200, model.Success(model.ParagraphCommentResponse{
		CommentID:      comment.CommentID,
		ChapterID:      comment.ChapterID,
		ParagraphIndex: comment.ParagraphIndex,
		UserID:         comment.UserID,
		UserName:       user.Nickname,
		UserAvatar:     user.Avatar,
		Content:        comment.Content,
		CreateTime:     comment.CreateTime,
		LikeCount:      0,
		IsLiked:        false,
	}))
}

func (h *Handler) Like(c *gin.Context) {
	userID := c.GetString("userId")
	commentID := c.Param("commentId")

	var existing model.Like
	if err := db.DB.Where("user_id = ? AND target_id = ? AND type = ?", userID, commentID, "paragraph_comment").First(&existing).Error; err == nil {
		c.JSON(200, model.Success(nil))
		return
	}

	like := model.Like{
		UserID:   userID,
		TargetID: commentID,
		Type:     "paragraph_comment",
	}

	if err := db.DB.Create(&like).Error; err != nil {
		c.JSON(500, model.Error(1005, "服务器内部错误"))
		return
	}

	db.DB.Model(&model.ParagraphComment{}).Where("comment_id = ?", commentID).UpdateColumn("like_count", gorm.Expr("like_count + 1"))

	c.JSON(200, model.Success(nil))
}

func (h *Handler) Unlike(c *gin.Context) {
	userID := c.GetString("userId")
	commentID := c.Param("commentId")

	if err := db.DB.Where("user_id = ? AND target_id = ? AND type = ?", userID, commentID, "paragraph_comment").Delete(&model.Like{}).Error; err != nil {
		c.JSON(500, model.Error(1005, "服务器内部错误"))
		return
	}

	db.DB.Model(&model.ParagraphComment{}).Where("comment_id = ?", commentID).UpdateColumn("like_count", gorm.Expr("like_count - 1"))

	c.JSON(200, model.Success(nil))
}