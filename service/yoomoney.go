package service

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting"
)

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

// YooMoneyPaymentType 支付方式常量
const (
	YooMoneyPaymentTypeWallet = "PC" // YooMoney 钱包
	YooMoneyPaymentTypeCard   = "AC" // 银行卡（俄罗斯的 Visa/Mastercard 已不可用，仅 Mir）
	YooMoneyPaymentTypeMobile = "MC" // 手机余额
	YooMoneyPaymentTypeSber   = "SB" // SberPay（通过 Sberbank Online）
)

// CreateYooMoneySubscriptionOrder 为 YooMoney 创建订阅订单
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

// CreateYooMoneyTopUpOrder 为 YooMoney 创建充值订单
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
