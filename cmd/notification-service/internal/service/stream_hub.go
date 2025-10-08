package service

import (
	"log"
	"sync"

	pb "qahub/api/proto/notification"
)

// StreamClient 表示一个通过 gRPC 流连接的客户端
type StreamClient struct {
	UserID int64
	Stream pb.NotificationService_SubscribeNotificationsServer
	Done   chan struct{}
}

// StreamHub 管理所有活跃的 gRPC 流式连接
type StreamHub struct {
	// key: userID, value: StreamClient
	clients map[int64]*StreamClient
	mu      sync.RWMutex

	// 用于注册新客户端
	register chan *StreamClient
	// 用于注销客户端
	unregister chan *StreamClient
	// 用于广播通知
	broadcast chan *NotificationMessage
}

// NotificationMessage 包含要发送的通知和目标用户ID
type NotificationMessage struct {
	UserID       int64
	Notification *pb.Notification
}

// NewStreamHub 创建一个新的流式通知中心
func NewStreamHub() *StreamHub {
	return &StreamHub{
		clients:    make(map[int64]*StreamClient),
		register:   make(chan *StreamClient, 10),
		unregister: make(chan *StreamClient, 10),
		broadcast:  make(chan *NotificationMessage, 100),
	}
}

// Run 启动 StreamHub 的事件循环
func (h *StreamHub) Run() {
	log.Println("StreamHub started")
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			// 如果同一用户已存在连接，关闭旧连接
			if oldClient, ok := h.clients[client.UserID]; ok {
				log.Printf("User %d reconnecting, closing old stream", client.UserID)
				close(oldClient.Done)
			}
			h.clients[client.UserID] = client
			log.Printf("User %d registered for streaming, total clients: %d", client.UserID, len(h.clients))
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; ok {
				delete(h.clients, client.UserID)
				close(client.Done)
				log.Printf("User %d unregistered from streaming, total clients: %d", client.UserID, len(h.clients))
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			if client, ok := h.clients[message.UserID]; ok {
				// 尝试发送通知到客户端流
				if err := client.Stream.Send(message.Notification); err != nil {
					log.Printf("Failed to send notification to user %d: %v", message.UserID, err)
					// 发送失败，注销该客户端
					go func() {
						h.unregister <- client
					}()
				} else {
					log.Printf("Notification sent to user %d via gRPC stream", message.UserID)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Register 注册一个新的流式客户端
func (h *StreamHub) Register(client *StreamClient) {
	h.register <- client
}

// Unregister 注销一个流式客户端
func (h *StreamHub) Unregister(client *StreamClient) {
	h.unregister <- client
}

// SendToUser 向指定用户发送通知（如果用户已连接）
func (h *StreamHub) SendToUser(userID int64, notification *pb.Notification) {
	h.broadcast <- &NotificationMessage{
		UserID:       userID,
		Notification: notification,
	}
}

// GetClientCount 返回当前连接的客户端数量
func (h *StreamHub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
