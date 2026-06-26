package controller

import (
	"fmt"
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting"

	"github.com/gin-gonic/gin"
)

// SendSMSVerificationRequest is the request to send an SMS verification code.
type SendSMSVerificationRequest struct {
	Phone   string `json:"phone" binding:"required"`
	Purpose string `json:"purpose"` // optional, defaults to sms_verification
}

// VerifySMSCodeRequest is the request to verify an SMS code.
type VerifySMSCodeRequest struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

// SendSMSVerification sends a verification code via SMS.
func SendSMSVerification(c *gin.Context) {
	if !setting.IsSMSEnabled() {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "SMS service is not enabled",
		})
		return
	}

	var req SendSMSVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Invalid request: phone is required",
		})
		return
	}

	// Normalize phone: strip non-digit chars except leading +
	phone := normalizePhone(req.Phone)
	if len(phone) < 10 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Invalid phone number format",
		})
		return
	}

	purpose := req.Purpose
	if purpose == "" {
		purpose = model.SMSVerificationPurpose
	}

	// Generate code and send
	code := common.GenerateVerificationCode(model.SMSVerificationCodeLength)
	message := fmt.Sprintf("Your verification code: %s", code)

	smsConfig := common.SMSConfig{
		Enabled:  setting.SMSEnabled,
		Provider: setting.SMSProvider,
		APIID:    setting.SMSApiID,
		Sender:   setting.SMSSender,
	}
	if err := common.SendSMS(phone, message, smsConfig); err != nil {
		common.SysError(fmt.Sprintf("SMS send failed to %s: %s", phone, err.Error()))
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Failed to send SMS. Please try again.",
		})
		return
	}

	// Save verification code
	if err := model.RecordSMSVerification(phone, code, purpose, common.VerificationValidMinutes); err != nil {
		common.SysError(fmt.Sprintf("Failed to record SMS verification: %s", err.Error()))
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Failed to save verification code",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}

// VerifySMSCode checks if the provided SMS code is valid.
func VerifySMSCode(c *gin.Context) {
	var req VerifySMSCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Invalid request",
		})
		return
	}

	phone := normalizePhone(req.Phone)
	if !model.VerifySMSCode(phone, req.Code, model.SMSVerificationPurpose) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Invalid or expired verification code",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Phone verified successfully",
	})
}

func normalizePhone(phone string) string {
	// Remove all non-digit characters except +
	normalized := ""
	for _, ch := range phone {
		if ch == '+' || (ch >= '0' && ch <= '9') {
			normalized += string(ch)
		}
	}
	return normalized
}
