package LLM

import (
	"context"
	"strings"

	"github.com/duluk/ask-ai/pkg/deepseek"
)

func NewDeepSeek() *DeepSeek {
	apiKey, err := getClientKey("deepseek")
	if err != nil {
		panic(err)
	}
	client := deepseek.NewClient(apiKey)

	return &DeepSeek{APIKey: apiKey, Client: client}
}

func (cs *DeepSeek) Chat(args ClientArgs, termWidth int, tabWidth int) (ClientResponse, error) {
	client := cs.Client

	var msgCtx string

	for _, msg := range args.Context {
		msg.Role = strings.ToLower(msg.Role)
		switch msg.Role {
		case "user":
			msgCtx += "User: " + msg.Content + "\n"
		case "assistant":
			msgCtx += "Assistant: " + msg.Content + "\n"
		}
	}

	const ChatModelDeepSeekChat = "deepseek-chat"
	const ChatModelDeepSeekReasoner = "deepseek-reasoner"

	myInputEstimate := EstimateTokens(msgCtx + *args.Prompt + *args.SystemPrompt)
	req := deepseek.ChatCompletionRequest{
		Model: ChatModelDeepSeekChat,
		Messages: []deepseek.Message{
			{
				Role:    "system",
				Content: *args.SystemPrompt,
			},
			{
				Role:    "assistant",
				Content: msgCtx,
			},
			{
				Role:    "user",
				Content: *args.Prompt,
			},
		},
		MaxTokens:   *args.MaxTokens,
		Temperature: float64(*args.Temperature),
		// Stream:      true,
	}

	ctx := context.Background()
	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		// Handle error
	}

	r := ClientResponse{
		// Text:         "NOT REAL",
		Text:         resp.Choices[0].Message.Content,
		InputTokens:  0,
		OutputTokens: 0,
		MyEstInput:   myInputEstimate,
	}

	return r, nil
}
