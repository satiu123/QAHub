package service

import (
	"context"
	"encoding/json"
	"log/slog"
	pb "qahub/api/proto/notification"
	"qahub/notification-service/internal/model"
	"qahub/notification-service/internal/store"
	"qahub/pkg/log"
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
	logger := log.FromContext(ctx)

	notifications, err := s.store.GetByRecipientID(ctx, userID, limit, offset)
	if err != nil {
		logger.Error("获取用户通知失败",
			slog.Int64("user_id", userID),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	logger.Debug("获取用户通知成功",
		slog.Int64("user_id", userID),
		slog.Int("count", len(notifications)),
		slog.Int("limit", int(limit)),
		slog.Int64("offset", offset),
	)
	return notifications, nil
}

// MarkNotificationsAsRead 标记通知为已读
// 如果 notificationIDs 为空，则将该用户所有未读通知标记为已读
func (s *notificationService) MarkNotificationsAsRead(ctx context.Context, userID int64, notificationIDs []string, markAll bool) (int64, error) {
	logger := log.FromContext(ctx)

	// 如果传入的id列表为空且markAll标记为true，则获取所有未读通知并标记
	if len(notificationIDs) == 0 && markAll {
		// 为了简化，这里我们获取所有通知（不分页），在实际应用中可能需要考虑性能
		notifications, err := s.store.GetByRecipientID(ctx, userID, 1000, 0) // 假设最多处理1000条
		if err != nil {
			logger.Error("获取用户未读通知失败",
				slog.Int64("user_id", userID),
				slog.String("error", err.Error()),
			)
			return 0, err
		}
		for _, n := range notifications {
			if n.Status == model.NotificationStatus(pb.NotificationStatus_UNREAD) {
				notificationIDs = append(notificationIDs, n.ID.Hex())
			}
		}
		// 如果没有未读通知，直接返回
		if len(notificationIDs) == 0 {
			logger.Debug("没有未读通知需要标记",
				slog.Int64("user_id", userID),
			)
			return 0, nil
		}
	}

	count, err := s.store.MarkManyAsRead(ctx, notificationIDs, userID)
	if err != nil {
		logger.Error("标记通知为已读失败",
			slog.Int64("user_id", userID),
			slog.Int("notification_count", len(notificationIDs)),
			slog.String("error", err.Error()),
		)
		return 0, err
	}

	logger.Info("通知标记为已读成功",
		slog.Int64("user_id", userID),
		slog.Int64("marked_count", count),
	)
	return count, nil
}

// DeleteNotification 删除一条通知
func (s *notificationService) DeleteNotification(ctx context.Context, userID int64, notificationID string) error {
	logger := log.FromContext(ctx)

	err := s.store.Delete(ctx, notificationID, userID)
	if err != nil {
		logger.Error("删除通知失败",
			slog.Int64("user_id", userID),
			slog.String("notification_id", notificationID),
			slog.String("error", err.Error()),
		)
		return err
	}

	logger.Info("通知删除成功",
		slog.Int64("user_id", userID),
		slog.String("notification_id", notificationID),
	)
	return nil
}

// DeleteNotifications 删除多条通知
func (s *notificationService) DeleteNotifications(ctx context.Context, userID int64, notificationIDs []string) (int64, error) {
	logger := log.FromContext(ctx)

	count, err := s.store.DeleteMany(ctx, notificationIDs, userID)
	if err != nil {
		logger.Error("删除多条通知失败",
			slog.Int64("user_id", userID),
			slog.Int("notification_count", len(notificationIDs)),
			slog.String("error", err.Error()),
		)
		return 0, err
	}

	logger.Info("多条通知删除成功",
		slog.Int64("user_id", userID),
		slog.Int64("deleted_count", count),
	)
	return count, nil
}

// GetUnreadCount 获取用户的未读通知数量
func (s *notificationService) GetUnreadCount(ctx context.Context, userID int64) (int64, error) {
	logger := log.FromContext(ctx)

	count, err := s.store.CountUnread(ctx, userID)
	if err != nil {
		logger.Error("获取未读通知数量失败",
			slog.Int64("user_id", userID),
			slog.String("error", err.Error()),
		)
		return 0, err
	}

	logger.Debug("获取未读通知数量成功",
		slog.Int64("user_id", userID),
		slog.Int64("unread_count", count),
	)
	return count, nil
}

// RegisterHandlers 返回此服务处理的事件及其对应的处理函数
func (s *notificationService) RegisterHandlers() map[messaging.EventType]messaging.EventHandler {
	return map[messaging.EventType]messaging.EventHandler{
		messaging.EventNotificationTriggered: s.handleNotificationTriggered,
	}
}

// handleNotificationTriggered 是处理来自 Kafka 的"通知触发"事件的核心方法
func (s *notificationService) handleNotificationTriggered(ctx context.Context, eventType string, eventMessage []byte) error {
	logger := log.FromContext(ctx)

	var event messaging.NotificationTriggeredEvent
	if err := json.Unmarshal(eventMessage, &event); err != nil {
		logger.Error("解析通知事件失败",
			slog.String("error", err.Error()),
		)
		return err
	}

	// 1. 根据事件创建通知对象
	notification := &model.Notification{
		RecipientID: event.Payload.RecipientID,
		SenderID:    event.Payload.SenderID,
		SenderName:  event.Payload.SenderName,
		Type:        model.NotificationType(event.Payload.NotificationType),
		Content:     event.Payload.Content,
		TargetURL:   event.Payload.TargetURL,
		Status:      model.NotificationStatus(pb.NotificationStatus_UNREAD),
		CreatedAt:   time.Now(),
	}

	// 2. 将通知存入数据库
	if err := s.store.Create(ctx, notification); err != nil {
		logger.Error("保存通知到数据库失败",
			slog.Int64("recipient_id", event.Payload.RecipientID),
			slog.Int64("sender_id", event.Payload.SenderID),
			slog.String("error", err.Error()),
		)
		return err
	}

	logger.Info("通知已创建并保存",
		slog.Int64("recipient_id", notification.RecipientID),
		slog.Int64("sender_id", notification.SenderID),
		slog.String("notification_type", notification.Type.Value()),
	)

	// 3. 通过 gRPC StreamHub 推送给在线用户
	if s.streamHub != nil {
		pbNotification := convertModelToProto(notification)
		s.streamHub.SendToUser(notification.RecipientID, pbNotification)

		logger.Debug("通知已通过gRPC流推送给用户",
			slog.Int64("recipient_id", notification.RecipientID),
		)
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
		Type:        pb.NotificationType(n.Type),
		Content:     n.Content,
		TargetUrl:   n.TargetURL,
		Status:      n.Status.ToProto(),
		CreatedAt:   timestamppb.New(n.CreatedAt),
	}
}
