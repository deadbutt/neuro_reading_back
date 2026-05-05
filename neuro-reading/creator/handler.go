package creator

import (
	"encoding/json"
	"fmt"
	"neuro-reading/config"
	"neuro-reading/db"
	"neuro-reading/model"
	"neuro-reading/utils"
	"os"
	"path"
	"regexp"
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
	return &Handler{
		cfg:        cfg,
		articleDir: articleDir,
	}
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

	if err := db.DB.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		c.JSON(400, model.Error(1001, "邮箱已被使用"))
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
		"avatar":    creator.Avatar,
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
		"avatar":    creator.Avatar,
		"token":     token,
	}))
}

func (h *Handler) GetProfile(c *gin.Context) {
	creatorID := c.GetString("userId")

	var creator model.Creator
	if err := db.DB.Where("creator_id = ?", creatorID).First(&creator).Error; err != nil {
		c.JSON(404, model.Error(1004, "创作者不存在"))
		return
	}

	var articleCount int64
	db.DB.Model(&model.ArticleMeta{}).Where("creator_id = ?", creatorID).Count(&articleCount)

	c.JSON(200, model.Success(model.CreatorProfileResponse{
		CreatorID:     creator.CreatorID,
		Account:       creator.Account,
		Name:          creator.Name,
		Avatar:        creator.Avatar,
		Description:   creator.Description,
		Email:         creator.Email,
		ArticleCount:  int(articleCount),
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
	if req.Email != "" {
		updates["email"] = req.Email
	}

	if len(updates) > 0 {
		if err := db.DB.Model(&model.Creator{}).Where("creator_id = ?", creatorID).Updates(updates).Error; err != nil {
			c.JSON(500, model.Error(1005, "更新失败"))
			return
		}
	}

	c.JSON(200, model.Success(nil))
}

func (h *Handler) GetArticles(c *gin.Context) {
	creatorID := c.GetString("userId")

	var req model.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req.Page = 1
		req.PageSize = 10
	}

	indexPath := path.Join(h.articleDir, "index.json")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		c.JSON(200, model.PageSuccess([]interface{}{}, 0, req.Page, req.PageSize))
		return
	}

	var allArticles []model.ArticleIndex
	json.Unmarshal(data, &allArticles)

	var creatorArticles []model.ArticleIndex
	for _, article := range allArticles {
		if article.CreatorID == creatorID {
			creatorArticles = append(creatorArticles, article)
		}
	}

	total := int64(len(creatorArticles))
	start := req.GetOffset()
	end := start + req.GetLimit()
	if start > len(creatorArticles) {
		start = len(creatorArticles)
	}
	if end > len(creatorArticles) {
		end = len(creatorArticles)
	}

	list := creatorArticles[start:end]

	c.JSON(200, model.PageSuccess(list, total, req.Page, req.PageSize))
}

func (h *Handler) UploadArticle(c *gin.Context) {
	creatorID := c.GetString("userId")

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(400, model.Error(1001, "请选择要上传的文件"))
		return
	}
	defer file.Close()

	title := c.PostForm("title")
	summary := c.PostForm("summary")
	tagsStr := c.PostForm("tags")

	if title == "" {
		c.JSON(400, model.Error(1001, "标题不能为空"))
		return
	}

	ext := strings.ToLower(path.Ext(header.Filename))
	if ext != ".txt" && ext != ".md" && ext != ".json" {
		c.JSON(400, model.Error(1001, "仅支持 txt、md、json 格式"))
		return
	}

	content := make([]byte, header.Size)
	_, err = file.Read(content)
	if err != nil {
		c.JSON(500, model.Error(1005, "读取文件失败"))
		return
	}

	articleID := utils.GenerateArticleID()
	articlePath := path.Join(h.articleDir, articleID)
	chaptersDir := path.Join(articlePath, "chapters")
	if err := os.MkdirAll(chaptersDir, 0755); err != nil {
		c.JSON(500, model.Error(1005, "创建目录失败"))
		return
	}

	now := time.Now().Format("2006-01-02 15:04:05")

	var chapters []model.ChapterMeta
	var totalWordCount int

	if ext == ".json" {
		var articleData struct {
			Title    string `json:"title"`
			Content  string `json:"content"`
			Chapters []struct {
				Title   string `json:"title"`
				Content string `json:"content"`
			} `json:"chapters"`
		}
		if err := json.Unmarshal(content, &articleData); err != nil {
			c.JSON(400, model.Error(1001, "JSON格式错误"))
			return
		}
		if articleData.Title != "" {
			title = articleData.Title
		}
		if len(articleData.Chapters) > 0 {
			for i, ch := range articleData.Chapters {
				wordCount := len([]rune(ch.Content))
				totalWordCount += wordCount
				chapters = append(chapters, model.ChapterMeta{
					Index:     i,
					ChapterID: utils.GenerateChapterID(),
					Title:     ch.Title,
					WordCount: wordCount,
				})
				os.WriteFile(path.Join(chaptersDir, fmt.Sprintf("%d.txt", i)), []byte(ch.Content), 0644)
			}
		} else if articleData.Content != "" {
			wordCount := len([]rune(articleData.Content))
			totalWordCount = wordCount
			chapters = append(chapters, model.ChapterMeta{
				Index:     0,
				ChapterID: utils.GenerateChapterID(),
				Title:     title,
				WordCount: wordCount,
			})
			os.WriteFile(path.Join(chaptersDir, "0.txt"), []byte(articleData.Content), 0644)
		}
	} else {
		text := string(content)
		parsedChapters := parseChapters(text, title)
		if len(parsedChapters) == 0 {
			wordCount := len([]rune(text))
			totalWordCount = wordCount
			chapters = []model.ChapterMeta{{
				Index:     0,
				ChapterID: utils.GenerateChapterID(),
				Title:     title,
				WordCount: wordCount,
			}}
			os.WriteFile(path.Join(chaptersDir, "0.txt"), content, 0644)
		} else {
			for i, ch := range parsedChapters {
				wordCount := len([]rune(ch.Content))
				totalWordCount += wordCount
				chapters = append(chapters, model.ChapterMeta{
					Index:     i,
					ChapterID: utils.GenerateChapterID(),
					Title:     ch.Title,
					WordCount: wordCount,
				})
				os.WriteFile(path.Join(chaptersDir, fmt.Sprintf("%d.txt", i)), []byte(ch.Content), 0644)
			}
		}
	}

	var tags []string
	if tagsStr != "" {
		tags = strings.Split(tagsStr, ",")
	}

	if summary == "" && len(chapters) > 0 {
		// Read first chapter for summary
		firstChapterPath := path.Join(chaptersDir, "0.txt")
		if data, err := os.ReadFile(firstChapterPath); err == nil {
			summary = string(data)
			if len(summary) > 200 {
				summary = summary[:200] + "..."
			}
		}
	}

	meta := &model.ArticleMeta{
		ArticleID:      articleID,
		Title:          title,
		Author:         "",
		Summary:        summary,
		Tags:           tags,
		WordCount:      totalWordCount,
		ChapterCount:   len(chapters),
		Status:         "published",
		PublishTime:    now,
		LastUpdateTime: now,
		Chapters:       chapters,
	}

	metaData, _ := json.MarshalIndent(meta, "", "  ")
	os.WriteFile(path.Join(articlePath, "meta.json"), metaData, 0644)

	indexPath := path.Join(h.articleDir, "index.json")
	var index []model.ArticleIndex
	if indexData, err := os.ReadFile(indexPath); err == nil {
		json.Unmarshal(indexData, &index)
	}

	indexItem := model.ArticleIndex{
		ArticleID:      articleID,
		Title:          title,
		Author:         "",
		Summary:        summary,
		WordCount:      totalWordCount,
		ChapterCount:   len(chapters),
		Tags:           tags,
		LastUpdateTime: now,
	}
	indexItem.CreatorID = creatorID

	index = append([]model.ArticleIndex{indexItem}, index...)
	indexData, _ := json.MarshalIndent(index, "", "  ")
	os.WriteFile(indexPath, indexData, 0644)

	c.JSON(200, model.Success(gin.H{
		"articleId":    articleID,
		"title":        title,
		"chapterCount": len(chapters),
		"wordCount":    totalWordCount,
	}))
}

func (h *Handler) DeleteArticle(c *gin.Context) {
	creatorID := c.GetString("userId")
	articleID := c.Param("articleId")

	articlePath := path.Join(h.articleDir, articleID)
	metaPath := path.Join(articlePath, "meta.json")

	data, err := os.ReadFile(metaPath)
	if err != nil {
		c.JSON(404, model.Error(1004, "文章不存在"))
		return
	}

	var meta model.ArticleMeta
	json.Unmarshal(data, &meta)

	if meta.CreatorID != creatorID {
		c.JSON(403, model.Error(1003, "无权删除此文章"))
		return
	}

	os.RemoveAll(articlePath)

	indexPath := path.Join(h.articleDir, "index.json")
	var index []model.ArticleIndex
	if indexData, err := os.ReadFile(indexPath); err == nil {
		json.Unmarshal(indexData, &index)
	}

	var newIndex []model.ArticleIndex
	for _, item := range index {
		if item.ArticleID != articleID {
			newIndex = append(newIndex, item)
		}
	}

	indexData, _ := json.MarshalIndent(newIndex, "", "  ")
	os.WriteFile(indexPath, indexData, 0644)

	c.JSON(200, model.Success(nil))
}

func (h *Handler) GetStats(c *gin.Context) {
	creatorID := c.GetString("userId")

	indexPath := path.Join(h.articleDir, "index.json")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		c.JSON(200, model.Success(model.CreatorStatsResponse{}))
		return
	}

	var allArticles []model.ArticleIndex
	json.Unmarshal(data, &allArticles)

	var creatorArticles []model.ArticleIndex
	for _, article := range allArticles {
		if article.CreatorID == creatorID {
			creatorArticles = append(creatorArticles, article)
		}
	}

	stats := model.CreatorStatsResponse{
		TotalArticles: len(creatorArticles),
	}

	for _, article := range creatorArticles {
		stats.TotalWords += article.WordCount
	}

	c.JSON(200, model.Success(stats))
}

type ParsedChapter struct {
	Title   string
	Content string
}

func parseChapters(content string, articleTitle string) []ParsedChapter {
	var chapters []ParsedChapter

	chapterPatterns := []string{
		`(?m)^【第[一二三四五六七八九十百千万零\d]+章[\s:：]*(.+?)】\s*$`,
		`(?m)^第[一二三四五六七八九十百千万零\d]+章[\s:：]*(.+?)$`,
		`(?m)^【第[一二三四五六七八九十百千万零\d]+章】\s*$`,
		`(?m)^第[一二三四五六七八九十百千万零\d]+章\s*$`,
		`(?m)^Chapter\s+\d+[\s:：]*(.+?)$`,
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
