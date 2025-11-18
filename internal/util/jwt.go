package util

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims はJWTトークンのクレームを表します
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

var (
	// ErrInvalidToken は無効なトークンエラーを表します
	ErrInvalidToken = errors.New("invalid token")
	// ErrExpiredToken は期限切れトークンエラーを表します
	ErrExpiredToken = errors.New("token has expired")
	// ErrMissingSecret はJWTシークレットが設定されていないエラーを表します
	ErrMissingSecret = errors.New("JWT_SECRET environment variable is not set")
	// ErrWeakSecret はJWTシークレットの強度が不足しているエラーを表します
	ErrWeakSecret = errors.New("JWT_SECRET must be at least 32 characters long")
)

const (
	// MinSecretLength はJWTシークレットの最小文字数
	MinSecretLength = 32
)

// GenerateToken はユーザーIDとメールアドレスからJWTトークンを生成します
func GenerateToken(userID, email string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", ErrMissingSecret
	}
	if len(secret) < MinSecretLength {
		return "", ErrWeakSecret
	}

	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateToken はJWTトークンを検証し、クレームを返します
func ValidateToken(tokenString string) (*JWTClaims, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, ErrMissingSecret
	}
	if len(secret) < MinSecretLength {
		return nil, ErrWeakSecret
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 署名方式の検証
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
