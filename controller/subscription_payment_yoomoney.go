package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting"
	"github.com/gin-gonic/gin"
	"github.com/thanhpk/randstr"
)

type YooMoneyTopUpRequest struct {
	Amount        int    `json:"amount"`
	PaymentMethod string `json:"payment_method"` // 可选：PC/AC/MC，默认 PC
}

// RequestYooMoneyTopUp 发起 YooMoney 余额充值
func RequestYooMoneyTopUp(c *gin.Context) {
	if !requirePaymentCompliance(c) {
		return
	}

	var req YooMoneyTopUpRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Amount <= 0 {
		common.ApiErrorMsg(c, "参数错误")
		return
	}

	if !setting.YoomoneyEnabled {
		common.ApiErrorMsg(c, "YooMoney 支付未启用")
		return
	}

	minAmount := setting.GetYoomoneyMinTopUp()
	if req.Amount < minAmount {
		common.ApiErrorMsg(c, fmt.Sprintf("最低充值金额为 %d", minAmount))
		return
	}

	userId := c.GetInt("id")
	user, err := model.GetUserById(userId, false)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if user == nil {
		common.ApiErrorMsg(c, "用户不存在")
		return
	}

	// 使用美元金额存储，Amount 为美元数额
	money := float64(req.Amount) // 前端传美元金额
	tradeNo := fmt.Sprintf("YM%d-%d-%s", userId, time.Now().UnixMilli(), randstr.String(6))

	topUp := &model.TopUp{
		UserId:          userId,
		Amount:          int64(req.Amount),
		Money:           money,
		TradeNo:         tradeNo,
		PaymentMethod:   model.PaymentMethodYoomoney,
		PaymentProvider: model.PaymentProviderYoomoney,
		CreateTime:      time.Now().Unix(),
		Status:          common.TopUpStatusPending,
	}
	if err := topUp.Insert(); err != nil {
		common.ApiErrorMsg(c, "创建充值订单失败")
		return
	}

	callBackAddress := service.GetCallbackAddress()
	returnURL := callBackAddress + "/api/yoomoney/return"

	payURL, err := service.CreateYooMoneyTopUpOrder(tradeNo, money, returnURL)
	if err != nil {
		_ = model.ExpireSubscriptionOrder(tradeNo, model.PaymentProviderYoomoney)
		common.ApiErrorMsg(c, "拉起支付失败："+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data": gin.H{
			"pay_url":  payURL,
			"order_id": tradeNo,
		},
	})
}

type YooMoneySubscriptionRequest struct {
	PlanId        int    `json:"plan_id"`
	PaymentMethod string `json:"payment_method"` // 可选
}

// RequestYooMoneySubscription 发起 YooMoney 订阅购买
func RequestYooMoneySubscription(c *gin.Context) {
	if !requirePaymentCompliance(c) {
		return
	}

	var req YooMoneySubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.PlanId <= 0 {
		common.ApiErrorMsg(c, "参数错误")
		return
	}

	if !setting.YoomoneyEnabled {
		common.ApiErrorMsg(c, "YooMoney 支付未启用")
		return
	}

	plan, err := model.GetSubscriptionPlanById(req.PlanId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if !plan.Enabled {
		common.ApiErrorMsg(c, "套餐未启用")
		return
	}
	if plan.PriceAmount < 0.01 {
		common.ApiErrorMsg(c, "套餐金额过低")
		return
	}

	userId := c.GetInt("id")
	if plan.MaxPurchasePerUser > 0 {
		count, err := model.CountUserSubscriptionsByPlan(userId, plan.Id)
		if err != nil {
			common.ApiError(c, err)
			return
		}
		if count >= int64(plan.MaxPurchasePerUser) {
			common.ApiErrorMsg(c, "已达到该套餐购买上限")
			return
		}
	}

	tradeNo := fmt.Sprintf("YMSUB%d-%d-%s", userId, time.Now().UnixMilli(), randstr.String(6))

	order := &model.SubscriptionOrder{
		UserId:          userId,
		PlanId:          plan.Id,
		Money:           plan.PriceAmount,
		TradeNo:         tradeNo,
		PaymentMethod:   model.PaymentMethodYoomoney,
		PaymentProvider: model.PaymentProviderYoomoney,
		CreateTime:      time.Now().Unix(),
		Status:          common.TopUpStatusPending,
	}
	if err := order.Insert(); err != nil {
		common.ApiErrorMsg(c, "创建订单失败")
		return
	}

	callBackAddress := service.GetCallbackAddress()
	returnURL := callBackAddress + "/api/yoomoney/subscription/return"
	notifyURL := callBackAddress + "/api/yoomoney/subscription/notify"

	payURL, err := service.CreateYooMoneySubscriptionOrder(tradeNo, plan.PriceAmount,
		fmt.Sprintf("订阅：%s", plan.Title), returnURL, notifyURL)
	if err != nil {
		_ = model.ExpireSubscriptionOrder(tradeNo, model.PaymentProviderYoomoney)
		common.ApiErrorMsg(c, "拉起支付失败："+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data": gin.H{
			"pay_url":   payURL,
			"order_id":  tradeNo,
			"plan_title": plan.Title,
		},
	})
}
