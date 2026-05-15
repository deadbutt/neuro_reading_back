package user

import (
	"neuro-reading/db"
	"neuro-reading/model"
	"neuro-reading/utils"

	"github.com/gin-gonic/gin"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) GetProfile(c *gin.Context) {
	userID := c.GetString("userId")

	var user model.User
	if err := db.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		c.JSON(500, model.Error(1005, "服务器内部错误"))
		return
	}

	var followingCount int64
	db.DB.Model(&model.Follow{}).Where("user_id = ?", userID).Count(&followingCount)

	var bookshelfCount int64
	db.DB.Model(&model.BookshelfItem{}).Where("user_id = ?", userID).Count(&bookshelfCount)

	c.JSON(200, model.Success(model.UserProfileResponse{
		UserID:         user.UserID,
		Account:        user.Account,
		Nickname:       user.Nickname,
		Avatar:         user.Avatar,
		Bio:            user.Bio,
		Gender:         user.Gender,
		FollowingCount: followingCount,
		BookshelfCount: bookshelfCount,
		ReadDuration:   user.ReadDuration,
	}))
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	userID := c.GetString("userId")

	var req model.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, model.Error(1001, "参数错误"))
		return
	}

	updates := make(map[string]interface{})
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	if req.Bio != "" {
		updates["bio"] = req.Bio
	}
	if req.Gender != nil {
		updates["gender"] = *req.Gender
	}

	if len(updates) == 0 {
		c.JSON(400, model.Error(1001, "参数错误"))
		return
	}

	if err := db.DB.Model(&model.User{}).Where("user_id = ?", userID).Updates(updates).Error; err != nil {
		c.JSON(500, model.Error(1005, "服务器内部错误"))
		return
	}

	var user model.User
	db.DB.Where("user_id = ?", userID).First(&user)

	var followingCount int64
	db.DB.Model(&model.Follow{}).Where("user_id = ?", userID).Count(&followingCount)

	var bookshelfCount int64
	db.DB.Model(&model.BookshelfItem{}).Where("user_id = ?", userID).Count(&bookshelfCount)

	c.JSON(200, model.Success(model.UserProfileResponse{
		UserID:         user.UserID,
		Account:        user.Account,
		Nickname:       user.Nickname,
		Avatar:         user.Avatar,
		Bio:            user.Bio,
		Gender:         user.Gender,
		FollowingCount: followingCount,
		BookshelfCount: bookshelfCount,
		ReadDuration:   user.ReadDuration,
	}))
}

func (h *Handler) Follow(c *gin.Context) {
	userID := c.GetString("userId")
	authorID := c.Param("authorId")

	if authorID == "" {
		c.JSON(400, model.Error(1001, "参数错误"))
		return
	}

	var existing model.Follow
	if err := db.DB.Where("user_id = ? AND author_id = ?", userID, authorID).First(&existing).Error; err == nil {
		c.JSON(200, model.Success(nil))
		return
	}

	follow := model.Follow{
		UserID:   userID,
		AuthorID: authorID,
	}

	if err := db.DB.Create(&follow).Error; err != nil {
		c.JSON(500, model.Error(1005, "服务器内部错误"))
		return
	}

	c.JSON(200, model.Success(nil))
}

func (h *Handler) Unfollow(c *gin.Context) {
	userID := c.GetString("userId")
	authorID := c.Param("authorId")

	if authorID == "" {
		c.JSON(400, model.Error(1001, "参数错误"))
		return
	}

	if err := db.DB.Where("user_id = ? AND author_id = ?", userID, authorID).Delete(&model.Follow{}).Error; err != nil {
		c.JSON(500, model.Error(1005, "服务器内部错误"))
		return
	}

	c.JSON(200, model.Success(nil))
}

func (h *Handler) GetFollowing(c *gin.Context) {
	userID := c.GetString("userId")

	var req model.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req.Page = 1
		req.PageSize = 20
	}

	var follows []model.Follow
	var total int64

	db.DB.Model(&model.Follow{}).Where("user_id = ?", userID).Count(&total)
	db.DB.Where("user_id = ?", userID).Offset(req.GetOffset()).Limit(req.GetLimit()).Find(&follows)

	var authorIDs []string
	for _, f := range follows {
		authorIDs = append(authorIDs, f.AuthorID)
	}

	var authors []model.Author
	if len(authorIDs) > 0 {
		db.DB.Where("author_id IN ?", authorIDs).Find(&authors)
	}

	list := make([]model.AuthorResponse, 0)
	for _, a := range authors {
		list = append(list, model.AuthorResponse{
			AuthorID:    a.AuthorID,
			Name:        a.Name,
			Avatar:      a.Avatar,
			Description: a.Description,
		})
	}

	c.JSON(200, model.PageSuccess(list, total, req.Page, req.PageSize))
}

func (h *Handler) GetReadingHistory(c *gin.Context) {
	userID := c.GetString("userId")

	var req model.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req.Page = 1
		req.PageSize = 20
	}

	var histories []model.ReadingHistory
	var total int64

	db.DB.Model(&model.ReadingHistory{}).Where("user_id = ?", userID).Count(&total)
	db.DB.Where("user_id = ?", userID).Order("last_read_time DESC").Offset(req.GetOffset()).Limit(req.GetLimit()).Find(&histories)

	list := make([]model.ReadingHistoryResponse, 0)
	for _, h := range histories {
		list = append(list, model.ReadingHistoryResponse{
			HistoryID:    h.HistoryID,
			ArticleID:    h.ArticleID,
			Title:        h.Title,
			Author:       h.Author,
			Cover:        h.Cover,
			ChapterIndex: h.ChapterIndex,
			ChapterTitle: h.ChapterTitle,
			Progress:     h.Progress,
			Position:     h.Position,
			ReadTime:     h.ReadTime,
			LastReadTime: h.LastReadTime,
		})
	}

	c.JSON(200, model.PageSuccess(list, total, req.Page, req.PageSize))
}

func (h *Handler) DeleteReadingHistory(c *gin.Context) {
	userID := c.GetString("userId")
	historyID := c.Param("historyId")

	if historyID == "" {
		c.JSON(400, model.Error(1001, "参数错误"))
		return
	}

	if err := db.DB.Where("user_id = ? AND history_id = ?", userID, historyID).Delete(&model.ReadingHistory{}).Error; err != nil {
		c.JSON(500, model.Error(1005, "删除失败"))
		return
	}

	c.JSON(200, model.Success(nil))
}

func (h *Handler) ClearReadingHistory(c *gin.Context) {
	userID := c.GetString("userId")

	if err := db.DB.Where("user_id = ?", userID).Delete(&model.ReadingHistory{}).Error; err != nil {
		c.JSON(500, model.Error(1005, "清空失败"))
		return
	}

	c.JSON(200, model.Success(nil))
}

func RecordReadingHistory(userID, articleID string, chapterIndex, progress, position int, chapterTitle string) {
	var article model.Article
	if err := db.DB.Where("article_id = ?", articleID).First(&article).Error; err != nil {
		return
	}

	var existing model.ReadingHistory
	if err := db.DB.Where("user_id = ? AND article_id = ?", userID, articleID).First(&existing).Error; err == nil {
		db.DB.Model(&existing).Updates(map[string]interface{}{
			"chapter_index":  chapterIndex,
			"chapter_title":  chapterTitle,
			"progress":       progress,
			"position":       position,
			"last_read_time": utils.CurrentTime(),
		})
		return
	}

	var count int64
	db.DB.Model(&model.ReadingHistory{}).Where("user_id = ?", userID).Count(&count)
	if count >= 100 {
		var oldest model.ReadingHistory
		db.DB.Where("user_id = ?", userID).Order("last_read_time ASC").First(&oldest)
		db.DB.Delete(&oldest)
	}

	history := model.ReadingHistory{
		HistoryID:    utils.GenerateHistoryID(),
		UserID:       userID,
		ArticleID:    articleID,
		Title:        article.Title,
		Author:       article.Author,
		Cover:        article.Cover,
		ChapterIndex: chapterIndex,
		ChapterTitle: chapterTitle,
		Progress:     progress,
		Position:     position,
		ReadTime:     0,
		LastReadTime: utils.CurrentTime(),
	}

	db.DB.Create(&history)
}