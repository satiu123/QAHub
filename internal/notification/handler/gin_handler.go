package handler

import (
	"net/http"
	"qahub/internal/notification/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service service.NotificationService
}

func NewHandler(s service.NotificationService) *Handler {
	return &Handler{service: s}
}

func (h *Handler) WsHandler(c *gin.Context) {
	// 1. 从context获取userID, userID在之前的auth middleware中被设置
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	// 2. 调用service中的ServeWs方法
	h.service.ServeWs(c, userID.(int64))
}

// GetNotifications 获取通知列表
func (h *Handler) GetNotifications(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	notifications, err := h.service.GetNotifications(c.Request.Context(), userID.(int64), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取通知失败"})
		return
	}

	c.JSON(http.StatusOK, notifications)
}

// MarkAsRead 标记通知为已读
func (h *Handler) MarkAsRead(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	type MarkReadRequest struct {
		NotificationIDs []string `json:"notification_ids"`
	}

	var req MarkReadRequest
	// 如果请求体为空，req.NotificationIDs会是一个空切片，service层会处理“全部已读”
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	modifiedCount, err := h.service.MarkNotificationsAsRead(c.Request.Context(), userID.(int64), req.NotificationIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "标记已读失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"modified_count": modifiedCount})
}

// DeleteNotification 删除一条通知
func (h *Handler) DeleteNotification(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	notificationID := c.Param("id")
	if notificationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少通知ID"})
		return
	}

	if err := h.service.DeleteNotification(c.Request.Context(), userID.(int64), notificationID);
		err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除通知失败"})
		return
	}

	c.Status(http.StatusNoContent)
}
