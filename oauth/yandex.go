package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

func init() {
	Register("yandex", &YandexProvider{})
}

// YandexProvider implements OAuth for Yandex ID — the most popular Russian single sign-on
type YandexProvider struct{}

type yandexTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Error       string `json:"error"`
	ErrorDesc   string `json:"error_description"`
}

type yandexUserInfo struct {
	ID           string   `json:"id"`
	Login        string   `json:"login"`
	RealName     string   `json:"real_name"`
	DisplayName  string   `json:"display_name"`
	DefaultEmail string   `json:"default_email"`
	DefaultPhone struct {
		ID     int    `json:"id"`
		Number string `json:"number"`
	} `json:"default_phone"`
	Emails     []string `json:"emails"`
	Sex        string   `json:"sex"`
	Birthday   string   `json:"birthday"`
	AvatarID   string   `json:"default_avatar_id"`
	IsAvatarEmpty bool  `json:"is_avatar_empty"`
}

func (p *YandexProvider) GetName() string {
	return "Yandex"
}

func (p *YandexProvider) IsEnabled() bool {
	return common.YandexOAuthEnabled
}

func (p *YandexProvider) ExchangeToken(ctx context.Context, code string, c *gin.Context) (*OAuthToken, error) {
	if code == "" {
		return nil, NewOAuthError(i18n.MsgOAuthInvalidCode, nil)
	}

	logger.LogDebug(ctx, "[OAuth-Yandex] ExchangeToken: code=%s...", code[:min(len(code), 10)])

	values := url.Values{}
	values.Set("grant_type", "authorization_code")
	values.Set("code", code)
	values.Set("client_id", common.YandexClientId)
	values.Set("client_secret", common.YandexClientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://oauth.yandex.ru/token", strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := http.Client{Timeout: 20 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		logger.LogError(ctx, fmt.Sprintf("[OAuth-Yandex] ExchangeToken error: %s", err.Error()))
		return nil, NewOAuthErrorWithRaw(i18n.MsgOAuthConnectFailed, map[string]any{"Provider": "Yandex"}, err.Error())
	}
	defer res.Body.Close()

	logger.LogDebug(ctx, "[OAuth-Yandex] ExchangeToken response status: %d", res.StatusCode)

	var tokenResp yandexTokenResponse
	if err := json.NewDecoder(res.Body).Decode(&tokenResp); err != nil {
		logger.LogError(ctx, fmt.Sprintf("[OAuth-Yandex] ExchangeToken decode error: %s", err.Error()))
		return nil, err
	}

	if tokenResp.Error != "" {
		logger.LogError(ctx, fmt.Sprintf("[OAuth-Yandex] ExchangeToken failed: error=%s desc=%s", tokenResp.Error, tokenResp.ErrorDesc))
		return nil, NewOAuthError(i18n.MsgOAuthTokenFailed, map[string]any{"Provider": "Yandex"})
	}

	if tokenResp.AccessToken == "" {
		logger.LogError(ctx, "[OAuth-Yandex] ExchangeToken failed: empty access token")
		return nil, NewOAuthError(i18n.MsgOAuthTokenFailed, map[string]any{"Provider": "Yandex"})
	}

	logger.LogDebug(ctx, "[OAuth-Yandex] ExchangeToken success: token_type=%s", tokenResp.TokenType)

	return &OAuthToken{
		AccessToken: tokenResp.AccessToken,
		TokenType:   tokenResp.TokenType,
		ExpiresIn:   tokenResp.ExpiresIn,
	}, nil
}

func (p *YandexProvider) GetUserInfo(ctx context.Context, token *OAuthToken) (*OAuthUser, error) {
	logger.LogDebug(ctx, "[OAuth-Yandex] GetUserInfo: fetching user info")

	req, err := http.NewRequestWithContext(ctx, "GET", "https://login.yandex.ru/info?format=json", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "OAuth "+token.AccessToken)

	client := http.Client{Timeout: 20 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		logger.LogError(ctx, fmt.Sprintf("[OAuth-Yandex] GetUserInfo error: %s", err.Error()))
		return nil, NewOAuthErrorWithRaw(i18n.MsgOAuthConnectFailed, map[string]any{"Provider": "Yandex"}, err.Error())
	}
	defer res.Body.Close()

	logger.LogDebug(ctx, "[OAuth-Yandex] GetUserInfo response status: %d", res.StatusCode)

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		bodyStr := string(body)
		if len(bodyStr) > 500 {
			bodyStr = bodyStr[:500] + "..."
		}
		logger.LogError(ctx, fmt.Sprintf("[OAuth-Yandex] GetUserInfo failed: status=%d, body=%s", res.StatusCode, bodyStr))
		return nil, NewOAuthErrorWithRaw(i18n.MsgOAuthGetUserErr, map[string]any{"Provider": "Yandex"}, fmt.Sprintf("status %d", res.StatusCode))
	}

	var yaUser yandexUserInfo
	if err := json.NewDecoder(res.Body).Decode(&yaUser); err != nil {
		logger.LogError(ctx, fmt.Sprintf("[OAuth-Yandex] GetUserInfo decode error: %s", err.Error()))
		return nil, err
	}

	if yaUser.ID == "" {
		logger.LogError(ctx, "[OAuth-Yandex] GetUserInfo failed: empty user id")
		return nil, NewOAuthError(i18n.MsgOAuthUserInfoEmpty, map[string]any{"Provider": "Yandex"})
	}

	username := yaUser.Login
	if username == "" {
		username = "ya_" + yaUser.ID
	}

	displayName := yaUser.RealName
	if displayName == "" {
		displayName = yaUser.DisplayName
	}
	if displayName == "" {
		displayName = username
	}

	email := yaUser.DefaultEmail

	logger.LogDebug(ctx, "[OAuth-Yandex] GetUserInfo success: id=%s, login=%s, name=%s, email=%s", yaUser.ID, yaUser.Login, displayName, email)

	return &OAuthUser{
		ProviderUserID: yaUser.ID,
		Username:       username,
		DisplayName:    displayName,
		Email:          email,
		Extra: map[string]any{
			"avatar_id": yaUser.AvatarID,
			"sex":       yaUser.Sex,
		},
	}, nil
}

func (p *YandexProvider) IsUserIDTaken(providerUserID string) bool {
	return model.IsYandexIdAlreadyTaken(providerUserID)
}

func (p *YandexProvider) FillUserByProviderID(user *model.User, providerUserID string) error {
	user.YandexId = providerUserID
	return user.FillUserByYandexId()
}

func (p *YandexProvider) SetProviderUserID(user *model.User, providerUserID string) {
	user.YandexId = providerUserID
}

func (p *YandexProvider) GetProviderPrefix() string {
	return "yandex_"
}
