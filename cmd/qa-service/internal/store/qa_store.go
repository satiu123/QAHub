package store

import (
	"context"
	"database/sql"
	"qahub/pkg/health"
	"qahub/qa-service/internal/model"
	"time"

	"github.com/jmoiron/sqlx"
)

type QAStore interface {
	// --- 问题相关 (Question) ---
	CreateQuestion(ctx context.Context, question *model.Question) (int64, error)
	GetQuestionByID(ctx context.Context, questionID int64) (*model.Question, error)
	ListQuestions(ctx context.Context, offset int64, limit int32) ([]*model.Question, error)
	ListQuestionsByUserID(ctx context.Context, userID int64, offset int64, limit int32) ([]*model.Question, error)
	CountQuestions(ctx context.Context) (int64, error)
	UpdateQuestion(ctx context.Context, question *model.Question) error
	DeleteQuestion(ctx context.Context, questionID int64) error
	GetAnswerCountByQuestionIDs(ctx context.Context, questionIDs []int64) (map[int64]int64, error)
	GetUsernamesByIDs(ctx context.Context, userIDs []int64) (map[int64]string, error)

	// --- 回答相关 (Answer) ---
	CreateAnswer(ctx context.Context, answer *model.Answer) (int64, error)
	GetAnswerByID(ctx context.Context, answerID int64) (*model.Answer, error)
	ListAnswersByQuestionID(ctx context.Context, questionID int64, offset int64, limit int32) ([]*model.Answer, error)
	// ListAnswersByUserID(ctx context.Context, userID int64, offset int, limit int) ([]*model.Answer, error)
	CountAnswersByQuestionID(ctx context.Context, questionID int64) (int64, error)
	GetUserVotesForAnswers(ctx context.Context, userID int64, answerIDs []int64) (map[int64]bool, error)

	CreateAnswerVote(ctx context.Context, answerID, userID int64, isUpvote bool) error
	DeleteAnswerVote(ctx context.Context, answerID, userID int64) error
	IncrementAnswerUpvoteCount(ctx context.Context, answerID int64) error
	DecrementAnswerUpvoteCount(ctx context.Context, answerID int64) error
	CountVotesByAnswerID(ctx context.Context, answerID int64) (int64, error)

	UpdateAnswer(ctx context.Context, answer *model.Answer) error
	DeleteAnswer(ctx context.Context, answerID int64) error

	// --- 评论相关 (Comment) ---
	CreateComment(ctx context.Context, comment *model.Comment) (int64, error)
	GetCommentByID(ctx context.Context, commentID int64) (*model.Comment, error)
	ListCommentsByAnswerID(ctx context.Context, answerID int64, offset int64, limit int32) ([]*model.Comment, error)
	CountCommentsByAnswerID(ctx context.Context, answerID int64) (int64, error)
	UpdateComment(ctx context.Context, comment *model.Comment) error
	DeleteComment(ctx context.Context, commentID int64) error

	ExecTx(ctx context.Context, fn func(QAStore) error) error
}
type querier interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	GetContext(ctx context.Context, dest any, query string, args ...any) error
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
}
type sqlxQAStore struct {
	db querier
	// 保存一个对 *sqlx.DB 的引用，用于开启新事务
	dbConn        *sqlx.DB
	healthChecker *health.Checker
}

func NewQAStore(db *sqlx.DB) *sqlxQAStore {
	return &sqlxQAStore{
		db:     db,
		dbConn: db, // 确保 dbConn 被正确初始化
	}
}

func (s *sqlxQAStore) SetHealthUpdater(updater health.StatusUpdater, serviceName string) {
	s.healthChecker = health.NewChecker(updater, serviceName)
	go s.startCheckHealth()
}

func (s *sqlxQAStore) startCheckHealth() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.healthChecker.CheckAndSetStatus(s.dbConn.PingContext, "MySQL")
	}
}

// --- 问题相关 (Question) ---

func (s *sqlxQAStore) CreateQuestion(ctx context.Context, question *model.Question) (int64, error) {
	query := "INSERT INTO questions (title, content, user_id) VALUES (?, ?, ?)"
	result, err := s.db.ExecContext(ctx, query, question.Title, question.Content, question.UserID)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *sqlxQAStore) GetQuestionByID(ctx context.Context, questionID int64) (*model.Question, error) {
	query := "SELECT id, title, content, user_id, created_at, updated_at FROM questions WHERE id = ?"
	var question model.Question
	err := s.db.GetContext(ctx, &question, query, questionID)
	if err != nil {
		return nil, err
	}
	return &question, nil
}

func (s *sqlxQAStore) ListQuestions(ctx context.Context, offset int64, limit int32) ([]*model.Question, error) {
	query := "SELECT id, title, content, user_id, created_at, updated_at FROM questions ORDER BY created_at DESC LIMIT ? OFFSET ?"
	var questions []*model.Question
	err := s.db.SelectContext(ctx, &questions, query, limit, offset)
	if err != nil {
		return nil, err
	}
	return questions, nil
}

func (s *sqlxQAStore) ListQuestionsByUserID(ctx context.Context, userID int64, offset int64, limit int32) ([]*model.Question, error) {
	query := "SELECT id, title, content, user_id, created_at, updated_at FROM questions WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?"
	var questions []*model.Question
	err := s.db.SelectContext(ctx, &questions, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	return questions, nil
}

func (s *sqlxQAStore) CountQuestions(ctx context.Context) (int64, error) {
	var count int64
	query := "SELECT COUNT(*) FROM questions"
	err := s.db.GetContext(ctx, &count, query)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *sqlxQAStore) UpdateQuestion(ctx context.Context, question *model.Question) error {
	query := "UPDATE questions SET title = ?, content = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?"
	_, err := s.db.ExecContext(ctx, query, question.Title, question.Content, question.ID)
	return err
}

func (s *sqlxQAStore) DeleteQuestion(ctx context.Context, questionID int64) error {
	query := "DELETE FROM questions WHERE id = ?"
	_, err := s.db.ExecContext(ctx, query, questionID)
	return err
}

// --- 回答相关 (Answer) ---

func (s *sqlxQAStore) CreateAnswer(ctx context.Context, answer *model.Answer) (int64, error) {
	query := "INSERT INTO answers (question_id, content, user_id) VALUES (?, ?, ?)"
	result, err := s.db.ExecContext(ctx, query, answer.QuestionID, answer.Content, answer.UserID)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *sqlxQAStore) GetAnswerByID(ctx context.Context, answerID int64) (*model.Answer, error) {
	query := "SELECT id, question_id, content, user_id, upvote_count, created_at, updated_at FROM answers WHERE id = ?"
	var answer model.Answer
	err := s.db.GetContext(ctx, &answer, query, answerID)
	if err != nil {
		return nil, err
	}
	return &answer, nil
}

func (s *sqlxQAStore) ListAnswersByQuestionID(ctx context.Context, questionID int64, offset int64, limit int32) ([]*model.Answer, error) {
	query := "SELECT id, question_id, content, user_id, upvote_count, created_at, updated_at FROM answers WHERE question_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?"
	var answers []*model.Answer
	err := s.db.SelectContext(ctx, &answers, query, questionID, limit, offset)
	if err != nil {
		return nil, err
	}
	return answers, nil
}

func (s *sqlxQAStore) CountAnswersByQuestionID(ctx context.Context, questionID int64) (int64, error) {
	var count int64
	query := "SELECT COUNT(*) FROM answers WHERE question_id = ?"
	err := s.db.GetContext(ctx, &count, query, questionID)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetAnswerCountByQuestionIDs 批量获取多个问题的回答数量
func (s *sqlxQAStore) GetAnswerCountByQuestionIDs(ctx context.Context, questionIDs []int64) (map[int64]int64, error) {
	counts := make(map[int64]int64)
	if len(questionIDs) == 0 {
		return counts, nil
	}

	query, args, err := sqlx.In("SELECT question_id, COUNT(*) AS cnt FROM answers WHERE question_id IN (?) GROUP BY question_id", questionIDs)
	if err != nil {
		return nil, err
	}

	query = s.dbConn.Rebind(query)
	var rows []struct {
		QuestionID int64 `db:"question_id"`
		Count      int64 `db:"cnt"`
	}

	if err := s.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, err
	}

	for _, row := range rows {
		counts[row.QuestionID] = row.Count
	}

	return counts, nil
}

// GetUserVotesForAnswers 获取用户对一组答案的投票状态
func (s *sqlxQAStore) GetUserVotesForAnswers(ctx context.Context, userID int64, answerIDs []int64) (map[int64]bool, error) {
	votes := make(map[int64]bool)
	if len(answerIDs) == 0 || userID == 0 {
		return votes, nil
	}

	query, args, err := sqlx.In("SELECT answer_id FROM answers_votes WHERE answer_id IN (?) AND user_id = ?", answerIDs, userID)
	if err != nil {
		return nil, err
	}

	query = s.dbConn.Rebind(query)

	var votedAnswerIDs []int64
	err = s.db.SelectContext(ctx, &votedAnswerIDs, query, args...)
	if err != nil {
		return nil, err
	}

	for _, id := range votedAnswerIDs {
		votes[id] = true
	}

	return votes, nil
}

// GetUsernamesByIDs 批量获取用户ID对应的用户名
func (s *sqlxQAStore) GetUsernamesByIDs(ctx context.Context, userIDs []int64) (map[int64]string, error) {
	result := make(map[int64]string)
	if len(userIDs) == 0 {
		return result, nil
	}

	query, args, err := sqlx.In("SELECT id, username FROM users WHERE id IN (?)", userIDs)
	if err != nil {
		return nil, err
	}

	query = s.dbConn.Rebind(query)
	var rows []struct {
		ID       int64  `db:"id"`
		Username string `db:"username"`
	}

	if err := s.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, err
	}

	for _, row := range rows {
		result[row.ID] = row.Username
	}

	return result, nil
}

func (s *sqlxQAStore) UpdateAnswer(ctx context.Context, answer *model.Answer) error {
	query := "UPDATE answers SET content = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?"
	_, err := s.db.ExecContext(ctx, query, answer.Content, answer.ID)
	return err
}

func (s *sqlxQAStore) DeleteAnswer(ctx context.Context, answerID int64) error {
	query := "DELETE FROM answers WHERE id = ?"
	_, err := s.db.ExecContext(ctx, query, answerID)
	return err
}

// --- 评论相关 (Comment) ---

func (s *sqlxQAStore) CreateComment(ctx context.Context, comment *model.Comment) (int64, error) {
	query := "INSERT INTO comments (answer_id, content, user_id) VALUES (?, ?, ?)"
	result, err := s.db.ExecContext(ctx, query, comment.AnswerID, comment.Content, comment.UserID)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *sqlxQAStore) GetCommentByID(ctx context.Context, commentID int64) (*model.Comment, error) {
	query := "SELECT id, answer_id, user_id, content, created_at, updated_at FROM comments WHERE id = ?"
	var comment model.Comment
	err := s.db.GetContext(ctx, &comment, query, commentID)
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (s *sqlxQAStore) ListCommentsByAnswerID(ctx context.Context, answerID int64, offset int64, limit int32) ([]*model.Comment, error) {
	query := "SELECT id, answer_id, user_id, content, created_at, updated_at FROM comments WHERE answer_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?"
	var comments []*model.Comment
	err := s.db.SelectContext(ctx, &comments, query, answerID, limit, offset)
	if err != nil {
		return nil, err
	}
	return comments, nil
}

func (s *sqlxQAStore) CountCommentsByAnswerID(ctx context.Context, answerID int64) (int64, error) {
	var count int64
	query := "SELECT COUNT(*) FROM comments WHERE answer_id = ?"
	err := s.db.GetContext(ctx, &count, query, answerID)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *sqlxQAStore) UpdateComment(ctx context.Context, comment *model.Comment) error {
	query := "UPDATE comments SET content = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?"
	_, err := s.db.ExecContext(ctx, query, comment.Content, comment.ID)
	return err
}

func (s *sqlxQAStore) DeleteComment(ctx context.Context, commentID int64) error {
	query := "DELETE FROM comments WHERE id = ?"
	_, err := s.db.ExecContext(ctx, query, commentID)
	return err
}

// --- 投票相关方法 ---

func (s *sqlxQAStore) CreateAnswerVote(ctx context.Context, answerID, userID int64, isUpvote bool) error {
	query := "INSERT INTO answers_votes (answer_id, user_id, is_upvote) VALUES (?, ?, ?)"
	_, err := s.db.ExecContext(ctx, query, answerID, userID, isUpvote)
	return err
}

func (s *sqlxQAStore) DeleteAnswerVote(ctx context.Context, answerID, userID int64) error {
	query := "DELETE FROM answers_votes WHERE answer_id = ? AND user_id = ?"
	_, err := s.db.ExecContext(ctx, query, answerID, userID)
	return err
}

func (s *sqlxQAStore) IncrementAnswerUpvoteCount(ctx context.Context, answerID int64) error {
	query := "UPDATE answers SET upvote_count = upvote_count + 1 WHERE id = ?"
	_, err := s.db.ExecContext(ctx, query, answerID)
	return err
}

func (s *sqlxQAStore) DecrementAnswerUpvoteCount(ctx context.Context, answerID int64) error {
	query := "UPDATE answers SET upvote_count = upvote_count - 1 WHERE id = ?"
	_, err := s.db.ExecContext(ctx, query, answerID)
	return err
}

func (s *sqlxQAStore) CountVotesByAnswerID(ctx context.Context, answerID int64) (int64, error) {
	var count int64
	query := "SELECT upvote_count FROM answers WHERE id = ?"
	err := s.db.GetContext(ctx, &count, query, answerID)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// ExecTx 用于执行一个包含多个数据库操作的事务
func (s *sqlxQAStore) ExecTx(ctx context.Context, fn func(QAStore) error) error {
	tx, err := s.dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	// 创建一个新的 store 实例，它持有的是事务对象 tx
	txStore := &sqlxQAStore{
		db:     tx,
		dbConn: s.dbConn,
	}
	// 执行回调函数
	err = fn(txStore)
	if err != nil {
		// 如果回调函数返回错误，则回滚事务
		if rbErr := tx.Rollback(); rbErr != nil {
			return rbErr
		}
		return err
	}
	return tx.Commit()
}
