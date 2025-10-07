package service

import (
	"context"
	"log"
	"qahub/pkg/auth"
	"qahub/pkg/messaging"
	"qahub/qa-service/internal/model"
	"time"

	"github.com/google/uuid"
)

// publishQuestionEvent 是一个辅助函数，用于发布与问题相关的事件
func (s *qaService) publishQuestionEvent(ctx context.Context, eventType messaging.EventType, question *model.Question) {
	identity, _ := auth.FromContext(ctx)
	
	event := messaging.QuestionCreatedEvent{
		Header: messaging.EventHeader{
			ID:        uuid.New().String(),
			Type:      eventType,
			Source:    "qa-service",
			Timestamp: time.Now(),
		},
		Payload: messaging.QuestionPayload{
			ID:         question.ID,
			Title:      question.Title,
			Content:    question.Content,
			AuthorID:   question.UserID,
			AuthorName: identity.Username,
			CreatedAt:  question.CreatedAt,
			UpdatedAt:  question.UpdatedAt,
			// Tags: question.Tags, // 如果有Tags字段的话
		},
	}

	err := s.kafkaProducer.SendMessage(ctx, s.cfg.Topics.QAEvents, event)
	if err != nil {
		log.Printf("Failed to publish event %s for question ID %d: %v", eventType, question.ID, err)
	} else {
		log.Printf("Published event %s for question ID %d", eventType, question.ID)
	}
}

// publishNotificationEvent 是一个辅助函数，用于发布通知事件
func (s *qaService) publishNotificationEvent(ctx context.Context, payload messaging.NotificationPayload) {
	event := messaging.NotificationTriggeredEvent{
		Header: messaging.EventHeader{
			ID:        uuid.New().String(),
			Type:      messaging.EventNotificationTriggered,
			Source:    "qa-service",
			Timestamp: time.Now(),
		},
		Payload: payload,
	}

	err := s.kafkaProducer.SendMessage(ctx, s.cfg.Topics.NotificationEvents, event)
	if err != nil {
		log.Printf("Failed to publish event %s for recipient ID %d: %v", messaging.EventNotificationTriggered, payload.RecipientID, err)
	} else {
		log.Printf("Published event %s for recipient ID %d", messaging.EventNotificationTriggered, payload.RecipientID)
	}
}
