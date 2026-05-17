package author

import (
	"neuro-reading/db"
	"neuro-reading/model"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) GetProfile(c *gin.Context) {
	authorID := c.Param("authorId")
	userID := c.GetString("userId")

	var user model.User
	if err := db.DB.Where("user_id = ?", authorID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, model.Error(1004, "用户不存在"))
			return
		}
		c.JSON(500, model.Error(1005, "服务器内部错误"))
		return
	}

	var worksCount int64
	db.DB.Model(&model.Article{}).Where("creator_id = ?", authorID).Count(&worksCount)

	var followersCount int64
	db.DB.Model(&model.Follow{}).Where("author_id = ?", authorID).Count(&followersCount)

	var totalWords int64
	var articles []model.Article
	db.DB.Where("creator_id = ?", authorID).Find(&articles)
	for _, article := range articles {
		totalWords += int64(article.WordCount)
	}

	var isFollowing bool
	if userID != "" {
		var follow model.Follow
		if err := db.DB.Where("user_id = ? AND author_id = ?", userID, authorID).First(&follow).Error; err == nil {
			isFollowing = true
		}
	}

	c.JSON(200, model.Success(model.AuthorProfileResponse{
		AuthorID:       user.UserID,
		Name:           user.Nickname,
		Avatar:         user.Avatar,
		Description:    user.Bio,
		WorksCount:     int(worksCount),
		FollowersCount: followersCount,
		TotalWords:     totalWords,
		IsFollowing:    isFollowing,
	}))
}

func (h *Handler) GetWorks(c *gin.Context) {
	authorID := c.Param("authorId")

	var req model.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req.Page = 1
		req.PageSize = 20
	}

	var articles []model.Article
	var total int64

	db.DB.Model(&model.Article{}).Where("creator_id = ?", authorID).Count(&total)
	db.DB.Where("creator_id = ?", authorID).Offset(req.GetOffset()).Limit(req.GetLimit()).Find(&articles)

	var user model.User
	db.DB.Where("user_id = ?", authorID).First(&user)

	list := make([]model.BookResponse, 0)
	for _, article := range articles {
		var tags []string
		if article.Tags != "" {
			tags = strings.Split(article.Tags, ",")
		}
		list = append(list, model.BookResponse{
			BookID: article.ArticleID,
			Title:  article.Title,
			Author: model.AuthorResponse{
				AuthorID:    user.UserID,
				Name:        user.Nickname,
				Avatar:      user.Avatar,
				Description: user.Bio,
			},
			Cover:          article.Cover,
			Description:    article.Summary,
			WordCount:      int64(article.WordCount),
			ChapterCount:   article.ChapterCount,
			Status:         article.Status,
			Tags:           tags,
			LastUpdateTime: article.UpdatedAt.Format("2006-01-02 15:04:05"),
			IsVip:          false,
		})
	}

	c.JSON(200, model.PageSuccess(list, total, req.Page, req.PageSize))
}

func (h *Handler) GetActivities(c *gin.Context) {
	authorID := c.Param("authorId")

	var req model.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req.Page = 1
		req.PageSize = 20
	}

	var activities []model.FeedActivity
	var total int64

	db.DB.Model(&model.FeedActivity{}).Where("author_id = ?", authorID).Count(&total)
	db.DB.Where("author_id = ?", authorID).Order("publish_time DESC").Offset(req.GetOffset()).Limit(req.GetLimit()).Find(&activities)

	list := make([]model.AuthorActivityResponse, 0)
	for _, activity := range activities {
		list = append(list, model.AuthorActivityResponse{
			ActivityID:     activity.FeedID,
			AuthorID:       activity.AuthorID,
			AuthorName:     activity.AuthorName,
			AuthorAvatar:   activity.AuthorAvatar,
			Type:           "update",
			Content:        activity.ActivityContent,
			BookID:         activity.BookID,
			BookTitle:      "",
			ChapterTitle:   "",
			ChapterPreview: activity.ChapterPreview,
			ReadHeat:       activity.ReadHeat,
			CreateTime:     activity.PublishTime,
			LikeCount:      activity.LikeCount,
			CommentCount:   activity.CommentCount,
		})
	}

	c.JSON(200, model.PageSuccess(list, total, req.Page, req.PageSize))
}

func (h *Handler) GetFollowStatus(c *gin.Context) {
	userID := c.GetString("userId")
	authorID := c.Param("authorId")

	if userID == "" {
		c.JSON(200, model.Success(map[string]bool{"isFollowing": false}))
		return
	}

	var isFollowing bool
	var follow model.Follow
	if err := db.DB.Where("user_id = ? AND author_id = ?", userID, authorID).First(&follow).Error; err == nil {
		isFollowing = true
	}

	c.JSON(200, model.Success(map[string]bool{"isFollowing": isFollowing}))
}
