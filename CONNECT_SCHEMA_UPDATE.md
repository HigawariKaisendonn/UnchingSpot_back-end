# Connect テーブルスキーマ更新

## 変更概要

図形描画機能に対応するため、connectテーブルのスキーマを変更しました。

## 変更内容

### データベーススキーマ

**変更前:**
- `pins_id_2`: 単一のUUID（2点間の接続のみ）

**変更後:**
- `pins_id_2`: UUID配列（複数の中間点を保持可能）

### 変更されたファイル

1. **マイグレーション**
   - `migrations/000003_create_connect_table.up.sql`
     - `pins_id_2`をUUID配列型に変更
     - GINインデックスを追加（配列検索の高速化）

2. **モデル**
   - `internal/model/connect.go`
     - `PinID2`を`pq.StringArray`型に変更
   - `internal/model/request.go`
     - `CreateConnectRequest.PinID2`を`[]string`に変更
     - `UpdateConnectRequest.PinID2`を`[]string`に変更

3. **サービス層**
   - `internal/service/connect_service.go`
     - `CreateConnect`メソッドのシグネチャを変更
     - `UpdateConnect`メソッドのシグネチャを変更
     - 複数のpin_id_2の存在確認ロジックを追加

4. **ハンドラー層**
   - `internal/handler/connect_handler.go`
     - バリデーションロジックを配列対応に変更

5. **テスト**
   - `internal/handler/connect_handler_test.go`
     - すべてのテストケースを配列型に対応
   - `internal/database/testhelper.go`
     - `CreateTestConnect`メソッドを配列型に対応
   - `internal/database/testdb_test.go`
     - テストケースを配列型に対応

## 使用例

### Connect作成リクエスト

```json
{
  "pin_id_1": "開始点と終了点のUUID",
  "pin_id_2": [
    "中間点1のUUID",
    "中間点2のUUID",
    "中間点3のUUID"
  ],
  "show": true
}
```

### Connect更新リクエスト

```json
{
  "pin_id_1": "開始点と終了点のUUID",
  "pin_id_2": [
    "新しい中間点1のUUID",
    "新しい中間点2のUUID"
  ],
  "show": false
}
```

## マイグレーション手順

既存のデータベースを更新する場合：

1. データベースのバックアップを取得
2. マイグレーションを実行：
   ```bash
   # Windows
   scripts\migrate.bat down
   scripts\migrate.bat up

   # Linux/macOS
   ./scripts/migrate.sh down
   ./scripts/migrate.sh up
   ```

## 注意事項

- 既存のconnectデータは、マイグレーションのdown/upで削除されます
- 本番環境では、データ移行スクリプトが必要な場合があります
- `pins_id_2`は最低1つの要素が必要です（空配列は不可）
