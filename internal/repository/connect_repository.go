package repository

import (
	"context"

	"github.com/yourusername/unchingspot/internal/model"
)

// ConnectRepository は接続データアクセスのインターフェースを定義します
type ConnectRepository interface {
	Create(ctx context.Context, connect *model.Connect) error
	Update(ctx context.Context, connect *model.Connect) error
	FindByID(ctx context.Context, id string) (*model.Connect, error)
	FindByUserID(ctx context.Context, userID string) ([]*model.Connect, error)
	Delete(ctx context.Context, id string) error
}
