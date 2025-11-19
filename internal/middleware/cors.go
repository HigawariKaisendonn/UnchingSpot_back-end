package middleware

import (
	"net/http"
	"os"
)

// CORSMiddleware はCORS（Cross-Origin Resource Sharing）を処理するミドルウェアです
// Next.jsフロントエンドとの疎通を可能にします
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 環境変数からフロントエンドURLを取得（デフォルト: http://localhost:3000）
		// 本番環境ではフロントエンドの実際のURLを指定してください
		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "http://localhost:3000"
		}

		// CORSヘッダーを設定
		w.Header().Set("Access-Control-Allow-Origin", frontendURL)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// プリフライトリクエスト（OPTIONS）の処理
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// 次のハンドラーを呼び出し
		next.ServeHTTP(w, r)
	})
}
