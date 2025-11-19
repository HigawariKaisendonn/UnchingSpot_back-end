package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/higawarikaisendonn/unchingspot-backend/internal/util"
)

// contextKey はコンテキストキーの型
type contextKey string

const (
	// UserIDKey はコンテキストに保存されるユーザーIDのキー
	UserIDKey contextKey = "user_id"
	// UserEmailKey はコンテキストに保存されるユーザーメールのキー
	UserEmailKey contextKey = "user_email"
)

// AuthMiddleware はJWTトークンを検証し、ユーザー情報をコンテキストに設定するミドルウェアです
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Authorizationヘッダーを取得
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		// "Bearer "プレフィックスを削除
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			// "Bearer "プレフィックスがない場合
			respondError(w, http.StatusUnauthorized, "invalid authorization header format")
			return
		}

		// トークンを検証
		claims, err := util.ValidateToken(tokenString)
		if err != nil {
			respondError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		// ユーザー情報をコンテキストに設定
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserEmailKey, claims.Email)

		// 次のハンドラーを呼び出し
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserIDFromContext はコンテキストからユーザーIDを取得します
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

// GetUserEmailFromContext はコンテキストからユーザーメールを取得します
func GetUserEmailFromContext(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(UserEmailKey).(string)
	return email, ok
}

// respondError はエラーレスポンスを返します
func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(`{"error":{"message":"` + message + `"}}`))
}
