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

// setupConnectTestRouter はConnect用のテストルーターをセットアップします
func setupConnectTestRouter(connectHandler *ConnectHandler) *chi.Mux {
	r := chi.NewRouter()
	
	r.Route("/api/connects", func(r chi.Router) {
		// 認証が必要なエンドポイント
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware)
			r.Post("/", connectHandler.CreateConnect)
			r.Get("/", connectHandler.GetConnects)
			r.Put("/{id}", connectHandler.UpdateConnect)
			r.Delete("/{id}", connectHandler.DeleteConnect)
		})
	})
	
	return r
}

// TestConnectHandler_CreateConnect はConnect作成エンドポイントのテスト
// 要件: 8.1, 8.7
func TestConnectHandler_CreateConnect(t *testing.T) {
	// テストデータベースのセットアップ
	testDB, err := database.SetupTestDB()
	require.NoError(t, err)
	defer testDB.Teardown()

	// リポジトリとサービスの初期化
	userRepo := repository.NewUserRepository(testDB.DB)
	pinRepo := repository.NewPinRepository(testDB.DB)
	connectRepo := repository.NewConnectRepository(testDB.DB)
	authService := service.NewAuthService(userRepo)
	// pinService := service.NewPinService(pinRepo)
	connectService := service.NewConnectService(connectRepo, pinRepo)
	connectHandler := NewConnectHandler(connectService)
	router := setupConnectTestRouter(connectHandler)

	// テストヘルパーの作成
	helper := database.NewTestHelper(testDB)

	t.Run("成功: 有効なリクエストでConnect作成", func(t *testing.T) {
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

		reqBody := model.CreateConnectRequest{
			PinID1: pin1.ID,
			PinID2: []string{pin2.ID},
			Show:   true,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/connects/", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// 要件: 8.7 - 作成されたConnect情報を含むレスポンスを返す
		assert.Equal(t, http.StatusCreated, w.Code)

		var connect model.Connect
		err = json.Unmarshal(w.Body.Bytes(), &connect)
		require.NoError(t, err)
		assert.NotEmpty(t, connect.ID)
		assert.Equal(t, user.ID, connect.UserID)
		assert.Equal(t, pin1.ID, connect.PinID1)
		assert.Len(t, connect.PinID2, 1)
		assert.Equal(t, pin2.ID, connect.PinID2[0])
		assert.True(t, connect.Show)
	})

	t.Run("成功: showフィールドがfalseのConnect作成", func(t *testing.T) {
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

		reqBody := model.CreateConnectRequest{
			PinID1: pin1.ID,
			PinID2: []string{pin2.ID},
			Show:   false,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/connects/", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var connect model.Connect
		err = json.Unmarshal(w.Body.Bytes(), &connect)
		require.NoError(t, err)
		assert.False(t, connect.Show)
	})

	t.Run("エラー: 未認証のリクエスト", func(t *testing.T) {
		defer testDB.CleanupData()

		reqBody := model.CreateConnectRequest{
			PinID1: "pin1-id",
			PinID2: []string{"pin2-id"},
			Show:   true,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/connects/", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("エラー: 存在しないPinを指定", func(t *testing.T) {
		defer testDB.CleanupData()

		// テストユーザーの作成
		user, err := helper.CreateTestUser("test@example.com", "password123", "Test User")
		require.NoError(t, err)

		// トークンの生成
		token, _, err := authService.Login(context.Background(), user.Email, "password123")
		require.NoError(t, err)

		reqBody := model.CreateConnectRequest{
			PinID1: "nonexistent-pin-1",
			PinID2: []string{"nonexistent-pin-2"},
			Show:   true,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/connects/", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// 要件: 8.6 - Pin存在確認
		assert.Equal(t, http.StatusNotFound, w.Code)

		var errResp model.ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Contains(t, errResp.Error.Message, "do not exist")
	})

	t.Run("エラー: 片方のPinのみ存在しない", func(t *testing.T) {
		defer testDB.CleanupData()

		// テストユーザーの作成
		user, err := helper.CreateTestUser("test@example.com", "password123", "Test User")
		require.NoError(t, err)

		// テストPinの作成（1つのみ）
		pin1, err := helper.CreateTestPin(user.ID, "トイレA", 35.6895, 139.6917)
		require.NoError(t, err)

		// トークンの生成
		token, _, err := authService.Login(context.Background(), user.Email, "password123")
		require.NoError(t, err)

		reqBody := model.CreateConnectRequest{
			PinID1: pin1.ID,
			PinID2: []string{"nonexistent-pin"},
			Show:   true,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/connects/", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// 要件: 8.6 - Pin存在確認
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("エラー: PinID1が空", func(t *testing.T) {
		defer testDB.CleanupData()

		// テストユーザーの作成
		user, err := helper.CreateTestUser("test@example.com", "password123", "Test User")
		require.NoError(t, err)

		// トークンの生成
		token, _, err := authService.Login(context.Background(), user.Email, "password123")
		require.NoError(t, err)

		reqBody := model.CreateConnectRequest{
			PinID1: "",
			PinID2: []string{"pin2-id"},
			Show:   true,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/connects/", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("エラー: PinID2が空", func(t *testing.T) {
		defer testDB.CleanupData()

		// テストユーザーの作成
		user, err := helper.CreateTestUser("test@example.com", "password123", "Test User")
		require.NoError(t, err)

		// トークンの生成
		token, _, err := authService.Login(context.Background(), user.Email, "password123")
		require.NoError(t, err)

		reqBody := model.CreateConnectRequest{
			PinID1: "pin1-id",
			PinID2: []string{},
			Show:   true,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/connects/", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestConnectHandler_UpdateConnect はConnect更新エンドポイントのテスト
// 要件: 9.1, 9.5, 9.6
func TestConnectHandler_UpdateConnect(t *testing.T) {
	// テストデータベースのセットアップ
	testDB, err := database.SetupTestDB()
	require.NoError(t, err)
	defer testDB.Teardown()

	// リポジトリとサービスの初期化
	userRepo := repository.NewUserRepository(testDB.DB)
	pinRepo := repository.NewPinRepository(testDB.DB)
	connectRepo := repository.NewConnectRepository(testDB.DB)
	authService := service.NewAuthService(userRepo)
	// pinService := service.NewPinService(pinRepo)
	connectService := service.NewConnectService(connectRepo, pinRepo)
	connectHandler := NewConnectHandler(connectService)
	router := setupConnectTestRouter(connectHandler)

	// テストヘルパーの作成
	helper := database.NewTestHelper(testDB)

	t.Run("成功: 自分が所有するConnectの更新", func(t *testing.T) {
		defer testDB.CleanupData()

		// テストユーザーの作成
		user, err := helper.CreateTestUser("test@example.com", "password123", "Test User")
		require.NoError(t, err)

		// テストPinの作成
		pin1, err := helper.CreateTestPin(user.ID, "トイレA", 35.6895, 139.6917)
		require.NoError(t, err)
		pin2, err := helper.CreateTestPin(user.ID, "トイレB", 35.7000, 139.7000)
		require.NoError(t, err)
		pin3, err := helper.CreateTestPin(user.ID, "トイレC", 35.7100, 139.7100)
		require.NoError(t, err)

		// テストConnectの作成
		connect, err := helper.CreateTestConnect(user.ID, pin1.ID, pin2.ID, true)
		require.NoError(t, err)

		// トークンの生成
		token, _, err := authService.Login(context.Background(), user.Email, "password123")
		require.NoError(t, err)

		showFalse := false
		reqBody := model.UpdateConnectRequest{
			PinID1: pin1.ID,
			PinID2: []string{pin3.ID},
			Show:   &showFalse,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/connects/"+connect.ID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// 要件: 9.6 - 更新されたConnect情報を含むレスポンスを返す
		assert.Equal(t, http.StatusOK, w.Code)

		var updatedConnect model.Connect
		err = json.Unmarshal(w.Body.Bytes(), &updatedConnect)
		require.NoError(t, err)
		assert.Equal(t, connect.ID, updatedConnect.ID)
		assert.Equal(t, pin1.ID, updatedConnect.PinID1)
		assert.Len(t, updatedConnect.PinID2, 1)
		assert.Equal(t, pin3.ID, updatedConnect.PinID2[0])
		assert.False(t, updatedConnect.Show)
	})

	t.Run("成功: showフィールドのみ更新", func(t *testing.T) {
		defer testDB.CleanupData()

		// テストユーザーの作成
		user, err := helper.CreateTestUser("test@example.com", "password123", "Test User")
		require.NoError(t, err)

		// テストPinの作成
		pin1, err := helper.CreateTestPin(user.ID, "トイレA", 35.6895, 139.6917)
		require.NoError(t, err)
		pin2, err := helper.CreateTestPin(user.ID, "トイレB", 35.7000, 139.7000)
		require.NoError(t, err)

		// テストConnectの作成
		connect, err := helper.CreateTestConnect(user.ID, pin1.ID, pin2.ID, true)
		require.NoError(t, err)

		// トークンの生成
		token, _, err := authService.Login(context.Background(), user.Email, "password123")
		require.NoError(t, err)

		showFalse := false
		reqBody := model.UpdateConnectRequest{
			Show: &showFalse,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/connects/"+connect.ID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var updatedConnect model.Connect
		err = json.Unmarshal(w.Body.Bytes(), &updatedConnect)
		require.NoError(t, err)
		assert.False(t, updatedConnect.Show)
	})

	t.Run("エラー: 他のユーザーのConnectを更新しようとする", func(t *testing.T) {
		defer testDB.CleanupData()

		// テストユーザー1の作成
		user1, err := helper.CreateTestUser("user1@example.com", "password123", "User 1")
		require.NoError(t, err)

		// テストユーザー2の作成
		user2, err := helper.CreateTestUser("user2@example.com", "password123", "User 2")
		require.NoError(t, err)

		// ユーザー1のPinを作成
		pin1, err := helper.CreateTestPin(user1.ID, "トイレA", 35.6895, 139.6917)
		require.NoError(t, err)
		pin2, err := helper.CreateTestPin(user1.ID, "トイレB", 35.7000, 139.7000)
		require.NoError(t, err)

		// ユーザー1のConnectを作成
		connect, err := helper.CreateTestConnect(user1.ID, pin1.ID, pin2.ID, true)
		require.NoError(t, err)

		// ユーザー2のトークンを生成
		token, _, err := authService.Login(context.Background(), user2.Email, "password123")
		require.NoError(t, err)

		showFalse := false
		reqBody := model.UpdateConnectRequest{
			Show: &showFalse,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/connects/"+connect.ID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// 要件: 9.4 - 権限エラーレスポンスを返す
		assert.Equal(t, http.StatusForbidden, w.Code)

		var errResp model.ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, model.ErrCodeForbidden, errResp.Error.Code)
	})

	t.Run("エラー: 存在しないConnectの更新", func(t *testing.T) {
		defer testDB.CleanupData()

		// テストユーザーの作成
		user, err := helper.CreateTestUser("test@example.com", "password123", "Test User")
		require.NoError(t, err)

		// トークンの生成
		token, _, err := authService.Login(context.Background(), user.Email, "password123")
		require.NoError(t, err)

		showFalse := false
		reqBody := model.UpdateConnectRequest{
			Show: &showFalse,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/connects/nonexistent-id", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("エラー: 存在しないPinを指定して更新", func(t *testing.T) {
		defer testDB.CleanupData()

		// テストユーザーの作成
		user, err := helper.CreateTestUser("test@example.com", "password123", "Test User")
		require.NoError(t, err)

		// テストPinの作成
		pin1, err := helper.CreateTestPin(user.ID, "トイレA", 35.6895, 139.6917)
		require.NoError(t, err)
		pin2, err := helper.CreateTestPin(user.ID, "トイレB", 35.7000, 139.7000)
		require.NoError(t, err)

		// テストConnectの作成
		connect, err := helper.CreateTestConnect(user.ID, pin1.ID, pin2.ID, true)
		require.NoError(t, err)

		// トークンの生成
		token, _, err := authService.Login(context.Background(), user.Email, "password123")
		require.NoError(t, err)

		reqBody := model.UpdateConnectRequest{
			PinID1: pin1.ID,
			PinID2: []string{"nonexistent-pin"},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/connects/"+connect.ID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// 要件: 9.5 - Pin存在確認
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// TestConnectHandler_GetConnects はConnect一覧取得エンドポイントのテスト
// 要件: 8.1, 9.1
func TestConnectHandler_GetConnects(t *testing.T) {
	// テストデータベースのセットアップ
	testDB, err := database.SetupTestDB()
	require.NoError(t, err)
	defer testDB.Teardown()

	// リポジトリとサービスの初期化
	userRepo := repository.NewUserRepository(testDB.DB)
	pinRepo := repository.NewPinRepository(testDB.DB)
	connectRepo := repository.NewConnectRepository(testDB.DB)
	authService := service.NewAuthService(userRepo)
	// pinService := service.NewPinService(pinRepo)
	connectService := service.NewConnectService(connectRepo, pinRepo)
	connectHandler := NewConnectHandler(connectService)
	router := setupConnectTestRouter(connectHandler)

	// テストヘルパーの作成
	helper := database.NewTestHelper(testDB)

	t.Run("成功: ユーザーのConnect一覧を取得", func(t *testing.T) {
		defer testDB.CleanupData()

		// テストユーザーの作成
		user, err := helper.CreateTestUser("test@example.com", "password123", "Test User")
		require.NoError(t, err)

		// テストPinの作成
		pin1, err := helper.CreateTestPin(user.ID, "トイレA", 35.6895, 139.6917)
		require.NoError(t, err)
		pin2, err := helper.CreateTestPin(user.ID, "トイレB", 35.7000, 139.7000)
		require.NoError(t, err)
		pin3, err := helper.CreateTestPin(user.ID, "トイレC", 35.7100, 139.7100)
		require.NoError(t, err)

		// テストConnectの作成
		connect1, err := helper.CreateTestConnect(user.ID, pin1.ID, pin2.ID, true)
		require.NoError(t, err)
		connect2, err := helper.CreateTestConnect(user.ID, pin2.ID, pin3.ID, false)
		require.NoError(t, err)

		// トークンの生成
		token, _, err := authService.Login(context.Background(), user.Email, "password123")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/connects/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var connects []model.Connect
		err = json.Unmarshal(w.Body.Bytes(), &connects)
		require.NoError(t, err)
		assert.Len(t, connects, 2)

		// Connect IDの確認
		connectIDs := []string{connects[0].ID, connects[1].ID}
		assert.Contains(t, connectIDs, connect1.ID)
		assert.Contains(t, connectIDs, connect2.ID)
	})

	t.Run("成功: Connectが0件の場合", func(t *testing.T) {
		defer testDB.CleanupData()

		// テストユーザーの作成
		user, err := helper.CreateTestUser("test@example.com", "password123", "Test User")
		require.NoError(t, err)

		// トークンの生成
		token, _, err := authService.Login(context.Background(), user.Email, "password123")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/connects/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var connects []model.Connect
		err = json.Unmarshal(w.Body.Bytes(), &connects)
		require.NoError(t, err)
		assert.Len(t, connects, 0)
	})

	t.Run("エラー: 未認証のリクエスト", func(t *testing.T) {
		defer testDB.CleanupData()

		req := httptest.NewRequest(http.MethodGet, "/api/connects/", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// TestConnectHandler_DeleteConnect はConnect削除エンドポイントのテスト
// 要件: 9.1, 9.6
func TestConnectHandler_DeleteConnect(t *testing.T) {
	// テストデータベースのセットアップ
	testDB, err := database.SetupTestDB()
	require.NoError(t, err)
	defer testDB.Teardown()

	// リポジトリとサービスの初期化
	userRepo := repository.NewUserRepository(testDB.DB)
	pinRepo := repository.NewPinRepository(testDB.DB)
	connectRepo := repository.NewConnectRepository(testDB.DB)
	authService := service.NewAuthService(userRepo)
	// pinService := service.NewPinService(pinRepo)
	connectService := service.NewConnectService(connectRepo, pinRepo)
	connectHandler := NewConnectHandler(connectService)
	router := setupConnectTestRouter(connectHandler)

	// テストヘルパーの作成
	helper := database.NewTestHelper(testDB)

	t.Run("成功: 自分が所有するConnectの削除", func(t *testing.T) {
		defer testDB.CleanupData()

		// テストユーザーの作成
		user, err := helper.CreateTestUser("test@example.com", "password123", "Test User")
		require.NoError(t, err)

		// テストPinの作成
		pin1, err := helper.CreateTestPin(user.ID, "トイレA", 35.6895, 139.6917)
		require.NoError(t, err)
		pin2, err := helper.CreateTestPin(user.ID, "トイレB", 35.7000, 139.7000)
		require.NoError(t, err)

		// テストConnectの作成
		connect, err := helper.CreateTestConnect(user.ID, pin1.ID, pin2.ID, true)
		require.NoError(t, err)

		// トークンの生成
		token, _, err := authService.Login(context.Background(), user.Email, "password123")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodDelete, "/api/connects/"+connect.ID, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// 要件: 9.6 - 成功レスポンスを返す
		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]string
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Contains(t, resp["message"], "deleted successfully")

		// Connectが削除されたことを確認
		connects, err := connectService.GetConnectsByUser(context.Background(), user.ID)
		require.NoError(t, err)
		assert.Len(t, connects, 0)
	})

	t.Run("エラー: 他のユーザーのConnectを削除しようとする", func(t *testing.T) {
		defer testDB.CleanupData()

		// テストユーザー1の作成
		user1, err := helper.CreateTestUser("user1@example.com", "password123", "User 1")
		require.NoError(t, err)

		// テストユーザー2の作成
		user2, err := helper.CreateTestUser("user2@example.com", "password123", "User 2")
		require.NoError(t, err)

		// ユーザー1のPinを作成
		pin1, err := helper.CreateTestPin(user1.ID, "トイレA", 35.6895, 139.6917)
		require.NoError(t, err)
		pin2, err := helper.CreateTestPin(user1.ID, "トイレB", 35.7000, 139.7000)
		require.NoError(t, err)

		// ユーザー1のConnectを作成
		connect, err := helper.CreateTestConnect(user1.ID, pin1.ID, pin2.ID, true)
		require.NoError(t, err)

		// ユーザー2のトークンを生成
		token, _, err := authService.Login(context.Background(), user2.Email, "password123")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodDelete, "/api/connects/"+connect.ID, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// 要件: 9.4 - 権限エラーレスポンスを返す
		assert.Equal(t, http.StatusForbidden, w.Code)

		var errResp model.ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, model.ErrCodeForbidden, errResp.Error.Code)
	})

	t.Run("エラー: 存在しないConnectの削除", func(t *testing.T) {
		defer testDB.CleanupData()

		// テストユーザーの作成
		user, err := helper.CreateTestUser("test@example.com", "password123", "Test User")
		require.NoError(t, err)

		// トークンの生成
		token, _, err := authService.Login(context.Background(), user.Email, "password123")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodDelete, "/api/connects/nonexistent-id", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("エラー: 未認証のリクエスト", func(t *testing.T) {
		defer testDB.CleanupData()

		req := httptest.NewRequest(http.MethodDelete, "/api/connects/some-id", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
