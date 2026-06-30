package gigachat

import (
	"bytes"
	"crypto/tls"
	"github.com/QuantumNous/new-api/common"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relay/helper"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
)

// ──────────────────────────────────────────────────────────────────────────────
// OAuth2 token cache
// ──────────────────────────────────────────────────────────────────────────────

type cachedToken struct {
	token     string
	expiresAt time.Time
}

var (
	tokenStore sync.Map
	// GigaChat uses a self-signed cert on their OAuth endpoint; we skip verification
	// only for the token fetch (not for the main API calls).
	oauthClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
		},
		Timeout: 10 * time.Second,
	}
)

const gigaChatOAuthURL = "https://ngw.devices.sberbank.ru:9443/api/v2/oauth"

// getAccessToken returns a valid Bearer token, fetching a new one if necessary.
// apiKey must be the Base64-encoded "ClientID:ClientSecret" pair issued by
// the Sber developer portal (the "Authorization" header value for the OAuth
// request, without the "Basic " prefix).
func getAccessToken(apiKey string) (string, error) {
	if val, ok := tokenStore.Load(apiKey); ok {
		ct := val.(cachedToken)
		if time.Now().Add(60 * time.Second).Before(ct.expiresAt) {
			return ct.token, nil
		}
	}
	return fetchToken(apiKey)
}

func fetchToken(apiKey string) (string, error) {
	body := bytes.NewBufferString("scope=GIGACHAT_API_PERS")
	req, err := http.NewRequest(http.MethodPost, gigaChatOAuthURL, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+apiKey)
	req.Header.Set("RqUID", common.GetUUID())

	resp, err := oauthClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("gigachat: OAuth request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gigachat: OAuth returned %d: %s", resp.StatusCode, string(respBody))
	}
	var tokenResp GigaChatTokenResponse
	if err := common.Unmarshal(respBody, &tokenResp); err != nil {
		return "", err
	}
	expiresAt := time.UnixMilli(tokenResp.ExpiresAt)
	tokenStore.Store(apiKey, cachedToken{
		token:     tokenResp.AccessToken,
		expiresAt: expiresAt,
	})
	return tokenResp.AccessToken, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Response helpers
// ──────────────────────────────────────────────────────────────────────────────

func gigaChat2OpenAI(resp *GigaChatResponse) *dto.OpenAITextResponse {
	choices := make([]dto.OpenAITextResponseChoice, 0, len(resp.Choices))
	for _, c := range resp.Choices {
		choices = append(choices, dto.OpenAITextResponseChoice{
			Index: c.Index,
			Message: dto.Message{
				Role:    c.Message.Role,
				Content: c.Message.Content,
			},
			FinishReason: c.FinishReason,
		})
	}
	return &dto.OpenAITextResponse{
		Object:  "chat.completion",
		Model:   resp.Model,
		Created: resp.Created,
		Choices: choices,
		Usage: dto.Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}
}

func gigaChatChunk2OpenAI(chunk *GigaChatStreamChunk) *dto.ChatCompletionsStreamResponse {
	choices := make([]dto.ChatCompletionsStreamResponseChoice, 0, len(chunk.Choices))
	for _, c := range chunk.Choices {
		var choice dto.ChatCompletionsStreamResponseChoice
		choice.Delta.SetContentString(c.Delta.Content)
		if c.FinishReason == "stop" || c.FinishReason == "length" {
			choice.FinishReason = &constant.FinishReasonStop
		}
		choices = append(choices, choice)
	}
	return &dto.ChatCompletionsStreamResponse{
		Object:  "chat.completion.chunk",
		Model:   chunk.Model,
		Created: chunk.Created,
		Choices: choices,
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Handlers
// ──────────────────────────────────────────────────────────────────────────────

func gigaChatStreamHandler(c *gin.Context, info *relaycommon.RelayInfo, resp *http.Response) (*types.NewAPIError, *dto.Usage) {
	usage := &dto.Usage{}
	helper.StreamScannerHandler(c, resp, info, func(data string, sr *helper.StreamResult) {
		var chunk GigaChatStreamChunk
		if err := common.Unmarshal([]byte(data), &chunk); err != nil {
			common.SysLog("gigachat: error unmarshalling stream chunk: " + err.Error())
			sr.Error(err)
			return
		}
		if chunk.Usage != nil {
			usage.PromptTokens = chunk.Usage.PromptTokens
			usage.CompletionTokens = chunk.Usage.CompletionTokens
			usage.TotalTokens = chunk.Usage.TotalTokens
		}
		response := gigaChatChunk2OpenAI(&chunk)
		if err := helper.ObjectData(c, response); err != nil {
			common.SysLog("gigachat: error writing stream chunk: " + err.Error())
			sr.Error(err)
		}
	})
	service.CloseResponseBodyGracefully(resp)
	return nil, usage
}

func gigaChatHandler(c *gin.Context, info *relaycommon.RelayInfo, resp *http.Response) (*types.NewAPIError, *dto.Usage) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return types.NewError(err, types.ErrorCodeBadResponseBody), nil
	}
	service.CloseResponseBodyGracefully(resp)

	var gcResp GigaChatResponse
	if err := common.Unmarshal(body, &gcResp); err != nil {
		return types.NewError(err, types.ErrorCodeBadResponseBody), nil
	}
	if len(gcResp.Choices) == 0 {
		return types.NewError(fmt.Errorf("gigachat: empty choices in response"), types.ErrorCodeBadResponseBody), nil
	}
	openAIResp := gigaChat2OpenAI(&gcResp)
	jsonResp, err := common.Marshal(openAIResp)
	if err != nil {
		return types.NewError(err, types.ErrorCodeBadResponseBody), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, _ = c.Writer.Write(jsonResp)
	return nil, &openAIResp.Usage
}
