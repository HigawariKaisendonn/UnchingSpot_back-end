package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

var (
	// ErrInvalidJSON はJSON解析エラーを表します
	ErrInvalidJSON = errors.New("invalid JSON format")
	// ErrValidationFailed はバリデーション失敗エラーを表します
	ErrValidationFailed = errors.New("validation failed")
)

// emailの正規表現パターン
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ParseJSONBody はリクエストボディをJSONとしてパースします
func ParseJSONBody(r *http.Request, v interface{}) error {
	if r.Body == nil {
		return ErrInvalidJSON
	}
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // 未知のフィールドを許可しない

	if err := decoder.Decode(v); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	return nil
}

// ValidateEmail はメールアドレスの形式を検証します
func ValidateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}
	return nil
}

// ValidatePassword はパスワードの強度を検証します
func ValidatePassword(password string) error {
	if password == "" {
		return errors.New("password is required")
	}
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	return nil
}

// ValidateRequired は必須フィールドを検証します
func ValidateRequired(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	return nil
}

// ValidateLatitude は緯度の範囲を検証します
func ValidateLatitude(lat float64) error {
	if lat < -90 || lat > 90 {
		return errors.New("latitude must be between -90 and 90")
	}
	return nil
}

// ValidateLongitude は経度の範囲を検証します
func ValidateLongitude(lng float64) error {
	if lng < -180 || lng > 180 {
		return errors.New("longitude must be between -180 and 180")
	}
	return nil
}

// ValidateUUID はUUID形式を検証します（簡易版）
func ValidateUUID(id string) error {
	if id == "" {
		return errors.New("id is required")
	}
	// UUID v4の基本的な形式チェック
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	if !uuidRegex.MatchString(id) {
		return errors.New("invalid UUID format")
	}
	return nil
}
