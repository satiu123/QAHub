package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"qahub/pkg/messaging"
)

type EventHandler func(ctx context.Context, eventType string, payload []byte) error

func (s *Service) registerHandlers() map[messaging.EventType]EventHandler {
	return map[messaging.EventType]EventHandler{
		messaging.EventQuestionCreated: s.handleQuestionCreated,
		messaging.EventQuestionUpdated: s.handleQuestionUpdated,
		messaging.EventQuestionDeleted: s.handleQuestionDeleted,
	}
}

func (s *Service) handleQuestionCreated(ctx context.Context, eventType string, payload []byte) error {
	var event messaging.QuestionCreatedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("解析 QuestionCreatedEvent 失败: %w", err)
	}
	if err := s.store.IndexQuestion(ctx, event.Payload); err != nil {
		return fmt.Errorf("索引问题文档失败 (ID: %d): %w", event.Payload.ID, err)
	}
	log.Printf("成功索引问题文档 (ID: %d)", event.Payload.ID)
	return nil
}
func (s *Service) handleQuestionUpdated(ctx context.Context, eventType string, payload []byte) error {
	var event messaging.QuestionUpdatedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("解析 QuestionUpdatedEvent 失败: %w", err)
	}
	if err := s.store.IndexQuestion(ctx, event.Payload); err != nil {
		return fmt.Errorf("更新问题索引失败 (ID: %d): %w", event.Payload.ID, err)
	}
	log.Printf("成功更新问题索引 (ID: %d)", event.Payload.ID)
	return nil
}

func (s *Service) handleQuestionDeleted(ctx context.Context, eventType string, payload []byte) error {
	var event messaging.QuestionDeletedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("解析 QuestionDeletedEvent 失败: %w", err)
	}
	if err := s.store.DeleteQuestion(ctx, event.Payload.ID); err != nil {
		return fmt.Errorf("删除问题索引失败 (ID: %d): %w", event.Payload.ID, err)
	}
	log.Printf("成功删除问题索引 (ID: %d)", event.Payload.ID)
	return nil
}
