package service

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	// 每次写入操作的超时时间
	writeWait = 10 * time.Second

	// 每次读取操作的超时时间
	pongWait = 60 * time.Second

	// 发送 ping 消息的周期，必须小于 pongWait
	pingPeriod = (pongWait * 9) / 10

	// 允许的最大消息大小
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允许跨域访问
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client 表示一个连接的用户
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID int64
}

// readPump 从 websocket 连接中读取消息
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { _ = c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		// 持续读取来自客户端的消息
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("websocket error: %v", err)
			}
			break
		}
	}
}

// writePump 负责向 websocket 连接写入消息
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub 已关闭发送通道
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Hub 维护活跃客户端集合并向客户端广播消息。
type Hub struct {
	clients    map[int64]*Client
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[int64]*Client),
	}
}

// Run 启动 hub 的事件循环。
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			// 如果同一用户存在旧客户端连接，先注销它以防止陈旧连接。
			if oldClient, ok := h.clients[client.userID]; ok {
				close(oldClient.send)
			}
			h.clients[client.userID] = client
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.userID]; ok {
				delete(h.clients, client.userID)
				close(client.send)
			}
			h.mu.Unlock()
		}
	}
}

// SendToUser 向指定用户发送消息（如果用户已连接）。
func (h *Hub) SendToUser(userID int64, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if client, ok := h.clients[userID]; ok {
		select {
		case client.send <- message:
		default:
			// 如果发送通道已满，说明客户端响应缓慢。
			// 我们关闭连接以防止资源泄漏。
			close(client.send)
			delete(h.clients, userID)
			log.Printf("client %d is lagging, connection closed", userID)
		}
	}
}

// ServeWs 处理来自对等方的 websocket 请求。
func ServeWs(hub *Hub, c *gin.Context, userID int64) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256), userID: userID}
	hub.register <- client

	// 通过在新的 goroutine 中执行所有工作，允许调用者释放所引用的内存。
	go client.writePump()
	go client.readPump()
}
