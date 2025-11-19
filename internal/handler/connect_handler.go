package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/higawarikaisendonn/unchingspot-backend/internal/middleware"
	"github.com/higawarikaisendonn/unchingspot-backend/internal/model"
	"github.com/higawarikaisendonn/unchingspot-backend/internal/service"
	"github.com/higawarikaisendonn/unchingspot-backend/internal/util"
)

// ConnectHandler はConnect関連のHTTPハンドラーを提供します
type ConnectHandler struct {
	connectService service.ConnectService
}

// NewConnectHandler は新しいConnectHandlerインスタンスを作成します
func NewConnectHandler(connectService service.ConnectService) *ConnectHandler {
	return &ConnectHandler{
		connectService: connectService,
	}
}

// CreateConnect はConnectを作成します
// POST /api/connects
// 要件: 8.1, 8.7
func (h *ConnectHandler) CreateConnect(w http.ResponseWriter, r *http.Request) {
	// コンテキストからユーザーIDを取得（認証ミドルウェアで設定される）
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		util.RespondUnauthorized(w, "Unauthorized")
		return
	}

	// リクエストボディのパース
	var req model.CreateConnectRequest
	if err := util.ParseJSONBody(r, &req); err != nil {
		util.RespondValidationError(w, "Invalid request body")
		return
	}

	// バリデーション
	if err := util.ValidateRequired(req.Name, "name"); err != nil {
		util.RespondValidationError(w, err.Error())
		return
	}
	if err := util.ValidateRequired(req.PinID1, "pin_id_1"); err != nil {
		util.RespondValidationError(w, err.Error())
		return
	}
	if len(req.PinID2) == 0 {
		util.RespondValidationError(w, "pin_id_2 is required and must contain at least one pin")
		return
	}

	// Connect作成処理（要件: 8.1）
	connect, err := h.connectService.CreateConnect(r.Context(), userID, req.Name, req.PinID1, req.PinID2, req.Show)
	if err != nil {
		if errors.Is(err, service.ErrPinNotExist) {
			util.RespondNotFound(w, "One or both pins do not exist")
			return
		}
		if errors.Is(err, service.ErrInvalidPinIDs) {
			util.RespondValidationError(w, "Invalid pin IDs")
			return
		}
		util.RespondInternalError(w, "Failed to create connect")
		return
	}

	// 成功レスポンス（要件: 8.7）
	util.RespondJSON(w, http.StatusCreated, connect)
}

// UpdateConnect はConnectを更新します
// PUT /api/connects/:id
// 要件: 9.1, 9.6
func (h *ConnectHandler) UpdateConnect(w http.ResponseWriter, r *http.Request) {
	// コンテキストからユーザーIDを取得
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		util.RespondUnauthorized(w, "Unauthorized")
		return
	}

	// URLパラメータからConnect IDを取得
	connectID := chi.URLParam(r, "id")
	if connectID == "" {
		util.RespondValidationError(w, "Connect ID is required")
		return
	}

	// リクエストボディのパース
	var req model.UpdateConnectRequest
	if err := util.ParseJSONBody(r, &req); err != nil {
		util.RespondValidationError(w, "Invalid request body")
		return
	}

	// showフィールドのデフォルト値設定
	show := false
	if req.Show != nil {
		show = *req.Show
	}

	// Connect更新処理（要件: 9.1）
	connect, err := h.connectService.UpdateConnect(r.Context(), connectID, userID, req.PinID1, req.PinID2, show)
	if err != nil {
		if errors.Is(err, service.ErrConnectNotFound) {
			util.RespondNotFound(w, "Connect not found")
			return
		}
		if errors.Is(err, service.ErrUnauthorizedConnectAccess) {
			util.RespondForbidden(w, "You don't have permission to update this connect")
			return
		}
		if errors.Is(err, service.ErrPinNotExist) {
			util.RespondNotFound(w, "One or both pins do not exist")
			return
		}
		util.RespondInternalError(w, "Failed to update connect")
		return
	}

	// 成功レスポンス（要件: 9.6）
	util.RespondJSON(w, http.StatusOK, connect)
}

// GetConnects はユーザーのConnect一覧を取得します
// GET /api/connects
// 要件: 8.1, 9.1
func (h *ConnectHandler) GetConnects(w http.ResponseWriter, r *http.Request) {
	// コンテキストからユーザーIDを取得
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		util.RespondUnauthorized(w, "Unauthorized")
		return
	}

	// ユーザーのConnect一覧を取得
	connects, err := h.connectService.GetConnectsByUser(r.Context(), userID)
	if err != nil {
		util.RespondInternalError(w, "Failed to get connects")
		return
	}

	// 成功レスポンス
	util.RespondJSON(w, http.StatusOK, connects)
}

// DeleteConnect はConnectを削除します
// DELETE /api/connects/:id
// 要件: 9.1, 9.6
func (h *ConnectHandler) DeleteConnect(w http.ResponseWriter, r *http.Request) {
	// コンテキストからユーザーIDを取得
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		util.RespondUnauthorized(w, "Unauthorized")
		return
	}

	// URLパラメータからConnect IDを取得
	connectID := chi.URLParam(r, "id")
	if connectID == "" {
		util.RespondValidationError(w, "Connect ID is required")
		return
	}

	// Connect削除処理
	err := h.connectService.DeleteConnect(r.Context(), connectID, userID)
	if err != nil {
		if errors.Is(err, service.ErrConnectNotFound) {
			util.RespondNotFound(w, "Connect not found")
			return
		}
		if errors.Is(err, service.ErrUnauthorizedConnectAccess) {
			util.RespondForbidden(w, "You don't have permission to delete this connect")
			return
		}
		util.RespondInternalError(w, "Failed to delete connect")
		return
	}

	// 成功レスポンス（要件: 9.6）
	util.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "Connect deleted successfully",
	})
}
