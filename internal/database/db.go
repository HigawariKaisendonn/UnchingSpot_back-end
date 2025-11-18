package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// DB はグローバルなデータベース接続インスタンス
var DB *sqlx.DB

// Config はデータベース接続設定
type Config struct {
	DatabaseURL      string
	MaxOpenConns     int
	MaxIdleConns     int
	ConnMaxLifetime  time.Duration
	ConnMaxIdleTime  time.Duration
}

// NewConfig は環境変数から設定を読み込む
func NewConfig() *Config {
	return &Config{
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		MaxOpenConns:    25,  // 最大オープン接続数
		MaxIdleConns:    5,   // 最大アイドル接続数
		ConnMaxLifetime: time.Hour,      // 接続の最大生存時間
		ConnMaxIdleTime: time.Minute * 5, // アイドル接続の最大時間
	}
}

// Connect はデータベースに接続し、接続プールを設定する
func Connect(config *Config) (*sqlx.DB, error) {
	if config.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	// データベースに接続
	db, err := sqlx.Connect("postgres", config.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 接続プールの設定
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// 接続テスト
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to database")
	return db, nil
}

// Init はデータベース接続を初期化し、グローバル変数に設定する
func Init() error {
	config := NewConfig()
	db, err := Connect(config)
	if err != nil {
		return err
	}
	
	DB = db
	return nil
}

// Close はデータベース接続を閉じる
func Close() error {
	if DB != nil {
		log.Println("Closing database connection")
		return DB.Close()
	}
	return nil
}

// GetDB はグローバルなデータベース接続を返す
func GetDB() *sqlx.DB {
	return DB
}

// HealthCheck はデータベース接続の健全性をチェックする
func HealthCheck() error {
	if DB == nil {
		return fmt.Errorf("database connection is not initialized")
	}
	
	if err := DB.Ping(); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	
	return nil
}
