package feed

import (
	"neuro-reading/db"
	"neuro-reading/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) GetFeed(c *gin.Context) {
	userID := c.GetString("userId")

	var req model.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req.Page = 1
		req.PageSize = 20
	}

	var follows []model.Follow
	db.DB.Where("user_id = ?", userID).Find(&follows)

	var authorIDs []string
	for _, f := range follows {
		authorIDs = append(authorIDs, f.AuthorID)
	}

	var activities []model.FeedActivity
	var total int64

	if len(authorIDs) > 0 {
		db.DB.Model(&model.FeedActivity{}).Where("author_id IN ?", authorIDs).Count(&total)
		db.DB.Where("author_id IN ?", authorIDs).Order("publish_time DESC").Offset(req.GetOffset()).Limit(req.GetLimit()).Find(&activities)
	}

	list := make([]model.FeedActivityResponse, 0)
	for _, activity := range activities {
		var isLiked bool
		var like model.Like
		if err := db.DB.Where("user_id = ? AND target_id = ? AND type = ?", userID, activity.FeedID, "feed").First(&like).Error; err == nil {
			isLiked = true
		}

		list = append(list, model.FeedActivityResponse{
			FeedID:          activity.FeedID,
			AuthorID:        activity.AuthorID,
			AuthorName:      activity.AuthorName,
			AuthorAvatar:    activity.AuthorAvatar,
			PublishTime:     activity.PublishTime,
			ActivityContent: activity.ActivityContent,
			BookID:          activity.BookID,
			BookCover:       activity.BookCover,
			ChapterPreview:  activity.ChapterPreview,
			ReadHeat:        activity.ReadHeat,
			LikeCount:       activity.LikeCount,
			CommentCount:    activity.CommentCount,
			IsLiked:         isLiked,
		})
	}

	c.JSON(200, model.PageSuccess(list, total, req.Page, req.PageSize))
}

func (h *Handler) LikeFeed(c *gin.Context) {
	userID := c.GetString("userId")
	feedID := c.Param("feedId")

	var existing model.Like
	if err := db.DB.Where("user_id = ? AND target_id = ? AND type = ?", userID, feedID, "feed").First(&existing).Error; err == nil {
		c.JSON(200, model.Success(nil))
		return
	}

	like := model.Like{
		UserID:   userID,
		TargetID: feedID,
		Type:     "feed",
	}

	if err := db.DB.Create(&like).Error; err != nil {
		c.JSON(500, model.Error(1005, "服务器内部错误"))
		return
	}

	db.DB.Model(&model.FeedActivity{}).Where("feed_id = ?", feedID).UpdateColumn("like_count", gorm.Expr("like_count + 1"))

	c.JSON(200, model.Success(nil))
}

func (h *Handler) UnlikeFeed(c *gin.Context) {
	userID := c.GetString("userId")
	feedID := c.Param("feedId")

	if err := db.DB.Where("user_id = ? AND target_id = ? AND type = ?", userID, feedID, "feed").Delete(&model.Like{}).Error; err != nil {
		c.JSON(500, model.Error(1005, "服务器内部错误"))
		return
	}

	db.DB.Model(&model.FeedActivity{}).Where("feed_id = ?", feedID).UpdateColumn("like_count", gorm.Expr("like_count - 1"))

	c.JSON(200, model.Success(nil))
}