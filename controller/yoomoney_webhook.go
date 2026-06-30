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
// 基于 YooMoney HTTP 通知协议（旧版快速支付表单回调）
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

	// 如果 payment 未完成（unaccepted=true），不处理
	if params["unaccepted"] == "true" {
		common.SysLog(fmt.Sprintf("YooMoney notify: payment unaccepted, label=%s", params["label"]))
		c.String(http.StatusOK, "success")
		return
	}

	// 验证签名
	notificationSecret := setting.YoomoneyNotifySecret
	if notificationSecret == "" {
		common.SysError("YooMoney notify: notification_secret not configured")
		c.String(http.StatusInternalServerError, "fail")
		return
	}

	if err := service.VerifyYooMoneyParams(params, notificationSecret); err != nil {
		common.SysError(fmt.Sprintf("YooMoney notify: signature verification failed: %v, label=%s", err, params["label"]))
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
// 注意：浏览器回跳不可靠（用户可能直接关闭页面），主要依赖 webhook
// 这里的处理作为辅助回退，不替代 webhook 通知
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

	if err := service.VerifyYooMoneyParams(params, notificationSecret); err != nil {
		common.SysError(fmt.Sprintf("YooMoney return: signature verification failed: %v, label=%s", err, params["label"]))
		c.Redirect(http.StatusFound, paymentReturnPath("/console/wallet?pay=fail"))
		return
	}

	tradeNo := params["label"]
	if tradeNo == "" {
		c.Redirect(http.StatusFound, paymentReturnPath("/console/wallet?pay=fail"))
		return
	}

	// 幂等处理：如果订单已经由 webhook 处理完成，直接重定向到成功页
	if strings.HasPrefix(tradeNo, "YMSUB") || strings.HasPrefix(tradeNo, "SUB") {
		LockOrder(tradeNo)
		err := model.CompleteSubscriptionOrder(tradeNo, common.GetJsonString(params), model.PaymentProviderYoomoney, "yoomoney")
		UnlockOrder(tradeNo)
		if err != nil {
			common.SysError("YooMoney return: complete subscription order failed: " + err.Error())
		}
	} else {
		LockOrder(tradeNo)
		err := model.RechargeYoomoney(tradeNo, c.ClientIP())
		UnlockOrder(tradeNo)
		if err != nil {
			common.SysError("YooMoney return: recharge failed: " + err.Error())
		}
	}

	c.Redirect(http.StatusFound, paymentReturnPath("/console/wallet?pay=success"))
}

// YooMoneySubscriptionReturn 处理订阅支付的浏览器回跳
func YooMoneySubscriptionReturn(c *gin.Context) {
	YooMoneyReturn(c)
}
