package auth

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// ErrInvalidToken 表示 token 无法解析或不合法。
	ErrInvalidToken = errors.New("invalid token")
	// ErrMissingUserID 表示 token 中缺少 user_id 声明。
	ErrMissingUserID = errors.New("token missing user_id claim")
)

// ParseToken 解析 JWT 并返回 Identity。
func ParseToken(tokenString string, secret []byte) (Identity, error) {
	if tokenString == "" {
		return Identity{}, ErrInvalidToken
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return Identity{}, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return Identity{}, ErrInvalidToken
	}

	rawUserID, ok := claims["user_id"]
	if !ok {
		return Identity{}, ErrMissingUserID
	}

	userID, err := toInt64(rawUserID)
	if err != nil {
		return Identity{}, ErrMissingUserID
	}

	username, _ := claims["username"].(string)

	return Identity{
		UserID:   userID,
		Username: username,
		Token:    tokenString,
		Claims:   claims,
	}, nil
}

func toInt64(value any) (int64, error) {
	switch v := value.(type) {
	case int64:
		return v, nil
	case int32:
		return int64(v), nil
	case int:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case float32:
		return int64(v), nil
	default:
		return 0, fmt.Errorf("unsupported type: %T", value)
	}
}
