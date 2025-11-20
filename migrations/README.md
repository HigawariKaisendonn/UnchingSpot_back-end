# Database Migrations

このディレクトリには、データベーススキーマのマイグレーションファイルが含まれています。

## マイグレーションファイル

- `000001_create_users_table.up.sql` / `down.sql` - usersテーブルの作成
- `000002_create_pins_table.up.sql` / `down.sql` - pinsテーブルの作成（PostGIS対応）
- `000003_create_connect_table.up.sql` / `down.sql` - connectテーブルの作成
  - `pins_id_1`: 開始点と終了点を表すUUID
  - `pins_id_2`: 複数の中間点を表すUUID配列（図形描画対応）

## マイグレーションの実行方法

### 前提条件

golang-migrate CLIツールをインストールする必要があります：

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### スクリプトを使用した実行

#### Windows:
```cmd
# マイグレーションを実行（最新まで）
scripts\migrate.bat up

# マイグレーションをロールバック
scripts\migrate.bat down

# 現在のバージョンを確認
scripts\migrate.bat version

# 特定のバージョンに強制設定
scripts\migrate.bat force <version>

# 新しいマイグレーションファイルを作成
scripts\migrate.bat create <migration_name>
```

#### Linux/macOS:
```bash
# スクリプトに実行権限を付与
chmod +x scripts/migrate.sh

# マイグレーションを実行（最新まで）
./scripts/migrate.sh up

# マイグレーションをロールバック
./scripts/migrate.sh down

# 現在のバージョンを確認
./scripts/migrate.sh version

# 特定のバージョンに強制設定
./scripts/migrate.sh force <version>

# 新しいマイグレーションファイルを作成
./scripts/migrate.sh create <migration_name>
```

### 直接コマンドで実行

```bash
# マイグレーションを実行
migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/unchingspot?sslmode=disable" up

# マイグレーションをロールバック
migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/unchingspot?sslmode=disable" down

# 現在のバージョンを確認
migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/unchingspot?sslmode=disable" version
```

### Goコードから実行

アプリケーション起動時に自動的にマイグレーションを実行する場合：

```go
import "github.com/yourusername/unchingspot-backend/internal/util"

func main() {
    databaseURL := os.Getenv("DATABASE_URL")
    
    // マイグレーションを実行
    if err := util.RunMigrations(databaseURL); err != nil {
        log.Fatalf("Failed to run migrations: %v", err)
    }
    
    // アプリケーションの起動処理...
}
```

## トラブルシューティング

### "Dirty database version" エラー

マイグレーションが途中で失敗した場合、データベースが "dirty" 状態になることがあります。
この場合、以下のコマンドで強制的にバージョンを設定できます：

```bash
# Windows
scripts\migrate.bat force <version>

# Linux/macOS
./scripts/migrate.sh force <version>
```

### PostgreSQL接続エラー

- データベースが起動していることを確認してください
- `.env`ファイルの`DATABASE_URL`が正しいことを確認してください
- Docker Composeを使用している場合は、`docker-compose up -d db`でデータベースを起動してください

## 新しいマイグレーションの作成

新しいマイグレーションファイルを作成する場合：

```bash
# Windows
scripts\migrate.bat create add_new_column

# Linux/macOS
./scripts/migrate.sh create add_new_column
```

これにより、`migrations/`ディレクトリに新しい`.up.sql`と`.down.sql`ファイルが作成されます。
