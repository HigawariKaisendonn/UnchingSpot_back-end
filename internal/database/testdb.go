package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

// TestDB はテスト用データベースの管理構造体
type TestDB struct {
	DB       *sqlx.DB
	pool     *dockertest.Pool
	resource *dockertest.Resource
}

// SetupTestDB はdockertestを使用してテスト用PostgreSQLコンテナを起動し、
// マイグレーションを実行してデータベースを初期化します
func SetupTestDB() (*TestDB, error) {
	// Dockerプールの作成
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, fmt.Errorf("could not construct pool: %w", err)
	}

	// Dockerデーモンへの接続確認
	err = pool.Client.Ping()
	if err != nil {
		return nil, fmt.Errorf("could not connect to Docker: %w", err)
	}

	// PostgreSQL + PostGISコンテナの起動
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgis/postgis",
		Tag:        "15-3.3",
		Env: []string{
			"POSTGRES_PASSWORD=testpass",
			"POSTGRES_USER=testuser",
			"POSTGRES_DB=testdb",
			"listen_addresses='*'",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		return nil, fmt.Errorf("could not start resource: %w", err)
	}

	// コンテナの有効期限を設定（テストが長時間実行される場合の保護）
	if err := resource.Expire(120); err != nil {
		return nil, fmt.Errorf("could not set expiration: %w", err)
	}

	// データベース接続文字列の構築
	hostAndPort := resource.GetHostPort("5432/tcp")
	databaseURL := fmt.Sprintf("postgres://testuser:testpass@%s/testdb?sslmode=disable", hostAndPort)

	log.Printf("Connecting to database on url: %s", databaseURL)

	// データベースが起動するまで待機（リトライ）
	var db *sqlx.DB
	pool.MaxWait = 60 * time.Second
	if err = pool.Retry(func() error {
		var err error
		db, err = sqlx.Connect("postgres", databaseURL)
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		return nil, fmt.Errorf("could not connect to database: %w", err)
	}

	log.Println("Successfully connected to test database")

	// PostGIS拡張機能の有効化
	if err := enablePostGIS(db.DB); err != nil {
		return nil, fmt.Errorf("could not enable PostGIS: %w", err)
	}

	// マイグレーションの実行
	if err := runMigrations(db.DB); err != nil {
		return nil, fmt.Errorf("could not run migrations: %w", err)
	}

	log.Println("Migrations completed successfully")

	return &TestDB{
		DB:       db,
		pool:     pool,
		resource: resource,
	}, nil
}

// enablePostGIS はPostGIS拡張機能を有効化します
func enablePostGIS(db *sql.DB) error {
	_, err := db.Exec("CREATE EXTENSION IF NOT EXISTS postgis")
	if err != nil {
		return fmt.Errorf("failed to enable PostGIS: %w", err)
	}
	return nil
}

// runMigrations はマイグレーションファイルを実行します
func runMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://../../migrations",
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not run migrations: %w", err)
	}

	return nil
}

// Teardown はテスト用データベースをクリーンアップし、Dockerコンテナを削除します
func (tdb *TestDB) Teardown() error {
	if tdb.DB != nil {
		if err := tdb.DB.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}

	if tdb.pool != nil && tdb.resource != nil {
		if err := tdb.pool.Purge(tdb.resource); err != nil {
			return fmt.Errorf("could not purge resource: %w", err)
		}
	}

	log.Println("Test database cleaned up successfully")
	return nil
}

// CleanupData はテストデータをクリーンアップします（テーブルのデータを削除）
func (tdb *TestDB) CleanupData() error {
	// 外部キー制約を考慮して、依存関係の逆順で削除
	tables := []string{"connect", "pins", "users"}
	
	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s", table)
		if _, err := tdb.DB.Exec(query); err != nil {
			return fmt.Errorf("could not cleanup table %s: %w", table, err)
		}
	}

	log.Println("Test data cleaned up successfully")
	return nil
}
