package service

import (
	"context"
	"errors"
	"qahub/internal/qa/model"
	"qahub/pkg/auth"
	"qahub/pkg/messaging"
)

// CreateQuestion 创建一个新问题
func (s *qaService) CreateQuestion(ctx context.Context, title, content string, userID int64) (*model.Question, error) {
	question := &model.Question{
		Title:   title,
		Content: content,
		UserID:  userID,
	}
	questionID, err := s.store.CreateQuestion(ctx, question)
	if err != nil {
		return nil, err
	}
	question.ID = questionID

	// 发布问题创建事件到 Kafka
	identity, _ := auth.FromContext(ctx)
	if identity.UserID == 0 {
		identity.UserID = userID
	}
	eventCtx := auth.WithIdentity(context.Background(), identity)
	go s.publishQuestionEvent(eventCtx, messaging.EventQuestionCreated, question)

	return question, nil
}

// GetQuestion 根据 ID 获取问题详情
func (s *qaService) GetQuestion(ctx context.Context, questionID int64) (*model.Question, error) {
	return s.store.GetQuestionByID(ctx, questionID)
}

// ListQuestions 返回分页的问题列表和总数
func (s *qaService) ListQuestions(ctx context.Context, page, pageSize int) ([]*model.Question, int64, error) {
	offset := (page - 1) * pageSize
	questions, err := s.store.ListQuestions(ctx, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}
	count, err := s.store.CountQuestions(ctx)
	if err != nil {
		return nil, 0, err
	}
	return questions, count, nil
}

func (s *qaService) ListQuestionsByUserID(ctx context.Context, userID int64, page, pageSize int) ([]*model.Question, int64, error) {
	offset := (page - 1) * pageSize
	questions, err := s.store.ListQuestionsByUserID(ctx, userID, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}
	count := int64(len(questions))
	return questions, count, nil
}

func (s *qaService) UpdateQuestion(ctx context.Context, questionID int64, title, content string, userID int64) (*model.Question, error) {
	question, err := s.store.GetQuestionByID(ctx, questionID)
	if err != nil {
		return nil, err
	}
	if question.UserID != userID {
		return nil, errors.New("无权限修改该问题")
	}
	question.Title = title
	question.Content = content
	if err := s.store.UpdateQuestion(ctx, question); err != nil {
		return nil, err
	}

	identity, _ := auth.FromContext(ctx)
	if identity.UserID == 0 {
		identity.UserID = userID
	}
	eventCtx := auth.WithIdentity(context.Background(), identity)
	// 发布问题更新事件到 Kafka
	go s.publishQuestionEvent(eventCtx, messaging.EventQuestionUpdated, question)

	return question, nil
}

func (s *qaService) DeleteQuestion(ctx context.Context, questionID, userID int64) error {
	question, err := s.store.GetQuestionByID(ctx, questionID)
	if err != nil {
		return err
	}
	if question.UserID != userID {
		return errors.New("无权限删除该问题")
	}
	identity, _ := auth.FromContext(ctx)
	if identity.UserID == 0 {
		identity.UserID = userID
	}
	eventCtx := auth.WithIdentity(context.Background(), identity)
	// 发布问题删除事件到 Kafka
	go s.publishQuestionEvent(eventCtx, messaging.EventQuestionDeleted, question)
	return s.store.DeleteQuestion(ctx, questionID)
}
