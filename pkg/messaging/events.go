package messaging

import "time"

// EventType 是用于区分不同事件类型的字符串
type EventType string

const (
	// EventQuestionCreated 表示一个问题被创建的事件
	EventQuestionCreated EventType = "question.created"
	// EventQuestionUpdated 表示一个问题被更新的事件
	EventQuestionUpdated EventType = "question.updated"
	// EventQuestionDeleted 表示一个问题被删除的事件
	EventQuestionDeleted EventType = "question.deleted"
	// EventAnswerCreated 表示一个回答被创建的事件
	EventAnswerCreated EventType = "answer.created"
	// EventAnswerUpdated 表示一个回答被更新的事件
	EventAnswerUpdated EventType = "answer.updated"
	// EventAnswerDeleted 表示一个回答被删除的事件
	EventAnswerDeleted EventType = "answer.deleted"
	// EventAnswerDownvoted 表示一个回答被点踩的事件
	EventAnswerDownvoted EventType = "answer.downvoted"
	// EventCommentCreated 表示一个评论被创建的事件
	EventCommentCreated EventType = "comment.created"
	// EventCommentUpdated 表示一个评论被更新的事件
	EventCommentUpdated EventType = "comment.updated"
	// EventCommentDeleted 表示一个评论被删除的事件
	EventCommentDeleted EventType = "comment.deleted"
)

// EventHeader 包含了所有事件共有的元数据
type EventHeader struct {
	ID        string    `json:"id"`        // 事件唯一ID
	Type      EventType `json:"type"`      // 事件类型
	Source    string    `json:"source"`    // 事件来源服务，例如 "qa-service"
	Timestamp time.Time `json:"timestamp"` // 事件发生时间
}

// QuestionPayload 是与问题相关的事件所携带的数据
type QuestionPayload struct {
	ID         int64    `json:"id"`
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	Tags       []string `json:"tags,omitempty"`
	AuthorID   int64    `json:"author_id"`
	AuthorName string   `json:"author_name,omitempty"`
}

// QuestionCreatedEvent 是问题创建事件的完整结构
type QuestionCreatedEvent struct {
	Header  EventHeader     `json:"header"`
	Payload QuestionPayload `json:"payload"`
}

// QustionUpdatedEvent 是问题更新事件的完整结构
type QuestionUpdatedEvent struct {
	Header  EventHeader     `json:"header"`
	Payload QuestionPayload `json:"payload"`
}

// QuestionDeletedEvent 是问题删除事件的完整结构
type QuestionDeletedEvent struct {
	Header  EventHeader `json:"header"`
	Payload struct {
		ID int64 `json:"id"`
	} `json:"payload"`
}

// EventNotificationTriggered 表示一个通知被触发的事件
const EventNotificationTriggered EventType = "notification.triggered"

const (
	NotificationTypeNewAnswer  = "new_answer"
	NotificationTypeNewComment = "new_comment"
)

// NotificationPayload 是与通知相关的事件所携带的数据
type NotificationPayload struct {
	RecipientID      int64  `json:"recipient_id"` // 接收通知的用户ID
	SenderID         int64  `json:"sender_id"`
	SenderName       string `json:"sender_name"`
	NotificationType string `json:"notification_type"` // e.g., "new_answer", "new_comment"
	Content          string `json:"content"`           // 通知内容
	TargetURL        string `json:"target_url"`        // 点击通知后跳转的URL
}

// NotificationTriggeredEvent 是通知触发事件的完整结构
type NotificationTriggeredEvent struct {
	Header  EventHeader         `json:"header"`
	Payload NotificationPayload `json:"payload"`
}
