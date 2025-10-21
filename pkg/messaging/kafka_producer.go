package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"qahub/pkg/config"
	"qahub/pkg/health"
	"qahub/pkg/util"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer interface {
	SendMessage(ctx context.Context, destination string, payload any) error
	Close() error
}

// KafkaProducer 是一个封装了 kafka.Writer 的生产者
type KafkaProducer struct {
	writer        *kafka.Writer
	brokers       []string
	healthChecker *health.Checker
}

// NewKafkaProducer 创建一个新的 KafkaProducer 实例
func NewKafkaProducer(cfg config.Kafka) *KafkaProducer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Balancer:     &kafka.LeastBytes{},
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}
	return &KafkaProducer{
		writer:  writer,
		brokers: cfg.Brokers,
	}
}

func (p *KafkaProducer) CheckConnection() error {
	if len(p.brokers) == 0 {
		return errors.New("没有配置 Kafka broker 地址")
	}

	// 尝试连接到第一个 broker
	conn, err := kafka.Dial("tcp", p.brokers[0])
	if err != nil {
		log.Printf("连接到 Kafka broker 失败: %v", err)
		return err
	}
	// 成功连接后立即关闭
	defer util.Cleanup("Kafka broker connection", conn.Close)
	return nil
}

func (p *KafkaProducer) SetHealthUpdater(updater health.StatusUpdater, serviceName string) {
	p.healthChecker = health.NewChecker(updater, serviceName)
	go p.startHealthCheck()
}

func (p *KafkaProducer) startHealthCheck() {
	{
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			p.healthChecker.CheckAndSetStatus(func(ctx context.Context) error {
				return p.CheckConnection()
			}, "KafkaProducer")
		}
	}
}

// SendMessage 向指定的 Kafka 主题发送消息
// 消息负载会自动被序列化为 JSON
func (p *KafkaProducer) SendMessage(ctx context.Context, destination string, payload any) error {
	msgBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("序列化 Kafka 消息负载失败: %v", err)
		return err
	}

	msg := kafka.Message{
		Topic: destination,
		Value: msgBytes,
	}

	err = p.writer.WriteMessages(ctx, msg)
	if err != nil {
		log.Printf("写入 Kafka 消息失败: %v", err)
		return err
	}
	log.Printf("Kafka 消息已发送至主题: %s", destination)
	return nil
}

// Close 关闭底层的 kafka writer
func (p *KafkaProducer) Close() error {
	log.Println("正在关闭 Kafka 生产者")
	return p.writer.Close()
}
