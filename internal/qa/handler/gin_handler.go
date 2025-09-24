package handler

import (
	"net/http"
	"qahub/internal/qa/model"
	"qahub/internal/qa/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type QAHandler struct {
	qaService service.QAService
}

func NewQAHandler(svc service.QAService) *QAHandler {
	return &QAHandler{
		qaService: svc,
	}
}

// --- Request & Response Structs ---

type CreateQuestionRequest struct {
	Title   string `json:"title" binding:"required,min=5,max=255"`
	Content string `json:"content" binding:"required,min=5"`
}

type UpdateQuestionRequest struct {
	Title   string `json:"title" binding:"required,min=5,max=255"`
	Content string `json:"content" binding:"required,min=5"`
}

type CreateAnswerRequest struct {
	Content string `json:"content" binding:"required,min=3"`
}

type UpdateAnswerRequest struct {
	Content string `json:"content" binding:"required,min=3"`
}

type CreateCommentRequest struct {
	Content string `json:"content" binding:"required,min=2"`
}

type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required,min=2"`
}

type ListResponse struct {
	Total int64 `json:"total"`
	Data  any   `json:"data"`
}

// --- Helper Functions ---

func getAuthUserID(c *gin.Context) (int64, bool) {
	authUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return 0, false
	}
	userID, ok := authUserID.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无效的用户ID格式"})
		return 0, false
	}
	return userID, true
}

// getAuthUserIDOrGuest 尝试从 context 获取 userID，如果不存在（访客），则返回 0 和 true
func getAuthUserIDOrGuest(c *gin.Context) (int64, bool) {
	authUserID, exists := c.Get("userID")
	if !exists {
		// 访客用户，返回0
		return 0, true
	}
	userID, ok := authUserID.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无效的用户ID格式"})
		return 0, false
	}
	return userID, true
}

func getIDFromParam(c *gin.Context, paramName string) (int64, bool) {
	idStr := c.Param(paramName)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID格式: " + paramName})
		return 0, false
	}
	return id, true
}

func getPagination(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	return page, pageSize
}

// --- Question Handlers ---

func (h *QAHandler) CreateQuestion(c *gin.Context) {
	userID, ok := getAuthUserID(c)
	if !ok {
		return
	}

	var req CreateQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	question, err := h.qaService.CreateQuestion(c.Request.Context(), req.Title, req.Content, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, question)
}

func (h *QAHandler) GetQuestion(c *gin.Context) {
	questionID, ok := getIDFromParam(c, "question_id")
	if !ok {
		return
	}

	question, err := h.qaService.GetQuestion(c.Request.Context(), questionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "问题未找到"})
		return
	}

	c.JSON(http.StatusOK, question)
}

func (h *QAHandler) ListQuestions(c *gin.Context) {
	userID, _ := getAuthUserIDOrGuest(c) // 访客用户 userID 为 0
	authorQuery := c.Query("author")
	page, pageSize := getPagination(c)

	var questions []*model.Question
	var total int64
	var err error

	var authorID int64
	if authorQuery == "me" {
		// 如果 author=me，使用当前用户ID
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录用户无法查看'我的'问题"})
			return
		}
		authorID = userID
	} else if authorQuery != "" {
		// 如果指定了具体的作者ID
		parsedID, err := strconv.ParseInt(authorQuery, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的作者ID格式"})
			return
		}
		authorID = parsedID
	}

	if authorID != 0 {
		questions, total, err = h.qaService.ListQuestionsByUserID(c.Request.Context(), authorID, page, pageSize)
	} else {
		questions, total, err = h.qaService.ListQuestions(c.Request.Context(), page, pageSize)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ListResponse{Total: total, Data: questions})
}

func (h *QAHandler) ListMyQuestions(c *gin.Context) {
	userID, ok := getAuthUserID(c)
	if !ok {
		return
	}
	page, pageSize := getPagination(c)
	questions, total, err := h.qaService.ListQuestionsByUserID(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ListResponse{Total: total, Data: questions})
}

func (h *QAHandler) ListQuestionsByUserID(c *gin.Context) {
	userID, ok := getIDFromParam(c, "user_id")
	if !ok {
		return
	}
	page, pageSize := getPagination(c)
	questions, total, err := h.qaService.ListQuestionsByUserID(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ListResponse{Total: total, Data: questions})
}

func (h *QAHandler) UpdateQuestion(c *gin.Context) {
	userID, ok := getAuthUserID(c)
	if !ok {
		return
	}
	questionID, ok := getIDFromParam(c, "question_id")
	if !ok {
		return
	}

	var req UpdateQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedQuestion, err := h.qaService.UpdateQuestion(c.Request.Context(), questionID, req.Title, req.Content, userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedQuestion)
}

func (h *QAHandler) DeleteQuestion(c *gin.Context) {
	userID, ok := getAuthUserID(c)
	if !ok {
		return
	}
	questionID, ok := getIDFromParam(c, "question_id")
	if !ok {
		return
	}

	err := h.qaService.DeleteQuestion(c.Request.Context(), questionID, userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// --- Answer Handlers ---

func (h *QAHandler) CreateAnswer(c *gin.Context) {
	userID, ok := getAuthUserID(c)
	if !ok {
		return
	}
	questionID, ok := getIDFromParam(c, "question_id")
	if !ok {
		return
	}

	var req CreateAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	answer, err := h.qaService.CreateAnswer(c.Request.Context(), questionID, req.Content, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, answer)
}

func (h *QAHandler) ListAnswers(c *gin.Context) {
	questionID, ok := getIDFromParam(c, "question_id")
	if !ok {
		return
	}
	page, pageSize := getPagination(c)

	userID, ok := getAuthUserIDOrGuest(c)
	if !ok {
		return
	}

	answers, total, err := h.qaService.ListAnswers(c.Request.Context(), questionID, page, pageSize, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ListResponse{Total: total, Data: answers})
}

func (h *QAHandler) UpdateAnswer(c *gin.Context) {
	userID, ok := getAuthUserID(c)
	if !ok {
		return
	}
	answerID, ok := getIDFromParam(c, "answer_id")
	if !ok {
		return
	}

	var req UpdateAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedAnswer, err := h.qaService.UpdateAnswer(c.Request.Context(), answerID, req.Content, userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedAnswer)
}

func (h *QAHandler) DeleteAnswer(c *gin.Context) {
	userID, ok := getAuthUserID(c)
	if !ok {
		return
	}
	answerID, ok := getIDFromParam(c, "answer_id")
	if !ok {
		return
	}

	err := h.qaService.DeleteAnswer(c.Request.Context(), answerID, userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// --- Comment Handlers ---

func (h *QAHandler) CreateComment(c *gin.Context) {
	userID, ok := getAuthUserID(c)
	if !ok {
		return
	}
	answerID, ok := getIDFromParam(c, "answer_id")
	if !ok {
		return
	}

	var req CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment, err := h.qaService.CreateComment(c.Request.Context(), answerID, req.Content, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, comment)
}

func (h *QAHandler) ListComments(c *gin.Context) {
	answerID, ok := getIDFromParam(c, "answer_id")
	if !ok {
		return
	}
	page, pageSize := getPagination(c)

	comments, total, err := h.qaService.ListComments(c.Request.Context(), answerID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ListResponse{Total: total, Data: comments})
}

func (h *QAHandler) UpdateComment(c *gin.Context) {
	userID, ok := getAuthUserID(c)
	if !ok {
		return
	}
	commentID, ok := getIDFromParam(c, "comment_id")
	if !ok {
		return
	}

	var req UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedComment, err := h.qaService.UpdateComment(c.Request.Context(), commentID, req.Content, userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedComment)
}

func (h *QAHandler) DeleteComment(c *gin.Context) {
	userID, ok := getAuthUserID(c)
	if !ok {
		return
	}
	commentID, ok := getIDFromParam(c, "comment_id")
	if !ok {
		return
	}

	err := h.qaService.DeleteComment(c.Request.Context(), commentID, userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// --- Vote Handlers ---

func (h *QAHandler) UpvoteAnswer(c *gin.Context) {
	userID, ok := getAuthUserID(c)
	if !ok {
		return
	}
	answerID, ok := getIDFromParam(c, "answer_id")
	if !ok {
		return
	}

	err := h.qaService.UpvoteAnswer(c.Request.Context(), answerID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "点赞成功"})
}

func (h *QAHandler) DownvoteAnswer(c *gin.Context) {
	userID, ok := getAuthUserID(c)
	if !ok {
		return
	}
	answerID, ok := getIDFromParam(c, "answer_id")
	if !ok {
		return
	}

	err := h.qaService.DownvoteAnswer(c.Request.Context(), answerID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "取消点赞成功"})
}
