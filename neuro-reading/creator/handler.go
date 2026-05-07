package creator

import (
	"archive/zip"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"neuro-reading/config"
	"neuro-reading/db"
	"neuro-reading/model"
	"neuro-reading/utils"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
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
	os.MkdirAll(articleDir, 0755)
	return &Handler{cfg: cfg, articleDir: articleDir}
}

func (h *Handler) Register(c *gin.Context) {
	var req model.CreatorRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, model.Error(1001, "参数错误"))
		return
	}

	if req.Password != req.ConfirmPassword {
		c.JSON(400, model.Error(1004, "两次密码不一致"))
		return
	}

	var existing model.Creator
	if err := db.DB.Where("account = ?", req.Account).First(&existing).Error; err == nil {
		c.JSON(400, model.Error(2001, "账号已存在"))
		return
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	creator := model.Creator{
		CreatorID:     utils.GenerateCreatorID(),
		Account:       req.Account,
		Password:      req.Password,
		Name:          req.Name,
		Email:         req.Email,
		CreateTime:    now,
		LastLoginTime: now,
		Status:        1,
	}

	if err := db.DB.Create(&creator).Error; err != nil {
		c.JSON(500, model.Error(1005, "注册失败"))
		return
	}

	token, _, _ := utils.GenerateToken(creator.CreatorID, "access", &h.cfg.JWT)

	c.JSON(200, model.Success(gin.H{
		"creatorId": creator.CreatorID,
		"account":   creator.Account,
		"name":      creator.Name,
		"token":     token,
	}))
}

func (h *Handler) Login(c *gin.Context) {
	var req model.CreatorLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, model.Error(1001, "参数错误"))
		return
	}

	var creator model.Creator
	if err := db.DB.Where("account = ?", req.Account).First(&creator).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(400, model.Error(2002, "账号或密码错误"))
			return
		}
		c.JSON(500, model.Error(1005, "服务器内部错误"))
		return
	}

	if creator.Password != req.Password {
		c.JSON(400, model.Error(2002, "账号或密码错误"))
		return
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	db.DB.Model(&creator).Update("last_login_time", now)

	token, _, _ := utils.GenerateToken(creator.CreatorID, "access", &h.cfg.JWT)

	c.JSON(200, model.Success(gin.H{
		"creatorId": creator.CreatorID,
		"account":   creator.Account,
		"name":      creator.Name,
		"token":     token,
	}))
}

func (h *Handler) GetProfile(c *gin.Context) {
	creatorID := c.GetString("userId")

	creator, err := h.getOrCreateCreator(creatorID)
	if err != nil {
		c.JSON(404, model.Error(1004, "用户不存在"))
		return
	}

	articles := h.getCreatorArticles(creatorID)

	c.JSON(200, model.Success(model.CreatorProfileResponse{
		CreatorID:     creator.CreatorID,
		Account:       creator.Account,
		Name:          creator.Name,
		Avatar:        creator.Avatar,
		Description:   creator.Description,
		Email:         creator.Email,
		ArticleCount:  len(articles),
		CreateTime:    creator.CreateTime,
		LastLoginTime: creator.LastLoginTime,
	}))
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	creatorID := c.GetString("userId")

	var req model.CreatorProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, model.Error(1001, "参数错误"))
		return
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}

	if len(updates) > 0 {
		db.DB.Model(&model.Creator{}).Where("creator_id = ?", creatorID).Updates(updates)
	}

	c.JSON(200, model.Success(nil))
}

func (h *Handler) getOrCreateCreator(userID string) (*model.Creator, error) {
	var creator model.Creator
	err := db.DB.Where("creator_id = ?", userID).First(&creator).Error
	if err == nil {
		return &creator, nil
	}

	var user model.User
	if err := db.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	creator = model.Creator{
		CreatorID:     user.UserID,
		Account:       user.Account,
		Password:      user.Password,
		Name:          user.Nickname,
		Avatar:        user.Avatar,
		Description:   user.Bio,
		CreateTime:    now,
		LastLoginTime: now,
		Status:        1,
	}

	if err := db.DB.Create(&creator).Error; err != nil {
		return nil, err
	}

	return &creator, nil
}

func (h *Handler) CreateWork(c *gin.Context) {
	creatorID := c.GetString("userId")

	var req model.CreateWorkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, model.Error(1001, "参数错误"))
		return
	}

	creator, err := h.getOrCreateCreator(creatorID)
	if err != nil {
		c.JSON(404, model.Error(1004, "用户不存在"))
		return
	}

	articleID := utils.GenerateArticleID()
	articlePath := path.Join(h.articleDir, articleID)
	chaptersDir := path.Join(articlePath, "chapters")
	os.MkdirAll(chaptersDir, 0755)

	now := time.Now().Format("2006-01-02 15:04:05")

	meta := &model.ArticleMeta{
		ArticleID:      articleID,
		CreatorID:      creatorID,
		Title:          req.Title,
		Author:         creator.Name,
		Summary:        req.Summary,
		Tags:           req.Tags,
		Cover:          req.Cover,
		Status:         "draft",
		WordCount:      0,
		ChapterCount:   0,
		PublishTime:    now,
		LastUpdateTime: now,
		Chapters:       []model.ChapterMeta{},
	}

	h.writeMeta(articleID, meta)
	h.addToIndex(meta)

	c.JSON(200, model.Success(gin.H{
		"articleId": articleID,
		"creatorId": creatorID,
		"title":     req.Title,
		"status":    "draft",
	}))
}

func (h *Handler) GetMyWorks(c *gin.Context) {
	creatorID := c.GetString("userId")

	var req model.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req.Page = 1
		req.PageSize = 20
	}

	status := c.Query("status")

	allArticles := h.getCreatorArticles(creatorID)

	var filtered []model.ArticleIndex
	for _, a := range allArticles {
		if status == "" || status == "all" || a.Status == status {
			filtered = append(filtered, a)
		}
	}

	total := int64(len(filtered))
	start := (req.Page - 1) * req.PageSize
	end := start + req.PageSize
	if start > len(filtered) {
		start = len(filtered)
	}
	if end > len(filtered) {
		end = len(filtered)
	}

	list := filtered[start:end]

	c.JSON(200, model.PageSuccess(list, total, req.Page, req.PageSize))
}

func (h *Handler) GetWork(c *gin.Context) {
	creatorID := c.GetString("userId")
	articleID := c.Param("workId")

	meta, err := h.readMeta(articleID)
	if err != nil || meta.CreatorID != creatorID {
		c.JSON(404, model.Error(1004, "作品不存在"))
		return
	}

	c.JSON(200, model.Success(meta))
}

func (h *Handler) UpdateWork(c *gin.Context) {
	creatorID := c.GetString("userId")
	articleID := c.Param("workId")

	meta, err := h.readMeta(articleID)
	if err != nil || meta.CreatorID != creatorID {
		c.JSON(404, model.Error(1004, "作品不存在"))
		return
	}

	var req model.UpdateWorkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, model.Error(1001, "参数错误"))
		return
	}

	if req.Title != "" {
		meta.Title = req.Title
	}
	if req.Summary != "" {
		meta.Summary = req.Summary
	}
	if req.Tags != nil {
		meta.Tags = req.Tags
	}
	if req.Cover != "" {
		meta.Cover = req.Cover
	}
	meta.LastUpdateTime = time.Now().Format("2006-01-02 15:04:05")

	h.writeMeta(articleID, meta)
	h.updateIndex(meta)

	c.JSON(200, model.Success(nil))
}

func (h *Handler) DeleteWork(c *gin.Context) {
	creatorID := c.GetString("userId")
	articleID := c.Param("workId")

	meta, err := h.readMeta(articleID)
	if err != nil || meta.CreatorID != creatorID {
		c.JSON(404, model.Error(1004, "作品不存在"))
		return
	}

	os.RemoveAll(path.Join(h.articleDir, articleID))
	h.removeFromIndex(articleID)

	c.JSON(200, model.Success(nil))
}

func (h *Handler) CreateChapter(c *gin.Context) {
	creatorID := c.GetString("userId")
	articleID := c.Param("workId")

	meta, err := h.readMeta(articleID)
	if err != nil || meta.CreatorID != creatorID {
		c.JSON(404, model.Error(1004, "作品不存在"))
		return
	}

	var req model.CreateChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, model.Error(1001, "参数错误"))
		return
	}

	chapterIndex := len(meta.Chapters)
	wordCount := len([]rune(req.Content))

	chapter := model.ChapterMeta{
		Index:     chapterIndex,
		ChapterID: utils.GenerateChapterID(),
		Title:     req.Title,
		WordCount: wordCount,
	}

	chapterPath := path.Join(h.articleDir, articleID, "chapters", fmt.Sprintf("%d.txt", chapterIndex))
	os.WriteFile(chapterPath, []byte(req.Content), 0644)

	meta.Chapters = append(meta.Chapters, chapter)
	meta.ChapterCount = len(meta.Chapters)
	meta.WordCount += wordCount
	meta.LastUpdateTime = time.Now().Format("2006-01-02 15:04:05")

	h.writeMeta(articleID, meta)
	h.updateIndex(meta)

	c.JSON(200, model.Success(gin.H{
		"chapterId":  chapter.ChapterID,
		"index":      chapterIndex,
		"title":      chapter.Title,
		"wordCount":  wordCount,
	}))
}

func (h *Handler) GetChapter(c *gin.Context) {
	creatorID := c.GetString("userId")
	articleID := c.Param("workId")
	chapterIndexStr := c.Param("chapterId")

	meta, err := h.readMeta(articleID)
	if err != nil || meta.CreatorID != creatorID {
		c.JSON(404, model.Error(1004, "作品不存在"))
		return
	}

	chapterIndex := 0
	fmt.Sscanf(chapterIndexStr, "%d", &chapterIndex)

	if chapterIndex >= len(meta.Chapters) {
		c.JSON(404, model.Error(1004, "章节不存在"))
		return
	}

	chapter := meta.Chapters[chapterIndex]
	content, err := h.readChapter(articleID, chapterIndex)
	if err != nil {
		c.JSON(500, model.Error(1005, "读取章节失败"))
		return
	}

	c.JSON(200, model.Success(gin.H{
		"chapterId":  chapter.ChapterID,
		"index":      chapterIndex,
		"title":      chapter.Title,
		"content":    content,
		"wordCount":  chapter.WordCount,
	}))
}

func (h *Handler) UpdateChapter(c *gin.Context) {
	creatorID := c.GetString("userId")
	articleID := c.Param("workId")
	chapterIndexStr := c.Param("chapterId")

	meta, err := h.readMeta(articleID)
	if err != nil || meta.CreatorID != creatorID {
		c.JSON(404, model.Error(1004, "作品不存在"))
		return
	}

	chapterIndex := 0
	fmt.Sscanf(chapterIndexStr, "%d", &chapterIndex)

	if chapterIndex >= len(meta.Chapters) {
		c.JSON(404, model.Error(1004, "章节不存在"))
		return
	}

	var req model.UpdateChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, model.Error(1001, "参数错误"))
		return
	}

	oldWordCount := meta.Chapters[chapterIndex].WordCount

	if req.Title != "" {
		meta.Chapters[chapterIndex].Title = req.Title
	}
	if req.Content != "" {
		chapterPath := path.Join(h.articleDir, articleID, "chapters", fmt.Sprintf("%d.txt", chapterIndex))
		os.WriteFile(chapterPath, []byte(req.Content), 0644)
		newWordCount := len([]rune(req.Content))
		meta.Chapters[chapterIndex].WordCount = newWordCount
		meta.WordCount = meta.WordCount - oldWordCount + newWordCount
	}
	meta.LastUpdateTime = time.Now().Format("2006-01-02 15:04:05")

	h.writeMeta(articleID, meta)
	h.updateIndex(meta)

	c.JSON(200, model.Success(nil))
}

func (h *Handler) DeleteChapter(c *gin.Context) {
	creatorID := c.GetString("userId")
	articleID := c.Param("workId")
	chapterIndexStr := c.Param("chapterId")

	meta, err := h.readMeta(articleID)
	if err != nil || meta.CreatorID != creatorID {
		c.JSON(404, model.Error(1004, "作品不存在"))
		return
	}

	chapterIndex := 0
	fmt.Sscanf(chapterIndexStr, "%d", &chapterIndex)

	if chapterIndex >= len(meta.Chapters) {
		c.JSON(404, model.Error(1004, "章节不存在"))
		return
	}

	removedWordCount := meta.Chapters[chapterIndex].WordCount

	chapterPath := path.Join(h.articleDir, articleID, "chapters", fmt.Sprintf("%d.txt", chapterIndex))
	os.Remove(chapterPath)

	meta.Chapters = append(meta.Chapters[:chapterIndex], meta.Chapters[chapterIndex+1:]...)
	for i := range meta.Chapters {
		meta.Chapters[i].Index = i
		oldPath := path.Join(h.articleDir, articleID, "chapters", fmt.Sprintf("%d.txt", i+1))
		newPath := path.Join(h.articleDir, articleID, "chapters", fmt.Sprintf("%d.txt", i))
		os.Rename(oldPath, newPath)
	}

	meta.ChapterCount = len(meta.Chapters)
	meta.WordCount -= removedWordCount
	meta.LastUpdateTime = time.Now().Format("2006-01-02 15:04:05")

	h.writeMeta(articleID, meta)
	h.updateIndex(meta)

	c.JSON(200, model.Success(nil))
}

func (h *Handler) PublishWork(c *gin.Context) {
	creatorID := c.GetString("userId")
	articleID := c.Param("workId")

	meta, err := h.readMeta(articleID)
	if err != nil || meta.CreatorID != creatorID {
		c.JSON(404, model.Error(1004, "作品不存在"))
		return
	}

	if len(meta.Chapters) == 0 {
		c.JSON(400, model.Error(1001, "发布前需至少有一个章节"))
		return
	}

	meta.Status = "published"
	meta.LastUpdateTime = time.Now().Format("2006-01-02 15:04:05")

	h.writeMeta(articleID, meta)
	h.updateIndex(meta)

	c.JSON(200, model.Success(gin.H{
		"articleId": articleID,
		"creatorId": creatorID,
		"status":    "published",
	}))
}

func (h *Handler) UploadDocx(c *gin.Context) {
	creatorID := c.GetString("userId")

	creator, err := h.getOrCreateCreator(creatorID)
	if err != nil {
		c.JSON(404, model.Error(1004, "用户不存在"))
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(400, model.Error(1001, "请选择要上传的文件"))
		return
	}
	defer file.Close()

	title := c.PostForm("title")
	if title == "" {
		title = strings.TrimSuffix(header.Filename, ".docx")
	}

	content, err := io.ReadAll(file)
	if err != nil {
		c.JSON(500, model.Error(1005, "读取文件失败"))
		return
	}

	text, err := h.parseDocx(content)
	if err != nil {
		c.JSON(400, model.Error(1001, "解析docx文件失败，请确保文件格式正确"))
		return
	}

	parsedChapters := h.parseChapters(text, ".txt")

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
	articlePath := path.Join(h.articleDir, articleID)
	chaptersDir := path.Join(articlePath, "chapters")
	os.MkdirAll(chaptersDir, 0755)

	now := time.Now().Format("2006-01-02 15:04:05")
	totalWordCount := 0

	chapters := make([]model.ChapterMeta, len(parsedChapters))
	for i, ch := range parsedChapters {
		ch.WordCount = len([]rune(ch.Content))
		totalWordCount += ch.WordCount

		chapters[i] = model.ChapterMeta{
			Index:     i,
			ChapterID: utils.GenerateChapterID(),
			Title:     ch.Title,
			WordCount: ch.WordCount,
		}

		chapterPath := path.Join(chaptersDir, fmt.Sprintf("%d.txt", i))
		os.WriteFile(chapterPath, []byte(ch.Content), 0644)
	}

	summary := c.PostForm("summary")
	if summary == "" {
		if len(text) > 200 {
			summary = text[:200] + "..."
		} else {
			summary = text
		}
	}

	tagsStr := c.PostForm("tags")
	var tags []string
	if tagsStr != "" {
		tags = strings.Split(tagsStr, ",")
	}

	cover := c.PostForm("cover")

	meta := &model.ArticleMeta{
		ArticleID:      articleID,
		CreatorID:      creatorID,
		Title:          title,
		Author:         creator.Name,
		Summary:        summary,
		Tags:           tags,
		Cover:          cover,
		Status:         "draft",
		WordCount:      totalWordCount,
		ChapterCount:   len(chapters),
		PublishTime:    now,
		LastUpdateTime: now,
		Chapters:       chapters,
	}

	h.writeMeta(articleID, meta)
	h.addToIndex(meta)

	c.JSON(200, model.Success(gin.H{
		"articleId":    articleID,
		"creatorId":    creatorID,
		"title":        title,
		"chapterCount": len(chapters),
		"wordCount":    totalWordCount,
	}))
}

func (h *Handler) UploadTxt(c *gin.Context) {
	creatorID := c.GetString("userId")

	creator, err := h.getOrCreateCreator(creatorID)
	if err != nil {
		c.JSON(404, model.Error(1004, "用户不存在"))
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(400, model.Error(1001, "请选择要上传的文件"))
		return
	}
	defer file.Close()

	title := c.PostForm("title")
	if title == "" {
		title = strings.TrimSuffix(header.Filename, path.Ext(header.Filename))
	}

	ext := strings.ToLower(path.Ext(header.Filename))
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
	articlePath := path.Join(h.articleDir, articleID)
	chaptersDir := path.Join(articlePath, "chapters")
	os.MkdirAll(chaptersDir, 0755)

	now := time.Now().Format("2006-01-02 15:04:05")
	totalWordCount := 0

	chapters := make([]model.ChapterMeta, len(parsedChapters))
	for i, ch := range parsedChapters {
		ch.WordCount = len([]rune(ch.Content))
		totalWordCount += ch.WordCount

		chapters[i] = model.ChapterMeta{
			Index:     i,
			ChapterID: utils.GenerateChapterID(),
			Title:     ch.Title,
			WordCount: ch.WordCount,
		}

		chapterPath := path.Join(chaptersDir, fmt.Sprintf("%d.txt", i))
		os.WriteFile(chapterPath, []byte(ch.Content), 0644)
	}

	summary := c.PostForm("summary")
	if summary == "" {
		if len(text) > 200 {
			summary = text[:200] + "..."
		} else {
			summary = text
		}
	}

	tagsStr := c.PostForm("tags")
	var tags []string
	if tagsStr != "" {
		tags = strings.Split(tagsStr, ",")
	}

	cover := c.PostForm("cover")

	meta := &model.ArticleMeta{
		ArticleID:      articleID,
		CreatorID:      creatorID,
		Title:          title,
		Author:         creator.Name,
		Summary:        summary,
		Tags:           tags,
		Cover:          cover,
		WordCount:      totalWordCount,
		ChapterCount:   len(chapters),
		Status:         "draft",
		PublishTime:    now,
		LastUpdateTime: now,
		Chapters:       chapters,
	}

	h.writeMeta(articleID, meta)
	h.addToIndex(meta)

	c.JSON(200, model.Success(gin.H{
		"articleId":    articleID,
		"creatorId":    creatorID,
		"title":        title,
		"chapterCount": len(chapters),
		"wordCount":    totalWordCount,
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

func (h *Handler) getIndexPath() string {
	return path.Join(h.articleDir, "index.json")
}

func (h *Handler) readIndex() ([]model.ArticleIndex, error) {
	data, err := os.ReadFile(h.getIndexPath())
	if err != nil {
		if os.IsNotExist(err) {
			return []model.ArticleIndex{}, nil
		}
		return nil, err
	}
	var index []model.ArticleIndex
	json.Unmarshal(data, &index)
	return index, nil
}

func (h *Handler) writeIndex(index []model.ArticleIndex) error {
	data, _ := json.MarshalIndent(index, "", "  ")
	return os.WriteFile(h.getIndexPath(), data, 0644)
}

func (h *Handler) readMeta(articleID string) (*model.ArticleMeta, error) {
	metaPath := path.Join(h.articleDir, articleID, "meta.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, err
	}
	var meta model.ArticleMeta
	json.Unmarshal(data, &meta)
	return &meta, nil
}

func (h *Handler) writeMeta(articleID string, meta *model.ArticleMeta) error {
	metaPath := path.Join(h.articleDir, articleID, "meta.json")
	data, _ := json.MarshalIndent(meta, "", "  ")
	return os.WriteFile(metaPath, data, 0644)
}

func (h *Handler) readChapter(articleID string, index int) (string, error) {
	chapterPath := path.Join(h.articleDir, articleID, "chapters", fmt.Sprintf("%d.txt", index))
	data, err := os.ReadFile(chapterPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (h *Handler) addToIndex(meta *model.ArticleMeta) {
	index, _ := h.readIndex()
	indexItem := model.ArticleIndex{
		ArticleID:      meta.ArticleID,
		CreatorID:      meta.CreatorID,
		Title:          meta.Title,
		Author:         meta.Author,
		Summary:        meta.Summary,
		Cover:          meta.Cover,
		WordCount:      meta.WordCount,
		ChapterCount:   meta.ChapterCount,
		Tags:           meta.Tags,
		Status:         meta.Status,
		LastUpdateTime: meta.LastUpdateTime,
	}
	index = append([]model.ArticleIndex{indexItem}, index...)
	h.writeIndex(index)
}

func (h *Handler) updateIndex(meta *model.ArticleMeta) {
	index, _ := h.readIndex()
	for i, item := range index {
		if item.ArticleID == meta.ArticleID {
			index[i] = model.ArticleIndex{
				ArticleID:      meta.ArticleID,
				CreatorID:      meta.CreatorID,
				Title:          meta.Title,
				Author:         meta.Author,
				Summary:        meta.Summary,
				Cover:          meta.Cover,
				WordCount:      meta.WordCount,
				ChapterCount:   meta.ChapterCount,
				Tags:           meta.Tags,
				Status:         meta.Status,
				LastUpdateTime: meta.LastUpdateTime,
			}
			break
		}
	}
	h.writeIndex(index)
}

func (h *Handler) removeFromIndex(articleID string) {
	index, _ := h.readIndex()
	var newIndex []model.ArticleIndex
	for _, item := range index {
		if item.ArticleID != articleID {
			newIndex = append(newIndex, item)
		}
	}
	h.writeIndex(newIndex)
}

func (h *Handler) getCreatorArticles(creatorID string) []model.ArticleIndex {
	index, _ := h.readIndex()
	var result []model.ArticleIndex
	for _, item := range index {
		if item.CreatorID == creatorID {
			result = append(result, item)
		}
	}
	return result
}

func parseChapters(content string, articleTitle string) []ParsedChapter {
	var chapters []ParsedChapter

	chapterPatterns := []string{
		`(?m)^【第[一二三四五六七八九十百千万零\d]+章[\s:：]*(.+?)】\s*$`,
		`(?m)^第[一二三四五六七八九十百千万零\d]+章[\s:：]*(.+?)$`,
		`(?m)^【第[一二三四五六七八九十百千万零\d]+章】\s*$`,
		`(?m)^第[一二三四五六七八九十百千万零\d]+章\s*$`,
	}

	var allMatches [][]int
	var matchTitles []string

	for _, pattern := range chapterPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatchIndex(content, -1)
		for _, match := range matches {
			start := match[0]
			end := match[1]
			var title string
			if len(match) >= 4 && match[2] >= 0 {
				title = strings.TrimSpace(content[match[2]:match[3]])
			}
			if title == "" {
				title = strings.TrimSpace(content[start:end])
				title = strings.Trim(title, "【】")
			}
			if title == "" {
				continue
			}
			allMatches = append(allMatches, match[:2])
			matchTitles = append(matchTitles, title)
		}
	}

	if len(allMatches) == 0 {
		bracketPattern := regexp.MustCompile(`(?m)^【(.+?)】\s*$`)
		matches := bracketPattern.FindAllStringSubmatchIndex(content, -1)
		for i, match := range matches {
			title := strings.TrimSpace(content[match[2]:match[3]])
			if title == "" {
				continue
			}
			if title == articleTitle && i > 0 {
				continue
			}
			allMatches = append(allMatches, match[:2])
			matchTitles = append(matchTitles, title)
		}
	}

	if len(allMatches) == 0 {
		return nil
	}

	for i := 0; i < len(allMatches)-1; i++ {
		for j := i + 1; j < len(allMatches); j++ {
			if allMatches[i][0] > allMatches[j][0] {
				allMatches[i], allMatches[j] = allMatches[j], allMatches[i]
				matchTitles[i], matchTitles[j] = matchTitles[j], matchTitles[i]
			}
		}
	}

	for i := 0; i < len(allMatches); i++ {
		end := allMatches[i][1]
		title := matchTitles[i]
		if title == "" {
			title = articleTitle
		}
		var chapterContent string
		if i < len(allMatches)-1 {
			nextStart := allMatches[i+1][0]
			chapterContent = strings.TrimSpace(content[end:nextStart])
		} else {
			chapterContent = strings.TrimSpace(content[end:])
		}
		if chapterContent != "" {
			chapters = append(chapters, ParsedChapter{
				Title:   title,
				Content: chapterContent,
			})
		}
	}

	return chapters
}

type ParsedChapter struct {
	Title   string
	Content string
}

type docxDocument struct {
	XMLName xml.Name `xml:"document"`
	Body    struct {
		Paragraphs []struct {
			Runs []struct {
				Text string `xml:"t"`
			} `xml:"r"`
		} `xml:"p"`
	} `xml:"body"`
}

func (h *Handler) parseDocx(data []byte) (string, error) {
	reader, err := zip.NewReader(strings.NewReader(string(data)), int64(len(data)))
	if err != nil {
		return "", err
	}

	var documentFile *zip.File
	for _, f := range reader.File {
		if f.Name == "word/document.xml" {
			documentFile = f
			break
		}
	}

	if documentFile == nil {
		return "", fmt.Errorf("invalid docx file: document.xml not found")
	}

	rc, err := documentFile.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	content, err := io.ReadAll(rc)
	if err != nil {
		return "", err
	}

	var doc docxDocument
	if err := xml.Unmarshal(content, &doc); err != nil {
		return "", err
	}

	var paragraphs []string
	for _, p := range doc.Body.Paragraphs {
		var text string
		for _, r := range p.Runs {
			text += r.Text
		}
		if text != "" {
			paragraphs = append(paragraphs, text)
		}
	}

	return strings.Join(paragraphs, "\n\n"), nil
}

func (h *Handler) UploadCover(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(400, model.Error(1001, "请选择要上传的文件"))
		return
	}
	defer file.Close()

	ext := strings.ToLower(path.Ext(header.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		c.JSON(400, model.Error(1001, "仅支持 jpg、png、webp 格式的图片"))
		return
	}

	if header.Size > 5*1024*1024 {
		c.JSON(400, model.Error(1001, "图片大小不能超过 5MB"))
		return
	}

	coverDir := path.Join(h.cfg.UploadDir, "covers")
	if err := os.MkdirAll(coverDir, 0755); err != nil {
		c.JSON(500, model.Error(1005, "创建目录失败"))
		return
	}

	filename := fmt.Sprintf("%d_%s.jpg", time.Now().Unix(), strconv.Itoa(int(time.Now().UnixNano()%10000)))
	filePath := path.Join(coverDir, filename)

	srcImage, _, err := image.Decode(file)
	if err != nil {
		c.JSON(400, model.Error(1001, "图片解析失败"))
		return
	}

	resizedImage := imaging.Fill(srcImage, 400, 533, imaging.Center, imaging.Lanczos)

	out, err := os.Create(filePath)
	if err != nil {
		c.JSON(500, model.Error(1005, "文件保存失败"))
		return
	}
	defer out.Close()

	if err := jpeg.Encode(out, resizedImage, &jpeg.Options{Quality: 85}); err != nil {
		c.JSON(500, model.Error(1005, "图片压缩保存失败"))
		return
	}

	baseURL := h.cfg.Server.PublicURL
	if baseURL == "" {
		baseURL = fmt.Sprintf("http://%s:%s", h.cfg.Server.Host, h.cfg.Server.Port)
	}
	fileURL := fmt.Sprintf("%s/uploads/covers/%s", baseURL, filename)

	c.JSON(200, model.Success(gin.H{
		"url": fileURL,
	}))
}
