package service

import (
	"errors"
	"time"

	"qahub/internal/user/model"
	"qahub/internal/user/store"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("satiu")

type UserService interface {
	// 定义用户服务接口方法
	Register(username, email, password string) (*model.User, error)
	Login(username, password string) (string, error)
	GetUserProfile(userID int64) (*model.User, error)
	UpdateUserProfile(user *model.User) error
	DeleteUser(userID int64) error
}

type userService struct {
	// 定义用户服务的依赖，例如存储接口
	userStore store.UserStore
}

func NewUserService(store store.UserStore) UserService {
	return &userService{userStore: store}
}

func (s *userService) Register(username, email, password string) (*model.User, error) {
	// 1. 验证输入
	if username == "" || email == "" || password == "" {
		return nil, errors.New("username, email, and password are required")
	}

	// 2. 检查邮箱是否已存在
	// 注意: GetUserByEmail 方法需要在 store 层中添加
	if existingUser, _ := s.userStore.GetUserByEmail(email); existingUser != nil {
		return nil, errors.New("email already exists")
	}

	// 3. 哈希密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 4. 创建用户记录
	newUser := &model.User{
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
	}

	// 5. 存入数据库
	newID, err := s.userStore.CreateUser(newUser)
	if err != nil {
		return nil, err
	}
	newUser.ID = newID

	// 6. 返回新创建的用户（不包含密码）
	newUser.Password = ""
	return newUser, nil
}

func (s *userService) Login(username, password string) (string, error) {
	// 1. 根据用户名查找用户
	user, err := s.userStore.GetUserByUsername(username)
	if err != nil {
		// 无论是数据库错误还是用户不存在，都返回相同的错误信息以防信息泄露
		return "", errors.New("invalid username or password")
	}

	// 2. 比较哈希密码和用户输入的密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		// 密码不匹配
		return "", errors.New("invalid username or password")
	}

	// 3. 密码正确，生成JWT
	// 创建载荷 (Claims)
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 72).Unix(), // 令牌有效期72小时
		"iat":      time.Now().Unix(),                     // 签发时间
	}

	// 创建令牌对象
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 使用密钥签名，获取完整的令牌字符串
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *userService) GetUserProfile(userID int64) (*model.User, error) {
	// 实现获取用户资料逻辑
	user, err := s.userStore.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	// 返回的用户信息不应包含密码
	user.Password = ""
	return user, nil
}

func (s *userService) UpdateUserProfile(user *model.User) error {
	// 实现更新用户资料逻辑
	// 注意：这里应该有权限检查，确保用户只能更新自己的资料
	// 密码更新应该有单独的接口和逻辑

	return s.userStore.UpdateUser(user)
}

func (s *userService) DeleteUser(userID int64) error {
	// 实现删除用户逻辑
	// 注意：这里应该有权限检查
	return s.userStore.DeleteUser(userID)
}
