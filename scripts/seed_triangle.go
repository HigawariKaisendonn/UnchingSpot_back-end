package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/higawarikaisendonn/unchingspot-backend/internal/database"
	"github.com/higawarikaisendonn/unchingspot-backend/internal/model"
	"github.com/higawarikaisendonn/unchingspot-backend/internal/repository"
	"github.com/joho/godotenv"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// .envファイルを読み込み
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// データベース接続
	config := database.NewConfig()
	db, err := database.Connect(config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// リポジトリの初期化
	userRepo := repository.NewUserRepository(db)
	pinRepo := repository.NewPinRepository(db)
	connectRepo := repository.NewConnectRepository(db)

	// 既存のテストデータをクリーンアップ
	fmt.Println("Cleaning up existing test data...")
	_, err = db.ExecContext(ctx, "DELETE FROM connect WHERE user_id IN (SELECT id FROM users WHERE email = 'triangle@example.com')")
	if err != nil {
		log.Printf("Warning: Failed to delete connects: %v", err)
	}
	_, err = db.ExecContext(ctx, "DELETE FROM pins WHERE user_id IN (SELECT id FROM users WHERE email = 'triangle@example.com')")
	if err != nil {
		log.Printf("Warning: Failed to delete pins: %v", err)
	}
	_, err = db.ExecContext(ctx, "DELETE FROM users WHERE email = 'triangle@example.com'")
	if err != nil {
		log.Printf("Warning: Failed to delete user: %v", err)
	}
	fmt.Println("✓ Cleanup completed")

	// テストユーザーの作成
	fmt.Println("\nCreating test user...")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("testpass123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	user := &model.User{
		ID:       uuid.New().String(),
		Email:    "triangle@example.com",
		Password: string(hashedPassword),
		Name:     "三角形テストユーザー",
	}

	if err := userRepo.Create(ctx, user); err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}
	fmt.Printf("✓ User created: %s (ID: %s)\n", user.Name, user.ID)

	// 3つの地点を作成（三角形の頂点）
	fmt.Println("\nCreating pins for triangle...")

	// 頂点1（開始/終了地点）
	pin1 := &model.Pin{
		ID:        uuid.New().String(),
		Name:      "頂点A",
		UserID:    user.ID,
		Latitude:  35.6812,
		Longitude: 139.7671,
	}
	if err := pinRepo.Create(ctx, pin1); err != nil {
		log.Fatalf("Failed to create pin1: %v", err)
	}
	fmt.Printf("✓ Pin 1 created: %s (%.4f, %.4f)\n", pin1.Name, pin1.Latitude, pin1.Longitude)

	// 頂点2
	pin2 := &model.Pin{
		ID:        uuid.New().String(),
		Name:      "頂点B",
		UserID:    user.ID,
		Latitude:  35.6900,
		Longitude: 139.7700,
	}
	if err := pinRepo.Create(ctx, pin2); err != nil {
		log.Fatalf("Failed to create pin2: %v", err)
	}
	fmt.Printf("✓ Pin 2 created: %s (%.4f, %.4f)\n", pin2.Name, pin2.Latitude, pin2.Longitude)

	// 頂点3
	pin3 := &model.Pin{
		ID:        uuid.New().String(),
		Name:      "頂点C",
		UserID:    user.ID,
		Latitude:  35.6850,
		Longitude: 139.7750,
	}
	if err := pinRepo.Create(ctx, pin3); err != nil {
		log.Fatalf("Failed to create pin3: %v", err)
	}
	fmt.Printf("✓ Pin 3 created: %s (%.4f, %.4f)\n", pin3.Name, pin3.Latitude, pin3.Longitude)

	// Connectを作成（三角形）
	fmt.Println("\nCreating triangle connect...")
	connect := &model.Connect{
		ID:     uuid.New().String(),
		UserID: user.ID,
		Name:   "三角形",
		PinID1: pin1.ID,
		PinID2: pq.StringArray{pin2.ID, pin3.ID},
		Show:   true,
	}

	if err := connectRepo.Create(ctx, connect); err != nil {
		log.Fatalf("Failed to create connect: %v", err)
	}
	fmt.Printf("✓ Connect created: %s (ID: %s)\n", connect.Name, connect.ID)
	fmt.Printf("  - Start/End Point: %s\n", pin1.Name)
	fmt.Printf("  - Intermediate Points:\n")
	fmt.Printf("    1. %s\n", pin2.Name)
	fmt.Printf("    2. %s\n", pin3.Name)

	// 作成されたデータを確認
	fmt.Println("\n=== Triangle Test Data Summary ===")
	fmt.Printf("User Email: %s\n", user.Email)
	fmt.Printf("Password: testpass123\n")
	fmt.Printf("User ID: %s\n", user.ID)
	fmt.Printf("\nPins (Triangle Vertices):\n")
	fmt.Printf("  1. %s (ID: %s) - %.4f, %.4f\n", pin1.Name, pin1.ID, pin1.Latitude, pin1.Longitude)
	fmt.Printf("  2. %s (ID: %s) - %.4f, %.4f\n", pin2.Name, pin2.ID, pin2.Latitude, pin2.Longitude)
	fmt.Printf("  3. %s (ID: %s) - %.4f, %.4f\n", pin3.Name, pin3.ID, pin3.Latitude, pin3.Longitude)
	fmt.Printf("\nConnect: %s (ID: %s)\n", connect.Name, connect.ID)
	fmt.Printf("Shape: Triangle with 3 vertices\n")

	fmt.Println("\n✅ Triangle test data created successfully!")
}
