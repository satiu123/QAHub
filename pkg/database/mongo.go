package database

import (
	"context"
	"log"
	"qahub/pkg/config"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewMongoConection(ctx context.Context, cfg config.MongoDB) (*mongo.Client, error) {
	clientOpts := options.Client().ApplyURI(cfg.URI())
	clientOpts.SetMaxPoolSize(100)
	clientOpts.SetMinPoolSize(10)

	var client *mongo.Client
	var err error

	maxRetries := 10
	retryInterval := time.Second * 5

	for i := range maxRetries {
		client, err = mongo.Connect(ctx, clientOpts)
		if err == nil {
			// 尝试 Ping 数据库以确保连接成功
			if pingErr := client.Ping(ctx, nil); pingErr == nil {
				log.Println("Successfully connected to MongoDB.")
				return client, nil
			} else {
				err = pingErr
			}
		}

		log.Printf("Failed to connect to MongoDB (attempt %d/%d): %v. Retrying in %v...", i+1, maxRetries, err, retryInterval)
		time.Sleep(retryInterval)
	}

	return nil, err // 返回最后一次的连接错误
}
