package service

import (
	"context"
	"errors"
	"fmt"
	"qahub/pkg/auth"
	"qahub/pkg/messaging"
	"qahub/qa-service/internal/dto"
	"qahub/qa-service/internal/model"
)

// CreateComment 创建一个新评论
func (s *qaService) CreateComment(ctx context.Context, answerID int64, content string, userID int64) (*model.Comment, error) {
	comment := &model.Comment{
		AnswerID: answerID,
		Content:  content,
		UserID:   userID,
	}
	commentID, err := s.store.CreateComment(ctx, comment)
	if err != nil {
		return nil, err
	}
	comment.ID = commentID

	// 发布回答创建事件
	go func() {
		//获取问题的作者ID
		answer, err := s.store.GetAnswerByID(ctx, answerID)
		if err != nil {
			return
		}
		if comment.UserID == userID {
			// 如果回答者是问题的作者自己，则不发送通知
			return
		}
		identity, _ := auth.FromContext(ctx)
		// 发布通知事件，通知问题的作者有了新的回答
		notificationPayload := messaging.NotificationPayload{
			RecipientID:      answer.UserID,
			SenderID:         comment.UserID,
			SenderName:       identity.Username,
			NotificationType: messaging.NotificationTypeNewComment,
			Content:          fmt.Sprintf("'%s' 评论了你的答案: '%s'", identity.Username, comment.Content),
			TargetURL:        fmt.Sprintf("/questions/%d#comment-%d", answer.QuestionID, comment.ID),
		}
		s.publishNotificationEvent(ctx, notificationPayload)
	}()

	return comment, nil
}

// GetComment 根据 ID 获取评论详情
func (s *qaService) GetComment(ctx context.Context, commentID int64) (*model.Comment, error) {
	return s.store.GetCommentByID(ctx, commentID)
}

// ListComments 返回分页的评论列表和总数
func (s *qaService) ListComments(ctx context.Context, answerID int64, page, pageSize int) ([]*dto.CommentResponse, int64, error) {
	offset := (page - 1) * pageSize
	comments, err := s.store.ListCommentsByAnswerID(ctx, answerID, offset, pageSize)
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
	comment, err := s.store.GetCommentByID(ctx, commentID)
	if err != nil {
		return nil, err
	}
	if comment.UserID != userID {
		return nil, errors.New("无权限修改该评论")
	}
	comment.Content = content
	if err := s.store.UpdateComment(ctx, comment); err != nil {
		return nil, err
	}
	return comment, nil
}

// DeleteComment 删除评论，只有评论的创建者可以删除
func (s *qaService) DeleteComment(ctx context.Context, commentID, userID int64) error {
	comment, err := s.store.GetCommentByID(ctx, commentID)
	if err != nil {
		return err
	}
	if comment.UserID != userID {
		return errors.New("无权限删除该评论")
	}
	return s.store.DeleteComment(ctx, commentID)
}
