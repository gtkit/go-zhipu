// @Author xiaozhaofu 2023/5/26 19:13:00
package test_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	zpai "github.com/gtkit/zhipuAi"
)

func TestZpChat(t *testing.T) {
	key := "797e0338809fdac6080dcdc72da95dbc.8H3t2c43RI4bhjPO"

	token, err := zpai.GetAPIToken(key, time.Hour*24)

	if err != nil {
		t.Log("---GenerateToken err:", err)
	}

	prompt := []zpai.ChatCompletionMessage{
		{Role: zpai.ChatMessageRoleUser, Content: "1=1等于多少"},
	}

	openConfig := zpai.DefaultConfig(token)

	openConfig.HTTPClient = &http.Client{
		Timeout: 180 * time.Second,
	}

	// 实例化一个客户端
	c := zpai.NewClientWithConfig(openConfig)

	req := zpai.ChatCompletionRequest{
		Model:       zpai.GLMPro,
		Messages:    prompt,
		Temperature: 0.7,
		Incremental: true,
	}
	jr, _ := json.Marshal(req)
	t.Log("---req:", string(jr))

	stream, err := c.CreateChatCompletionStream(context.Background(), req)
	if err != nil {
		t.Logf("ChatCompletionStream error: %v\n", err)

		return
	}
	defer stream.Close()

	for {
		response, reserr := stream.Recv()

		if errors.Is(reserr, io.EOF) {
			t.Log("\nStream finished")
			return
		}

		if reserr != nil {
			t.Logf("\nStream error: %v\n", reserr)
			return
		}

		var s string
		for _, choice := range response.Choices {
			s += choice.Delta.Content
		}
		// t.Log("---- response.Choices Content: ", s)
	}
}
