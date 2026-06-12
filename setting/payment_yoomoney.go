package setting

import (
	"os"
	"strconv"
	"strings"
)

// 从环境变量读取，若未设置则使用默认值
func getEnvBool(key string, defaultVal bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return strings.ToLower(val) == "true" || val == "1"
}

func getEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return intVal
}

func getEnvString(key string, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}

var (
	YoomoneyEnabled     bool   = getEnvBool("YOOMONEY_ENABLED", true)       // Russia localized: default enabled
	YoomoneyWalletId    string = getEnvString("YOOMONEY_WALLET_ID", "")    // YooMoney 钱包 ID（收款号）
	YoomoneyApiKey      string = getEnvString("YOOMONEY_API_KEY", "")      // API 密钥（用于签名验证）
	YoomoneyNotifySecret string = getEnvString("YOOMONEY_NOTIFY_SECRET", "") // 通知密钥（webhook 签名）
	YoomoneyTestMode    bool   = getEnvBool("YOOMONEY_TEST_MODE", false)   // 沙箱/测试模式
	YoomoneyCurrency    string = getEnvString("YOOMONEY_CURRENCY", "RUB")  // 货币：RUB / USD 等
	YoomoneyMinTopUp    int    = getEnvInt("YOOMONEY_MIN_TOPUP", 50)       // 最小充值金额（RUB）
)

// IsYoomoneyEnabled 返回 YooMoney 是否启用
func IsYoomoneyEnabled() bool {
	return YoomoneyEnabled && YoomoneyWalletId != ""
}

// GetYoomoneyMinTopUp 返回最小充值金额
func GetYoomoneyMinTopUp() int {
	if YoomoneyMinTopUp <= 0 {
		return 50
	}
	return YoomoneyMinTopUp
}

// InitYoomoneyFromEnv 用环境变量覆盖数据库配置（环境变量优先）
// 在 loadOptionsFromDatabase 之后调用
func InitYoomoneyFromEnv() {
	if os.Getenv("YOOMONEY_ENABLED") != "" {
		YoomoneyEnabled = getEnvBool("YOOMONEY_ENABLED", YoomoneyEnabled)
	}
	if os.Getenv("YOOMONEY_WALLET_ID") != "" {
		YoomoneyWalletId = os.Getenv("YOOMONEY_WALLET_ID")
	}
	if os.Getenv("YOOMONEY_API_KEY") != "" {
		YoomoneyApiKey = os.Getenv("YOOMONEY_API_KEY")
	}
	if os.Getenv("YOOMONEY_NOTIFY_SECRET") != "" {
		YoomoneyNotifySecret = os.Getenv("YOOMONEY_NOTIFY_SECRET")
	}
	if os.Getenv("YOOMONEY_TEST_MODE") != "" {
		YoomoneyTestMode = getEnvBool("YOOMONEY_TEST_MODE", YoomoneyTestMode)
	}
	if os.Getenv("YOOMONEY_CURRENCY") != "" {
		YoomoneyCurrency = os.Getenv("YOOMONEY_CURRENCY")
	}
	if os.Getenv("YOOMONEY_MIN_TOPUP") != "" {
		YoomoneyMinTopUp = getEnvInt("YOOMONEY_MIN_TOPUP", YoomoneyMinTopUp)
	}
}
