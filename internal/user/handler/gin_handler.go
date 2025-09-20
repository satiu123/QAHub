package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"qahub/internal/user/model"
	"qahub/internal/user/service"
	"qahub/pkg/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{
		userService: svc,
	}
}

// --- Request & Response Structs ---

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=72"`
	Bio      string `json:"bio,omitempty"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type UpdateProfileRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Email    string `json:"email" binding:"required,email"`
	Bio      string `json:"bio,omitempty"`
}

// --- Handler Methods ---

func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userResponse, err := h.userService.Register(c.Request.Context(), req.Username, req.Email, req.Bio, req.Password)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, userResponse)
}

func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.userService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{Token: token})
}

func (h *UserHandler) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的授权标头"})
		return
	}
	tokenString := parts[1]

	var jwtSecret = []byte(config.Conf.Services.UserService.JWTSecret)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("非预期的签名方法: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "登出成功"})
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		err := h.userService.Logout(c.Request.Context(), tokenString, claims)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "登出成功"})
}

// ValidateToken 是为 Nginx auth_request 设计的端点
func (h *UserHandler) ValidateToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "无效的授权标头"})
		return
	}
	tokenString := parts[1]

	userID, err := h.userService.ValidateToken(c.Request.Context(), tokenString)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// 验证成功，将 User-ID 放入响应头，供 Nginx 读取
	c.Header("X-User-ID", strconv.FormatInt(userID, 10))
	c.JSON(http.StatusOK, gin.H{"message": "valid"})
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	userResponse, err := h.userService.GetUserProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, userResponse)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	authUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取认证用户信息"})
		return
	}

	idStr := c.Param("id")
	targetUserID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	if authUserID.(int64) != targetUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "没有权限执行此操作"})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updateModel := &model.User{
		ID:       targetUserID,
		Username: req.Username,
		Email:    req.Email,
		Bio:      req.Bio,
	}

	err = h.userService.UpdateUserProfile(c.Request.Context(), updateModel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户信息更新成功"})
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	authUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取认证用户信息"})
		return
	}

	idStr := c.Param("id")
	targetUserID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	if authUserID.(int64) != targetUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "没有权限执行此操作"})
		return
	}

	err = h.userService.DeleteUser(c.Request.Context(), targetUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户删除成功"})
}
