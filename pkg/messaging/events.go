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
	ID       uint64   `json:"id"`
	Title    string   `json:"title"`
	Content  string   `json:"content"`
	Tags     []string `json:"tags,omitempty"`
	AuthorID uint64   `json:"author_id"`
}

// QuestionCreatedEvent 是问题创建事件的完整结构
type QuestionCreatedEvent struct {
	Header  EventHeader     `json:"header"`
	Payload QuestionPayload `json:"payload"`
}
