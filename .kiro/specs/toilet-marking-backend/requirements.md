# 要件定義書

## イントロダクション

「うんちんぐすぽっと」は、ユーザーがトイレを使用した場所をマーキングし、複数の位置情報を接続して管理できるバックエンドシステムです。Go言語で実装され、PostgreSQL（PostGIS拡張）をデータベースとして使用し、Docker環境で動作し、Fly.ioにデプロイされます。

## 用語集

- **System**: うんちんぐすぽっとバックエンドAPI
- **User**: アプリケーションを使用する登録済みユーザー
- **Pin**: ユーザーが作成したトイレの位置情報マーカー
- **Connect**: 2つのPinを接続するエリア情報
- **Frontend**: Next.jsで実装されたクライアントアプリケーション
- **PostGIS**: PostgreSQLの地理空間データ拡張機能
- **Docker**: コンテナ化されたPostgreSQLデータベース環境

## 要件

### 要件1: ユーザー登録

**ユーザーストーリー:** ユーザーとして、メールアドレスとパスワードを使用してアカウントを作成したい。これにより、アプリケーションの機能にアクセスできるようになる。

#### 受入基準

1. WHEN Frontendが有効なメールアドレスとパスワードを含むPOSTリクエストを/api/auth/signupエンドポイントに送信するとき、THE System SHALL 新しいUserレコードをusersテーブルに作成する
2. WHEN Systemが新しいUserを作成するとき、THE System SHALL 一意のUUID形式のidを生成し割り当てる
3. WHEN Systemが新しいUserを作成するとき、THE System SHALL パスワードをハッシュ化してから保存する
4. IF 既に登録済みのメールアドレスでユーザー登録リクエストを受信したとき、THEN THE System SHALL エラーレスポンスを返し、重複登録を防止する
5. WHEN ユーザー登録が成功したとき、THE System SHALL 成功ステータスとユーザー情報を含むレスポンスをFrontendに返す

### 要件2: ユーザーログイン

**ユーザーストーリー:** 登録済みユーザーとして、メールアドレスとパスワードでログインしたい。これにより、自分のデータにアクセスし、機能を使用できるようになる。

#### 受入基準

1. WHEN Frontendがメールアドレスとパスワードを含むPOSTリクエストを/api/auth/loginエンドポイントに送信するとき、THE System SHALL 提供された認証情報を検証する
2. WHEN 認証情報が有効であるとき、THE System SHALL 認証トークンまたはセッションを生成する
3. WHEN ログインが成功したとき、THE System SHALL 認証トークンとユーザー情報を含むレスポンスをFrontendに返す
4. IF 無効な認証情報を受信したとき、THEN THE System SHALL 認証エラーレスポンスを返す

### 要件3: ユーザーログアウト

**ユーザーストーリー:** ログイン中のユーザーとして、セキュアにログアウトしたい。これにより、セッションを終了し、アカウントを保護できる。

#### 受入基準

1. WHEN 認証済みUserがPOSTリクエストを/api/auth/logoutエンドポイントに送信するとき、THE System SHALL 現在のセッションまたは認証トークンを無効化する
2. WHEN ログアウトが成功したとき、THE System SHALL 成功ステータスを含むレスポンスをFrontendに返す

### 要件4: ログイン中のユーザー情報取得

**ユーザーストーリー:** ログイン中のユーザーとして、自分のユーザー名を確認したい。これにより、正しいアカウントでログインしていることを確認できる。

#### 受入基準

1. WHEN 認証済みUserがGETリクエストを/api/auth/meエンドポイントに送信するとき、THE System SHALL 現在のUserのname情報を取得する
2. WHEN ユーザー情報の取得が成功したとき、THE System SHALL ユーザー名を含むレスポンスをFrontendに返す
3. IF 未認証のリクエストを受信したとき、THEN THE System SHALL 認証エラーレスポンスを返す

### 要件5: データベース接続テスト

**ユーザーストーリー:** 開発者として、PostgreSQLデータベースへの接続が正常に機能していることを確認したい。これにより、システムの健全性を検証できる。

#### 受入基準

1. WHEN GETリクエストを/api/auth/testエンドポイントに送信するとき、THE System SHALL PostgreSQLデータベースへの接続をテストする
2. WHEN 接続テストが成功したとき、THE System SHALL 成功ステータスを含むレスポンスを返す
3. IF データベース接続に失敗したとき、THEN THE System SHALL エラーステータスと詳細情報を含むレスポンスを返す

### 要件6: Pin作成

**ユーザーストーリー:** ログイン中のユーザーとして、トイレを使用した場所の位置情報を保存したい。これにより、訪れた場所を記録できる。

#### 受入基準

1. WHEN 認証済みUserがPin名と位置情報（緯度経度）を含むリクエストをFrontendから送信するとき、THE System SHALL 新しいPinレコードをpinsテーブルに作成する
2. WHEN Systemが新しいPinを作成するとき、THE System SHALL 一意のUUID形式のidを生成し割り当てる
3. WHEN Systemが新しいPinを作成するとき、THE System SHALL 位置情報をPostGISのPoint型として保存する
4. WHEN Systemが新しいPinを作成するとき、THE System SHALL 現在のUserのidをuser_idとして関連付ける
5. WHEN Pin作成が成功したとき、THE System SHALL 作成されたPin情報を含むレスポンスをFrontendに返す

### 要件7: Pin更新

**ユーザーストーリー:** ログイン中のユーザーとして、既存のPinの位置情報を更新したい。これにより、誤った情報を修正できる。

#### 受入基準

1. WHEN 認証済みUserが自分が所有するPinの更新リクエストをFrontendから送信するとき、THE System SHALL pinsテーブル内の対象Pinレコードを更新する
2. WHEN Systemが Pinを更新するとき、THE System SHALL 新しい位置情報をPostGISのPoint型として保存する
3. WHEN Systemが Pinを更新するとき、THE System SHALL edit_adフィールドを現在の日時で更新する
4. IF Userが所有していないPinの更新を試みたとき、THEN THE System SHALL 権限エラーレスポンスを返す
5. WHEN Pin更新が成功したとき、THE System SHALL 更新されたPin情報を含むレスポンスをFrontendに返す

### 要件8: Connect作成

**ユーザーストーリー:** ログイン中のユーザーとして、2つのPinを接続するエリアを作成したい。これにより、複数の場所の関係性を管理できる。

#### 受入基準

1. WHEN 認証済みUserが2つのPin IDを含むリクエストをFrontendから送信するとき、THE System SHALL 新しいConnectレコードをconnectテーブルに作成する
2. WHEN Systemが新しいConnectを作成するとき、THE System SHALL 一意のUUID形式のidを生成し割り当てる
3. WHEN Systemが新しいConnectを作成するとき、THE System SHALL 現在のUserのidをuser_idとして関連付ける
4. WHEN Systemが新しいConnectを作成するとき、THE System SHALL 提供された2つのPin IDをpins_id_1とpins_id_2として保存する
5. WHEN Systemが新しいConnectを作成するとき、THE System SHALL showフィールドをデフォルト値（true）で設定する
6. IF 指定されたPin IDが存在しないとき、THEN THE System SHALL エラーレスポンスを返す
7. WHEN Connect作成が成功したとき、THE System SHALL 作成されたConnect情報を含むレスポンスをFrontendに返す

### 要件9: Connect更新

**ユーザーストーリー:** ログイン中のユーザーとして、既存のConnectの接続情報を更新したい。これにより、エリアの定義を変更できる。

#### 受入基準

1. WHEN 認証済みUserが自分が所有するConnectの更新リクエストをFrontendから送信するとき、THE System SHALL connectテーブル内の対象Connectレコードを更新する
2. WHEN SystemがConnectを更新するとき、THE System SHALL 新しいPin IDをpins_id_1またはpins_id_2として保存する
3. WHEN SystemがConnectを更新するとき、THE System SHALL showフィールドの値を更新できる
4. IF Userが所有していないConnectの更新を試みたとき、THEN THE System SHALL 権限エラーレスポンスを返す
5. IF 指定されたPin IDが存在しないとき、THEN THE System SHALL エラーレスポンスを返す
6. WHEN Connect更新が成功したとき、THE System SHALL 更新されたConnect情報を含むレスポンスをFrontendに返す

### 要件10: データベーススキーマ

**ユーザーストーリー:** システム管理者として、適切なデータベーススキーマが実装されていることを確認したい。これにより、データの整合性とパフォーマンスが保証される。

#### 受入基準

1. THE System SHALL usersテーブルを以下のカラムで定義する: id (UUID, PRIMARY KEY), name (text), email (Email), password (text), created_at (timestamp), updated_at (timestamp), deleted_at (timestamp, nullable)
2. THE System SHALL pinsテーブルを以下のカラムで定義する: id (UUID, PRIMARY KEY), name (text), user_id (UUID, FOREIGN KEY), location (geometry Point), created_at (timestamp), edit_ad (timestamp), deleted_at (timestamp, nullable)
3. THE System SHALL connectテーブルを以下のカラムで定義する: id (UUID, PRIMARY KEY), user_id (UUID, FOREIGN KEY), pins_id_1 (UUID, FOREIGN KEY), pins_id_2 (UUID, FOREIGN KEY), show (boolean)
4. THE System SHALL pinsテーブルのuser_idカラムにusersテーブルへの外部キー制約を設定する
5. THE System SHALL connectテーブルのuser_idカラムにusersテーブルへの外部キー制約を設定する
6. THE System SHALL connectテーブルのpins_id_1とpins_id_2カラムにpinsテーブルへの外部キー制約を設定する
7. THE System SHALL PostGIS拡張機能を有効化し、地理空間データ型をサポートする
