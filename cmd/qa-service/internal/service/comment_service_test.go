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

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCreateComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := service.NewMockQAStore(ctrl)
	producer := messaging.NewKafkaProducer(config.Conf.Kafka)
	qaService := service.NewQAService(mockStore, producer, &config.Conf)

	t.Run("成功创建评论", func(t *testing.T) {
		answerID := int64(200)
		content := "这是一个测试评论"
		userID := int64(100)
		username := "testuser"

		// 创建带有用户身份的 context
		identity := auth.Identity{
			UserID:   userID,
			Username: username,
		}
		ctx := auth.WithIdentity(context.Background(), identity)

		// Mock: 创建评论成功
		mockStore.EXPECT().
			CreateComment(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, c *model.Comment) (int64, error) {
				assert.Equal(t, answerID, c.AnswerID)
				assert.Equal(t, content, c.Content)
				assert.Equal(t, userID, c.UserID)
				return int64(300), nil
			}).
			Times(1)

		// Mock: 获取回答（用于通知）- 这是异步的，可能不会被调用
		mockStore.EXPECT().
			GetAnswerByID(gomock.Any(), answerID).
			Return(&model.Answer{
				ID:         answerID,
				QuestionID: 1,
				UserID:     999, // 不同的用户，会触发通知
			}, nil).
			AnyTimes()

		// 执行测试
		result, err := qaService.CreateComment(ctx, answerID, content, userID)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(300), result.ID)
		assert.Equal(t, answerID, result.AnswerID)
		assert.Equal(t, content, result.Content)
		assert.Equal(t, userID, result.UserID)
	})

	t.Run("缺少用户身份信息", func(t *testing.T) {
		ctx := context.Background() // 没有用户身份信息

		// 执行测试
		result, err := qaService.CreateComment(ctx, 200, "内容", 100)

		// 验证结果
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "user identity not found in context", err.Error())
	})

	t.Run("创建评论失败-数据库错误", func(t *testing.T) {
		answerID := int64(200)
		content := "测试内容"
		userID := int64(100)

		identity := auth.Identity{
			UserID:   userID,
			Username: "testuser",
		}
		ctx := auth.WithIdentity(context.Background(), identity)

		// Mock: 数据库错误
		mockStore.EXPECT().
			CreateComment(ctx, gomock.Any()).
			Return(int64(0), errors.New("database error")).
			Times(1)

		// 执行测试
		result, err := qaService.CreateComment(ctx, answerID, content, userID)

		// 验证结果
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "database error", err.Error())
	})
}

func TestGetComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := service.NewMockQAStore(ctrl)
	producer := messaging.NewKafkaProducer(config.Conf.Kafka)
	qaService := service.NewQAService(mockStore, producer, &config.Conf)
	ctx := context.Background()

	t.Run("成功获取评论", func(t *testing.T) {
		commentID := int64(300)
		comment := &model.Comment{
			ID:       commentID,
			AnswerID: 200,
			Content:  "测试评论",
			UserID:   100,
		}

		// Mock: 获取评论
		mockStore.EXPECT().
			GetCommentByID(ctx, commentID).
			Return(comment, nil).
			Times(1)

		// 执行测试
		result, err := qaService.GetComment(ctx, commentID)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, commentID, result.ID)
		assert.Equal(t, "测试评论", result.Content)
	})

	t.Run("评论不存在", func(t *testing.T) {
		commentID := int64(999)

		// Mock: 评论不存在
		mockStore.EXPECT().
			GetCommentByID(ctx, commentID).
			Return(nil, errors.New("comment not found")).
			Times(1)

		// 执行测试
		result, err := qaService.GetComment(ctx, commentID)

		// 验证结果
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestListComments(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := service.NewMockQAStore(ctrl)
	producer := messaging.NewKafkaProducer(config.Conf.Kafka)
	qaService := service.NewQAService(mockStore, producer, &config.Conf)
	ctx := context.Background()

	t.Run("成功获取评论列表", func(t *testing.T) {
		answerID := int64(200)
		page := int64(1)
		pageSize := int32(10)

		comments := []*model.Comment{
			{ID: 1, AnswerID: answerID, Content: "评论1", UserID: 100},
			{ID: 2, AnswerID: answerID, Content: "评论2", UserID: 101},
		}

		// Mock: 获取评论列表
		mockStore.EXPECT().
			ListCommentsByAnswerID(ctx, answerID, int64(0), pageSize).
			Return(comments, nil).
			Times(1)

		// Mock: 获取评论总数
		mockStore.EXPECT().
			CountCommentsByAnswerID(ctx, answerID).
			Return(int64(15), nil).
			Times(1)

		// Mock: 获取用户名
		mockStore.EXPECT().
			GetUsernamesByIDs(ctx, gomock.Any()).
			Return(map[int64]string{100: "user1", 101: "user2"}, nil).
			Times(1)

		// 执行测试
		results, total, err := qaService.ListComments(ctx, answerID, page, pageSize)

		// 验证结果
		assert.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, int64(15), total)
		assert.Equal(t, "user1", results[0].Username)
		assert.Equal(t, "评论1", results[0].Content)
	})

	t.Run("空评论列表", func(t *testing.T) {
		answerID := int64(200)
		page := int64(1)
		pageSize := int32(10)

		// Mock: 返回空列表
		mockStore.EXPECT().
			ListCommentsByAnswerID(ctx, answerID, int64(0), pageSize).
			Return([]*model.Comment{}, nil).
			Times(1)

		// Mock: 总数为0
		mockStore.EXPECT().
			CountCommentsByAnswerID(ctx, answerID).
			Return(int64(0), nil).
			Times(1)

		// 执行测试
		results, total, err := qaService.ListComments(ctx, answerID, page, pageSize)

		// 验证结果
		assert.NoError(t, err)
		assert.Len(t, results, 0)
		assert.Equal(t, int64(0), total)
	})

	t.Run("获取评论列表失败-数据库错误", func(t *testing.T) {
		answerID := int64(200)
		page := int64(1)
		pageSize := int32(10)

		// Mock: 数据库错误
		mockStore.EXPECT().
			ListCommentsByAnswerID(ctx, answerID, int64(0), pageSize).
			Return(nil, errors.New("database error")).
			Times(1)

		// 执行测试
		results, total, err := qaService.ListComments(ctx, answerID, page, pageSize)

		// 验证结果
		assert.Error(t, err)
		assert.Nil(t, results)
		assert.Equal(t, int64(0), total)
	})
}

func TestUpdateComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := service.NewMockQAStore(ctrl)
	producer := messaging.NewKafkaProducer(config.Conf.Kafka)
	qaService := service.NewQAService(mockStore, producer, &config.Conf)
	ctx := context.Background()

	t.Run("成功更新评论", func(t *testing.T) {
		commentID := int64(300)
		userID := int64(100)
		newContent := "更新后的评论"

		existingComment := &model.Comment{
			ID:       commentID,
			AnswerID: 200,
			Content:  "原评论",
			UserID:   userID,
		}

		// Mock: 获取评论
		mockStore.EXPECT().
			GetCommentByID(ctx, commentID).
			Return(existingComment, nil).
			Times(1)

		// Mock: 更新评论
		mockStore.EXPECT().
			UpdateComment(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, c *model.Comment) error {
				assert.Equal(t, newContent, c.Content)
				assert.Equal(t, userID, c.UserID)
				return nil
			}).
			Times(1)

		// 执行测试
		result, err := qaService.UpdateComment(ctx, commentID, newContent, userID)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, newContent, result.Content)
	})

	t.Run("无权限更新评论", func(t *testing.T) {
		commentID := int64(300)
		userID := int64(100)
		otherUserID := int64(200)

		existingComment := &model.Comment{
			ID:       commentID,
			AnswerID: 200,
			Content:  "原评论",
			UserID:   userID,
		}

		// Mock: 获取评论
		mockStore.EXPECT().
			GetCommentByID(ctx, commentID).
			Return(existingComment, nil).
			Times(1)

		// 执行测试 - 使用其他用户ID
		result, err := qaService.UpdateComment(ctx, commentID, "新评论", otherUserID)

		// 验证结果
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "无权限修改该评论", err.Error())
	})

	t.Run("评论不存在", func(t *testing.T) {
		commentID := int64(999)
		userID := int64(100)

		// Mock: 评论不存在
		mockStore.EXPECT().
			GetCommentByID(ctx, commentID).
			Return(nil, errors.New("comment not found")).
			Times(1)

		// 执行测试
		result, err := qaService.UpdateComment(ctx, commentID, "新评论", userID)

		// 验证结果
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestDeleteComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := service.NewMockQAStore(ctrl)
	producer := messaging.NewKafkaProducer(config.Conf.Kafka)
	qaService := service.NewQAService(mockStore, producer, &config.Conf)
	ctx := context.Background()

	t.Run("成功删除评论", func(t *testing.T) {
		commentID := int64(300)
		userID := int64(100)

		comment := &model.Comment{
			ID:       commentID,
			AnswerID: 200,
			UserID:   userID,
		}

		// Mock: 获取评论
		mockStore.EXPECT().
			GetCommentByID(ctx, commentID).
			Return(comment, nil).
			Times(1)

		// Mock: 删除评论
		mockStore.EXPECT().
			DeleteComment(ctx, commentID).
			Return(nil).
			Times(1)

		// 执行测试
		err := qaService.DeleteComment(ctx, commentID, userID)

		// 验证结果
		assert.NoError(t, err)
	})

	t.Run("无权限删除评论", func(t *testing.T) {
		commentID := int64(300)
		userID := int64(100)
		otherUserID := int64(200)

		comment := &model.Comment{
			ID:       commentID,
			AnswerID: 200,
			UserID:   userID,
		}

		// Mock: 获取评论
		mockStore.EXPECT().
			GetCommentByID(ctx, commentID).
			Return(comment, nil).
			Times(1)

		// 执行测试 - 使用其他用户ID
		err := qaService.DeleteComment(ctx, commentID, otherUserID)

		// 验证结果
		assert.Error(t, err)
		assert.Equal(t, "无权限删除该评论", err.Error())
	})

	t.Run("评论不存在", func(t *testing.T) {
		commentID := int64(999)
		userID := int64(100)

		// Mock: 评论不存在
		mockStore.EXPECT().
			GetCommentByID(ctx, commentID).
			Return(nil, errors.New("comment not found")).
			Times(1)

		// 执行测试
		err := qaService.DeleteComment(ctx, commentID, userID)

		// 验证结果
		assert.Error(t, err)
	})

	t.Run("删除失败-数据库错误", func(t *testing.T) {
		commentID := int64(300)
		userID := int64(100)

		comment := &model.Comment{
			ID:       commentID,
			AnswerID: 200,
			UserID:   userID,
		}

		// Mock: 获取评论
		mockStore.EXPECT().
			GetCommentByID(ctx, commentID).
			Return(comment, nil).
			Times(1)

		// Mock: 删除失败
		mockStore.EXPECT().
			DeleteComment(ctx, commentID).
			Return(errors.New("database error")).
			Times(1)

		// 执行测试
		err := qaService.DeleteComment(ctx, commentID, userID)

		// 验证结果
		assert.Error(t, err)
		assert.Equal(t, "database error", err.Error())
	})
}
