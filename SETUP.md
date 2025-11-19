# うんちんぐすぽっと バックエンド環境構築手順

このドキュメントでは、Docker Composeを使用してバックエンドAPIサーバーをローカル環境で起動する手順を説明します。

## 前提条件

以下のソフトウェアがインストールされていること：
- Docker Desktop（macOS/Windows）または Docker Engine + Docker Compose（Linux）
- Git

## 1. リポジトリのクローン

```bash
git clone <repository-url>
cd UnchingSpot_back-end
```

## 2. 環境変数ファイルの設定

**環境変数の値は管理者に確認してください。**

`.env.example`をコピーして`.env`ファイルを作成します：

```bash
cp .env.example .env
```

`.env`ファイルを開いて、管理者から提供された値を設定してください。

必要な環境変数：
- `DATABASE_URL` - データベース接続URL
- `POSTGRES_USER` - PostgreSQLユーザー名
- `POSTGRES_PASSWORD` - PostgreSQLパスワード
- `POSTGRES_DB` - データベース名
- `POSTGRES_PORT` - PostgreSQLポート番号
- `JWT_SECRET` - JWT認証用のシークレットキー
- `PORT` - APIサーバーのポート番号
- `FRONTEND_URL` - フロントエンドのURL（CORS設定用）
- `ENV` - 環境（development/production）
- `TEST_DATABASE_URL` - テスト用データベース接続URL

## 3. Dockerコンテナの起動

Docker ComposeでデータベースとAPIサーバーを起動します：

```bash
docker-compose up -d
```

初回起動時は、Dockerイメージのビルドに数分かかる場合があります。

起動が完了したら、コンテナの状態を確認します：

```bash
docker-compose ps
```

以下の2つのコンテナが表示されればOKです：
- `unchingspot_back-end-db-1` (PostgreSQL + PostGIS) - STATUS: `Up (healthy)`
- `unchingspot_back-end-api-1` (APIサーバー) - STATUS: `Up`

APIログを確認して、データベース接続の状態をチェック：

```bash
docker-compose logs api
```

**初回起動時は以下のエラーが出ますが、これは正常です（テーブルがまだ作成されていないため）：**
```
Database connection failed: database connection test failed: failed to find user by email: pq: relation "users" does not exist
```

次のステップでテーブルを作成します。

## 4. データベースマイグレーションの実行

データベースにテーブルを作成します。以下のコマンドを**順番に**実行してください：

### 4-1. PostGIS拡張機能を有効化

```bash
docker-compose exec db psql -U postgres -d unchingspot -c "CREATE EXTENSION IF NOT EXISTS postgis;"
```

成功すると `CREATE EXTENSION` と表示されます。

### 4-2. usersテーブルの作成

```bash
cat migrations/000001_create_users_table.up.sql | docker-compose exec -T db psql -U postgres -d unchingspot
```

以下のような出力が表示されればOK：
```
CREATE EXTENSION
CREATE TABLE
CREATE INDEX
CREATE INDEX
```

### 4-3. pinsテーブルの作成

```bash
cat migrations/000002_create_pins_table.up.sql | docker-compose exec -T db psql -U postgres -d unchingspot
```

以下のような出力が表示されればOK：
```
CREATE TABLE
CREATE INDEX
CREATE INDEX
CREATE INDEX
```

### 4-4. connectテーブルの作成

```bash
cat migrations/000003_create_connect_table.up.sql | docker-compose exec -T db psql -U postgres -d unchingspot
```

以下のような出力が表示されればOK：
```
CREATE TABLE
CREATE INDEX
CREATE INDEX
```

### 4-5. テーブル作成の確認

データベースに接続して、テーブルが正しく作成されたか確認できます：

```bash
docker-compose exec db psql -U postgres -d unchingspot -c "\dt"
```

以下の3つのテーブルが表示されればOK：
```
           List of relations
 Schema |  Name   | Type  |  Owner
--------+---------+-------+----------
 public | connect | table | postgres
 public | pins    | table | postgres
 public | users   | table | postgres
```

## 5. 動作確認

APIサーバーが正常に起動しているか、テストエンドポイントを呼び出して確認します：

```bash
curl http://localhost:8088/api/auth/test
```

**成功した場合：**
```json
{
  "message": "Database connection successful",
  "status": "ok"
}
```

このレスポンスが返ってくれば、環境構築は完了です！🎉

**エラーが出た場合：**

APIログを確認してエラー内容を調べます：
```bash
docker-compose logs api --tail=50
```

よくあるエラーと対処法は「トラブルシューティング」セクションを参照してください。

## 利用可能なエンドポイント

サーバーが起動したら、以下のエンドポイントが利用可能になります：

### 認証API
- `POST /api/auth/signup` - ユーザー登録
- `POST /api/auth/login` - ログイン
- `POST /api/auth/logout` - ログアウト（認証必要）
- `GET /api/auth/me` - ユーザー情報取得（認証必要）
- `GET /api/auth/test` - DB接続テスト

### Pin API（全て認証必要）
- `POST /api/pins` - Pin作成
- `GET /api/pins` - Pin一覧取得
- `GET /api/pins/:id` - Pin取得
- `PUT /api/pins/:id` - Pin更新
- `DELETE /api/pins/:id` - Pin削除

### Connect API（全て認証必要）
- `POST /api/connects` - Connect作成
- `GET /api/connects` - Connect一覧取得
- `PUT /api/connects/:id` - Connect更新
- `DELETE /api/connects/:id` - Connect削除

## よく使うコマンド

### コンテナの起動
```bash
docker-compose up -d
```

### コンテナの停止
```bash
docker-compose down
```

### コンテナの再起動
```bash
docker-compose restart
```

### ログの確認
```bash
# 全てのログを表示
docker-compose logs

# APIサーバーのログのみ表示（最新50行）
docker-compose logs api --tail=50

# データベースのログのみ表示
docker-compose logs db --tail=50

# リアルタイムでログを表示
docker-compose logs -f api
```

### データベースに直接接続
```bash
docker-compose exec db psql -U postgres -d unchingspot
```

### データベースのリセット（全データ削除）
```bash
# コンテナとボリュームを削除
docker-compose down -v

# 再起動してマイグレーション実行
docker-compose up -d
# （マイグレーションコマンドを再実行）
```

## トラブルシューティング

### ポートが既に使用されている

エラー：`Bind for 0.0.0.0:8088 failed: port is already allocated`

別のプロセスがポート8088または5432を使用しています。以下で確認：

```bash
# macOS/Linux
lsof -i :8088
lsof -i :5432

# 使用中のプロセスを停止するか、.envファイルでポート番号を変更
```

### データベース接続エラー

**エラー例：**
```
Failed to initialize database: failed to connect to database: pq: password authentication failed for user "postgres"
```

または

```
parse "postgres://<username>:<strong_password>@db:5432/unchingspot?sslmode=disable": net/url: invalid userinfo
```

**原因：**
`.env`ファイルの環境変数がプレースホルダーのままになっているか、古いボリュームデータが残っている

**対処法：**

1. `.env`ファイルを確認して、`<username>`や`<password>`などのプレースホルダーが残っていないかチェック：
   ```bash
   cat .env
   ```

2. `.env`を修正したら、古いボリュームデータを削除して再起動：
   ```bash
   docker-compose down -v
   docker-compose up -d
   ```

   **注意：`-v`オプションを付けるとデータベースの全データが削除されます**

3. マイグレーションを再実行（手順4に戻る）

### マイグレーションエラー

`relation already exists`エラーが出る場合は既にテーブルが作成されています。問題ありません。

テーブルを再作成したい場合：
```bash
# downマイグレーションでテーブル削除
cat migrations/000003_create_connect_table.down.sql | docker-compose exec -T db psql -U postgres -d unchingspot
cat migrations/000002_create_pins_table.down.sql | docker-compose exec -T db psql -U postgres -d unchingspot
cat migrations/000001_create_users_table.down.sql | docker-compose exec -T db psql -U postgres -d unchingspot

# upマイグレーションで再作成
cat migrations/000001_create_users_table.up.sql | docker-compose exec -T db psql -U postgres -d unchingspot
cat migrations/000002_create_pins_table.up.sql | docker-compose exec -T db psql -U postgres -d unchingspot
cat migrations/000003_create_connect_table.up.sql | docker-compose exec -T db psql -U postgres -d unchingspot
```

## 開発時のヒント

### ホットリロード

現在の設定では、コードを変更してもコンテナは自動的に再ビルドされません。変更を反映するには：

```bash
docker-compose restart api
```

または、コンテナを再ビルド：

```bash
docker-compose up -d --build api
```

### ローカルでGoを直接実行

Dockerを使わずにローカルでAPIサーバーを実行する場合：

```bash
# データベースコンテナのみ起動
docker-compose up -d db

# ローカルでAPIサーバー起動
go run cmd/api/main.go
```

この場合、`.env`の`DATABASE_URL`は`localhost`を使用してください。

## 次のステップ

- API仕様の詳細は`API.md`を参照
- テストの実行方法は`TEST.md`を参照
- 開発ガイドラインは`CONTRIBUTING.md`を参照
