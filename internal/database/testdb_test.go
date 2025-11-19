package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupTestDB(t *testing.T) {
	// テストデータベースのセットアップ
	testDB, err := SetupTestDB()
	require.NoError(t, err, "テストデータベースのセットアップに失敗しました")
	defer testDB.Teardown()

	// データベース接続の確認
	err = testDB.DB.Ping()
	assert.NoError(t, err, "データベースへのPingに失敗しました")

	// PostGIS拡張機能の確認
	var version string
	err = testDB.DB.Get(&version, "SELECT PostGIS_Version()")
	assert.NoError(t, err, "PostGISバージョンの取得に失敗しました")
	assert.NotEmpty(t, version, "PostGISバージョンが空です")
	t.Logf("PostGIS version: %s", version)
}

func TestCleanupData(t *testing.T) {
	// テストデータベースのセットアップ
	testDB, err := SetupTestDB()
	require.NoError(t, err)
	defer testDB.Teardown()

	helper := NewTestHelper(testDB)

	// テストユーザーの作成
	user, err := helper.CreateTestUser("test@example.com", "password123", "Test User")
	require.NoError(t, err)
	assert.NotEmpty(t, user.ID)

	// テストPinの作成
	pin, err := helper.CreateTestPin(user.ID, "Test Pin", 35.6895, 139.6917)
	require.NoError(t, err)
	assert.NotEmpty(t, pin.ID)

	// データのクリーンアップ
	err = testDB.CleanupData()
	assert.NoError(t, err, "データのクリーンアップに失敗しました")

	// データが削除されたことを確認
	var count int
	err = testDB.DB.Get(&count, "SELECT COUNT(*) FROM users")
	assert.NoError(t, err)
	assert.Equal(t, 0, count, "usersテーブルにデータが残っています")

	err = testDB.DB.Get(&count, "SELECT COUNT(*) FROM pins")
	assert.NoError(t, err)
	assert.Equal(t, 0, count, "pinsテーブルにデータが残っています")
}

func TestTestHelper_CreateTestUser(t *testing.T) {
	testDB, err := SetupTestDB()
	require.NoError(t, err)
	defer testDB.Teardown()

	helper := NewTestHelper(testDB)

	// ユーザーの作成
	user, err := helper.CreateTestUser("user@example.com", "securepass", "John Doe")
	require.NoError(t, err)
	assert.NotEmpty(t, user.ID)
	assert.Equal(t, "user@example.com", user.Email)
	assert.Equal(t, "John Doe", user.Name)
	assert.NotEqual(t, "securepass", user.Password, "パスワードがハッシュ化されていません")

	// データベースから取得して確認
	retrievedUser, err := helper.GetUserByID(user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.ID, retrievedUser.ID)
	assert.Equal(t, user.Email, retrievedUser.Email)
}

func TestTestHelper_CreateTestPin(t *testing.T) {
	testDB, err := SetupTestDB()
	require.NoError(t, err)
	defer testDB.Teardown()

	helper := NewTestHelper(testDB)

	// テストユーザーの作成
	user, err := helper.CreateTestUser("user@example.com", "password", "Test User")
	require.NoError(t, err)

	// Pinの作成
	pin, err := helper.CreateTestPin(user.ID, "Tokyo Station", 35.6812, 139.7671)
	require.NoError(t, err)
	assert.NotEmpty(t, pin.ID)
	assert.Equal(t, "Tokyo Station", pin.Name)
	assert.Equal(t, user.ID, pin.UserID)
	assert.Equal(t, 35.6812, pin.Latitude)
	assert.Equal(t, 139.7671, pin.Longitude)

	// データベースから取得して確認
	retrievedPin, err := helper.GetPinByID(pin.ID)
	require.NoError(t, err)
	assert.Equal(t, pin.ID, retrievedPin.ID)
	assert.Equal(t, pin.Name, retrievedPin.Name)
	assert.InDelta(t, pin.Latitude, retrievedPin.Latitude, 0.0001)
	assert.InDelta(t, pin.Longitude, retrievedPin.Longitude, 0.0001)
}

func TestTestHelper_CreateTestConnect(t *testing.T) {
	testDB, err := SetupTestDB()
	require.NoError(t, err)
	defer testDB.Teardown()

	helper := NewTestHelper(testDB)

	// テストユーザーの作成
	user, err := helper.CreateTestUser("user@example.com", "password", "Test User")
	require.NoError(t, err)

	// 2つのPinの作成
	pin1, err := helper.CreateTestPin(user.ID, "Pin 1", 35.6812, 139.7671)
	require.NoError(t, err)

	pin2, err := helper.CreateTestPin(user.ID, "Pin 2", 35.6895, 139.6917)
	require.NoError(t, err)

	// Connectの作成
	connect, err := helper.CreateTestConnect(user.ID, pin1.ID, pin2.ID, true)
	require.NoError(t, err)
	assert.NotEmpty(t, connect.ID)
	assert.Equal(t, user.ID, connect.UserID)
	assert.Equal(t, pin1.ID, connect.PinID1)
	assert.Equal(t, pin2.ID, connect.PinID2)
	assert.True(t, connect.Show)

	// データベースから取得して確認
	retrievedConnect, err := helper.GetConnectByID(connect.ID)
	require.NoError(t, err)
	assert.Equal(t, connect.ID, retrievedConnect.ID)
	assert.Equal(t, connect.UserID, retrievedConnect.UserID)
	assert.Equal(t, connect.PinID1, retrievedConnect.PinID1)
	assert.Equal(t, connect.PinID2, retrievedConnect.PinID2)
}
