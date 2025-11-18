# Database Connection Utility

このパッケージは、PostgreSQL + PostGISデータベースへの接続とコネクションプール管理を提供します。

## 機能

- sqlxを使用したデータベース接続
- コネクションプールの設定と管理
- 接続のヘルスチェック
- グレースフルシャットダウン対応

## 使用方法

### 基本的な使用

```go
package main

import (
    "log"
    "github.com/higawarikaisendonn/unchingspot-backend/internal/database"
)

func main() {
    // データベース接続の初期化
    if err := database.Init(); err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    defer database.Close()

    // グローバルなDB接続を取得
    db := database.GetDB()
    
    // クエリの実行例
    var count int
    err := db.Get(&count, "SELECT COUNT(*) FROM users")
    if err != nil {
        log.Printf("Query failed: %v", err)
    }
}
```

### カスタム設定での接続

```go
config := &database.Config{
    DatabaseURL:     "postgres://user:pass@localhost:5432/dbname",  // 例
    MaxOpenConns:    50,
    MaxIdleConns:    10,
    ConnMaxLifetime: time.Hour * 2,
    ConnMaxIdleTime: time.Minute * 10,
}

db, err := database.Connect(config)
if err != nil {
    log.Fatalf("Failed to connect: %v", err)
}
defer db.Close()
```

### ヘルスチェック

```go
if err := database.HealthCheck(); err != nil {
    log.Printf("Database is unhealthy: %v", err)
}
```

## 環境変数

以下の環境変数が必要です：

- `DATABASE_URL`: PostgreSQL接続文字列（例: `postgres://user:password@localhost:5432/dbname?sslmode=disable`）

## コネクションプール設定

デフォルトの設定値：

- **MaxOpenConns**: 25 - 同時に開ける最大接続数
- **MaxIdleConns**: 5 - アイドル状態で保持する最大接続数
- **ConnMaxLifetime**: 1時間 - 接続の最大生存時間
- **ConnMaxIdleTime**: 5分 - アイドル接続の最大時間

これらの値は本番環境の負荷に応じて調整してください。

## PostGIS対応

このユーティリティはPostGIS拡張を使用するPostgreSQLデータベースに対応しています。
PostGISの地理空間クエリを実行する際は、sqlxの標準的なクエリメソッドを使用できます。

```go
type Location struct {
    ID        string  `db:"id"`
    Name      string  `db:"name"`
    Longitude float64 `db:"longitude"`
    Latitude  float64 `db:"latitude"`
}

var locations []Location
query := `
    SELECT 
        id, 
        name,
        ST_X(location) as longitude,
        ST_Y(location) as latitude
    FROM pins
    WHERE deleted_at IS NULL
`
err := db.Select(&locations, query)
```
