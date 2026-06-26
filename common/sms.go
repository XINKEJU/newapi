package common

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// SMSConfig holds SMS provider configuration.
type SMSConfig struct {
	Enabled  bool
	Provider string // "smsru" or "smsc"
	APIID    string // SMS.ru API ID
	Sender   string // Sender name
}

// SMSResponse is the response from SMS.ru API
type SMSResponse struct {
	Status     string              `json:"status"`
	StatusCode int                 `json:"status_code"`
	SMS        map[string]SMSInfo  `json:"sms,omitempty"`
	Balance    float64             `json:"balance,omitempty"`
	Message    string              `json:"message,omitempty"`
}

// SMSInfo holds per-phone SMS sending result
type SMSInfo struct {
	Status     string `json:"status"`
	StatusCode int    `json:"status_code"`
	SMSID      string `json:"sms_id,omitempty"`
	StatusText string `json:"status_text,omitempty"`
}

// SendSMS sends an SMS message via the configured provider.
// phone: international format (e.g. "79161234567")
// message: text content
// config: SMS configuration from settings
func SendSMS(phone string, message string, config SMSConfig) error {
	if phone == "" {
		return fmt.Errorf("phone number is empty")
	}
	if !config.Enabled || config.APIID == "" {
		return fmt.Errorf("SMS is not enabled")
	}

	switch config.Provider {
	case "smsru":
		return sendViaSMSRu(phone, message, config)
	default:
		return fmt.Errorf("unsupported SMS provider: %s", config.Provider)
	}
}

func sendViaSMSRu(phone string, message string, config SMSConfig) error {
	baseURL := "https://sms.ru/sms/send"
	params := url.Values{}
	params.Set("api_id", config.APIID)
	params.Set("to", phone)
	params.Set("msg", message)
	params.Set("json", "1")
	if config.Sender != "" {
		params.Set("from", config.Sender)
	}

	client := http.Client{Timeout: 15 * time.Second}
	resp, err := client.PostForm(baseURL, params)
	if err != nil {
		return fmt.Errorf("SMS.ru request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("SMS.ru read response failed: %w", err)
	}

	var smsResp SMSResponse
	if err := json.Unmarshal(body, &smsResp); err != nil {
		return fmt.Errorf("SMS.ru parse response failed: %w (body: %s)", err, string(body))
	}

	if smsResp.Status != "OK" {
		return fmt.Errorf("SMS.ru API error: status=%s, code=%d, msg=%s", smsResp.Status, smsResp.StatusCode, smsResp.Message)
	}

	if smsInfo, ok := smsResp.SMS[phone]; ok {
		if smsInfo.Status != "OK" {
			return fmt.Errorf("SMS.ru send failed for %s: code=%d, text=%s", phone, smsInfo.StatusCode, smsInfo.StatusText)
		}
	} else {
		return fmt.Errorf("SMS.ru response missing status for %s", phone)
	}

	SysLog(fmt.Sprintf("SMS sent to %s, sms_id=%s", phone, smsResp.SMS[phone].SMSID))
	return nil
}
