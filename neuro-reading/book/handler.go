package book

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

func (h *Handler) buildBookResponse(book model.Book, author model.Author) model.BookResponse {
	var tags []string
	if book.Tags != "" {
		tags = strings.Split(book.Tags, ",")
	}
	return model.BookResponse{
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
	}
}

func (h *Handler) GetRecommend(c *gin.Context) {
	var req model.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req.Page = 1
		req.PageSize = 10
	}

	var books []model.Book
	var total int64

	db.DB.Model(&model.Book{}).Count(&total)
	db.DB.Order("rating DESC").Offset(req.GetOffset()).Limit(req.GetLimit()).Find(&books)

	list := h.buildBookListWithAuthors(books)
	c.JSON(200, model.PageSuccess(list, total, req.Page, req.PageSize))
}

func (h *Handler) GetHot(c *gin.Context) {
	var req model.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req.Page = 1
		req.PageSize = 10
	}

	var books []model.Book
	var total int64

	db.DB.Model(&model.Book{}).Count(&total)
	db.DB.Order("rating_count DESC").Offset(req.GetOffset()).Limit(req.GetLimit()).Find(&books)

	list := h.buildBookListWithAuthors(books)
	c.JSON(200, model.PageSuccess(list, total, req.Page, req.PageSize))
}

func (h *Handler) GetLatest(c *gin.Context) {
	var req model.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req.Page = 1
		req.PageSize = 10
	}

	var books []model.Book
	var total int64

	db.DB.Model(&model.Book{}).Count(&total)
	db.DB.Order("last_update_time DESC").Offset(req.GetOffset()).Limit(req.GetLimit()).Find(&books)

	list := h.buildBookListWithAuthors(books)
	c.JSON(200, model.PageSuccess(list, total, req.Page, req.PageSize))
}

func (h *Handler) buildBookListWithAuthors(books []model.Book) []model.BookResponse {
	if len(books) == 0 {
		return []model.BookResponse{}
	}

	authorIDs := make([]string, 0, len(books))
	for _, book := range books {
		authorIDs = append(authorIDs, book.AuthorID)
	}

	var authors []model.Author
	authorMap := make(map[string]model.Author)
	if len(authorIDs) > 0 {
		db.DB.Where("author_id IN ?", authorIDs).Find(&authors)
		for _, author := range authors {
			authorMap[author.AuthorID] = author
		}
	}

	list := make([]model.BookResponse, 0, len(books))
	for _, book := range books {
		author := authorMap[book.AuthorID]
		list = append(list, h.buildBookResponse(book, author))
	}
	return list
}

func (h *Handler) GetDetail(c *gin.Context) {
	bookID := c.Param("bookId")
	userID := c.GetString("userId")

	var book model.Book
	if err := db.DB.Where("book_id = ?", bookID).First(&book).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, model.Error(1004, "资源不存在"))
			return
		}
		c.JSON(500, model.Error(1005, "服务器内部错误"))
		return
	}

	var author model.Author
	db.DB.Where("author_id = ?", book.AuthorID).First(&author)

	var chapters []model.Chapter
	db.DB.Where("book_id = ?", bookID).Order("index ASC").Limit(10).Find(&chapters)

	var chapterResponses []model.ChapterResponse
	for _, ch := range chapters {
		chapterResponses = append(chapterResponses, model.ChapterResponse{
			ChapterID:  ch.ChapterID,
			BookID:     ch.BookID,
			Title:      ch.Title,
			Index:      ch.Index,
			WordCount:  ch.WordCount,
			IsVip:      ch.IsVip,
			IsTrial:    ch.IsTrial,
			UpdateTime: ch.UpdateTime,
		})
	}

	var isInBookshelf bool
	if userID != "" {
		var bookshelfItem model.BookshelfItem
		if err := db.DB.Where("user_id = ? AND book_id = ?", userID, bookID).First(&bookshelfItem).Error; err == nil {
			isInBookshelf = true
		}
	}

	var tags []string
	if book.Tags != "" {
		tags = strings.Split(book.Tags, ",")
	}

	c.JSON(200, model.Success(model.BookDetailResponse{
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
		Rating:         book.Rating,
		RatingCount:    book.RatingCount,
		CommentCount:   book.CommentCount,
		IsInBookshelf:  isInBookshelf,
		Chapters:       chapterResponses,
	}))
}

func (h *Handler) GetChapters(c *gin.Context) {
	bookID := c.Param("bookId")

	var chapters []model.Chapter
	if err := db.DB.Where("book_id = ?", bookID).Order("index ASC").Find(&chapters).Error; err != nil {
		c.JSON(500, model.Error(1005, "服务器内部错误"))
		return
	}

	list := make([]model.ChapterResponse, 0)
	for _, ch := range chapters {
		list = append(list, model.ChapterResponse{
			ChapterID:  ch.ChapterID,
			BookID:     ch.BookID,
			Title:      ch.Title,
			Index:      ch.Index,
			WordCount:  ch.WordCount,
			IsVip:      ch.IsVip,
			IsTrial:    ch.IsTrial,
			UpdateTime: ch.UpdateTime,
		})
	}

	c.JSON(200, model.Success(list))
}

func (h *Handler) GetChapterContent(c *gin.Context) {
	bookID := c.Param("bookId")
	chapterID := c.Param("chapterId")

	var chapter model.Chapter
	if err := db.DB.Where("chapter_id = ? AND book_id = ?", chapterID, bookID).First(&chapter).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, model.Error(1004, "资源不存在"))
			return
		}
		c.JSON(500, model.Error(1005, "服务器内部错误"))
		return
	}

	var prevChapter, nextChapter *model.Chapter
	db.DB.Where("book_id = ? AND index < ?", bookID, chapter.Index).Order("index DESC").First(&prevChapter)
	db.DB.Where("book_id = ? AND index > ?", bookID, chapter.Index).Order("index ASC").First(&nextChapter)

	var paragraphComments []model.ParagraphComment
	db.DB.Where("chapter_id = ?", chapterID).Find(&paragraphComments)

	paragraphCommentMap := make(map[string]int)
	for _, pc := range paragraphComments {
		key := string(rune(pc.ParagraphIndex))
		paragraphCommentMap[key]++
	}

	var prevID, nextID *int
	if prevChapter != nil {
		prevID = &prevChapter.Index
	}
	if nextChapter != nil {
		nextID = &nextChapter.Index
	}

	c.JSON(200, model.Success(model.ChapterContentResponse{
		ChapterID:         chapter.ChapterID,
		BookID:            chapter.BookID,
		Title:             chapter.Title,
		Content:           chapter.Content,
		PrevChapterID:     prevID,
		NextChapterID:     nextID,
		ParagraphComments: paragraphCommentMap,
	}))
}

func (h *Handler) Search(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(400, model.Error(1001, "参数错误"))
		return
	}

	var req model.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req.Page = 1
		req.PageSize = 20
	}

	var books []model.Book
	var total int64

	db.DB.Model(&model.Book{}).Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%").Count(&total)
	db.DB.Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%").
		Offset(req.GetOffset()).Limit(req.GetLimit()).Find(&books)

	list := h.buildBookListWithAuthors(books)
	c.JSON(200, model.PageSuccess(list, total, req.Page, req.PageSize))
}

func (h *Handler) GetHotKeywords(c *gin.Context) {
	keywords := []string{"斗破苍穹", "诡秘之主", "凡人修仙传", "斗罗大陆", "全职高手", "庆余年"}
	c.JSON(200, model.Success(keywords))
}
