package model

type User struct {
	ID       int64  `json:"id"`       // 唯一标识符
	Username string `json:"username"` // 不唯一
	Email    string `json:"email"`    // 唯一
	Password string `json:"-"`        // 密码哈希值，不应直接暴露
}
