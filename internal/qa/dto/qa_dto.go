package dto

import "qahub/internal/qa/model"

//暂时不使用

type AnswerResponse struct {
	model.Answer
	IsUpvotedByUser bool `json:"is_upvoted_by_user"` // 当前用户是否点赞了该答案
}
