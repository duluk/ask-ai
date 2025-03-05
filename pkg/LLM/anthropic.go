package LLM

// Add support for the anthropic model, Claude Sonnet

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/duluk/ask-ai/pkg/linewrap"

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

func NewAnthropic() *Anthropic {
	api_key := getClientKey("anthropic")
	client := anthropic.NewClient(api_key)

	return &Anthropic{APIKey: api_key, Client: client}
}

func (cs *Anthropic) Chat(args ClientArgs, termWidth int, tabWidth int) (ClientResponse, error) {
	prompt := args.Prompt
	client := cs.Client

	msgCtx := convertToAnthropicMessages(args.Context)
	msgCtx = append(msgCtx, anthropic.NewUserTextMessage(*prompt))

	myInputEstimate := EstimateTokens(*args.Prompt + *args.SystemPrompt)

	wrapper := linewrap.NewLineWrapper(termWidth, tabWidth, os.Stdout)
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
				wrapper.Write([]byte(*data.Delta.Text))
			},
		})
	if err != nil {
		var e *anthropic.APIError
		if errors.As(err, &e) {
			fmt.Printf("Messages stream error, type: %s, message: %s", e.Type, e.Message)
		} else {
			fmt.Printf("Messages stream error: %v\n", err)
		}
		return ClientResponse{}, err
	}

	// I believe the stats object will be usable even if the response is empty
	stats := resp.Usage
	r := ClientResponse{
		Text:         resp.Content[0].GetText(),
		InputTokens:  int32(stats.InputTokens),
		OutputTokens: int32(stats.OutputTokens),
		MyEstInput:   myInputEstimate,
	}
	return r, nil
}

// StreamChat implements streaming response for the Anthropic client
func (cs *Anthropic) StreamChat(args ClientArgs, termWidth int, tabWidth int, callback func(chunk string)) (ClientResponse, error) {
	prompt := args.Prompt
	client := cs.Client

	msgCtx := convertToAnthropicMessages(args.Context)
	msgCtx = append(msgCtx, anthropic.NewUserTextMessage(*prompt))

	myInputEstimate := EstimateTokens(*args.Prompt + *args.SystemPrompt)

	var fullResponse string

	// Use the context from args if provided, otherwise use background context
	ctx := context.Background()
	if args.Ctx != nil {
		ctx = args.Ctx
	}

	// Use stream API and collect chunks through callback
	resp, err := client.CreateMessagesStream(
		ctx,
		anthropic.MessagesStreamRequest{
			MessagesRequest: anthropic.MessagesRequest{
				Model:       anthropic.ModelClaude3Dot5Haiku20241022,
				Messages:    msgCtx,
				MaxTokens:   *args.MaxTokens,
				Temperature: args.Temperature,
				System:      *args.SystemPrompt,
			},
			// Stream the response as it comes in, sending each chunk to the callback
			OnContentBlockDelta: func(data anthropic.MessagesEventContentBlockDeltaData) {
				if data.Delta.Text != nil {
					chunk := *data.Delta.Text
					fullResponse += chunk
					callback(chunk)
				}
			},
		})
	if err != nil {
		var e *anthropic.APIError
		if errors.As(err, &e) {
			return ClientResponse{}, fmt.Errorf("messages stream error, type: %s, message: %s", e.Type, e.Message)
		} else {
			return ClientResponse{}, fmt.Errorf("messages stream error: %v", err)
		}
	}

	// Return complete response with stats
	stats := resp.Usage
	r := ClientResponse{
		Text:         fullResponse,
		InputTokens:  int32(stats.InputTokens),
		OutputTokens: int32(stats.OutputTokens),
		MyEstInput:   myInputEstimate,
	}
	return r, nil
}
