package services

import (
	"context"
	"fmt"

	qapb "qahub/api/proto/qa"
)

type QAService struct {
	client *GRPCClient
}

func NewQAService(client *GRPCClient) *QAService {
	return &QAService{client: client}
}

// Question 问题结构
type Question struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	UserID      int64  `json:"user_id"`
	AuthorName  string `json:"author_name"`
	AnswerCount int64  `json:"answer_count"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// Answer 回答结构
type Answer struct {
	ID          int64  `json:"id"`
	QuestionID  int64  `json:"question_id"`
	Content     string `json:"content"`
	UserID      int64  `json:"user_id"`
	Username    string `json:"username"`
	UpvoteCount int32  `json:"upvote_count"`
	IsUpvoted   bool   `json:"is_upvoted"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// Comment 评论结构
type Comment struct {
	ID        int64  `json:"id"`
	AnswerID  int64  `json:"answer_id"`
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ListQuestions 获取问题列表
func (s *QAService) ListQuestions(ctx context.Context, page, pageSize int32) ([]Question, int64, error) {
	authCtx := s.client.NewAuthContext(ctx)
	resp, err := s.client.QAClient.ListQuestions(authCtx, &qapb.ListQuestionsRequest{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("获取问题列表失败: %w", err)
	}

	questions := make([]Question, 0, len(resp.Questions))
	for _, q := range resp.Questions {
		questions = append(questions, Question{
			ID:          q.Id,
			Title:       q.Title,
			Content:     q.Content,
			UserID:      q.UserId,
			AuthorName:  q.AuthorName,
			AnswerCount: q.AnswerCount,
			CreatedAt:   q.CreatedAt.AsTime().Format("2006-01-02 15:04:05"),
			UpdatedAt:   q.UpdatedAt.AsTime().Format("2006-01-02 15:04:05"),
		})
	}

	return questions, resp.TotalCount, nil
}

// GetQuestion 获取问题详情
func (s *QAService) GetQuestion(ctx context.Context, id int64) (*Question, error) {
	authCtx := s.client.NewAuthContext(ctx)
	resp, err := s.client.QAClient.GetQuestion(authCtx, &qapb.GetQuestionRequest{
		Id: id,
	})
	if err != nil {
		return nil, fmt.Errorf("获取问题详情失败: %w", err)
	}

	return &Question{
		ID:          resp.Id,
		Title:       resp.Title,
		Content:     resp.Content,
		UserID:      resp.UserId,
		AuthorName:  resp.AuthorName,
		AnswerCount: resp.AnswerCount,
		CreatedAt:   resp.CreatedAt.AsTime().Format("2006-01-02 15:04:05"),
		UpdatedAt:   resp.UpdatedAt.AsTime().Format("2006-01-02 15:04:05"),
	}, nil
}

// CreateQuestion 创建问题
func (s *QAService) CreateQuestion(ctx context.Context, title, content string) (*Question, error) {
	authCtx := s.client.NewAuthContext(ctx)
	resp, err := s.client.QAClient.CreateQuestion(authCtx, &qapb.CreateQuestionRequest{
		Title:   title,
		Content: content,
	})
	if err != nil {
		return nil, fmt.Errorf("创建问题失败: %w", err)
	}

	return &Question{
		ID:          resp.Id,
		Title:       resp.Title,
		Content:     resp.Content,
		UserID:      resp.UserId,
		AuthorName:  resp.AuthorName,
		AnswerCount: resp.AnswerCount,
		CreatedAt:   resp.CreatedAt.AsTime().Format("2006-01-02 15:04:05"),
		UpdatedAt:   resp.UpdatedAt.AsTime().Format("2006-01-02 15:04:05"),
	}, nil
}

// UpdateQuestion 更新问题
func (s *QAService) UpdateQuestion(ctx context.Context, id int64, title, content string) (*Question, error) {
	authCtx := s.client.NewAuthContext(ctx)
	resp, err := s.client.QAClient.UpdateQuestion(authCtx, &qapb.UpdateQuestionRequest{
		Id:      id,
		Title:   title,
		Content: content,
	})
	if err != nil {
		return nil, fmt.Errorf("更新问题失败: %w", err)
	}

	return &Question{
		ID:          resp.Id,
		Title:       resp.Title,
		Content:     resp.Content,
		UserID:      resp.UserId,
		AuthorName:  resp.AuthorName,
		AnswerCount: resp.AnswerCount,
		CreatedAt:   resp.CreatedAt.AsTime().Format("2006-01-02 15:04:05"),
		UpdatedAt:   resp.UpdatedAt.AsTime().Format("2006-01-02 15:04:05"),
	}, nil
}

// DeleteQuestion 删除问题
func (s *QAService) DeleteQuestion(ctx context.Context, id int64) error {
	authCtx := s.client.NewAuthContext(ctx)
	_, err := s.client.QAClient.DeleteQuestion(authCtx, &qapb.DeleteQuestionRequest{
		Id: id,
	})
	if err != nil {
		return fmt.Errorf("删除问题失败: %w", err)
	}
	return nil
}

// ListAnswers 获取问题的回答列表
func (s *QAService) ListAnswers(ctx context.Context, questionID int64, page, pageSize int32) ([]Answer, int64, error) {
	authCtx := s.client.NewAuthContext(ctx)
	resp, err := s.client.QAClient.ListAnswers(authCtx, &qapb.ListAnswersRequest{
		QuestionId: questionID,
		Page:       page,
		PageSize:   pageSize,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("获取回答列表失败: %w", err)
	}

	answers := make([]Answer, 0, len(resp.Answers))
	for _, a := range resp.Answers {
		answers = append(answers, Answer{
			ID:          a.Id,
			QuestionID:  a.QuestionId,
			Content:     a.Content,
			UserID:      a.UserId,
			Username:    a.Username,
			UpvoteCount: a.UpvoteCount,
			IsUpvoted:   a.IsUpvotedByUser,
			CreatedAt:   a.CreatedAt.AsTime().Format("2006-01-02 15:04:05"),
			UpdatedAt:   a.UpdatedAt.AsTime().Format("2006-01-02 15:04:05"),
		})
	}

	return answers, resp.TotalCount, nil
}

// CreateAnswer 创建回答
func (s *QAService) CreateAnswer(ctx context.Context, questionID int64, content string) (*Answer, error) {
	authCtx := s.client.NewAuthContext(ctx)
	resp, err := s.client.QAClient.CreateAnswer(authCtx, &qapb.CreateAnswerRequest{
		QuestionId: questionID,
		Content:    content,
	})
	if err != nil {
		return nil, fmt.Errorf("创建回答失败: %w", err)
	}

	return &Answer{
		ID:          resp.Id,
		QuestionID:  resp.QuestionId,
		Content:     resp.Content,
		UserID:      resp.UserId,
		Username:    resp.Username,
		UpvoteCount: resp.UpvoteCount,
		IsUpvoted:   resp.IsUpvotedByUser,
		CreatedAt:   resp.CreatedAt.AsTime().Format("2006-01-02 15:04:05"),
		UpdatedAt:   resp.UpdatedAt.AsTime().Format("2006-01-02 15:04:05"),
	}, nil
}

// UpdateAnswer 更新回答
func (s *QAService) UpdateAnswer(ctx context.Context, id int64, content string) (*Answer, error) {
	authCtx := s.client.NewAuthContext(ctx)
	resp, err := s.client.QAClient.UpdateAnswer(authCtx, &qapb.UpdateAnswerRequest{
		Id:      id,
		Content: content,
	})
	if err != nil {
		return nil, fmt.Errorf("更新回答失败: %w", err)
	}

	return &Answer{
		ID:          resp.Id,
		QuestionID:  resp.QuestionId,
		Content:     resp.Content,
		UserID:      resp.UserId,
		Username:    resp.Username,
		UpvoteCount: resp.UpvoteCount,
		IsUpvoted:   resp.IsUpvotedByUser,
		CreatedAt:   resp.CreatedAt.AsTime().Format("2006-01-02 15:04:05"),
		UpdatedAt:   resp.UpdatedAt.AsTime().Format("2006-01-02 15:04:05"),
	}, nil
}

// DeleteAnswer 删除回答
func (s *QAService) DeleteAnswer(ctx context.Context, id int64) error {
	authCtx := s.client.NewAuthContext(ctx)
	_, err := s.client.QAClient.DeleteAnswer(authCtx, &qapb.DeleteAnswerRequest{
		Id: id,
	})
	if err != nil {
		return fmt.Errorf("删除回答失败: %w", err)
	}
	return nil
}

// UpvoteAnswer 点赞回答
func (s *QAService) UpvoteAnswer(ctx context.Context, answerID int64) error {
	authCtx := s.client.NewAuthContext(ctx)
	_, err := s.client.QAClient.UpvoteAnswer(authCtx, &qapb.UpvoteAnswerRequest{
		AnswerId: answerID,
	})
	if err != nil {
		return fmt.Errorf("点赞失败: %w", err)
	}
	return nil
}

// DownvoteAnswer 取消点赞
func (s *QAService) DownvoteAnswer(ctx context.Context, answerID int64) error {
	authCtx := s.client.NewAuthContext(ctx)
	_, err := s.client.QAClient.DownvoteAnswer(authCtx, &qapb.DownvoteAnswerRequest{
		AnswerId: answerID,
	})
	if err != nil {
		return fmt.Errorf("取消点赞失败: %w", err)
	}
	return nil
}

// ListComments 获取回答的评论列表
func (s *QAService) ListComments(ctx context.Context, answerID int64, page, pageSize int32) ([]Comment, int64, error) {
	authCtx := s.client.NewAuthContext(ctx)
	resp, err := s.client.QAClient.ListComments(authCtx, &qapb.ListCommentsRequest{
		AnswerId: answerID,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("获取评论列表失败: %w", err)
	}

	comments := make([]Comment, 0, len(resp.Comments))
	for _, c := range resp.Comments {
		comments = append(comments, Comment{
			ID:        c.Id,
			AnswerID:  c.AnswerId,
			UserID:    c.UserId,
			Username:  c.Username,
			Content:   c.Content,
			CreatedAt: c.CreatedAt.AsTime().Format("2006-01-02 15:04:05"),
			UpdatedAt: c.UpdatedAt.AsTime().Format("2006-01-02 15:04:05"),
		})
	}

	return comments, resp.TotalCount, nil
}

// CreateComment 创建评论
func (s *QAService) CreateComment(ctx context.Context, answerID int64, content string) (*Comment, error) {
	authCtx := s.client.NewAuthContext(ctx)
	resp, err := s.client.QAClient.CreateComment(authCtx, &qapb.CreateCommentRequest{
		AnswerId: answerID,
		Content:  content,
	})
	if err != nil {
		return nil, fmt.Errorf("创建评论失败: %w", err)
	}

	return &Comment{
		ID:        resp.Id,
		AnswerID:  resp.AnswerId,
		UserID:    resp.UserId,
		Username:  resp.Username,
		Content:   resp.Content,
		CreatedAt: resp.CreatedAt.AsTime().Format("2006-01-02 15:04:05"),
		UpdatedAt: resp.UpdatedAt.AsTime().Format("2006-01-02 15:04:05"),
	}, nil
}

// UpdateComment 更新评论
func (s *QAService) UpdateComment(ctx context.Context, id int64, content string) (*Comment, error) {
	authCtx := s.client.NewAuthContext(ctx)
	resp, err := s.client.QAClient.UpdateComment(authCtx, &qapb.UpdateCommentRequest{
		Id:      id,
		Content: content,
	})
	if err != nil {
		return nil, fmt.Errorf("更新评论失败: %w", err)
	}

	return &Comment{
		ID:        resp.Id,
		AnswerID:  resp.AnswerId,
		UserID:    resp.UserId,
		Username:  resp.Username,
		Content:   resp.Content,
		CreatedAt: resp.CreatedAt.AsTime().Format("2006-01-02 15:04:05"),
		UpdatedAt: resp.UpdatedAt.AsTime().Format("2006-01-02 15:04:05"),
	}, nil
}

// DeleteComment 删除评论
func (s *QAService) DeleteComment(ctx context.Context, id int64) error {
	authCtx := s.client.NewAuthContext(ctx)
	_, err := s.client.QAClient.DeleteComment(authCtx, &qapb.DeleteCommentRequest{
		Id: id,
	})
	if err != nil {
		return fmt.Errorf("删除评论失败: %w", err)
	}
	return nil
}
