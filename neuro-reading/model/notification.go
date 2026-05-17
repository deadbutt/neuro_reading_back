package model

type Notification struct {
	ID             uint64 `gorm:"primaryKey;autoIncrement" json:"-"`
	NotificationID string `gorm:"uniqueIndex;size:32" json:"notificationId"`
	UserID         string `gorm:"size:32;index" json:"-"`
	Type           string `gorm:"size:16" json:"type"`
	Title          string `gorm:"size:128" json:"title"`
	Content        string `gorm:"type:text" json:"content"`
	RelatedID      string `gorm:"size:32" json:"relatedId"`
	FromUserID     string `gorm:"size:32" json:"fromUserId"`
	FromUserName   string `gorm:"size:64" json:"fromUserName"`
	FromUserAvatar string `gorm:"size:255" json:"fromUserAvatar"`
	IsRead         bool   `gorm:"default:false" json:"isRead"`
	CreateTime     string `gorm:"size:32" json:"createTime"`
}

type NotificationResponse struct {
	NotificationID string `json:"notificationId"`
	Type           string `json:"type"`
	Title          string `json:"title"`
	Content        string `json:"content"`
	RelatedID      string `json:"relatedId"`
	FromUserID     string `json:"fromUserId"`
	FromUserName   string `json:"fromUserName"`
	FromUserAvatar string `json:"fromUserAvatar"`
	IsRead         bool   `json:"isRead"`
	CreateTime     string `json:"createTime"`
}

type NotificationCountResponse struct {
	UnreadCount int64 `json:"unreadCount"`
	TotalCount  int64 `json:"totalCount"`
}