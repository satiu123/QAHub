package store

import (
	"context"
	"qahub/notification-service/internal/model"
	"qahub/pkg/health"
	"time"

	"github.com/jmoiron/sqlx"
)

type mysqlNotificationStore struct {
	db            *sqlx.DB
	healthChecker *health.Checker
}

func NewMySQLNotificationStore(db *sqlx.DB) *mysqlNotificationStore {
	return &mysqlNotificationStore{db: db}
}

func (m *mysqlNotificationStore) SetHealthUpdater(checker health.StatusUpdater, serviceName string) {
	m.healthChecker = health.NewChecker(checker, serviceName)
	go m.startHealthCheck()
}

func (m *mysqlNotificationStore) startHealthCheck() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		m.healthChecker.CheckAndSetStatus(m.db.PingContext, "MySQL")
	}
}

func (m *mysqlNotificationStore) Create(ctx context.Context, notification *model.Notification) error {
	return nil
}

func (m *mysqlNotificationStore) GetByRecipientID(ctx context.Context, userID int64, limit int32, offset int64) ([]*model.Notification, error) {
	return []*model.Notification{}, nil
}

func (m *mysqlNotificationStore) CountByRecipientID(ctx context.Context, userID int64) (int64, error) {
	return 0, nil
}

func (m *mysqlNotificationStore) MarkAsRead(ctx context.Context, notificationID string, userID int64) error {
	return nil
}
func (m *mysqlNotificationStore) MarkManyAsRead(ctx context.Context, notificationIDs []string, userID int64) (int64, error) {
	return 0, nil
}

func (m *mysqlNotificationStore) Delete(ctx context.Context, notificationID string, userID int64) error {
	return nil
}
func (m *mysqlNotificationStore) DeleteMany(ctx context.Context, notificationIDs []string, userID int64) (int64, error) {
	return 0, nil
}

func (m *mysqlNotificationStore) CountUnread(ctx context.Context, userID int64) (int64, error) {
	return 0, nil
}
