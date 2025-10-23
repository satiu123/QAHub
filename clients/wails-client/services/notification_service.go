package services

import (
	"context"
	"fmt"

	ntpb "wails-client/api/proto/notification"
)

type NotificationService struct {
	client *GRPCClient
}

func NewNotificationService(client *GRPCClient) *NotificationService {
	return &NotificationService{client: client}
}

// Notification 通知结构
type Notification struct {
	ID          string `json:"id"`
	RecipientID int64  `json:"recipient_id"`
	SenderID    int64  `json:"sender_id"`
	SenderName  string `json:"sender_name"`
	Type        string `json:"type"`
	Content     string `json:"content"`
	TargetURL   string `json:"target_url"`
	IsRead      bool   `json:"is_read"`
	CreatedAt   string `json:"created_at"`
}

// GetNotifications 获取通知列表
func (s *NotificationService) GetNotifications(ctx context.Context, page int32, pageSize int64, unreadOnly bool) ([]Notification, int64, int64, error) {
	if s.client == nil || s.client.NotificationClient == nil {
		return nil, 0, 0, fmt.Errorf("通知服务未初始化")
	}

	authCtx := s.client.NewAuthContext(ctx)
	userID := s.client.GetUserID()

	resp, err := s.client.NotificationClient.GetNotifications(authCtx, &ntpb.GetNotificationsRequest{
		UserId:     userID,
		Page:       page,
		PageSize:   int32(pageSize),
		UnreadOnly: unreadOnly,
	})
	if err != nil {
		return nil, 0, 0, fmt.Errorf("获取通知失败: %w", err)
	}

	notifications := make([]Notification, 0, len(resp.Notifications))
	for _, n := range resp.Notifications {
		notifications = append(notifications, Notification{
			ID:          n.Id,
			RecipientID: n.RecipientId,
			SenderID:    n.SenderId,
			SenderName:  n.SenderName,
			Type:        n.Type,
			Content:     n.Content,
			TargetURL:   n.TargetUrl,
			IsRead:      n.IsRead,
			CreatedAt:   n.CreatedAt.AsTime().Format("2006-01-02 15:04:05"),
		})
	}

	return notifications, resp.Total, resp.UnreadCount, nil
}

// GetUnreadCount 获取未读通知数量
func (s *NotificationService) GetUnreadCount(ctx context.Context) (int64, error) {
	if s.client == nil || s.client.NotificationClient == nil {
		return 0, fmt.Errorf("通知服务未初始化")
	}

	authCtx := s.client.NewAuthContext(ctx)
	userID := s.client.GetUserID()

	resp, err := s.client.NotificationClient.GetUnreadCount(authCtx, &ntpb.GetUnreadCountRequest{
		UserId: userID,
	})
	if err != nil {
		return 0, fmt.Errorf("获取未读数量失败: %w", err)
	}

	return resp.UnreadCount, nil
}

// MarkAsRead 标记通知为已读
func (s *NotificationService) MarkAsRead(ctx context.Context, notificationIDs []string, markAll bool) (int64, error) {
	if s.client == nil || s.client.NotificationClient == nil {
		return 0, fmt.Errorf("通知服务未初始化")
	}

	authCtx := s.client.NewAuthContext(ctx)
	userID := s.client.GetUserID()

	resp, err := s.client.NotificationClient.MarkAsRead(authCtx, &ntpb.MarkAsReadRequest{
		UserId:          userID,
		NotificationIds: notificationIDs,
		MarkAll:         markAll,
	})
	if err != nil {
		return 0, fmt.Errorf("标记已读失败: %w", err)
	}

	return resp.ModifiedCount, nil
}

// DeleteNotification 删除通知
func (s *NotificationService) DeleteNotification(ctx context.Context, notificationID string) error {
	if s.client == nil || s.client.NotificationClient == nil {
		return fmt.Errorf("通知服务未初始化")
	}

	authCtx := s.client.NewAuthContext(ctx)
	userID := s.client.GetUserID()

	_, err := s.client.NotificationClient.DeleteNotification(authCtx, &ntpb.DeleteNotificationRequest{
		UserId:         userID,
		NotificationId: notificationID,
	})
	if err != nil {
		return fmt.Errorf("删除通知失败: %w", err)
	}

	return nil
}
