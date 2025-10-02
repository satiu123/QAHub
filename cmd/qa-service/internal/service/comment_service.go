package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"qahub/pkg/auth"
	"qahub/pkg/messaging"
	"qahub/qa-service/internal/dto"
	"qahub/qa-service/internal/model"
	"time"
)

// CreateComment 创建一个新评论
func (s *qaService) CreateComment(ctx context.Context, answerID int64, content string, userID int64) (*model.Comment, error) {
	comment := &model.Comment{
		AnswerID: answerID,
		Content:  content,
		UserID:   userID,
	}

	// 从上下文中提前获取用户信息
	identity, ok := auth.FromContext(ctx)
	if !ok {
		// 如果无法获取用户信息，可以根据业务逻辑决定是返回错误还是继续
		return nil, errors.New("user identity not found in context")
	}

	commentID, err := s.store.CreateComment(ctx, comment)
	if err != nil {
		return nil, err
	}
	comment.ID = commentID

	// 启动后台协程发布通知事件
	go func(senderUsername string, newComment model.Comment) {
		//  使用 context.Background() 为后台任务创建独立的上下文
		notifyCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		// 获取答案的作者ID
		answer, err := s.store.GetAnswerByID(notifyCtx, newComment.AnswerID)
		if err != nil {
			// 在后台任务中，错误应该被记录下来，而不是被忽略
			log.Printf("error getting answer by id: %v", err)
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
	offset := calculateOffset(page, pageSize)
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
