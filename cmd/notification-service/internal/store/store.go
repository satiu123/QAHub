package store

import (
	"context"
	"qahub/notification-service/internal/model"
)

type NotificationStore interface {
	Create(ctx context.Context, notification *model.Notification) error
	GetByRecipientID(ctx context.Context, userID int64, limit int32, offset int64) ([]*model.Notification, error)
	MarkAsRead(ctx context.Context, notificationID string, userID int64) error
	MarkManyAsRead(ctx context.Context, notificationIDs []string, userID int64) (int64, error)
	Delete(ctx context.Context, notificationID string, userID int64) error
	DeleteMany(ctx context.Context, notificationIDs []string, userID int64) (int64, error)
	CountUnread(ctx context.Context, userID int64) (int64, error)
}
