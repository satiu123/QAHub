package store

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"qahub/pkg/config"
	"qahub/pkg/messaging"

	"github.com/elastic/go-elasticsearch/v8"
)

const (
	IndexQuestions = "questions" // 问题索引的名称
)

// Store 结构体封装了与 Elasticsearch 的所有交互
type Store struct {
	client *elasticsearch.Client
}

// New 函数创建一个新的 Store 实例
func New(cfg config.Elasticsearch) (*Store, error) {
	// 创建 Elasticsearch 客户端配置
	esCfg := elasticsearch.Config{
		Addresses: cfg.URLs,
		// 在这里可以添加其他配置，例如用户名、密码、证书等
	}

	// 创建客户端
	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, fmt.Errorf("创建 Elasticsearch 客户端失败: %w", err)
	}

	// Ping Elasticsearch 服务器以验证连接
	res, err := client.Ping()
	if err != nil {
		return nil, fmt.Errorf("无法 Ping通 Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("Elasticsearch Ping 响应错误: %s", res.String())
	}

	fmt.Println("成功连接到 Elasticsearch")

	return &Store{client: client}, nil
}

// IndexQuestion 将一个问题文档索引到 Elasticsearch 中
func (s *Store) IndexQuestion(ctx context.Context, question messaging.QuestionPayload) error {
	// 将 question 对象序列化为 JSON
	body, err := json.Marshal(question)
	if err != nil {
		return fmt.Errorf("序列化问题失败: %w", err)
	}

	// 使用 client.Index 方法发送索引请求
	res, err := s.client.Index(
		IndexQuestions,        // 索引名称
		bytes.NewReader(body), // 文档内容
		s.client.Index.WithDocumentID(strconv.FormatInt(question.ID, 10)), // 设置文档ID
		s.client.Index.WithContext(ctx),                                   // 传递上下文
		s.client.Index.WithRefresh("true"),                                // 索引后立即刷新，以便可以立即搜索到（在生产环境中可能会考虑其他策略）
	)
	if err != nil {
		return fmt.Errorf("索引文档失败: %w", err)
	}
	defer res.Body.Close()

	// 检查响应中是否有错误
	if res.IsError() {
		return fmt.Errorf("索引响应错误: %s", res.String())
	}

	return nil
}

// SearchQuestions 在 Elasticsearch 中搜索问题
func (s *Store) SearchQuestions(ctx context.Context, query string) ([]messaging.QuestionPayload, error) {
	var buf bytes.Buffer
	// 定义 Elasticsearch 查询体
	searchQuery := map[string]any{
		"query": map[string]any{
			"multi_match": map[string]any{
				"query":  query,
				"fields": []string{"title", "content"},
			},
		},
	}

	if err := json.NewEncoder(&buf).Encode(searchQuery); err != nil {
		return nil, fmt.Errorf("编码查询体失败: %w", err)
	}

	// 执行搜索请求
	res, err := s.client.Search(
		s.client.Search.WithContext(ctx),
		s.client.Search.WithIndex(IndexQuestions),
		s.client.Search.WithBody(&buf),
		s.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, fmt.Errorf("执行搜索失败: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("搜索响应错误: %s", res.String())
	}

	// 解析响应
	var r map[string]any
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("解析响应体失败: %w", err)
	}

	// 从响应中提取文档
	var questions []messaging.QuestionPayload
	hits, found := r["hits"].(map[string]any)["hits"].([]any)
	if !found {
		return questions, nil // 没有命中，返回空切片
	}

	for _, hit := range hits {
		source, ok := hit.(map[string]any)["_source"].(map[string]any)
		if !ok {
			log.Printf("解析 _source 失败")
			continue
		}
		var q messaging.QuestionPayload
		// 将 map[string]any 转换为 JSON bytes，再反序列化到结构体
		payloadBytes, err := json.Marshal(source)
		if err != nil {
			log.Printf("重新序列化 payload 失败: %v", err)
			continue
		}
		if err := json.Unmarshal(payloadBytes, &q); err != nil {
			log.Printf("反序列化到 QuestionPayload 失败: %v", err)
			continue
		}
		questions = append(questions, q)
	}

	return questions, nil
}

// DeleteQuestion 从 Elasticsearch 中删除一个问题文档
func (s *Store) DeleteQuestion(ctx context.Context, questionID int64) error {
	// 使用 client.Delete 方法发送删除请求
	res, err := s.client.Delete(
		IndexQuestions,                      // 索引名称
		strconv.FormatInt(questionID, 10),   // 文档ID
		s.client.Delete.WithContext(ctx),    // 传递上下文
		s.client.Delete.WithRefresh("true"), // 删除后立即刷新
	)
	if err != nil {
		return fmt.Errorf("删除文档失败: %w", err)
	}
	defer res.Body.Close()

	// 检查响应中是否有错误
	if res.IsError() {
		return fmt.Errorf("删除响应错误: %s", res.String())
	}

	return nil
}

func (s *Store) ClearIndex(ctx context.Context) error {
	// 删除索引
	res, err := s.client.Indices.Delete([]string{IndexQuestions}, s.client.Indices.Delete.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("删除索引失败: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("删除索引响应错误: %s", res.String())
	}

	// 重新创建索引
	res, err = s.client.Indices.Create(IndexQuestions, s.client.Indices.Create.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("创建索引失败: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("创建索引响应错误: %s", res.String())
	}

	return nil
}
