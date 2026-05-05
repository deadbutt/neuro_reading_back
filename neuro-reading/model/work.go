package model

import "time"

type Work struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"-"`
	WorkID       string    `gorm:"uniqueIndex;size:32" json:"workId"`
	CreatorID    string    `gorm:"index;size:32" json:"creatorId"`
	Title        string    `gorm:"size:128" json:"title"`
	Summary      string    `gorm:"size:1000" json:"summary"`
	Tags         string    `gorm:"size:255" json:"-"`
	Cover        string    `gorm:"size:255" json:"cover"`
	Status       string    `gorm:"size:16;default:draft" json:"status"`
	WordCount    int       `gorm:"default:0" json:"wordCount"`
	ChapterCount int       `gorm:"default:0" json:"chapterCount"`
	CreateTime   time.Time `json:"createTime"`
	UpdateTime   time.Time `json:"updateTime"`
}

type WorkChapter struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"-"`
	ChapterID  string    `gorm:"uniqueIndex;size:32" json:"chapterId"`
	WorkID     string    `gorm:"index;size:32" json:"workId"`
	Title      string    `gorm:"size:128" json:"title"`
	Content    string    `gorm:"type:longtext" json:"content"`
	WordCount  int       `gorm:"default:0" json:"wordCount"`
	Status     string    `gorm:"size:16;default:draft" json:"status"`
	SortOrder  int       `gorm:"default:0" json:"sortOrder"`
	CreateTime time.Time `json:"createTime"`
	UpdateTime time.Time `json:"updateTime"`
}

type CreateWorkRequest struct {
	Title   string   `json:"title" binding:"required,min=1,max=50"`
	Summary string   `json:"summary" binding:"max=500"`
	Tags    []string `json:"tags"`
	Cover   string   `json:"cover"`
}

type UpdateWorkRequest struct {
	Title   string   `json:"title"`
	Summary string   `json:"summary"`
	Tags    []string `json:"tags"`
	Cover   string   `json:"cover"`
}

type CreateChapterRequest struct {
	Title   string `json:"title" binding:"required,min=1,max=50"`
	Content string `json:"content" binding:"required"`
	Status  string `json:"status"`
}

type UpdateChapterRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Status  string `json:"status"`
}

type ChapterOrderRequest struct {
	ChapterOrder []string `json:"chapterOrder" binding:"required"`
}

type WorkResponse struct {
	WorkID       string   `json:"workId"`
	Title        string   `json:"title"`
	Summary      string   `json:"summary"`
	Tags         []string `json:"tags"`
	Cover        string   `json:"cover"`
	Status       string   `json:"status"`
	ChapterCount int      `json:"chapterCount"`
	WordCount    int      `json:"wordCount"`
	CreateTime   string   `json:"createTime"`
	UpdateTime   string   `json:"updateTime"`
}

type WorkDetailResponse struct {
	WorkID       string                  `json:"workId"`
	Title        string                  `json:"title"`
	Summary      string                  `json:"summary"`
	Tags         []string                `json:"tags"`
	Cover        string                  `json:"cover"`
	Status       string                  `json:"status"`
	WordCount    int                     `json:"wordCount"`
	Chapters     []WorkChapterResponse   `json:"chapters"`
	CreateTime   string                  `json:"createTime"`
	UpdateTime   string                  `json:"updateTime"`
}

type WorkChapterResponse struct {
	ChapterID  string `json:"chapterId"`
	Title      string `json:"title"`
	WordCount  int    `json:"wordCount"`
	Status     string `json:"status"`
	SortOrder  int    `json:"sortOrder"`
	CreateTime string `json:"createTime"`
	UpdateTime string `json:"updateTime"`
}

type WorkChapterContentResponse struct {
	ChapterID  string `json:"chapterId"`
	WorkID     string `json:"workId"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	WordCount  int    `json:"wordCount"`
	Status     string `json:"status"`
	CreateTime string `json:"createTime"`
	UpdateTime string `json:"updateTime"`
}
