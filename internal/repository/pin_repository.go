package repository

import (
	"context"

	"github.com/yourusername/unchingspot-backend/internal/model"
)

// PinRepository はピンデータアクセスのインターフェースを定義します
type PinRepository interface {
	Create(ctx context.Context, pin *model.Pin) error
	Update(ctx context.Context, pin *model.Pin) error
	FindByID(ctx context.Context, id string) (*model.Pin, error)
	FindByUserID(ctx context.Context, userID string) ([]*model.Pin, error)
	SoftDelete(ctx context.Context, id string) error
}
