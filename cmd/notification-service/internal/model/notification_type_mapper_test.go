package model

import (
	"encoding/json"
	"testing"

	pb "qahub/api/proto/notification"

	"go.mongodb.org/mongo-driver/bson"
)

// TestNotificationTypeValue 测试 NotificationType.Value() 方法
func TestNotificationTypeValue(t *testing.T) {
	tests := []struct {
		name     string
		input    NotificationType
		expected string
	}{
		{
			name:     "NEWANSWER 类型",
			input:    NotificationType(pb.NotificationType_NEWANSWER),
			expected: "new_answer",
		},
		{
			name:     "NEWCOMMENT 类型",
			input:    NotificationType(pb.NotificationType_NEWCOMMENT),
			expected: "new_comment",
		},
		{
			name:     "未知类型（默认值）",
			input:    NotificationType(999),
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.Value()
			if result != tt.expected {
				t.Errorf("Value() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestNotificationTypeFromString 测试从字符串转换
func TestNotificationTypeFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected NotificationType
	}{
		{
			name:     "new_answer 字符串",
			input:    "new_answer",
			expected: NotificationType(pb.NotificationType_NEWANSWER),
		},
		{
			name:     "new_comment 字符串",
			input:    "new_comment",
			expected: NotificationType(pb.NotificationType_NEWCOMMENT),
		},
		{
			name:     "未知字符串（返回默认值）",
			input:    "unknown_type",
			expected: NotificationType(pb.NotificationType_NEWANSWER),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NotificationTypeFromString(tt.input)
			if result != tt.expected {
				t.Errorf("NotificationTypeFromString(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestNotificationTypeFromProto 测试从 protobuf 转换
func TestNotificationTypeFromProto(t *testing.T) {
	tests := []struct {
		name     string
		input    pb.NotificationType
		expected NotificationType
	}{
		{
			name:     "NEWANSWER proto",
			input:    pb.NotificationType_NEWANSWER,
			expected: NotificationType(pb.NotificationType_NEWANSWER),
		},
		{
			name:     "NEWCOMMENT proto",
			input:    pb.NotificationType_NEWCOMMENT,
			expected: NotificationType(pb.NotificationType_NEWCOMMENT),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NotificationTypeFromProto(tt.input)
			if result != tt.expected {
				t.Errorf("NotificationTypeFromProto(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestNotificationTypeToProto 测试转换到 protobuf
func TestNotificationTypeToProto(t *testing.T) {
	tests := []struct {
		name     string
		input    NotificationType
		expected pb.NotificationType
	}{
		{
			name:     "转换到 NEWANSWER",
			input:    NotificationType(pb.NotificationType_NEWANSWER),
			expected: pb.NotificationType_NEWANSWER,
		},
		{
			name:     "转换到 NEWCOMMENT",
			input:    NotificationType(pb.NotificationType_NEWCOMMENT),
			expected: pb.NotificationType_NEWCOMMENT,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.ToProto()
			if result != tt.expected {
				t.Errorf("ToProto() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestNotificationTypeJSONSerialization 测试 JSON 序列化和反序列化
func TestNotificationTypeJSONSerialization(t *testing.T) {
	tests := []struct {
		name         string
		input        NotificationType
		expectedJSON string
	}{
		{
			name:         "NEWANSWER 序列化",
			input:        NotificationType(pb.NotificationType_NEWANSWER),
			expectedJSON: `"new_answer"`,
		},
		{
			name:         "NEWCOMMENT 序列化",
			input:        NotificationType(pb.NotificationType_NEWCOMMENT),
			expectedJSON: `"new_comment"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试序列化
			jsonData, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("MarshalJSON() error = %v", err)
			}
			if string(jsonData) != tt.expectedJSON {
				t.Errorf("MarshalJSON() = %v, want %v", string(jsonData), tt.expectedJSON)
			}

			// 测试反序列化
			var result NotificationType
			err = json.Unmarshal(jsonData, &result)
			if err != nil {
				t.Fatalf("UnmarshalJSON() error = %v", err)
			}
			if result != tt.input {
				t.Errorf("UnmarshalJSON() = %v, want %v", result, tt.input)
			}
		})
	}
}

// TestNotificationTypeBSONSerialization 测试 BSON 序列化和反序列化
func TestNotificationTypeBSONSerialization(t *testing.T) {
	tests := []struct {
		name  string
		input NotificationType
	}{
		{
			name:  "NEWANSWER BSON",
			input: NotificationType(pb.NotificationType_NEWANSWER),
		},
		{
			name:  "NEWCOMMENT BSON",
			input: NotificationType(pb.NotificationType_NEWCOMMENT),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试序列化
			bsonType, data, err := tt.input.MarshalBSONValue()
			if err != nil {
				t.Fatalf("MarshalBSONValue() error = %v", err)
			}

			// 测试反序列化
			var result NotificationType
			err = result.UnmarshalBSONValue(bsonType, data)
			if err != nil {
				t.Fatalf("UnmarshalBSONValue() error = %v", err)
			}
			if result != tt.input {
				t.Errorf("UnmarshalBSONValue() = %v, want %v", result, tt.input)
			}
		})
	}
}

// TestNotificationTypeString 测试 String() 方法
func TestNotificationTypeString(t *testing.T) {
	tests := []struct {
		name     string
		input    NotificationType
		expected string
	}{
		{
			name:     "NEWANSWER 字符串",
			input:    NotificationType(pb.NotificationType_NEWANSWER),
			expected: "new_answer",
		},
		{
			name:     "NEWCOMMENT 字符串",
			input:    NotificationType(pb.NotificationType_NEWCOMMENT),
			expected: "new_comment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.String()
			if result != tt.expected {
				t.Errorf("String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestNotificationStatusValue 测试 NotificationStatus.Value() 方法
func TestNotificationStatusValue(t *testing.T) {
	tests := []struct {
		name     string
		input    NotificationStatus
		expected string
	}{
		{
			name:     "UNREAD 状态",
			input:    NotificationStatus(pb.NotificationStatus_UNREAD),
			expected: "unread",
		},
		{
			name:     "READ 状态",
			input:    NotificationStatus(pb.NotificationStatus_READ),
			expected: "read",
		},
		{
			name:     "未知状态（默认值）",
			input:    NotificationStatus(999),
			expected: "unread",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.Value()
			if result != tt.expected {
				t.Errorf("Value() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestNotificationStatusFromString 测试从字符串转换
func TestNotificationStatusFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected NotificationStatus
	}{
		{
			name:     "read 字符串",
			input:    "read",
			expected: NotificationStatus(pb.NotificationStatus_READ),
		},
		{
			name:     "unread 字符串",
			input:    "unread",
			expected: NotificationStatus(pb.NotificationStatus_UNREAD),
		},
		{
			name:     "未知字符串（返回默认值）",
			input:    "unknown_status",
			expected: NotificationStatus(pb.NotificationStatus_UNREAD),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NotificationStatusFromString(tt.input)
			if result != tt.expected {
				t.Errorf("NotificationStatusFromString(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestNotificationStatusFromProto 测试从 protobuf 转换
func TestNotificationStatusFromProto(t *testing.T) {
	tests := []struct {
		name     string
		input    pb.NotificationStatus
		expected NotificationStatus
	}{
		{
			name:     "UNREAD proto",
			input:    pb.NotificationStatus_UNREAD,
			expected: NotificationStatus(pb.NotificationStatus_UNREAD),
		},
		{
			name:     "READ proto",
			input:    pb.NotificationStatus_READ,
			expected: NotificationStatus(pb.NotificationStatus_READ),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NotificationStatusFromProto(tt.input)
			if result != tt.expected {
				t.Errorf("NotificationStatusFromProto(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestNotificationStatusToProto 测试转换到 protobuf
func TestNotificationStatusToProto(t *testing.T) {
	tests := []struct {
		name     string
		input    NotificationStatus
		expected pb.NotificationStatus
	}{
		{
			name:     "转换到 UNREAD",
			input:    NotificationStatus(pb.NotificationStatus_UNREAD),
			expected: pb.NotificationStatus_UNREAD,
		},
		{
			name:     "转换到 READ",
			input:    NotificationStatus(pb.NotificationStatus_READ),
			expected: pb.NotificationStatus_READ,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.ToProto()
			if result != tt.expected {
				t.Errorf("ToProto() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestNotificationStatusJSONSerialization 测试 JSON 序列化和反序列化
func TestNotificationStatusJSONSerialization(t *testing.T) {
	tests := []struct {
		name         string
		input        NotificationStatus
		expectedJSON string
	}{
		{
			name:         "UNREAD 序列化",
			input:        NotificationStatus(pb.NotificationStatus_UNREAD),
			expectedJSON: `"unread"`,
		},
		{
			name:         "READ 序列化",
			input:        NotificationStatus(pb.NotificationStatus_READ),
			expectedJSON: `"read"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试序列化
			jsonData, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("MarshalJSON() error = %v", err)
			}
			if string(jsonData) != tt.expectedJSON {
				t.Errorf("MarshalJSON() = %v, want %v", string(jsonData), tt.expectedJSON)
			}

			// 测试反序列化
			var result NotificationStatus
			err = json.Unmarshal(jsonData, &result)
			if err != nil {
				t.Fatalf("UnmarshalJSON() error = %v", err)
			}
			if result != tt.input {
				t.Errorf("UnmarshalJSON() = %v, want %v", result, tt.input)
			}
		})
	}
}

// TestNotificationStatusBSONSerialization 测试 BSON 序列化和反序列化
func TestNotificationStatusBSONSerialization(t *testing.T) {
	tests := []struct {
		name  string
		input NotificationStatus
	}{
		{
			name:  "UNREAD BSON",
			input: NotificationStatus(pb.NotificationStatus_UNREAD),
		},
		{
			name:  "READ BSON",
			input: NotificationStatus(pb.NotificationStatus_READ),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试序列化
			bsonType, data, err := tt.input.MarshalBSONValue()
			if err != nil {
				t.Fatalf("MarshalBSONValue() error = %v", err)
			}

			// 测试反序列化
			var result NotificationStatus
			err = result.UnmarshalBSONValue(bsonType, data)
			if err != nil {
				t.Fatalf("UnmarshalBSONValue() error = %v", err)
			}
			if result != tt.input {
				t.Errorf("UnmarshalBSONValue() = %v, want %v", result, tt.input)
			}
		})
	}
}

// TestNotificationStatusString 测试 String() 方法
func TestNotificationStatusString(t *testing.T) {
	tests := []struct {
		name     string
		input    NotificationStatus
		expected string
	}{
		{
			name:     "UNREAD 字符串",
			input:    NotificationStatus(pb.NotificationStatus_UNREAD),
			expected: "unread",
		},
		{
			name:     "READ 字符串",
			input:    NotificationStatus(pb.NotificationStatus_READ),
			expected: "read",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.String()
			if result != tt.expected {
				t.Errorf("String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestNotificationTypeRoundTrip 测试完整的往返转换
func TestNotificationTypeRoundTrip(t *testing.T) {
	original := NotificationType(pb.NotificationType_NEWANSWER)

	// Proto -> NotificationType -> String -> NotificationType -> Proto
	proto := original.ToProto()
	fromProto := NotificationTypeFromProto(proto)
	str := fromProto.Value()
	fromString := NotificationTypeFromString(str)
	finalProto := fromString.ToProto()

	if finalProto != proto {
		t.Errorf("Round trip failed: got %v, want %v", finalProto, proto)
	}
}

// TestNotificationStatusRoundTrip 测试完整的往返转换
func TestNotificationStatusRoundTrip(t *testing.T) {
	original := NotificationStatus(pb.NotificationStatus_READ)

	// Proto -> NotificationStatus -> String -> NotificationStatus -> Proto
	proto := original.ToProto()
	fromProto := NotificationStatusFromProto(proto)
	str := fromProto.Value()
	fromString := NotificationStatusFromString(str)
	finalProto := fromString.ToProto()

	if finalProto != proto {
		t.Errorf("Round trip failed: got %v, want %v", finalProto, proto)
	}
}

// TestNotificationWithBSON 测试在完整结构中的 BSON 序列化
func TestNotificationWithBSON(t *testing.T) {
	notification := &Notification{
		RecipientID: 123,
		SenderID:    456,
		SenderName:  "test_user",
		Type:        NotificationType(pb.NotificationType_NEWANSWER),
		Content:     "Test notification",
		TargetURL:   "/test/url",
		Status:      NotificationStatus(pb.NotificationStatus_UNREAD),
	}

	// 序列化
	data, err := bson.Marshal(notification)
	if err != nil {
		t.Fatalf("bson.Marshal() error = %v", err)
	}

	// 反序列化
	var result Notification
	err = bson.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("bson.Unmarshal() error = %v", err)
	}

	// 验证
	if result.Type != notification.Type {
		t.Errorf("Type mismatch: got %v, want %v", result.Type, notification.Type)
	}
	if result.Status != notification.Status {
		t.Errorf("Status mismatch: got %v, want %v", result.Status, notification.Status)
	}
	if result.RecipientID != notification.RecipientID {
		t.Errorf("RecipientID mismatch: got %v, want %v", result.RecipientID, notification.RecipientID)
	}
}

// TestNotificationWithJSON 测试在完整结构中的 JSON 序列化
func TestNotificationWithJSON(t *testing.T) {
	notification := &Notification{
		RecipientID: 123,
		SenderID:    456,
		SenderName:  "test_user",
		Type:        NotificationType(pb.NotificationType_NEWCOMMENT),
		Content:     "Test notification",
		TargetURL:   "/test/url",
		Status:      NotificationStatus(pb.NotificationStatus_READ),
	}

	// 序列化
	data, err := json.Marshal(notification)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// 反序列化
	var result Notification
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	// 验证
	if result.Type != notification.Type {
		t.Errorf("Type mismatch: got %v, want %v", result.Type, notification.Type)
	}
	if result.Status != notification.Status {
		t.Errorf("Status mismatch: got %v, want %v", result.Status, notification.Status)
	}
	if result.Content != notification.Content {
		t.Errorf("Content mismatch: got %v, want %v", result.Content, notification.Content)
	}
}
