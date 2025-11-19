package database

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/higawarikaisendonn/unchingspot-backend/internal/model"
	"golang.org/x/crypto/bcrypt"
)

// TestHelper はテストデータの作成を支援する構造体
type TestHelper struct {
	DB *TestDB
}

// NewTestHelper は新しいTestHelperを作成します
func NewTestHelper(db *TestDB) *TestHelper {
	return &TestHelper{DB: db}
}

// CreateTestUser はテスト用のユーザーを作成します
func (h *TestHelper) CreateTestUser(email, password, name string) (*model.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		ID:        uuid.New().String(),
		Email:     email,
		Password:  string(hashedPassword),
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	query := `
		INSERT INTO users (id, email, password, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err = h.DB.DB.ExecContext(context.Background(), query,
		user.ID, user.Email, user.Password, user.Name, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// CreateTestPin はテスト用のPinを作成します
func (h *TestHelper) CreateTestPin(userID, name string, lat, lng float64) (*model.Pin, error) {
	pin := &model.Pin{
		ID:        uuid.New().String(),
		Name:      name,
		UserID:    userID,
		Latitude:  lat,
		Longitude: lng,
		CreatedAt: time.Now(),
		EditedAt:  time.Now(),
	}

	query := `
		INSERT INTO pins (id, name, user_id, location, created_at, edit_at)
		VALUES ($1, $2, $3, ST_SetSRID(ST_MakePoint($4, $5), 4326), $6, $7)
	`

	_, err := h.DB.DB.ExecContext(context.Background(), query,
		pin.ID, pin.Name, pin.UserID, pin.Longitude, pin.Latitude, pin.CreatedAt, pin.EditedAt)
	if err != nil {
		return nil, err
	}

	return pin, nil
}

// CreateTestConnect はテスト用のConnectを作成します
func (h *TestHelper) CreateTestConnect(userID, pinID1, pinID2 string, show bool) (*model.Connect, error) {
	connect := &model.Connect{
		ID:     uuid.New().String(),
		UserID: userID,
		PinID1: pinID1,
		PinID2: pinID2,
		Show:   show,
	}

	query := `
		INSERT INTO connect (id, user_id, pins_id_1, pins_id_2, show)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := h.DB.DB.ExecContext(context.Background(), query,
		connect.ID, connect.UserID, connect.PinID1, connect.PinID2, connect.Show)
	if err != nil {
		return nil, err
	}

	return connect, nil
}

// GetUserByID はIDでユーザーを取得します
func (h *TestHelper) GetUserByID(id string) (*model.User, error) {
	var user model.User
	query := `SELECT id, email, password, name, created_at, updated_at, deleted_at FROM users WHERE id = $1`
	err := h.DB.DB.GetContext(context.Background(), &user, query, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetPinByID はIDでPinを取得します
func (h *TestHelper) GetPinByID(id string) (*model.Pin, error) {
	var pin model.Pin
	query := `
		SELECT
			id, name, user_id,
			ST_Y(location) as latitude,
			ST_X(location) as longitude,
			created_at, edit_at, deleted_at
		FROM pins
		WHERE id = $1
	`
	err := h.DB.DB.GetContext(context.Background(), &pin, query, id)
	if err != nil {
		return nil, err
	}
	return &pin, nil
}

// GetConnectByID はIDでConnectを取得します
func (h *TestHelper) GetConnectByID(id string) (*model.Connect, error) {
	var connect model.Connect
	query := `SELECT id, user_id, pins_id_1, pins_id_2, show FROM connect WHERE id = $1`
	err := h.DB.DB.GetContext(context.Background(), &connect, query, id)
	if err != nil {
		return nil, err
	}
	return &connect, nil
}
