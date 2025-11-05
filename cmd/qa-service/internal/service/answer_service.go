package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"qahub/pkg/auth"
	"qahub/pkg/log"
	"qahub/pkg/messaging"
	"qahub/pkg/pagination"
	"qahub/qa-service/internal/dto"
	"qahub/qa-service/internal/model"
	"qahub/qa-service/internal/store"
	"time"
)

func (s *qaService) CreateAnswer(ctx context.Context, questionID int64, content string, userID int64) (*model.Answer, error) {
	logger := log.FromContext(ctx)
	
	answer := &model.Answer{
		QuestionID: questionID,
		Content:    content,
		UserID:     userID,
	}
	// 从上下文中提前获取用户信息
	identity, ok := auth.FromContext(ctx)
	if !ok {
		// 如果无法获取用户信息，可以根据业务逻辑决定是返回错误还是继续
		logger.Error("用户身份不在context中",
			slog.Int64("question_id", questionID),
			slog.Int64("user_id", userID),
		)
		return nil, errors.New("user identity not found in context")
	}
	answerID, err := s.store.CreateAnswer(ctx, answer)
	if err != nil {
		logger.Error("创建回答失败",
			slog.Int64("question_id", questionID),
			slog.Int64("user_id", userID),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	answer.ID = answerID

	logger.Info("回答创建成功",
		slog.Int64("answer_id", answerID),
		slog.Int64("question_id", questionID),
		slog.Int64("user_id", userID),
	)

	// 发布回答创建事件
	go func(senderUsername string, newAnswer model.Answer) {

		notifyCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		//获取问题的作者ID
		question, err := s.store.GetQuestionByID(notifyCtx, questionID)
		if err != nil {
			return
		}
		if question.UserID == userID {
			// 如果回答者是问题的作者自己，则不发送通知
			return
		}
		// 发布通知事件，通知问题的作者有了新的回答
		notificationPayload := messaging.NotificationPayload{
			RecipientID:      question.UserID,
			SenderID:         newAnswer.UserID,
			SenderName:       senderUsername,
			NotificationType: messaging.NotificationTypeNewAnswer,
			Content:          fmt.Sprintf("'%s' 回答了你的问题: '%s',内容是'%s'", senderUsername, question.Title, newAnswer.Content),
			TargetURL:        fmt.Sprintf("/questions/%d#answer-%d", question.ID, newAnswer.ID),
		}
		s.publishNotificationEvent(notifyCtx, notificationPayload)
	}(identity.Username, *answer)

	return answer, nil
}

func (s *qaService) GetAnswer(ctx context.Context, answerID int64) (*model.Answer, error) {
	return s.store.GetAnswerByID(ctx, answerID)
}

func (s *qaService) ListAnswers(ctx context.Context, questionID int64, page int64, pageSize int32, userID int64) ([]*dto.AnswerResponse, int64, error) {
	limit, offset := pagination.CalculateOffset(page, pageSize)
	answers, err := s.store.ListAnswersByQuestionID(ctx, questionID, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	count, err := s.store.CountAnswersByQuestionID(ctx, questionID)
	if err != nil {
		return nil, 0, err
	}

	// 如果没有回答，直接返回
	if len(answers) == 0 {
		return []*dto.AnswerResponse{}, count, nil
	}

	userIDSet := make(map[int64]struct{})
	for _, answer := range answers {
		userIDSet[answer.UserID] = struct{}{}
	}
	userIDs := make([]int64, 0, len(userIDSet))
	for id := range userIDSet {
		userIDs = append(userIDs, id)
	}

	usernames, err := s.store.GetUsernamesByIDs(ctx, userIDs)
	if err != nil {
		return nil, 0, err
	}

	// 提取所有回答的 ID
	answerIDs := make([]int64, len(answers))
	for i, answer := range answers {
		answerIDs[i] = answer.ID
	}

	// 获取当前用户对这些回答的点赞状态
	votes, err := s.store.GetUserVotesForAnswers(ctx, userID, answerIDs)
	if err != nil {
		return nil, 0, err
	}

	// 构建响应
	answerResponses := make([]*dto.AnswerResponse, len(answers))
	for i, answer := range answers {
		answerResponses[i] = &dto.AnswerResponse{
			Answer:          *answer,
			Username:        usernames[answer.UserID],
			IsUpvotedByUser: votes[answer.ID],
		}
	}

	return answerResponses, count, nil
}

func (s *qaService) UpdateAnswer(ctx context.Context, answerID int64, content string, userID int64) (*model.Answer, error) {
	logger := log.FromContext(ctx)
	
	answer, err := s.store.GetAnswerByID(ctx, answerID)
	if err != nil {
		logger.Error("获取回答失败",
			slog.Int64("answer_id", answerID),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	if answer.UserID != userID {
		logger.Warn("无权限修改回答",
			slog.Int64("answer_id", answerID),
			slog.Int64("user_id", userID),
			slog.Int64("owner_id", answer.UserID),
		)
		return nil, errors.New("无权限修改该回答")
	}
	answer.Content = content
	if err := s.store.UpdateAnswer(ctx, answer); err != nil {
		logger.Error("更新回答失败",
			slog.Int64("answer_id", answerID),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	
	logger.Info("回答更新成功",
		slog.Int64("answer_id", answerID),
		slog.Int64("user_id", userID),
	)
	// TODO: 发布回答更新事件
	return answer, nil
}

func (s *qaService) DeleteAnswer(ctx context.Context, answerID, userID int64) error {
	logger := log.FromContext(ctx)
	
	answer, err := s.store.GetAnswerByID(ctx, answerID)
	if err != nil {
		logger.Error("获取回答失败",
			slog.Int64("answer_id", answerID),
			slog.String("error", err.Error()),
		)
		return err
	}
	if answer.UserID != userID {
		logger.Warn("无权限删除回答",
			slog.Int64("answer_id", answerID),
			slog.Int64("user_id", userID),
			slog.Int64("owner_id", answer.UserID),
		)
		return errors.New("无权限删除该回答")
	}
	
	err = s.store.DeleteAnswer(ctx, answerID)
	if err != nil {
		logger.Error("删除回答失败",
			slog.Int64("answer_id", answerID),
			slog.String("error", err.Error()),
		)
		return err
	}
	
	logger.Info("回答删除成功",
		slog.Int64("answer_id", answerID),
		slog.Int64("user_id", userID),
	)
	// TODO: 发布回答删除事件
	return nil
}

func (s *qaService) UpvoteAnswer(ctx context.Context, answerID, userID int64) error {
	logger := log.FromContext(ctx)
	
	err := s.store.ExecTx(ctx, func(tx store.QAStore) error {
		err := tx.CreateAnswerVote(ctx, answerID, userID, true)
		if err != nil {
			return err
		}
		err = tx.IncrementAnswerUpvoteCount(ctx, answerID)
		if err != nil {
			return err
		}
		return nil
	})
	
	if err != nil {
		logger.Error("点赞回答失败",
			slog.Int64("answer_id", answerID),
			slog.Int64("user_id", userID),
			slog.String("error", err.Error()),
		)
		return err
	}
	
	logger.Debug("回答被点赞",
		slog.Int64("answer_id", answerID),
		slog.Int64("user_id", userID),
	)
	return nil
}

func (s *qaService) DownvoteAnswer(ctx context.Context, answerID, userID int64) error {
	logger := log.FromContext(ctx)
	
	err := s.store.ExecTx(ctx, func(tx store.QAStore) error {
		err := tx.DeleteAnswerVote(ctx, answerID, userID)
		if err != nil {
			return err
		}
		err = tx.DecrementAnswerUpvoteCount(ctx, answerID)
		if err != nil {
			return err
		}
		return nil
	})
	
	if err != nil {
		logger.Error("取消点赞回答失败",
			slog.Int64("answer_id", answerID),
			slog.Int64("user_id", userID),
			slog.String("error", err.Error()),
		)
		return err
	}
	
	logger.Debug("取消对回答的点赞",
		slog.Int64("answer_id", answerID),
		slog.Int64("user_id", userID),
	)
	return nil
}

func (s *qaService) CountVotes(ctx context.Context, answerID int64) (int64, error) {
	return s.store.CountVotesByAnswerID(ctx, answerID)
}
