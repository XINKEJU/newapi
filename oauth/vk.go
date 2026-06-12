package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

func init() {
	Register("vk", &VKProvider{})
}

// VKProvider implements OAuth for VK (VKontakte) — the largest Russian social network
type VKProvider struct{}

// vkTokenResponse is the response from VK's OAuth token endpoint
type vkTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	UserID      int64  `json:"user_id"`
	Email       string `json:"email"`
	Error       string `json:"error"`
	ErrorDesc   string `json:"error_description"`
}

// vkUsersResponse is the response from VK's users.get API
type vkUsersResponse struct {
	Response []vkUser `json:"response"`
}

type vkUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Photo100  string `json:"photo_100"`
}

func (p *VKProvider) GetName() string {
	return "VK"
}

func (p *VKProvider) IsEnabled() bool {
	return common.VKOAuthEnabled
}

func (p *VKProvider) ExchangeToken(ctx context.Context, code string, c *gin.Context) (*OAuthToken, error) {
	if code == "" {
		return nil, NewOAuthError(i18n.MsgOAuthInvalidCode, nil)
	}

	logger.LogDebug(ctx, "[OAuth-VK] ExchangeToken: code=%s...", code[:min(len(code), 10)])

	redirectUri := fmt.Sprintf("%s/oauth/vk", common.OptionMap["ServerAddress"])
	url := fmt.Sprintf(
		"https://oauth.vk.com/access_token?client_id=%s&client_secret=%s&redirect_uri=%s&code=%s",
		common.VKClientId, common.VKClientSecret, redirectUri, code,
	)

	logger.LogDebug(ctx, "[OAuth-VK] ExchangeToken: redirect_uri=%s", redirectUri)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := http.Client{Timeout: 20 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		logger.LogError(ctx, fmt.Sprintf("[OAuth-VK] ExchangeToken error: %s", err.Error()))
		return nil, NewOAuthErrorWithRaw(i18n.MsgOAuthConnectFailed, map[string]any{"Provider": "VK"}, err.Error())
	}
	defer res.Body.Close()

	logger.LogDebug(ctx, "[OAuth-VK] ExchangeToken response status: %d", res.StatusCode)

	var tokenResp vkTokenResponse
	if err := json.NewDecoder(res.Body).Decode(&tokenResp); err != nil {
		logger.LogError(ctx, fmt.Sprintf("[OAuth-VK] ExchangeToken decode error: %s", err.Error()))
		return nil, err
	}

	if tokenResp.Error != "" {
		logger.LogError(ctx, fmt.Sprintf("[OAuth-VK] ExchangeToken failed: error=%s desc=%s", tokenResp.Error, tokenResp.ErrorDesc))
		return nil, NewOAuthError(i18n.MsgOAuthTokenFailed, map[string]any{"Provider": "VK"})
	}

	if tokenResp.AccessToken == "" {
		logger.LogError(ctx, "[OAuth-VK] ExchangeToken failed: empty access token")
		return nil, NewOAuthError(i18n.MsgOAuthTokenFailed, map[string]any{"Provider": "VK"})
	}

	logger.LogDebug(ctx, "[OAuth-VK] ExchangeToken success: user_id=%d, email=%s", tokenResp.UserID, tokenResp.Email)

	return &OAuthToken{
		AccessToken: tokenResp.AccessToken,
		ExpiresIn:   tokenResp.ExpiresIn,
		Extra: map[string]any{
			"vk_user_id": strconv.FormatInt(tokenResp.UserID, 10),
			"vk_email":   tokenResp.Email,
		},
	}, nil
}

func (p *VKProvider) GetUserInfo(ctx context.Context, token *OAuthToken) (*OAuthUser, error) {
	logger.LogDebug(ctx, "[OAuth-VK] GetUserInfo: fetching user info")

	// Get basic profile from users.get
	userID := ""
	if extra, ok := token.Extra["vk_user_id"].(string); ok {
		userID = extra
	}
	email := ""
	if em, ok := token.Extra["vk_email"].(string); ok {
		email = em
	}

	url := fmt.Sprintf("https://api.vk.com/method/users.get?access_token=%s&v=5.131&fields=photo_100", token.AccessToken)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := http.Client{Timeout: 20 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		logger.LogError(ctx, fmt.Sprintf("[OAuth-VK] GetUserInfo error: %s", err.Error()))
		return nil, NewOAuthErrorWithRaw(i18n.MsgOAuthConnectFailed, map[string]any{"Provider": "VK"}, err.Error())
	}
	defer res.Body.Close()

	logger.LogDebug(ctx, "[OAuth-VK] GetUserInfo response status: %d", res.StatusCode)

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		bodyStr := string(body)
		if len(bodyStr) > 500 {
			bodyStr = bodyStr[:500] + "..."
		}
		logger.LogError(ctx, fmt.Sprintf("[OAuth-VK] GetUserInfo failed: status=%d, body=%s", res.StatusCode, bodyStr))
		return nil, NewOAuthErrorWithRaw(i18n.MsgOAuthGetUserErr, map[string]any{"Provider": "VK"}, fmt.Sprintf("status %d", res.StatusCode))
	}

	var usersResp vkUsersResponse
	if err := json.NewDecoder(res.Body).Decode(&usersResp); err != nil {
		logger.LogError(ctx, fmt.Sprintf("[OAuth-VK] GetUserInfo decode error: %s", err.Error()))
		return nil, err
	}

	if len(usersResp.Response) == 0 {
		logger.LogError(ctx, "[OAuth-VK] GetUserInfo failed: empty response array")
		return nil, NewOAuthError(i18n.MsgOAuthUserInfoEmpty, map[string]any{"Provider": "VK"})
	}

	vkUser := usersResp.Response[0]
	if vkUser.ID == 0 {
		logger.LogError(ctx, "[OAuth-VK] GetUserInfo failed: empty user id")
		return nil, NewOAuthError(i18n.MsgOAuthUserInfoEmpty, map[string]any{"Provider": "VK"})
	}

	providerUserID := strconv.FormatInt(vkUser.ID, 10)
	displayName := vkUser.FirstName + " " + vkUser.LastName

	// Use email from token exchange response if available; fall back to construct one
	userEmail := email

	// If we didn't get userID from token extra, use the one from users.get response
	if userID == "" {
		userID = providerUserID
	}

	logger.LogDebug(ctx, "[OAuth-VK] GetUserInfo success: id=%s, name=%s, email=%s", providerUserID, displayName, userEmail)

	return &OAuthUser{
		ProviderUserID: providerUserID,
		Username:       "vk_" + userID,
		DisplayName:    displayName,
		Email:          userEmail,
		Extra: map[string]any{
			"photo_url": vkUser.Photo100,
		},
	}, nil
}

func (p *VKProvider) IsUserIDTaken(providerUserID string) bool {
	return model.IsVkIdAlreadyTaken(providerUserID)
}

func (p *VKProvider) FillUserByProviderID(user *model.User, providerUserID string) error {
	user.VkId = providerUserID
	return user.FillUserByVkId()
}

func (p *VKProvider) SetProviderUserID(user *model.User, providerUserID string) {
	user.VkId = providerUserID
}

func (p *VKProvider) GetProviderPrefix() string {
	return "vk_"
}
