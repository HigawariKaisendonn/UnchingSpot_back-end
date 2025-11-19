package util

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrInvalidPassword はパスワード検証失敗エラーを表します
	ErrInvalidPassword = errors.New("invalid password")
)

const (
	// BcryptCost はbcryptのコスト値（2024年推奨値）
	BcryptCost = 12
)

// HashPassword はパスワードをbcryptでハッシュ化します
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword はパスワードとハッシュを比較して検証します
func CheckPassword(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidPassword
		}
		return err
	}
	return nil
}
