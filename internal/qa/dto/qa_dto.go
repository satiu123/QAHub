package dto

import "qahub/internal/qa/model"

// AnswerResponse contains an answer and additional metadata for API responses
type AnswerResponse struct {
	model.Answer
	IsUpvotedByUser bool `json:"is_upvoted_by_user"` // 当前用户是否点赞了该答案
}
