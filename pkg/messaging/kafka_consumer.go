package messaging

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"qahub/pkg/config"
	"qahub/pkg/health"
	"qahub/pkg/util"

	"github.com/segmentio/kafka-go"
)

// EventHandler 是一个函数类型，用于处理特定类型的事件
type EventHandler func(ctx context.Context, eventType string, payload []byte) error

type Consumer interface {
	Start(ctx context.Context)
	Close() error
}

// KafkaConsumer 封装了 Kafka 消费者的通用逻辑
type KafkaConsumer struct {
	reader        *kafka.Reader
	handlers      map[EventType]EventHandler
	brokers       []string
	healthChecker *health.Checker
}

// NewKafkaConsumer 创建一个新的 KafkaConsumer 实例
func NewKafkaConsumer(cfg config.Kafka, topic string, groupID string, handlers map[EventType]EventHandler) *KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &KafkaConsumer{
		reader:   reader,
		handlers: handlers,
		brokers:  cfg.Brokers,
	}
}

func (c *KafkaConsumer) SetHealthUpdater(updater health.StatusUpdater, serviceName string) {
	c.healthChecker = health.NewChecker(updater, serviceName)
	go c.startHealthCheck()
}

func (c *KafkaConsumer) startHealthCheck() {
	{
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			c.healthChecker.CheckAndSetStatus(func(ctx context.Context) error {
				return c.CheckConnection()
			}, "KafkaProducer")
		}
	}
}

func (c *KafkaConsumer) CheckConnection() error {
	if len(c.brokers) == 0 {
		return nil
	}

	// 尝试连接到第一个 broker
	conn, err := kafka.Dial("tcp", c.brokers[0])
	if err != nil {
		log.Printf("连接到 Kafka broker 失败: %v", err)
		return err
	}
	// 成功连接后立即关闭
	defer util.Cleanup("Kafka broker connection", func() error { return conn.Close() })
	return nil
}

// SetHandlers 设置事件处理器映射
func (c *KafkaConsumer) SetHandlers(handlers map[EventType]EventHandler) {
	c.handlers = handlers
}

// Start 在一个无限循环中启动消费者
func (c *KafkaConsumer) Start(ctx context.Context) {
	log.Printf("开始消费 Kafka topic '%s' (Group: %s)...", c.reader.Config().Topic, c.reader.Config().GroupID)
	log.Printf("注册的事件处理器数量: %d", len(c.handlers))
	for eventType := range c.handlers {
		log.Printf("已注册事件类型: %s", eventType)
	}

	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("读取 Kafka 消息失败: %v", err)
			continue
		}

		log.Printf("收到消息, Topic: %s, Offset: %d, Value: %s", msg.Topic, msg.Offset, string(msg.Value))

		var eventData struct {
			Header EventHeader `json:"header"`
		}
		if err := json.Unmarshal(msg.Value, &eventData); err != nil {
			log.Printf("解析事件头失败: %v", err)
			continue
		}
		eventType := EventType(eventData.Header.Type)
		// 根据事件类型调用对应的处理器
		if handler, exists := c.handlers[eventType]; exists {
			if err := handler(ctx, string(eventType), msg.Value); err != nil {
				log.Printf("处理事件失败 (Type: %s): %v", eventData.Header.Type, err)
			}
		} else {
			log.Printf("未注册的事件类型: %s", eventData.Header.Type)
		}
	}
}

// Close 关闭 Kafka reader
func (c *KafkaConsumer) Close() error {
	if c.reader != nil {
		log.Printf("正在关闭 Kafka reader for topic '%s'...", c.reader.Config().Topic)
		return c.reader.Close()
	}
	return nil
}
