package database

import (
	"qahub/pkg/config"

	_ "github.com/go-sql-driver/mysql" // 匿名导入 MySQL 驱动
	"github.com/jmoiron/sqlx"
)

// NewMySQLConnection 使用提供的配置创建一个新的 sqlx 数据库实例
func NewMySQLConnection(cfg config.MySQL) (*sqlx.DB, error) {
	// 从配置中获取动态生成的 DSN
	dsn := cfg.DSN()

	// sqlx.Connect 会执行 Open 和 Ping
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// 设置最大打开连接数
	db.SetMaxOpenConns(100)

	// 设置最大空闲连接数
	db.SetMaxIdleConns(10)

	return db, nil
}
