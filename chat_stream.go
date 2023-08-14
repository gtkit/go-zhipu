package openai

import (
	"context"
	"fmt"
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
	Meta    Usage                        `json:"meta"`
}

type GlmChatCompletionStream struct {
	*streamReader[GlmChatCompletionStreamResponseResponse]
}

func (c *Client) CreateChatCompletionStream(ctx context.Context, request ChatCompletionRequest) (stream *GlmChatCompletionStream, err error) {
	urlSuffix := chatCompletionsSuffix

	fmt.Println("---c.fullURL(urlSuffix, request.Model)--", c.fullURL(urlSuffix, request.Model))

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
