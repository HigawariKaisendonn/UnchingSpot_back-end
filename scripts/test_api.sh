#!/bin/bash

# うんちんぐすぽっと API テストスクリプト
# 使用方法: ./scripts/test_api.sh

API_BASE="http://localhost:8088"
EMAIL="test@example.com"
PASSWORD="password123"

echo "========================================="
echo "うんちんぐすぽっと API テスト"
echo "========================================="
echo ""

# 1. データベース接続テスト
echo "1. データベース接続テスト"
echo "---------------------------"
curl -s -X GET "$API_BASE/api/auth/test" | jq .
echo ""
echo ""

# 2. ユーザー登録（既に登録済みの場合はスキップ）
echo "2. ユーザー登録"
echo "---------------------------"
SIGNUP_RESPONSE=$(curl -s -X POST "$API_BASE/api/auth/signup" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"${EMAIL}\",\"password\":\"${PASSWORD}\",\"name\":\"テストユーザー\"}")
echo "$SIGNUP_RESPONSE" | jq .
echo ""

# 3. ログイン
echo "3. ログイン"
echo "---------------------------"
LOGIN_RESPONSE=$(curl -s -X POST "$API_BASE/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"${EMAIL}\",\"password\":\"${PASSWORD}\"}")
TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token')
echo "$LOGIN_RESPONSE" | jq .
echo ""
echo "取得したトークン: ${TOKEN:0:50}..."
echo ""

# 4. ユーザー情報取得
echo "4. ユーザー情報取得 (/api/auth/me)"
echo "---------------------------"
curl -s -X GET "$API_BASE/api/auth/me" \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""
echo ""

# 5. Pin作成
echo "5. Pin作成"
echo "---------------------------"
PIN_CREATE_RESPONSE=$(curl -s -X POST "$API_BASE/api/pins" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"テストトイレ","latitude":35.6895,"longitude":139.6917}')
PIN_ID=$(echo "$PIN_CREATE_RESPONSE" | jq -r '.id')
echo "$PIN_CREATE_RESPONSE" | jq .
echo ""
echo "作成したPin ID: $PIN_ID"
echo ""

# 6. Pin一覧取得
echo "6. Pin一覧取得"
echo "---------------------------"
curl -s -X GET "$API_BASE/api/pins" \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""
echo ""

# 7. Pin詳細取得
echo "7. Pin詳細取得"
echo "---------------------------"
curl -s -X GET "$API_BASE/api/pins/$PIN_ID" \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""
echo ""

# 8. Pin更新
echo "8. Pin更新"
echo "---------------------------"
curl -s -X PUT "$API_BASE/api/pins/$PIN_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"更新されたトイレ","latitude":35.6900,"longitude":139.6920}' | jq .
echo ""
echo ""

# 9. Connect作成（2つのPinが必要）
echo "9. Connect作成"
echo "---------------------------"
# 既存のPinを取得
PINS=$(curl -s -X GET "$API_BASE/api/pins" \
  -H "Authorization: Bearer $TOKEN")
PIN_IDS=($(echo "$PINS" | jq -r '.[].id'))
if [ ${#PIN_IDS[@]} -ge 2 ]; then
  PIN1_ID=${PIN_IDS[0]}
  PIN2_ID=${PIN_IDS[1]}
  CONNECT_CREATE_RESPONSE=$(curl -s -X POST "$API_BASE/api/connects" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "{\"pin_id_1\":\"$PIN1_ID\",\"pin_id_2\":\"$PIN2_ID\",\"show\":true}")
  CONNECT_ID=$(echo "$CONNECT_CREATE_RESPONSE" | jq -r '.id')
  echo "$CONNECT_CREATE_RESPONSE" | jq .
  echo ""
  echo "作成したConnect ID: $CONNECT_ID"
else
  echo "Connect作成には最低2つのPinが必要です。"
fi
echo ""
echo ""

# 10. Connect一覧取得
echo "10. Connect一覧取得"
echo "---------------------------"
curl -s -X GET "$API_BASE/api/connects" \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""
echo ""

# 11. ログアウト
echo "11. ログアウト"
echo "---------------------------"
curl -s -X POST "$API_BASE/api/auth/logout" \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""
echo ""

echo "========================================="
echo "テスト完了！"
echo "========================================="

