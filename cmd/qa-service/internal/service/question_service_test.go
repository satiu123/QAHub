//go:generate mockgen -source=../store/qa_store.go -destination=qa_store_mock.go -package=service

package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"qahub/pkg/config"
	"qahub/qa-service/internal/model"
	"qahub/qa-service/internal/service"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCreateQuestion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := service.NewMockQAStore(ctrl)
	qaService := service.NewQAService(mockStore, config.Kafka{})
	ctx := context.Background()

	t.Run("成功创建问题", func(t *testing.T) {
		title := "测试问题标题"
		content := "测试问题内容"
		userID := int64(1)

		// Mock: 创建问题成功，返回新问题ID
		mockStore.EXPECT().
			CreateQuestion(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, q *model.Question) (int64, error) {
				// 验证传入的问题数据
				assert.Equal(t, title, q.Title)
				assert.Equal(t, content, q.Content)
				assert.Equal(t, userID, q.UserID)
				return int64(100), nil
			}).
			Times(1)

		// 执行测试
		result, err := qaService.CreateQuestion(ctx, title, content, userID)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(100), result.ID)
		assert.Equal(t, title, result.Title)
		assert.Equal(t, content, result.Content)
		assert.Equal(t, userID, result.UserID)
	})

	t.Run("创建问题失败-数据库错误", func(t *testing.T) {
		title := "测试问题"
		content := "测试内容"
		userID := int64(1)

		// Mock: 数据库错误
		mockStore.EXPECT().
			CreateQuestion(ctx, gomock.Any()).
			Return(int64(0), errors.New("database error")).
			Times(1)

		// 执行测试
		result, err := qaService.CreateQuestion(ctx, title, content, userID)

		// 验证结果
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "database error", err.Error())
	})
}

func TestGetQuestion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := service.NewMockQAStore(ctrl)
	qaService := service.NewQAService(mockStore, config.Kafka{})
	ctx := context.Background()

	t.Run("成功获取问题详情", func(t *testing.T) {
		questionID := int64(1)
		question := &model.Question{
			ID:        questionID,
			Title:     "测试问题",
			Content:   "测试内容",
			UserID:    100,
			CreatedAt: time.Now(),
		}

		// Mock: 获取问题
		mockStore.EXPECT().
			GetQuestionByID(ctx, questionID).
			Return(question, nil).
			Times(1)

		// Mock: 获取用户名
		mockStore.EXPECT().
			GetUsernamesByIDs(ctx, []int64{100}).
			Return(map[int64]string{100: "testuser"}, nil).
			Times(1)

		// Mock: 获取回答数量
		mockStore.EXPECT().
			GetAnswerCountByQuestionIDs(ctx, []int64{questionID}).
			Return(map[int64]int64{questionID: 5}, nil).
			Times(1)

		// 执行测试
		result, err := qaService.GetQuestion(ctx, questionID)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, questionID, result.ID)
		assert.Equal(t, "测试问题", result.Title)
		assert.Equal(t, "testuser", result.AuthorName)
		assert.Equal(t, int64(5), result.AnswerCount)
	})

	t.Run("问题不存在", func(t *testing.T) {
		questionID := int64(999)

		// Mock: 问题不存在
		mockStore.EXPECT().
			GetQuestionByID(ctx, questionID).
			Return(nil, nil).
			Times(1)

		// 执行测试
		result, err := qaService.GetQuestion(ctx, questionID)

		// 验证结果
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "问题未找到", err.Error())
	})

	t.Run("数据库错误", func(t *testing.T) {
		questionID := int64(1)

		// Mock: 数据库错误
		mockStore.EXPECT().
			GetQuestionByID(ctx, questionID).
			Return(nil, errors.New("database error")).
			Times(1)

		// 执行测试
		result, err := qaService.GetQuestion(ctx, questionID)

		// 验证结果
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestListQuestions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := service.NewMockQAStore(ctrl)
	qaService := service.NewQAService(mockStore, config.Kafka{})
	ctx := context.Background()

	t.Run("成功获取问题列表", func(t *testing.T) {
		page := int64(1)
		pageSize := int32(10)

		questions := []*model.Question{
			{ID: 1, Title: "问题1", Content: "内容1", UserID: 100},
			{ID: 2, Title: "问题2", Content: "内容2", UserID: 101},
		}

		// Mock: 获取问题列表
		mockStore.EXPECT().
			ListQuestions(ctx, int64(0), pageSize).
			Return(questions, nil).
			Times(1)

		// Mock: 获取问题总数
		mockStore.EXPECT().
			CountQuestions(ctx).
			Return(int64(20), nil).
			Times(1)

		// Mock: 获取用户名 (使用 gomock.Any() 因为 map 遍历顺序不确定)
		mockStore.EXPECT().
			GetUsernamesByIDs(ctx, gomock.Any()).
			Return(map[int64]string{100: "user1", 101: "user2"}, nil).
			Times(1)

		// Mock: 获取回答数量 (使用 gomock.Any() 因为切片可能按任意顺序)
		mockStore.EXPECT().
			GetAnswerCountByQuestionIDs(ctx, gomock.Any()).
			Return(map[int64]int64{1: 3, 2: 5}, nil).
			Times(1)

		// 执行测试
		results, total, err := qaService.ListQuestions(ctx, page, pageSize)

		// 验证结果
		assert.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, int64(20), total)
		assert.Equal(t, "user1", results[0].AuthorName)
		assert.Equal(t, int64(3), results[0].AnswerCount)
	})

	t.Run("空列表", func(t *testing.T) {
		page := int64(1)
		pageSize := int32(10)

		// Mock: 返回空列表
		mockStore.EXPECT().
			ListQuestions(ctx, int64(0), pageSize).
			Return([]*model.Question{}, nil).
			Times(1)

		// Mock: 总数为0
		mockStore.EXPECT().
			CountQuestions(ctx).
			Return(int64(0), nil).
			Times(1)

		// 执行测试
		results, total, err := qaService.ListQuestions(ctx, page, pageSize)

		// 验证结果
		assert.NoError(t, err)
		assert.Len(t, results, 0)
		assert.Equal(t, int64(0), total)
	})
}

func TestUpdateQuestion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := service.NewMockQAStore(ctrl)
	qaService := service.NewQAService(mockStore, config.Kafka{})
	ctx := context.Background()

	t.Run("成功更新问题", func(t *testing.T) {
		questionID := int64(1)
		userID := int64(100)
		newTitle := "更新后的标题"
		newContent := "更新后的内容"

		existingQuestion := &model.Question{
			ID:      questionID,
			Title:   "原标题",
			Content: "原内容",
			UserID:  userID,
		}

		// Mock: 获取问题
		mockStore.EXPECT().
			GetQuestionByID(ctx, questionID).
			Return(existingQuestion, nil).
			Times(1)

		// Mock: 更新问题
		mockStore.EXPECT().
			UpdateQuestion(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, q *model.Question) error {
				assert.Equal(t, newTitle, q.Title)
				assert.Equal(t, newContent, q.Content)
				assert.Equal(t, userID, q.UserID)
				return nil
			}).
			Times(1)

		// 执行测试
		result, err := qaService.UpdateQuestion(ctx, questionID, newTitle, newContent, userID)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, newTitle, result.Title)
		assert.Equal(t, newContent, result.Content)
	})

	t.Run("无权限更新问题", func(t *testing.T) {
		questionID := int64(1)
		userID := int64(100)
		otherUserID := int64(200)

		existingQuestion := &model.Question{
			ID:      questionID,
			Title:   "原标题",
			Content: "原内容",
			UserID:  userID, // 属于其他用户
		}

		// Mock: 获取问题
		mockStore.EXPECT().
			GetQuestionByID(ctx, questionID).
			Return(existingQuestion, nil).
			Times(1)

		// 执行测试 - 尝试用其他用户ID更新
		result, err := qaService.UpdateQuestion(ctx, questionID, "新标题", "新内容", otherUserID)

		// 验证结果
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "无权限修改该问题", err.Error())
	})

	t.Run("问题不存在", func(t *testing.T) {
		questionID := int64(999)
		userID := int64(100)

		// Mock: 问题不存在
		mockStore.EXPECT().
			GetQuestionByID(ctx, questionID).
			Return(nil, errors.New("question not found")).
			Times(1)

		// 执行测试
		result, err := qaService.UpdateQuestion(ctx, questionID, "标题", "内容", userID)

		// 验证结果
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestDeleteQuestion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := service.NewMockQAStore(ctrl)
	qaService := service.NewQAService(mockStore, config.Kafka{})
	ctx := context.Background()

	t.Run("成功删除问题", func(t *testing.T) {
		questionID := int64(1)
		userID := int64(100)

		question := &model.Question{
			ID:     questionID,
			UserID: userID,
		}

		// Mock: 获取问题
		mockStore.EXPECT().
			GetQuestionByID(ctx, questionID).
			Return(question, nil).
			Times(1)

		// Mock: 删除问题
		mockStore.EXPECT().
			DeleteQuestion(ctx, questionID).
			Return(nil).
			Times(1)

		// 执行测试
		err := qaService.DeleteQuestion(ctx, questionID, userID)

		// 验证结果
		assert.NoError(t, err)
	})

	t.Run("无权限删除问题", func(t *testing.T) {
		questionID := int64(1)
		userID := int64(100)
		otherUserID := int64(200)

		question := &model.Question{
			ID:     questionID,
			UserID: userID,
		}

		// Mock: 获取问题
		mockStore.EXPECT().
			GetQuestionByID(ctx, questionID).
			Return(question, nil).
			Times(1)

		// 执行测试 - 尝试用其他用户ID删除
		err := qaService.DeleteQuestion(ctx, questionID, otherUserID)

		// 验证结果
		assert.Error(t, err)
		assert.Equal(t, "无权限删除该问题", err.Error())
	})

	t.Run("问题不存在", func(t *testing.T) {
		questionID := int64(999)
		userID := int64(100)

		// Mock: 问题不存在
		mockStore.EXPECT().
			GetQuestionByID(ctx, questionID).
			Return(nil, errors.New("question not found")).
			Times(1)

		// 执行测试
		err := qaService.DeleteQuestion(ctx, questionID, userID)

		// 验证结果
		assert.Error(t, err)
	})
}
