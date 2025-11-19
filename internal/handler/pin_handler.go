package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/higawarikaisendonn/unchingspot-backend/internal/model"
	"github.com/higawarikaisendonn/unchingspot-backend/internal/service"
	"github.com/higawarikaisendonn/unchingspot-backend/internal/util"
)

// PinHandler はPin関連のHTTPハンドラーを提供します
type PinHandler struct {
	pinService service.PinService
}

// NewPinHandler は新しいPinHandlerインスタンスを作成します
func NewPinHandler(pinService service.PinService) *PinHandler {
	return &PinHandler{
		pinService: pinService,
	}
}

// CreatePin はPinを作成します
// POST /api/pins
// 要件: 6.1, 6.5
func (h *PinHandler) CreatePin(w http.ResponseWriter, r *http.Request) {
	// コンテキストからユーザーIDを取得（認証ミドルウェアで設定される）
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		util.RespondUnauthorized(w, "Unauthorized")
		return
	}

	// リクエストボディのパース
	var req model.CreatePinRequest
	if err := util.ParseJSONBody(r, &req); err != nil {
		util.RespondValidationError(w, "Invalid request body")
		return
	}

	// バリデーション
	if err := util.ValidateRequired(req.Name, "name"); err != nil {
		util.RespondValidationError(w, err.Error())
		return
	}
	if err := util.ValidateCoordinates(req.Latitude, req.Longitude); err != nil {
		util.RespondValidationError(w, err.Error())
		return
	}

	// Pin作成処理（要件: 6.1）
	pin, err := h.pinService.CreatePin(r.Context(), userID, req.Name, req.Latitude, req.Longitude)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCoordinates) {
			util.RespondValidationError(w, "Invalid coordinates")
			return
		}
		util.RespondInternalError(w, "Failed to create pin")
		return
	}

	// 成功レスポンス（要件: 6.5）
	util.RespondJSON(w, http.StatusCreated, pin)
}

// UpdatePin はPinを更新します
// PUT /api/pins/:id
// 要件: 7.1, 7.5
func (h *PinHandler) UpdatePin(w http.ResponseWriter, r *http.Request) {
	// コンテキストからユーザーIDを取得
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		util.RespondUnauthorized(w, "Unauthorized")
		return
	}

	// URLパラメータからPin IDを取得
	pinID := chi.URLParam(r, "id")
	if pinID == "" {
		util.RespondValidationError(w, "Pin ID is required")
		return
	}

	// リクエストボディのパース
	var req model.UpdatePinRequest
	if err := util.ParseJSONBody(r, &req); err != nil {
		util.RespondValidationError(w, "Invalid request body")
		return
	}

	// バリデーション
	if req.Name == "" {
		util.RespondValidationError(w, "Name is required")
		return
	}
	if err := util.ValidateCoordinates(req.Latitude, req.Longitude); err != nil {
		util.RespondValidationError(w, err.Error())
		return
	}

	// Pin更新処理（要件: 7.1）
	pin, err := h.pinService.UpdatePin(r.Context(), pinID, userID, req.Name, req.Latitude, req.Longitude)
	if err != nil {
		if errors.Is(err, service.ErrPinNotFound) {
			util.RespondNotFound(w, "Pin not found")
			return
		}
		if errors.Is(err, service.ErrUnauthorizedPinAccess) {
			// 要件: 7.4 - 権限エラー
			util.RespondForbidden(w, "You don't have permission to update this pin")
			return
		}
		if errors.Is(err, service.ErrInvalidCoordinates) {
			util.RespondValidationError(w, "Invalid coordinates")
			return
		}
		util.RespondInternalError(w, "Failed to update pin")
		return
	}

	// 成功レスポンス（要件: 7.5）
	util.RespondJSON(w, http.StatusOK, pin)
}

// GetPins はユーザーのPin一覧を取得します
// GET /api/pins
// 要件: 6.1, 7.1
func (h *PinHandler) GetPins(w http.ResponseWriter, r *http.Request) {
	// コンテキストからユーザーIDを取得
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		util.RespondUnauthorized(w, "Unauthorized")
		return
	}

	// ユーザーのPin一覧を取得
	pins, err := h.pinService.GetPinsByUser(r.Context(), userID)
	if err != nil {
		util.RespondInternalError(w, "Failed to get pins")
		return
	}

	// 成功レスポンス
	util.RespondJSON(w, http.StatusOK, pins)
}

// GetPin は指定されたPinを取得します
// GET /api/pins/:id
// 要件: 6.1, 7.1
func (h *PinHandler) GetPin(w http.ResponseWriter, r *http.Request) {
	// URLパラメータからPin IDを取得
	pinID := chi.URLParam(r, "id")
	if pinID == "" {
		util.RespondValidationError(w, "Pin ID is required")
		return
	}

	// Pinを取得
	pin, err := h.pinService.GetPin(r.Context(), pinID)
	if err != nil {
		if errors.Is(err, service.ErrPinNotFound) {
			util.RespondNotFound(w, "Pin not found")
			return
		}
		util.RespondInternalError(w, "Failed to get pin")
		return
	}

	// 成功レスポンス
	util.RespondJSON(w, http.StatusOK, pin)
}

// DeletePin はPinを削除します
// DELETE /api/pins/:id
// 要件: 7.1, 7.5
func (h *PinHandler) DeletePin(w http.ResponseWriter, r *http.Request) {
	// コンテキストからユーザーIDを取得
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		util.RespondUnauthorized(w, "Unauthorized")
		return
	}

	// URLパラメータからPin IDを取得
	pinID := chi.URLParam(r, "id")
	if pinID == "" {
		util.RespondValidationError(w, "Pin ID is required")
		return
	}

	// Pin削除処理
	err := h.pinService.DeletePin(r.Context(), pinID, userID)
	if err != nil {
		if errors.Is(err, service.ErrPinNotFound) {
			util.RespondNotFound(w, "Pin not found")
			return
		}
		if errors.Is(err, service.ErrUnauthorizedPinAccess) {
			// 要件: 7.4 - 権限エラー
			util.RespondForbidden(w, "You don't have permission to delete this pin")
			return
		}
		util.RespondInternalError(w, "Failed to delete pin")
		return
	}

	// 成功レスポンス（要件: 7.5）
	util.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "Pin deleted successfully",
	})
}
