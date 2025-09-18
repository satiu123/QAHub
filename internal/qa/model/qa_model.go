package model

// Question 对应于数据库中的 questions 表
type Question struct {
	ID        uint64 `db:"id"`
	Title     string `db:"title"`
	Content   string `db:"content"`
	UserID    uint64 `db:"user_id"`
	CreatedAt int64  `db:"created_at"`
	UpdatedAt int64  `db:"updated_at"`
}

// Answer 对应于数据库中的 answers 表
type Answer struct {
	ID          uint64 `db:"id"`
	QuestionID  uint64 `db:"question_id"`
	Content     string `db:"content"`
	UserID      uint64 `db:"user_id"`
	UpvoteCount int    `db:"upvote_count"`
	CreatedAt   int64  `db:"created_at"`
	UpdatedAt   int64  `db:"updated_at"`
}

// Comment 对应于数据库中的 comments 表
type Comment struct {
	ID        uint64 `db:"id"`
	AnswerID  uint64 `db:"answer_id"`
	UserID    uint64 `db:"user_id"`
	Content   string `db:"content"`
	CreatedAt int64  `db:"created_at"`
	UpdatedAt int64  `db:"updated_at"`
}
