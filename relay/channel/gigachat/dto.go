package gigachat

import "github.com/QuantumNous/new-api/dto"

// ──────────────────────────────────────────────────────────────────────────────
// GigaChat REST API DTO
// Docs: https://developers.sber.ru/docs/ru/gigachat/api/reference/rest/post-chat
// ──────────────────────────────────────────────────────────────────────────────

// GigaChatMessage maps directly to the GigaChat message format.
// Role values: "system" | "user" | "assistant"
type GigaChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// GigaChatRequest is the body sent to /chat/completions.
type GigaChatRequest struct {
	Model             string            `json:"model"`
	Messages          []GigaChatMessage `json:"messages"`
	Stream            bool              `json:"stream,omitempty"`
	Temperature       float64           `json:"temperature,omitempty"`
	TopP              float64           `json:"top_p,omitempty"`
	MaxTokens         *int              `json:"max_tokens,omitempty"`
	RepetitionPenalty float64           `json:"repetition_penalty,omitempty"`
	UpdateInterval    float64           `json:"update_interval,omitempty"`
}

// ──────────────────────────────────────────────────────────────────────────────
// Response DTOs
// ──────────────────────────────────────────────────────────────────────────────

type GigaChatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type GigaChatChoice struct {
	Message      GigaChatMessage `json:"message"`
	Delta        GigaChatMessage `json:"delta"`
	Index        int             `json:"index"`
	FinishReason string          `json:"finish_reason"` // "stop" | "length" | "function_call"
}

type GigaChatResponse struct {
	Choices []GigaChatChoice `json:"choices"`
	Created int64            `json:"created"`
	Model   string           `json:"model"`
	Usage   GigaChatUsage    `json:"usage"`
	Object  string           `json:"object"` // "chat.completion"
}

// GigaChatStreamChunk is a single data event from the SSE stream.
type GigaChatStreamChunk struct {
	Choices []GigaChatChoice `json:"choices"`
	Created int64            `json:"created"`
	Model   string           `json:"model"`
	Object  string           `json:"object"` // "chat.completion.chunk"
	Usage   *GigaChatUsage   `json:"usage,omitempty"`
}

// ──────────────────────────────────────────────────────────────────────────────
// OAuth2 token response
// ──────────────────────────────────────────────────────────────────────────────

type GigaChatTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresAt   int64  `json:"expires_at"` // Unix ms
}

// ──────────────────────────────────────────────────────────────────────────────
// Conversion helpers
// ──────────────────────────────────────────────────────────────────────────────

func openAI2GigaChat(req dto.GeneralOpenAIRequest, isStream bool) *GigaChatRequest {
	gcReq := &GigaChatRequest{
		Model:  req.Model,
		Stream: isStream,
	}
	if req.Temperature != nil {
		gcReq.Temperature = *req.Temperature
	}
	if req.TopP != nil {
		gcReq.TopP = *req.TopP
	}
	if req.FrequencyPenalty != nil {
		gcReq.RepetitionPenalty = *req.FrequencyPenalty
	}
	if maxTok := req.GetMaxTokens(); maxTok > 0 {
		mt := int(maxTok)
		gcReq.MaxTokens = &mt
	}
	for _, m := range req.Messages {
		gcReq.Messages = append(gcReq.Messages, GigaChatMessage{
			Role:    m.Role,
			Content: m.StringContent(),
		})
	}
	return gcReq
}
