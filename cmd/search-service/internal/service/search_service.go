package service

import (
	"context"
	"log/slog"

	"qahub/pkg/log"
	"qahub/search-service/internal/store"

	"qahub/pkg/messaging"
)

const (
	TopicQuestions = "qa_events"       // 定义消费的主题
	GroupID        = "search-consumer" // 定义消费者组ID
)

type SearchService interface {
	SearchQuestions(ctx context.Context, query string) ([]messaging.QuestionPayload, error)
	IndexAllQuestions(ctx context.Context) error
	DeleteIndexAllQuestions(ctx context.Context) error
}

// searchService 结构体封装了搜索服务的所有业务逻辑
type searchService struct {
	store store.SearchStore
}

// New 函数创建一个新的 searchService 实例
func NewSearchService(s store.SearchStore) *searchService {
	return &searchService{
		store: s,
	}
}

// SearchQuestions 调用 store 层执行搜索
func (s *searchService) SearchQuestions(ctx context.Context, query string) ([]messaging.QuestionPayload, error) {
	logger := log.FromContext(ctx)
	
	results, err := s.store.SearchQuestions(ctx, query)
	if err != nil {
		logger.Error("搜索问题失败",
			slog.String("query", query),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	
	logger.Info("问题搜索成功",
		slog.String("query", query),
		slog.Int("result_count", len(results)),
	)
	return results, nil
}

// IndexAllQuestions 从 QA Service 获取所有问题并建立索引
func (s *searchService) IndexAllQuestions(ctx context.Context) error {
	logger := log.FromContext(ctx)
	
	err := s.store.IndexAllQuestions(ctx)
	if err != nil {
		logger.Error("索引所有问题失败",
			slog.String("error", err.Error()),
		)
		return err
	}
	
	logger.Info("所有问题索引完成")
	return nil
}

// DeleteIndexAllQuestions 删除所有问题索引并重新创建空索引
func (s *searchService) DeleteIndexAllQuestions(ctx context.Context) error {
	logger := log.FromContext(ctx)
	
	err := s.store.DeleteIndexAllQuestions(ctx)
	if err != nil {
		logger.Error("删除所有问题索引失败",
			slog.String("error", err.Error()),
		)
		return err
	}
	
	logger.Info("所有问题索引已删除")
	return nil
}
