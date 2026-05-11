package model

type Creator struct {
	ID            uint64 `gorm:"primaryKey;autoIncrement" json:"-"`
	CreatorID     string `gorm:"uniqueIndex;size:32" json:"creatorId"`
	Account       string `gorm:"uniqueIndex;size:64" json:"account"`
	Password      string `gorm:"size:255" json:"-"`
	Name          string `gorm:"size:64" json:"name"`
	Avatar        string `gorm:"size:255" json:"avatar"`
	Description   string `gorm:"size:500" json:"description"`
	Email         string `gorm:"size:128" json:"email"`
	CreateTime    string `gorm:"size:32" json:"createTime"`
	LastLoginTime string `gorm:"size:32" json:"lastLoginTime"`
	Status        int    `gorm:"default:1" json:"status"`
}

type CreatorLoginRequest struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type CreatorRegisterRequest struct {
	Account         string `json:"account" binding:"required,min=3,max=32"`
	Password        string `json:"password" binding:"required,min=6,max=32"`
	ConfirmPassword string `json:"confirmPassword" binding:"required"`
	Name            string `json:"name" binding:"required,min=2,max=32"`
	Email           string `json:"email" binding:"required,email"`
}

type CreatorProfileRequest struct {
	Name        string `json:"name"`
	Avatar      string `json:"avatar"`
	Description string `json:"description"`
	Email       string `json:"email"`
}

type CreatorProfileResponse struct {
	CreatorID     string `json:"creatorId"`
	Account       string `json:"account"`
	Name          string `json:"name"`
	Avatar        string `json:"avatar"`
	Description   string `json:"description"`
	Email         string `json:"email"`
	ArticleCount  int    `json:"articleCount"`
	TotalWords    int    `json:"totalWords"`
	TotalReads    int    `json:"totalReads"`
	TotalLikes    int    `json:"totalLikes"`
	TotalComments int    `json:"totalComments"`
	CreateTime    string `json:"createTime"`
	LastLoginTime string `json:"lastLoginTime"`
}

type CreatorArticleResponse struct {
	ArticleID      string   `json:"articleId"`
	Title          string   `json:"title"`
	Summary        string   `json:"summary"`
	Tags           []string `json:"tags"`
	WordCount      int      `json:"wordCount"`
	ChapterCount   int      `json:"chapterCount"`
	Status         string   `json:"status"`
	ReadCount      int      `json:"readCount"`
	LikeCount      int      `json:"likeCount"`
	CommentCount   int      `json:"commentCount"`
	PublishTime    string   `json:"publishTime"`
	LastUpdateTime string   `json:"lastUpdateTime"`
}

type CreatorStatsResponse struct {
	TotalArticles int `json:"totalArticles"`
	TotalWords    int `json:"totalWords"`
	TotalReads    int `json:"totalReads"`
	TotalLikes    int `json:"totalLikes"`
	TotalComments int `json:"totalComments"`
	TodayReads    int `json:"todayReads"`
	TodayLikes    int `json:"todayLikes"`
	TodayComments int `json:"todayComments"`
	WeekReads     int `json:"weekReads"`
	MonthReads    int `json:"monthReads"`
}
