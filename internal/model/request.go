package model

// SignUpRequest はユーザー登録リクエストを表します
type SignUpRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"required"`
}

// LoginRequest はユーザーログインリクエストを表します
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// CreatePinRequest はピン作成リクエストを表します
type CreatePinRequest struct {
	Name      string  `json:"name" validate:"required"`
	Latitude  float64 `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude float64 `json:"longitude" validate:"required,min=-180,max=180"`
}

// UpdatePinRequest はピン更新リクエストを表します
type UpdatePinRequest struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude" validate:"min=-90,max=90"`
	Longitude float64 `json:"longitude" validate:"min=-180,max=180"`
}

// CreateConnectRequest は接続作成リクエストを表します
type CreateConnectRequest struct {
	PinID1 string `json:"pin_id_1" validate:"required,uuid"`
	PinID2 string `json:"pin_id_2" validate:"required,uuid"`
	Show   bool   `json:"show"`
}

// UpdateConnectRequest は接続更新リクエストを表します
type UpdateConnectRequest struct {
	PinID1 string `json:"pin_id_1" validate:"uuid"`
	PinID2 string `json:"pin_id_2" validate:"uuid"`
	Show   *bool  `json:"show"`
}
