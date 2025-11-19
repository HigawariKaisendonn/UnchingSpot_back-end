package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/higawarikaisendonn/unchingspot-backend/internal/model"
)

// pinRepositoryImpl はPinRepositoryの実装
type pinRepositoryImpl struct {
	db *sqlx.DB
}

// NewPinRepository は新しいPinRepositoryインスタンスを作成します
func NewPinRepository(db *sqlx.DB) PinRepository {
	return &pinRepositoryImpl{
		db: db,
	}
}

// Create は新しいPinをデータベースに作成します
// PostGISのST_MakePointを使用して位置情報を保存
// 要件: 6.1, 6.2, 6.3, 6.4
func (r *pinRepositoryImpl) Create(ctx context.Context, pin *model.Pin) error {
	// UUIDを生成
	if pin.ID == "" {
		pin.ID = uuid.New().String()
	}

	query := `
		INSERT INTO pins (id, name, user_id, location, created_at, edit_at)
		VALUES ($1, $2, $3, ST_SetSRID(ST_MakePoint($4, $5), 4326), NOW(), NOW())
		RETURNING id, created_at, edit_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		pin.ID,
		pin.Name,
		pin.UserID,
		pin.Longitude, // ST_MakePoint(longitude, latitude)の順序
		pin.Latitude,
	).Scan(&pin.ID, &pin.CreatedAt, &pin.EditedAt)

	if err != nil {
		return fmt.Errorf("failed to create pin: %w", err)
	}

	return nil
}

// Update はPinの情報を更新します
// PostGISのST_MakePointを使用して位置情報を更新
// 要件: 7.1, 7.2, 7.3
func (r *pinRepositoryImpl) Update(ctx context.Context, pin *model.Pin) error {
	query := `
		UPDATE pins
		SET name = $1, location = ST_SetSRID(ST_MakePoint($2, $3), 4326), edit_at = NOW()
		WHERE id = $4 AND deleted_at IS NULL
		RETURNING edit_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		pin.Name,
		pin.Longitude, // ST_MakePoint(longitude, latitude)の順序
		pin.Latitude,
		pin.ID,
	).Scan(&pin.EditedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("pin not found or already deleted: %s", pin.ID)
		}
		return fmt.Errorf("failed to update pin: %w", err)
	}

	return nil
}

// FindByID はIDでPinを検索します
// PostGISのST_X/ST_Yを使用して緯度経度を抽出
// 要件: 6.1
func (r *pinRepositoryImpl) FindByID(ctx context.Context, id string) (*model.Pin, error) {
	var pin model.Pin

	query := `
		SELECT
			id,
			name,
			user_id,
			ST_X(location) as longitude,
			ST_Y(location) as latitude,
			created_at,
			edit_at,
			deleted_at
		FROM pins
		WHERE id = $1 AND deleted_at IS NULL
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&pin.ID,
		&pin.Name,
		&pin.UserID,
		&pin.Longitude,
		&pin.Latitude,
		&pin.CreatedAt,
		&pin.EditedAt,
		&pin.DeletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("pin not found with id: %s", id)
		}
		return nil, fmt.Errorf("failed to find pin by id: %w", err)
	}

	return &pin, nil
}

// FindByUserID はユーザーIDで全てのPinを検索します
// PostGISのST_X/ST_Yを使用して緯度経度を抽出
// 要件: 6.1
func (r *pinRepositoryImpl) FindByUserID(ctx context.Context, userID string) ([]*model.Pin, error) {
	var pins []*model.Pin

	query := `
		SELECT
			id,
			name,
			user_id,
			ST_X(location) as longitude,
			ST_Y(location) as latitude,
			created_at,
			edit_at,
			deleted_at
		FROM pins
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find pins by user id: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var pin model.Pin
		err := rows.Scan(
			&pin.ID,
			&pin.Name,
			&pin.UserID,
			&pin.Longitude,
			&pin.Latitude,
			&pin.CreatedAt,
			&pin.EditedAt,
			&pin.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pin: %w", err)
		}
		pins = append(pins, &pin)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating pins: %w", err)
	}

	return pins, nil
}

// SoftDelete はPinを論理削除します
// 要件: 7.1
func (r *pinRepositoryImpl) SoftDelete(ctx context.Context, id string) error {
	query := `
		UPDATE pins
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete pin: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("pin not found or already deleted: %s", id)
	}

	return nil
}
