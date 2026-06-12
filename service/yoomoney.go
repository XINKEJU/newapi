package service

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting"
)

// YooMoneyQuickPayParams 封装 YooMoney 快速支付表单参数
type YooMoneyQuickPayParams struct {
	Receiver     string // 收款钱包 ID（wallet_id）
	QuickpayForm string // 表单类型，固定 "shop"
	Targets      string // 支付描述（显示在支付页）
	PaymentType  string // 支付方式：PC=钱包、AC=银行卡、MC=手机余额
	Sum          float64  // 金额
	Label        string // 订单号（trade_no，用于回调识别）
	SuccessURL   string // 支付成功回跳地址
	FailURL      string // 支付失败回跳地址
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
// 文档：https://yoomoney.ru/docs/payment-notifications/using/epay/notifications
type YooMoneyNotification struct {
	NotificationType string `form:"notification_type" json:"notification_type"`
	WalletId         string `form:"wallet_id"         json:"wallet_id"`
	Amount           string `form:"amount"              json:"amount"`
	Currency         string `form:"currency"            json:"currency"`
	TransId          string `form:"transaction_id"      json:"transaction_id"`
	Label            string `form:"label"               json:"label"`
	Sha1Hash         string `form:"sha1_hash"          json:"sha1_hash"`
	Unaccepted       string `form:"unaccepted"          json:"unaccepted"`
	CreatedAt        string `form:"created_at"          json:"created_at"`
	WithdrawId       string `form:"withdraw_id"        json:"withdraw_id"`
}

// VerifyYooMoneyNotification 验证 YooMoney 通知签名
// 签名算法：SHA1(notification_type&amount&currency&wallet_id&transaction_id&label&<notification_secret>)
func VerifyYooMoneyNotification(n *YooMoneyNotification, notificationSecret string) bool {
	if notificationSecret == "" {
		common.SysError("YooMoney notification secret not configured")
		return false
	}

	// 按文档要求的顺序拼接
	toSign := strings.Join([]string{
		n.NotificationType,
		n.Amount,
		n.Currency,
		n.WalletId,
		n.TransId,
		n.Label,
		notificationSecret,
	}, "&")

	h := sha1.New()
	h.Write([]byte(toSign))
	expected := hex.EncodeToString(h.Sum(nil))

	return expected == strings.ToLower(n.Sha1Hash)
}

// YooMoneyPaymentType 支付方式常量
const (
	YooMoneyPaymentTypeWallet  = "PC" // YooMoney 钱包
	YooMoneyPaymentTypeCard    = "AC" // 银行卡（俄罗斯的 Visa/Mastercard 已不可用，仅 Mir）
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

func init() {
	// 注册 YooMoney 配置热更新监听（如有需要可在此添加）
	_ = time.Now() // 占位，避免空 init 编译错误
}
