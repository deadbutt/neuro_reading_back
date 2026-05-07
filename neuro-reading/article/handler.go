package article

import (
	"encoding/json"
	"fmt"
	"neuro-reading/config"
	"neuro-reading/model"
	"neuro-reading/utils"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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

func (h *Handler) getIndexPath() string {
	return path.Join(h.articleDir, "index.json")
}

func (h *Handler) getArticleDir(articleID string) string {
	return path.Join(h.articleDir, articleID)
}

func (h *Handler) getMetaPath(articleID string) string {
	return path.Join(h.getArticleDir(articleID), "meta.json")
}

func (h *Handler) getChaptersDir(articleID string) string {
	return path.Join(h.getArticleDir(articleID), "chapters")
}

func (h *Handler) getChapterPath(articleID string, index int) string {
	return path.Join(h.getChaptersDir(articleID), fmt.Sprintf("%d.txt", index))
}

func (h *Handler) readIndex() ([]model.ArticleIndex, error) {
	indexPath := h.getIndexPath()
	data, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []model.ArticleIndex{}, nil
		}
		return nil, err
	}

	var index []model.ArticleIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, err
	}
	return index, nil
}

func (h *Handler) writeIndex(index []model.ArticleIndex) error {
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(h.getIndexPath(), data, 0644)
}

func (h *Handler) readMeta(articleID string) (*model.ArticleMeta, error) {
	metaPath := h.getMetaPath(articleID)
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, err
	}

	var meta model.ArticleMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

func (h *Handler) writeMeta(articleID string, meta *model.ArticleMeta) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(h.getMetaPath(articleID), data, 0644)
}

func (h *Handler) readChapter(articleID string, index int) (string, error) {
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

	index, err := h.readIndex()
	if err != nil {
		c.JSON(500, model.Error(1005, "读取文章列表失败"))
		return
	}

	var published []model.ArticleIndex
	for _, item := range index {
		if item.Status == "published" || item.Status == "" {
			published = append(published, item)
		}
	}

	total := int64(len(published))

	sort.Slice(published, func(i, j int) bool {
		return published[i].LastUpdateTime > published[j].LastUpdateTime
	})

	start := req.GetOffset()
	end := start + req.GetLimit()
	if start > len(published) {
		start = len(published)
	}
	if end > len(published) {
		end = len(published)
	}

	list := published[start:end]

	c.JSON(200, model.PageSuccess(list, total, req.Page, req.PageSize))
}

func (h *Handler) Detail(c *gin.Context) {
	articleID := c.Param("articleId")
	if articleID == "" {
		c.JSON(400, model.Error(1001, "参数错误"))
		return
	}

	meta, err := h.readMeta(articleID)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(404, model.Error(1004, "文章不存在"))
			return
		}
		c.JSON(500, model.Error(1005, "读取文章失败"))
		return
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

	meta, err := h.readMeta(articleID)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(404, model.Error(1004, "文章不存在"))
			return
		}
		c.JSON(500, model.Error(1005, "读取文章失败"))
		return
	}

	if chapterIndex >= len(meta.Chapters) {
		c.JSON(404, model.Error(1004, "章节不存在"))
		return
	}

	chapter := meta.Chapters[chapterIndex]

	content, err := h.readChapter(articleID, chapterIndex)
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

	var prevID, nextID *int
	if chapterIndex > 0 {
		p := chapterIndex - 1
		prevID = &p
	}
	if chapterIndex < len(meta.Chapters)-1 {
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

	content := make([]byte, header.Size)
	_, err = file.Read(content)
	if err != nil {
		c.JSON(500, model.Error(1005, "读取文件失败"))
		return
	}

	text := string(content)
	parsedChapters := h.parseChapters(text, ext)

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

	now := time.Now().Format("2006-01-02 15:04:05")
	totalWordCount := 0

	chapters := make([]model.ChapterMeta, len(parsedChapters))
	for i, ch := range parsedChapters {
		ch.WordCount = len([]rune(ch.Content))
		totalWordCount += ch.WordCount

		chapters[i] = model.ChapterMeta{
			Index:     ch.Index,
			ChapterID: ch.ChapterID,
			Title:     ch.Title,
			WordCount: ch.WordCount,
		}

		chapterPath := h.getChapterPath(articleID, ch.Index)
		if err := os.WriteFile(chapterPath, []byte(ch.Content), 0644); err != nil {
			c.JSON(500, model.Error(1005, "保存章节失败"))
			return
		}
	}

	var tags []string
	if tagsStr != "" {
		tags = strings.Split(tagsStr, ",")
	}

	meta := &model.ArticleMeta{
		ArticleID:      articleID,
		Title:          title,
		Author:         author,
		Summary:        summary,
		Tags:           tags,
		WordCount:      totalWordCount,
		ChapterCount:   len(chapters),
		Status:         "published",
		PublishTime:    now,
		LastUpdateTime: now,
		Chapters:       chapters,
	}

	if err := h.writeMeta(articleID, meta); err != nil {
		c.JSON(500, model.Error(1005, "保存文章元数据失败"))
		return
	}

	indexItem := model.ArticleIndex{
		ArticleID:      articleID,
		Title:          title,
		Author:         author,
		Summary:        summary,
		WordCount:      totalWordCount,
		ChapterCount:   len(chapters),
		Status:         "published",
		LastUpdateTime: now,
	}

	index, _ := h.readIndex()
	index = append([]model.ArticleIndex{indexItem}, index...)
	if err := h.writeIndex(index); err != nil {
		c.JSON(500, model.Error(1005, "更新索引失败"))
		return
	}

	c.JSON(200, model.Success(gin.H{
		"articleId": articleID,
		"title":     title,
		"chapters":  len(chapters),
	}))
}

func (h *Handler) parseChapters(text string, ext string) []model.ChapterMeta {
	var chapters []model.ChapterMeta
	lines := strings.Split(text, "\n")

	var currentTitle string
	var currentContent []string
	var chapterIndex int

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		isChapterTitle := false
		if ext == ".md" {
			isChapterTitle = strings.HasPrefix(line, "# ") || strings.HasPrefix(line, "## ")
			if isChapterTitle {
				line = strings.TrimPrefix(line, "# ")
				line = strings.TrimPrefix(line, "## ")
			}
		} else {
			if strings.HasPrefix(line, "第") && (strings.Contains(line, "章") || strings.Contains(line, "节") || strings.Contains(line, "序")) {
				isChapterTitle = true
			}
		}

		if isChapterTitle {
			if currentTitle != "" && len(currentContent) > 0 {
				chapters = append(chapters, model.ChapterMeta{
					Index:     chapterIndex,
					ChapterID: utils.GenerateChapterID(),
					Title:     currentTitle,
					WordCount: 0,
					Content:   strings.Join(currentContent, "\n\n"),
				})
				chapterIndex++
			}
			currentTitle = line
			currentContent = nil
		} else {
			currentContent = append(currentContent, line)
		}
	}

	if currentTitle != "" && len(currentContent) > 0 {
		chapters = append(chapters, model.ChapterMeta{
			Index:     chapterIndex,
			ChapterID: utils.GenerateChapterID(),
			Title:     currentTitle,
			WordCount: 0,
			Content:   strings.Join(currentContent, "\n\n"),
		})
	}

	return chapters
}

func (h *Handler) Delete(c *gin.Context) {
	articleID := c.Param("articleId")
	if articleID == "" {
		c.JSON(400, model.Error(1001, "参数错误"))
		return
	}

	articleDir := h.getArticleDir(articleID)
	if err := os.RemoveAll(articleDir); err != nil {
		c.JSON(500, model.Error(1005, "删除文章失败"))
		return
	}

	index, _ := h.readIndex()
	var newIndex []model.ArticleIndex
	for _, item := range index {
		if item.ArticleID != articleID {
			newIndex = append(newIndex, item)
		}
	}
	h.writeIndex(newIndex)

	c.JSON(200, model.Success(nil))
}

func (h *Handler) Search(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(400, model.Error(1001, "搜索关键词不能为空"))
		return
	}

	index, err := h.readIndex()
	if err != nil {
		c.JSON(500, model.Error(1005, "读取文章列表失败"))
		return
	}

	var results []model.ArticleIndex
	keyword = strings.ToLower(keyword)
	for _, item := range index {
		if item.Status != "published" && item.Status != "" {
			continue
		}
		if strings.Contains(strings.ToLower(item.Title), keyword) ||
			strings.Contains(strings.ToLower(item.Author), keyword) ||
			strings.Contains(strings.ToLower(item.Summary), keyword) {
			results = append(results, item)
		}
	}

	c.JSON(200, model.Success(results))
}
