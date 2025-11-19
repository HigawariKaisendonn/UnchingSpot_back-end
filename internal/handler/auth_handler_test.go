package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/higawarikaisendonn/unchingspot-backend/internal/database"
	"github.com/higawarikaisendonn/unchingspot-backend/internal/middleware"
	"github.com/higawarikaisendonn/unchingspot-backend/internal/model"
	"github.com/higawarikaisendonn/unchingspot-backend/internal/repository"
	"github.com/higawarikaisendonn/unchingspot-backend/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain はテストの前に環境変数を設定します
func TestMain(m *testing.M) {
	// JWT_SECRETの設定（テスト用）
	os.Setenv("JWT_SECRET", "test-secret-key-for-jwt-token-generation-32chars")
	
	// テストの実行
	code := m.Run()
	
	// 終了
	os.Exit(code)
}

// setupTestRouter はテスト用のルーターをセットアップします
func setupTestRouter(authHandler *AuthHandler) *chi.Mux {
	r := chi.NewRouter()
	
	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/signup", authHandler.SignUp)
		r.Post("/login", authHandler.Login)
		r.Get("/test", authHandler.TestConnection)
		
		// 認証が必要なエンドポイント
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware)
			r.Post("/logout", authHandler.Logout)
			r.Get("/me", authHandler.GetMe)
		})
	})
	
	return r
}

// TestAuthHandler_SignUp はユーザー登録エンドポイントのテスト
// 要件: 1.1, 1.4, 1.5
func TestAuthHandler_SignUp(t *testing.T) {
	// テストデータベースのセットアップ
	testDB, err := database.SetupTestDB()
	require.NoError(t, err)
	defer testDB.Teardown()

	// リポジトリとサービスの初期化
	userRepo := repository.NewUserRepository(testDB.DB)
	authService := service.NewAuthService(userRepo)
	authHandler := NewAuthHandler(authService)
	router := setupTestRouter(authHandler)

	t.Run("成功: 有効なリクエストでユーザー登録", func(t *testing.T) {
		defer testDB.CleanupData()

		reqBody := model.SignUpRequest{
			Email:    "test@example.com",
			Password: "password123",
			Name:     "Test User",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/signup", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// 要件: 1.5 - 成功ステータスとユーザー情報を返す
		assert.Equal(t, http.StatusCreated, w.Code)

		var user model.User
		err := json.Unmarshal(w.Body.Bytes(), &user)
		require.NoError(t, err)
		assert.NotEmpty(t, user.ID)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, "Test User", user.Name)
		assert.Empty(t, user.Password) // パスワードはレスポンスに含まれない
	})

	t.Run("エラー: 既に登録済みのメールアドレス", func(t *testing.T) {
		defer testDB.CleanupData()

		// 最初のユーザー登録
		reqBody := model.SignUpRequest{
			Email:    "duplicate@example.com",
			Password: "password123",
			Name:     "First User",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/auth/signup", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		// 要件: 1.4 - 重複登録を防止
		req = httptest.NewRequest(http.MethodPost, "/api/auth/signup", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)

		var errResp model.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, model.ErrCodeConflict, errResp.Error.Code)
	})

	t.Run("エラー: 無効なメールアドレス", func(t *testing.T) {
		reqBody := model.SignUpRequest{
			Email:    "invalid-email",
			Password: "password123",
			Name:     "Test User",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/signup", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("エラー: パスワードが短すぎる", func(t *testing.T) {
		reqBody := model.SignUpRequest{
			Email:    "test@example.com",
			Password: "short",
			Name:     "Test User",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/signup", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("エラー: 名前が空", func(t *testing.T) {
		reqBody := model.SignUpRequest{
			Email:    "test@example.com",
			Password: "password123",
			Name:     "",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/signup", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestAuthHandler_Login はログインエンドポイントのテスト
// 要件: 2.1, 2.4
func TestAuthHandler_Login(t *testing.T) {
	// テストデータベースのセットアップ
	testDB, err := database.SetupTestDB()
	require.NoError(t, err)
	defer testDB.Teardown()

	// リポジトリとサービスの初期化
	userRepo := repository.NewUserRepository(testDB.DB)
	authService := service.NewAuthService(userRepo)
	authHandler := NewAuthHandler(authService)
	router := setupTestRouter(authHandler)

	// テストヘルパーの作成
	helper := database.NewTestHelper(testDB)

	t.Run("成功: 有効な認証情報でログイン", func(t *testing.T) {
		defer testDB.CleanupData()

		// テストユーザーの作成
		_, err := helper.CreateTestUser("terakai@gmail.com", "12345678", "tera")
		require.NoError(t, err)

		reqBody := model.LoginRequest{
			Email:    "terakai@gmail.com",
			Password: "12345678",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// デバッグ: レスポンスの内容を確認
		if w.Code != http.StatusOK {
			t.Logf("Response body: %s", w.Body.String())
		}

		// 要件: 2.3 - 認証トークンとユーザー情報を返す
		assert.Equal(t, http.StatusOK, w.Code)

		var authResp model.AuthResponse
		err = json.Unmarshal(w.Body.Bytes(), &authResp)
		require.NoError(t, err)
		assert.NotEmpty(t, authResp.Token)
		assert.NotNil(t, authResp.User)
		assert.Equal(t, "terakai@gmail.com", authResp.User.Email)
		assert.Equal(t, "tera", authResp.User.Name)
	})

	t.Run("エラー: 無効なメールアドレス", func(t *testing.T) {
		defer testDB.CleanupData()

		reqBody := model.LoginRequest{
			Email:    "nonexistent@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// 要件: 2.4 - 無効な認証情報の場合
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var errResp model.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, model.ErrCodeUnauthorized, errResp.Error.Code)
	})

	t.Run("エラー: 無効なパスワード", func(t *testing.T) {
		defer testDB.CleanupData()

		// テストユーザーの作成
		_, err := helper.CreateTestUser("terakai@gmail.com", "12345678", "tera")
		require.NoError(t, err)

		reqBody := model.LoginRequest{
			Email:    "terakai@gmail.com",
			Password: "wrongpassword",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// 要件: 2.4 - 無効な認証情報の場合
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// TestAuthHandler_Logout はログアウトエンドポイントのテスト
// 要件: 3.1, 3.2
func TestAuthHandler_Logout(t *testing.T) {
	// テストデータベースのセットアップ
	testDB, err := database.SetupTestDB()
	require.NoError(t, err)
	defer testDB.Teardown()
	defer testDB.CleanupData()

	// リポジトリとサービスの初期化
	userRepo := repository.NewUserRepository(testDB.DB)
	authService := service.NewAuthService(userRepo)
	authHandler := NewAuthHandler(authService)
	router := setupTestRouter(authHandler)

	// テストヘルパーの作成
	helper := database.NewTestHelper(testDB)

	t.Run("成功: 認証済みユーザーのログアウト", func(t *testing.T) {
		// テストユーザーの作成
		user, err := helper.CreateTestUser("terakai@gmail.com", "12345678", "tera")
		require.NoError(t, err)

		// トークンの生成
		token, _, err := authService.Login(context.Background(), user.Email, "12345678")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// 要件: 3.2 - 成功ステータスを返す
		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]string
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Contains(t, resp["message"], "Logged out successfully")
	})

	t.Run("エラー: 未認証のログアウト試行", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// 要件: 4.3 - 未認証の場合
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// TestAuthHandler_GetMe はユーザー情報取得エンドポイントのテスト
// 要件: 4.1, 4.2, 4.3
func TestAuthHandler_GetMe(t *testing.T) {
	// テストデータベースのセットアップ
	testDB, err := database.SetupTestDB()
	require.NoError(t, err)
	defer testDB.Teardown()
	defer testDB.CleanupData()

	// リポジトリとサービスの初期化
	userRepo := repository.NewUserRepository(testDB.DB)
	authService := service.NewAuthService(userRepo)
	authHandler := NewAuthHandler(authService)
	router := setupTestRouter(authHandler)

	// テストヘルパーの作成
	helper := database.NewTestHelper(testDB)

	t.Run("成功: 認証済みユーザーの情報取得", func(t *testing.T) {
		// テストユーザーの作成
		user, err := helper.CreateTestUser("terakai@gmail.com", "12345678", "tera")
		require.NoError(t, err)

		// トークンの生成
		token, _, err := authService.Login(context.Background(), user.Email, "12345678")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// 要件: 4.2 - ユーザー名を含むレスポンスを返す
		assert.Equal(t, http.StatusOK, w.Code)

		var respUser model.User
		err = json.Unmarshal(w.Body.Bytes(), &respUser)
		require.NoError(t, err)
		assert.Equal(t, user.ID, respUser.ID)
		assert.Equal(t, "terakai@gmail.com", respUser.Email)
		assert.Equal(t, "tera", respUser.Name)
	})

	t.Run("エラー: 未認証のリクエスト", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// 要件: 4.3 - 未認証の場合
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("エラー: 無効なトークン", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// TestAuthHandler_TestConnection はデータベース接続テストエンドポイントのテスト
// 要件: 5.1, 5.2, 5.3
func TestAuthHandler_TestConnection(t *testing.T) {
	// テストデータベースのセットアップ
	testDB, err := database.SetupTestDB()
	require.NoError(t, err)
	defer testDB.Teardown()

	// リポジトリとサービスの初期化
	userRepo := repository.NewUserRepository(testDB.DB)
	authService := service.NewAuthService(userRepo)
	authHandler := NewAuthHandler(authService)
	router := setupTestRouter(authHandler)

	t.Run("成功: データベース接続テスト", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/auth/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// 要件: 5.2 - 接続成功の場合
		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "ok", resp["status"])
		assert.Contains(t, resp["message"], "Database connection successful")
	})
}
