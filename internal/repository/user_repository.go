package repository

import (
	"context"

	"github.com/higawarikaisendonn/unchingspot-backend/internal/model"
)

// UserRepository はユーザーデータアクセスのインターフェースを定義します
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
}
