package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// QuestionPayload 定义了与 pkg/messaging/events.go 中一致的结构，
// 以使此脚本可以独立运行。
type QuestionPayload struct {
	ID       uint64   `json:"id"`
	Title    string   `json:"title"`
	Content  string   `json:"content"`
	Tags     []string `json:"tags,omitempty"`
	AuthorID uint64   `json:"author_id"`
}

const (
	esURL      = "http://localhost:9200"
	gatewayURL = "http://localhost:8080"
	testDocID  = "9999"
	indexName  = "questions"
)
const (
	token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTg2MDQ2MjQsImlhdCI6MTc1ODUxODIyNCwidXNlcl9pZCI6MSwidXNlcm5hbWUiOiJzYXRpdSJ9.OoUm68LG6Zc-KsW-NKcokiMK9UT5F5CX8KiH1RHmyjw"
)

func main() {
	log.Println("--- 开始运行 Search Service E2E 测试 ---")

	// 准备测试数据
	testDoc := QuestionPayload{
		ID:       9999,
		Title:    "一个关于Go语言的测试问题",
		Content:  "如何用Go实现一个高性能的搜索服务？",
		AuthorID: 1,
	}

	// 步骤 1: 直接向 Elasticsearch 索引一个测试文档
	log.Println("步骤 1: 正在索引测试文档...")
	if err := indexDocumentDirectly(testDoc); err != nil {
		log.Fatalf("❌ 测试失败: 索引文档时出错: %v", err)
	}
	log.Println("文档索引成功.")

	// 步骤 2: 等待 ES 完成索引刷新
	log.Println("等待2秒以确保索引刷新...")
	time.Sleep(2 * time.Second)

	// 步骤 3: 通过 API 网关调用搜索服务
	log.Println("步骤 3: 使用关键词 'Go语言' 调用搜索API...")
	results, err := searchViaAPI("Go语言")
	if err != nil {
		log.Fatalf("❌ 测试失败: 调用搜索API时出错: %v", err)
	}
	log.Printf("API 返回 %d 个结果.", len(results))

	// 步骤 4: 验证搜索结果
	log.Println("步骤 4: 正在验证结果...")
	if len(results) == 0 {
		log.Fatalf("❌ 测试失败: 期望至少1个结果, 但返回了0个.")
	}

	found := false
	for _, q := range results {
		if q.ID == testDoc.ID {
			log.Println("在搜索结果中找到了测试文档.")
			found = true
			break
		}
	}

	if !found {
		log.Fatalf("❌ 测试失败: 在搜索结果中未找到ID为 %s 的测试文档.", testDocID)
	}

	log.Println("✅ 测试通过!")
}

// indexDocumentDirectly 直接调用 ES API 索引文档
func indexDocumentDirectly(doc QuestionPayload) error {
	url := fmt.Sprintf("%s/%s/_doc/%s?refresh=true", esURL, indexName, testDocID)
	body, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ES返回了非2xx状态码: %s, 响应: %s", resp.Status, string(bodyBytes))
	}
	return nil
}

// searchViaAPI 通过网关调用搜索 API
func searchViaAPI(query string) ([]QuestionPayload, error) {
	url := fmt.Sprintf("%s/api/v1/search?q=%s", gatewayURL, query)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API返回了非200状态码: %s, 响应: %s", resp.Status, string(bodyBytes))
	}

	var results []QuestionPayload
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("解析API响应失败: %w", err)
	}

	return results, nil
}
