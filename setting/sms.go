package setting

var (
	SMSEnabled   bool   = getEnvBool("SMS_ENABLED", false)
	SMSProvider  string = getEnvString("SMS_PROVIDER", "smsru") // smsru or smsc
	SMSApiID     string = getEnvString("SMS_API_ID", "")       // SMS.ru API ID (api_id)
	SMSSender    string = getEnvString("SMS_SENDER", "")       // Sender name (SMS.ru registered name)
	SMSTestPhone string = getEnvString("SMS_TEST_PHONE", "")   // Test phone for dev mode
)

func IsSMSEnabled() bool {
	return SMSEnabled && SMSApiID != ""
}
