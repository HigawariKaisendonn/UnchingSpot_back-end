# 設計書

## 概要

「うんちんぐすぽっと」バックエンドは、Go言語で実装されるRESTful APIサーバーです。PostgreSQL（PostGIS拡張）をデータベースとして使用し、Docker環境で動作します。クリーンアーキテクチャの原則に従い、レイヤー分離とテスタビリティを重視した設計とします。

## アーキテクチャ

### レイヤー構成

```
┌─────────────────────────────────────┐
│         Handler Layer               │  ← HTTPリクエスト/レスポンス処理
│  (auth_handler, pin_handler, etc.)  │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│         Service Layer               │  ← ビジネスロジック
│  (auth_service, pin_service, etc.)  │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│       Repository Layer              │  ← データアクセス
│ (user_repo, pin_repo, connect_repo) │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│      PostgreSQL + PostGIS           │  ← データストレージ
└─────────────────────────────────────┘
```

### 技術スタック

- **言語**: Go 1.21+
- **Webフレームワーク**: Chi router（軽量で高速）
- **データベース**: PostgreSQL 15+ with PostGIS 3.3+
- **ORMライブラリ**: sqlx（軽量なSQL拡張）
- **認証**: JWT（JSON Web Token）
- **パスワードハッシュ**: bcrypt
- **環境変数管理**: godotenv
- **マイグレーション**: golang-migrate
- **コンテナ**: Docker & Docker Compose

## コンポーネントとインターフェース

### 1. Handler Layer

#### AuthHandler
```go
type AuthHandler struct {
    authService AuthService
}

// エンドポイント
POST   /api/auth/signup   - SignUp(w http.ResponseWriter, r *http.Request)
POST   /api/auth/login    - Login(w http.ResponseWriter, r *http.Request)
POST   /api/auth/logout   - Logout(w http.ResponseWriter, r *http.Request)
GET    /api/auth/me       - GetMe(w http.ResponseWriter, r *http.Request)
GET    /api/auth/test     - TestConnection(w http.ResponseWriter, r *http.Request)
```

#### PinHandler
```go
type PinHandler struct {
    pinService PinService
}

// エンドポイント
POST   /api/pins          - CreatePin(w http.ResponseWriter, r *http.Request)
PUT    /api/pins/:id      - UpdatePin(w http.ResponseWriter, r *http.Request)
GET    /api/pins          - GetPins(w http.ResponseWriter, r *http.Request)
GET    /api/pins/:id      - GetPin(w http.ResponseWriter, r *http.Request)
DELETE /api/pins/:id      - DeletePin(w http.ResponseWriter, r *http.Request)
```

#### ConnectHandler
```go
type ConnectHandler struct {
    connectService ConnectService
}

// エンドポイント
POST   /api/connects      - CreateConnect(w http.ResponseWriter, r *http.Request)
PUT    /api/connects/:id  - UpdateConnect(w http.ResponseWriter, r *http.Request)
GET    /api/connects      - GetConnects(w http.ResponseWriter, r *http.Request)
DELETE /api/connects/:id  - DeleteConnect(w http.ResponseWriter, r *http.Request)
```

### 2. Service Layer

#### AuthService
```go
type AuthService interface {
    SignUp(ctx context.Context, email, password, name string) (*User, error)
    Login(ctx context.Context, email, password string) (string, *User, error) // token, user
    ValidateToken(ctx context.Context, token string) (*User, error)
    TestConnection(ctx context.Context) error
}
```

#### PinService
```go
type PinService interface {
    CreatePin(ctx context.Context, userID string, name string, lat, lng float64) (*Pin, error)
    UpdatePin(ctx context.Context, pinID, userID string, name string, lat, lng float64) (*Pin, error)
    GetPin(ctx context.Context, pinID string) (*Pin, error)
    GetPinsByUser(ctx context.Context, userID string) ([]*Pin, error)
    DeletePin(ctx context.Context, pinID, userID string) error
}
```

#### ConnectService
```go
type ConnectService interface {
    CreateConnect(ctx context.Context, userID, pinID1, pinID2 string, show bool) (*Connect, error)
    UpdateConnect(ctx context.Context, connectID, userID, pinID1, pinID2 string, show bool) (*Connect, error)
    GetConnectsByUser(ctx context.Context, userID string) ([]*Connect, error)
    DeleteConnect(ctx context.Context, connectID, userID string) error
}
```

### 3. Repository Layer

#### UserRepository
```go
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    FindByEmail(ctx context.Context, email string) (*User, error)
    FindByID(ctx context.Context, id string) (*User, error)
    Update(ctx context.Context, user *User) error
}
```

#### PinRepository
```go
type PinRepository interface {
    Create(ctx context.Context, pin *Pin) error
    Update(ctx context.Context, pin *Pin) error
    FindByID(ctx context.Context, id string) (*Pin, error)
    FindByUserID(ctx context.Context, userID string) ([]*Pin, error)
    SoftDelete(ctx context.Context, id string) error
}
```

#### ConnectRepository
```go
type ConnectRepository interface {
    Create(ctx context.Context, connect *Connect) error
    Update(ctx context.Context, connect *Connect) error
    FindByID(ctx context.Context, id string) (*Connect, error)
    FindByUserID(ctx context.Context, userID string) ([]*Connect, error)
    Delete(ctx context.Context, id string) error
}
```

## データモデル

### User
```go
type User struct {
    ID        string    `db:"id" json:"id"`
    Name      string    `db:"name" json:"name"`
    Email     string    `db:"email" json:"email"`
    Password  string    `db:"password" json:"-"` // JSONには含めない
    CreatedAt time.Time `db:"created_at" json:"created_at"`
    UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
    DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}
```

### Pin
```go
type Pin struct {
    ID        string     `db:"id" json:"id"`
    Name      string     `db:"name" json:"name"`
    UserID    string     `db:"user_id" json:"user_id"`
    Latitude  float64    `json:"latitude"`  // PostGISから抽出
    Longitude float64    `json:"longitude"` // PostGISから抽出
    CreatedAt time.Time  `db:"created_at" json:"created_at"`
    EditedAt  time.Time  `db:"edit_ad" json:"edited_at"`
    DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

// データベース保存時はST_MakePoint(longitude, latitude)を使用
```

### Connect
```go
type Connect struct {
    ID       string `db:"id" json:"id"`
    UserID   string `db:"user_id" json:"user_id"`
    PinID1   string `db:"pins_id_1" json:"pin_id_1"`
    PinID2   string `db:"pins_id_2" json:"pin_id_2"`
    Show     bool   `db:"show" json:"show"`
}
```

### リクエスト/レスポンス型

```go
// 認証関連
type SignUpRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
    Name     string `json:"name" validate:"required"`
}

type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
    Token string `json:"token"`
    User  *User  `json:"user"`
}

// Pin関連
type CreatePinRequest struct {
    Name      string  `json:"name" validate:"required"`
    Latitude  float64 `json:"latitude" validate:"required,min=-90,max=90"`
    Longitude float64 `json:"longitude" validate:"required,min=-180,max=180"`
}

type UpdatePinRequest struct {
    Name      string  `json:"name"`
    Latitude  float64 `json:"latitude" validate:"min=-90,max=90"`
    Longitude float64 `json:"longitude" validate:"min=-180,max=180"`
}

// Connect関連
type CreateConnectRequest struct {
    PinID1 string `json:"pin_id_1" validate:"required,uuid"`
    PinID2 string `json:"pin_id_2" validate:"required,uuid"`
    Show   bool   `json:"show"`
}

type UpdateConnectRequest struct {
    PinID1 string `json:"pin_id_1" validate:"uuid"`
    PinID2 string `json:"pin_id_2" validate:"uuid"`
    Show   *bool  `json:"show"`
}
```

## データベース設計

### スキーマ定義

```sql
-- UUIDとPostGIS拡張を有効化
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS postgis;

-- usersテーブル
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

-- pinsテーブル
CREATE TABLE pins (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    location GEOMETRY(Point, 4326) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    edit_ad TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_pins_user_id ON pins(user_id);
CREATE INDEX idx_pins_location ON pins USING GIST(location);
CREATE INDEX idx_pins_deleted_at ON pins(deleted_at);

-- connectテーブル
CREATE TABLE connect (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    pins_id_1 UUID NOT NULL REFERENCES pins(id) ON DELETE CASCADE,
    pins_id_2 UUID NOT NULL REFERENCES pins(id) ON DELETE CASCADE,
    show BOOLEAN NOT NULL DEFAULT true
);

CREATE INDEX idx_connect_user_id ON connect(user_id);
CREATE INDEX idx_connect_pins ON connect(pins_id_1, pins_id_2);
```

### PostGISの使用方法

```sql
-- Pinの挿入
INSERT INTO pins (name, user_id, location)
VALUES ('トイレA', 'user-uuid', ST_SetSRID(ST_MakePoint(139.6917, 35.6895), 4326));

-- Pinの取得（緯度経度を抽出）
SELECT 
    id, 
    name, 
    user_id,
    ST_X(location) as longitude,
    ST_Y(location) as latitude,
    created_at,
    edit_ad,
    deleted_at
FROM pins
WHERE id = 'pin-uuid';

-- 距離検索（例：1km以内のPin）
SELECT * FROM pins
WHERE ST_DWithin(
    location::geography,
    ST_SetSRID(ST_MakePoint(139.6917, 35.6895), 4326)::geography,
    1000
);
```

## エラーハンドリング

### エラー型定義

```go
type AppError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Status  int    `json:"-"`
}

// エラーコード
const (
    ErrCodeInvalidInput      = "INVALID_INPUT"
    ErrCodeUnauthorized      = "UNAUTHORIZED"
    ErrCodeForbidden         = "FORBIDDEN"
    ErrCodeNotFound          = "NOT_FOUND"
    ErrCodeConflict          = "CONFLICT"
    ErrCodeInternalServer    = "INTERNAL_SERVER_ERROR"
    ErrCodeDatabaseError     = "DATABASE_ERROR"
)
```

### エラーレスポンス形式

```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid credentials"
  }
}
```

### HTTPステータスコードマッピング

- 200 OK: 成功
- 201 Created: リソース作成成功
- 400 Bad Request: 入力検証エラー
- 401 Unauthorized: 認証エラー
- 403 Forbidden: 権限エラー
- 404 Not Found: リソースが見つからない
- 409 Conflict: 重複エラー（メール登録済みなど）
- 500 Internal Server Error: サーバーエラー

## 認証とセキュリティ

### JWT実装

```go
type JWTClaims struct {
    UserID string `json:"user_id"`
    Email  string `json:"email"`
    jwt.StandardClaims
}

// トークン生成
func GenerateToken(userID, email string) (string, error) {
    claims := JWTClaims{
        UserID: userID,
        Email:  email,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
            IssuedAt:  time.Now().Unix(),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}
```

### 認証ミドルウェア

```go
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            respondError(w, http.StatusUnauthorized, "Missing authorization header")
            return
        }
        
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        claims, err := ValidateToken(tokenString)
        if err != nil {
            respondError(w, http.StatusUnauthorized, "Invalid token")
            return
        }
        
        ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### パスワードハッシュ化

```go
import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

func CheckPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
```

## Docker構成

### docker-compose.yml

```yaml
version: '3.8'

services:
  db:
    image: postgis/postgis:15-3.3
    environment:
      POSTGRES_USER: ${POSTGRES_USER}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      POSTGRES_DB: ${POSTGRES_DB}
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  api:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8088:8088"
    environment:
      DATABASE_URL: ${DATABASE_URL}
      JWT_SECRET: ${JWT_SECRET}
      PORT: ${PORT}
    depends_on:
      db:
        condition: service_healthy
    volumes:
      - .:/app

volumes:
  postgres_data:
```

### Dockerfile

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 依存関係のインストール
COPY go.mod go.sum ./
RUN go mod download

# ソースコードのコピーとビルド
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api

# 実行用の軽量イメージ
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8088

CMD ["./main"]
```

## プロジェクト構造

```
.
├── cmd/
│   └── api/
│       └── main.go              # エントリーポイント
├── internal/
│   ├── handler/
│   │   ├── auth_handler.go
│   │   ├── pin_handler.go
│   │   └── connect_handler.go
│   ├── service/
│   │   ├── auth_service.go
│   │   ├── pin_service.go
│   │   └── connect_service.go
│   ├── repository/
│   │   ├── user_repository.go
│   │   ├── pin_repository.go
│   │   └── connect_repository.go
│   ├── model/
│   │   ├── user.go
│   │   ├── pin.go
│   │   └── connect.go
│   ├── middleware/
│   │   ├── auth.go
│   │   ├── cors.go
│   │   └── logger.go
│   └── util/
│       ├── response.go
│       ├── validator.go
│       └── jwt.go
├── migrations/
│   ├── 000001_create_users_table.up.sql
│   ├── 000001_create_users_table.down.sql
│   ├── 000002_create_pins_table.up.sql
│   ├── 000002_create_pins_table.down.sql
│   ├── 000003_create_connect_table.up.sql
│   └── 000003_create_connect_table.down.sql
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
├── .env.example
└── README.md
```

## テスト戦略

### テストレベル

1. **単体テスト（Unit Tests）**
   - Service層のビジネスロジックテスト
   - Repository層のモック使用
   - カバレッジ目標: 70%以上

2. **統合テスト（Integration Tests）**
   - Handler層のエンドポイントテスト
   - テスト用PostgreSQLコンテナ使用
   - 実際のデータベース操作を含む

### テストツール

- `testing`: Go標準テストパッケージ
- `testify`: アサーションとモック
- `httptest`: HTTPハンドラーテスト
- `dockertest`: テスト用Dockerコンテナ管理

### テストデータベース

```go
// テスト用DB接続
func SetupTestDB(t *testing.T) *sqlx.DB {
    // テスト用の接続情報は環境変数から取得
    testDBURL := os.Getenv("DATABASE_URL")
    db, err := sqlx.Connect("postgres", testDBURL)
    require.NoError(t, err)
    
    // マイグレーション実行
    RunMigrations(db)
    
    t.Cleanup(func() {
        db.Close()
    })
    
    return db
}
```

## CORS設定

Next.jsフロントエンドとの疎通のため、CORSミドルウェアを実装します。

```go
func CORSMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", os.Getenv("FRONTEND_URL"))
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        w.Header().Set("Access-Control-Allow-Credentials", "true")
        
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

## 環境変数

```.env
# Database
# 本番環境では強力なパスワードを使用してください
DATABASE_URL=postgres://<username>:<password>@localhost:5432/unchingspot?sslmode=disable

# JWT
# 生成方法: openssl rand -base64 32
JWT_SECRET=<generate_with_openssl_rand_base64_32>

# Server
PORT=8088

# CORS
# 本番環境ではフロントエンドの実際のURLを指定してください
FRONTEND_URL=http://localhost:3000

# Environment
ENV=development
```

## デプロイ（Fly.io）

### fly.toml

```toml
app = "unchingspot-api"
primary_region = "nrt"

[build]
  dockerfile = "Dockerfile"

[env]
  PORT = "8088"

[[services]]
  internal_port = 8088
  protocol = "tcp"

  [[services.ports]]
    handlers = ["http"]
    port = 80

  [[services.ports]]
    handlers = ["tls", "http"]
    port = 443

[services.concurrency]
  hard_limit = 25
  soft_limit = 20

[[services.tcp_checks]]
  interval = "15s"
  timeout = "2s"
  grace_period = "5s"
```

### デプロイコマンド

```bash
# Fly.io CLIインストール
curl -L https://fly.io/install.sh | sh

# アプリ作成
fly launch

# PostgreSQLデータベース作成
fly postgres create --name unchingspot-db --region nrt

# データベース接続
fly postgres attach unchingspot-db

# シークレット設定（openssl rand -base64 32で生成した値を使用）
fly secrets set JWT_SECRET=<your-generated-secret-key>

# デプロイ
fly deploy
```
