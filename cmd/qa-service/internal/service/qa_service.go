package service

import (
	"context"
	"qahub/pkg/config"
	"qahub/pkg/messaging"
	"qahub/qa-service/internal/dto"
	"qahub/qa-service/internal/model"
	"qahub/qa-service/internal/store"
)

// QAService 定义了问答服务的业务逻辑接口
type QAService interface {
	// --- 问题相关 ---

	CreateQuestion(ctx context.Context, title, content string, userID int64) (*model.Question, error)
	GetQuestion(ctx context.Context, questionID int64) (*dto.QuestionResponse, error)
	ListQuestions(ctx context.Context, page, pageSize int) ([]*dto.QuestionResponse, int64, error)
	ListQuestionsByUserID(ctx context.Context, userID int64, page, pageSize int) ([]*dto.QuestionResponse, int64, error)
	UpdateQuestion(ctx context.Context, questionID int64, title, content string, userID int64) (*model.Question, error)
	DeleteQuestion(ctx context.Context, questionID, userID int64) error

	// --- 回答相关 ---

	CreateAnswer(ctx context.Context, questionID int64, content string, userID int64) (*model.Answer, error)
	GetAnswer(ctx context.Context, answerID int64) (*model.Answer, error)
	ListAnswers(ctx context.Context, questionID int64, page, pageSize int, userID int64) ([]*dto.AnswerResponse, int64, error)

	UpvoteAnswer(ctx context.Context, answerID, userID int64) error
	DownvoteAnswer(ctx context.Context, answerID, userID int64) error
	CountVotes(ctx context.Context, answerID int64) (int64, error)

	UpdateAnswer(ctx context.Context, answerID int64, content string, userID int64) (*model.Answer, error)
	DeleteAnswer(ctx context.Context, answerID, userID int64) error

	// --- 评论相关 ---

	CreateComment(ctx context.Context, answerID int64, content string, userID int64) (*model.Comment, error)
	GetComment(ctx context.Context, commentID int64) (*model.Comment, error)
	ListComments(ctx context.Context, answerID int64, page, pageSize int) ([]*dto.CommentResponse, int64, error)
	UpdateComment(ctx context.Context, commentID int64, content string, userID int64) (*model.Comment, error)
	DeleteComment(ctx context.Context, commentID, userID int64) error
}

// qaService 是 QAService 接口的实现
type qaService struct {
	store         store.QAStore
	kafkaProducer *messaging.KafkaProducer
	cfg           config.Kafka
}

// NewQAService 创建一个新的 QAService
func NewQAService(s store.QAStore, cfg config.Kafka) QAService {
	producer := messaging.NewKafkaProducer(cfg)
	return &qaService{
		store:         s,
		kafkaProducer: producer,
		cfg:           cfg,
	}
}
