/**
此文件为旧版支付设置文件，如需增加新的参数、变量等，请在 payment_setting.go 中添加
This file is the old version of the payment settings file. If you need to add new parameters, variables, etc., please add them in payment_setting.go
*/

package operation_setting

import (
	"github.com/QuantumNous/new-api/common"
)

var PayAddress = ""
var CustomCallbackAddress = ""
var EpayId = ""
var EpayKey = ""
var Price = 90.0
var MinTopUp = 1
var USDExchangeRate = 7.25    // 1 USD = 7.25 CNY（原值 90 是 RUB 汇率，现已拆分）
var CNYExchangeRate = 7.25    // 1 USD = X CNY
var RUBExchangeRate = 90.0    // 1 USD = X RUB

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
	},
	{
		"name":  "SberPay",
		"color": "rgba(34, 167, 77, 1)",
		"type":  "sberpay",
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
