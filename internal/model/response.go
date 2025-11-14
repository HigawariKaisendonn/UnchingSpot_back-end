package model

// AuthResponse はトークンとユーザー情報を含む認証レスポンスを表します
type AuthResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

// ErrorResponse はエラーレスポンスを表します
type ErrorResponse struct {
	Error *AppError `json:"error"`
}

// AppError はアプリケーションエラーの詳細を表します
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

// エラーコード
const (
	ErrCodeInvalidInput   = "INVALID_INPUT"
	ErrCodeUnauthorized   = "UNAUTHORIZED"
	ErrCodeForbidden      = "FORBIDDEN"
	ErrCodeNotFound       = "NOT_FOUND"
	ErrCodeConflict       = "CONFLICT"
	ErrCodeInternalServer = "INTERNAL_SERVER_ERROR"
	ErrCodeDatabaseError  = "DATABASE_ERROR"
)
