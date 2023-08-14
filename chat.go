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

// const chatCompletionsSuffix = "/chat/completions"
const chatCompletionsSuffix = "/sse-invoke"

type ChatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type ChatCompletionPrompt struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest represents a request structure for chat completion API. 请求模型参数
type ChatCompletionRequest struct {
	Model    string                  `json:"model"`    // 模型
	Messages []ChatCompletionMessage `json:"messages"` // prompt
	Prompt   []ChatCompletionPrompt  `json:"prompt"`
	// MaxTokens        int                     `json:"max_tokens,omitempty"`
	Temperature float32 `json:"temperature,omitempty"`
	TopP        float32 `json:"top_p,omitempty"`
	// 智谱 SSE接口调用时，用于控制每次返回内容方式是增量还是全量，不提供此参数时默认为增量返回 - true 为增量返回 - false 为全量返回
	Incremental bool `json:"incremental"`
}

// GlmChatCompletionResponse Api文本返回
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
func (c *Client) CreateChatCompletion(ctx context.Context, request ChatCompletionRequest) (response GlmChatCompletionResponse, err error) {
	urlSuffix := chatCompletionsSuffix

	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(urlSuffix, request.Model), withBody(request))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}
