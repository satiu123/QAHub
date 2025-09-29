package handler

import (
	"net/http"

	"qahub/search-service/internal/service"

	"github.com/gin-gonic/gin"
)

// Handler 结构体持有 service 的引用
type Handler struct {
	service service.SearchService
}

// New 函数创建一个新的 Handler 实例
func New(s service.SearchService) *Handler {
	return &Handler{service: s}
}

// Search 是处理搜索请求的 Gin handler
func (h *Handler) Search(c *gin.Context) {
	// 从查询参数中获取搜索关键词
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "查询参数 'q' 不能为空"})
		return
	}

	// 调用 service 层执行搜索
	results, err := h.service.SearchQuestions(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "执行搜索时发生内部错误"})
		return
	}

	// 返回搜索结果
	c.JSON(http.StatusOK, results)
}
