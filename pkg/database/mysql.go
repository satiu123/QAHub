package database

import (
	"log"
	"time"

	"qahub/pkg/config"

	_ "github.com/go-sql-driver/mysql" // 匿名导入 MySQL 驱动
	"github.com/jmoiron/sqlx"
)

// NewMySQLConnection 使用提供的配置创建一个新的 sqlx 数据库实例
// 增加了重试逻辑以应对数据库启动慢的问题
func NewMySQLConnection(cfg config.MySQL) (*sqlx.DB, error) {
	dsn := cfg.DSN()

	var db *sqlx.DB
	var err error

	maxRetries := 10
	retryInterval := time.Second * 5

	for i := range maxRetries {
		db, err = sqlx.Connect("mysql", dsn)
		if err == nil {
			log.Println("Successfully connected to the database.")
			db.SetMaxOpenConns(100)
			db.SetMaxIdleConns(10)
			return db, nil
		}

		log.Printf("Failed to connect to database (attempt %d/%d): %v. Retrying in %v...", i+1, maxRetries, err, retryInterval)
		time.Sleep(retryInterval)
	}

	return nil, err // 返回最后一次的连接错误
}
