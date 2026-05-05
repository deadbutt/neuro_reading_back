package model

type FeedActivity struct {
	ID              uint64 `gorm:"primaryKey;autoIncrement" json:"-"`
	FeedID          string `gorm:"uniqueIndex;size:32" json:"feedId"`
	AuthorID        string `gorm:"size:32;index" json:"authorId"`
	AuthorName      string `gorm:"size:64" json:"authorName"`
	AuthorAvatar    string `gorm:"size:255" json:"authorAvatar"`
	PublishTime     string `gorm:"size:32" json:"publishTime"`
	ActivityContent string `gorm:"type:text" json:"activityContent"`
	BookID          string `gorm:"size:32" json:"bookId"`
	BookCover       string `gorm:"size:255" json:"bookCover"`
	ChapterPreview  string `gorm:"size:500" json:"chapterPreview"`
	ReadHeat        string `gorm:"size:64" json:"readHeat"`
	LikeCount       int    `gorm:"default:0" json:"likeCount"`
	CommentCount    int    `gorm:"default:0" json:"commentCount"`
}

type FeedActivityResponse struct {
	FeedID          string `json:"feedId"`
	AuthorID        string `json:"authorId"`
	AuthorName      string `json:"authorName"`
	AuthorAvatar    string `json:"authorAvatar"`
	PublishTime     string `json:"publishTime"`
	ActivityContent string `json:"activityContent"`
	BookID          string `json:"bookId"`
	BookCover       string `json:"bookCover"`
	ChapterPreview  string `json:"chapterPreview"`
	ReadHeat        string `json:"readHeat"`
	LikeCount       int    `json:"likeCount"`
	CommentCount    int    `json:"commentCount"`
	IsLiked         bool   `json:"isLiked"`
}

type Follow struct {
	ID         uint64 `gorm:"primaryKey;autoIncrement" json:"-"`
	UserID     string `gorm:"size:32;index" json:"-"`
	AuthorID   string `gorm:"size:32;index" json:"-"`
}

type Like struct {
	ID       uint64 `gorm:"primaryKey;autoIncrement" json:"-"`
	UserID   string `gorm:"size:32;index" json:"-"`
	TargetID string `gorm:"size:32;index" json:"-"`
	Type     string `gorm:"size:16;index" json:"-"`
}