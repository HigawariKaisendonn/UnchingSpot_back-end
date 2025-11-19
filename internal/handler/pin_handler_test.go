package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

// setupPinTestRouter はPin用のテストルーターをセットアップします
func setupPinTestRouter(pinHandler *PinHandler) *chi.Mux {
	r := chi.NewRouter()
	
	r.Route("/api/pins", func(r chi.Router) {
		// 認証が必要なエンドポイント
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware)
			r.Post("/", pinHandler.CreatePin)
			r.Get("/", pinHandler.GetPins)
			r.Get("/{id}", pinHandler.GetPin)
			r.Put("/{id}", pinHandler.UpdatePin)
			r.Delete("/{id}", pinHandler.DeletePin)
		})
	})
	
	return r
}

// TestPinHandler_CreatePin はPin作成エンドポイントのテスト
// 要件: 6.1, 6.5
func TestPinHandler_CreatePin(t *testing.T) {
	// テストデータベースのセットアップ
	testDB, err := database.SetupTestDB()
	require.NoError(t, err)
	defer testDB.Teardown()

	// リポジトリとサービスの初期化
	userRepo := repository.NewUserRepository(testDB.DB)
	pinRepo := repository.NewPinRepository(testDB.DB)
	authService := service.NewAuthService(userRepo)
	pinService := service.NewPinService(pinRepo)
	pinHandler := NewPinHandler(pinService)
	router := setupPinTestRouter(pinHandler)

	// テストヘルパーの作成
	helper := database.NewTestHelper(testDB)

	t.Run("成功: 有効なリクエストでPin作成", func(t *testing.T) {
		defer testDB.CleanupData()

		// テストユーザーの作成
		user, err := helper.CreateTestUser("test@example.com", "password123", "Test User")
		require.NoError(t, err)

		// トークンの生成
		token, _, err := authService.Login(context.Background(), user.Email, "password123")
		require.NoError(t, err)

		reqBody := model.CreatePinRequest{
			Name:      "トイレA",
			Latitude:  35.6895,
			Longitude: 139.6917,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/pins/", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// 要件: 6.5 - 作成されたPin情報を含むレスポンスを返す
		assert.Equal(t, http.StatusCreated, w.Code)

		var pin model.Pin
		err = json.Unmarshal(w.Body.Bytes(), &pin)
		require.NoError(t, err)
		assert.NotEmpty(t, pin.ID)
		assert.Equal(t, "トイレA", pin.Name)
		assert.Equal(t, user.ID, pin.UserID)
		assert.Equal(t, 35.6895, pin.Latitude)
		assert.Equal(t, 139.6917, pin.Longitude)
	})

	t.Run("エラー: 未認証のリクエスト", func(t *testing.T) {
		defer testDB.CleanupData()

		reqBody := model.CreatePinRequest{
			Name:      "トイレA",
			Latitude:  35.6895,
			Longitude: 139.6917,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/pins/", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("エラー: 無効な緯度", func(t *testing.T) {
		defer testDB.CleanupData()

		// テストユーザーの作成
		user, err := helper.CreateTestUser("test@example.com", "password123", "Test User")
		require.NoError(t, err)

		// トークンの生成
		token, _, err := authService.Login(context.Background(), user.Email, "password123")
		require.NoError(t, err)

		reqBody := model.CreatePinRequest{
			Name:      "トイレA",
			Latitude:  91.0, // 無効な緯度
			Longitude: 139.6917,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/pins/", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("エラー: 名前が空", func(t *testing.T) {
		defer testDB.CleanupData()

		// テストユーザーの作成
		user, err := helper.CreateTestUser("test@example.com", "password123", "Test User")
		require.NoError(t, err)

		// トークンの生成
		token, _, err := authService.Login(context.Background(), user.Email, "password123")
		require.NoError(t, err)

		reqBody := model.CreatePinRequest{
			Name:      "",
			Latitude:  35.6895,
			Longitude: 139.6917,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/pins/", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestPinHandler_UpdatePin はPin更新エンドポイントのテスト
// 要件: 7.1, 7.4, 7.5
func TestPinHandler_UpdatePin(t *testing.T) {
	// テストデータベースのセットアップ
	testDB, err := database.SetupTestDB()
	require.NoError(t, err)
	defer testDB.Teardown()

	// リポジトリとサービスの初期化
	userRepo := repository.NewUserRepository(testDB.DB)
	pinRepo := repository.NewPinRepository(testDB.DB)
	authService := service.NewAuthService(userRepo)
	pinService := service.NewPinService(pinRepo)
	pinHandler := NewPinHandler(pinService)
	router := setupPinTestRouter(pinHandler)

	// テストヘルパーの作成
	helper := database.NewTestHelper(testDB)

	t.Run("成功: 自分が所有するPinの更新", func(t *testing.T) {
		defer testDB.CleanupData()

		// テストユーザーの作成
		user, err := helper.CreateTestUser("test@example.com", "password123", "Test User")
		require.NoError(t, err)

		// テストPinの作成
		pin, err := helper.CreateTestPin(user.ID, "トイレA", 35.6895, 139.6917)
		require.NoError(t, err)

		// トークンの生成
		token, _, err := authService.Login(context.Background(), user.Email, "password123")
		require.NoError(t, err)

		reqBody := model.UpdatePinRequest{
			Name:      "トイレB",
			Latitude:  35.7000,
			Longitude: 139.7000,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/pins/"+pin.ID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// 要件: 7.5 - 更新されたPin情報を含むレスポンスを返す
		assert.Equal(t, http.StatusOK, w.Code)

		var updatedPin model.Pin
		err = json.Unmarshal(w.Body.Bytes(), &updatedPin)
		require.NoError(t, err)
		assert.Equal(t, pin.ID, updatedPin.ID)
		assert.Equal(t, "トイレB", updatedPin.Name)
		assert.Equal(t, 35.7000, updatedPin.Latitude)
		assert.Equal(t, 139.7000, updatedPin.Longitude)
	})

	t.Run("エラー: 他のユーザーのPinを更新しようとする", func(t *testing.T) {
		defer testDB.CleanupData()

		// テストユーザー1の作成
		user1, err := helper.CreateTestUser("user1@example.com", "password123", "User 1")
		require.NoError(t, err)

		// テストユーザー2の作成
		user2, err := helper.CreateTestUser("user2@example.com", "password123", "User 2")
		require.NoError(t, err)

		// ユーザー1のPinを作成
		pin, err := helper.CreateTestPin(user1.ID, "トイレA", 35.6895, 139.6917)
		require.NoError(t, err)

		// ユーザー2のトークンを生成
		token, _, err := authService.Login(context.Background(), user2.Email, "password123")
		require.NoError(t, err)

		reqBody := model.UpdatePinRequest{
			Name:      "トイレB",
			Latitude:  35.7000,
			Longitude: 139.7000,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/pins/"+pin.ID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// 要件: 7.4 - 権限エラーレスポンスを返す
		assert.Equal(t, http.StatusForbidden, w.Code)

		var errResp model.ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, model.ErrCodeForbidden, errResp.Error.Code)
	})

	t.Run("エラー: 存在しないPinの更新", func(t *testing.T) {
		defer testDB.CleanupData()

		// テストユーザーの作成
		user, err := helper.CreateTestUser("test@example.com", "password123", "Test User")
		require.NoError(t, err)

		// トークンの生成
		token, _, err := authService.Login(context.Background(), user.Email, "password123")
		require.NoError(t, err)

		reqBody := model.UpdatePinRequest{
			Name:      "トイレB",
			Latitude:  35.7000,
			Longitude: 139.7000,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/pins/nonexistent-id", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// TestPinHandler_GetPins はPin一覧取得エンドポイントのテスト
// 要件: 6.1, 7.1
func TestPinHandler_GetPins(t *testing.T) {
	// テストデータベースのセットアップ
	testDB, err := database.SetupTestDB()
	require.NoError(t, err)
	defer testDB.Teardown()

	// リポジトリとサービスの初期化
	userRepo := repository.NewUserRepository(testDB.DB)
	pinRepo := repository.NewPinRepository(testDB.DB)
	authService := service.NewAuthService(userRepo)
	pinService := service.NewPinService(pinRepo)
	pinHandler := NewPinHandler(pinService)
	router := setupPinTestRouter(pinHandler)

	// テストヘルパーの作成
	helper := database.NewTestHelper(testDB)

	t.Run("成功: ユーザーのPin一覧を取得", func(t *testing.T) {
		defer testDB.CleanupData()

		// テストユーザーの作成
		user, err := helper.CreateTestUser("test@example.com", "password123", "Test User")
		require.NoError(t, err)

		// テストPinの作成
		pin1, err := helper.CreateTestPin(user.ID, "トイレA", 35.6895, 139.6917)
		require.NoError(t, err)
		pin2, err := helper.CreateTestPin(user.ID, "トイレB", 35.7000, 139.7000)
		require.NoError(t, err)

		// トークンの生成
		token, _, err := authService.Login(context.Background(), user.Email, "password123")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/pins/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var pins []model.Pin
		err = json.Unmarshal(w.Body.Bytes(), &pins)
		require.NoError(t, err)
		assert.Len(t, pins, 2)
		
		// Pin IDの確認
		pinIDs := []string{pins[0].ID, pins[1].ID}
		assert.Contains(t, pinIDs, pin1.ID)
		assert.Contains(t, pinIDs, pin2.ID)
	})

	t.Run("成功: Pinが0件の場合", func(t *testing.T) {
		defer testDB.CleanupData()

		// テストユーザーの作成
		user, err := helper.CreateTestUser("test@example.com", "password123", "Test User")
		require.NoError(t, err)

		// トークンの生成
		token, _, err := authService.Login(context.Background(), user.Email, "password123")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/pins/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var pins []model.Pin
		err = json.Unmarshal(w.Body.Bytes(), &pins)
		require.NoError(t, err)
		assert.Len(t, pins, 0)
	})
}

// TestPinHandler_GetPin は個別Pin取得エンドポイントのテスト
// 要件: 6.1, 7.1
func TestPinHandler_GetPin(t *testing.T) {
	// テストデータベースのセットアップ
	testDB, err := database.SetupTestDB()
	require.NoError(t, err)
	defer testDB.Teardown()

	// リポジトリとサービスの初期化
	userRepo := repository.NewUserRepository(testDB.DB)
	pinRepo := repository.NewPinRepository(testDB.DB)
	authService := service.NewAuthService(userRepo)
	pinService := service.NewPinService(pinRepo)
	pinHandler := NewPinHandler(pinService)
	router := setupPinTestRouter(pinHandler)

	// テストヘルパーの作成
	helper := database.NewTestHelper(testDB)

	t.Run("成功: 指定されたPinを取得", func(t *testing.T) {
		defer testDB.CleanupData()

		// テストユーザーの作成
		user, err := helper.CreateTestUser("test@example.com", "password123", "Test User")
		require.NoError(t, err)

		// テストPinの作成
		pin, err := helper.CreateTestPin(user.ID, "トイレA", 35.6895, 139.6917)
		require.NoError(t, err)

		// トークンの生成
		token, _, err := authService.Login(context.Background(), user.Email, "password123")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/pins/"+pin.ID, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var retrievedPin model.Pin
		err = json.Unmarshal(w.Body.Bytes(), &retrievedPin)
		require.NoError(t, err)
		assert.Equal(t, pin.ID, retrievedPin.ID)
		assert.Equal(t, "トイレA", retrievedPin.Name)
		assert.Equal(t, 35.6895, retrievedPin.Latitude)
		assert.Equal(t, 139.6917, retrievedPin.Longitude)
	})

	t.Run("エラー: 存在しないPinの取得", func(t *testing.T) {
		defer testDB.CleanupData()

		// テストユーザーの作成
		user, err := helper.CreateTestUser("test@example.com", "password123", "Test User")
		require.NoError(t, err)

		// トークンの生成
		token, _, err := authService.Login(context.Background(), user.Email, "password123")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/pins/nonexistent-id", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// TestPinHandler_DeletePin はPin削除エンドポイントのテスト
// 要件: 7.1, 7.4, 7.5
func TestPinHandler_DeletePin(t *testing.T) {
	// テストデータベースのセットアップ
	testDB, err := database.SetupTestDB()
	require.NoError(t, err)
	defer testDB.Teardown()

	// リポジトリとサービスの初期化
	userRepo := repository.NewUserRepository(testDB.DB)
	pinRepo := repository.NewPinRepository(testDB.DB)
	authService := service.NewAuthService(userRepo)
	pinService := service.NewPinService(pinRepo)
	pinHandler := NewPinHandler(pinService)
	router := setupPinTestRouter(pinHandler)

	// テストヘルパーの作成
	helper := database.NewTestHelper(testDB)

	t.Run("成功: 自分が所有するPinの削除", func(t *testing.T) {
		defer testDB.CleanupData()

		// テストユーザーの作成
		user, err := helper.CreateTestUser("test@example.com", "password123", "Test User")
		require.NoError(t, err)

		// テストPinの作成
		pin, err := helper.CreateTestPin(user.ID, "トイレA", 35.6895, 139.6917)
		require.NoError(t, err)

		// トークンの生成
		token, _, err := authService.Login(context.Background(), user.Email, "password123")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodDelete, "/api/pins/"+pin.ID, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// 要件: 7.5 - 成功レスポンスを返す
		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]string
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Contains(t, resp["message"], "deleted successfully")

		// Pinが削除されたことを確認
		_, err = pinService.GetPin(context.Background(), pin.ID)
		assert.Error(t, err)
	})

	t.Run("エラー: 他のユーザーのPinを削除しようとする", func(t *testing.T) {
		defer testDB.CleanupData()

		// テストユーザー1の作成
		user1, err := helper.CreateTestUser("user1@example.com", "password123", "User 1")
		require.NoError(t, err)

		// テストユーザー2の作成
		user2, err := helper.CreateTestUser("user2@example.com", "password123", "User 2")
		require.NoError(t, err)

		// ユーザー1のPinを作成
		pin, err := helper.CreateTestPin(user1.ID, "トイレA", 35.6895, 139.6917)
		require.NoError(t, err)

		// ユーザー2のトークンを生成
		token, _, err := authService.Login(context.Background(), user2.Email, "password123")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodDelete, "/api/pins/"+pin.ID, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// 要件: 7.4 - 権限エラーレスポンスを返す
		assert.Equal(t, http.StatusForbidden, w.Code)

		var errResp model.ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, model.ErrCodeForbidden, errResp.Error.Code)
	})

	t.Run("エラー: 存在しないPinの削除", func(t *testing.T) {
		defer testDB.CleanupData()

		// テストユーザーの作成
		user, err := helper.CreateTestUser("test@example.com", "password123", "Test User")
		require.NoError(t, err)

		// トークンの生成
		token, _, err := authService.Login(context.Background(), user.Email, "password123")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodDelete, "/api/pins/nonexistent-id", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
