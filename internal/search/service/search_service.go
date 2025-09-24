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

	return &Service{
		store:       s,
		kafkaReader: reader,
	}
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

		// 根据事件类型进行不同的处理
		switch eventData.Header.Type {
		case messaging.EventQuestionCreated:
			var event messaging.QuestionCreatedEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("解析 QuestionCreatedEvent 失败: %v", err)
				continue
			}
			if err := s.store.IndexQuestion(ctx, event.Payload); err != nil {
				log.Printf("索引问题文档失败 (ID: %d): %v", event.Payload.ID, err)
			} else {
				log.Printf("成功索引问题文档 (ID: %d)", event.Payload.ID)
			}
		case messaging.EventQuestionUpdated:
			var event messaging.QuestionCreatedEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("解析 QuestionUpdatedEvent 失败: %v", err)
				continue
			}
			if err := s.store.IndexQuestion(ctx, event.Payload); err != nil {
				log.Printf("更新问题索引失败 (ID: %d): %v", event.Payload.ID, err)
			} else {
				log.Printf("成功更新问题索引 (ID: %d)", event.Payload.ID)
			}
		// TODO: 添加对 EventAnswerCreated 的处理
		default:
			log.Printf("收到未知的事件类型: %s", eventData.Header.Type)
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
