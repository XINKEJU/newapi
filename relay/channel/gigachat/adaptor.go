package gigachat

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/relay/channel"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
)

// Adaptor implements channel.Adaptor for Sber GigaChat.
// API reference: https://developers.sber.ru/docs/ru/gigachat/api/reference/rest/
//
// Authentication:
//   The channel key must be the Base64-encoded "ClientID:ClientSecret" string
//   obtained from the Sber developer portal. The adaptor fetches and caches an
//   OAuth2 Bearer token automatically.
//
// Base URL: https://gigachat.devices.sberbank.ru/api/v1
type Adaptor struct{}

func (a *Adaptor) Init(_ *relaycommon.RelayInfo) {}

// GetRequestURL returns the GigaChat chat completions endpoint.
func (a *Adaptor) GetRequestURL(info *relaycommon.RelayInfo) (string, error) {
	base := strings.TrimRight(info.ChannelBaseUrl, "/")
	return base + "/chat/completions", nil
}

// SetupRequestHeader fetches an OAuth token and sets Authorization + Content-Type.
func (a *Adaptor) SetupRequestHeader(_ *gin.Context, req *http.Header, info *relaycommon.RelayInfo) error {
	channel.SetupApiRequestHeader(info, nil, req)
	token, err := getAccessToken(info.ApiKey)
	if err != nil {
		return err
	}
	req.Set("Authorization", "Bearer "+token)
	req.Set("Content-Type", "application/json")
	return nil
}

// ConvertOpenAIRequest converts a standard OpenAI chat request to GigaChat format.
func (a *Adaptor) ConvertOpenAIRequest(_ *gin.Context, info *relaycommon.RelayInfo, request *dto.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	return openAI2GigaChat(*request, info.IsStream), nil
}

func (a *Adaptor) ConvertRerankRequest(_ *gin.Context, _ int, _ dto.RerankRequest) (any, error) {
	return nil, errors.New("not supported")
}

func (a *Adaptor) ConvertEmbeddingRequest(_ *gin.Context, _ *relaycommon.RelayInfo, _ dto.EmbeddingRequest) (any, error) {
	return nil, errors.New("not supported")
}

func (a *Adaptor) ConvertAudioRequest(_ *gin.Context, _ *relaycommon.RelayInfo, _ dto.AudioRequest) (io.Reader, error) {
	return nil, errors.New("not supported")
}

func (a *Adaptor) ConvertImageRequest(_ *gin.Context, _ *relaycommon.RelayInfo, _ dto.ImageRequest) (any, error) {
	return nil, errors.New("not supported")
}

func (a *Adaptor) ConvertOpenAIResponsesRequest(_ *gin.Context, _ *relaycommon.RelayInfo, _ dto.OpenAIResponsesRequest) (any, error) {
	return nil, errors.New("not supported")
}

func (a *Adaptor) ConvertClaudeRequest(_ *gin.Context, _ *relaycommon.RelayInfo, _ *dto.ClaudeRequest) (any, error) {
	return nil, errors.New("not supported")
}

func (a *Adaptor) ConvertGeminiRequest(_ *gin.Context, _ *relaycommon.RelayInfo, _ *dto.GeminiChatRequest) (any, error) {
	return nil, errors.New("not supported")
}

func (a *Adaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (any, error) {
	return channel.DoApiRequest(a, c, info, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (usage any, err *types.NewAPIError) {
	if info.IsStream {
		apiErr, u := gigaChatStreamHandler(c, info, resp)
		return u, apiErr
	}
	apiErr, u := gigaChatHandler(c, info, resp)
	return u, apiErr
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return ChannelName
}
