# 通知服务 (Notification Service) - 详细实施计划

## 1. 概述

通知服务的目标是实现一个实时的消息推送系统。当触发特定事件（如新回答、新评论）时，能够将通知实时推送给相关用户，并保存历史通知记录。

根据项目总体规划，技术栈选型如下：

- **语言**: Go
- **实时通信**: WebSocket (面向浏览器客户端)
- **消息队列**: Kafka (用于服务间异步解耦)
- **数据库**: MongoDB (用于存储通知历史)

---

## 2. 数据流设计

```
+-------------+      (1) Event      +---------------+      (2) Consume      +------------------------+
|             |-------------------->|               |-------------------->|                        |
|  Q&A Service|   (e.g., New Answer) |     Kafka     |  (Notification Svc) |  Kafka Consumer        |
|             |                     |               |                     |                        |
+-------------+                     +---------------+                     +-----------+------------+
                                                                                      | (3) Process & Save
                                                                                      |
                                                                                      v
+-------------+      (5) Push Msg     +------------------------+      (4) Store      +------------------------+
|             |<--------------------|                        |-------------------->|                        |
|   Browser   |  (WebSocket Conn)   |    WebSocket Manager   |                     |       MongoDB          |
|             |                     |                        |                     | (Notification History) |
+-------------+                     +------------------------+                     +------------------------+

```

1.  **事件触发**: 问答服务（Q&A Service）中发生业务事件（例如，用户A回答了用户B的问题）。
2.  **发送消息**: 问答服务向 Kafka 的特定主题（Topic）发送一条消息，如 `qa_events`。
3.  **消费消息**: 通知服务的 Kafka 消费者监听到这条消息。
4.  **处理与存储**:
    *   消费者解析消息，生成一条结构化的通知数据。
    *   将这条通知数据存入 MongoDB 中，作为历史记录。
5.  **实时推送**:
    *   通知服务通过 WebSocket 管理器，查找该通知的目标用户（用户B）是否在线。
    *   如果在线，则通过对应的 WebSocket 连接将通知实时推送到其客户端（浏览器）。
    *   如果用户不在线，则不做任何事。用户下次上线时，可以通过 API 拉取历史通知。

---

## 3. 数据库设计 (MongoDB)

在 MongoDB 中，创建一个名为 `notifications` 的集合。每个文档代表一条通知，结构如下：

```json
{
  "_id": "ObjectID",
  "recipient_id": "integer", // 接收通知的用户ID
  "sender_id": "integer",    // 触发通知的用户ID (可选)
  "sender_name": "string",   // 触发通知的用户名 (可选)
  "type": "string",          // 通知类型, e.g., "new_answer", "new_comment"
  "content": "string",       // 通知摘要内容, e.g., "张三 回答了你的问题: 如何学习Go语言?"
  "target_url": "string",    // 点击通知后跳转的URL, e.g., "/questions/123"
  "is_read": "boolean",      // 是否已读
  "created_at": "timestamp"
}
```

---

## 4. 核心组件实现 (Go)

项目结构将主要在 `internal/notification/` 目录下展开。

*   **`internal/notification/model/notification.go`**: 定义 `Notification` 的 Go 结构体，与 MongoDB 文档对应。
*   **`internal/notification/store/mongo_store.go`**: 数据访问层。
    *   `NewMongoStore(client *mongo.Client)`: 创建一个新的 Store 实例。
    *   `Create(ctx context.Context, notification *model.Notification) error`: 创建一条新通知。
    *   `GetByRecipientID(ctx context.Context, userID int64, limit, offset int) ([]*model.Notification, error)`: 分页获取用户的历史通知。
    *   `MarkAsRead(ctx context.Context, notificationID string, userID int64) error`: 将通知标记为已读。
*   **`internal/notification/service/hub.go`**: WebSocket 连接管理器。
    *   管理所有客户端的 WebSocket 连接，维护一个 `UserID -> WebSocket Connection` 的映射。
    *   提供 `Register`, `Unregister`, `SendToUser` 等方法。
*   **`internal/notification/handler/kafka_consumer.go`**: Kafka 消息处理器。
    *   订阅来自问答服务的事件主题。
    *   调用 `NotificationService` 来处理消息。
*   **`internal/notification/service/notification_service.go`**: 核心业务逻辑层。
    *   `HandleEvent(eventMessage []byte) error`: 解析 Kafka 消息，创建通知，存入数据库，并尝试通过 `Hub` 推送。
*   **`internal/notification/handler/gin_handler.go`**: HTTP 和 WebSocket 处理器。
    *   `wsHandler(c *gin.Context)`: 处理 WebSocket 升级请求，并将新连接注册到 `Hub`。
    *   `getNotifications(c *gin.Context)`: 提供 REST API，用于获取历史通知。
    *   `markNotificationAsRead(c *gin.Context)`: 提供 REST API，用于标记通知已读。
*   **`cmd/notification-service/main.go`**: 服务入口。
    *   初始化配置。
    *   连接 MongoDB。
    *   初始化并启动 Kafka 消费者。
    *   初始化 `Hub`。
    *   初始化 Gin 引擎并注册路由。
    *   启动服务。

---

## 5. 开发步骤

1.  **环境配置**: 在 `compose.yml` 中添加 MongoDB 服务。
2.  **定义模型**: 在 `internal/notification/model/` 中创建 `notification.go`。
3.  **实现 Store**: 编写 `mongo_store.go`，实现对 MongoDB 的增删改查。
4.  **实现 WebSocket Hub**: 编写 `hub.go` 来管理客户端连接。
5.  **实现 Service**: 编写 `notification_service.go`，处理核心业务逻辑。
6.  **实现 Kafka Consumer**: 编写 `kafka_consumer.go`，监听并处理上游消息。
7.  **实现 Gin Handler**: 编写 `gin_handler.go`，暴露 WebSocket 和 HTTP API。
8.  **组装服务**: 修改 `cmd/notification-service/main.go`，将所有组件串联起来并启动。
9.  **修改上游服务**: 在问答服务（qa-service）中，当有新回答、新评论等事件发生时，通过 Kafka Producer 发送消息。
10. **配置网关**: 修改 `nginx/nginx.conf`，为 WebSocket 连接 `/ws/notifications` 添加反向代理规则。
