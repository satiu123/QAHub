package store

import (
	"context"
	"errors"
	"qahub/internal/notification/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	notificationCollection = "notifications"
)

type NotificationStore interface {
	Create(ctx context.Context, notification *model.Notification) error
	GetByRecipientID(ctx context.Context, userID int64, limit, offset int) ([]*model.Notification, error)
	MarkAsRead(ctx context.Context, notificationID string, userID int64) error
}

type MongoNotificationStore struct {
	// MongoDB 连接和集合等字段
	db *mongo.Database
}

func NewMongoNotificationStore(db *mongo.Database) NotificationStore {
	return &MongoNotificationStore{db: db}
}

// Create 插入一条新的通知记录
func (m *MongoNotificationStore) Create(ctx context.Context, notification *model.Notification) error {
	_, err := m.db.Collection(notificationCollection).InsertOne(ctx, notification)
	return err
}

// GetByRecipientID 分页查询某个用户的通知列表，按时间倒序排列
func (m *MongoNotificationStore) GetByRecipientID(ctx context.Context, userID int64, limit, offset int) ([]*model.Notification, error) {
	var notifications []*model.Notification

	opts := options.Find()
	opts.SetSort(bson.D{{Key: "created_at", Value: -1}}) // 按创建时间降序
	opts.SetLimit(int64(limit))
	opts.SetSkip(int64(offset))

	filter := bson.M{"recipient_id": userID}

	cursor, err := m.db.Collection(notificationCollection).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &notifications); err != nil {
		return nil, err
	}
	return notifications, nil
}

// MarkAsRead 将指定的通知标记为已读
// 它会校验通知ID和用户ID，确保用户只能修改自己的通知
func (m *MongoNotificationStore) MarkAsRead(ctx context.Context, notificationID string, userID int64) error {
	objID, err := primitive.ObjectIDFromHex(notificationID)
	if err != nil {
		return errors.New("invalid notification id format")
	}

	filter := bson.M{
		"_id":          objID,
		"recipient_id": userID, // 确保用户只能标记自己的通知
	}
	update := bson.M{
		"$set": bson.M{"is_read": true},
	}

	result, err := m.db.Collection(notificationCollection).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("notification not found or permission denied")
	}

	return nil
}
