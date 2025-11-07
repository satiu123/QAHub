package model

import (
	"encoding/json"
	"fmt"

	pb "qahub/api/proto/notification"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

type NotificationType pb.NotificationType
type NotificationStatus pb.NotificationStatus

// NotificationType 数据库映射方法

// Value 将 NotificationType 转换为数据库存储的字符串值
func (t NotificationType) Value() string {
	switch t {
	case NotificationType(pb.NotificationType_NEWANSWER):
		return "new_answer"
	case NotificationType(pb.NotificationType_NEWCOMMENT):
		return "new_comment"
	default:
		return "unknown"
	}
}

// ToProto 将 NotificationType 转换为 protobuf 类型
func (t NotificationType) ToProto() pb.NotificationType {
	return pb.NotificationType(t)
}

// NotificationTypeFromString 从数据库字符串值转换为 NotificationType
func NotificationTypeFromString(s string) NotificationType {
	switch s {
	case "new_answer":
		return NotificationType(pb.NotificationType_NEWANSWER)
	case "new_comment":
		return NotificationType(pb.NotificationType_NEWCOMMENT)
	default:
		return NotificationType(pb.NotificationType_NEWANSWER) // 默认值
	}
}

// NotificationTypeFromProto 从 protobuf 类型转换为 NotificationType
func NotificationTypeFromProto(t pb.NotificationType) NotificationType {
	return NotificationType(t)
}

// NotificationStatus 数据库映射方法

// Value 将 NotificationStatus 转换为数据库存储的字符串值
func (s NotificationStatus) Value() string {
	switch s {
	case NotificationStatus(pb.NotificationStatus_UNREAD):
		return "unread"
	case NotificationStatus(pb.NotificationStatus_READ):
		return "read"
	default:
		return "unread"
	}
}

// ToProto 将 NotificationStatus 转换为 protobuf 类型
func (s NotificationStatus) ToProto() pb.NotificationStatus {
	return pb.NotificationStatus(s)
}

// NotificationStatusFromString 从数据库字符串值转换为 NotificationStatus
func NotificationStatusFromString(str string) NotificationStatus {
	switch str {
	case "read":
		return NotificationStatus(pb.NotificationStatus_READ)
	case "unread":
		return NotificationStatus(pb.NotificationStatus_UNREAD)
	default:
		return NotificationStatus(pb.NotificationStatus_UNREAD) // 默认值
	}
}

// NotificationStatusFromProto 从 protobuf 类型转换为 NotificationStatus
func NotificationStatusFromProto(s pb.NotificationStatus) NotificationStatus {
	return NotificationStatus(s)
}

// NotificationType BSON 序列化方法

// MarshalBSONValue 将 NotificationType 序列化为 BSON 字符串
func (t NotificationType) MarshalBSONValue() (bsontype.Type, []byte, error) {
	return bson.MarshalValue(t.Value())
}

// UnmarshalBSONValue 从 BSON 字符串反序列化 NotificationType
func (t *NotificationType) UnmarshalBSONValue(bsonType bsontype.Type, data []byte) error {
	var str string
	if err := bson.UnmarshalValue(bsonType, data, &str); err != nil {
		return err
	}
	*t = NotificationTypeFromString(str)
	return nil
}

// NotificationType JSON 序列化方法

// MarshalJSON 将 NotificationType 序列化为 JSON 字符串
func (t NotificationType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Value())
}

// UnmarshalJSON 从 JSON 字符串反序列化 NotificationType
func (t *NotificationType) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	*t = NotificationTypeFromString(str)
	return nil
}

// String 实现 Stringer 接口
func (t NotificationType) String() string {
	return t.Value()
}

// NotificationStatus BSON 序列化方法

// MarshalBSONValue 将 NotificationStatus 序列化为 BSON 字符串
func (s NotificationStatus) MarshalBSONValue() (bsontype.Type, []byte, error) {
	return bson.MarshalValue(s.Value())
}

// UnmarshalBSONValue 从 BSON 字符串反序列化 NotificationStatus
func (s *NotificationStatus) UnmarshalBSONValue(bsonType bsontype.Type, data []byte) error {
	var str string
	if err := bson.UnmarshalValue(bsonType, data, &str); err != nil {
		return err
	}
	*s = NotificationStatusFromString(str)
	return nil
}

// NotificationStatus JSON 序列化方法

// MarshalJSON 将 NotificationStatus 序列化为 JSON 字符串
func (s NotificationStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Value())
}

// UnmarshalJSON 从 JSON 字符串反序列化 NotificationStatus
func (s *NotificationStatus) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("unmarshal notification status: %w", err)
	}
	*s = NotificationStatusFromString(str)
	return nil
}

// String 实现 Stringer 接口
func (s NotificationStatus) String() string {
	return s.Value()
}
