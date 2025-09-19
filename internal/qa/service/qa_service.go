package service

import (
	"context"
	"errors"
	"qahub/internal/qa/model"
	"qahub/internal/qa/store"
)

// QAService 定义了问答服务的业务逻辑接口
type QAService interface {
	// --- 问题相关 ---

	CreateQuestion(ctx context.Context, title, content string, userID int64) (*model.Question, error)
	GetQuestion(ctx context.Context, questionID int64) (*model.Question, error)
	ListQuestions(ctx context.Context, page, pageSize int) ([]*model.Question, int64, error)
	UpdateQuestion(ctx context.Context, questionID int64, title, content string, userID int64) (*model.Question, error)
	DeleteQuestion(ctx context.Context, questionID, userID int64) error

	// --- 回答相关 ---

	CreateAnswer(ctx context.Context, questionID int64, content string, userID int64) (*model.Answer, error)
	GetAnswer(ctx context.Context, answerID int64) (*model.Answer, error)
	ListAnswers(ctx context.Context, questionID int64, page, pageSize int) ([]*model.Answer, int64, error)

	UpvoteAnswer(ctx context.Context, answerID, userID int64) error
	DownvoteAnswer(ctx context.Context, answerID, userID int64) error
	CountVotes(ctx context.Context, answerID int64) (int64, error)

	UpdateAnswer(ctx context.Context, answerID int64, content string, userID int64) (*model.Answer, error)
	DeleteAnswer(ctx context.Context, answerID, userID int64) error

	// --- 评论相关 ---

	CreateComment(ctx context.Context, answerID int64, content string, userID int64) (*model.Comment, error)
	GetComment(ctx context.Context, commentID int64) (*model.Comment, error)
	ListComments(ctx context.Context, answerID int64, page, pageSize int) ([]*model.Comment, int64, error)
	UpdateComment(ctx context.Context, commentID int64, content string, userID int64) (*model.Comment, error)
	DeleteComment(ctx context.Context, commentID, userID int64) error
}

// qaService 是 QAService 接口的实现
type qaService struct {
	store store.QAStore
}

// NewQAService 创建一个新的 QAService
func NewQAService(s store.QAStore) QAService {
	return &qaService{store: s}
}

// --- 问题实现 ---

// CreateQuestion 创建一个新问题
func (s *qaService) CreateQuestion(ctx context.Context, title, content string, userID int64) (*model.Question, error) {
	if title == "" || content == "" {
		return nil, errors.New("标题和内容不能为空")
	}
	question := &model.Question{
		Title:   title,
		Content: content,
		UserID:  userID,
	}
	quetion_id, err := s.store.CreateQuestion(ctx, question)
	if err != nil {
		return nil, err
	}
	question.ID = quetion_id
	return question, nil
}

// GetQuestion 根据 ID 获取问题详情
func (s *qaService) GetQuestion(ctx context.Context, questionID int64) (*model.Question, error) {
	return s.store.GetQuestionByID(ctx, questionID)
}

// ListQuestions 返回分页的问题列表和总数
func (s *qaService) ListQuestions(ctx context.Context, page, pageSize int) ([]*model.Question, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
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

func (s *qaService) UpdateQuestion(ctx context.Context, questionID int64, title, content string, userID int64) (*model.Question, error) {
	if title == "" || content == "" {
		return nil, errors.New("标题和内容不能为空")
	}
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
	return s.store.DeleteQuestion(ctx, questionID)
}

// --- 回答实现 ---
func (s *qaService) CreateAnswer(ctx context.Context, questionID int64, content string, userID int64) (*model.Answer, error) {
	if content == "" {
		return nil, errors.New("内容不能为空")
	}
	answer := &model.Answer{
		QuestionID: questionID,
		Content:    content,
		UserID:     userID,
	}
	answer_id, err := s.store.CreateAnswer(ctx, answer)
	if err != nil {
		return nil, err
	}
	answer.ID = answer_id
	return answer, nil
}

func (s *qaService) GetAnswer(ctx context.Context, answerID int64) (*model.Answer, error) {
	return s.store.GetAnswerByID(ctx, answerID)
}

func (s *qaService) ListAnswers(ctx context.Context, questionID int64, page, pageSize int) ([]*model.Answer, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	answers, err := s.store.ListAnswersByQuestionID(ctx, questionID, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}
	count, err := s.store.CountAnswersByQuestionID(ctx, questionID)
	if err != nil {
		return nil, 0, err
	}
	return answers, count, nil
}

func (s *qaService) UpdateAnswer(ctx context.Context, answerID int64, content string, userID int64) (*model.Answer, error) {
	if content == "" {
		return nil, errors.New("内容不能为空")
	}
	answer, err := s.store.GetAnswerByID(ctx, answerID)
	if err != nil {
		return nil, err
	}
	if answer.UserID != userID {
		return nil, errors.New("无权限修改该回答")
	}
	answer.Content = content
	if err := s.store.UpdateAnswer(ctx, answer); err != nil {
		return nil, err
	}
	return answer, nil
}

func (s *qaService) DeleteAnswer(ctx context.Context, answerID, userID int64) error {
	answer, err := s.store.GetAnswerByID(ctx, answerID)
	if err != nil {
		return err
	}
	if answer.UserID != userID {
		return errors.New("无权限删除该回答")
	}
	return s.store.DeleteAnswer(ctx, answerID)
}

func (s *qaService) UpvoteAnswer(ctx context.Context, answerID, userID int64) error {
	return s.store.ExecTx(ctx, func(tx store.QAStore) error {
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
}

func (s *qaService) DownvoteAnswer(ctx context.Context, answerID, userID int64) error {
	return s.store.ExecTx(ctx, func(tx store.QAStore) error {
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
}

func (s *qaService) CountVotes(ctx context.Context, answerID int64) (int64, error) {
	return s.store.CountVotesByAnswerID(ctx, answerID)
}

// --- 评论实现 ---

// CreateComment 创建一个新评论
func (s *qaService) CreateComment(ctx context.Context, answerID int64, content string, userID int64) (*model.Comment, error) {
	if content == "" {
		return nil, errors.New("内容不能为空")
	}
	comment := &model.Comment{
		AnswerID: answerID,
		Content:  content,
		UserID:   userID,
	}
	comment_id, err := s.store.CreateComment(ctx, comment)
	if err != nil {
		return nil, err
	}
	comment.ID = comment_id
	return comment, nil
}

// GetComment 根据 ID 获取评论详情
func (s *qaService) GetComment(ctx context.Context, commentID int64) (*model.Comment, error) {
	return s.store.GetCommentByID(ctx, commentID)
}

// ListComments 返回分页的评论列表和总数
func (s *qaService) ListComments(ctx context.Context, answerID int64, page, pageSize int) ([]*model.Comment, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
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
	if content == "" {
		return nil, errors.New("内容不能为空")
	}
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
