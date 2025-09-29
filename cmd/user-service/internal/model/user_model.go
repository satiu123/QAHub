package model

import "time"

// User 对应于数据库中的 users 表
type User struct {
	ID        int64     `db:"id"`
	Username  string    `db:"username"`
	Email     string    `db:"email"`
	Bio       string    `db:"bio"`
	Password  string    `db:"password"` // 在实际应用中应存储哈希值
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
