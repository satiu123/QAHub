package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	ntpb "wails-client/api/proto/notification"
)

// NotificationHandler 是处理接收到的通知的回调函数类型
type NotificationHandler func(notification *Notification)

// NotificationStream 管理通知流连接
type NotificationStream struct {
	client      *GRPCClient
	stream      ntpb.NotificationService_SubscribeNotificationsClient
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.RWMutex
	connected   bool
	handlers    []NotificationHandler
	reconnectCh chan struct{}
}

// NewNotificationStream 创建新的通知流管理器
func NewNotificationStream(client *GRPCClient) *NotificationStream {
	return &NotificationStream{
		client:      client,
		handlers:    make([]NotificationHandler, 0),
		reconnectCh: make(chan struct{}, 1),
	}
}

// AddHandler 添加通知处理器
func (ns *NotificationStream) AddHandler(handler NotificationHandler) {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	ns.handlers = append(ns.handlers, handler)
}

// Start 启动通知流连接
func (ns *NotificationStream) Start() error {
	ns.mu.Lock()
	if ns.connected {
		ns.mu.Unlock()
		return nil
	}
	ns.mu.Unlock()

	// 启动连接和重连逻辑
	go ns.connectLoop()
	return nil
}

// connectLoop 连接循环，支持自动重连
func (ns *NotificationStream) connectLoop() {
	backoff := time.Second
	maxBackoff := time.Minute

	for {
		select {
		case <-ns.reconnectCh:
			// 收到重连信号
			log.Println("Attempting to reconnect notification stream...")
		default:
		}

		if err := ns.connect(); err != nil {
			log.Printf("Failed to connect notification stream: %v", err)
			log.Printf("Retrying in %v...", backoff)
			time.Sleep(backoff)

			// 指数退避
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue
		}

		// 连接成功，重置退避时间
		backoff = time.Second

		// 开始接收消息
		ns.receiveLoop()

		// 如果退出接收循环，等待一段时间后重连
		log.Println("Notification stream disconnected, will reconnect...")
		time.Sleep(2 * time.Second)
	}
}

// connect 建立流连接
func (ns *NotificationStream) connect() error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	if ns.client == nil || ns.client.NotificationClient == nil {
		return fmt.Errorf("notification client not initialized")
	}

	// 创建新的上下文
	ns.ctx, ns.cancel = context.WithCancel(context.Background())
	authCtx := ns.client.NewAuthContext(ns.ctx)
	userID := ns.client.GetUserID()

	// 订阅通知流
	stream, err := ns.client.NotificationClient.SubscribeNotifications(authCtx, &ntpb.SubscribeNotificationsRequest{
		UserId: userID,
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe notifications: %w", err)
	}

	ns.stream = stream
	ns.connected = true
	log.Printf("✅ Notification stream connected for user %d", userID)
	return nil
}

// receiveLoop 接收通知消息的循环
func (ns *NotificationStream) receiveLoop() {
	defer func() {
		ns.mu.Lock()
		ns.connected = false
		ns.stream = nil
		ns.mu.Unlock()
	}()

	for {
		notification, err := ns.stream.Recv()
		if err != nil {
			if err == io.EOF {
				log.Println("Notification stream closed by server")
			} else {
				log.Printf("Error receiving notification: %v", err)
			}
			return
		}

		// 转换为本地通知结构
		localNotification := &Notification{
			ID:          notification.Id,
			RecipientID: notification.RecipientId,
			SenderID:    notification.SenderId,
			SenderName:  notification.SenderName,
			Type:        notification.Type,
			Content:     notification.Content,
			TargetURL:   notification.TargetUrl,
			IsRead:      notification.IsRead,
			CreatedAt:   notification.CreatedAt.AsTime().Format("2006-01-02 15:04:05"),
		}

		log.Printf("📨 Received notification: %s", localNotification.Content)

		// 调用所有注册的处理器
		ns.mu.RLock()
		handlers := make([]NotificationHandler, len(ns.handlers))
		copy(handlers, ns.handlers)
		ns.mu.RUnlock()

		for _, handler := range handlers {
			go handler(localNotification)
		}
	}
}

// Stop 停止通知流
func (ns *NotificationStream) Stop() {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	if ns.cancel != nil {
		ns.cancel()
	}
	ns.connected = false
	log.Println("Notification stream stopped")
}

// IsConnected 检查是否已连接
func (ns *NotificationStream) IsConnected() bool {
	ns.mu.RLock()
	defer ns.mu.RUnlock()
	return ns.connected
}

// Reconnect 触发重连
func (ns *NotificationStream) Reconnect() {
	select {
	case ns.reconnectCh <- struct{}{}:
	default:
	}
}
