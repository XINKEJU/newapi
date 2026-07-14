package setting

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
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
	YoomoneyEnabled      bool   = getEnvBool("YOOMONEY_ENABLED", true)       // Russia localized: default enabled
	YoomoneyWalletId     string = getEnvString("YOOMONEY_WALLET_ID", "")     // YooMoney 钱包 ID（收款号，用于旧 Quick Pay 协议）
	YoomoneyApiKey       string = getEnvString("YOOMONEY_API_KEY", "")       // 商店密码/API密钥（预留，当前未用于签名验证；webhook 验证使用 NotifySecret）
	YoomoneyNotifySecret string = getEnvString("YOOMONEY_NOTIFY_SECRET", "") // 通知密钥（webhook SHA-1 签名验证用，仅旧协议模式）
	YoomoneyTestMode     bool   = getEnvBool("YOOMONEY_TEST_MODE", false)    // 沙箱/测试模式
	YoomoneyCurrency     string = getEnvString("YOOMONEY_CURRENCY", "RUB")   // 货币：RUB / USD 等
	YoomoneyMinTopUp     int    = getEnvInt("YOOMONEY_MIN_TOPUP", 50)        // 最小充值金额（RUB）

	// YooKassa v3 REST API 凭证（新式商户 API）
	// 设置此项后将自动切换到 YooKassa API 模式，替代旧 Quick Pay 协议
	YoomoneyShopId   string = getEnvString("YOOMONEY_SHOP_ID", "")     // YooKassa 商户 Shop ID
	YoomoneySecretKey string = getEnvString("YOOMONEY_SECRET_KEY", "") // YooKassa 商户 Secret Key
)

// IsYoomoneyEnabled 返回 YooMoney 是否启用
// 旧 Quick Pay 模式需要 YoomoneyWalletId
// YooKassa API 模式需要 YoomoneyShopId + YoomoneySecretKey
func IsYoomoneyEnabled() bool {
	if !YoomoneyEnabled {
		return false
	}
	return YoomoneyWalletId != "" || (YoomoneyShopId != "" && YoomoneySecretKey != "")
}

// IsYookassaMode 返回是否使用 YooKassa v3 REST API 模式
// 当 YOOMONEY_SHOP_ID 和 YOOMONEY_SECRET_KEY 都配置时自动启用
func IsYookassaMode() bool {
	return YoomoneyEnabled && YoomoneyShopId != "" && YoomoneySecretKey != ""
}

// GetYookassaAPIEndpoint 返回 YooKassa API 基础 URL
func GetYookassaAPIEndpoint() string {
	return "https://api.yookassa.ru/v3"
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
		common.SysLog("env override: YoomoneyEnabled set")
	}
	if os.Getenv("YOOMONEY_WALLET_ID") != "" {
		YoomoneyWalletId = os.Getenv("YOOMONEY_WALLET_ID")
		common.SysLog("env override: YoomoneyWalletId set")
	}
	if os.Getenv("YOOMONEY_API_KEY") != "" {
		YoomoneyApiKey = os.Getenv("YOOMONEY_API_KEY")
		common.SysLog("env override: YoomoneyApiKey set")
	}
	if os.Getenv("YOOMONEY_NOTIFY_SECRET") != "" {
		YoomoneyNotifySecret = os.Getenv("YOOMONEY_NOTIFY_SECRET")
		common.SysLog("env override: YoomoneyNotifySecret set")
	}
	if os.Getenv("YOOMONEY_TEST_MODE") != "" {
		YoomoneyTestMode = getEnvBool("YOOMONEY_TEST_MODE", YoomoneyTestMode)
		common.SysLog("env override: YoomoneyTestMode set")
	}
	if os.Getenv("YOOMONEY_CURRENCY") != "" {
		YoomoneyCurrency = os.Getenv("YOOMONEY_CURRENCY")
		common.SysLog(fmt.Sprintf("env override: YoomoneyCurrency = %s", YoomoneyCurrency))
	}
	if os.Getenv("YOOMONEY_MIN_TOPUP") != "" {
		YoomoneyMinTopUp = getEnvInt("YOOMONEY_MIN_TOPUP", YoomoneyMinTopUp)
		common.SysLog(fmt.Sprintf("env override: YoomoneyMinTopUp = %d", YoomoneyMinTopUp))
	}
	if os.Getenv("YOOMONEY_SHOP_ID") != "" {
		YoomoneyShopId = os.Getenv("YOOMONEY_SHOP_ID")
		common.SysLog("env override: YoomoneyShopId set")
	}
	if os.Getenv("YOOMONEY_SECRET_KEY") != "" {
		YoomoneySecretKey = os.Getenv("YOOMONEY_SECRET_KEY")
		common.SysLog("env override: YoomoneySecretKey set")
	}
}
