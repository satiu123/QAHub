package store

import (
	"context"
	"qahub/internal/qa/model"

	"github.com/jmoiron/sqlx"
)

type QAStore interface {
	// --- 问题相关 (Question) ---
	CreateQuestion(ctx context.Context, question *model.Question) (int64, error)
	GetQuestionByID(ctx context.Context, questionID int64) (*model.Question, error)
	ListQuestions(ctx context.Context, offset int, limit int) ([]*model.Question, error)
	CountQuestions(ctx context.Context) (int64, error)
	UpdateQuestion(ctx context.Context, question *model.Question) error
	DeleteQuestion(ctx context.Context, questionID int64) error

	// --- 回答相关 (Answer) ---
	CreateAnswer(ctx context.Context, answer *model.Answer) (int64, error)
	GetAnswerByID(ctx context.Context, answerID int64) (*model.Answer, error)
	ListAnswersByQuestionID(ctx context.Context, questionID int64, offset int, limit int) ([]*model.Answer, error)
	CountAnswersByQuestionID(ctx context.Context, questionID int64) (int64, error)

	IncrementAnswerUpvoteCount(ctx context.Context, answerID int64) error
	DecrementAnswerUpvoteCount(ctx context.Context, answerID int64) error
	CountVotesByAnswerID(ctx context.Context, answerID int64) (int64, error)

	UpdateAnswer(ctx context.Context, answer *model.Answer) error
	DeleteAnswer(ctx context.Context, answerID int64) error

	// --- 评论相关 (Comment) ---
	CreateComment(ctx context.Context, comment *model.Comment) (int64, error)
	GetCommentByID(ctx context.Context, commentID int64) (*model.Comment, error)
	ListCommentsByAnswerID(ctx context.Context, answerID int64, offset int, limit int) ([]*model.Comment, error)
	CountCommentsByAnswerID(ctx context.Context, answerID int64) (int64, error)
	UpdateComment(ctx context.Context, comment *model.Comment) error
	DeleteComment(ctx context.Context, commentID int64) error
}

type sqlxQAStore struct {
	db *sqlx.DB
}

func NewQAStore(db *sqlx.DB) QAStore {
	return &sqlxQAStore{db: db}
}

// --- 问题相关 (Question) ---

func (s *sqlxQAStore) CreateQuestion(ctx context.Context, question *model.Question) (int64, error) {
	query := "INSERT INTO questions (title, content, user_id) VALUES (?, ?, ?)"
	result, err := s.db.ExecContext(ctx, query, question.Title, question.Content, question.UserID)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *sqlxQAStore) GetQuestionByID(ctx context.Context, questionID int64) (*model.Question, error) {
	query := "SELECT id, title, content, user_id, created_at, updated_at FROM questions WHERE id = ?"
	var question model.Question
	err := s.db.GetContext(ctx, &question, query, questionID)
	if err != nil {
		return nil, err
	}
	return &question, nil
}

func (s *sqlxQAStore) ListQuestions(ctx context.Context, offset int, limit int) ([]*model.Question, error) {
	query := "SELECT id, title, content, user_id, created_at, updated_at FROM questions ORDER BY created_at DESC LIMIT ? OFFSET ?"
	var questions []*model.Question
	err := s.db.SelectContext(ctx, &questions, query, limit, offset)
	if err != nil {
		return nil, err
	}
	return questions, nil
}

func (s *sqlxQAStore) CountQuestions(ctx context.Context) (int64, error) {
	var count int64
	query := "SELECT COUNT(*) FROM questions"
	err := s.db.GetContext(ctx, &count, query)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *sqlxQAStore) UpdateQuestion(ctx context.Context, question *model.Question) error {
	query := "UPDATE questions SET title = ?, content = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?"
	_, err := s.db.ExecContext(ctx, query, question.Title, question.Content, question.ID)
	return err
}

func (s *sqlxQAStore) DeleteQuestion(ctx context.Context, questionID int64) error {
	query := "DELETE FROM questions WHERE id = ?"
	_, err := s.db.ExecContext(ctx, query, questionID)
	return err
}

// --- 回答相关 (Answer) ---

func (s *sqlxQAStore) CreateAnswer(ctx context.Context, answer *model.Answer) (int64, error) {
	query := "INSERT INTO answers (question_id, content, user_id) VALUES (?, ?, ?)"
	result, err := s.db.ExecContext(ctx, query, answer.QuestionID, answer.Content, answer.UserID)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *sqlxQAStore) GetAnswerByID(ctx context.Context, answerID int64) (*model.Answer, error) {
	query := "SELECT id, question_id, content, user_id, upvote_count, created_at, updated_at FROM answers WHERE id = ?"
	var answer model.Answer
	err := s.db.GetContext(ctx, &answer, query, answerID)
	if err != nil {
		return nil, err
	}
	return &answer, nil
}

func (s *sqlxQAStore) ListAnswersByQuestionID(ctx context.Context, questionID int64, offset int, limit int) ([]*model.Answer, error) {
	query := "SELECT id, question_id, content, user_id, upvote_count, created_at, updated_at FROM answers WHERE question_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?"
	var answers []*model.Answer
	err := s.db.SelectContext(ctx, &answers, query, questionID, limit, offset)
	if err != nil {
		return nil, err
	}
	return answers, nil
}

func (s *sqlxQAStore) CountAnswersByQuestionID(ctx context.Context, questionID int64) (int64, error) {
	var count int64
	query := "SELECT COUNT(*) FROM answers WHERE question_id = ?"
	err := s.db.GetContext(ctx, &count, query, questionID)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *sqlxQAStore) UpdateAnswer(ctx context.Context, answer *model.Answer) error {
	query := "UPDATE answers SET content = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?"
	_, err := s.db.ExecContext(ctx, query, answer.Content, answer.ID)
	return err
}

func (s *sqlxQAStore) DeleteAnswer(ctx context.Context, answerID int64) error {
	query := "DELETE FROM answers WHERE id = ?"
	_, err := s.db.ExecContext(ctx, query, answerID)
	return err
}

// --- 评论相关 (Comment) ---

func (s *sqlxQAStore) CreateComment(ctx context.Context, comment *model.Comment) (int64, error) {
	query := "INSERT INTO comments (answer_id, content, user_id) VALUES (?, ?, ?)"
	result, err := s.db.ExecContext(ctx, query, comment.AnswerID, comment.Content, comment.UserID)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *sqlxQAStore) GetCommentByID(ctx context.Context, commentID int64) (*model.Comment, error) {
	query := "SELECT id, answer_id, user_id, content, created_at, updated_at FROM comments WHERE id = ?"
	var comment model.Comment
	err := s.db.GetContext(ctx, &comment, query, commentID)
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (s *sqlxQAStore) ListCommentsByAnswerID(ctx context.Context, answerID int64, offset int, limit int) ([]*model.Comment, error) {
	query := "SELECT id, answer_id, user_id, content, created_at, updated_at FROM comments WHERE answer_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?"
	var comments []*model.Comment
	err := s.db.SelectContext(ctx, &comments, query, answerID, limit, offset)
	if err != nil {
		return nil, err
	}
	return comments, nil
}

func (s *sqlxQAStore) CountCommentsByAnswerID(ctx context.Context, answerID int64) (int64, error) {
	var count int64
	query := "SELECT COUNT(*) FROM comments WHERE answer_id = ?"
	err := s.db.GetContext(ctx, &count, query, answerID)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *sqlxQAStore) UpdateComment(ctx context.Context, comment *model.Comment) error {
	query := "UPDATE comments SET content = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?"
	_, err := s.db.ExecContext(ctx, query, comment.Content, comment.ID)
	return err
}

func (s *sqlxQAStore) DeleteComment(ctx context.Context, commentID int64) error {
	query := "DELETE FROM comments WHERE id = ?"
	_, err := s.db.ExecContext(ctx, query, commentID)
	return err
}

// --- 投票相关方法 ---

func (s *sqlxQAStore) IncrementAnswerUpvoteCount(ctx context.Context, answerID int64) error {
	query := "UPDATE answers SET upvote_count = upvote_count + 1 WHERE id = ?"
	_, err := s.db.ExecContext(ctx, query, answerID)
	return err
}

func (s *sqlxQAStore) DecrementAnswerUpvoteCount(ctx context.Context, answerID int64) error {
	query := "UPDATE answers SET upvote_count = upvote_count - 1 WHERE id = ?"
	_, err := s.db.ExecContext(ctx, query, answerID)
	return err
}

func (s *sqlxQAStore) CountVotesByAnswerID(ctx context.Context, answerID int64) (int64, error) {
	var count int64
	query := "SELECT upvote_count FROM answers WHERE id = ?"
	err := s.db.GetContext(ctx, &count, query, answerID)
	if err != nil {
		return 0, err
	}
	return count, nil
}
