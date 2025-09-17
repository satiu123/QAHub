package store

import (
	"github.com/jmoiron/sqlx"

	_ "github.com/go-sql-driver/mysql" // MySQL driver

	"qahub/internal/user/model"
)

type UserStore interface {
	CreateUser(user *model.User) (int64, error)
	GetUserByID(id int64) (*model.User, error)
	GetUserByUsername(username string) (*model.User, error)
	GetUserByEmail(email string) (*model.User, error)
	UpdateUser(user *model.User) error
	DeleteUser(id int64) error
}

type mySQLUserStore struct {
	db *sqlx.DB // 存储数据库连接
}

func NewMySQLUserStore(db *sqlx.DB) UserStore {
	return &mySQLUserStore{db: db}
}

func (s *mySQLUserStore) CreateUser(user *model.User) (int64, error) {
	query := "INSERT INTO users (username, email, password) VALUES (?, ?, ?)"
	result, err := s.db.Exec(query, user.Username, user.Email, user.Password)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *mySQLUserStore) GetUserByID(id int64) (*model.User, error) {
	var user model.User

	query := "SELECT id, username, email, password FROM users WHERE id = ?"
	err := s.db.Get(&user, query, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *mySQLUserStore) GetUserByUsername(username string) (*model.User, error) {
	var user model.User

	query := "SELECT id, username, email, password FROM users WHERE username = ?"
	err := s.db.Get(&user, query, username)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *mySQLUserStore) GetUserByEmail(email string) (*model.User, error) {
	var user model.User

	query := "SELECT id, username, email, password FROM users WHERE email = ?"
	err := s.db.Get(&user, query, email)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *mySQLUserStore) UpdateUser(user *model.User) error {
	query := "UPDATE users SET username = ?, email = ?, password = ? WHERE id = ?"
	_, err := s.db.Exec(query, user.Username, user.Email, user.Password, user.ID)
	if err != nil {
		return err
	}
	return nil
}

func (s *mySQLUserStore) DeleteUser(id int64) error {
	query := "DELETE FROM users WHERE id = ?"
	_, err := s.db.Exec(query, id)
	if err != nil {
		return err
	}
	return nil
}
