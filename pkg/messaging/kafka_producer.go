package messaging

import (
	"context"
	"encoding/json"
	"log"
	"qahub/pkg/config"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaProducer 是一个封装了 kafka.Writer 的生产者
type KafkaProducer struct {
	writer *kafka.Writer
}

// NewKafkaProducer 创建一个新的 KafkaProducer 实例
func NewKafkaProducer(cfg config.Kafka) *KafkaProducer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Balancer:     &kafka.LeastBytes{},
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}
	return &KafkaProducer{writer: writer}
}

// SendMessage 向指定的 Kafka 主题发送消息
// 消息负载会自动被序列化为 JSON
func (p *KafkaProducer) SendMessage(ctx context.Context, topic string, payload any) error {
	msgBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("序列化 Kafka 消息负载失败: %v", err)
		return err
	}

	msg := kafka.Message{
		Topic: topic,
		Value: msgBytes,
	}

	err = p.writer.WriteMessages(ctx, msg)
	if err != nil {
		log.Printf("写入 Kafka 消息失败: %v", err)
		return err
	}
	log.Printf("Kafka 消息已发送至主题: %s", topic)
	return nil
}

// Close 关闭底层的 kafka writer
func (p *KafkaProducer) Close() error {
	log.Println("正在关闭 Kafka 生产者")
	return p.writer.Close()
}
