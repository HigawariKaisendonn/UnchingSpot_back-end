package util

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse はエラーレスポンスの構造を表します
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail はエラーの詳細情報を表します
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// SuccessResponse は成功レスポンスの構造を表します
type SuccessResponse struct {
	Data interface{} `json:"data,omitempty"`
}

// エラーコード定数
const (
	ErrCodeInvalidInput   = "INVALID_INPUT"
	ErrCodeUnauthorized   = "UNAUTHORIZED"
	ErrCodeForbidden      = "FORBIDDEN"
	ErrCodeNotFound       = "NOT_FOUND"
	ErrCodeConflict       = "CONFLICT"
	ErrCodeInternalServer = "INTERNAL_SERVER_ERROR"
	ErrCodeDatabaseError  = "DATABASE_ERROR"
)

// RespondJSON はJSON形式で成功レスポンスを返します
func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data != nil {
		// データを直接エンコード（ラップしない）
		if err := json.NewEncoder(w).Encode(data); err != nil {
			// エンコードエラーの場合はログに記録（実装は後で追加可能）
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}

// RespondError はJSON形式でエラーレスポンスを返します
func RespondError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		// エンコードエラーの場合は標準エラーレスポンスを返す
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
	}
}

// RespondValidationError はバリデーションエラーレスポンスを返します
func RespondValidationError(w http.ResponseWriter, message string) {
	RespondError(w, http.StatusBadRequest, ErrCodeInvalidInput, message)
}

// RespondUnauthorized は認証エラーレスポンスを返します
func RespondUnauthorized(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Unauthorized"
	}
	RespondError(w, http.StatusUnauthorized, ErrCodeUnauthorized, message)
}

// RespondForbidden は権限エラーレスポンスを返します
func RespondForbidden(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Forbidden"
	}
	RespondError(w, http.StatusForbidden, ErrCodeForbidden, message)
}

// RespondNotFound はリソース未検出エラーレスポンスを返します
func RespondNotFound(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Resource not found"
	}
	RespondError(w, http.StatusNotFound, ErrCodeNotFound, message)
}

// RespondConflict は競合エラーレスポンスを返します
func RespondConflict(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Resource conflict"
	}
	RespondError(w, http.StatusConflict, ErrCodeConflict, message)
}

// RespondInternalError は内部サーバーエラーレスポンスを返します
func RespondInternalError(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Internal server error"
	}
	RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, message)
}

// RespondDatabaseError はデータベースエラーレスポンスを返します
func RespondDatabaseError(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Database error"
	}
	RespondError(w, http.StatusInternalServerError, ErrCodeDatabaseError, message)
}
