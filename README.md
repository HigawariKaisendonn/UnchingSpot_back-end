# UnchingSpot Backend
うんちんぐすぽっと - バックエンドAPI

## 概要

UnchingSpotは、ユーザーがトイレを使用した場所をマーキングし、複数の位置情報を接続して管理できるバックエンドシステムです。Go言語で実装されたRESTful APIサーバーで、PostgreSQL（PostGIS拡張）をデータベースとして使用し、Docker環境で動作します。

## 主な機能

- **ユーザー認証**: JWT（JSON Web Token）を使用した安全な認証システム
- **Pin管理**: トイレの位置情報（緯度経度）を登録・更新・削除
- **Connect管理**: 2つのPinを接続してエリアを作成
- **地理空間検索**: PostGISを使用した位置情報ベースの検索
- **権限管理**: ユーザーは自分が作成したPinとConnectのみ操作可能

## 技術スタック

- **言語**: Go 1.23+
- **Webフレームワーク**: Chi router
- **データベース**: PostgreSQL 15+ with PostGIS 3.3+
- **ORMライブラリ**: sqlx
- **認証**: JWT (JSON Web Token)
- **パスワードハッシュ**: bcrypt
- **環境変数管理**: godotenv
- **マイグレーション**: golang-migrate
- **コンテナ**: Docker & Docker Compose
- **テスト**: Go標準testing、testify、dockertest

## クイックスタート

### 前提条件

- Go 1.23以上
- Docker & Docker Compose
- Git
- golang-migrate（マイグレーション用）

### セットアップ手順

#### 1. リポジトリをクローン

```bash
git clone <repository-url>
cd UnchingSpot_back-end
```

#### 2. 環境変数を設定

```bash
cp .env.example .env
```

`.env`ファイルを編集して、必要な環境変数を設定してください：

```env
# Database
DATABASE_URL=postgres://postgres:postgres@localhost:5432/unchingspot?sslmode=disable

# JWT
JWT_SECRET=your-super-secret-jwt-key-change-in-production

# Server
PORT=8088

# CORS
FRONTEND_URL=http://localhost:3000

# Environment
ENV=development
```

#### 3. 依存関係をインストール

```bash
go mod download
```

#### 4. マイグレーションツールをインストール

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

#### 5. Docker環境を起動

```bash
# データベースとAPIサーバーを起動
docker-compose up -d

# ログを確認
docker-compose logs -f
```

または、開発環境でデータベースのみ起動する場合：

```bash
# データベースのみ起動
docker-compose up -d db

# データベースの起動を確認
docker-compose ps
```

#### 6. マイグレーションを実行

```bash
# Windows
scripts\migrate.bat up

# Linux/macOS
chmod +x scripts/migrate.sh
./scripts/migrate.sh up
```

#### 7. APIサーバーを起動（ローカル開発の場合）

```bash
go run cmd/api/main.go
```

サーバーは `http://localhost:8088` で起動します。

### 動作確認

データベース接続テスト：

```bash
curl http://localhost:8088/api/auth/test
```

成功すると以下のレスポンスが返ります：

```json
{
  "message": "Database connection successful"
}
```

## API仕様書

### 認証

すべての保護されたエンドポイントには、Authorizationヘッダーが必要です：

```
Authorization: Bearer <JWT_TOKEN>
```

### エンドポイント一覧

#### 認証エンドポイント

##### POST /api/auth/signup
ユーザー登録

**リクエスト:**
```json
{
  "email": "user@example.com",
  "password": "password123",
  "name": "ユーザー名"
}
```

**レスポンス (201 Created):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "uuid",
    "name": "ユーザー名",
    "email": "user@example.com",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

##### POST /api/auth/login
ログイン

**リクエスト:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**レスポンス (200 OK):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "uuid",
    "name": "ユーザー名",
    "email": "user@example.com",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

##### POST /api/auth/logout
ログアウト（認証必須）

**レスポンス (200 OK):**
```json
{
  "message": "Logged out successfully"
}
```

##### GET /api/auth/me
ログイン中のユーザー情報取得（認証必須）

**レスポンス (200 OK):**
```json
{
  "id": "uuid",
  "name": "ユーザー名",
  "email": "user@example.com",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

##### GET /api/auth/test
データベース接続テスト

**レスポンス (200 OK):**
```json
{
  "message": "Database connection successful"
}
```

#### Pinエンドポイント（すべて認証必須）

##### POST /api/pins
Pin作成

**リクエスト:**
```json
{
  "name": "トイレA",
  "latitude": 35.6895,
  "longitude": 139.6917
}
```

**レスポンス (201 Created):**
```json
{
  "id": "uuid",
  "name": "トイレA",
  "user_id": "uuid",
  "latitude": 35.6895,
  "longitude": 139.6917,
  "created_at": "2024-01-01T00:00:00Z",
  "edited_at": "2024-01-01T00:00:00Z"
}
```

##### GET /api/pins
ログイン中のユーザーのPin一覧取得

**レスポンス (200 OK):**
```json
[
  {
    "id": "uuid",
    "name": "トイレA",
    "user_id": "uuid",
    "latitude": 35.6895,
    "longitude": 139.6917,
    "created_at": "2024-01-01T00:00:00Z",
    "edited_at": "2024-01-01T00:00:00Z"
  }
]
```

##### GET /api/pins/:id
Pin詳細取得

**レスポンス (200 OK):**
```json
{
  "id": "uuid",
  "name": "トイレA",
  "user_id": "uuid",
  "latitude": 35.6895,
  "longitude": 139.6917,
  "created_at": "2024-01-01T00:00:00Z",
  "edited_at": "2024-01-01T00:00:00Z"
}
```

##### PUT /api/pins/:id
Pin更新（自分が作成したPinのみ）

**リクエスト:**
```json
{
  "name": "トイレB",
  "latitude": 35.6896,
  "longitude": 139.6918
}
```

**レスポンス (200 OK):**
```json
{
  "id": "uuid",
  "name": "トイレB",
  "user_id": "uuid",
  "latitude": 35.6896,
  "longitude": 139.6918,
  "created_at": "2024-01-01T00:00:00Z",
  "edited_at": "2024-01-01T01:00:00Z"
}
```

##### DELETE /api/pins/:id
Pin削除（自分が作成したPinのみ）

**レスポンス (200 OK):**
```json
{
  "message": "Pin deleted successfully"
}
```

#### Connectエンドポイント（すべて認証必須）

##### POST /api/connects
Connect作成

**リクエスト:**
```json
{
  "pin_id_1": "uuid1",
  "pin_id_2": "uuid2",
  "show": true
}
```

**レスポンス (201 Created):**
```json
{
  "id": "uuid",
  "user_id": "uuid",
  "pin_id_1": "uuid1",
  "pin_id_2": "uuid2",
  "show": true
}
```

##### GET /api/connects
ログイン中のユーザーのConnect一覧取得

**レスポンス (200 OK):**
```json
[
  {
    "id": "uuid",
    "user_id": "uuid",
    "pin_id_1": "uuid1",
    "pin_id_2": "uuid2",
    "show": true
  }
]
```

##### PUT /api/connects/:id
Connect更新（自分が作成したConnectのみ）

**リクエスト:**
```json
{
  "pin_id_1": "uuid3",
  "pin_id_2": "uuid4",
  "show": false
}
```

**レスポンス (200 OK):**
```json
{
  "id": "uuid",
  "user_id": "uuid",
  "pin_id_1": "uuid3",
  "pin_id_2": "uuid4",
  "show": false
}
```

##### DELETE /api/connects/:id
Connect削除（自分が作成したConnectのみ）

**レスポンス (200 OK):**
```json
{
  "message": "Connect deleted successfully"
}
```

### エラーレスポンス

すべてのエラーは以下の形式で返されます：

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "エラーメッセージ"
  }
}
```

#### エラーコード一覧

- `INVALID_INPUT` (400): 入力検証エラー
- `UNAUTHORIZED` (401): 認証エラー
- `FORBIDDEN` (403): 権限エラー
- `NOT_FOUND` (404): リソースが見つからない
- `CONFLICT` (409): 重複エラー（メール登録済みなど）
- `INTERNAL_SERVER_ERROR` (500): サーバーエラー
- `DATABASE_ERROR` (500): データベースエラー

## ドキュメント

- [マイグレーションガイド](migrations/README.md) - データベースマイグレーションの詳細
- [データベース設計](internal/database/README.md) - データベーススキーマの詳細

## プロジェクト構造

```
.
├── cmd/                    # アプリケーションエントリーポイント
├── internal/               # 内部パッケージ
│   ├── model/             # データモデル
│   ├── repository/        # データアクセス層
│   ├── service/           # ビジネスロジック層
│   ├── handler/           # HTTPハンドラー
│   └── util/              # ユーティリティ関数
├── migrations/            # データベースマイグレーションファイル
├── scripts/               # ビルド・マイグレーションスクリプト
├── docker-compose.yml     # Docker Compose設定
├── .env.example           # 環境変数のサンプル
└── README.md              # このファイル
```

## 開発環境の起動方法

### 方法1: Docker Composeを使用（推奨）

すべてのサービスをコンテナで起動：

```bash
# すべてのサービスを起動
docker-compose up -d

# ログを確認
docker-compose logs -f api

# サービスの状態を確認
docker-compose ps

# サービスを停止
docker-compose down
```

この方法では、データベースとAPIサーバーの両方が自動的に起動し、マイグレーションも実行されます。

### 方法2: ローカル開発（データベースのみDocker）

データベースはDockerで起動し、APIサーバーはローカルで実行：

```bash
# 1. データベースを起動
docker-compose up -d db

# 2. マイグレーションを実行
# Windows
scripts\migrate.bat up

# Linux/macOS
./scripts/migrate.sh up

# 3. APIサーバーをローカルで起動
go run cmd/api/main.go
```

この方法は、コードの変更を即座に反映できるため、開発時に便利です。

### 開発時のワークフロー

#### コードの変更を反映

```bash
# ローカル開発の場合
# サーバーを再起動（Ctrl+Cで停止後）
go run cmd/api/main.go

# Docker開発の場合
docker-compose restart api
```

#### データベースのリセット

```bash
# マイグレーションをすべて戻す
# Windows
scripts\migrate.bat down

# Linux/macOS
./scripts/migrate.sh down

# 再度マイグレーションを適用
# Windows
scripts\migrate.bat up

# Linux/macOS
./scripts/migrate.sh up
```

#### データベースに直接接続

```bash
# PostgreSQLクライアントで接続
docker-compose exec db psql -U postgres -d unchingspot

# または
psql postgres://postgres:postgres@localhost:5432/unchingspot
```

### データベースマイグレーション

#### 新しいマイグレーションを作成

```bash
# Windows
scripts\migrate.bat create migration_name

# Linux/macOS
./scripts/migrate.sh create migration_name
```

これにより、`migrations/`ディレクトリに新しい`.up.sql`と`.down.sql`ファイルが作成されます。

#### マイグレーションを適用

```bash
# すべてのマイグレーションを適用
# Windows
scripts\migrate.bat up

# Linux/macOS
./scripts/migrate.sh up

# 特定のバージョンまで適用
# Windows
scripts\migrate.bat goto 2

# Linux/macOS
./scripts/migrate.sh goto 2
```

#### マイグレーションを戻す

```bash
# 1つ前のバージョンに戻す
# Windows
scripts\migrate.bat down 1

# Linux/macOS
./scripts/migrate.sh down 1

# すべてのマイグレーションを戻す
# Windows
scripts\migrate.bat down

# Linux/macOS
./scripts/migrate.sh down
```

### テスト実行

#### すべてのテストを実行

```bash
go test ./...
```

#### カバレッジ付きでテストを実行

```bash
go test -cover ./...
```

#### 詳細なテスト結果を表示

```bash
go test -v ./...
```

#### 特定のパッケージのテストを実行

```bash
# ハンドラーのテスト
go test ./internal/handler/...

# サービスのテスト
go test ./internal/service/...

# リポジトリのテスト
go test ./internal/repository/...
```

### デバッグ

#### ログレベルの設定

`.env`ファイルで環境変数を設定：

```env
ENV=development  # development, production
```

#### APIリクエストのテスト

```bash
# ユーザー登録
curl -X POST http://localhost:8088/api/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123","name":"テストユーザー"}'

# ログイン
curl -X POST http://localhost:8088/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# Pin作成（認証トークンが必要）
curl -X POST http://localhost:8088/api/pins \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{"name":"トイレA","latitude":35.6895,"longitude":139.6917}'
```

## トラブルシューティング

### データベース接続エラー

```
Error: failed to connect to database
```

**解決方法:**
1. データベースコンテナが起動しているか確認：`docker-compose ps`
2. 環境変数が正しく設定されているか確認：`.env`ファイルを確認
3. データベースの起動を待つ：`docker-compose logs db`

### マイグレーションエラー

```
Error: Dirty database version
```

**解決方法:**
```bash
# マイグレーションのバージョンを強制的にリセット
# Windows
scripts\migrate.bat force 1

# Linux/macOS
./scripts/migrate.sh force 1
```

### ポート競合エラー

```
Error: bind: address already in use
```

**解決方法:**
1. 既存のプロセスを停止
2. `.env`ファイルでポート番号を変更
3. `docker-compose.yml`のポートマッピングを変更

### JWT認証エラー

```
Error: Invalid token
```

**解決方法:**
1. トークンが期限切れでないか確認（24時間有効）
2. 正しいAuthorizationヘッダー形式を使用：`Bearer <token>`
3. JWT_SECRETが正しく設定されているか確認

## デプロイ

### Fly.ioへのデプロイ

```bash
# Fly.io CLIをインストール
curl -L https://fly.io/install.sh | sh

# アプリを作成
fly launch

# PostgreSQLデータベースを作成
fly postgres create --name unchingspot-db --region nrt

# データベースを接続
fly postgres attach unchingspot-db

# シークレットを設定
fly secrets set JWT_SECRET=your-secret-key

# デプロイ
fly deploy
```

## アーキテクチャ

このプロジェクトはクリーンアーキテクチャの原則に従っています：

```
Handler Layer (HTTPリクエスト/レスポンス)
    ↓
Service Layer (ビジネスロジック)
    ↓
Repository Layer (データアクセス)
    ↓
Database (PostgreSQL + PostGIS)
```

各レイヤーは独立しており、テストが容易で保守性が高い設計になっています。

## セキュリティ

- パスワードはbcryptでハッシュ化して保存
- JWT認証による安全なセッション管理
- SQL Injectionを防ぐためのプリペアドステートメント使用
- CORS設定によるクロスオリジンリクエストの制御
- 環境変数による機密情報の管理

## パフォーマンス

- データベース接続プールによる効率的な接続管理
- PostGISのGiSTインデックスによる高速な地理空間検索
- 適切なデータベースインデックスの設定

## ライセンス

[ライセンス情報を記載]

## コントリビューション

プルリクエストを歓迎します。大きな変更の場合は、まずissueを開いて変更内容を議論してください。

## サポート

問題が発生した場合は、GitHubのissueを作成してください。
