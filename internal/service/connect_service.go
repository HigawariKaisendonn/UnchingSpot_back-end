package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/higawarikaisendonn/unchingspot-backend/internal/model"
	"github.com/higawarikaisendonn/unchingspot-backend/internal/repository"
)

var (
	// ErrConnectNotFound はConnectが見つからないエラー
	ErrConnectNotFound = errors.New("connect not found")
	// ErrUnauthorizedConnectAccess はConnectへの不正アクセスエラー
	ErrUnauthorizedConnectAccess = errors.New("unauthorized access to connect")
	// ErrPinNotExist は指定されたPinが存在しないエラー
	ErrPinNotExist = errors.New("specified pin does not exist")
	// ErrInvalidPinIDs は無効なPin IDエラー
	ErrInvalidPinIDs = errors.New("invalid pin IDs")
)

// ConnectService はConnect関連のビジネスロジックを提供します
type ConnectService interface {
	CreateConnect(ctx context.Context, userID, name, pinID1 string, pinID2 []string, show bool) (*model.Connect, error)
	UpdateConnect(ctx context.Context, connectID, userID, name, pinID1 string, pinID2 []string, show bool) (*model.Connect, error)
	GetConnectsByUser(ctx context.Context, userID string) ([]*model.Connect, error)
	DeleteConnect(ctx context.Context, connectID, userID string) error
}

// connectServiceImpl はConnectServiceの実装
type connectServiceImpl struct {
	connectRepo repository.ConnectRepository
	pinRepo     repository.PinRepository
}

// NewConnectService は新しいConnectServiceインスタンスを作成します
func NewConnectService(connectRepo repository.ConnectRepository, pinRepo repository.PinRepository) ConnectService {
	return &connectServiceImpl{
		connectRepo: connectRepo,
		pinRepo:     pinRepo,
	}
}

// CreateConnect は新しいConnectを作成します
// 要件: 8.1, 8.2, 8.3, 8.4, 8.5, 8.6, 8.7
func (s *connectServiceImpl) CreateConnect(ctx context.Context, userID, name, pinID1 string, pinID2 []string, show bool) (*model.Connect, error) {
	// Pin IDの検証
	if pinID1 == "" || len(pinID2) == 0 {
		return nil, ErrInvalidPinIDs
	}

	// Pin1の存在確認（要件: 8.6）
	pin1, err := s.pinRepo.FindByID(ctx, pinID1)
	if err != nil || pin1 == nil {
		return nil, ErrPinNotExist
	}

	// Pin2の各IDの存在確認（要件: 8.6）
	for _, pid := range pinID2 {
		pin, err := s.pinRepo.FindByID(ctx, pid)
		if err != nil || pin == nil {
			return nil, ErrPinNotExist
		}
	}

	// 新しいConnectの作成（要件: 8.1, 8.2, 8.3, 8.4, 8.5）
	connect := &model.Connect{
		UserID: userID,
		Name:   name,
		PinID1: pinID1,
		PinID2: pinID2,
		Show:   show,
	}

	if err := s.connectRepo.Create(ctx, connect); err != nil {
		return nil, fmt.Errorf("failed to create connect: %w", err)
	}

	// 要件: 8.7 - 作成されたConnect情報を返す
	return connect, nil
}

// UpdateConnect は既存のConnectを更新します
// 要件: 9.1, 9.2, 9.3, 9.4, 9.5, 9.6
func (s *connectServiceImpl) UpdateConnect(ctx context.Context, connectID, userID, name, pinID1 string, pinID2 []string, show bool) (*model.Connect, error) {
	// 既存のConnectを取得（要件: 9.1）
	connect, err := s.connectRepo.FindByID(ctx, connectID)
	if err != nil {
		return nil, ErrConnectNotFound
	}

	// 所有権の確認（要件: 9.4）
	if connect.UserID != userID {
		return nil, ErrUnauthorizedConnectAccess
	}

	// 名前の更新
	if name != "" {
		connect.Name = name
	}

	// Pin IDが指定されている場合は存在確認（要件: 9.5）
	if pinID1 != "" {
		pin1, err := s.pinRepo.FindByID(ctx, pinID1)
		if err != nil || pin1 == nil {
			return nil, ErrPinNotExist
		}
		connect.PinID1 = pinID1
	}

	if len(pinID2) > 0 {
		// Pin2の各IDの存在確認
		for _, pid := range pinID2 {
			pin, err := s.pinRepo.FindByID(ctx, pid)
			if err != nil || pin == nil {
				return nil, ErrPinNotExist
			}
		}
		connect.PinID2 = pinID2
	}

	// showフィールドの更新（要件: 9.2, 9.3）
	connect.Show = show

	// Connectの更新（要件: 9.1）
	if err := s.connectRepo.Update(ctx, connect); err != nil {
		return nil, fmt.Errorf("failed to update connect: %w", err)
	}

	// 要件: 9.6 - 更新されたConnect情報を返す
	return connect, nil
}

// GetConnectsByUser は指定されたユーザーの全Connectを取得します
// 要件: 8.1, 9.1
func (s *connectServiceImpl) GetConnectsByUser(ctx context.Context, userID string) ([]*model.Connect, error) {
	connects, err := s.connectRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get connects: %w", err)
	}

	// 結果が空の場合は空のスライスを返す
	if connects == nil {
		connects = []*model.Connect{}
	}

	return connects, nil
}

// DeleteConnect は指定されたConnectを削除します
// 要件: 9.1, 9.4, 9.6
func (s *connectServiceImpl) DeleteConnect(ctx context.Context, connectID, userID string) error {
	// 既存のConnectを取得
	connect, err := s.connectRepo.FindByID(ctx, connectID)
	if err != nil {
		return ErrConnectNotFound
	}

	// 所有権の確認（要件: 9.4）
	if connect.UserID != userID {
		return ErrUnauthorizedConnectAccess
	}

	// Connectの削除
	if err := s.connectRepo.Delete(ctx, connectID); err != nil {
		return fmt.Errorf("failed to delete connect: %w", err)
	}

	return nil
}
