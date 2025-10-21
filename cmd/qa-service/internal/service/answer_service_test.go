package service_test

import (
	"context"
	"errors"
	"testing"

	"qahub/pkg/auth"
	"qahub/pkg/config"
	"qahub/pkg/messaging"
	"qahub/qa-service/internal/model"
	"qahub/qa-service/internal/service"
	"qahub/qa-service/internal/store"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCreateAnswer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := service.NewMockQAStore(ctrl)
	producer := messaging.NewKafkaProducer(config.Conf.Kafka)
	qaService := service.NewQAService(mockStore, producer, &config.Conf)

	t.Run("成功创建回答", func(t *testing.T) {
		questionID := int64(1)
		content := "这是一个测试回答"
		userID := int64(100)
		username := "testuser"

		// 创建带有用户身份的 context
		identity := auth.Identity{
			UserID:   userID,
			Username: username,
		}
		ctx := auth.WithIdentity(context.Background(), identity)

		// Mock: 创建回答成功
		mockStore.EXPECT().
			CreateAnswer(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, a *model.Answer) (int64, error) {
				assert.Equal(t, questionID, a.QuestionID)
				assert.Equal(t, content, a.Content)
				assert.Equal(t, userID, a.UserID)
				return int64(200), nil
			}).
			Times(1)

		// Mock: 获取问题（用于通知）- 这是异步的，可能不会被调用
		mockStore.EXPECT().
			GetQuestionByID(gomock.Any(), questionID).
			Return(&model.Question{
				ID:     questionID,
				UserID: 999, // 不同的用户，会触发通知
			}, nil).
			AnyTimes()

		// 执行测试
		result, err := qaService.CreateAnswer(ctx, questionID, content, userID)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(200), result.ID)
		assert.Equal(t, questionID, result.QuestionID)
		assert.Equal(t, content, result.Content)
		assert.Equal(t, userID, result.UserID)
	})

	t.Run("缺少用户身份信息", func(t *testing.T) {
		ctx := context.Background() // 没有用户身份信息

		// 执行测试
		result, err := qaService.CreateAnswer(ctx, 1, "内容", 100)

		// 验证结果
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "user identity not found in context", err.Error())
	})

	t.Run("创建回答失败-数据库错误", func(t *testing.T) {
		questionID := int64(1)
		content := "测试内容"
		userID := int64(100)

		identity := auth.Identity{
			UserID:   userID,
			Username: "testuser",
		}
		ctx := auth.WithIdentity(context.Background(), identity)

		// Mock: 数据库错误
		mockStore.EXPECT().
			CreateAnswer(ctx, gomock.Any()).
			Return(int64(0), errors.New("database error")).
			Times(1)

		// 执行测试
		result, err := qaService.CreateAnswer(ctx, questionID, content, userID)

		// 验证结果
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "database error", err.Error())
	})
}

func TestGetAnswer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := service.NewMockQAStore(ctrl)
	producer := messaging.NewKafkaProducer(config.Conf.Kafka)
	qaService := service.NewQAService(mockStore, producer, &config.Conf)
	ctx := context.Background()

	t.Run("成功获取回答", func(t *testing.T) {
		answerID := int64(200)
		answer := &model.Answer{
			ID:         answerID,
			QuestionID: 1,
			Content:    "测试回答",
			UserID:     100,
		}

		// Mock: 获取回答
		mockStore.EXPECT().
			GetAnswerByID(ctx, answerID).
			Return(answer, nil).
			Times(1)

		// 执行测试
		result, err := qaService.GetAnswer(ctx, answerID)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, answerID, result.ID)
		assert.Equal(t, "测试回答", result.Content)
	})

	t.Run("回答不存在", func(t *testing.T) {
		answerID := int64(999)

		// Mock: 回答不存在
		mockStore.EXPECT().
			GetAnswerByID(ctx, answerID).
			Return(nil, errors.New("answer not found")).
			Times(1)

		// 执行测试
		result, err := qaService.GetAnswer(ctx, answerID)

		// 验证结果
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestListAnswers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := service.NewMockQAStore(ctrl)
	producer := messaging.NewKafkaProducer(config.Conf.Kafka)
	qaService := service.NewQAService(mockStore, producer, &config.Conf)
	ctx := context.Background()

	t.Run("成功获取回答列表", func(t *testing.T) {
		questionID := int64(1)
		userID := int64(100)
		page := int64(1)
		pageSize := int32(10)

		answers := []*model.Answer{
			{ID: 1, QuestionID: questionID, Content: "回答1", UserID: 100},
			{ID: 2, QuestionID: questionID, Content: "回答2", UserID: 101},
		}

		// Mock: 获取回答列表
		mockStore.EXPECT().
			ListAnswersByQuestionID(ctx, questionID, int64(0), pageSize).
			Return(answers, nil).
			Times(1)

		// Mock: 获取回答总数
		mockStore.EXPECT().
			CountAnswersByQuestionID(ctx, questionID).
			Return(int64(20), nil).
			Times(1)

		// Mock: 获取用户名
		mockStore.EXPECT().
			GetUsernamesByIDs(ctx, gomock.Any()).
			Return(map[int64]string{100: "user1", 101: "user2"}, nil).
			Times(1)

		// Mock: 获取用户投票信息
		mockStore.EXPECT().
			GetUserVotesForAnswers(ctx, userID, gomock.Any()).
			Return(map[int64]bool{1: true, 2: false}, nil).
			Times(1)

		// 执行测试
		results, total, err := qaService.ListAnswers(ctx, questionID, page, pageSize, userID)

		// 验证结果
		assert.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, int64(20), total)
		assert.Equal(t, "user1", results[0].Username)
		assert.True(t, results[0].IsUpvotedByUser)
	})

	t.Run("空回答列表", func(t *testing.T) {
		questionID := int64(1)
		userID := int64(100)
		page := int64(1)
		pageSize := int32(10)

		// Mock: 返回空列表
		mockStore.EXPECT().
			ListAnswersByQuestionID(ctx, questionID, int64(0), pageSize).
			Return([]*model.Answer{}, nil).
			Times(1)

		// Mock: 总数为0
		mockStore.EXPECT().
			CountAnswersByQuestionID(ctx, questionID).
			Return(int64(0), nil).
			Times(1)

		// 执行测试
		results, total, err := qaService.ListAnswers(ctx, questionID, page, pageSize, userID)

		// 验证结果
		assert.NoError(t, err)
		assert.Len(t, results, 0)
		assert.Equal(t, int64(0), total)
	})
}

func TestUpvoteAnswer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := service.NewMockQAStore(ctrl)
	producer := messaging.NewKafkaProducer(config.Conf.Kafka)
	qaService := service.NewQAService(mockStore, producer, &config.Conf)
	ctx := context.Background()

	t.Run("成功点赞回答", func(t *testing.T) {
		answerID := int64(200)
		userID := int64(100)

		// Mock: 执行事务
		mockStore.EXPECT().
			ExecTx(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, fn func(store.QAStore) error) error {
				// 模拟事务内的操作
				return fn(mockStore)
			}).
			Times(1)

		// Mock: 创建投票记录
		mockStore.EXPECT().
			CreateAnswerVote(ctx, answerID, userID, true).
			Return(nil).
			Times(1)

		// Mock: 增加点赞数
		mockStore.EXPECT().
			IncrementAnswerUpvoteCount(ctx, answerID).
			Return(nil).
			Times(1)

		// 执行测试
		err := qaService.UpvoteAnswer(ctx, answerID, userID)

		// 验证结果
		assert.NoError(t, err)
	})

	t.Run("点赞失败-事务错误", func(t *testing.T) {
		answerID := int64(200)
		userID := int64(100)

		// Mock: 事务失败
		mockStore.EXPECT().
			ExecTx(ctx, gomock.Any()).
			Return(errors.New("transaction error")).
			Times(1)

		// 执行测试
		err := qaService.UpvoteAnswer(ctx, answerID, userID)

		// 验证结果
		assert.Error(t, err)
		assert.Equal(t, "transaction error", err.Error())
	})
}

func TestDownvoteAnswer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := service.NewMockQAStore(ctrl)
	producer := messaging.NewKafkaProducer(config.Conf.Kafka)
	qaService := service.NewQAService(mockStore, producer, &config.Conf)
	ctx := context.Background()

	t.Run("成功取消点赞", func(t *testing.T) {
		answerID := int64(200)
		userID := int64(100)

		// Mock: 执行事务
		mockStore.EXPECT().
			ExecTx(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, fn func(store.QAStore) error) error {
				return fn(mockStore)
			}).
			Times(1)

		// Mock: 删除投票记录
		mockStore.EXPECT().
			DeleteAnswerVote(ctx, answerID, userID).
			Return(nil).
			Times(1)

		// Mock: 减少点赞数
		mockStore.EXPECT().
			DecrementAnswerUpvoteCount(ctx, answerID).
			Return(nil).
			Times(1)

		// 执行测试
		err := qaService.DownvoteAnswer(ctx, answerID, userID)

		// 验证结果
		assert.NoError(t, err)
	})
}

func TestUpdateAnswer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := service.NewMockQAStore(ctrl)
	producer := messaging.NewKafkaProducer(config.Conf.Kafka)
	qaService := service.NewQAService(mockStore, producer, &config.Conf)
	ctx := context.Background()

	t.Run("成功更新回答", func(t *testing.T) {
		answerID := int64(200)
		userID := int64(100)
		newContent := "更新后的内容"

		existingAnswer := &model.Answer{
			ID:         answerID,
			QuestionID: 1,
			Content:    "原内容",
			UserID:     userID,
		}

		// Mock: 获取回答
		mockStore.EXPECT().
			GetAnswerByID(ctx, answerID).
			Return(existingAnswer, nil).
			Times(1)

		// Mock: 更新回答
		mockStore.EXPECT().
			UpdateAnswer(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, a *model.Answer) error {
				assert.Equal(t, newContent, a.Content)
				assert.Equal(t, userID, a.UserID)
				return nil
			}).
			Times(1)

		// 执行测试
		result, err := qaService.UpdateAnswer(ctx, answerID, newContent, userID)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, newContent, result.Content)
	})

	t.Run("无权限更新回答", func(t *testing.T) {
		answerID := int64(200)
		userID := int64(100)
		otherUserID := int64(200)

		existingAnswer := &model.Answer{
			ID:         answerID,
			QuestionID: 1,
			Content:    "原内容",
			UserID:     userID,
		}

		// Mock: 获取回答
		mockStore.EXPECT().
			GetAnswerByID(ctx, answerID).
			Return(existingAnswer, nil).
			Times(1)

		// 执行测试 - 使用其他用户ID
		result, err := qaService.UpdateAnswer(ctx, answerID, "新内容", otherUserID)

		// 验证结果
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "无权限修改该回答", err.Error())
	})
}

func TestDeleteAnswer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := service.NewMockQAStore(ctrl)
	producer := messaging.NewKafkaProducer(config.Conf.Kafka)
	qaService := service.NewQAService(mockStore, producer, &config.Conf)
	ctx := context.Background()

	t.Run("成功删除回答", func(t *testing.T) {
		answerID := int64(200)
		userID := int64(100)

		answer := &model.Answer{
			ID:         answerID,
			QuestionID: 1,
			UserID:     userID,
		}

		// Mock: 获取回答
		mockStore.EXPECT().
			GetAnswerByID(ctx, answerID).
			Return(answer, nil).
			Times(1)

		// Mock: 删除回答
		mockStore.EXPECT().
			DeleteAnswer(ctx, answerID).
			Return(nil).
			Times(1)

		// 执行测试
		err := qaService.DeleteAnswer(ctx, answerID, userID)

		// 验证结果
		assert.NoError(t, err)
	})

	t.Run("无权限删除回答", func(t *testing.T) {
		answerID := int64(200)
		userID := int64(100)
		otherUserID := int64(200)

		answer := &model.Answer{
			ID:         answerID,
			QuestionID: 1,
			UserID:     userID,
		}

		// Mock: 获取回答
		mockStore.EXPECT().
			GetAnswerByID(ctx, answerID).
			Return(answer, nil).
			Times(1)

		// 执行测试 - 使用其他用户ID
		err := qaService.DeleteAnswer(ctx, answerID, otherUserID)

		// 验证结果
		assert.Error(t, err)
		assert.Equal(t, "无权限删除该回答", err.Error())
	})

	t.Run("回答不存在", func(t *testing.T) {
		answerID := int64(999)
		userID := int64(100)

		// Mock: 回答不存在
		mockStore.EXPECT().
			GetAnswerByID(ctx, answerID).
			Return(nil, errors.New("answer not found")).
			Times(1)

		// 执行测试
		err := qaService.DeleteAnswer(ctx, answerID, userID)

		// 验证结果
		assert.Error(t, err)
	})
}
