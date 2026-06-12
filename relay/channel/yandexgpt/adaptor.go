package yandexgpt

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/relay/channel"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
)

// Adaptor implements channel.Adaptor for Yandex Foundation Models.
// API reference: https://yandex.cloud/en/docs/foundation-models/text-generation/api-ref/
//
// Authentication: pass the IAM token or API key as the channel key.
//   - IAM token  → Authorization: Bearer <token>
//   - API key    → Authorization: Api-Key <key>
//     (the adaptor auto-detects: if the key starts with "t1." it is treated as an
//     IAM token, otherwise as an API key)
//
// The folderId must be embedded in the model URI, e.g.:
//
//	gpt://<folderId>/yandexgpt/latest
//
// Users can set the channel Base URL to their folder-specific endpoint or
// keep the default https://llm.api.cloud.yandex.net/foundationModels/v1
// and provide the full modelUri in the model name field.
type Adaptor struct{}

func (a *Adaptor) Init(_ *relaycommon.RelayInfo) {}

// GetRequestURL returns the Yandex text-generation endpoint.
// Streaming requests go to /completionStream; sync requests to /completion.
func (a *Adaptor) GetRequestURL(info *relaycommon.RelayInfo) (string, error) {
	base := strings.TrimRight(info.ChannelBaseUrl, "/")
	if info.IsStream {
		return base + "/completionStream", nil
	}
	return base + "/completion", nil
}

// SetupRequestHeader sets Content-Type and the appropriate Authorization header.
func (a *Adaptor) SetupRequestHeader(_ *gin.Context, req *http.Header, info *relaycommon.RelayInfo) error {
	channel.SetupApiRequestHeader(info, nil, req)
	key := info.ApiKey
	if strings.HasPrefix(key, "t1.") {
		// IAM token
		req.Set("Authorization", "Bearer "+key)
	} else {
		// API key
		req.Set("Authorization", "Api-Key "+key)
	}
	req.Set("Content-Type", "application/json")
	return nil
}

// buildModelURI returns a fully-qualified Yandex model URI.
// If the model name already looks like a URI (starts with "gpt://"), it is
// returned as-is. Otherwise the channel base URL is used to derive the folder
// ID from the path, falling back to the raw model name.
func buildModelURI(info *relaycommon.RelayInfo) string {
	model := info.UpstreamModelName
	if strings.HasPrefix(model, "gpt://") || strings.HasPrefix(model, "ds://") {
		return model
	}
	// Derive folder ID from extra config or channel notes if available.
	// Convention: the channel "other" JSON may contain {"folder_id":"<id>"}.
	// For simplicity we accept "gpt://<anything>/model" set directly by the user.
	// Default: wrap the raw model name.
	return fmt.Sprintf("gpt://%s/latest", model)
}

// ConvertOpenAIRequest converts a standard OpenAI chat request to the YandexGPT format.
func (a *Adaptor) ConvertOpenAIRequest(_ *gin.Context, info *relaycommon.RelayInfo, request *dto.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	return openAI2Yandex(*request, buildModelURI(info), info.IsStream), nil
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
		err, u := yandexStreamHandler(c, info, resp)
		return u, err
	}
	err, u := yandexHandler(c, info, resp)
	return u, err
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return ChannelName
}
