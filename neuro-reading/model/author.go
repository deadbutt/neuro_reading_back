package model

type Author struct {
	ID          uint64 `gorm:"primaryKey;autoIncrement" json:"-"`
	AuthorID    string `gorm:"uniqueIndex;size:32" json:"authorId"`
	Name        string `gorm:"size:64" json:"name"`
	Avatar      string `gorm:"size:255" json:"avatar"`
	Description string `gorm:"size:500" json:"description"`
}

type AuthorResponse struct {
	AuthorID    string `json:"authorId"`
	Name        string `json:"name"`
	Avatar      string `json:"avatar"`
	Description string `json:"description"`
}

type AuthorProfileResponse struct {
	AuthorID      string `json:"authorId"`
	Name          string `json:"name"`
	Avatar        string `json:"avatar"`
	Description   string `json:"description"`
	WorksCount    int    `json:"worksCount"`
	FollowersCount int64 `json:"followersCount"`
	TotalWords    int64  `json:"totalWords"`
	IsFollowing   bool   `json:"isFollowing"`
}

type AuthorActivityResponse struct {
	ActivityID      string `json:"activityId"`
	AuthorID        string `json:"authorId"`
	AuthorName      string `json:"authorName"`
	AuthorAvatar    string `json:"authorAvatar"`
	Type            string `json:"type"`
	Content         string `json:"content"`
	BookID          string `json:"bookId"`
	BookTitle       string `json:"bookTitle"`
	ChapterTitle    string `json:"chapterTitle"`
	ChapterPreview  string `json:"chapterPreview"`
	ReadHeat        string `json:"readHeat"`
	CreateTime      string `json:"createTime"`
	LikeCount       int    `json:"likeCount"`
	CommentCount    int    `json:"commentCount"`
}