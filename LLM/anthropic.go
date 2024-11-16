package LLM

// Add support for the anthropic model, Claude Sonnet

import (
	"context"
	"errors"
	"fmt"

	"github.com/liushuangls/go-anthropic/v2"
)

func convertToAnthropicMessages(chatHist []LLMConversations) []anthropic.Message {
	anthropicMsgs := make([]anthropic.Message, len(chatHist))

	for i, msg := range chatHist {
		var role anthropic.ChatRole

		switch msg.Role {
		case "User":
			role = anthropic.RoleUser
		case "Assistant":
			role = anthropic.RoleAssistant
		}

		content := []anthropic.MessageContent{
			{
				Type: "text",
				Text: &msg.Content,
			},
		}
		anthropicMsgs[i] = anthropic.Message{
			Role:    role,
			Content: content,
		}
	}

	return anthropicMsgs
}

func NewAnthropic(maxTokens int) *Anthropic {
	api_key := getClientKey("anthropic")
	client := anthropic.NewClient(api_key)

	return &Anthropic{APIKey: api_key, Tokens: maxTokens, Client: client}
}

func (cs *Anthropic) Chat(args ClientArgs) (string, error) {
	prompt := args.Prompt
	client := cs.Client

	msgCtx := convertToAnthropicMessages(args.Context)
	msgCtx = append(msgCtx, anthropic.NewUserTextMessage(*prompt))
	resp, err := client.CreateMessagesStream(
		context.Background(),
		anthropic.MessagesStreamRequest{
			MessagesRequest: anthropic.MessagesRequest{
				// TODO: figure out how to specify different anthropic models
				// Model: anthropic.ModelClaude3Dot5Sonnet20241022,
				Model:       anthropic.ModelClaude3Dot5Haiku20241022,
				Messages:    msgCtx,
				MaxTokens:   *args.MaxTokens,
				Temperature: args.Temperature,
				System:      *args.SystemPrompt,
				// TopP:        1.0,
				// TopK:        40,
			},
			// Print the response as it comes in, as a streaming chat...
			OnContentBlockDelta: func(data anthropic.MessagesEventContentBlockDeltaData) {
				fmt.Printf(*data.Delta.Text)
			},
		})
	if err != nil {
		var e *anthropic.APIError
		if errors.As(err, &e) {
			fmt.Printf("Messages stream error, type: %s, message: %s", e.Type, e.Message)
		} else {
			fmt.Printf("Messages stream error: %v\n", err)
		}
		return "", err
	}

	return resp.Content[0].GetText(), nil
}
