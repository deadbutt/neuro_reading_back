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

	var author model.Author
	if err := db.DB.Where("author_id = ?", authorID).First(&author).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, model.Error(1004, "资源不存在"))
			return
		}
		c.JSON(500, model.Error(1005, "服务器内部错误"))
		return
	}

	var worksCount int64
	db.DB.Model(&model.Book{}).Where("author_id = ?", authorID).Count(&worksCount)

	var followersCount int64
	db.DB.Model(&model.Follow{}).Where("author_id = ?", authorID).Count(&followersCount)

	var totalWords int64
	var books []model.Book
	db.DB.Where("author_id = ?", authorID).Find(&books)
	for _, book := range books {
		totalWords += book.WordCount
	}

	var isFollowing bool
	if userID != "" {
		var follow model.Follow
		if err := db.DB.Where("user_id = ? AND author_id = ?", userID, authorID).First(&follow).Error; err == nil {
			isFollowing = true
		}
	}

	c.JSON(200, model.Success(model.AuthorProfileResponse{
		AuthorID:       author.AuthorID,
		Name:           author.Name,
		Avatar:         author.Avatar,
		Description:    author.Description,
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

	var books []model.Book
	var total int64

	db.DB.Model(&model.Book{}).Where("author_id = ?", authorID).Count(&total)
	db.DB.Where("author_id = ?", authorID).Offset(req.GetOffset()).Limit(req.GetLimit()).Find(&books)

	var author model.Author
	db.DB.Where("author_id = ?", authorID).First(&author)

	list := make([]model.BookResponse, 0)
	for _, book := range books {
		var tags []string
		if book.Tags != "" {
			tags = strings.Split(book.Tags, ",")
		}
		list = append(list, model.BookResponse{
			BookID: book.BookID,
			Title:  book.Title,
			Author: model.AuthorResponse{
				AuthorID:    author.AuthorID,
				Name:        author.Name,
				Avatar:      author.Avatar,
				Description: author.Description,
			},
			Cover:          book.Cover,
			Description:    book.Description,
			HotText:        book.HotText,
			WordCount:      book.WordCount,
			ChapterCount:   book.ChapterCount,
			Status:         book.Status,
			Tags:           tags,
			LastUpdateTime: book.LastUpdateTime,
			IsVip:          book.IsVip,
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