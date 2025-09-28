package service

import (
	"context"

	"qahub/internal/search/store"
	"qahub/pkg/config"
	"qahub/pkg/messaging"
)

const (
	TopicQuestions = "qa_events"       // 定义消费的主题
	GroupID        = "search-consumer" // 定义消费者组ID
)

// Service 结构体封装了搜索服务的所有业务逻辑
type Service struct {
	store         *store.Store
	kafkaComsumer *messaging.KafkaConsumer
}

// New 函数创建一个新的 Service 实例
func New(s *store.Store, cfg config.Kafka) *Service {
	// 配置 Kafka Comsumer
	consumer := messaging.NewKafkaConsumer(cfg, TopicQuestions, GroupID, nil)

	service := &Service{
		store:         s,
		kafkaComsumer: consumer,
	}
	handlers := service.registerHandlers()
	service.kafkaComsumer.SetHandlers(handlers)
	return service
}

// StartConsumer 启动 Kafka 消费者，在一个无限循环中读取消息
func (s *Service) StartConsumer(ctx context.Context) {
	go s.kafkaComsumer.Start(ctx)
}

// SearchQuestions 调用 store 层执行搜索
func (s *Service) SearchQuestions(ctx context.Context, query string) ([]messaging.QuestionPayload, error) {
	return s.store.SearchQuestions(ctx, query)
}

// Close 方法用于优雅地关闭服务资源，例如 Kafka reader
func (s *Service) Close() error {
	return s.kafkaComsumer.Close()
}
