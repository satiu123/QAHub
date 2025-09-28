package service

import (
	"context"
	"errors"
	"qahub/internal/qa/model"
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
	return comment, nil
}

// GetComment 根据 ID 获取评论详情
func (s *qaService) GetComment(ctx context.Context, commentID int64) (*model.Comment, error) {
	return s.store.GetCommentByID(ctx, commentID)
}

// ListComments 返回分页的评论列表和总数
func (s *qaService) ListComments(ctx context.Context, answerID int64, page, pageSize int) ([]*model.Comment, int64, error) {
	offset := (page - 1) * pageSize
	comments, err := s.store.ListCommentsByAnswerID(ctx, answerID, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}
	count, err := s.store.CountCommentsByAnswerID(ctx, answerID)
	if err != nil {
		return nil, 0, err
	}
	return comments, count, nil
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
