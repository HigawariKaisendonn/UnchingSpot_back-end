package middleware

import (
	"log"
	"net/http"
	"time"
)

// responseWriter はhttp.ResponseWriterをラップしてステータスコードとレスポンスサイズを記録します
type responseWriter struct {
	http.ResponseWriter
	statusCode  int
	size        int
	wroteHeader bool
}

// WriteHeader はステータスコードを記録します
func (rw *responseWriter) WriteHeader(statusCode int) {
	if !rw.wroteHeader {
		rw.statusCode = statusCode
		rw.wroteHeader = true
		rw.ResponseWriter.WriteHeader(statusCode)
	}
}

// Write はレスポンスボディを書き込み、サイズを記録します
func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// LoggerMiddleware はHTTPリクエストとレスポンスをログ出力するミドルウェアです
func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// リクエスト開始時刻を記録
		start := time.Now()

		// responseWriterでラップ
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // デフォルトは200
		}

		// 次のハンドラーを呼び出し
		next.ServeHTTP(rw, r)

		// リクエスト処理時間を計算
		duration := time.Since(start)

		// ログ出力
		log.Printf(
			"%s %s %d %s %dB",
			r.Method,
			r.RequestURI,
			rw.statusCode,
			duration,
			rw.size,
		)
	})
}
