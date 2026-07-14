package service

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting"
)

// ============================================================================
// 旧版 YooMoney Quick Pay 协议（向后兼容）
// ============================================================================

// YooMoneyQuickPayParams 封装 YooMoney 快速支付表单参数
type YooMoneyQuickPayParams struct {
	Receiver     string  // 收款钱包 ID（wallet_id）
	QuickpayForm string  // 表单类型，固定 "shop"
	Targets      string  // 支付描述（显示在支付页）
	PaymentType  string  // 支付方式：PC=钱包、AC=银行卡、MC=手机余额
	Sum          float64 // 金额
	Label        string  // 订单号（trade_no，用于回调识别）
	SuccessURL   string  // 支付成功回跳地址
	FailURL      string  // 支付失败回跳地址
}

// BuildYooMoneyQuickPayURL 构造 YooMoney 快速支付重定向 URL（GET 跳转）
// 文档：https://yoomoney.ru/docs/payment-buttons/using/epay/payments-list-html
func BuildYooMoneyQuickPayURL(p *YooMoneyQuickPayParams) string {
	formData := url.Values{}
	formData.Set("receiver", p.Receiver)
	formData.Set("quickpay-form", p.QuickpayForm)
	formData.Set("targets", p.Targets)
	formData.Set("paymentType", p.PaymentType)
	formData.Set("sum", strconv.FormatFloat(p.Sum, 'f', 2, 64))
	formData.Set("label", p.Label)
	if p.SuccessURL != "" {
		formData.Set("successURL", p.SuccessURL)
	}
	if p.FailURL != "" {
		formData.Set("failURL", p.FailURL)
	}

	return "https://yoomoney.ru/quickpay/confirm.xml?" + formData.Encode()
}

// YooMoneyNotification YooMoney 异步通知（HTTP 回调）参数
// 基于 YooMoney HTTP 通知协议（旧版 YooKassa 快速支付表单回调）
// 签名依据的是回调中实际包含的字段，缺失字段不参加签名
type YooMoneyNotification struct {
	NotificationType string `form:"notification_type" json:"notification_type"`
	WalletId         string `form:"wallet_id"         json:"wallet_id"`
	Amount           string `form:"amount"            json:"amount"`
	Currency         string `form:"currency"          json:"currency"`
	TransId          string `form:"transaction_id"    json:"transaction_id"`
	Label            string `form:"label"             json:"label"`
	Sha1Hash         string `form:"sha1_hash"         json:"sha1_hash"`
	Unaccepted       string `form:"unaccepted"        json:"unaccepted"`
	CreatedAt        string `form:"created_at"        json:"created_at"`
	WithdrawId       string `form:"withdraw_id"       json:"withdraw_id"`
	Datestamp        string `form:"datestamp"         json:"datestamp"`
	SenderPhone      string `form:"sender_phone"      json:"sender_phone"`
	Codepro          string `form:"codepro"           json:"codepro"`
	OperationLabel   string `form:"operation_label"   json:"operation_label"`
	OperationId      string `form:"operation_id"      json:"operation_id"`
	BillId           string `form:"bill_id"           json:"bill_id"`
	Firstname        string `form:"firstname"         json:"firstname"`
	Lastname         string `form:"lastname"          json:"lastname"`
	Fathersname      string `form:"fathersname"       json:"fathersname"`
	Email            string `form:"email"             json:"email"`
	Phone            string `form:"phone"             json:"phone"`
	City             string `form:"city"              json:"city"`
	Street           string `form:"street"            json:"street"`
	Building         string `form:"building"          json:"building"`
	Suite            string `form:"suite"             json:"suite"`
	Flat             string `form:"flat"              json:"flat"`
	Zip              string `form:"zip"               json:"zip"`
	TestNotification string `form:"test_notification" json:"test_notification"`
}

// normalAmountForSignature 将金额标准化为两位小数格式（文档要求 amount 在签名中必须为 X.XX 格式）
func normalAmountForSignature(amount string) string {
	if amount == "" {
		return ""
	}
	val, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return amount
	}
	return strconv.FormatFloat(val, 'f', 2, 64)
}

// buildSignatureString 按文档顺序拼接非空参数并连接 notification_secret 生成签名源字符串
// 参数字段顺序：notification_type, amount, currency, wallet_id, transaction_id, label, <secret>
func buildSignatureString(params map[string]string, notificationSecret string) string {
	parts := []string{}

	// 按文档固定顺序遍历字段
	for _, key := range []string{"notification_type", "amount", "currency", "wallet_id", "transaction_id", "label"} {
		val := params[key]
		if key == "amount" && val != "" {
			val = normalAmountForSignature(val)
		}
		if val != "" {
			parts = append(parts, val)
		}
	}

	parts = append(parts, notificationSecret)
	return strings.Join(parts, "&")
}

// VerifyYooMoneyNotification 验证 YooMoney 通知签名
// 签名算法：SHA1(notification_type&amount&currency&wallet_id&transaction_id&label&<notification_secret>)
func VerifyYooMoneyNotification(n *YooMoneyNotification, notificationSecret string) bool {
	if notificationSecret == "" {
		common.SysError("YooMoney notification secret not configured")
		return false
	}

	params := map[string]string{
		"notification_type": n.NotificationType,
		"amount":            n.Amount,
		"currency":          n.Currency,
		"wallet_id":         n.WalletId,
		"transaction_id":    n.TransId,
		"label":             n.Label,
	}

	toSign := buildSignatureString(params, notificationSecret)

	h := sha1.New()
	h.Write([]byte(toSign))
	expected := hex.EncodeToString(h.Sum(nil))

	return expected == strings.ToLower(n.Sha1Hash)
}

// VerifyYooMoneyParams 从原始参数 map 验证 YooMoney 通知签名（适用于浏览器回跳 GET 参数）
func VerifyYooMoneyParams(params map[string]string, notificationSecret string) error {
	if notificationSecret == "" {
		return fmt.Errorf("notification secret not configured")
	}

	if params["sha1_hash"] == "" {
		return fmt.Errorf("sha1_hash missing")
	}

	toSign := buildSignatureString(params, notificationSecret)

	h := sha1.New()
	h.Write([]byte(toSign))
	expected := hex.EncodeToString(h.Sum(nil))

	if expected != strings.ToLower(params["sha1_hash"]) {
		return fmt.Errorf("signature mismatch")
	}
	return nil
}

// YooMoneyPaymentType 支付方式常量（旧 Quick Pay 协议）
const (
	YooMoneyPaymentTypeWallet = "PC" // YooMoney 钱包
	YooMoneyPaymentTypeCard   = "AC" // 银行卡（俄罗斯的 Visa/Mastercard 已不可用，仅 Mir）
	YooMoneyPaymentTypeMobile = "MC" // 手机余额
	YooMoneyPaymentTypeSber   = "SB" // SberPay（通过 Sberbank Online）
)

// CreateYooMoneySubscriptionOrder 为 YooMoney 创建订阅订单（旧 Quick Pay 协议）
// 返回支付重定向 URL
func CreateYooMoneySubscriptionOrder(tradeNo string, amount float64, description string, returnURL string, notifyURL string) (payURL string, err error) {
	cfgWalletId := setting.YoomoneyWalletId
	if cfgWalletId == "" {
		return "", fmt.Errorf("YooMoney wallet_id 未配置")
	}

	paymentType := YooMoneyPaymentTypeWallet // 默认使用 YooMoney 钱包支付
	if setting.YoomoneyTestMode {
		// 测试模式也使用钱包支付
	}

	p := &YooMoneyQuickPayParams{
		Receiver:     cfgWalletId,
		QuickpayForm: "shop",
		Targets:      description,
		PaymentType:  paymentType,
		Sum:          amount,
		Label:        tradeNo,
		SuccessURL:   returnURL,
		FailURL:      returnURL + "?pay=fail",
	}

	payURL = BuildYooMoneyQuickPayURL(p)
	return payURL, nil
}

// CreateYooMoneyTopUpOrder 为 YooMoney 创建充值订单（旧 Quick Pay 协议）
// 返回支付重定向 URL
func CreateYooMoneyTopUpOrder(tradeNo string, amount float64, returnURL string) (payURL string, err error) {
	cfgWalletId := setting.YoomoneyWalletId
	if cfgWalletId == "" {
		return "", fmt.Errorf("YooMoney wallet_id 未配置")
	}

	p := &YooMoneyQuickPayParams{
		Receiver:     cfgWalletId,
		QuickpayForm: "shop",
		Targets:      fmt.Sprintf("余额充值 #%s", tradeNo),
		PaymentType:  YooMoneyPaymentTypeWallet,
		Sum:          amount,
		Label:        tradeNo,
		SuccessURL:   returnURL,
		FailURL:      returnURL + "?pay=fail",
	}

	payURL = BuildYooMoneyQuickPayURL(p)
	return payURL, nil
}

// ============================================================================
// YooKassa v3 REST API（新协议）
// ============================================================================

// YookassaAmount 金额结构
type YookassaAmount struct {
	Value    string `json:"value"`    // 金额，点号分隔字符串，如 "100.00"
	Currency string `json:"currency"` // 货币代码 ISO-4217，如 "RUB"
}

// YookassaConfirmation 确认场景
type YookassaConfirmation struct {
	Type      string `json:"type"`       // 固定 "redirect"
	ReturnURL string `json:"return_url"` // 支付后回跳地址
}

// YookassaPaymentRequest 创建支付请求体
type YookassaPaymentRequest struct {
	Amount        *YookassaAmount      `json:"amount"`
	Description   string               `json:"description,omitempty"`
	Confirmation  *YookassaConfirmation `json:"confirmation"`
	Capture       bool                 `json:"capture"`
	Metadata      map[string]string    `json:"metadata,omitempty"`
	PaymentMethodData map[string]string `json:"payment_method_data,omitempty"`
}

// YookassaPaymentResponse 创建支付响应
type YookassaPaymentResponse struct {
	ID           string              `json:"id"`
	Status       string              `json:"status"`
	Paid         bool                `json:"paid"`
	Amount       *YookassaAmount     `json:"amount"`
	Confirmation *YookassaConfirmationResponse `json:"confirmation,omitempty"`
	CreatedAt    string              `json:"created_at"`
	Metadata     map[string]string   `json:"metadata,omitempty"`
	Test         bool                `json:"test"`
}

// YookassaConfirmationResponse 确认场景响应（含 confirmation_url）
type YookassaConfirmationResponse struct {
	Type            string `json:"type"`
	ConfirmationURL string `json:"confirmation_url"`
	ReturnURL       string `json:"return_url,omitempty"`
}

// YookassaWebhookNotification YooKassa webhook 通知结构
type YookassaWebhookNotification struct {
	Type   string          `json:"type"`   // 固定 "notification"
	Event  string          `json:"event"`  // 如 "payment.succeeded"
	Object json.RawMessage `json:"object"` // 支付对象 JSON
}

// YookassaError YooKassa API 错误响应
type YookassaError struct {
	Type        string `json:"type"`
	ID          string `json:"id"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

func (e *YookassaError) Error() string {
	return fmt.Sprintf("YooKassa API error: [%s] %s (id=%s)", e.Code, e.Description, e.ID)
}

// YookassaKnownIPs YooKassa webhook 来源 IP 白名单
// https://yookassa.ru/developers/using-api/webhooks
var YookassaKnownIPs = []string{
	"185.71.76.0/27",
	"185.71.77.0/27",
	"77.75.153.0/25",
	"77.75.156.11",
	"77.75.156.35",
	"77.75.154.128/25",
	"2a02:5180::/32",
}

// createYookassaHTTPClient 创建带有 Basic Auth 的 HTTP 客户端
func createYookassaHTTPClient() *http.Client {
	return &http.Client{}
}

// createYookassaRequest 创建一个 YooKassa API 请求
func createYookassaRequest(method, path string, body io.Reader) (*http.Request, error) {
	apiEndpoint := setting.GetYookassaAPIEndpoint()
	req, err := http.NewRequest(method, apiEndpoint+path, body)
	if err != nil {
		return nil, fmt.Errorf("创建 YooKassa 请求失败: %w", err)
	}

	// Basic Auth: shopId:secretKey
	req.SetBasicAuth(setting.YoomoneyShopId, setting.YoomoneySecretKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Idempotence-Key 由调用方设置

	return req, nil
}

// CreateYookassaPayment 通过 YooKassa v3 API 创建支付
// 返回确认 URL（confirmation_url）和支付 ID
func CreateYookassaPayment(tradeNo string, amountRub float64, description string, returnURL string) (confirmationURL string, paymentID string, err error) {
	shopID := setting.YoomoneyShopId
	secretKey := setting.YoomoneySecretKey
	if shopID == "" || secretKey == "" {
		return "", "", fmt.Errorf("YooKassa shop_id 或 secret_key 未配置")
	}

	// 金额格式化为两位小数
	amountStr := strconv.FormatFloat(amountRub, 'f', 2, 64)

	reqBody := &YookassaPaymentRequest{
		Amount: &YookassaAmount{
			Value:    amountStr,
			Currency: setting.YoomoneyCurrency,
		},
		Description: description,
		Confirmation: &YookassaConfirmation{
			Type:      "redirect",
			ReturnURL: returnURL,
		},
		Capture: true, // 单阶段模式，立即扣款
		Metadata: map[string]string{
			"order_id": tradeNo,
		},
	}

	bodyBytes, err := common.Marshal(reqBody)
	if err != nil {
		return "", "", fmt.Errorf("序列化请求体失败: %w", err)
	}

	req, err := createYookassaRequest("POST", "/payments", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", "", err
	}

	// 设置幂等性 Key（使用 tradeNo 作为幂等 Key，确保同一订单不会重复创建）
	req.Header.Set("Idempotence-Key", tradeNo)

	resp, err := createYookassaHTTPClient().Do(req)
	if err != nil {
		return "", "", fmt.Errorf("YooKassa API 请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("读取 YooKassa 响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		// 解析错误响应
		var apiErr YookassaError
		if common.Unmarshal(respBody, &apiErr) == nil && apiErr.Code != "" {
			return "", "", &apiErr
		}
		return "", "", fmt.Errorf("YooKassa API 返回 HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var paymentResp YookassaPaymentResponse
	if err := common.Unmarshal(respBody, &paymentResp); err != nil {
		return "", "", fmt.Errorf("解析 YooKassa 响应失败: %w", err)
	}

	if paymentResp.Confirmation == nil || paymentResp.Confirmation.ConfirmationURL == "" {
		return "", "", fmt.Errorf("YooKassa 响应缺少 confirmation_url")
	}

	common.SysLog(fmt.Sprintf("YooKassa payment created: id=%s, status=%s, tradeNo=%s", paymentResp.ID, paymentResp.Status, tradeNo))
	return paymentResp.Confirmation.ConfirmationURL, paymentResp.ID, nil
}

// HandleYookassaWebhook 验证并处理 YooKassa webhook 通知
// 返回 event 类型、支付对象、order_id（即 tradeNo）
func HandleYookassaWebhook(body []byte) (event string, payment *YookassaPaymentResponse, orderID string, err error) {
	var notification YookassaWebhookNotification
	if err := common.Unmarshal(body, &notification); err != nil {
		return "", nil, "", fmt.Errorf("解析 webhook 通知失败: %w", err)
	}

	if notification.Type != "notification" {
		return "", nil, "", fmt.Errorf("非法的通知类型: %s", notification.Type)
	}

	var paymentObj YookassaPaymentResponse
	if err := common.Unmarshal(notification.Object, &paymentObj); err != nil {
		return "", nil, "", fmt.Errorf("解析支付对象失败: %w", err)
	}

	// 从 metadata 中提取 order_id
	orderID = ""
	if paymentObj.Metadata != nil {
		orderID = paymentObj.Metadata["order_id"]
	}

	common.SysLog(fmt.Sprintf("YooKassa webhook received: event=%s, payment_id=%s, status=%s, order_id=%s",
		notification.Event, paymentObj.ID, paymentObj.Status, orderID))

	return notification.Event, &paymentObj, orderID, nil
}

// QueryYookassaPayment 查询 YooKassa 支付状态
// 用于 webhook 验证（对象状态认证）
func QueryYookassaPayment(paymentID string) (*YookassaPaymentResponse, error) {
	req, err := createYookassaRequest("GET", "/payments/"+url.PathEscape(paymentID), nil)
	if err != nil {
		return nil, err
	}

	resp, err := createYookassaHTTPClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("查询 YooKassa 支付状态失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取 YooKassa 响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr YookassaError
		if common.Unmarshal(respBody, &apiErr) == nil && apiErr.Code != "" {
			return nil, &apiErr
		}
		return nil, fmt.Errorf("YooKassa API 返回 HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var paymentResp YookassaPaymentResponse
	if err := common.Unmarshal(respBody, &paymentResp); err != nil {
		return nil, fmt.Errorf("解析 YooKassa 响应失败: %w", err)
	}

	return &paymentResp, nil
}

// YookassaRefundRequest 退款请求体
type YookassaRefundRequest struct {
	Amount    *YookassaAmount `json:"amount"`
	PaymentID string          `json:"payment_id"`
}

// CreateYookassaRefund 发起 YooKassa 退款
func CreateYookassaRefund(paymentID string, amountRub float64) error {
	amountStr := strconv.FormatFloat(amountRub, 'f', 2, 64)

	reqBody := &YookassaRefundRequest{
		Amount: &YookassaAmount{
			Value:    amountStr,
			Currency: setting.YoomoneyCurrency,
		},
		PaymentID: paymentID,
	}

	bodyBytes, err := common.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("序列化退款请求失败: %w", err)
	}

	req, err := createYookassaRequest("POST", "/refunds", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Idempotence-Key", common.GetUUID())

	resp, err := createYookassaHTTPClient().Do(req)
	if err != nil {
		return fmt.Errorf("YooKassa 退款请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var apiErr YookassaError
		if common.Unmarshal(respBody, &apiErr) == nil && apiErr.Code != "" {
			return &apiErr
		}
		return fmt.Errorf("YooKassa 退款返回 HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	common.SysLog(fmt.Sprintf("YooKassa refund created: payment_id=%s, amount=%s RUB", paymentID, amountStr))
	return nil
}
