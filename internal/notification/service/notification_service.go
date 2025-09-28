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
)

const (
	TopicNotifications = "notification_events"   // 定义消费的主题
	GroupID            = "notification-consumer" // 定义消费者组ID
)

// NotificationService 是处理通知相关业务逻辑的核心服务
type NotificationService struct {
	store         store.NotificationStore
	hub           *Hub
	kafkaConsumer *messaging.KafkaConsumer
}

// NewNotificationService 创建一个新的 NotificationService 实例
func NewNotificationService(store store.NotificationStore, hub *Hub, cfg config.Kafka) *NotificationService {
	service := &NotificationService{
		store: store,
		hub:   hub,
	}
	// 必须在service实例化后设置consumer，因为handler需要引用service
	consumer := messaging.NewKafkaConsumer(cfg, TopicNotifications, GroupID, service.getEventHandlers())
	service.kafkaConsumer = consumer
	return service
}

// getEventHandlers 返回此服务处理的事件及其对应的处理函数
func (s *NotificationService) getEventHandlers() map[messaging.EventType]messaging.EventHandler {
	return map[messaging.EventType]messaging.EventHandler{
		messaging.EventNotificationTriggered: s.handleNotificationTriggered,
	}
}

// handleNotificationTriggered 是处理来自 Kafka 的“通知触发”事件的核心方法
func (s *NotificationService) handleNotificationTriggered(ctx context.Context, eventType string, eventMessage []byte) error {
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
func (s *NotificationService) StartConsumer(ctx context.Context) {
	s.kafkaConsumer.Start(ctx)
}

// Close 方法用于优雅地关闭服务资源，例如 Kafka reader
func (s *NotificationService) Close() error {
	return s.kafkaConsumer.Close()
}
