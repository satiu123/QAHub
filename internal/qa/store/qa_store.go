package store

import (
	"qahub/internal/qa/model"

	"github.com/jmoiron/sqlx"
)

type QAStore interface {
	// --- 问题相关 (Question) ---
	CreateQuestion(title string, content string, userID int64) (int64, error)
	GetQuestionByID(questionID int64) (*model.Question, error)
	ListQuestions(offset int, limit int) ([]*model.Question, error)
	UpdateQuestion(questionID int64, title string, content string) error
	DeleteQuestion(questionID int64) error

	// --- 回答相关 (Answer) ---
	CreateAnswer(questionID int64, content string, userID int64) (int64, error)
	GetAnswerByID(answerID int64) (*model.Answer, error)
	ListAnswersByQuestionID(questionID int64, offset int, limit int) ([]*model.Answer, error)
	UpdateAnswer(answerID int64, content string) error
	DeleteAnswer(answerID int64) error

	// --- 评论相关 (Comment) ---
	CreateComment(answerID int64, content string, userID int64) (int64, error)
	GetCommentByID(commentID int64) (*model.Comment, error)
	ListCommentsByAnswerID(answerID int64, offset int, limit int) ([]*model.Comment, error)
	UpdateComment(commentID int64, content string) error
	DeleteComment(commentID int64) error
}

type sqlxQAStore struct {
	db *sqlx.DB
}

func NewQAStore(db *sqlx.DB) QAStore {
	return &sqlxQAStore{db: db}
}

// --- 问题相关 (Question) ---

func (s *sqlxQAStore) CreateQuestion(title string, content string, userID int64) (int64, error) {
	query := "INSERT INTO questions (title, content, user_id) VALUES (?, ?, ?)"
	result, err := s.db.Exec(query, title, content, userID)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *sqlxQAStore) GetQuestionByID(questionID int64) (*model.Question, error) {
	query := "SELECT id, title, content, user_id, created_at, updated_at FROM questions WHERE id = ?"
	var question model.Question
	err := s.db.Get(&question, query, questionID)
	if err != nil {
		return nil, err
	}
	return &question, nil
}

func (s *sqlxQAStore) ListQuestions(offset int, limit int) ([]*model.Question, error) {
	query := "SELECT id, title, content, user_id, created_at, updated_at FROM questions ORDER BY created_at DESC LIMIT ? OFFSET ?"
	var questions []*model.Question
	err := s.db.Select(&questions, query, limit, offset)
	if err != nil {
		return nil, err
	}
	return questions, nil
}

func (s *sqlxQAStore) UpdateQuestion(questionID int64, title string, content string) error {
	query := "UPDATE questions SET title = ?, content = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?"
	_, err := s.db.Exec(query, title, content, questionID)
	return err
}

func (s *sqlxQAStore) DeleteQuestion(questionID int64) error {
	query := "DELETE FROM questions WHERE id = ?"
	_, err := s.db.Exec(query, questionID)
	return err
}

// --- 回答相关 (Answer) ---

func (s *sqlxQAStore) CreateAnswer(questionID int64, content string, userID int64) (int64, error) {
	query := "INSERT INTO answers (question_id, content, user_id) VALUES (?, ?, ?)"
	result, err := s.db.Exec(query, questionID, content, userID)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *sqlxQAStore) GetAnswerByID(answerID int64) (*model.Answer, error) {
	query := "SELECT id, question_id, content, user_id, upvote_count, created_at, updated_at FROM answers WHERE id = ?"
	var answer model.Answer
	err := s.db.Get(&answer, query, answerID)
	if err != nil {
		return nil, err
	}
	return &answer, nil
}

func (s *sqlxQAStore) ListAnswersByQuestionID(questionID int64, offset int, limit int) ([]*model.Answer, error) {
	query := "SELECT id, question_id, content, user_id, upvote_count, created_at, updated_at FROM answers WHERE question_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?"
	var answers []*model.Answer
	err := s.db.Select(&answers, query, questionID, limit, offset)
	if err != nil {
		return nil, err
	}
	return answers, nil
}

func (s *sqlxQAStore) UpdateAnswer(answerID int64, content string) error {
	query := "UPDATE answers SET content = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?"
	_, err := s.db.Exec(query, content, answerID)
	return err
}

func (s *sqlxQAStore) DeleteAnswer(answerID int64) error {
	query := "DELETE FROM answers WHERE id = ?"
	_, err := s.db.Exec(query, answerID)
	return err
}

// --- 评论相关 (Comment) ---

func (s *sqlxQAStore) CreateComment(answerID int64, content string, userID int64) (int64, error) {
	query := "INSERT INTO comments (answer_id, content, user_id) VALUES (?, ?, ?)"
	result, err := s.db.Exec(query, answerID, content, userID)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *sqlxQAStore) GetCommentByID(commentID int64) (*model.Comment, error) {
	query := "SELECT id, answer_id, user_id, content, created_at, updated_at FROM comments WHERE id = ?"
	var comment model.Comment
	err := s.db.Get(&comment, query, commentID)
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (s *sqlxQAStore) ListCommentsByAnswerID(answerID int64, offset int, limit int) ([]*model.Comment, error) {
	query := "SELECT id, answer_id, user_id, content, created_at, updated_at FROM comments WHERE answer_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?"
	var comments []*model.Comment
	err := s.db.Select(&comments, query, answerID, limit, offset)
	if err != nil {
		return nil, err
	}
	return comments, nil
}

func (s *sqlxQAStore) UpdateComment(commentID int64, content string) error {
	query := "UPDATE comments SET content = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?"
	_, err := s.db.Exec(query, content, commentID)
	return err
}

func (s *sqlxQAStore) DeleteComment(commentID int64) error {
	query := "DELETE FROM comments WHERE id = ?"
	_, err := s.db.Exec(query, commentID)
	return err
}
