package service_test

import (
	"context"
	"errors"
	"testing"

	"qahub/user-service/internal/dto"
	"qahub/user-service/internal/model"
	"qahub/user-service/internal/service"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestRegister(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := service.NewMockUserStore(ctrl)
	userService := service.NewUserService(mockStore)
	ctx := context.Background()

	t.Run("成功注册新用户", func(t *testing.T) {
		req := dto.RegisterRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Bio:      "Test bio",
			Password: "password123",
		}

		// Mock: 邮箱不存在
		mockStore.EXPECT().
			GetUserByEmail(ctx, req.Email).
			Return(nil, errors.New("user not found")).
			Times(1)

		// Mock: 创建用户成功，返回新用户ID
		mockStore.EXPECT().
			CreateUser(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, user *model.User) (int64, error) {
				// 验证传入的用户数据
				assert.Equal(t, req.Username, user.Username)
				assert.Equal(t, req.Email, user.Email)
				assert.Equal(t, req.Bio, user.Bio)
				// 验证密码已被哈希
				err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
				assert.NoError(t, err, "密码应该被正确哈希")
				return int64(1), nil
			}).
			Times(1)

		// 执行测试
		resp, err := userService.Register(ctx, req)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int64(1), resp.ID)
		assert.Equal(t, req.Username, resp.Username)
		assert.Equal(t, req.Email, resp.Email)
		assert.Equal(t, req.Bio, resp.Bio)
	})

	t.Run("邮箱已存在时注册失败", func(t *testing.T) {
		req := dto.RegisterRequest{
			Username: "testuser",
			Email:    "existing@example.com",
			Bio:      "Test bio",
			Password: "password123",
		}

		existingUser := &model.User{
			ID:       1,
			Username: "existinguser",
			Email:    req.Email,
		}

		// Mock: 邮箱已存在
		mockStore.EXPECT().
			GetUserByEmail(ctx, req.Email).
			Return(existingUser, nil).
			Times(1)

		// 执行测试
		resp, err := userService.Register(ctx, req)

		// 验证结果
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "该邮箱已被注册", err.Error())
	})

	t.Run("创建用户时数据库错误", func(t *testing.T) {
		req := dto.RegisterRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Bio:      "Test bio",
			Password: "password123",
		}

		// Mock: 邮箱不存在
		mockStore.EXPECT().
			GetUserByEmail(ctx, req.Email).
			Return(nil, errors.New("user not found")).
			Times(1)

		// Mock: 创建用户失败
		mockStore.EXPECT().
			CreateUser(ctx, gomock.Any()).
			Return(int64(0), errors.New("database error")).
			Times(1)

		// 执行测试
		resp, err := userService.Register(ctx, req)

		// 验证结果
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "database error", err.Error())
	})
}

func TestLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := service.NewMockUserStore(ctrl)
	userService := service.NewUserService(mockStore)
	ctx := context.Background()

	// 准备测试数据：哈希密码
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)

	t.Run("成功登录", func(t *testing.T) {
		username := "testuser"
		password := "correctpassword"

		user := &model.User{
			ID:       1,
			Username: username,
			Password: string(hashedPassword),
			Email:    "test@example.com",
		}

		// Mock: 根据用户名查找用户
		mockStore.EXPECT().
			GetUserByUsername(ctx, username).
			Return(user, nil).
			Times(1)

		// 执行测试
		token, err := userService.Login(ctx, username, password)

		// 验证结果
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("用户名不存在", func(t *testing.T) {
		username := "nonexistent"
		password := "password123"

		// Mock: 用户不存在
		mockStore.EXPECT().
			GetUserByUsername(ctx, username).
			Return(nil, errors.New("user not found")).
			Times(1)

		// 执行测试
		token, err := userService.Login(ctx, username, password)

		// 验证结果
		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Equal(t, "invalid username or password", err.Error())
	})

	t.Run("密码错误", func(t *testing.T) {
		username := "testuser"
		password := "wrongpassword"

		user := &model.User{
			ID:       1,
			Username: username,
			Password: string(hashedPassword),
			Email:    "test@example.com",
		}

		// Mock: 用户存在
		mockStore.EXPECT().
			GetUserByUsername(ctx, username).
			Return(user, nil).
			Times(1)

		// 执行测试
		token, err := userService.Login(ctx, username, password)

		// 验证结果
		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Equal(t, "invalid username or password", err.Error())
	})
}

func TestGetUserProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := service.NewMockUserStore(ctrl)
	userService := service.NewUserService(mockStore)
	ctx := context.Background()

	t.Run("成功获取用户信息", func(t *testing.T) {
		userID := int64(1)
		user := &model.User{
			ID:       userID,
			Username: "testuser",
			Email:    "test@example.com",
			Bio:      "Test bio",
			Password: "hashedpassword",
		}

		// Mock: 根据ID查找用户
		mockStore.EXPECT().
			GetUserByID(ctx, userID).
			Return(user, nil).
			Times(1)

		// 执行测试
		resp, err := userService.GetUserProfile(ctx, userID)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, userID, resp.ID)
		assert.Equal(t, "testuser", resp.Username)
		assert.Equal(t, "test@example.com", resp.Email)
		assert.Equal(t, "Test bio", resp.Bio)
	})

	t.Run("用户不存在", func(t *testing.T) {
		userID := int64(999)

		// Mock: 用户不存在
		mockStore.EXPECT().
			GetUserByID(ctx, userID).
			Return(nil, errors.New("user not found")).
			Times(1)

		// 执行测试
		resp, err := userService.GetUserProfile(ctx, userID)

		// 验证结果
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "user not found", err.Error())
	})

	t.Run("数据库错误", func(t *testing.T) {
		userID := int64(1)

		// Mock: 数据库错误
		mockStore.EXPECT().
			GetUserByID(ctx, userID).
			Return(nil, errors.New("database connection error")).
			Times(1)

		// 执行测试
		resp, err := userService.GetUserProfile(ctx, userID)

		// 验证结果
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestUpdateUserProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := service.NewMockUserStore(ctrl)
	userService := service.NewUserService(mockStore)
	ctx := context.Background()

	t.Run("成功更新用户信息", func(t *testing.T) {
		user := &model.User{
			ID:       1,
			Username: "testuser",
			Email:    "updated@example.com",
			Bio:      "Updated bio",
		}

		// Mock: 更新用户成功
		mockStore.EXPECT().
			UpdateUser(ctx, user).
			Return(nil).
			Times(1)

		// 执行测试
		err := userService.UpdateUserProfile(ctx, user)

		// 验证结果
		assert.NoError(t, err)
	})

	t.Run("更新用户失败", func(t *testing.T) {
		user := &model.User{
			ID:       1,
			Username: "testuser",
			Email:    "updated@example.com",
		}

		// Mock: 更新失败
		mockStore.EXPECT().
			UpdateUser(ctx, user).
			Return(errors.New("update failed")).
			Times(1)

		// 执行测试
		err := userService.UpdateUserProfile(ctx, user)

		// 验证结果
		assert.Error(t, err)
		assert.Equal(t, "update failed", err.Error())
	})

	t.Run("数据库连接错误", func(t *testing.T) {
		user := &model.User{
			ID:       1,
			Username: "testuser",
		}

		// Mock: 数据库连接错误
		mockStore.EXPECT().
			UpdateUser(ctx, user).
			Return(errors.New("database connection error")).
			Times(1)

		// 执行测试
		err := userService.UpdateUserProfile(ctx, user)

		// 验证结果
		assert.Error(t, err)
	})
}

func TestDeleteUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := service.NewMockUserStore(ctrl)
	userService := service.NewUserService(mockStore)
	ctx := context.Background()

	t.Run("成功删除用户", func(t *testing.T) {
		userID := int64(1)

		// Mock: 删除用户成功
		mockStore.EXPECT().
			DeleteUser(ctx, userID).
			Return(nil).
			Times(1)

		// 执行测试
		err := userService.DeleteUser(ctx, userID)

		// 验证结果
		assert.NoError(t, err)
	})

	t.Run("删除不存在的用户", func(t *testing.T) {
		userID := int64(999)

		// Mock: 用户不存在
		mockStore.EXPECT().
			DeleteUser(ctx, userID).
			Return(errors.New("user not found")).
			Times(1)

		// 执行测试
		err := userService.DeleteUser(ctx, userID)

		// 验证结果
		assert.Error(t, err)
		assert.Equal(t, "user not found", err.Error())
	})

	t.Run("数据库错误", func(t *testing.T) {
		userID := int64(1)

		// Mock: 数据库错误
		mockStore.EXPECT().
			DeleteUser(ctx, userID).
			Return(errors.New("database error")).
			Times(1)

		// 执行测试
		err := userService.DeleteUser(ctx, userID)

		// 验证结果
		assert.Error(t, err)
	})
}
