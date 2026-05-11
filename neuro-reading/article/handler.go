package article

import (
	"fmt"
	"io"
	"neuro-reading/config"
	"neuro-reading/db"
	"neuro-reading/model"
	"neuro-reading/utils"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	cfg        *config.Config
	articleDir string
}

func NewHandler(cfg *config.Config) *Handler {
	articleDir := cfg.ArticleDir
	if articleDir == "" {
		articleDir = "./articles"
	}

	if err := os.MkdirAll(articleDir, 0755); err != nil {
		fmt.Printf("创建文章目录失败: %v\n", err)
	}

	return &Handler{
		cfg:        cfg,
		articleDir: articleDir,
	}
}

func (h *Handler) getArticleDir(articleID string) string {
	return path.Join(h.articleDir, articleID)
}

func (h *Handler) getChaptersDir(articleID string) string {
	return path.Join(h.getArticleDir(articleID), "chapters")
}

func (h *Handler) getChapterPath(articleID string, index int) string {
	return path.Join(h.getChaptersDir(articleID), fmt.Sprintf("%d.txt", index))
}

func (h *Handler) readChapterContent(articleID string, index int) (string, error) {
	chapterPath := h.getChapterPath(articleID, index)
	data, err := os.ReadFile(chapterPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (h *Handler) List(c *gin.Context) {
	var req model.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req.Page = 1
		req.PageSize = 10
	}

	var total int64
	db.DB.Model(&model.Article{}).Where("status = ?", "published").Count(&total)

	var articles []model.Article
	offset := (req.Page - 1) * req.PageSize
	if err := db.DB.Where("status = ?", "published").
		Order("updated_at DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&articles).Error; err != nil {
		c.JSON(500, model.Error(1005, "读取文章列表失败"))
		return
	}

	list := make([]model.ArticleIndex, len(articles))
	for i, article := range articles {
		list[i] = model.ArticleIndex{
			ArticleID:      article.ArticleID,
			CreatorID:      article.CreatorID,
			Title:          article.Title,
			Author:         article.Author,
			Summary:        article.Summary,
			Cover:          article.Cover,
			WordCount:      article.WordCount,
			ChapterCount:   article.ChapterCount,
			Tags:           strings.Split(article.Tags, ","),
			Status:         article.Status,
			LastUpdateTime: article.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	c.JSON(200, model.PageSuccess(list, total, req.Page, req.PageSize))
}

func (h *Handler) Detail(c *gin.Context) {
	articleID := c.Param("articleId")
	if articleID == "" {
		c.JSON(400, model.Error(1001, "参数错误"))
		return
	}

	var article model.Article
	if err := db.DB.Where("article_id = ?", articleID).First(&article).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, model.Error(1004, "文章不存在"))
			return
		}
		c.JSON(500, model.Error(1005, "读取文章失败"))
		return
	}

	var chapters []model.Chapter
	db.DB.Where("book_id = ?", articleID).Order("`index` ASC").Find(&chapters)

	chapterMetas := make([]model.ChapterMeta, len(chapters))
	for i, ch := range chapters {
		chapterMetas[i] = model.ChapterMeta{
			Index:     ch.Index,
			ChapterID: ch.ChapterID,
			Title:     ch.Title,
			WordCount: ch.WordCount,
		}
	}

	meta := model.ArticleMeta{
		ArticleID:      article.ArticleID,
		CreatorID:      article.CreatorID,
		Title:          article.Title,
		Author:         article.Author,
		Summary:        article.Summary,
		Cover:          article.Cover,
		Tags:           strings.Split(article.Tags, ","),
		WordCount:      article.WordCount,
		ChapterCount:   article.ChapterCount,
		Status:         article.Status,
		PublishTime:    article.PublishTime.Format("2006-01-02 15:04:05"),
		LastUpdateTime: article.UpdatedAt.Format("2006-01-02 15:04:05"),
		Chapters:       chapterMetas,
	}

	c.JSON(200, model.Success(meta))
}

func (h *Handler) Chapter(c *gin.Context) {
	articleID := c.Param("articleId")
	chapterIndexStr := c.Param("chapterIndex")

	if articleID == "" || chapterIndexStr == "" {
		c.JSON(400, model.Error(1001, "参数错误"))
		return
	}

	chapterIndex, err := strconv.Atoi(chapterIndexStr)
	if err != nil || chapterIndex < 0 {
		c.JSON(400, model.Error(1001, "章节索引无效"))
		return
	}

	var article model.Article
	if err := db.DB.Where("article_id = ?", articleID).First(&article).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, model.Error(1004, "文章不存在"))
			return
		}
		c.JSON(500, model.Error(1005, "读取文章失败"))
		return
	}

	var chapter model.Chapter
	if err := db.DB.Where("book_id = ? AND `index` = ?", articleID, chapterIndex).First(&chapter).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, model.Error(1004, "章节不存在"))
			return
		}
		c.JSON(500, model.Error(1005, "读取章节失败"))
		return
	}

	content, err := h.readChapterContent(articleID, chapterIndex)
	if err != nil {
		c.JSON(500, model.Error(1005, "读取章节内容失败"))
		return
	}

	paragraphs := strings.Split(content, "\n\n")
	var filtered []string
	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p != "" {
			filtered = append(filtered, p)
		}
	}

	var totalChapters int64
	db.DB.Model(&model.Chapter{}).Where("book_id = ?", articleID).Count(&totalChapters)

	var prevID, nextID *int
	if chapterIndex > 0 {
		p := chapterIndex - 1
		prevID = &p
	}
	if chapterIndex < int(totalChapters)-1 {
		n := chapterIndex + 1
		nextID = &n
	}

	c.JSON(200, model.Success(model.ChapterContentResponse{
		ChapterID:         chapter.ChapterID,
		BookID:            articleID,
		Title:             chapter.Title,
		Content:           content,
		Paragraphs:        filtered,
		PrevChapterID:     prevID,
		NextChapterID:     nextID,
		ParagraphComments: make(map[string]int),
	}))
}

func (h *Handler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(400, model.Error(1001, "请选择要上传的文件"))
		return
	}
	defer file.Close()

	title := c.PostForm("title")
	author := c.PostForm("author")
	summary := c.PostForm("summary")
	tagsStr := c.PostForm("tags")

	if title == "" {
		c.JSON(400, model.Error(1001, "标题不能为空"))
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".txt" && ext != ".md" {
		c.JSON(400, model.Error(1001, "仅支持 txt 或 md 格式"))
		return
	}

	if header.Size > 10*1024*1024 {
		c.JSON(400, model.Error(1001, "文件大小不能超过 10MB"))
		return
	}

	content, err := io.ReadAll(file)
	if err != nil {
		c.JSON(500, model.Error(1005, "读取文件失败"))
		return
	}

	text := string(content)
	parsedChapters := utils.ParseChapters(text, ext)

	if len(parsedChapters) == 0 {
		parsedChapters = []model.ChapterMeta{
			{
				Index:     0,
				ChapterID: utils.GenerateChapterID(),
				Title:     "正文",
				WordCount: len([]rune(text)),
				Content:   text,
			},
		}
	}

	articleID := utils.GenerateArticleID()
	chaptersDir := h.getChaptersDir(articleID)

	if err := os.MkdirAll(chaptersDir, 0755); err != nil {
		c.JSON(500, model.Error(1005, "创建目录失败"))
		return
	}

	now := time.Now()
	totalWordCount := 0

	tags := tagsStr
	if tagsStr == "" {
		tags = ""
	}

	article := model.Article{
		ArticleID:    articleID,
		Title:        title,
		Author:       author,
		Summary:      summary,
		Tags:         tags,
		WordCount:    0,
		ChapterCount: len(parsedChapters),
		Status:       "published",
		PublishTime:  now,
	}

	if err := db.DB.Create(&article).Error; err != nil {
		c.JSON(500, model.Error(1005, "保存文章失败"))
		return
	}

	for i, ch := range parsedChapters {
		ch.WordCount = len([]rune(ch.Content))
		totalWordCount += ch.WordCount

		chapter := model.Chapter{
			ChapterID: ch.ChapterID,
			BookID:    articleID,
			Title:     ch.Title,
			Index:     i,
			WordCount: ch.WordCount,
		}

		if err := db.DB.Create(&chapter).Error; err != nil {
			c.JSON(500, model.Error(1005, "保存章节失败"))
			return
		}

		chapterPath := h.getChapterPath(articleID, i)
		if err := os.WriteFile(chapterPath, []byte(ch.Content), 0644); err != nil {
			c.JSON(500, model.Error(1005, "保存章节文件失败"))
			return
		}
	}

	db.DB.Model(&article).Updates(map[string]interface{}{
		"word_count": totalWordCount,
	})

	c.JSON(200, model.Success(gin.H{
		"articleId": articleID,
		"title":     title,
		"chapters":  len(parsedChapters),
	}))
}

func (h *Handler) Delete(c *gin.Context) {
	articleID := c.Param("articleId")
	if articleID == "" {
		c.JSON(400, model.Error(1001, "参数错误"))
		return
	}

	var article model.Article
	if err := db.DB.Where("article_id = ?", articleID).First(&article).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, model.Error(1004, "文章不存在"))
			return
		}
		c.JSON(500, model.Error(1005, "查询文章失败"))
		return
	}

	if err := db.DB.Where("book_id = ?", articleID).Delete(&model.Chapter{}).Error; err != nil {
		c.JSON(500, model.Error(1005, "删除章节失败"))
		return
	}

	if err := db.DB.Delete(&article).Error; err != nil {
		c.JSON(500, model.Error(1005, "删除文章失败"))
		return
	}

	articleDir := h.getArticleDir(articleID)
	os.RemoveAll(articleDir)

	c.JSON(200, model.Success(nil))
}

func (h *Handler) Search(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(400, model.Error(1001, "搜索关键词不能为空"))
		return
	}

	var articles []model.Article
	keyword = "%" + strings.ToLower(keyword) + "%"
	if err := db.DB.Where("status = ? AND (LOWER(title) LIKE ? OR LOWER(author) LIKE ? OR LOWER(summary) LIKE ?)",
		"published", keyword, keyword, keyword).
		Order("updated_at DESC").
		Find(&articles).Error; err != nil {
		c.JSON(500, model.Error(1005, "搜索失败"))
		return
	}

	results := make([]model.ArticleIndex, len(articles))
	for i, article := range articles {
		results[i] = model.ArticleIndex{
			ArticleID:      article.ArticleID,
			CreatorID:      article.CreatorID,
			Title:          article.Title,
			Author:         article.Author,
			Summary:        article.Summary,
			Cover:          article.Cover,
			WordCount:      article.WordCount,
			ChapterCount:   article.ChapterCount,
			Tags:           strings.Split(article.Tags, ","),
			Status:         article.Status,
			LastUpdateTime: article.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	c.JSON(200, model.Success(results))
}
