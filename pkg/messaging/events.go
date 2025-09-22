package messaging

import "time"

// EventType 是用于区分不同事件类型的字符串
type EventType string

const (
	// EventQuestionCreated 表示一个问题被创建的事件
	EventQuestionCreated EventType = "question.created"
	// EventQuestionUpdated 表示一个问题被更新的事件
	EventQuestionUpdated EventType = "question.updated"
	// EventAnswerCreated 表示一个回答被创建的事件
	EventAnswerCreated EventType = "answer.created"
)

// EventHeader 包含了所有事件共有的元数据
type EventHeader struct {
	ID        string    `json:"id"`         // 事件唯一ID
	Type      EventType `json:"type"`       // 事件类型
	Source    string    `json:"source"`     // 事件来源服务，例如 "qa-service"
	Timestamp time.Time `json:"timestamp"`  // 事件发生时间
}

// QuestionPayload 是与问题相关的事件所携带的数据
type QuestionPayload struct {
	ID      uint64   `json:"id"`
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Tags    []string `json:"tags,omitempty"`
	AuthorID uint64 `json:"author_id"`
}

// QuestionCreatedEvent 是问题创建事件的完整结构
type QuestionCreatedEvent struct {
	Header  EventHeader     `json:"header"`
	Payload QuestionPayload `json:"payload"`
}
