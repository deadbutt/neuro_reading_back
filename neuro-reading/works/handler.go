package works

import (
	"neuro-reading/config"
	"neuro-reading/db"
	"neuro-reading/model"
	"neuro-reading/utils"
	"os"
	"path"
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
	os.MkdirAll(articleDir, 0755)
	return &Handler{cfg: cfg, articleDir: articleDir}
}

func (h *Handler) CreateWork(c *gin.Context) {
	creatorID := c.GetString("userId")

	var req model.CreateWorkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, model.Error(1001, "参数错误"))
		return
	}

	now := time.Now()
	work := model.Work{
		WorkID:     utils.GenerateWorkID(),
		CreatorID:  creatorID,
		Title:      req.Title,
		Summary:    req.Summary,
		Cover:      req.Cover,
		Status:     "draft",
		CreateTime: now,
		UpdateTime: now,
	}

	if len(req.Tags) > 0 {
		work.Tags = strings.Join(req.Tags, ",")
	}

	if err := db.DB.Create(&work).Error; err != nil {
		c.JSON(500, model.Error(1005, "创建失败"))
		return
	}

	c.JSON(200, model.Success(gin.H{
		"workId": work.WorkID,
		"title":  work.Title,
		"status": work.Status,
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
	if status == "" {
		status = "all"
	}

	query := db.DB.Model(&model.Work{}).Where("creator_id = ?", creatorID)
	if status != "all" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var works []model.Work
	offset := (req.Page - 1) * req.PageSize
	query.Order("update_time DESC").Offset(offset).Limit(req.PageSize).Find(&works)

	list := make([]model.WorkResponse, len(works))
	for i, w := range works {
		list[i] = model.WorkResponse{
			WorkID:       w.WorkID,
			Title:        w.Title,
			Summary:      w.Summary,
			Tags:         parseTags(w.Tags),
			Cover:        w.Cover,
			Status:       w.Status,
			ChapterCount: w.ChapterCount,
			WordCount:    w.WordCount,
			CreateTime:   w.CreateTime.Format("2006-01-02 15:04:05"),
			UpdateTime:   w.UpdateTime.Format("2006-01-02 15:04:05"),
		}
	}

	c.JSON(200, model.PageSuccess(list, total, req.Page, req.PageSize))
}

func (h *Handler) GetWork(c *gin.Context) {
	creatorID := c.GetString("userId")
	workID := c.Param("workId")

	var work model.Work
	if err := db.DB.Where("work_id = ? AND creator_id = ?", workID, creatorID).First(&work).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, model.Error(1004, "作品不存在"))
			return
		}
		c.JSON(500, model.Error(1005, "服务器错误"))
		return
	}

	var chapters []model.WorkChapter
	db.DB.Where("work_id = ?", workID).Order("sort_order ASC").Find(&chapters)

	chapterList := make([]model.WorkChapterResponse, len(chapters))
	for i, ch := range chapters {
		chapterList[i] = model.WorkChapterResponse{
			ChapterID:  ch.ChapterID,
			Title:      ch.Title,
			WordCount:  ch.WordCount,
			Status:     ch.Status,
			SortOrder:  ch.SortOrder,
			CreateTime: ch.CreateTime.Format("2006-01-02 15:04:05"),
			UpdateTime: ch.UpdateTime.Format("2006-01-02 15:04:05"),
		}
	}

	c.JSON(200, model.Success(model.WorkDetailResponse{
		WorkID:     work.WorkID,
		Title:      work.Title,
		Summary:    work.Summary,
		Tags:       parseTags(work.Tags),
		Cover:      work.Cover,
		Status:     work.Status,
		WordCount:  work.WordCount,
		Chapters:   chapterList,
		CreateTime: work.CreateTime.Format("2006-01-02 15:04:05"),
		UpdateTime: work.UpdateTime.Format("2006-01-02 15:04:05"),
	}))
}

func (h *Handler) UpdateWork(c *gin.Context) {
	creatorID := c.GetString("userId")
	workID := c.Param("workId")

	var work model.Work
	if err := db.DB.Where("work_id = ? AND creator_id = ?", workID, creatorID).First(&work).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, model.Error(1004, "作品不存在"))
			return
		}
		c.JSON(500, model.Error(1005, "服务器错误"))
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
	if req.Cover != "" {
		updates["cover"] = req.Cover
	}
	if req.Tags != nil {
		updates["tags"] = strings.Join(req.Tags, ",")
	}
	updates["update_time"] = time.Now()

	if err := db.DB.Model(&work).Updates(updates).Error; err != nil {
		c.JSON(500, model.Error(1005, "更新失败"))
		return
	}

	c.JSON(200, model.Success(nil))
}

func (h *Handler) DeleteWork(c *gin.Context) {
	creatorID := c.GetString("userId")
	workID := c.Param("workId")

	var work model.Work
	if err := db.DB.Where("work_id = ? AND creator_id = ?", workID, creatorID).First(&work).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, model.Error(1004, "作品不存在"))
			return
		}
		c.JSON(500, model.Error(1005, "服务器错误"))
		return
	}

	db.DB.Where("work_id = ?", workID).Delete(&model.WorkChapter{})
	db.DB.Delete(&work)

	os.RemoveAll(path.Join(h.articleDir, workID))

	c.JSON(200, model.Success(nil))
}

func (h *Handler) CreateChapter(c *gin.Context) {
	creatorID := c.GetString("userId")
	workID := c.Param("workId")

	var work model.Work
	if err := db.DB.Where("work_id = ? AND creator_id = ?", workID, creatorID).First(&work).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, model.Error(1004, "作品不存在"))
			return
		}
		c.JSON(500, model.Error(1005, "服务器错误"))
		return
	}

	var req model.CreateChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, model.Error(1001, "参数错误"))
		return
	}

	status := req.Status
	if status == "" {
		status = "draft"
	}

	var maxOrder int
	db.DB.Model(&model.WorkChapter{}).Where("work_id = ?", workID).Select("COALESCE(MAX(sort_order), 0)").Scan(&maxOrder)

	now := time.Now()
	chapter := model.WorkChapter{
		ChapterID:  utils.GenerateChapterID(),
		WorkID:     workID,
		Title:      req.Title,
		Content:    req.Content,
		WordCount:  len([]rune(req.Content)),
		Status:     status,
		SortOrder:  maxOrder + 1,
		CreateTime: now,
		UpdateTime: now,
	}

	if err := db.DB.Create(&chapter).Error; err != nil {
		c.JSON(500, model.Error(1005, "创建失败"))
		return
	}

	db.DB.Model(&work).Updates(map[string]interface{}{
		"chapter_count": gorm.Expr("chapter_count + 1"),
		"word_count":    gorm.Expr("word_count + ?", chapter.WordCount),
		"update_time":   now,
	})

	c.JSON(200, model.Success(gin.H{
		"chapterId": chapter.ChapterID,
		"title":     chapter.Title,
		"wordCount": chapter.WordCount,
		"status":    chapter.Status,
	}))
}

func (h *Handler) GetChapter(c *gin.Context) {
	creatorID := c.GetString("userId")
	workID := c.Param("workId")
	chapterID := c.Param("chapterId")

	var work model.Work
	if err := db.DB.Where("work_id = ? AND creator_id = ?", workID, creatorID).First(&work).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, model.Error(1004, "作品不存在"))
			return
		}
		c.JSON(500, model.Error(1005, "服务器错误"))
		return
	}

	var chapter model.WorkChapter
	if err := db.DB.Where("chapter_id = ? AND work_id = ?", chapterID, workID).First(&chapter).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, model.Error(1004, "章节不存在"))
			return
		}
		c.JSON(500, model.Error(1005, "服务器错误"))
		return
	}

	c.JSON(200, model.Success(model.WorkChapterContentResponse{
		ChapterID:  chapter.ChapterID,
		WorkID:     chapter.WorkID,
		Title:      chapter.Title,
		Content:    chapter.Content,
		WordCount:  chapter.WordCount,
		Status:     chapter.Status,
		CreateTime: chapter.CreateTime.Format("2006-01-02 15:04:05"),
		UpdateTime: chapter.UpdateTime.Format("2006-01-02 15:04:05"),
	}))
}

func (h *Handler) UpdateChapter(c *gin.Context) {
	creatorID := c.GetString("userId")
	workID := c.Param("workId")
	chapterID := c.Param("chapterId")

	var work model.Work
	if err := db.DB.Where("work_id = ? AND creator_id = ?", workID, creatorID).First(&work).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, model.Error(1004, "作品不存在"))
			return
		}
		c.JSON(500, model.Error(1005, "服务器错误"))
		return
	}

	var chapter model.WorkChapter
	if err := db.DB.Where("chapter_id = ? AND work_id = ?", chapterID, workID).First(&chapter).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, model.Error(1004, "章节不存在"))
			return
		}
		c.JSON(500, model.Error(1005, "服务器错误"))
		return
	}

	var req model.UpdateChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, model.Error(1001, "参数错误"))
		return
	}

	updates := make(map[string]interface{})
	oldWordCount := chapter.WordCount

	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Content != "" {
		updates["content"] = req.Content
		newWordCount := len([]rune(req.Content))
		updates["word_count"] = newWordCount
		oldWordCount = newWordCount - chapter.WordCount
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	updates["update_time"] = time.Now()

	if err := db.DB.Model(&chapter).Updates(updates).Error; err != nil {
		c.JSON(500, model.Error(1005, "更新失败"))
		return
	}

	if oldWordCount != 0 {
		db.DB.Model(&work).Updates(map[string]interface{}{
			"word_count":  gorm.Expr("word_count + ?", oldWordCount),
			"update_time": time.Now(),
		})
	}

	c.JSON(200, model.Success(nil))
}

func (h *Handler) DeleteChapter(c *gin.Context) {
	creatorID := c.GetString("userId")
	workID := c.Param("workId")
	chapterID := c.Param("chapterId")

	var work model.Work
	if err := db.DB.Where("work_id = ? AND creator_id = ?", workID, creatorID).First(&work).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, model.Error(1004, "作品不存在"))
			return
		}
		c.JSON(500, model.Error(1005, "服务器错误"))
		return
	}

	var chapter model.WorkChapter
	if err := db.DB.Where("chapter_id = ? AND work_id = ?", chapterID, workID).First(&chapter).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, model.Error(1004, "章节不存在"))
			return
		}
		c.JSON(500, model.Error(1005, "服务器错误"))
		return
	}

	db.DB.Delete(&chapter)

	db.DB.Model(&work).Updates(map[string]interface{}{
		"chapter_count": gorm.Expr("chapter_count - 1"),
		"word_count":    gorm.Expr("word_count - ?", chapter.WordCount),
		"update_time":   time.Now(),
	})

	c.JSON(200, model.Success(nil))
}

func (h *Handler) PublishWork(c *gin.Context) {
	creatorID := c.GetString("userId")
	workID := c.Param("workId")

	var work model.Work
	if err := db.DB.Where("work_id = ? AND creator_id = ?", workID, creatorID).First(&work).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, model.Error(1004, "作品不存在"))
			return
		}
		c.JSON(500, model.Error(1005, "服务器错误"))
		return
	}

	var publishedCount int64
	db.DB.Model(&model.WorkChapter{}).Where("work_id = ? AND status = ?", workID, "published").Count(&publishedCount)

	if publishedCount == 0 {
		c.JSON(400, model.Error(1001, "发布前需至少有一个已发布的章节"))
		return
	}

	now := time.Now()
	db.DB.Model(&work).Updates(map[string]interface{}{
		"status":      "published",
		"update_time": now,
	})

	c.JSON(200, model.Success(gin.H{
		"workId": work.WorkID,
		"status": "published",
	}))
}

func (h *Handler) UpdateChapterOrder(c *gin.Context) {
	creatorID := c.GetString("userId")
	workID := c.Param("workId")

	var work model.Work
	if err := db.DB.Where("work_id = ? AND creator_id = ?", workID, creatorID).First(&work).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, model.Error(1004, "作品不存在"))
			return
		}
		c.JSON(500, model.Error(1005, "服务器错误"))
		return
	}

	var req model.ChapterOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, model.Error(1001, "参数错误"))
		return
	}

	for i, chapterID := range req.ChapterOrder {
		db.DB.Model(&model.WorkChapter{}).Where("chapter_id = ? AND work_id = ?", chapterID, workID).Update("sort_order", i+1)
	}

	db.DB.Model(&work).Update("update_time", time.Now())

	c.JSON(200, model.Success(nil))
}

func parseTags(tags string) []string {
	if tags == "" {
		return []string{}
	}
	return strings.Split(tags, ",")
}
