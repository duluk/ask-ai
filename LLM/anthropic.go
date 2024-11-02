package LLM

// Add support for the anthropic model, Claude Sonnet

import (
	"context"
	"errors"
	"fmt"

	"github.com/liushuangls/go-anthropic/v2"
)

func New_Anthropic(max_tokens int) *Anthropic {
	api_key := get_client_key("anthropic")
	client := anthropic.NewClient(api_key)

	return &Anthropic{API_Key: api_key, Tokens: max_tokens, Client: client}
}

func (cs *Anthropic) Chat(args Client_Args) error {
	prompt := args.Prompt
	log := args.Log
	client := cs.Client

	resp, err := client.CreateMessagesStream(context.Background(), anthropic.MessagesStreamRequest{
		MessagesRequest: anthropic.MessagesRequest{
			// TODO: figure out how to specify different anthropic models
			Model: anthropic.ModelClaudeInstant1Dot2,
			Messages: []anthropic.Message{
				anthropic.NewUserTextMessage(prompt),
			},
			MaxTokens: 1000,
		},
		OnContentBlockDelta: func(data anthropic.MessagesEventContentBlockDeltaData) {
			print(*data.Delta.Text)
		},
	})
	if err != nil {
		var e *anthropic.APIError
		if errors.As(err, &e) {
			fmt.Printf("Messages stream error, type: %s, message: %s", e.Type, e.Message)
		} else {
			fmt.Printf("Messages stream error: %v\n", err)
		}
		return err
	}
	println()
	log.WriteString("Assistant: " + resp.Content[0].GetText())
	log.WriteString("\n<------>\n")

	return nil
}
