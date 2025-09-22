package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Conf 是一个全局变量，用于存储所有应用程序的配置
var Conf Config

// Config 对应于 config.yaml 文件的顶层结构
type Config struct {
	Server        Server        `mapstructure:"server"`
	Log           Log           `mapstructure:"log"`
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

// Log 对应于 [log] 配置部分
type Log struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// MySQL 对应于 [mysql] 配置部分
type MySQL struct {
	Host      string `mapstructure:"host"`
	Port      int    `mapstructure:"port"`
	User      string `mapstructure:"user"`
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
		m.User,
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
}

// Elasticsearch 对应于 [elasticsearch] 配置部分
type Elasticsearch struct {
	URLs []string `mapstructure:"urls"`
}

// MongoDB 对应于 [mongodb] 配置部分
type MongoDB struct {
	URI      string `mapstructure:"uri"`
	Database string `mapstructure:"database"`
}

// Services 对应于 [services] 配置部分
type Services struct {
	UserService   UserService   `mapstructure:"user_service"`
	QAService     QAService     `mapstructure:"qa_service"`
	SearchService SearchService `mapstructure:"search_service"`
}

// UserService 对应于 [services.user_service] 配置部分
type UserService struct {
	JWTSecret        string `mapstructure:"jwt_secret"`
	TokenExpireHours int    `mapstructure:"token_expire_hours"`
	GrpcPort         string `mapstructure:"grpc_port"`
	HttpPort         string `mapstructure:"http_port"`
}

// QAService 对应于 [services.qa_service] 配置部分
type QAService struct {
	GrpcPort string `mapstructure:"grpc_port"`
	HttpPort string `mapstructure:"http_port"`
}

// SearchService 对应于 [services.search_service] 配置部分
type SearchService struct {
	GrpcPort string `mapstructure:"grpc_port"`
	HttpPort string `mapstructure:"http_port"`
}

// Init 函数用于初始化配置加载
// 它会读取指定路径下的配置文件，并将其解析到全局的 Conf 变量中
// 也支持环境变量覆盖，例如 SERVER_PORT 将会覆盖 server.port 的值
func Init(configPath string) error {
	// 设置配置文件路径
	viper.AddConfigPath(configPath)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// 启用环境变量覆盖
	// 例如，要覆盖数据库DSN，可以设置环境变量 MYSQL_DSN
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	// 将配置解析到 Conf 变量中
	return viper.Unmarshal(&Conf)
}
