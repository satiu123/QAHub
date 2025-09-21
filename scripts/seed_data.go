package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/yaml.v3"
)

// Config 数据库配置结构
type Config struct {
	MySQL struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
		Charset  string `yaml:"charset"`
	} `yaml:"mysql"`
}

// SampleData 示例数据
var sampleUsers = []struct {
	Username string
	Email    string
	Bio      string
	Password string
}{
	{"tech_expert", "tech@example.com", "资深技术专家，专注于后端开发", "$2a$10$example.hash.password1"},
	{"code_lover", "code@example.com", "热爱编程的开发者", "$2a$10$example.hash.password2"},
	{"ai_researcher", "ai@example.com", "AI研究员，专注于机器学习", "$2a$10$example.hash.password3"},
	{"web_developer", "web@example.com", "前端开发工程师", "$2a$10$example.hash.password4"},
	{"database_admin", "dba@example.com", "数据库管理员", "$2a$10$example.hash.password5"},
}

var sampleQuestions = []struct {
	Title   string
	Content string
}{
	{
		"如何优化MySQL查询性能？",
		"我有一个包含百万条记录的表，查询速度很慢。请问有什么优化方法？包括索引优化、查询语句优化等方面的建议都可以。",
	},
	{
		"Go语言中的并发编程最佳实践",
		"最近在学习Go语言的goroutine和channel，想了解一些并发编程的最佳实践。比如如何避免竞态条件，如何正确使用sync包等。",
	},
	{
		"React Hook的使用场景和注意事项",
		"刚开始学习React Hook，对useState和useEffect比较熟悉了，但不太清楚useCallback、useMemo等其他Hook的使用场景。",
	},
	{
		"微服务架构中的服务发现机制",
		"在微服务架构中，服务之间需要相互调用，请问常见的服务发现机制有哪些？各有什么优缺点？",
	},
	{
		"Redis缓存穿透和缓存雪崩的解决方案",
		"在高并发场景下，Redis可能会遇到缓存穿透和缓存雪崩的问题，请问有哪些有效的解决方案？",
	},
	{
		"Docker容器化部署的最佳实践",
		"正在学习Docker，想了解容器化部署的最佳实践，比如镜像优化、多阶段构建、安全配置等方面。",
	},
	{
		"分布式系统中的一致性问题",
		"在分布式系统中，如何保证数据的一致性？CAP理论、BASE理论在实际应用中如何权衡？",
	},
	{
		"前端性能优化的常用手段",
		"网站加载速度比较慢，想了解前端性能优化的方法，比如代码分割、懒加载、CDN等技术。",
	},
}

var sampleAnswers = []string{
	"可以从以下几个方面优化MySQL查询性能：1. 创建合适的索引 2. 优化查询语句，避免全表扫描 3. 使用EXPLAIN分析执行计划 4. 合理使用分页查询 5. 考虑读写分离和分库分表",
	"MySQL查询优化建议：添加索引时要注意不要过多，选择合适的索引类型，定期分析慢查询日志，优化表结构设计。",
	"Go并发编程要点：1. 使用channel进行goroutine通信 2. 避免共享内存，通过通信共享内存 3. 使用sync.WaitGroup等待goroutine完成 4. 使用context控制goroutine生命周期",
	"建议使用Go的race detector来检测竞态条件，合理使用sync.Mutex和sync.RWMutex保护共享资源。",
	"Hook使用建议：useState管理组件状态，useEffect处理副作用，useCallback缓存函数引用，useMemo缓存计算结果，useContext共享状态。",
	"React Hook要注意依赖数组的正确使用，避免无限循环渲染，合理使用优化类Hook。",
	"常见服务发现机制：1. 客户端发现（Eureka） 2. 服务端发现（AWS ELB） 3. 服务网格（Istio） 4. DNS发现，各有性能和复杂度的权衡。",
	"缓存问题解决方案：缓存穿透可使用布隆过滤器，缓存雪崩可设置不同过期时间，缓存击穿可使用互斥锁或双重检查。",
	"Docker最佳实践：使用多阶段构建减小镜像大小，不在容器中运行root用户，合理设置资源限制，使用.dockerignore文件。",
	"分布式一致性：强一致性使用2PC/3PC，最终一致性使用消息队列，根据业务需求选择合适的一致性级别。",
	"前端优化手段：代码压缩、图片优化、使用CDN、启用Gzip、减少HTTP请求、使用浏览器缓存、代码分割等。",
}

var sampleComments = []string{
	"这个回答很详细，学到了很多！",
	"补充一点：还可以考虑使用连接池优化数据库连接。",
	"实践过了，效果确实不错。",
	"有没有具体的代码示例？",
	"感谢分享，正好遇到了类似的问题。",
	"这种方法在生产环境中稳定吗？",
	"可以结合实际项目案例来说明吗？",
	"非常实用的建议！",
	"还有其他的解决思路吗？",
	"这个方案的性能如何？",
}

func loadConfig() (*Config, error) {
	file, err := os.Open("configs/config.yaml")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&config)
	return &config, err
}

func connectDB(config *Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		config.MySQL.User,
		config.MySQL.Password,
		config.MySQL.Host,
		config.MySQL.Port,
		config.MySQL.DBName,
		config.MySQL.Charset,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func seedUsers(db *sql.DB) ([]int64, error) {
	var userIDs []int64

	for _, user := range sampleUsers {
		// 检查用户是否已存在
		var exists int
		err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", user.Email).Scan(&exists)
		if err != nil {
			return nil, err
		}

		if exists > 0 {
			// 获取已存在用户的ID
			var userID int64
			err = db.QueryRow("SELECT id FROM users WHERE email = ?", user.Email).Scan(&userID)
			if err != nil {
				return nil, err
			}
			userIDs = append(userIDs, userID)
			continue
		}

		result, err := db.Exec(
			"INSERT INTO users (username, email, bio, password) VALUES (?, ?, ?, ?)",
			user.Username, user.Email, user.Bio, user.Password,
		)
		if err != nil {
			return nil, err
		}

		userID, err := result.LastInsertId()
		if err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	log.Printf("成功插入/获取 %d 个用户", len(userIDs))
	return userIDs, nil
}

func seedQuestions(db *sql.DB, userIDs []int64) ([]int64, error) {
	var questionIDs []int64

	for _, question := range sampleQuestions {
		// 随机选择一个用户
		userID := userIDs[rand.Intn(len(userIDs))]

		result, err := db.Exec(
			"INSERT INTO questions (title, content, user_id) VALUES (?, ?, ?)",
			question.Title, question.Content, userID,
		)
		if err != nil {
			return nil, err
		}

		questionID, err := result.LastInsertId()
		if err != nil {
			return nil, err
		}
		questionIDs = append(questionIDs, questionID)
	}

	log.Printf("成功插入 %d 个问题", len(questionIDs))
	return questionIDs, nil
}

func seedAnswers(db *sql.DB, questionIDs []int64, userIDs []int64) ([]int64, error) {
	var answerIDs []int64

	for _, questionID := range questionIDs {
		// 每个问题随机生成1-3个回答
		answerCount := rand.Intn(3) + 1

		for i := 0; i < answerCount; i++ {
			// 随机选择一个用户和一个回答内容
			userID := userIDs[rand.Intn(len(userIDs))]
			content := sampleAnswers[rand.Intn(len(sampleAnswers))]
			upvoteCount := rand.Intn(20) // 随机点赞数

			result, err := db.Exec(
				"INSERT INTO answers (question_id, content, user_id, upvote_count) VALUES (?, ?, ?, ?)",
				questionID, content, userID, upvoteCount,
			)
			if err != nil {
				return nil, err
			}

			answerID, err := result.LastInsertId()
			if err != nil {
				return nil, err
			}
			answerIDs = append(answerIDs, answerID)
		}
	}

	log.Printf("成功插入 %d 个回答", len(answerIDs))
	return answerIDs, nil
}

func seedComments(db *sql.DB, answerIDs []int64, userIDs []int64) error {
	commentCount := 0

	for _, answerID := range answerIDs {
		// 每个回答随机生成0-2个评论
		numComments := rand.Intn(3)

		for i := 0; i < numComments; i++ {
			// 随机选择一个用户和一个评论内容
			userID := userIDs[rand.Intn(len(userIDs))]
			content := sampleComments[rand.Intn(len(sampleComments))]

			_, err := db.Exec(
				"INSERT INTO comments (answer_id, user_id, content) VALUES (?, ?, ?)",
				answerID, userID, content,
			)
			if err != nil {
				return err
			}
			commentCount++
		}
	}

	log.Printf("成功插入 %d 个评论", commentCount)
	return nil
}

func main() {
	// 设置随机种子
	rand.Seed(time.Now().UnixNano())

	// 加载配置
	config, err := loadConfig()
	if err != nil {
		log.Fatal("加载配置文件失败:", err)
	}

	// 连接数据库
	db, err := connectDB(config)
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	log.Println("开始生成测试数据...")

	// 插入用户数据
	userIDs, err := seedUsers(db)
	if err != nil {
		log.Fatal("插入用户数据失败:", err)
	}

	// 插入问题数据
	questionIDs, err := seedQuestions(db, userIDs)
	if err != nil {
		log.Fatal("插入问题数据失败:", err)
	}

	// 插入回答数据
	answerIDs, err := seedAnswers(db, questionIDs, userIDs)
	if err != nil {
		log.Fatal("插入回答数据失败:", err)
	}

	// 插入评论数据
	err = seedComments(db, answerIDs, userIDs)
	if err != nil {
		log.Fatal("插入评论数据失败:", err)
	}

	log.Println("测试数据生成完成！")
	log.Printf("总计: %d 用户, %d 问题, %d 回答", len(userIDs), len(questionIDs), len(answerIDs))
}
