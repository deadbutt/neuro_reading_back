package model

import (
	"time"
)

type User struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"-"`
	UserID       string    `gorm:"uniqueIndex;size:32" json:"userId"`
	Account      string    `gorm:"uniqueIndex;size:64" json:"account"`
	Password     string    `gorm:"size:255" json:"-"`
	Nickname     string    `gorm:"size:64" json:"nickname"`
	Avatar       string    `gorm:"size:255" json:"avatar"`
	Bio          string    `gorm:"size:500" json:"bio"`
	Gender       int       `gorm:"default:0" json:"gender"`
	ReadDuration int64     `gorm:"default:0" json:"readDuration"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
}

type UserProfileResponse struct {
	UserID         string `json:"userId"`
	Account        string `json:"account"`
	Nickname       string `json:"nickname"`
	Avatar         string `json:"avatar"`
	Bio            string `json:"bio"`
	Gender         int    `json:"gender"`
	FollowingCount int64  `json:"followingCount"`
	BookshelfCount int64  `json:"bookshelfCount"`
	ReadDuration   int64  `json:"readDuration"`
}

type UpdateProfileRequest struct {
	Nickname string `json:"nickname,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
	Bio      string `json:"bio,omitempty"`
	Gender   *int   `json:"gender,omitempty"`
}

type Follow struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"-"`
	UserID    string    `gorm:"index;size:32" json:"userId"`
	AuthorID  string    `gorm:"index;size:32" json:"authorId"`
	CreatedAt time.Time `json:"-"`
}

type Like struct {
	ID       uint64    `gorm:"primaryKey;autoIncrement" json:"-"`
	UserID   string    `gorm:"index;size:32" json:"userId"`
	TargetID string    `gorm:"index;size:32" json:"targetId"`
	Type     string    `gorm:"index;size:16" json:"type"`
	CreatedAt time.Time `json:"-"`
}

type ReadingHistory struct {
	ID           uint64 `gorm:"primaryKey;autoIncrement" json:"-"`
	HistoryID    string `gorm:"uniqueIndex;size:32" json:"historyId"`
	UserID       string `gorm:"index;size:32" json:"-"`
	ArticleID    string `gorm:"index;size:32" json:"articleId"`
	Title        string `gorm:"size:128" json:"title"`
	Author       string `gorm:"size:50" json:"author"`
	Cover        string `gorm:"size:255" json:"cover"`
	ChapterIndex int    `json:"chapterIndex"`
	ChapterTitle string `gorm:"size:128" json:"chapterTitle"`
	Progress     int    `json:"progress"`
	Position     int    `json:"position"`
	ReadTime     int    `json:"readTime"`
	LastReadTime string `gorm:"size:32" json:"lastReadTime"`
}

type ReadingHistoryResponse struct {
	HistoryID    string `json:"historyId"`
	ArticleID    string `json:"articleId"`
	Title        string `json:"title"`
	Author       string `json:"author"`
	Cover        string `json:"cover"`
	ChapterIndex int    `json:"chapterIndex"`
	ChapterTitle string `json:"chapterTitle"`
	Progress     int    `json:"progress"`
	Position     int    `json:"position"`
	ReadTime     int    `json:"readTime"`
	LastReadTime string `json:"lastReadTime"`
}
