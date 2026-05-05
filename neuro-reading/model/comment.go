package model

type Comment struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement" json:"-"`
	CommentID string `gorm:"uniqueIndex;size:32" json:"commentId"`
	BookID    string `gorm:"size:32;index" json:"bookId"`
	UserID    string `gorm:"size:32;index" json:"userId"`
	Content   string `gorm:"type:text" json:"content"`
	ParentID  *string `gorm:"size:32;index" json:"-"`
	LikeCount int    `gorm:"default:0" json:"likeCount"`
	ReplyCount int   `gorm:"default:0" json:"replyCount"`
	CreateTime string `gorm:"size:32" json:"createTime"`
}

type CommentResponse struct {
	CommentID string               `json:"commentId"`
	BookID    string               `json:"bookId"`
	UserID    string               `json:"userId"`
	UserName  string               `json:"userName"`
	UserAvatar string              `json:"userAvatar"`
	Content   string               `json:"content"`
	CreateTime string              `json:"createTime"`
	LikeCount int                  `json:"likeCount"`
	IsLiked   bool                 `json:"isLiked"`
	ReplyCount int                 `json:"replyCount"`
	Replies   []CommentReplyResponse `json:"replies"`
}

type CommentReplyResponse struct {
	ReplyID    string `json:"replyId"`
	UserID     string `json:"userId"`
	UserName   string `json:"userName"`
	UserAvatar string `json:"userAvatar"`
	Content    string `json:"content"`
	CreateTime string `json:"createTime"`
	ToUserName string `json:"toUserName"`
}

type ParagraphComment struct {
	ID             uint64 `gorm:"primaryKey;autoIncrement" json:"-"`
	CommentID      string `gorm:"uniqueIndex;size:32" json:"commentId"`
	ChapterID      string `gorm:"size:32;index" json:"chapterId"`
	ParagraphIndex int    `json:"paragraphIndex"`
	UserID         string `gorm:"size:32;index" json:"userId"`
	Content        string `gorm:"type:text" json:"content"`
	LikeCount      int    `gorm:"default:0" json:"likeCount"`
	CreateTime     string `gorm:"size:32" json:"createTime"`
}

type ParagraphCommentResponse struct {
	CommentID      string `json:"commentId"`
	ChapterID      string `json:"chapterId"`
	ParagraphIndex int    `json:"paragraphIndex"`
	UserID         string `json:"userId"`
	UserName       string `json:"userName"`
	UserAvatar     string `json:"userAvatar"`
	Content        string `json:"content"`
	CreateTime     string `json:"createTime"`
	LikeCount      int    `json:"likeCount"`
	IsLiked        bool   `json:"isLiked"`
}