package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"qahub/pkg/messaging"
)

func (s *searchService) registerHandlers() map[messaging.EventType]messaging.EventHandler {
	return map[messaging.EventType]messaging.EventHandler{
		messaging.EventQuestionCreated: s.handleQuestionCreated,
		messaging.EventQuestionUpdated: s.handleQuestionUpdated,
		messaging.EventQuestionDeleted: s.handleQuestionDeleted,
	}
}

func (s *searchService) handleQuestionCreated(ctx context.Context, eventType string, payload []byte) error {
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
func (s *searchService) handleQuestionUpdated(ctx context.Context, eventType string, payload []byte) error {
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

func (s *searchService) handleQuestionDeleted(ctx context.Context, eventType string, payload []byte) error {
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
