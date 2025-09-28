package dto

import "qahub/qa-service/internal/model"

type QuestionResponse struct {
	model.Question
	AuthorName  string `json:"author_name"`  // 提问者的用户名
	AnswerCount int64  `json:"answer_count"` // 回答数量
}

type AnswerResponse struct {
	model.Answer
	Username        string `json:"username"`           // 回答者的用户名
	IsUpvotedByUser bool   `json:"is_upvoted_by_user"` // 当前用户是否点赞了该答案
}

type CommentResponse struct {
	model.Comment
	Username string `json:"username"` // 评论者的用户名
}
