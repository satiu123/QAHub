package dto

import (
	"time"

	"qahub/user-service/internal/model"
)

// UserResponse 定义了API响应中返回的用户信息,不包含敏感数据。
type UserResponse struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Bio       string    `json:"bio,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// FromUser 从 User 模型创建 UserResponse
func (r *UserResponse) FromUser(user *model.User) *UserResponse {
	r.ID = user.ID
	r.Username = user.Username
	r.Email = user.Email
	r.Bio = user.Bio
	r.CreatedAt = user.CreatedAt
	return r
}

// NewUserResponse 创建一个新的 UserResponse 从 User 模型
func NewUserResponse(user *model.User) *UserResponse {
	return &UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Bio:       user.Bio,
		CreatedAt: user.CreatedAt,
	}
}

// RegisterRequest 定义了用户注册请求的结构。
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=30"`
	Email    string `json:"email" binding:"required,email"`
	Bio      string `json:"bio,omitempty" binding:"max=500"`
	Password string `json:"password" binding:"required,min=6,max=128"`
}

// ToUser 将 RegisterRequest 转换为 User 模型
func (r *RegisterRequest) ToUser(hashedPassword string) *model.User {
	return &model.User{
		Username: r.Username,
		Email:    r.Email,
		Password: hashedPassword,
		Bio:      r.Bio,
	}
}
