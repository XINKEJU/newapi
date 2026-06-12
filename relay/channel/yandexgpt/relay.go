package yandexgpt

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relay/helper"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
)

// parseTokenCount converts a YandexGPT string token count to int.
func parseTokenCount(s string) int {
	if s == "" {
		return 0
	}
	n, _ := strconv.Atoi(s)
	return n
}

// yandexUsage converts YandexGPT usage to the standard dto.Usage struct.
func yandexUsage(u YandexUsage) *dto.Usage {
	prompt := parseTokenCount(u.InputTextTokens)
	completion := parseTokenCount(u.CompletionTokens)
	total := parseTokenCount(u.TotalTokens)
	if total == 0 {
		total = prompt + completion
	}
	return &dto.Usage{
		PromptTokens:     prompt,
		CompletionTokens: completion,
		TotalTokens:      total,
	}
}

// yandexResponse2OpenAI converts a synchronous YandexGPT response to the
// OpenAI chat completion format.
func yandexResponse2OpenAI(model string, resp *YandexChatResponse) *dto.OpenAITextResponse {
	choices := make([]dto.OpenAITextResponseChoice, 0, len(resp.Result.Alternatives))
	for i, alt := range resp.Result.Alternatives {
		finishReason := "stop"
		if alt.Status == "ALTERNATIVE_STATUS_TRUNCATED_FINAL" {
			finishReason = "length"
		}
		choices = append(choices, dto.OpenAITextResponseChoice{
			Index: i,
			Message: dto.Message{
				Role:    alt.Message.Role,
				Content: alt.Message.Text,
			},
			FinishReason: finishReason,
		})
	}
	return &dto.OpenAITextResponse{
		Object:  "chat.completion",
		Model:   model,
		Choices: choices,
		Usage:   *yandexUsage(resp.Result.Usage),
	}
}

// yandexStreamChunk2OpenAI converts a streaming chunk to the OpenAI SSE format.
func yandexStreamChunk2OpenAI(model string, chunk *YandexStreamChunk) *dto.ChatCompletionsStreamResponse {
	choices := make([]dto.ChatCompletionsStreamResponseChoice, 0, len(chunk.Result.Alternatives))
	for _, alt := range chunk.Result.Alternatives {
		var choice dto.ChatCompletionsStreamResponseChoice
		choice.Delta.SetContentString(alt.Message.Text)
		if alt.Status == "ALTERNATIVE_STATUS_FINAL" ||
			alt.Status == "ALTERNATIVE_STATUS_TRUNCATED_FINAL" {
			choice.FinishReason = &constant.FinishReasonStop
		}
		choices = append(choices, choice)
	}
	return &dto.ChatCompletionsStreamResponse{
		Object:  "chat.completion.chunk",
		Model:   model,
		Choices: choices,
	}
}

// yandexStreamHandler handles Server-Sent Events from the streaming endpoint.
func yandexStreamHandler(c *gin.Context, info *relaycommon.RelayInfo, resp *http.Response) (*types.NewAPIError, *dto.Usage) {
	usage := &dto.Usage{}
	helper.StreamScannerHandler(c, resp, info, func(data string, sr *helper.StreamResult) {
		var chunk YandexStreamChunk
		if err := common.Unmarshal([]byte(data), &chunk); err != nil {
			common.SysLog("yandexgpt: error unmarshalling stream chunk: " + err.Error())
			sr.Error(err)
			return
		}
		// The last chunk carries the final usage counters.
		if chunk.Result.Usage.TotalTokens != "" {
			usage = yandexUsage(chunk.Result.Usage)
		}
		response := yandexStreamChunk2OpenAI(info.UpstreamModelName, &chunk)
		if err := helper.ObjectData(c, response); err != nil {
			common.SysLog("yandexgpt: error writing stream chunk: " + err.Error())
			sr.Error(err)
		}
	})
	service.CloseResponseBodyGracefully(resp)
	return nil, usage
}

// yandexHandler handles a synchronous completion response.
func yandexHandler(c *gin.Context, info *relaycommon.RelayInfo, resp *http.Response) (*types.NewAPIError, *dto.Usage) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return types.NewError(err, types.ErrorCodeBadResponseBody), nil
	}
	service.CloseResponseBodyGracefully(resp)

	// Check for API-level error first.
	var apiErr YandexErrorResponse
	if json.Unmarshal(body, &apiErr) == nil && apiErr.Error.Message != "" {
		return types.NewError(
			fmt.Errorf("yandexgpt error %d: %s", apiErr.Error.GRPCStatusCode, apiErr.Error.Message),
			types.ErrorCodeBadResponseBody,
		), nil
	}

	var yResp YandexChatResponse
	if err := json.Unmarshal(body, &yResp); err != nil {
		return types.NewError(err, types.ErrorCodeBadResponseBody), nil
	}
	openAIResp := yandexResponse2OpenAI(info.UpstreamModelName, &yResp)
	jsonResp, err := json.Marshal(openAIResp)
	if err != nil {
		return types.NewError(err, types.ErrorCodeBadResponseBody), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, _ = c.Writer.Write(jsonResp)
	return nil, &openAIResp.Usage
}
