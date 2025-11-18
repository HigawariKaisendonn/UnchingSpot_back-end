package model

// Connect は2つのピン間の接続を表します
type Connect struct {
	ID     string `db:"id" json:"id"`
	UserID string `db:"user_id" json:"user_id"`
	PinID1 string `db:"pins_id_1" json:"pin_id_1"`
	PinID2 string `db:"pins_id_2" json:"pin_id_2"`
	Show   bool   `db:"show" json:"show"`
}
