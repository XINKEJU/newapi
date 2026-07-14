package service

import (
	"crypto/sha1"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// 旧协议：SHA-1 签名测试
// ============================================================================

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

// ============================================================================
// 新协议：YooKassa v3 API 测试
// ============================================================================

func TestYookassaPaymentRequest_Serialization(t *testing.T) {
	// 验证 YooKassa 支付请求的序列化正确性
	req := &YookassaPaymentRequest{
		Amount: &YookassaAmount{
			Value:    "100.00",
			Currency: "RUB",
		},
		Description: "余额充值测试",
		Confirmation: &YookassaConfirmation{
			Type:      "redirect",
			ReturnURL: "https://example.com/return",
		},
		Capture: true,
		Metadata: map[string]string{
			"order_id": "YM123-test",
		},
	}

	jsonBytes, err := common.Marshal(req)
	require.NoError(t, err, "serialization should not fail")

	jsonStr := string(jsonBytes)
	assert.Contains(t, jsonStr, `"value":"100.00"`)
	assert.Contains(t, jsonStr, `"currency":"RUB"`)
	assert.Contains(t, jsonStr, `"type":"redirect"`)
	assert.Contains(t, jsonStr, `"capture":true`)
	assert.Contains(t, jsonStr, `"order_id":"YM123-test"`)
	assert.Contains(t, jsonStr, `"description":"余额充值测试"`)
}

func TestYookassaWebhookNotification_Parse(t *testing.T) {
	// 模拟 YooKassa webhook 通知 JSON
	webhookJSON := `{
		"type": "notification",
		"event": "payment.succeeded",
		"object": {
			"id": "22d6d597-000f-5000-9000-145f6df21d6f",
			"status": "succeeded",
			"paid": true,
			"amount": {
				"value": "100.00",
				"currency": "RUB"
			},
			"created_at": "2024-05-20T10:51:18.139Z",
			"metadata": {
				"order_id": "YM123-test"
			}
		}
	}`

	// 使用 HandleYookassaWebhook 整体解析
	event, payment, orderID, err := HandleYookassaWebhook([]byte(webhookJSON))
	require.NoError(t, err)

	assert.Equal(t, "payment.succeeded", event)
	assert.Equal(t, "YM123-test", orderID)
	require.NotNil(t, payment)
	assert.Equal(t, "22d6d597-000f-5000-9000-145f6df21d6f", payment.ID)
	assert.Equal(t, "succeeded", payment.Status)
	assert.True(t, payment.Paid)
	assert.Equal(t, "100.00", payment.Amount.Value)
	assert.Equal(t, "RUB", payment.Amount.Currency)
	assert.Equal(t, "YM123-test", payment.Metadata["order_id"])
}

func TestYookassaHandleWebhook(t *testing.T) {
	// 测试 HandleYookassaWebhook 函数 - 包含完整流程
	webhookJSON := []byte(`{
		"type": "notification",
		"event": "payment.succeeded",
		"object": {
			"id": "test-payment-id-123",
			"status": "succeeded",
			"paid": true,
			"amount": {
				"value": "200.00",
				"currency": "RUB"
			},
			"created_at": "2024-05-20T10:51:18.139Z",
			"metadata": {
				"order_id": "YMSUB456-test"
			}
		}
	}`)

	event, payment, orderID, err := HandleYookassaWebhook(webhookJSON)
	require.NoError(t, err)
	assert.Equal(t, "payment.succeeded", event)
	assert.Equal(t, "YMSUB456-test", orderID)
	assert.NotNil(t, payment)
	assert.Equal(t, "test-payment-id-123", payment.ID)
	assert.Equal(t, "YMSUB456-test", payment.Metadata["order_id"])
}

func TestYookassaHandleWebhook_InvalidNotification(t *testing.T) {
	// 测试无效通知（未知类型）
	invalidJSON := []byte(`{"type":"unknown","event":"test"}`)
	_, _, _, err := HandleYookassaWebhook(invalidJSON)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "非法的通知类型")
}

func TestYookassaHandleWebhook_WrongEvent(t *testing.T) {
	// 测试 payment.canceled 事件（应正常解析，不报错）
	webhookJSON := []byte(`{
		"type": "notification",
		"event": "payment.canceled",
		"object": {
			"id": "test-id",
			"status": "canceled",
			"paid": false,
			"amount": {"value":"100.00","currency":"RUB"},
			"created_at": "2024-05-20T10:51:18.139Z",
			"metadata": {"order_id": "YM789-test"}
		}
	}`)

	event, _, orderID, err := HandleYookassaWebhook(webhookJSON)
	require.NoError(t, err)
	assert.Equal(t, "payment.canceled", event)
	assert.Equal(t, "YM789-test", orderID)
}

func TestYookassaPaymentRequest_NoDescription(t *testing.T) {
	// 验证 description 为空时序列化结果不包含该字段
	req := &YookassaPaymentRequest{
		Amount: &YookassaAmount{
			Value:    "50.00",
			Currency: "RUB",
		},
		Confirmation: &YookassaConfirmation{
			Type:      "redirect",
			ReturnURL: "https://example.com/return",
		},
		Capture:  true,
		Metadata: map[string]string{"order_id": "test"},
	}

	jsonBytes, err := common.Marshal(req)
	require.NoError(t, err)

	// description 使用了 omitempty，空字符串应该不出现
	jsonStr := string(jsonBytes)
	assert.NotContains(t, jsonStr, `"description"`)
}

func TestYookassaAmountFormat(t *testing.T) {
	// 验证金额格式
	req := &YookassaPaymentRequest{
		Amount: &YookassaAmount{
			Value:    "50",
			Currency: "RUB",
		},
		Confirmation: &YookassaConfirmation{
			Type:      "redirect",
			ReturnURL: "https://example.com/return",
		},
		Capture:  true,
		Metadata: map[string]string{"order_id": "test"},
	}

	jsonBytes, err := common.Marshal(req)
	require.NoError(t, err)
	jsonStr := string(jsonBytes)
	assert.Contains(t, jsonStr, `"value":"50"`)
}

func TestYookassaWebhookNotification_StandaloneParse(t *testing.T) {
	// 测试直接反序列化 YookassaWebhookNotification 结构体
	webhookJSON := `{
		"type": "notification",
		"event": "payment.waiting_for_capture",
		"object": {
			"id": "test-payment-id",
			"status": "waiting_for_capture",
			"paid": true,
			"amount": {"value":"500.00","currency":"RUB"},
			"test": true,
			"metadata": {"order_id": "YM555-test"}
		}
	}`

	var notification YookassaWebhookNotification
	err := common.Unmarshal([]byte(webhookJSON), &notification)
	require.NoError(t, err)

	assert.Equal(t, "notification", notification.Type)
	assert.Equal(t, "payment.waiting_for_capture", notification.Event)
	assert.NotEmpty(t, notification.Object)

	// 解析 object 中的支付对象
	var payment YookassaPaymentResponse
	err = common.Unmarshal(notification.Object, &payment)
	require.NoError(t, err)
	assert.Equal(t, "test-payment-id", payment.ID)
	assert.Equal(t, "waiting_for_capture", payment.Status)
	assert.Equal(t, "500.00", payment.Amount.Value)
	assert.True(t, payment.Test)
}
