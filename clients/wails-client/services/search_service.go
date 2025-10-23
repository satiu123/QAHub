package services

import (
	"context"
	"fmt"

	searchpb "wails-client/api/proto/search"
)

type SearchService struct {
	client *GRPCClient
}

func NewSearchService(client *GRPCClient) *SearchService {
	return &SearchService{client: client}
}

// SearchResult 搜索结果
type SearchResult struct {
	ID         int64  `json:"id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	AuthorID   int64  `json:"author_id"`
	AuthorName string `json:"author_name"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// SearchQuestions 搜索问题
func (s *SearchService) SearchQuestions(ctx context.Context, query string, limit, offset int32) ([]SearchResult, error) {
	if s.client == nil || s.client.SearchClient == nil {
		return nil, fmt.Errorf("搜索服务未初始化")
	}

	authCtx := s.client.NewAuthContext(ctx)

	resp, err := s.client.SearchClient.SearchQuestions(authCtx, &searchpb.SearchQuestionsRequest{
		Query:  query,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("搜索失败: %w", err)
	}

	results := make([]SearchResult, 0, len(resp.Questions))
	for _, q := range resp.Questions {
		results = append(results, SearchResult{
			ID:         q.Id,
			Title:      q.Title,
			Content:    q.Content,
			AuthorID:   q.AuthorId,
			AuthorName: q.AuthorName,
			CreatedAt:  q.CreatedAt.AsTime().Format("2006-01-02 15:04:05"),
			UpdatedAt:  q.UpdatedAt.AsTime().Format("2006-01-02 15:04:05"),
		})
	}

	return results, nil
}

// IndexAllQuestions 索引所有问题（仅用于测试/管理）
func (s *SearchService) IndexAllQuestions(ctx context.Context) (string, error) {
	if s.client == nil || s.client.SearchClient == nil {
		return "", fmt.Errorf("搜索服务未初始化")
	}

	authCtx := s.client.NewAuthContext(ctx)

	resp, err := s.client.SearchClient.IndexAllQuestions(authCtx, &searchpb.IndexAllQuestionsRequest{})
	if err != nil {
		return "", fmt.Errorf("索引失败: %w", err)
	}

	return resp.Message, nil
}

// DeleteIndexAllQuestions 删除所有问题索引（仅用于测试/管理）
func (s *SearchService) DeleteIndexAllQuestions(ctx context.Context) (string, error) {
	if s.client == nil || s.client.SearchClient == nil {
		return "", fmt.Errorf("搜索服务未初始化")
	}

	authCtx := s.client.NewAuthContext(ctx)

	resp, err := s.client.SearchClient.DeleteIndexAllQuestions(authCtx, &searchpb.DeleteIndexAllQuestionsRequest{})
	if err != nil {
		return "", fmt.Errorf("删除索引失败: %w", err)
	}

	return resp.Message, nil
}
