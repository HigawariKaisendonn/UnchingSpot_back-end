package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/higawarikaisendonn/unchingspot-backend/internal/model"
	"github.com/higawarikaisendonn/unchingspot-backend/internal/repository"
	"github.com/higawarikaisendonn/unchingspot-backend/internal/util"
)

var (
	// ErrEmailAlreadyExists はメールアドレスが既に登録されているエラー
	ErrEmailAlreadyExists = errors.New("email already exists")
	// ErrInvalidCredentials は認証情報が無効なエラー
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrUserNotFound はユーザーが見つからないエラー
	ErrUserNotFound = errors.New("user not found")
)

// AuthService は認証関連のビジネスロジックを提供します
type AuthService interface {
	SignUp(ctx context.Context, email, password, name string) (*model.User, error)
	Login(ctx context.Context, email, password string) (string, *model.User, error)
	ValidateToken(ctx context.Context, token string) (*model.User, error)
	TestConnection(ctx context.Context) error
}

// authServiceImpl はAuthServiceの実装
type authServiceImpl struct {
	userRepo repository.UserRepository
}

// NewAuthService は新しいAuthServiceインスタンスを作成します
func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authServiceImpl{
		userRepo: userRepo,
	}
}

// SignUp は新しいユーザーを登録します
// 要件: 1.1, 1.2, 1.3, 1.4, 1.5
func (s *authServiceImpl) SignUp(ctx context.Context, email, password, name string) (*model.User, error) {
	// メールアドレスの重複チェック（要件: 1.4）
	existingUser, err := s.userRepo.FindByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, ErrEmailAlreadyExists
	}

	// パスワードのハッシュ化（要件: 1.3）
	hashedPassword, err := util.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 新しいユーザーの作成（要件: 1.1, 1.2）
	user := &model.User{
		Name:     name,
		Email:    email,
		Password: hashedPassword,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		// データベースの一意制約違反の場合
		if isUniqueViolation(err) {
			return nil, ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 要件: 1.5 - 成功ステータスとユーザー情報を返す
	return user, nil
}

// Login はユーザーのログイン処理を行います
// 要件: 2.1, 2.2, 2.3, 2.4
func (s *authServiceImpl) Login(ctx context.Context, email, password string) (string, *model.User, error) {
	// ユーザーの検索（要件: 2.1）
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", nil, ErrInvalidCredentials
	}

	// パスワードの検証（要件: 2.1）
	if err := util.CheckPassword(password, user.Password); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	// JWTトークンの生成（要件: 2.2）
	token, err := util.GenerateToken(user.ID, user.Email)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// 要件: 2.3 - 認証トークンとユーザー情報を返す
	return token, user, nil
}

// ValidateToken はJWTトークンを検証し、ユーザー情報を返します
// 要件: 4.3, 5.1, 5.2, 5.3
func (s *authServiceImpl) ValidateToken(ctx context.Context, token string) (*model.User, error) {
	// トークンの検証
	claims, err := util.ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// ユーザーの取得
	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// TestConnection はデータベース接続をテストします
// 要件: 5.1, 5.2, 5.3
func (s *authServiceImpl) TestConnection(ctx context.Context) error {
	// 簡単なクエリを実行してデータベース接続を確認
	// ダミーのメールアドレスで検索を試みる（存在しなくてもOK）
	_, err := s.userRepo.FindByEmail(ctx, "test@connection.check")
	
	// ユーザーが見つからないエラーは正常（接続は成功している）
	if err != nil && !isNotFoundError(err) {
		return fmt.Errorf("database connection test failed: %w", err)
	}

	return nil
}

// isUniqueViolation はエラーが一意制約違反かどうかを判定します
func isUniqueViolation(err error) bool {
	// PostgreSQLの一意制約違反エラーコード: 23505
	return err != nil && (
		contains(err.Error(), "duplicate key") ||
		contains(err.Error(), "unique constraint") ||
		contains(err.Error(), "23505"))
}

// isNotFoundError はエラーがレコード未検出エラーかどうかを判定します
func isNotFoundError(err error) bool {
	return err != nil && contains(err.Error(), "not found")
}

// contains は文字列に部分文字列が含まれているかチェックします
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(findSubstring(s, substr) >= 0))
}

// findSubstring は文字列内の部分文字列の位置を返します
func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
