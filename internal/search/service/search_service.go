package service

import (
	"context"
	"encoding/json"
	"log"

	"qahub/internal/search/store"
	"qahub/pkg/config"
	"qahub/pkg/messaging"

	"github.com/segmentio/kafka-go"
)

const (
	TopicQuestions = "qa_events"       // 定义消费的主题
	GroupID        = "search-consumer" // 定义消费者组ID
)

// Service 结构体封装了搜索服务的所有业务逻辑
type Service struct {
	store       *store.Store
	kafkaReader *kafka.Reader
	handlers    map[messaging.EventType]EventHandler
}

// New 函数创建一个新的 Service 实例
func New(s *store.Store, cfg config.Kafka) *Service {
	// 配置 Kafka Reader
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		Topic:    TopicQuestions,
		GroupID:  GroupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	service := &Service{
		store:       s,
		kafkaReader: reader,
	}
	service.handlers = service.registerHandlers()
	return service
}

// StartConsumer 启动 Kafka 消费者，在一个无限循环中读取消息
func (s *Service) StartConsumer(ctx context.Context) {
	log.Println("开始消费 Kafka 消息...")
	for {
		// 从 Kafka 读取消息
		msg, err := s.kafkaReader.ReadMessage(ctx)
		if err != nil {
			log.Printf("读取 Kafka 消息失败: %v", err)
			continue // 或者根据错误类型决定是否退出
		}

		log.Printf("收到消息, Topic: %s, Offset: %d, Value: %s\n", msg.Topic, msg.Offset, string(msg.Value))

		// 解析事件头以获取事件类型
		var eventData struct {
			Header messaging.EventHeader `json:"header"`
		}
		if err := json.Unmarshal(msg.Value, &eventData); err != nil {
			log.Printf("解析事件头失败: %v", err)
			continue
		}

		// 根据事件类型调用相应的处理函数
		if handler, exists := s.handlers[eventData.Header.Type]; exists {
			if err := handler(ctx, string(eventData.Header.Type), msg.Value); err != nil {
				log.Printf("处理事件失败 (Type: %s): %v", eventData.Header.Type, err)
			}
		} else {
			log.Printf("未注册的事件类型: %s", eventData.Header.Type)
		}
	}
}

// SearchQuestions 调用 store 层执行搜索
func (s *Service) SearchQuestions(ctx context.Context, query string) ([]messaging.QuestionPayload, error) {
	return s.store.SearchQuestions(ctx, query)
}

// Close 方法用于优雅地关闭服务资源，例如 Kafka reader
func (s *Service) Close() error {
	if s.kafkaReader != nil {
		log.Println("正在关闭 Kafka reader...")
		return s.kafkaReader.Close()
	}
	return nil
}
