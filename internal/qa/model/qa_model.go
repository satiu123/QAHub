package model

// Question 对应于数据库中的 questions 表
type Question struct {
	ID        int64  `db:"id"`
	Title     string `db:"title"`
	Content   string `db:"content"`
	UserID    int64  `db:"user_id"`
	CreatedAt int64  `db:"created_at"`
	UpdatedAt int64  `db:"updated_at"`
}

// Answer 对应于数据库中的 answers 表
type Answer struct {
	ID          int64  `db:"id"`
	QuestionID  int64  `db:"question_id"`
	Content     string `db:"content"`
	UserID      int64  `db:"user_id"`
	UpvoteCount int    `db:"upvote_count"`
	CreatedAt   int64  `db:"created_at"`
	UpdatedAt   int64  `db:"updated_at"`
}

// Comment 对应于数据库中的 comments 表
type Comment struct {
	ID        int64  `db:"id"`
	AnswerID  int64  `db:"answer_id"`
	UserID    int64  `db:"user_id"`
	Content   string `db:"content"`
	CreatedAt int64  `db:"created_at"`
	UpdatedAt int64  `db:"updated_at"`
}
