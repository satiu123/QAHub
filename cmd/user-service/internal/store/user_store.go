package store

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"

	_ "github.com/go-sql-driver/mysql" // MySQL driver

	"qahub/user-service/internal/model"
)

// TokenBlacklister 定义了 JWT 黑名单所需的方法
type TokenBlacklister interface {
	AddToBlacklist(token string, expiration time.Duration) error
	IsBlacklisted(token string) (bool, error)
}

type UserStore interface {
	CreateUser(ctx context.Context, user *model.User) (int64, error)
	GetUserByID(ctx context.Context, id int64) (*model.User, error)
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error
	DeleteUser(ctx context.Context, id int64) error
}

type mySQLUserStore struct {
	db *sqlx.DB // 存储数据库连接
}

func NewMySQLUserStore(db *sqlx.DB) UserStore {
	return &mySQLUserStore{db: db}
}

func (s *mySQLUserStore) CreateUser(ctx context.Context, user *model.User) (int64, error) {
	query := "INSERT INTO users (username, email,bio, password) VALUES (?, ?,?, ?)"
	result, err := s.db.ExecContext(ctx, query, user.Username, user.Email, user.Bio, user.Password)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *mySQLUserStore) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	var user model.User

	query := "SELECT id, username, email,bio, password FROM users WHERE id = ?"
	err := s.db.GetContext(ctx, &user, query, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *mySQLUserStore) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User

	query := "SELECT id, username, email,bio, password FROM users WHERE username = ?"
	err := s.db.GetContext(ctx, &user, query, username)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *mySQLUserStore) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User

	query := "SELECT id, username, email,bio, password FROM users WHERE email = ?"
	err := s.db.GetContext(ctx, &user, query, email)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *mySQLUserStore) UpdateUser(ctx context.Context, user *model.User) error {
	query := "UPDATE users SET username = ?, email = ?, bio = ? WHERE id = ?"
	_, err := s.db.ExecContext(ctx, query, user.Username, user.Email, user.Bio, user.ID)
	if err != nil {
		return err
	}
	return nil
}

func (s *mySQLUserStore) DeleteUser(ctx context.Context, id int64) error {
	query := "DELETE FROM users WHERE id = ?"
	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	return nil
}
