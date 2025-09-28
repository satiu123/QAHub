package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Notification 结构体定义了通知的存储模型，用于MongoDB
type Notification struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`                            // MongoDB的唯一标识符
	RecipientID int64              `bson:"recipient_id" json:"recipient_id"`                   // 接收者的用户ID
	SenderID    int64              `bson:"sender_id,omitempty" json:"sender_id,omitempty"`     // 发送者的用户ID，可选
	SenderName  string             `bson:"sender_name,omitempty" json:"sender_name,omitempty"` // 发送者的用户名，可选
	Type        string             `bson:"type" json:"type"`                                   // 通知类型，如 "comment", "like", "follow"
	Content     string             `bson:"content" json:"content"`                             // 通知内容
	TargetURL   string             `bson:"target_url" json:"target_url"`                       // 相关链接，如评论或帖子链接
	IsRead      bool               `bson:"is_read" json:"is_read"`                             // 是否已读
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`                       // 创建时间
}
