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
	"time"
)

// CreateComment 创建一个新评论
func (s *qaService) CreateComment(ctx context.Context, answerID int64, content string, userID int64) (*model.Comment, error) {
	logger := log.FromContext(ctx)
	
	comment := &model.Comment{
		AnswerID: answerID,
		Content:  content,
		UserID:   userID,
	}

	// 从上下文中提前获取用户信息
	identity, ok := auth.FromContext(ctx)
	if !ok {
		// 如果无法获取用户信息，可以根据业务逻辑决定是返回错误还是继续
		logger.Error("用户身份不在context中",
			slog.Int64("answer_id", answerID),
			slog.Int64("user_id", userID),
		)
		return nil, errors.New("user identity not found in context")
	}

	commentID, err := s.store.CreateComment(ctx, comment)
	if err != nil {
		logger.Error("创建评论失败",
			slog.Int64("answer_id", answerID),
			slog.Int64("user_id", userID),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	comment.ID = commentID

	logger.Info("评论创建成功",
		slog.Int64("comment_id", commentID),
		slog.Int64("answer_id", answerID),
		slog.Int64("user_id", userID),
	)

	// 启动后台协程发布通知事件
	go func(senderUsername string, newComment model.Comment) {
		//  使用 context.Background() 为后台任务创建独立的上下文
		notifyCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		// 获取答案的作者ID
		answer, err := s.store.GetAnswerByID(notifyCtx, newComment.AnswerID)
		if err != nil {
			// 在后台任务中，错误应该被记录下来，而不是被忽略
			logger.Error("后台任务：获取答案失败",
				slog.Int64("answer_id", newComment.AnswerID),
				slog.String("error", err.Error()),
			)
			return
		}

		// 判断评论者是否是答案的作者本人
		if newComment.UserID == answer.UserID {
			// 如果评论者是答案的作者自己，则不发送通知
			return
		}

		// 构建并发布通知事件
		notificationPayload := messaging.NotificationPayload{
			RecipientID:      answer.UserID,
			SenderID:         newComment.UserID,
			SenderName:       senderUsername,
			NotificationType: messaging.NotificationTypeNewComment,
			Content:          fmt.Sprintf("'%s' 评论了你的答案: '%s'", senderUsername, newComment.Content),
			TargetURL:        fmt.Sprintf("/questions/%d#comment-%d", answer.QuestionID, newComment.ID),
		}
		s.publishNotificationEvent(notifyCtx, notificationPayload)
	}(identity.Username, *comment)

	return comment, nil
}

// GetComment 根据 ID 获取评论详情
func (s *qaService) GetComment(ctx context.Context, commentID int64) (*model.Comment, error) {
	return s.store.GetCommentByID(ctx, commentID)
}

// ListComments 返回分页的评论列表和总数
func (s *qaService) ListComments(ctx context.Context, answerID int64, page int64, pageSize int32) ([]*dto.CommentResponse, int64, error) {
	limit, offset := pagination.CalculateOffset(page, pageSize)
	comments, err := s.store.ListCommentsByAnswerID(ctx, answerID, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	count, err := s.store.CountCommentsByAnswerID(ctx, answerID)
	if err != nil {
		return nil, 0, err
	}
	if len(comments) == 0 {
		return []*dto.CommentResponse{}, count, nil
	}

	userIDSet := make(map[int64]struct{})
	for _, comment := range comments {
		userIDSet[comment.UserID] = struct{}{}
	}

	userIDs := make([]int64, 0, len(userIDSet))
	for id := range userIDSet {
		userIDs = append(userIDs, id)
	}

	usernames, err := s.store.GetUsernamesByIDs(ctx, userIDs)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*dto.CommentResponse, len(comments))
	for i, comment := range comments {
		responses[i] = &dto.CommentResponse{
			Comment:  *comment,
			Username: usernames[comment.UserID],
		}
	}

	return responses, count, nil
}

// UpdateComment 修改评论，只有评论的创建者可以修改
func (s *qaService) UpdateComment(ctx context.Context, commentID int64, content string, userID int64) (*model.Comment, error) {
	logger := log.FromContext(ctx)
	
	comment, err := s.store.GetCommentByID(ctx, commentID)
	if err != nil {
		logger.Error("获取评论失败",
			slog.Int64("comment_id", commentID),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	if comment.UserID != userID {
		logger.Warn("无权限修改评论",
			slog.Int64("comment_id", commentID),
			slog.Int64("user_id", userID),
			slog.Int64("owner_id", comment.UserID),
		)
		return nil, errors.New("无权限修改该评论")
	}
	comment.Content = content
	if err := s.store.UpdateComment(ctx, comment); err != nil {
		logger.Error("更新评论失败",
			slog.Int64("comment_id", commentID),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	
	logger.Info("评论更新成功",
		slog.Int64("comment_id", commentID),
		slog.Int64("user_id", userID),
	)
	return comment, nil
}

// DeleteComment 删除评论，只有评论的创建者可以删除
func (s *qaService) DeleteComment(ctx context.Context, commentID, userID int64) error {
	logger := log.FromContext(ctx)
	
	comment, err := s.store.GetCommentByID(ctx, commentID)
	if err != nil {
		logger.Error("获取评论失败",
			slog.Int64("comment_id", commentID),
			slog.String("error", err.Error()),
		)
		return err
	}
	if comment.UserID != userID {
		logger.Warn("无权限删除评论",
			slog.Int64("comment_id", commentID),
			slog.Int64("user_id", userID),
			slog.Int64("owner_id", comment.UserID),
		)
		return errors.New("无权限删除该评论")
	}
	
	err = s.store.DeleteComment(ctx, commentID)
	if err != nil {
		logger.Error("删除评论失败",
			slog.Int64("comment_id", commentID),
			slog.String("error", err.Error()),
		)
		return err
	}
	
	logger.Info("评论删除成功",
		slog.Int64("comment_id", commentID),
		slog.Int64("user_id", userID),
	)
	return nil
}
