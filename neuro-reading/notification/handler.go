package notification

import (
	"neuro-reading/db"
	"neuro-reading/model"
	"neuro-reading/utils"

	"github.com/gin-gonic/gin"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) GetList(c *gin.Context) {
	userID := c.GetString("userId")

	var req model.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req.Page = 1
		req.PageSize = 20
	}

	var notifications []model.Notification
	var total int64

	db.DB.Model(&model.Notification{}).Where("user_id = ?", userID).Count(&total)
	db.DB.Where("user_id = ?", userID).Order("create_time DESC").Offset(req.GetOffset()).Limit(req.GetLimit()).Find(&notifications)

	list := make([]model.NotificationResponse, 0)
	for _, n := range notifications {
		list = append(list, model.NotificationResponse{
			NotificationID: n.NotificationID,
			Type:           n.Type,
			Title:          n.Title,
			Content:        n.Content,
			RelatedID:      n.RelatedID,
			FromUserID:     n.FromUserID,
			FromUserName:   n.FromUserName,
			FromUserAvatar: n.FromUserAvatar,
			IsRead:         n.IsRead,
			CreateTime:     n.CreateTime,
		})
	}

	c.JSON(200, model.PageSuccess(list, total, req.Page, req.PageSize))
}

func (h *Handler) GetUnreadCount(c *gin.Context) {
	userID := c.GetString("userId")

	var unreadCount int64
	var totalCount int64

	db.DB.Model(&model.Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Count(&unreadCount)
	db.DB.Model(&model.Notification{}).Where("user_id = ?", userID).Count(&totalCount)

	c.JSON(200, model.Success(model.NotificationCountResponse{
		UnreadCount: unreadCount,
		TotalCount:  totalCount,
	}))
}

func (h *Handler) MarkAsRead(c *gin.Context) {
	userID := c.GetString("userId")
	notificationID := c.Param("notificationId")

	if notificationID == "" {
		c.JSON(400, model.Error(1001, "参数错误"))
		return
	}

	if err := db.DB.Model(&model.Notification{}).Where("user_id = ? AND notification_id = ?", userID, notificationID).Update("is_read", true).Error; err != nil {
		c.JSON(500, model.Error(1005, "服务器内部错误"))
		return
	}

	c.JSON(200, model.Success(nil))
}

func (h *Handler) MarkAllAsRead(c *gin.Context) {
	userID := c.GetString("userId")

	if err := db.DB.Model(&model.Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Update("is_read", true).Error; err != nil {
		c.JSON(500, model.Error(1005, "服务器内部错误"))
		return
	}

	c.JSON(200, model.Success(nil))
}

func CreateNotification(userID, nType, title, content, relatedID, fromUserID, fromUserName, fromUserAvatar string) {
	notification := model.Notification{
		NotificationID: utils.GenerateNotificationID(),
		UserID:         userID,
		Type:           nType,
		Title:          title,
		Content:        content,
		RelatedID:      relatedID,
		FromUserID:     fromUserID,
		FromUserName:   fromUserName,
		FromUserAvatar: fromUserAvatar,
		IsRead:         false,
		CreateTime:     utils.CurrentTime(),
	}

	db.DB.Create(&notification)
}