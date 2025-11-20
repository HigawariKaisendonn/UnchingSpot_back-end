package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/higawarikaisendonn/unchingspot-backend/internal/model"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// connectRepositoryImpl はConnectRepositoryの実装
type connectRepositoryImpl struct {
	db *sqlx.DB
}

// NewConnectRepository は新しいConnectRepositoryインスタンスを作成します
func NewConnectRepository(db *sqlx.DB) ConnectRepository {
	return &connectRepositoryImpl{
		db: db,
	}
}

// Create は新しい接続をデータベースに作成します
// 要件: 8.1, 8.2, 8.3
func (r *connectRepositoryImpl) Create(ctx context.Context, connect *model.Connect) error {
	// UUIDを生成
	if connect.ID == "" {
		connect.ID = uuid.New().String()
	}

	query := `
		INSERT INTO connect (id, user_id, name, pins_id_1, pins_id_2, show)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		connect.ID,
		connect.UserID,
		connect.Name,
		connect.PinID1,
		pq.Array(connect.PinID2),
		connect.Show,
	).Scan(&connect.ID)

	if err != nil {
		return fmt.Errorf("failed to create connect: %w", err)
	}

	return nil
}

// Update は接続情報を更新します
// 要件: 9.1, 9.2
func (r *connectRepositoryImpl) Update(ctx context.Context, connect *model.Connect) error {
	query := `
		UPDATE connect
		SET name = $1, pins_id_1 = $2, pins_id_2 = $3, show = $4
		WHERE id = $5
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		connect.Name,
		connect.PinID1,
		pq.Array(connect.PinID2),
		connect.Show,
		connect.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update connect: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("connect not found: %s", connect.ID)
	}

	return nil
}

// FindByID はIDで接続を検索します
// 要件: 8.1, 9.1
func (r *connectRepositoryImpl) FindByID(ctx context.Context, id string) (*model.Connect, error) {
	var connect model.Connect

	query := `
		SELECT id, user_id, name, pins_id_1, pins_id_2, show
		FROM connect
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &connect, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("connect not found with id: %s", id)
		}
		return nil, fmt.Errorf("failed to find connect by id: %w", err)
	}

	return &connect, nil
}

// FindByUserID はユーザーIDで接続一覧を検索します
// 要件: 8.1, 9.1
func (r *connectRepositoryImpl) FindByUserID(ctx context.Context, userID string) ([]*model.Connect, error) {
	var connects []*model.Connect

	query := `
		SELECT id, user_id, name, pins_id_1, pins_id_2, show
		FROM connect
		WHERE user_id = $1
		ORDER BY id
	`

	err := r.db.SelectContext(ctx, &connects, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find connects by user id: %w", err)
	}

	return connects, nil
}

// Delete は接続を削除します
// 要件: 9.1
func (r *connectRepositoryImpl) Delete(ctx context.Context, id string) error {
	query := `
		DELETE FROM connect
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete connect: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("connect not found: %s", id)
	}

	return nil
}
