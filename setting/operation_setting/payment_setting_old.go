/**
此文件为旧版支付设置文件，如需增加新的参数、变量等，请在 payment_setting.go 中添加
This file is the old version of the payment settings file. If you need to add new parameters, variables, etc., please add them in payment_setting.go
*/

package operation_setting

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
)

var PayAddress = ""
var CustomCallbackAddress = ""
var EpayId = ""
var EpayKey = ""
var Price = 7.3 // 1 USD = X units of local currency (override via PRICE env or DB)
var MinTopUp = 1
var USDExchangeRate = 7.25    // 1 USD = 7.25 CNY（人民币对美元汇率）
var CNYExchangeRate = 7.25    // 1 USD = X CNY
var RUBExchangeRate = 90.0    // 1 USD = X RUB
var CNYRUBExchangeRate = 0.08 // 1 RUB = X 元（人民币对卢布汇率）

func init() {
	// Price and exchange rate defaults — override via env for regional deployments
	if v := os.Getenv("PRICE"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			Price = f
			common.SysLog(fmt.Sprintf("env override: Price = %.2f", f))
		}
	}
	if v := os.Getenv("CNY_EXCHANGE_RATE"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			CNYExchangeRate = f
			common.SysLog(fmt.Sprintf("env override: CNYExchangeRate = %.2f", f))
		}
	}
	if v := os.Getenv("RUB_EXCHANGE_RATE"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			RUBExchangeRate = f
			common.SysLog(fmt.Sprintf("env override: RUBExchangeRate = %.2f", f))
		}
	}
	if v := os.Getenv("CNY_RUB_EXCHANGE_RATE"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			CNYRUBExchangeRate = f
			common.SysLog(fmt.Sprintf("env override: CNYRUBExchangeRate = %.4f", f))
		}
	}
	if v := os.Getenv("QUOTA_DISPLAY_TYPE"); v != "" {
		switch strings.ToUpper(v) {
		case "USD", "CNY", "RUB", "TOKENS", "CUSTOM":
			GetGeneralSetting().QuotaDisplayType = strings.ToUpper(v)
			common.SysLog(fmt.Sprintf("env override: QuotaDisplayType = %s", v))
		}
	}
}

var PayMethods = []map[string]string{
	{
		"name":  "YooMoney",
		"color": "rgba(140, 60, 230, 1)",
		"type":  "yoomoney",
	},
	{
		"name":  "Stripe",
		"color": "rgba(99, 91, 255, 1)",
		"type":  "stripe",
	},	{
		"name": "支付宝",
		"icon": "SiAlipay",
		"type": "alipay",
	},
	{
		"name": "微信",
		"icon": "SiWechat",
		"type": "wxpay",
	},
	{
		"name":      "自定义1",
		"icon":      "LuCreditCard",
		"type":      "custom1",
		"min_topup": "50",
	},
}

func UpdatePayMethodsByJsonString(jsonString string) error {
	PayMethods = make([]map[string]string, 0)
	return common.Unmarshal([]byte(jsonString), &PayMethods)
}

func PayMethods2JsonString() string {
	jsonBytes, err := common.Marshal(PayMethods)
	if err != nil {
		return "[]"
	}
	return string(jsonBytes)
}

func ContainsPayMethod(method string) bool {
	for _, payMethod := range PayMethods {
		if payMethod["type"] == method {
			return true
		}
	}
	return false
}
