package service

import (
	"context"

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
	return s.store.SearchQuestions(ctx, query)
}

// IndexAllQuestions 从 QA Service 获取所有问题并建立索引
func (s *searchService) IndexAllQuestions(ctx context.Context) error {
	return s.store.IndexAllQuestions(ctx)
}

// DeleteIndexAllQuestions 删除所有问题索引并重新创建空索引
func (s *searchService) DeleteIndexAllQuestions(ctx context.Context) error {
	return s.store.DeleteIndexAllQuestions(ctx)
}
