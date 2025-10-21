package service

import (
	"context"
	"encoding/json"
	"log"
	pb "qahub/api/proto/notification"
	"qahub/notification-service/internal/model"
	"qahub/notification-service/internal/store"
	"qahub/pkg/messaging"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	TopicNotifications = "notification_events"   // 定义消费的主题
	GroupID            = "notification-consumer" // 定义消费者组ID
)

// NotificationService 是通知服务的接口
type NotificationService interface {
	GetNotifications(ctx context.Context, userID int64, limit int32, offset int64) ([]*model.Notification, error)
	MarkNotificationsAsRead(ctx context.Context, userID int64, notificationIDs []string, markAll bool) (int64, error)
	DeleteNotification(ctx context.Context, userID int64, notificationID string) error
	DeleteNotifications(ctx context.Context, userID int64, notificationIDs []string) (int64, error)
	GetUnreadCount(ctx context.Context, userID int64) (int64, error)
	GetStreamHub() *StreamHub
}

// notificationService 是 NotificationService 接口的具体实现
type notificationService struct {
	store     store.NotificationStore
	streamHub *StreamHub
}

// NewNotificationService 创建一个新的 NotificationService 实例
func NewNotificationService(store store.NotificationStore, streamHub *StreamHub) *notificationService {
	service := &notificationService{
		store:     store,
		streamHub: streamHub,
	}
	return service
}

// GetStreamHub 返回 StreamHub 实例
func (s *notificationService) GetStreamHub() *StreamHub {
	return s.streamHub
}

// GetNotifications 获取用户的通知列表
func (s *notificationService) GetNotifications(ctx context.Context, userID int64, limit int32, offset int64) ([]*model.Notification, error) {
	return s.store.GetByRecipientID(ctx, userID, limit, offset)
}

// MarkNotificationsAsRead 标记通知为已读
// 如果 notificationIDs 为空，则将该用户所有未读通知标记为已读
func (s *notificationService) MarkNotificationsAsRead(ctx context.Context, userID int64, notificationIDs []string, markAll bool) (int64, error) {
	// 如果传入的id列表为空且markAll标记为true，则获取所有未读通知并标记
	if len(notificationIDs) == 0 && markAll {
		// 为了简化，这里我们获取所有通知（不分页），在实际应用中可能需要考虑性能
		notifications, err := s.store.GetByRecipientID(ctx, userID, 1000, 0) // 假设最多处理1000条
		if err != nil {
			return 0, err
		}
		for _, n := range notifications {
			if !n.IsRead {
				notificationIDs = append(notificationIDs, n.ID.Hex())
			}
		}
		// 如果没有未读通知，直接返回
		if len(notificationIDs) == 0 {
			return 0, nil
		}
	}

	return s.store.MarkManyAsRead(ctx, notificationIDs, userID)
}

// DeleteNotification 删除一条通知
func (s *notificationService) DeleteNotification(ctx context.Context, userID int64, notificationID string) error {
	return s.store.Delete(ctx, notificationID, userID)
}

// DeleteNotifications 删除多条通知
func (s *notificationService) DeleteNotifications(ctx context.Context, userID int64, notificationIDs []string) (int64, error) {
	return s.store.DeleteMany(ctx, notificationIDs, userID)
}

// GetUnreadCount 获取用户的未读通知数量
func (s *notificationService) GetUnreadCount(ctx context.Context, userID int64) (int64, error) {
	return s.store.CountUnread(ctx, userID)
}

// RegisterHandlers 返回此服务处理的事件及其对应的处理函数
func (s *notificationService) RegisterHandlers() map[messaging.EventType]messaging.EventHandler {
	return map[messaging.EventType]messaging.EventHandler{
		messaging.EventNotificationTriggered: s.handleNotificationTriggered,
	}
}

// handleNotificationTriggered 是处理来自 Kafka 的“通知触发”事件的核心方法
func (s *notificationService) handleNotificationTriggered(ctx context.Context, eventType string, eventMessage []byte) error {
	var event messaging.NotificationTriggeredEvent
	if err := json.Unmarshal(eventMessage, &event); err != nil {
		log.Printf("failed to unmarshal event: %v", err)
		return err
	}

	// 1. 根据事件创建通知对象
	notification := &model.Notification{
		RecipientID: event.Payload.RecipientID,
		SenderID:    event.Payload.SenderID,
		SenderName:  event.Payload.SenderName,
		Type:        event.Payload.NotificationType,
		Content:     event.Payload.Content,
		TargetURL:   event.Payload.TargetURL,
		IsRead:      false,
		CreatedAt:   time.Now(),
	}

	// 2. 将通知存入数据库
	if err := s.store.Create(ctx, notification); err != nil {
		log.Printf("failed to save notification: %v", err)
		return err
	}

	// 3. 通过 gRPC StreamHub 推送给在线用户
	if s.streamHub != nil {
		pbNotification := convertModelToProto(notification)
		s.streamHub.SendToUser(notification.RecipientID, pbNotification)
		log.Printf("Notification pushed to user %d via gRPC stream", notification.RecipientID)
	}

	return nil
}

// convertModelToProto 将 model.Notification 转换为 pb.Notification
func convertModelToProto(n *model.Notification) *pb.Notification {
	return &pb.Notification{
		Id:          n.ID.Hex(),
		RecipientId: n.RecipientID,
		SenderId:    n.SenderID,
		SenderName:  n.SenderName,
		Type:        n.Type,
		Content:     n.Content,
		TargetUrl:   n.TargetURL,
		IsRead:      n.IsRead,
		CreatedAt:   timestamppb.New(n.CreatedAt),
	}
}
