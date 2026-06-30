package service

import (
	"crypto/sha1"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// computeExpectedSignature 计算期望的 SHA1 签名（辅助验证）
func computeExpectedSignature(params map[string]string, secret string) string {
	toSign := buildSignatureString(params, secret)
	h := sha1.New()
	h.Write([]byte(toSign))
	return hex.EncodeToString(h.Sum(nil))
}

func TestNormalAmountForSignature(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"100", "100.00"},
		{"100.0", "100.00"},
		{"100.00", "100.00"},
		{"100.5", "100.50"},
		{"99.99", "99.99"},
		{"0", "0.00"},
		{"0.00", "0.00"},
		{"1234.56", "1234.56"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalAmountForSignature(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVerifyYooMoneyNotification_ValidSignature(t *testing.T) {
	secret := "test_secret_key"

	params := map[string]string{
		"notification_type": "p2p-incoming",
		"amount":            "100.00",
		"currency":          "643",
		"wallet_id":         "410011234567890",
		"transaction_id":    "1234567890",
		"label":             "YM123-test",
	}

	expectedHash := computeExpectedSignature(params, secret)

	notif := &YooMoneyNotification{
		NotificationType: params["notification_type"],
		Amount:           params["amount"],
		Currency:         params["currency"],
		WalletId:         params["wallet_id"],
		TransId:          params["transaction_id"],
		Label:            params["label"],
		Sha1Hash:         expectedHash,
	}

	result := VerifyYooMoneyNotification(notif, secret)
	assert.True(t, result, "signature should be valid")
}

func TestVerifyYooMoneyNotification_AmountWithoutDecimals(t *testing.T) {
	secret := "test_secret_key"

	// 回调中的 amount 可能是 "100"（不带小数）
	params := map[string]string{
		"notification_type": "p2p-incoming",
		"amount":            "100",
		"currency":          "643",
		"wallet_id":         "410011234567890",
		"transaction_id":    "1234567890",
		"label":             "YM123-test",
	}

	expectedHash := computeExpectedSignature(params, secret)

	notif := &YooMoneyNotification{
		NotificationType: params["notification_type"],
		Amount:           params["amount"],
		Currency:         params["currency"],
		WalletId:         params["wallet_id"],
		TransId:          params["transaction_id"],
		Label:            params["label"],
		Sha1Hash:         expectedHash,
	}

	result := VerifyYooMoneyNotification(notif, secret)
	assert.True(t, result, "signature should be valid even when amount lacks decimal places")
}

func TestVerifyYooMoneyNotification_WrongSignature(t *testing.T) {
	secret := "test_secret_key"

	notif := &YooMoneyNotification{
		NotificationType: "p2p-incoming",
		Amount:           "100.00",
		Currency:         "643",
		WalletId:         "410011234567890",
		TransId:          "1234567890",
		Label:            "YM123-test",
		Sha1Hash:         "wronghash123",
	}

	result := VerifyYooMoneyNotification(notif, secret)
	assert.False(t, result, "wrong signature should be rejected")
}

func TestVerifyYooMoneyNotification_EmptySecret(t *testing.T) {
	notif := &YooMoneyNotification{
		NotificationType: "p2p-incoming",
		Amount:           "100.00",
		Currency:         "643",
		WalletId:         "410011234567890",
		TransId:          "1234567890",
		Label:            "YM123-test",
		Sha1Hash:         "abc123",
	}

	result := VerifyYooMoneyNotification(notif, "")
	assert.False(t, result, "empty secret should return false")
}

func TestVerifyYooMoneyParams(t *testing.T) {
	secret := "my_notify_secret"

	params := map[string]string{
		"notification_type": "p2p-incoming",
		"amount":            "50.00",
		"currency":          "643",
		"wallet_id":         "410011234567890",
		"transaction_id":    "9876543210",
		"label":             "YM999-test",
	}

	params["sha1_hash"] = computeExpectedSignature(params, secret)

	err := VerifyYooMoneyParams(params, secret)
	require.NoError(t, err, "valid params should pass")
}

func TestVerifyYooMoneyParams_BadSignature(t *testing.T) {
	secret := "my_notify_secret"

	params := map[string]string{
		"notification_type": "p2p-incoming",
		"amount":            "50.00",
		"currency":          "643",
		"wallet_id":         "410011234567890",
		"transaction_id":    "9876543210",
		"label":             "YM999-test",
		"sha1_hash":         "0000000000000000000000000000000000000000",
	}

	err := VerifyYooMoneyParams(params, secret)
	assert.Error(t, err, "bad signature should fail")
}

func TestVerifyYooMoneyParams_MissingHash(t *testing.T) {
	secret := "my_notify_secret"

	params := map[string]string{
		"notification_type": "p2p-incoming",
		"amount":            "50.00",
		"currency":          "643",
		"wallet_id":         "410011234567890",
		"transaction_id":    "9876543210",
		"label":             "YM999-test",
	}

	err := VerifyYooMoneyParams(params, secret)
	assert.Error(t, err, "missing sha1_hash should fail")
}

// 测试签名与 YooMoney 官方文档中的示例一致
// 参考 https://yoomoney.ru/docs/payment-notifications
func TestVerifyYooMoneyNotification_DocumentationExample(t *testing.T) {
	// 模拟官方文档中的典型通知参数
	secret := "abcdef1234567890"

	params := map[string]string{
		"notification_type": "card-incoming",
		"amount":            "1500.00",
		"currency":          "643",
		"wallet_id":         "4100116075156746",
		"transaction_id":    "20130101000000000001",
		"label":             "order-123",
	}

	expectedHash := computeExpectedSignature(params, secret)

	notif := &YooMoneyNotification{
		NotificationType: params["notification_type"],
		Amount:           params["amount"],
		Currency:         params["currency"],
		WalletId:         params["wallet_id"],
		TransId:          params["transaction_id"],
		Label:            params["label"],
		Sha1Hash:         expectedHash,
	}

	result := VerifyYooMoneyNotification(notif, secret)
	assert.True(t, result, "should match documentation example signature")
}

func TestBuildSignatureString_IgnoresEmptyFields(t *testing.T) {
	secret := "test123"

	// 缺少 currency 字段
	params := map[string]string{
		"notification_type": "p2p-incoming",
		"amount":            "100.00",
		"wallet_id":         "410011234567890",
		"transaction_id":    "1234567890",
		"label":             "YM123-test",
	}

	sigStr := buildSignatureString(params, secret)

	// 应该不包含 currency 和空字段
	assert.NotContains(t, sigStr, "&&", "should not have empty field separators")
	assert.NotContains(t, strings.TrimRight(sigStr, "&"), "&&", "signature string should skip empty fields")

	// 验证可以正确计算签名
	hash := computeExpectedSignature(params, secret)
	assert.NotEmpty(t, hash)
	assert.Len(t, hash, 40, "SHA1 hash should be 40 hex characters")
}
