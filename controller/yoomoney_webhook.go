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

// YooMoneyNotify 处理 YooMoney / YooKassa 异步通知（webhook）
// 支持两种协议：
//   1. 旧 YooMoney Quick Pay 表单通知（form-encoded + SHA-1 签名）
//   2. 新 YooKassa v3 REST API JSON webhook
func YooMoneyNotify(c *gin.Context) {
	contentType := c.GetHeader("Content-Type")

	// =====================================================================
	// 检测是否为 YooKassa JSON webhook
	// =====================================================================
	if strings.Contains(contentType, "application/json") || setting.IsYookassaMode() {
		handleYookassaWebhook(c)
		return
	}

	// =====================================================================
	// 旧协议：YooMoney Quick Pay 表单通知
	// =====================================================================
	handleYoomoneyFormNotify(c)
}

// handleYookassaWebhook 处理 YooKassa v3 API JSON webhook 通知
func handleYookassaWebhook(c *gin.Context) {
	// 读取请求体
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		common.SysError("YooKassa webhook: read body failed: " + err.Error())
		c.String(http.StatusBadRequest, "fail")
		return
	}

	// 解析通知
	event, paymentObj, orderID, err := service.HandleYookassaWebhook(body)
	if err != nil {
		common.SysError("YooKassa webhook: parse failed: " + err.Error())
		c.String(http.StatusBadRequest, "fail")
		return
	}

	// 只处理支付成功事件
	if event != "payment.succeeded" {
		common.SysLog(fmt.Sprintf("YooKassa webhook: ignoring event=%s, order_id=%s", event, orderID))
		c.String(http.StatusOK, "success")
		return
	}

	tradeNo := orderID
	if tradeNo == "" {
		common.SysError("YooKassa webhook: order_id missing in metadata")
		c.String(http.StatusBadRequest, "fail")
		return
	}

	// 根据订单号前缀判断是充值单还是订阅单
	if strings.HasPrefix(tradeNo, "YMSUB") || strings.HasPrefix(tradeNo, "SUB") {
		// 订阅订单
		LockOrder(tradeNo)
		defer UnlockOrder(tradeNo)

		actualPaymentMethod := "yoomoney"
		if paymentObj != nil && paymentObj.ID != "" {
			actualPaymentMethod = "yookassa_" + strings.ReplaceAll(paymentObj.ID, "-", "")
		}

		if err := model.CompleteSubscriptionOrder(tradeNo, common.GetJsonString(body), model.PaymentProviderYoomoney, actualPaymentMethod); err != nil {
			common.SysError("YooKassa webhook: complete subscription order failed: " + err.Error())
			c.String(http.StatusOK, "fail")
			return
		}
	} else {
		// 充值订单
		LockOrder(tradeNo)
		defer UnlockOrder(tradeNo)

		if err := model.RechargeYoomoney(tradeNo, c.ClientIP()); err != nil {
			common.SysError("YooKassa webhook: recharge failed: " + err.Error())
			c.String(http.StatusOK, "fail")
			return
		}
	}

	c.String(http.StatusOK, "success")
}

// handleYoomoneyFormNotify 处理旧 YooMoney Quick Pay 表单通知
func handleYoomoneyFormNotify(c *gin.Context) {
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
	// YooKassa 模式下，成功页回跳直接重定向到钱包页
	// 无需验证签名，YooKassa 会在 return_url 后附加 ?payment_id=xxx
	// 真正的支付成功处理由 webhook 完成
	if setting.IsYookassaMode() {
		c.Redirect(http.StatusFound, paymentReturnPath("/console/wallet?pay=success"))
		return
	}

	// 旧 Quick Pay 协议：验证 SHA-1 签名
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

// YookassaWebhookDebug 处理 YooKassa webhook 调试回跳（附加 payment_id 参数）
// 当用户从 YooKassa 支付页返回时，YooKassa 会在 return_url 后附加查询参数
// 此函数仅作日志记录，实际支付处理由 webhook 异步完成
func YookassaWebhookDebug(c *gin.Context) {
	paymentID := c.Query("payment_id")
	if paymentID != "" {
		common.SysLog(fmt.Sprintf("YooKassa return: payment_id=%s", paymentID))

		// 异步查询支付状态作为日志（不阻塞响应）
		go func() {
			payment, err := service.QueryYookassaPayment(paymentID)
			if err != nil {
				common.SysError(fmt.Sprintf("YooKassa return: query payment failed: %v", err))
				return
			}
			common.SysLog(fmt.Sprintf("YooKassa return: payment status=%s for payment_id=%s", payment.Status, paymentID))
		}()
	}

	c.Redirect(http.StatusFound, paymentReturnPath("/console/wallet?pay=success"))
}
