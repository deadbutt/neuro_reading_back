package model

import (
	"time"
)

type User struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"-"`
	UserID       string    `gorm:"uniqueIndex;size:32" json:"userId"`
	Account      string    `gorm:"uniqueIndex;size:64" json:"account"`
	Password     string    `gorm:"size:64" json:"-"`
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