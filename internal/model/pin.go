package model

import "time"

// Pin はトイレの位置マーカーを表します
type Pin struct {
	ID        string     `db:"id" json:"id"`
	Name      string     `db:"name" json:"name"`
	UserID    string     `db:"user_id" json:"user_id"`
	Latitude  float64    `json:"latitude"`  // PostGISから抽出
	Longitude float64    `json:"longitude"` // PostGISから抽出
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	EditedAt  time.Time  `db:"edit_ad" json:"edited_at"`
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}
