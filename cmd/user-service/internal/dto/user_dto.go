package dto

import "time"

// UserResponse 定义了API响应中返回的用户信息，不包含敏感数据。
type UserResponse struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Bio       string    `json:"bio,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
