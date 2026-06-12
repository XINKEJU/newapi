package controller

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)

// YooMoneyNotify 处理 YooMoney 异步通知（webhook）
// 文档：https://yoomoney.ru/docs/payment-notifications/using/epay/notifications
func YooMoneyNotify(c *gin.Context) {
	var params map[string]string

	if c.Request.Method == "POST" {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			common.SysError("YooMoney notify: read body failed: " + err.Error())
			c.String(http.StatusBadRequest, "fail")
			return
		}
		c.Request.Body = io.NopCloser(strings.NewReader(string(body)))

		if err := c.Request.ParseForm(); err != nil {
			c.String(http.StatusBadRequest, "fail")
			return
		}
		params = lo.Reduce(lo.Keys(c.Request.PostForm), func(r map[string]string, t string, i int) map[string]string {
			r[t] = c.Request.PostForm.Get(t)
			return r
		}, map[string]string{})
	} else {
		params = lo.Reduce(lo.Keys(c.Request.URL.Query()), func(r map[string]string, t string, i int) map[string]string {
			r[t] = c.Request.URL.Query().Get(t)
			return r
		}, map[string]string{})
	}

	if len(params) == 0 {
		c.String(http.StatusBadRequest, "fail")
		return
	}

	// 验证签名
	notificationSecret := setting.YoomoneyNotifySecret
	if notificationSecret == "" {
		common.SysError("YooMoney notify: notification_secret not configured")
		c.String(http.StatusInternalServerError, "fail")
		return
	}

	// 构造 YooMoneyNotification 用于验证
	notif := &service.YooMoneyNotification{
		NotificationType: params["notification_type"],
		WalletId:         params["wallet_id"],
		Amount:           params["amount"],
		Currency:         params["currency"],
		TransId:          params["transaction_id"],
		Label:            params["label"],
		Sha1Hash:         params["sha1_hash"],
	}

	if !service.VerifyYooMoneyNotification(notif, notificationSecret) {
		common.SysError(fmt.Sprintf("YooMoney notify: signature verification failed, params: %+v", params))
		c.String(http.StatusOK, "fail")
		return
	}

	tradeNo := params["label"]
	if tradeNo == "" {
		c.String(http.StatusOK, "fail")
		return
	}

	// 根据订单号前缀判断是充值单还是订阅单
	if strings.HasPrefix(tradeNo, "YMSUB") || strings.HasPrefix(tradeNo, "SUB") {
		// 订阅订单
		LockOrder(tradeNo)
		defer UnlockOrder(tradeNo)

		actualPaymentMethod := "yoomoney"
		if params["payment_type"] != "" {
			actualPaymentMethod = "yoomoney_" + params["payment_type"]
		}

		if err := model.CompleteSubscriptionOrder(tradeNo, common.GetJsonString(params), model.PaymentProviderYoomoney, actualPaymentMethod); err != nil {
			common.SysError("YooMoney notify: complete subscription order failed: " + err.Error())
			c.String(http.StatusOK, "fail")
			return
		}
	} else {
		// 充值订单
		LockOrder(tradeNo)
		defer UnlockOrder(tradeNo)

		if err := model.RechargeYoomoney(tradeNo, c.ClientIP()); err != nil {
			common.SysError("YooMoney notify: recharge failed: " + err.Error())
			c.String(http.StatusOK, "fail")
			return
		}
	}

	c.String(http.StatusOK, "success")
}

// YooMoneyReturn 处理 YooMoney 支付后浏览器回跳
func YooMoneyReturn(c *gin.Context) {
	params := lo.Reduce(lo.Keys(c.Request.URL.Query()), func(r map[string]string, t string, i int) map[string]string {
		r[t] = c.Request.URL.Query().Get(t)
		return r
	}, map[string]string{})

	if len(params) == 0 {
		c.Redirect(http.StatusFound, paymentReturnPath("/console/wallet?pay=fail"))
		return
	}

	notificationSecret := setting.YoomoneyNotifySecret
	if notificationSecret == "" {
		c.Redirect(http.StatusFound, paymentReturnPath("/console/wallet?pay=fail"))
		return
	}

	notif := &service.YooMoneyNotification{
		NotificationType: params["notification_type"],
		WalletId:         params["wallet_id"],
		Amount:           params["amount"],
		Currency:         params["currency"],
		TransId:          params["transaction_id"],
		Label:            params["label"],
		Sha1Hash:         params["sha1_hash"],
	}

	if !service.VerifyYooMoneyNotification(notif, notificationSecret) {
		c.Redirect(http.StatusFound, paymentReturnPath("/console/wallet?pay=fail"))
		return
	}

	tradeNo := params["label"]
	if tradeNo == "" {
		c.Redirect(http.StatusFound, paymentReturnPath("/console/wallet?pay=fail"))
		return
	}

	// 完成订单
	if strings.HasPrefix(tradeNo, "YMSUB") || strings.HasPrefix(tradeNo, "SUB") {
		LockOrder(tradeNo)
		defer UnlockOrder(tradeNo)
		_ = model.CompleteSubscriptionOrder(tradeNo, common.GetJsonString(params), model.PaymentProviderYoomoney, "yoomoney")
	} else {
		LockOrder(tradeNo)
		defer UnlockOrder(tradeNo)
		_ = model.RechargeYoomoney(tradeNo, c.ClientIP())
	}

	c.Redirect(http.StatusFound, paymentReturnPath("/console/wallet?pay=success"))
}

// YooMoneySubscriptionReturn 处理订阅支付的浏览器回跳
func YooMoneySubscriptionReturn(c *gin.Context) {
	YooMoneyReturn(c)
}
