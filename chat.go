package openai

import (
	"context"
	"net/http"
)

// Chat message role defined by the OpenAI API.
const (
	ChatMessageRoleSystem    = "system"
	ChatMessageRoleUser      = "user"
	ChatMessageRoleAssistant = "assistant"
)

const chatCompletionsSuffix = "/sse-invoke"

type ChatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type ChatCompletionPrompt struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest  请求模型参数.
type ChatCompletionRequest struct {
	Model       string                  `json:"model"`  // 模型
	Messages    []ChatCompletionMessage `json:"prompt"` // prompt
	Temperature float32                 `json:"temperature,omitempty"`
	TopP        float32                 `json:"top_p,omitempty"`
	// 智谱 SSE接口调用时，用于控制每次返回内容方式是增量还是全量，不提供此参数时默认为增量返回 - true 为增量返回 - false 为全量返回
	Incremental bool `json:"incremental"`
}
type FinishReason string

const (
	FinishReasonStop          FinishReason = "stop"
	FinishReasonLength        FinishReason = "length"
	FinishReasonFunctionCall  FinishReason = "function_call"
	FinishReasonContentFilter FinishReason = "content_filter"
	FinishReasonNull          FinishReason = "null"
)

type ChatCompletionChoice struct {
	Index   int                   `json:"index"`
	Message ChatCompletionMessage `json:"message"`
	// FinishReason
	// stop: API returned complete message,
	// or a message terminated by one of the stop sequences provided via the stop parameter
	// length: Incomplete model output due to max_tokens parameter or token limit
	// function_call: The model decided to call a function
	// content_filter: Omitted content due to a flag from our content filters
	// null: API response still in progress or incomplete
	FinishReason FinishReason `json:"finish_reason"`
}

type ChatCompletionResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []ChatCompletionChoice `json:"choices"`
	Usage   Usage                  `json:"usage"`
}

// GlmChatCompletionResponse Api文本返回.
type GlmChatCompletionResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		RequestID  string `json:"request_id"`
		TaskID     string `json:"task_id"`
		TaskStatus string `json:"task_status"`
		Choices    []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"choices"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
	} `json:"data"`
	Success bool `json:"success"`
}

// CreateChatCompletion — API call to Create a completion for the chat message.
func (c *Client) CreateChatCompletion(
	ctx context.Context,
	request ChatCompletionRequest,
) (response GlmChatCompletionResponse, err error) {
	urlSuffix := chatCompletionsSuffix

	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(urlSuffix, request.Model), withBody(request))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}
