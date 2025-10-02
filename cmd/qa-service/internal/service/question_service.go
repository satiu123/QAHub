package service

import (
	"context"
	"errors"
	"qahub/pkg/auth"
	"qahub/pkg/messaging"
	"qahub/pkg/pagination"
	"qahub/qa-service/internal/dto"
	"qahub/qa-service/internal/model"
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
func (s *qaService) GetQuestion(ctx context.Context, questionID int64) (*dto.QuestionResponse, error) {
	question, err := s.store.GetQuestionByID(ctx, questionID)
	if err != nil {
		return nil, err
	}
	if question == nil {
		return nil, errors.New("问题未找到")
	}

	usernames, err := s.store.GetUsernamesByIDs(ctx, []int64{question.UserID})
	if err != nil {
		return nil, err
	}
	authorName := usernames[question.UserID]

	answerCounts, err := s.store.GetAnswerCountByQuestionIDs(ctx, []int64{question.ID})
	if err != nil {
		return nil, err
	}
	answerCount := answerCounts[question.ID]

	response := &dto.QuestionResponse{
		Question:    *question,
		AuthorName:  authorName,
		AnswerCount: answerCount,
	}
	return response, nil
}

// ListQuestions 返回分页的问题列表和总数

func (s *qaService) buildQuestionResponses(ctx context.Context, questions []*model.Question) ([]*dto.QuestionResponse, error) {
	responses := make([]*dto.QuestionResponse, 0, len(questions))
	if len(questions) == 0 {
		return responses, nil
	}

	questionIDs := make([]int64, 0, len(questions))
	userIDSet := make(map[int64]struct{})
	for _, q := range questions {
		questionIDs = append(questionIDs, q.ID)
		userIDSet[q.UserID] = struct{}{}
	}

	userIDs := make([]int64, 0, len(userIDSet))
	for id := range userIDSet {
		userIDs = append(userIDs, id)
	}

	usernames, err := s.store.GetUsernamesByIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	answerCounts, err := s.store.GetAnswerCountByQuestionIDs(ctx, questionIDs)
	if err != nil {
		return nil, err
	}

	for _, q := range questions {
		responses = append(responses, &dto.QuestionResponse{
			Question:    *q,
			AuthorName:  usernames[q.UserID],
			AnswerCount: answerCounts[q.ID],
		})
	}

	return responses, nil
}

func (s *qaService) ListQuestions(ctx context.Context, page int64, pageSize int32) ([]*dto.QuestionResponse, int64, error) {
	limit, offset := pagination.CalculateOffset(page, pageSize)
	questions, err := s.store.ListQuestions(ctx, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	count, err := s.store.CountQuestions(ctx)
	if err != nil {
		return nil, 0, err
	}
	responses, err := s.buildQuestionResponses(ctx, questions)
	if err != nil {
		return nil, 0, err
	}
	return responses, count, nil
}

func (s *qaService) ListQuestionsByUserID(ctx context.Context, userID int64, page int64, pageSize int32) ([]*dto.QuestionResponse, int64, error) {
	limit, offset := pagination.CalculateOffset(page, pageSize)
	questions, err := s.store.ListQuestionsByUserID(ctx, userID, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	responses, err := s.buildQuestionResponses(ctx, questions)
	if err != nil {
		return nil, 0, err
	}
	count := int64(len(questions))
	return responses, count, nil
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
