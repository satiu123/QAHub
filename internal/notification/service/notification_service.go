package service

import (
	"context"
	"encoding/json"
	"log"
	"qahub/internal/notification/model"
	"qahub/internal/notification/store"
	"qahub/pkg/config"
	"qahub/pkg/messaging"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	TopicNotifications = "notification_events"   // 定义消费的主题
	GroupID            = "notification-consumer" // 定义消费者组ID
)

// NotificationService 是通知服务的接口
type NotificationService interface {
	StartConsumer(ctx context.Context)
	Close() error
	ServeWs(c *gin.Context, userID int64)
	GetNotifications(ctx context.Context, userID int64, limit, offset int) ([]*model.Notification, error)
	MarkNotificationsAsRead(ctx context.Context, userID int64, notificationIDs []string) (int64, error)
	DeleteNotification(ctx context.Context, userID int64, notificationID string) error
}

// notificationService 是 NotificationService 接口的具体实现
type notificationService struct {
	store         store.NotificationStore
	hub           *Hub
	kafkaConsumer *messaging.KafkaConsumer
}

// NewNotificationService 创建一个新的 NotificationService 实例
func NewNotificationService(store store.NotificationStore, hub *Hub, cfg config.Kafka) NotificationService {
	s := &notificationService{
		store: store,
		hub:   hub,
	}
	// 必须在service实例化后设置consumer，因为handler需要引用service
	consumer := messaging.NewKafkaConsumer(cfg, TopicNotifications, GroupID, s.getEventHandlers())
	s.kafkaConsumer = consumer
	return s
}

// ServeWs 处理来自客户端的 websocket 请求
func (s *notificationService) ServeWs(c *gin.Context, userID int64) {
	ServeWs(s.hub, c, userID)
}

// GetNotifications 获取用户的通知列表
func (s *notificationService) GetNotifications(ctx context.Context, userID int64, limit, offset int) ([]*model.Notification, error) {
	return s.store.GetByRecipientID(ctx, userID, limit, offset)
}

// MarkNotificationsAsRead 标记通知为已读
// 如果 notificationIDs 为空，则将该用户所有未读通知标记为已读
func (s *notificationService) MarkNotificationsAsRead(ctx context.Context, userID int64, notificationIDs []string) (int64, error) {
	// 如果传入的id列表为空，则获取所有未读通知并标记
	if len(notificationIDs) == 0 {
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

// getEventHandlers 返回此服务处理的事件及其对应的处理函数
func (s *notificationService) getEventHandlers() map[messaging.EventType]messaging.EventHandler {
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

	// 3. 将新创建的通知（包含数据库生成的ID和时间戳）序列化为JSON
	notificationJSON, err := json.Marshal(notification)
	if err != nil {
		log.Printf("failed to marshal notification for push: %v", err)
		// 即使序列化失败，数据也已入库，所以这里只记录日志而不返回错误
		return nil
	}

	// 4. 通过WebSocket Hub推送给在线用户
	s.hub.SendToUser(notification.RecipientID, notificationJSON)

	return nil
}

// StartConsumer 启动 Kafka 消费者，在一个无限循环中读取消息
func (s *notificationService) StartConsumer(ctx context.Context) {
	s.kafkaConsumer.Start(ctx)
}

// Close 方法用于优雅地关闭服务资源，例如 Kafka reader
func (s *notificationService) Close() error {
	return s.kafkaConsumer.Close()
}
