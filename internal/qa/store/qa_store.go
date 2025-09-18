package store

import "qahub/internal/qa/model"

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
