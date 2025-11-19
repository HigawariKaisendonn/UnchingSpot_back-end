package model

import "github.com/lib/pq"

// Connect は開始点と複数の中間点を持つ図形の接続を表します
// PinID1は開始点と終了点を意味し、PinID2は複数の中間点を保持します
type Connect struct {
	ID     string         `db:"id" json:"id"`
	UserID string         `db:"user_id" json:"user_id"`
	Name   string         `db:"name" json:"name"`
	PinID1 string         `db:"pins_id_1" json:"pin_id_1"`
	PinID2 pq.StringArray `db:"pins_id_2" json:"pin_id_2"`
	Show   bool           `db:"show" json:"show"`
}
