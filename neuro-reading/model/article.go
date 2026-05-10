package model

import "time"

type Article struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"-"`
	ArticleID    string    `gorm:"uniqueIndex;size:32" json:"articleId"`
	CreatorID    string    `gorm:"index;size:32" json:"creatorId,omitempty"`
	Title        string    `gorm:"index;size:100;not null" json:"title"`
	Author       string    `gorm:"index;size:50" json:"author"`
	Summary      string    `gorm:"size:500" json:"summary"`
	Cover        string    `gorm:"size:255" json:"cover,omitempty"`
	Tags         string    `gorm:"size:200" json:"tags,omitempty"`
	WordCount    int       `gorm:"default:0" json:"wordCount"`
	ChapterCount int       `gorm:"default:0" json:"chapterCount"`
	Status       string    `gorm:"index;size:20;default:draft" json:"status"`
	PublishTime  time.Time `json:"publishTime,omitempty"`
	UpdatedAt    time.Time `gorm:"index" json:"-"`
	CreatedAt    time.Time `json:"-"`
}

type ArticleIndex struct {
	ArticleID      string   `json:"articleId"`
	CreatorID      string   `json:"creatorId,omitempty"`
	Title          string   `json:"title"`
	Author         string   `json:"author"`
	Summary        string   `json:"summary"`
	Cover          string   `json:"cover,omitempty"`
	WordCount      int      `json:"wordCount"`
	ChapterCount   int      `json:"chapterCount"`
	Tags           []string `json:"tags,omitempty"`
	Status         string   `json:"status"`
	LastUpdateTime string   `json:"lastUpdateTime"`
}

type ArticleMeta struct {
	ArticleID      string        `json:"articleId"`
	CreatorID      string        `json:"creatorId,omitempty"`
	Title          string        `json:"title"`
	Author         string        `json:"author"`
	Summary        string        `json:"summary"`
	Cover          string        `json:"cover,omitempty"`
	Tags           []string      `json:"tags,omitempty"`
	WordCount      int           `json:"wordCount"`
	ChapterCount   int           `json:"chapterCount"`
	Status         string        `json:"status"`
	PublishTime    string        `json:"publishTime"`
	LastUpdateTime string        `json:"lastUpdateTime"`
	Chapters       []ChapterMeta `json:"chapters"`
}

type ChapterMeta struct {
	Index     int    `json:"index"`
	ChapterID string `json:"chapterId"`
	Title     string `json:"title"`
	WordCount int    `json:"wordCount"`
	Content   string `json:"-"`
}
