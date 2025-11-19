package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/higawarikaisendonn/unchingspot-backend/internal/database"
	"github.com/higawarikaisendonn/unchingspot-backend/internal/handler"
	"github.com/higawarikaisendonn/unchingspot-backend/internal/middleware"
	"github.com/higawarikaisendonn/unchingspot-backend/internal/repository"
	"github.com/higawarikaisendonn/unchingspot-backend/internal/service"
	"github.com/joho/godotenv"
)

func main() {
	// 環境変数の読み込み
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// ポート番号の取得
	port := os.Getenv("PORT")
	if port == "" {
		port = "8088"
	}

	// データベース接続の初期化
	if err := database.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// データベース接続のヘルスチェック
	if err := database.HealthCheck(); err != nil {
		log.Fatalf("Database health check failed: %v", err)
	}

	log.Println("Database connection established successfully")

	// リポジトリの初期化
	db := database.GetDB()
	userRepo := repository.NewUserRepository(db)
	pinRepo := repository.NewPinRepository(db)
	connectRepo := repository.NewConnectRepository(db)

	// サービスの初期化
	authService := service.NewAuthService(userRepo)
	pinService := service.NewPinService(pinRepo)
	connectService := service.NewConnectService(connectRepo, pinRepo)

	// ハンドラーの初期化
	authHandler := handler.NewAuthHandler(authService)
	pinHandler := handler.NewPinHandler(pinService)
	connectHandler := handler.NewConnectHandler(connectService)

	// Chi routerのセットアップ
	r := chi.NewRouter()

	// グローバルミドルウェアの適用
	r.Use(middleware.LoggerMiddleware)
	r.Use(middleware.CORSMiddleware)

	// ルーティング設定
	r.Route("/api", func(r chi.Router) {
		// 認証エンドポイント（認証不要）
		r.Route("/auth", func(r chi.Router) {
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

		// Pinエンドポイント（全て認証が必要）
		r.Route("/pins", func(r chi.Router) {
			r.Use(middleware.AuthMiddleware)
			r.Post("/", pinHandler.CreatePin)
			r.Get("/", pinHandler.GetPins)
			r.Get("/{id}", pinHandler.GetPin)
			r.Put("/{id}", pinHandler.UpdatePin)
			r.Delete("/{id}", pinHandler.DeletePin)
		})

		// Connectエンドポイント（全て認証が必要）
		r.Route("/connects", func(r chi.Router) {
			r.Use(middleware.AuthMiddleware)
			r.Post("/", connectHandler.CreateConnect)
			r.Get("/", connectHandler.GetConnects)
			r.Put("/{id}", connectHandler.UpdateConnect)
			r.Delete("/{id}", connectHandler.DeleteConnect)
		})
	})

	// HTTPサーバーの設定
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// サーバーを別のゴルーチンで起動
	go func() {
		log.Printf("うんちんぐすぽっと API server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// グレースフルシャットダウンの設定
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// シャットダウンのタイムアウト設定（30秒）
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// サーバーのグレースフルシャットダウン
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
