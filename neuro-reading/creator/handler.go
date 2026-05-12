package creator

import (
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"neuro-reading/config"
	"neuro-reading/db"
	"neuro-reading/model"
	"neuro-reading/utils"
	"os"
	"path"
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
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(500, model.Error(1005, "注册失败"))
		return
	}

	creator := model.Creator{
		CreatorID:     utils.GenerateCreatorID(),
		Account:       req.Account,
		Password:      hashedPassword,
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

	if !utils.CheckPassword(req.Password, creator.Password) {
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

	var articleCount int64
	db.DB.Model(&model.Article{}).Where("creator_id = ?", creatorID).Count(&articleCount)

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
		// Note: user.Password is already hashed
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

	tags := strings.Join(req.Tags, ",")

	article := model.Article{
		ArticleID:  articleID,
		CreatorID:  creatorID,
		Title:      req.Title,
		Author:     creator.Name,
		Summary:    req.Summary,
		Tags:       tags,
		Cover:      req.Cover,
		Status:     "draft",
		WordCount:  0,
		PublishTime: time.Now(),
	}

	if err := db.DB.Create(&article).Error; err != nil {
		c.JSON(500, model.Error(1005, "创建作品失败"))
		return
	}

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

	var total int64
	query := db.DB.Model(&model.Article{}).Where("creator_id = ?", creatorID)
	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)

	var articles []model.Article
	offset := (req.Page - 1) * req.PageSize
	query.Order("updated_at DESC").Offset(offset).Limit(req.PageSize).Find(&articles)

	list := make([]gin.H, len(articles))
	for i, article := range articles {
		list[i] = gin.H{
			"articleId":      article.ArticleID,
			"creatorId":      article.CreatorID,
			"title":          article.Title,
			"summary":        article.Summary,
			"cover":          article.Cover,
			"status":         article.Status,
			"chapterCount":   article.ChapterCount,
			"wordCount":      article.WordCount,
			"lastUpdateTime": article.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	c.JSON(200, model.PageSuccess(list, total, req.Page, req.PageSize))
}

func (h *Handler) GetWork(c *gin.Context) {
	creatorID := c.GetString("userId")
	articleID := c.Param("workId")

	var article model.Article
	if err := db.DB.Where("article_id = ? AND creator_id = ?", articleID, creatorID).First(&article).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, model.Error(1004, "作品不存在"))
			return
		}
		c.JSON(500, model.Error(1005, "查询失败"))
		return
	}

	var chapters []model.Chapter
	db.DB.Where("book_id = ?", articleID).Order("`index` ASC").Find(&chapters)

	chapterList := make([]gin.H, len(chapters))
	for i, ch := range chapters {
		chapterList[i] = gin.H{
			"index":     ch.Index,
			"chapterId": ch.ChapterID,
			"title":     ch.Title,
			"wordCount": ch.WordCount,
		}
	}

	c.JSON(200, model.Success(gin.H{
		"articleId":      article.ArticleID,
		"creatorId":      article.CreatorID,
		"title":          article.Title,
		"summary":        article.Summary,
		"tags":           strings.Split(article.Tags, ","),
		"cover":          article.Cover,
		"status":         article.Status,
		"chapters":       chapterList,
		"wordCount":      article.WordCount,
		"publishTime":    article.PublishTime.Format("2006-01-02 15:04:05"),
		"lastUpdateTime": article.UpdatedAt.Format("2006-01-02 15:04:05"),
	}))
}

func (h *Handler) UpdateWork(c *gin.Context) {
	creatorID := c.GetString("userId")
	articleID := c.Param("workId")

	var article model.Article
	if err := db.DB.Where("article_id = ? AND creator_id = ?", articleID, creatorID).First(&article).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, model.Error(1004, "作品不存在"))
			return
		}
		c.JSON(500, model.Error(1005, "查询失败"))
		return
	}

	var req model.UpdateWorkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, model.Error(1001, "参数错误"))
		return
	}

	updates := make(map[string]interface{})
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Summary != "" {
		updates["summary"] = req.Summary
	}
	if req.Tags != nil {
		updates["tags"] = strings.Join(req.Tags, ",")
	}
	if req.Cover != "" {
		updates["cover"] = req.Cover
	}

	if len(updates) > 0 {
		db.DB.Model(&article).Updates(updates)
	}

	c.JSON(200, model.Success(nil))
}

func (h *Handler) DeleteWork(c *gin.Context) {
	creatorID := c.GetString("userId")
	articleID := c.Param("workId")

	var article model.Article
	if err := db.DB.Where("article_id = ? AND creator_id = ?", articleID, creatorID).First(&article).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, model.Error(1004, "作品不存在"))
			return
		}
		c.JSON(500, model.Error(1005, "查询失败"))
		return
	}

	db.DB.Where("book_id = ?", articleID).Delete(&model.Chapter{})
	db.DB.Delete(&article)

	os.RemoveAll(path.Join(h.articleDir, articleID))

	c.JSON(200, model.Success(nil))
}

func (h *Handler) CreateChapter(c *gin.Context) {
	creatorID := c.GetString("userId")
	articleID := c.Param("workId")

	var article model.Article
	if err := db.DB.Where("article_id = ? AND creator_id = ?", articleID, creatorID).First(&article).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, model.Error(1004, "作品不存在"))
			return
		}
		c.JSON(500, model.Error(1005, "查询失败"))
		return
	}

	var req model.CreateChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, model.Error(1001, "参数错误"))
		return
	}

	var maxIndex int
	db.DB.Model(&model.Chapter{}).Where("book_id = ?", articleID).Select("COALESCE(MAX(`index`), -1)").Scan(&maxIndex)
	chapterIndex := maxIndex + 1

	wordCount := len([]rune(req.Content))

	chapter := model.Chapter{
		ChapterID: utils.GenerateChapterID(),
		BookID:    articleID,
		Title:     req.Title,
		Index:     chapterIndex,
		WordCount: wordCount,
	}

	if err := db.DB.Create(&chapter).Error; err != nil {
		c.JSON(500, model.Error(1005, "创建章节失败"))
		return
	}

	chapterPath := path.Join(h.articleDir, articleID, "chapters", fmt.Sprintf("%d.txt", chapterIndex))
	os.WriteFile(chapterPath, []byte(req.Content), 0644)

	db.DB.Model(&article).Updates(map[string]interface{}{
		"chapter_count": gorm.Expr("chapter_count + 1"),
		"word_count":    gorm.Expr("word_count + ?", wordCount),
	})

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

	var article model.Article
	if err := db.DB.Where("article_id = ? AND creator_id = ?", articleID, creatorID).First(&article).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, model.Error(1004, "作品不存在"))
			return
		}
		c.JSON(500, model.Error(1005, "查询失败"))
		return
	}

	chapterIndex := 0
	fmt.Sscanf(chapterIndexStr, "%d", &chapterIndex)

	var chapter model.Chapter
	if err := db.DB.Where("book_id = ? AND `index` = ?", articleID, chapterIndex).First(&chapter).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, model.Error(1004, "章节不存在"))
			return
		}
		c.JSON(500, model.Error(1005, "查询失败"))
		return
	}

	chapterPath := path.Join(h.articleDir, articleID, "chapters", fmt.Sprintf("%d.txt", chapterIndex))
	content, _ := os.ReadFile(chapterPath)

	c.JSON(200, model.Success(gin.H{
		"chapterId":  chapter.ChapterID,
		"index":      chapter.Index,
		"title":      chapter.Title,
		"content":    string(content),
		"wordCount":  chapter.WordCount,
	}))
}

func (h *Handler) UpdateChapter(c *gin.Context) {
	creatorID := c.GetString("userId")
	articleID := c.Param("workId")
	chapterIndexStr := c.Param("chapterId")

	var article model.Article
	if err := db.DB.Where("article_id = ? AND creator_id = ?", articleID, creatorID).First(&article).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, model.Error(1004, "作品不存在"))
			return
		}
		c.JSON(500, model.Error(1005, "查询失败"))
		return
	}

	chapterIndex := 0
	fmt.Sscanf(chapterIndexStr, "%d", &chapterIndex)

	var chapter model.Chapter
	if err := db.DB.Where("book_id = ? AND `index` = ?", articleID, chapterIndex).First(&chapter).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, model.Error(1004, "章节不存在"))
			return
		}
		c.JSON(500, model.Error(1005, "查询失败"))
		return
	}

	var req model.UpdateChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, model.Error(1001, "参数错误"))
		return
	}

	oldWordCount := chapter.WordCount

	updates := make(map[string]interface{})
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Content != "" {
		chapterPath := path.Join(h.articleDir, articleID, "chapters", fmt.Sprintf("%d.txt", chapterIndex))
		os.WriteFile(chapterPath, []byte(req.Content), 0644)
		newWordCount := len([]rune(req.Content))
		updates["word_count"] = newWordCount
		db.DB.Model(&article).Update("word_count", gorm.Expr("word_count - ? + ?", oldWordCount, newWordCount))
	}

	if len(updates) > 0 {
		db.DB.Model(&chapter).Updates(updates)
	}

	c.JSON(200, model.Success(nil))
}

func (h *Handler) DeleteChapter(c *gin.Context) {
	creatorID := c.GetString("userId")
	articleID := c.Param("workId")
	chapterIndexStr := c.Param("chapterId")

	var article model.Article
	if err := db.DB.Where("article_id = ? AND creator_id = ?", articleID, creatorID).First(&article).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, model.Error(1004, "作品不存在"))
			return
		}
		c.JSON(500, model.Error(1005, "查询失败"))
		return
	}

	chapterIndex := 0
	fmt.Sscanf(chapterIndexStr, "%d", &chapterIndex)

	var chapter model.Chapter
	if err := db.DB.Where("book_id = ? AND `index` = ?", articleID, chapterIndex).First(&chapter).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, model.Error(1004, "章节不存在"))
			return
		}
		c.JSON(500, model.Error(1005, "查询失败"))
		return
	}

	removedWordCount := chapter.WordCount

	chapterPath := path.Join(h.articleDir, articleID, "chapters", fmt.Sprintf("%d.txt", chapterIndex))
	os.Remove(chapterPath)

	db.DB.Delete(&chapter)

	var laterChapters []model.Chapter
	db.DB.Where("book_id = ? AND `index` > ?", articleID, chapterIndex).Order("`index` ASC").Find(&laterChapters)
	for _, ch := range laterChapters {
		oldPath := path.Join(h.articleDir, articleID, "chapters", fmt.Sprintf("%d.txt", ch.Index))
		newPath := path.Join(h.articleDir, articleID, "chapters", fmt.Sprintf("%d.txt", ch.Index-1))
		os.Rename(oldPath, newPath)
		db.DB.Model(&ch).Update("index", ch.Index-1)
	}

	db.DB.Model(&article).Updates(map[string]interface{}{
		"chapter_count": gorm.Expr("chapter_count - 1"),
		"word_count":    gorm.Expr("word_count - ?", removedWordCount),
	})

	c.JSON(200, model.Success(nil))
}

func (h *Handler) PublishWork(c *gin.Context) {
	creatorID := c.GetString("userId")
	articleID := c.Param("workId")

	var article model.Article
	if err := db.DB.Where("article_id = ? AND creator_id = ?", articleID, creatorID).First(&article).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, model.Error(1004, "作品不存在"))
			return
		}
		c.JSON(500, model.Error(1005, "查询失败"))
		return
	}

	var chapterCount int64
	db.DB.Model(&model.Chapter{}).Where("book_id = ?", articleID).Count(&chapterCount)

	if chapterCount == 0 {
		c.JSON(400, model.Error(1001, "发布前需至少有一个章节"))
		return
	}

	db.DB.Model(&article).Update("status", "published")

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
	if len([]rune(title)) > 100 {
		title = string([]rune(title)[:100])
	}

	content, err := io.ReadAll(file)
	if err != nil {
		c.JSON(500, model.Error(1005, "读取文件失败"))
		return
	}

	text, err := utils.ParseDocx(content)
	if err != nil {
		c.JSON(400, model.Error(1001, "解析docx文件失败，请确保文件格式正确"))
		return
	}

	parsedChapters := utils.ParseChapters(text, ".txt")

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

	totalWordCount := 0

	summary := c.PostForm("summary")
	if summary == "" {
		runes := []rune(text)
		if len(runes) > 200 {
			summary = string(runes[:200]) + "..."
		} else {
			summary = text
		}
	}

	tagsStr := c.PostForm("tags")
	cover := c.PostForm("cover")

	author := creator.Name
	if len([]rune(author)) > 50 {
		author = string([]rune(author)[:50])
	}

	article := model.Article{
		ArticleID:  articleID,
		CreatorID:  creatorID,
		Title:      title,
		Author:     author,
		Summary:    summary,
		Tags:       tagsStr,
		Cover:      cover,
		Status:     "draft",
		WordCount:  0,
		PublishTime: time.Now(),
	}

	if err := db.DB.Create(&article).Error; err != nil {
		c.JSON(500, model.Error(1005, "创建作品失败"))
		return
	}

	for i, ch := range parsedChapters {
		ch.WordCount = len([]rune(ch.Content))
		totalWordCount += ch.WordCount

		chapter := model.Chapter{
			ChapterID: utils.GenerateChapterID(),
			BookID:    articleID,
			Title:     ch.Title,
			Index:     i,
			WordCount: ch.WordCount,
		}

		if err := db.DB.Create(&chapter).Error; err != nil {
			c.JSON(500, model.Error(1005, "创建章节失败"))
			return
		}

		chapterPath := path.Join(chaptersDir, fmt.Sprintf("%d.txt", i))
		os.WriteFile(chapterPath, []byte(ch.Content), 0644)
	}

	db.DB.Model(&article).Updates(map[string]interface{}{
		"word_count":    totalWordCount,
		"chapter_count": len(parsedChapters),
	})

	c.JSON(200, model.Success(gin.H{
		"articleId":    articleID,
		"creatorId":    creatorID,
		"title":        title,
		"chapterCount": len(parsedChapters),
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
	if len([]rune(title)) > 100 {
		title = string([]rune(title)[:100])
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
	articlePath := path.Join(h.articleDir, articleID)
	chaptersDir := path.Join(articlePath, "chapters")
	os.MkdirAll(chaptersDir, 0755)

	totalWordCount := 0

	summary := c.PostForm("summary")
	if summary == "" {
		runes := []rune(text)
		if len(runes) > 200 {
			summary = string(runes[:200]) + "..."
		} else {
			summary = text
		}
	}

	tagsStr := c.PostForm("tags")
	cover := c.PostForm("cover")

	author := creator.Name
	if len([]rune(author)) > 50 {
		author = string([]rune(author)[:50])
	}

	article := model.Article{
		ArticleID:  articleID,
		CreatorID:  creatorID,
		Title:      title,
		Author:     author,
		Summary:    summary,
		Tags:       tagsStr,
		Cover:      cover,
		Status:     "draft",
		WordCount:  0,
		PublishTime: time.Now(),
	}

	if err := db.DB.Create(&article).Error; err != nil {
		c.JSON(500, model.Error(1005, "创建作品失败"))
		return
	}

	for i, ch := range parsedChapters {
		ch.WordCount = len([]rune(ch.Content))
		totalWordCount += ch.WordCount

		chapter := model.Chapter{
			ChapterID: utils.GenerateChapterID(),
			BookID:    articleID,
			Title:     ch.Title,
			Index:     i,
			WordCount: ch.WordCount,
		}

		if err := db.DB.Create(&chapter).Error; err != nil {
			c.JSON(500, model.Error(1005, "创建章节失败"))
			return
		}

		chapterPath := path.Join(chaptersDir, fmt.Sprintf("%d.txt", i))
		os.WriteFile(chapterPath, []byte(ch.Content), 0644)
	}

	db.DB.Model(&article).Updates(map[string]interface{}{
		"word_count":    totalWordCount,
		"chapter_count": len(parsedChapters),
	})

	c.JSON(200, model.Success(gin.H{
		"articleId":    articleID,
		"creatorId":    creatorID,
		"title":        title,
		"chapterCount": len(parsedChapters),
		"wordCount":    totalWordCount,
	}))
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

	bounds := srcImage.Bounds()
	if bounds.Dx() > 8192 || bounds.Dy() > 8192 {
		c.JSON(400, model.Error(1001, "图片尺寸过大，最大支持 8192x8192"))
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
