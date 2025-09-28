package auth

import (
	"context"

	"github.com/gin-gonic/gin"
)

// Identity 表示经过认证的用户身份信息，供各微服务共享。
type Identity struct {
	UserID   int64  // 用户唯一ID
	Username string // 用户名
	Token    string // 原始JWT，便于链路中重用
}

// contextKey 是内部使用的上下文键类型，避免与外部冲突。
type contextKey string

const (
	identityContextKey contextKey = "auth.identity"
	// ContextUserIDKey 与历史字符串键保持一致，方便旧代码读取。
	ContextUserIDKey   = "userID"
	ContextUsernameKey = "username"
)

// WithIdentity 将身份信息写入 Context。
func WithIdentity(ctx context.Context, identity Identity) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx = context.WithValue(ctx, identityContextKey, identity)
	ctx = context.WithValue(ctx, ContextUserIDKey, identity.UserID)
	if identity.Username != "" {
		ctx = context.WithValue(ctx, ContextUsernameKey, identity.Username)
	}
	return ctx
}

// FromContext 尝试从 Context 中读取 Identity。
func FromContext(ctx context.Context) (Identity, bool) {
	if ctx == nil {
		return Identity{}, false
	}
	value := ctx.Value(identityContextKey)
	if value == nil {
		return Identity{}, false
	}
	identity, ok := value.(Identity)
	if !ok {
		return Identity{}, false
	}
	return identity, true
}

const ginIdentityKey = "auth.identity"

// InjectIntoGin 将身份信息写入 gin.Context 与底层 *http.Request.Context。
func InjectIntoGin(c *gin.Context, identity Identity) {
	if c == nil {
		return
	}
	c.Set(ginIdentityKey, identity)
	c.Set(ContextUserIDKey, identity.UserID)
	if identity.Username != "" {
		c.Set(ContextUsernameKey, identity.Username)
	}
	// 同步写入 request context，便于 service 层读取。
	ctx := WithIdentity(c.Request.Context(), identity)
	c.Request = c.Request.WithContext(ctx)
}

// FromGinContext 读取 gin.Context 中的身份信息。
func FromGinContext(c *gin.Context) (Identity, bool) {
	if c == nil {
		return Identity{}, false
	}
	value, exists := c.Get(ginIdentityKey)
	if !exists {
		return Identity{}, false
	}
	identity, ok := value.(Identity)
	if !ok {
		return Identity{}, false
	}
	return identity, true
}
