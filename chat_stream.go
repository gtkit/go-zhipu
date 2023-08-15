package openai

import (
	"context"
	"net/http"
)

type ChatCompletionStreamChoiceDelta struct {
	Content string `json:"content"`
}

type ChatCompletionStreamChoice struct {
	Delta ChatCompletionStreamChoiceDelta `json:"delta"`
}

type GlmChatCompletionStreamResponseResponse struct {
	ID      string                       `json:"id"`
	Event   string                       `json:"event"`
	Data    string                       `json:"data"`
	Choices []ChatCompletionStreamChoice `json:"choices"`
	Meta    GlmMeta                      `json:"meta"`
}

type GlmMeta struct {
	TaskStatus string `json:"task_status"`
	Usage      struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
	TaskID    string `json:"task_id"`
	RequestID string `json:"request_id"`
}

type GlmChatCompletionStream struct {
	*streamReader[GlmChatCompletionStreamResponseResponse]
}

func (c *Client) CreateChatCompletionStream(
	ctx context.Context,
	request ChatCompletionRequest,
) (stream *GlmChatCompletionStream, err error) {
	urlSuffix := chatCompletionsSuffix

	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(urlSuffix, request.Model), withBody(request))
	if err != nil {
		return nil, err
	}

	resp, err := sendRequestStream[GlmChatCompletionStreamResponseResponse](c, req)
	if err != nil {
		return
	}
	stream = &GlmChatCompletionStream{
		streamReader: resp,
	}
	return
}
