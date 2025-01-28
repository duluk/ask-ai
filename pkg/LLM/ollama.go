package LLM

import (
	"os"
	"strings"

	"github.com/duluk/ask-ai/pkg/linewrap"
	"github.com/duluk/ask-ai/pkg/ollama"
)

func NewOllama() *Ollama {
	apiKey := ""
	client := ollama.NewClient(apiKey)

	return &Ollama{APIKey: apiKey, Client: client}
}

func (cs *Ollama) Chat(args ClientArgs, termWidth int, tabWidth int) (ClientResponse, error) {
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

	const OllamaModelDeepseekR1 = "deepseek-r1:8b"

	myInputEstimate := EstimateTokens(msgCtx + *args.Prompt + *args.SystemPrompt)
	req := ollama.ChatCompletionRequest{
		Model: OllamaModelDeepseekR1,
		Messages: []ollama.Message{
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

	resp, err := client.ChatCompletion(req)
	if err != nil {
		return ClientResponse{}, err
	}

	respText := resp.Choices[0].Message.Content
	wrapper := linewrap.NewLineWrapper(termWidth, tabWidth, os.Stdout)

	if _, err := wrapper.Write([]byte(respText)); err != nil {
		return ClientResponse{}, err
	}

	return ClientResponse{
		Text:         respText,
		InputTokens:  myInputEstimate,
		OutputTokens: int32(len(respText)),
	}, nil
}
