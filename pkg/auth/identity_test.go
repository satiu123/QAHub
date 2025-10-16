package auth

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// TestContextIdentity 测试 WithIdentity 和 FromContext 函数。
func TestContextIdentity(t *testing.T) {
	identity := Identity{
		UserID:   1,
		Username: "testuser",
		Token:    "some.jwt.token",
	}

	// 场景1: 成功注入和提取
	ctx := WithIdentity(context.Background(), identity)
	retrieved, ok := FromContext(ctx)
	if !ok {
		t.Fatalf("期望在 context 中找到 identity，但没有找到")
	}
	if retrieved.UserID != identity.UserID {
		t.Errorf("期望 UserID 为 %d, 得到 %d", identity.UserID, retrieved.UserID)
	}
	if retrieved.Username != identity.Username {
		t.Errorf("期望 Username 为 %s, 得到 %s", identity.Username, retrieved.Username)
	}

	// 同时检查直接注入的值
	userID, _ := ctx.Value(ContextUserIDKey).(int64)
	if userID != identity.UserID {
		t.Errorf("期望 context 中 ContextUserIDKey 的值为 %d, 得到 %d", identity.UserID, userID)
	}

	// 场景2: context 中没有 identity
	_, ok = FromContext(context.Background())
	if ok {
		t.Fatalf("期望在空 context 中找不到 identity，但找到了")
	}

	// 场景3: nil context
	//nolint:staticcheck
	_, ok = FromContext(nil)
	if ok {
		t.Fatalf("期望在 nil context 中找不到 identity，但找到了")
	}

	// 场景4: context 中存在错误的类型
	wrongCtx := context.WithValue(context.Background(), identityContextKey, "not-an-identity")
	_, ok = FromContext(wrongCtx)
	if ok {
		t.Fatalf("期望在类型错误的 context 中找不到 identity，但找到了")
	}
}

// TestGinContextIdentity 测试 InjectIntoGin 和 FromGinContext 函数。
func TestGinContextIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// 为测试上下文附加一个真实的 http.Request，以避免 c.Request 为 nil
	c.Request = httptest.NewRequest("GET", "/", nil)

	identity := Identity{
		UserID:   2,
		Username: "ginuser",
	}

	// 场景1: 成功注入和提取
	InjectIntoGin(c, identity)
	retrieved, ok := FromGinContext(c)
	if !ok {
		t.Fatalf("期望在 gin context 中找到 identity，但没有找到")
	}
	if retrieved.UserID != identity.UserID {
		t.Errorf("期望 UserID 为 %d, 得到 %d", identity.UserID, retrieved.UserID)
	}
	if retrieved.Username != identity.Username {
		t.Errorf("期望 Username 为 %s, 得到 %s", identity.Username, retrieved.Username)
	}

	// 检查底层的 request context
	reqCtxRetrieved, ok := FromContext(c.Request.Context())
	if !ok {
		t.Fatalf("期望在底层的 request context 中找到 identity，但没有找到")
	}
	if reqCtxRetrieved.UserID != identity.UserID {
		t.Errorf("期望 request context 中的 UserID 为 %d, 得到 %d", identity.UserID, reqCtxRetrieved.UserID)
	}

	// 场景2: context 中没有 identity
	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	_, ok = FromGinContext(c)
	if ok {
		t.Fatalf("期望在空 gin context 中找不到 identity，但找到了")
	}

	// 场景3: nil context
	FromGinContext(nil) // 不应引发 panic
}

// TestIdentityClaims 测试 Identity 结构体上的声明辅助方法。
func TestIdentityClaims(t *testing.T) {
	identity := Identity{
		Claims: jwt.MapClaims{
			"sub":         "subject",
			"user_id":     float64(123), // JWT 中的数字通常解码为 float64
			"user_id_int": int64(456),
			"admin":       true,
		},
	}

	// 测试 GetStringClaim
	if val, ok := identity.GetStringClaim("sub"); !ok || val != "subject" {
		t.Errorf("GetStringClaim('sub') 失败: 期望 'subject', 得到 '%s'", val)
	}
	if _, ok := identity.GetStringClaim("admin"); ok {
		t.Error("GetStringClaim('admin') 对非字符串类型应失败")
	}

	// 测试 GetInt64Claim
	if val, ok := identity.GetInt64Claim("user_id"); !ok || val != 123 {
		t.Errorf("GetInt64Claim('user_id') 失败: 期望 123, 得到 '%d'", val)
	}
	if val, ok := identity.GetInt64Claim("user_id_int"); !ok || val != 456 {
		t.Errorf("GetInt64Claim('user_id_int') 失败: 期望 456, 得到 '%d'", val)
	}
	if _, ok := identity.GetInt64Claim("sub"); ok {
		t.Error("GetInt64Claim('sub') 对非数字类型应失败")
	}

	// 测试 HasClaim
	if !identity.HasClaim("admin") {
		t.Error("HasClaim('admin') 应为 true")
	}
	if identity.HasClaim("nonexistent") {
		t.Error("HasClaim('nonexistent') 应为 false")
	}

	// 测试 nil claims
	nilClaimsIdentity := Identity{}
	if _, ok := nilClaimsIdentity.GetClaim("any"); ok {
		t.Error("在 nil claims 上调用 GetClaim 应返回 false")
	}
	if nilClaimsIdentity.HasClaim("any") {
		t.Error("在 nil claims 上调用 HasClaim 应为 false")
	}
}
