package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/higawarikaisendonn/unchingspot-backend/internal/model"
	"github.com/higawarikaisendonn/unchingspot-backend/internal/repository"
)

var (
	// ErrPinNotFound はPinが見つからないエラー
	ErrPinNotFound = errors.New("pin not found")
	// ErrUnauthorizedPinAccess はPinへの不正アクセスエラー
	ErrUnauthorizedPinAccess = errors.New("unauthorized access to pin")
	// ErrInvalidCoordinates は無効な座標エラー
	ErrInvalidCoordinates = errors.New("invalid coordinates")
)

// PinService はPin関連のビジネスロジックを提供します
type PinService interface {
	CreatePin(ctx context.Context, userID string, name string, lat, lng float64) (*model.Pin, error)
	UpdatePin(ctx context.Context, pinID, userID string, name string, lat, lng float64) (*model.Pin, error)
	GetPin(ctx context.Context, pinID string) (*model.Pin, error)
	GetPinsByUser(ctx context.Context, userID string) ([]*model.Pin, error)
	DeletePin(ctx context.Context, pinID, userID string) error
}

// pinServiceImpl はPinServiceの実装
type pinServiceImpl struct {
	pinRepo repository.PinRepository
}

// NewPinService は新しいPinServiceインスタンスを作成します
func NewPinService(pinRepo repository.PinRepository) PinService {
	return &pinServiceImpl{
		pinRepo: pinRepo,
	}
}

// CreatePin は新しいPinを作成します
// 要件: 6.1, 6.2, 6.3, 6.4, 6.5
func (s *pinServiceImpl) CreatePin(ctx context.Context, userID string, name string, lat, lng float64) (*model.Pin, error) {
	// 座標の検証
	if !isValidCoordinates(lat, lng) {
		return nil, ErrInvalidCoordinates
	}

	// 新しいPinの作成（要件: 6.1, 6.2, 6.3, 6.4）
	now := time.Now()
	pin := &model.Pin{
		Name:      name,
		UserID:    userID,
		Latitude:  lat,
		Longitude: lng,
		CreatedAt: now,
		EditedAt:  now,
	}

	if err := s.pinRepo.Create(ctx, pin); err != nil {
		return nil, fmt.Errorf("failed to create pin: %w", err)
	}

	// 要件: 6.5 - 作成されたPin情報を返す
	return pin, nil
}

// UpdatePin は既存のPinを更新します
// 要件: 7.1, 7.2, 7.3, 7.4, 7.5
func (s *pinServiceImpl) UpdatePin(ctx context.Context, pinID, userID string, name string, lat, lng float64) (*model.Pin, error) {
	// 座標の検証
	if !isValidCoordinates(lat, lng) {
		return nil, ErrInvalidCoordinates
	}

	// 既存のPinを取得（要件: 7.1）
	pin, err := s.pinRepo.FindByID(ctx, pinID)
	if err != nil {
		return nil, ErrPinNotFound
	}

	// 所有権の確認（要件: 7.4）
	if pin.UserID != userID {
		return nil, ErrUnauthorizedPinAccess
	}

	// Pinの更新（要件: 7.1, 7.2, 7.3）
	pin.Name = name
	pin.Latitude = lat
	pin.Longitude = lng
	pin.EditedAt = time.Now()

	if err := s.pinRepo.Update(ctx, pin); err != nil {
		return nil, fmt.Errorf("failed to update pin: %w", err)
	}

	// 要件: 7.5 - 更新されたPin情報を返す
	return pin, nil
}

// GetPin は指定されたIDのPinを取得します
// 要件: 6.1, 7.1
func (s *pinServiceImpl) GetPin(ctx context.Context, pinID string) (*model.Pin, error) {
	pin, err := s.pinRepo.FindByID(ctx, pinID)
	if err != nil {
		return nil, ErrPinNotFound
	}

	return pin, nil
}

// GetPinsByUser は指定されたユーザーの全Pinを取得します
// 要件: 6.1, 7.1
func (s *pinServiceImpl) GetPinsByUser(ctx context.Context, userID string) ([]*model.Pin, error) {
	pins, err := s.pinRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pins: %w", err)
	}

	// 結果が空の場合は空のスライスを返す
	if pins == nil {
		pins = []*model.Pin{}
	}

	return pins, nil
}

// DeletePin は指定されたPinを削除します（ソフトデリート）
// 要件: 7.1, 7.4, 7.5
func (s *pinServiceImpl) DeletePin(ctx context.Context, pinID, userID string) error {
	// 既存のPinを取得
	pin, err := s.pinRepo.FindByID(ctx, pinID)
	if err != nil {
		return ErrPinNotFound
	}

	// 所有権の確認（要件: 7.4）
	if pin.UserID != userID {
		return ErrUnauthorizedPinAccess
	}

	// ソフトデリートの実行
	if err := s.pinRepo.SoftDelete(ctx, pinID); err != nil {
		return fmt.Errorf("failed to delete pin: %w", err)
	}

	return nil
}

// isValidCoordinates は座標が有効範囲内かチェックします
func isValidCoordinates(lat, lng float64) bool {
	return lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180
}
