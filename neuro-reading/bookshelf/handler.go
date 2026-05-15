package bookshelf

import (
	"neuro-reading/db"
	"neuro-reading/model"
	"neuro-reading/user"
	"neuro-reading/utils"

	"github.com/gin-gonic/gin"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) GetList(c *gin.Context) {
	userID := c.GetString("userId")

	var req model.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req.Page = 1
		req.PageSize = 20
	}

	var items []model.BookshelfItem
	var total int64

	db.DB.Model(&model.BookshelfItem{}).Where("user_id = ?", userID).Count(&total)
	db.DB.Where("user_id = ?", userID).Offset(req.GetOffset()).Limit(req.GetLimit()).Find(&items)

	list := make([]model.BookshelfItemResponse, 0)
	for _, item := range items {
		var article model.Article
		if err := db.DB.Where("article_id = ?", item.BookID).First(&article).Error; err != nil {
			continue
		}

		var chapterIndex int
		var chapter model.Chapter
		if err := db.DB.Where("chapter_id = ?", item.ChapterID).First(&chapter).Error; err == nil {
			chapterIndex = chapter.Index
		}

		list = append(list, model.BookshelfItemResponse{
			ArticleID:       item.BookID,
			Title:           article.Title,
			Author:          article.Author,
			Cover:           article.Cover,
			LastReadChapter: item.LastReadChapter,
			LastReadTime:    item.LastReadTime,
			Progress:        item.Progress,
			ChapterIndex:    chapterIndex,
			IsFinished:      item.IsFinished,
		})
	}

	c.JSON(200, model.PageSuccess(list, total, req.Page, req.PageSize))
}

func (h *Handler) Add(c *gin.Context) {
	userID := c.GetString("userId")
	bookID := c.Param("bookId")

	if bookID == "" {
		c.JSON(400, model.Error(1001, "参数错误"))
		return
	}

	var existing model.BookshelfItem
	if err := db.DB.Where("user_id = ? AND book_id = ?", userID, bookID).First(&existing).Error; err == nil {
		c.JSON(200, model.Success(nil))
		return
	}

	item := model.BookshelfItem{
		UserID: userID,
		BookID: bookID,
	}

	if err := db.DB.Create(&item).Error; err != nil {
		c.JSON(500, model.Error(1005, "服务器内部错误"))
		return
	}

	var article model.Article
	if err := db.DB.Where("article_id = ?", bookID).First(&article).Error; err == nil {
		var chapter model.Chapter
		chapterTitle := ""
		if err := db.DB.Where("book_id = ?", bookID).Order("`index` ASC").First(&chapter).Error; err == nil {
			chapterTitle = chapter.Title
		}
		user.RecordReadingHistory(userID, bookID, 0, 0, 0, chapterTitle)
	}

	c.JSON(200, model.Success(nil))
}

func (h *Handler) Remove(c *gin.Context) {
	userID := c.GetString("userId")
	bookID := c.Param("bookId")

	if bookID == "" {
		c.JSON(400, model.Error(1001, "参数错误"))
		return
	}

	if err := db.DB.Where("user_id = ? AND book_id = ?", userID, bookID).Delete(&model.BookshelfItem{}).Error; err != nil {
		c.JSON(500, model.Error(1005, "服务器内部错误"))
		return
	}

	c.JSON(200, model.Success(nil))
}

func (h *Handler) UpdateProgress(c *gin.Context) {
	userID := c.GetString("userId")
	bookID := c.Param("bookId")

	if bookID == "" {
		c.JSON(400, model.Error(1001, "参数错误"))
		return
	}

	var req model.UpdateProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, model.Error(1001, "参数错误"))
		return
	}

	var chapter model.Chapter
	if err := db.DB.Where("book_id = ? AND `index` = ?", bookID, req.ChapterIndex).First(&chapter).Error; err != nil {
		c.JSON(400, model.Error(1004, "资源不存在"))
		return
	}

	var item model.BookshelfItem
	if err := db.DB.Where("user_id = ? AND book_id = ?", userID, bookID).First(&item).Error; err != nil {
		item = model.BookshelfItem{
			UserID: userID,
			BookID: bookID,
		}
		if err := db.DB.Create(&item).Error; err != nil {
			c.JSON(500, model.Error(1005, "服务器内部错误"))
			return
		}
	}

	updates := map[string]interface{}{
		"chapter_id":        chapter.ChapterID,
		"progress":          req.Progress,
		"position":          req.Position,
		"last_read_chapter": chapter.Title,
		"last_read_time":    utils.CurrentTime(),
		"is_finished":       req.Progress >= 100,
	}

	if err := db.DB.Model(&item).Updates(updates).Error; err != nil {
		c.JSON(500, model.Error(1005, "服务器内部错误"))
		return
	}

	user.RecordReadingHistory(userID, bookID, chapter.Index, req.Progress, req.Position, chapter.Title)

	c.JSON(200, model.Success(nil))
}