package model

import (
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Notification 结构体定义了通知的存储模型，用于MongoDB
type Notification struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"-" db:"id"`                                      // MongoDB的唯一标识符
	RecipientID int64              `bson:"recipient_id" json:"recipient_id" db:"recipient_id"`                  // 接收者的用户ID
	SenderID    int64              `bson:"sender_id,omitempty" json:"sender_id,omitempty" db:"sender_id"`       // 发送者的用户ID，可选
	SenderName  string             `bson:"sender_name,omitempty" json:"sender_name,omitempty" db:"sender_name"` // 发送者的用户名，可选
	Type        NotificationType   `bson:"type" json:"type" db:"notification_type"`                             // 通知类型，如 "new_answer", "new_comment"
	Content     string             `bson:"content" json:"content"`                                              // 通知内容
	TargetURL   string             `bson:"target_url" json:"target_url"`                                        // 相关链接，如评论或帖子链接
	Status      NotificationStatus `bson:"status" json:"status" db:"status"`                                    // 通知状态，如 "unread", "read"
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`                                        // 创建时间
}

// MarshalJSON 自定义JSON序列化，将ObjectID转换为字符串
func (n *Notification) MarshalJSON() ([]byte, error) {
	type Alias Notification
	return json.Marshal(&struct {
		ID string `json:"id"`
		*Alias
	}{
		ID:    n.ID.Hex(),
		Alias: (*Alias)(n),
	})
}
