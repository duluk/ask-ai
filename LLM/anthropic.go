package LLM

// Add support for the anthropic model, Claude Sonnet

import (
	"context"
	"errors"
	"fmt"

	"github.com/liushuangls/go-anthropic/v2"
)

func convert_to_anthropic_messages(chat_hist []LLM_Conversations) []anthropic.Message {
	anthropicMsgs := make([]anthropic.Message, len(chat_hist))

	for i, msg := range chat_hist {
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

func New_Anthropic(max_tokens int) *Anthropic {
	api_key := get_client_key("anthropic")
	client := anthropic.NewClient(api_key)

	return &Anthropic{API_Key: api_key, Tokens: max_tokens, Client: client}
}

func (cs *Anthropic) Chat(args Client_Args) (string, error) {
	prompt := args.Prompt
	client := cs.Client

	msg_ctx := convert_to_anthropic_messages(args.Context)
	msg_ctx = append(msg_ctx, anthropic.NewUserTextMessage(prompt))
	resp, err := client.CreateMessagesStream(
		context.Background(),
		anthropic.MessagesStreamRequest{
			MessagesRequest: anthropic.MessagesRequest{
				// TODO: figure out how to specify different anthropic models
				// Model: anthropic.ModelClaude3Dot5Sonnet20241022,
				Model:     anthropic.ModelClaude3Dot5Haiku20241022,
				Messages:  msg_ctx,
				MaxTokens: args.Max_Tokens,
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
