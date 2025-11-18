# UnchingSpot Backend
うんちんぐすぽっと - バックエンドAPI

## 概要

UnchingSpotは、トイレの位置情報を共有・管理するためのWebアプリケーションのバックエンドAPIです。
Go言語とPostgreSQL（PostGIS）を使用して構築されています。

## 主な機能

- ユーザー認証（JWT）
- トイレ位置情報の登録・検索（PostGIS使用）
- 位置情報ベースの検索
- ユーザー間でのトイレ情報の共有

## 技術スタック

- **言語**: Go 1.23
- **データベース**: PostgreSQL 15 + PostGIS 3.3
- **認証**: JWT (JSON Web Token)
- **コンテナ**: Docker & Docker Compose
- **マイグレーション**: golang-migrate

## クイックスタート

### 前提条件

- Go 1.23以上
- Docker & Docker Compose
- Git

### セットアップ

```bash
# 1. リポジトリをクローン
git clone <repository-url>
cd UnchingSpot_back-end

# 2. 環境変数を設定
cp .env.example .env

# 3. 依存関係をインストール
go mod download

# 4. マイグレーションツールをインストール
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# 5. データベースを起動
docker-compose up -d db

# 6. マイグレーションを実行
# Windows
scripts\migrate.bat up

# Linux/macOS
chmod +x scripts/migrate.sh
./scripts/migrate.sh up
```

## ドキュメント

- [マイグレーションガイド](migrations/README.md) - データベースマイグレーションの詳細

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

## 開発

### データベースマイグレーション

新しいマイグレーションを作成:
```bash
# Windows
scripts\migrate.bat create migration_name

# Linux/macOS
./scripts/migrate.sh create migration_name
```

マイグレーションを適用:
```bash
# Windows
scripts\migrate.bat up

# Linux/macOS
./scripts/migrate.sh up
```

### テスト実行

```bash
go test ./...
```

## ライセンス

[ライセンス情報を記載]

## コントリビューション

[コントリビューションガイドラインを記載]
