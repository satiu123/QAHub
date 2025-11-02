package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Conf 是一个全局变量，用于存储所有应用程序的配置
var Conf Config

// Config 对应于 config.yaml 文件的顶层结构
type Config struct {
	Server        Server        `mapstructure:"server"`
	Log           LogConfig     `mapstructure:"log"`
	MySQL         MySQL         `mapstructure:"mysql"`
	Redis         Redis         `mapstructure:"redis"`
	Kafka         Kafka         `mapstructure:"kafka"`
	Elasticsearch Elasticsearch `mapstructure:"elasticsearch"`
	MongoDB       MongoDB       `mapstructure:"mongodb"`
	Services      Services      `mapstructure:"services"`
}

// Server 对应于 [server] 配置部分
type Server struct {
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

// LogConfig 对应于 [log] 配置部分
type LogConfig struct {
	Level         string            `mapstructure:"level"`          // 日志级别: debug, info, warn, error
	Format        string            `mapstructure:"format"`         // 日志格式: json, text
	AddSource     bool              `mapstructure:"add_source"`     // 是否在日志中添加源码位置
	Output        string            `mapstructure:"output"`         // 日志输出位置: stdout, stderr, file
	File          lumberjack.Logger `mapstructure:"file"`           // 当 output 为 file 时，文件的归档配置
	InitialFields map[string]any    `mapstructure:"initial_fields"` // 添加到所有日志的初始字段 (如 service_name)
}

// MySQL 对应于 [mysql] 配置部分
type MySQL struct {
	Host      string `mapstructure:"host"`
	Port      int    `mapstructure:"port"`
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
	DBName    string `mapstructure:"dbname"`
	Charset   string `mapstructure:"charset"`
	ParseTime string `mapstructure:"parseTime"`
	Loc       string `mapstructure:"loc"`
}

// DSN 方法根据配置字段动态构建 MySQL DSN 字符串
func (m *MySQL) DSN() string {
	// "user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%s&loc=%s",
		m.Username,
		m.Password,
		m.Host,
		m.Port,
		m.DBName,
		m.Charset,
		m.ParseTime,
		m.Loc,
	)
}

// Redis 对应于 [redis] 配置部分
type Redis struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// Kafka 对应于 [kafka] 配置部分
type Kafka struct {
	Brokers []string `mapstructure:"brokers"`
	Topics  Topics   `mapstructure:"topics"`
}

// Topics 对应于 [kafka.topics] 配置部分
type Topics struct {
	QAEvents           string `mapstructure:"qa_events"`
	NotificationEvents string `mapstructure:"notification_events"`
}

// Elasticsearch 对应于 [elasticsearch] 配置部分
type Elasticsearch struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
}

func (e *Elasticsearch) URLs() []string {
	return []string{fmt.Sprintf("http://%s:%d", e.Host, e.Port)}
}

// MongoDB 对应于 [mongodb] 配置部分
type MongoDB struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
}

func (m *MongoDB) URI() string {
	return fmt.Sprintf("mongodb://%s:%s@%s:%d/%s?authSource=%s",
		m.Username, // 用户名
		m.Password, // 密码
		m.Host,     // 主机
		m.Port,     // 端口
		m.Database, // 默认数据库
		m.Database, // 认证数据库 (authSource)
	)
}

// Services 对应于 [services] 配置部分
type Services struct {
	UserService         UserService         `mapstructure:"user_service"`
	QAService           QAService           `mapstructure:"qa_service"`
	SearchService       SearchService       `mapstructure:"search_service"`
	NotificationService NotificationService `mapstructure:"notification_service"`
	Gateway             Gateway             `mapstructure:"gateway"`
}

// UserService 对应于 [services.user_service] 配置部分
type UserService struct {
	JWTSecret        string   `mapstructure:"jwt_secret"`
	TokenExpireHours int      `mapstructure:"token_expire_hours"`
	GrpcPort         string   `mapstructure:"grpc_port"`
	HttpPort         string   `mapstructure:"http_port"`
	PublicMethods    []string `mapstructure:"public_methods"`
}

// QAService 对应于 [services.qa_service] 配置部分
type QAService struct {
	GrpcPort      string   `mapstructure:"grpc_port"`
	HttpPort      string   `mapstructure:"http_port"`
	PublicMethods []string `mapstructure:"public_methods"`
}

// SearchService 对应于 [services.search_service] 配置部分
type SearchService struct {
	GrpcPort      string   `mapstructure:"grpc_port"`
	HttpPort      string   `mapstructure:"http_port"`
	PublicMethods []string `mapstructure:"public_methods"`
}

// NotificationService 对应于 [services.notification_service] 配置部分
type NotificationService struct {
	GrpcPort      string   `mapstructure:"grpc_port"`
	HttpPort      string   `mapstructure:"http_port"`
	PublicMethods []string `mapstructure:"public_methods"`
}

// Gateway 对应于 [service.gateway] 配置部分
type Gateway struct {
	Port                        string `mapstructure:"port"`
	UserServiceEndpoint         string `mapstructure:"user_service_endpoint"`
	QaServiceEndpoint           string `mapstructure:"qa_service_endpoint"`
	SearchServiceEndpoint       string `mapstructure:"search_service_endpoint"`
	NotificationServiceEndpoint string `mapstructure:"notification_service_endpoint"`
}

// Init 函数用于初始化配置加载
// 它会读取指定路径下的配置文件，并将其解析到全局的 Conf 变量中
// 也支持环境变量覆盖，例如 SERVER_PORT 将会覆盖 server.port 的值
func Init(configPath string) error {
	// 设置配置文件路径，如果 configPath 不为空
	if configPath != "" {
		viper.AddConfigPath(configPath)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	// 启用环境变量覆盖
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		} else {
			return err
		}
	}

	// 将配置解析到 Conf 变量中
	return viper.Unmarshal(&Conf)
}

func (c *Config) QuestionCreatedDestination() string {
	return c.Kafka.Topics.QAEvents
}

func (c *Config) NotificationDestination() string {
	return c.Kafka.Topics.NotificationEvents
}
