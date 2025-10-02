package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestParseToken(t *testing.T) {
	secret := []byte("secret")
	claims := jwt.MapClaims{
		"user_id":  int64(42),
		"username": "alice",
		"exp":      time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	identity, err := ParseToken(tokenString, secret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if identity.UserID != 42 {
		t.Fatalf("expected userID 42, got %d", identity.UserID)
	}

	if identity.Username != "alice" {
		t.Fatalf("expected username alice, got %s", identity.Username)
	}

	if identity.Token != tokenString {
		t.Fatalf("expected raw token to be preserved")
	}

	// 测试 claims 字段
	if identity.Claims == nil {
		t.Fatalf("expected claims to be populated")
	}

	// 测试便捷方法
	if username, ok := identity.GetStringClaim("username"); !ok || username != "alice" {
		t.Fatalf("expected username claim to be alice, got %s", username)
	}

	if userID, ok := identity.GetInt64Claim("user_id"); !ok || userID != 42 {
		t.Fatalf("expected user_id claim to be 42, got %d", userID)
	}

	if !identity.HasClaim("exp") {
		t.Fatalf("expected exp claim to exist")
	}

	if identity.HasClaim("non_existent") {
		t.Fatalf("expected non_existent claim to not exist")
	}
}

func TestParseTokenMissingUserID(t *testing.T) {
	secret := []byte("secret")
	claims := jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	if _, err := ParseToken(tokenString, secret); err == nil {
		t.Fatalf("expected error when user_id is missing")
	}
}
