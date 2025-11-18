package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/higawarikaisendonn/unchingspot-backend/internal/model"
)

// userRepositoryImpl はUserRepositoryの実装
type userRepositoryImpl struct {
	db *sqlx.DB
}

// NewUserRepository は新しいUserRepositoryインスタンスを作成します
func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepositoryImpl{
		db: db,
	}
}

// Create は新しいユーザーをデータベースに作成します
// 要件: 1.1, 1.2
func (r *userRepositoryImpl) Create(ctx context.Context, user *model.User) error {
	// UUIDを生成
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	query := `
		INSERT INTO users (id, name, email, password, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		user.ID,
		user.Name,
		user.Email,
		user.Password,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// FindByEmail はメールアドレスでユーザーを検索します
// 要件: 2.1
func (r *userRepositoryImpl) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User

	query := `
		SELECT id, name, email, password, created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`

	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found with email: %s", email)
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	return &user, nil
}

// FindByID はIDでユーザーを検索します
// 要件: 4.1
func (r *userRepositoryImpl) FindByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User

	query := `
		SELECT id, name, email, password, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found with id: %s", id)
		}
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}

	return &user, nil
}

// Update はユーザー情報を更新します
// 要件: 4.1
func (r *userRepositoryImpl) Update(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users
		SET name = $1, email = $2, password = $3, updated_at = NOW()
		WHERE id = $4 AND deleted_at IS NULL
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		user.Name,
		user.Email,
		user.Password,
		user.ID,
	).Scan(&user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found or already deleted: %s", user.ID)
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}
