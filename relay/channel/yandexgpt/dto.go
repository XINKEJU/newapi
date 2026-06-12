package yandexgpt

import "github.com/QuantumNous/new-api/dto"

// ──────────────────────────────────────────────────────────────────────────────
// YandexGPT REST API DTO
// Docs: https://yandex.cloud/en/docs/foundation-models/text-generation/api-ref/
// ──────────────────────────────────────────────────────────────────────────────

// YandexMessage is a single turn in the conversation.
type YandexMessage struct {
	Role string `json:"role"` // "user" | "assistant" | "system"
	Text string `json:"text"`
}

// YandexCompletionOptions contains sampling parameters.
type YandexCompletionOptions struct {
	Stream      bool    `json:"stream"`
	Temperature float64 `json:"temperature,omitempty"`
	MaxTokens   *int    `json:"maxTokens,omitempty"`
}

// YandexChatRequest is the body sent to /completionAsync or /completion.
type YandexChatRequest struct {
	ModelURI          string                   `json:"modelUri"`
	CompletionOptions YandexCompletionOptions  `json:"completionOptions"`
	Messages          []YandexMessage          `json:"messages"`
}

// ──────────────────────────────────────────────────────────────────────────────
// Response DTOs
// ──────────────────────────────────────────────────────────────────────────────

type YandexAlternative struct {
	Message YandexMessage `json:"message"`
	Status  string        `json:"status"` // ALTERNATIVE_STATUS_FINAL, etc.
}

type YandexUsage struct {
	InputTextTokens  string `json:"inputTextTokens"`
	CompletionTokens string `json:"completionTokens"`
	TotalTokens      string `json:"totalTokens"`
}

type YandexResult struct {
	Alternatives []YandexAlternative `json:"alternatives"`
	Usage        YandexUsage         `json:"usage"`
	ModelVersion string              `json:"modelVersion"`
}

// YandexChatResponse wraps the synchronous completion endpoint.
type YandexChatResponse struct {
	Result YandexResult `json:"result"`
}

// YandexStreamChunk is a single SSE payload from the streaming endpoint.
type YandexStreamChunk struct {
	Result YandexResult `json:"result"`
}

// YandexErrorResponse is returned when the API signals an error.
type YandexErrorResponse struct {
	Error struct {
		GRPCStatusCode int    `json:"grpcStatusCode"`
		Message        string `json:"message"`
	} `json:"error"`
}

// ──────────────────────────────────────────────────────────────────────────────
// Conversion helpers
// ──────────────────────────────────────────────────────────────────────────────

func openAI2Yandex(req dto.GeneralOpenAIRequest, modelURI string, isStream bool) *YandexChatRequest {
	yReq := &YandexChatRequest{
		ModelURI: modelURI,
		CompletionOptions: YandexCompletionOptions{
			Stream: isStream,
		},
	}
	if req.Temperature != nil {
		temp := *req.Temperature
		// YandexGPT temperature is [0, 1]; clamp just in case
		if temp > 1.0 {
			temp = 1.0
		}
		yReq.CompletionOptions.Temperature = temp
	}
	if maxTok := req.GetMaxTokens(); maxTok > 0 {
		mt := int(maxTok)
		yReq.CompletionOptions.MaxTokens = &mt
	}
	for _, m := range req.Messages {
		role := m.Role
		if role == "system" {
			role = "system"
		}
		yReq.Messages = append(yReq.Messages, YandexMessage{
			Role: role,
			Text: m.StringContent(),
		})
	}
	return yReq
}
