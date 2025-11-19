package handler

import (
	"errors"
	"net/http"

	"github.com/higawarikaisendonn/unchingspot-backend/internal/model"
	"github.com/higawarikaisendonn/unchingspot-backend/internal/service"
	"github.com/higawarikaisendonn/unchingspot-backend/internal/util"
)

// AuthHandler は認証関連のHTTPハンドラーを提供します
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler は新しいAuthHandlerインスタンスを作成します
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// SignUp はユーザー登録を処理します
// POST /api/auth/signup
// 要件: 1.1, 1.4, 1.5
func (h *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	// リクエストボディのパース
	var req model.SignUpRequest
	if err := util.ParseJSONBody(r, &req); err != nil {
		util.RespondValidationError(w, "Invalid request body")
		return
	}

	// バリデーション
	if err := util.ValidateEmail(req.Email); err != nil {
		util.RespondValidationError(w, err.Error())
		return
	}
	if err := util.ValidatePassword(req.Password); err != nil {
		util.RespondValidationError(w, err.Error())
		return
	}
	if err := util.ValidateRequired(req.Name, "name"); err != nil {
		util.RespondValidationError(w, err.Error())
		return
	}

	// ユーザー登録処理
	user, err := h.authService.SignUp(r.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		if errors.Is(err, service.ErrEmailAlreadyExists) {
			util.RespondConflict(w, "Email already registered")
			return
		}
		util.RespondInternalError(w, "Failed to create user")
		return
	}

	// 成功レスポンス（要件: 1.5）
	util.RespondJSON(w, http.StatusCreated, user)
}

// Login はユーザーログインを処理します
// POST /api/auth/login
// 要件: 2.1, 2.3, 2.4
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// リクエストボディのパース
	var req model.LoginRequest
	if err := util.ParseJSONBody(r, &req); err != nil {
		util.RespondValidationError(w, "Invalid request body")
		return
	}

	// バリデーション
	if err := util.ValidateEmail(req.Email); err != nil {
		util.RespondValidationError(w, err.Error())
		return
	}
	if err := util.ValidateRequired(req.Password, "password"); err != nil {
		util.RespondValidationError(w, err.Error())
		return
	}

	// ログイン処理
	token, user, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			// 要件: 2.4 - 無効な認証情報の場合
			util.RespondUnauthorized(w, "Invalid email or password")
			return
		}
		util.RespondInternalError(w, "Failed to login")
		return
	}

	// 成功レスポンス（要件: 2.3）
	response := model.AuthResponse{
		Token: token,
		User:  user,
	}
	util.RespondJSON(w, http.StatusOK, response)
}

// Logout はユーザーログアウトを処理します
// POST /api/auth/logout
// 要件: 3.1, 3.2
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// JWTベースの認証では、クライアント側でトークンを削除するだけで十分
	// サーバー側では特に処理は不要（ステートレス）
	// 将来的にトークンのブラックリスト機能を追加する場合はここで実装
	
	// 要件: 3.2 - 成功ステータスを返す
	util.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
}

// GetMe は現在のユーザー情報を取得します
// GET /api/auth/me
// 要件: 4.1, 4.2, 4.3
func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	// コンテキストからユーザーIDを取得（認証ミドルウェアで設定される）
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		// 要件: 4.3 - 未認証の場合
		util.RespondUnauthorized(w, "Unauthorized")
		return
	}

	// Authorizationヘッダーからトークンを取得
	token := r.Header.Get("Authorization")
	if token == "" {
		util.RespondUnauthorized(w, "Unauthorized")
		return
	}

	// "Bearer "プレフィックスを削除
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// ユーザー情報の取得（要件: 4.1）
	user, err := h.authService.ValidateToken(r.Context(), token)
	if err != nil {
		util.RespondUnauthorized(w, "Invalid token")
		return
	}

	// 要件: 4.2 - ユーザー名を含むレスポンスを返す
	util.RespondJSON(w, http.StatusOK, user)
}

// TestConnection はデータベース接続をテストします
// GET /api/auth/test
// 要件: 5.1, 5.2, 5.3
func (h *AuthHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	// データベース接続テスト（要件: 5.1）
	err := h.authService.TestConnection(r.Context())
	if err != nil {
		// 要件: 5.3 - 接続失敗の場合
		util.RespondDatabaseError(w, "Database connection failed: "+err.Error())
		return
	}

	// 要件: 5.2 - 接続成功の場合
	util.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "Database connection successful",
		"status":  "ok",
	})
}
