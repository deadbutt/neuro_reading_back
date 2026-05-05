package model

type BookshelfItem struct {
	ID              uint64 `gorm:"primaryKey;autoIncrement" json:"-"`
	UserID          string `gorm:"size:32;index" json:"-"`
	BookID          string `gorm:"size:32;index" json:"bookId"`
	LastReadChapter string `gorm:"size:128" json:"lastReadChapter"`
	LastReadTime    string `gorm:"size:32" json:"lastReadTime"`
	Progress        int    `gorm:"default:0" json:"progress"`
	IsFinished      bool   `gorm:"default:false" json:"isFinished"`
	ChapterID       string `gorm:"size:32" json:"-"`
	Position        int    `gorm:"default:0" json:"-"`
}

type BookshelfItemResponse struct {
	BookID          string         `json:"bookId"`
	Title           string         `json:"title"`
	Author          AuthorResponse `json:"author"`
	Cover           string         `json:"cover"`
	LastReadChapter string         `json:"lastReadChapter"`
	LastReadTime    string         `json:"lastReadTime"`
	Progress        int            `json:"progress"`
	IsFinished      bool           `json:"isFinished"`
}

type UpdateProgressRequest struct {
	ChapterID string `json:"chapterId" binding:"required"`
	Progress  int    `json:"progress" binding:"required,min=0,max=100"`
	Position  int    `json:"position" binding:"required"`
}