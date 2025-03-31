package LLM

// Add support for the anthropic model, Claude Sonnet

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/liushuangls/go-anthropic/v2"

	"github.com/duluk/ask-ai/pkg/logger"
)

func convertToAnthropicMessages(chatHist []LLMConversations) []anthropic.Message {
	anthropicMsgs := make([]anthropic.Message, len(chatHist))

	for i, msg := range chatHist {
		logger.Debug("Anthropic LLMConversations", "msg", msg)
		var role anthropic.ChatRole

		switch strings.ToLower(msg.Role) {
		case "user":
			role = anthropic.RoleUser
		case "assistant":
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

func (cs *Anthropic) Chat(args ClientArgs, termWidth int, tabWidth int) (ClientResponse, <-chan StreamResponse, error) {
	responseChan := make(chan StreamResponse)

	var resp ClientResponse
	var err error
	go func() {
		defer close(responseChan)

		// Use the streaming implementation
		resp, err = cs.ChatStream(args, termWidth, tabWidth, responseChan)
		if err != nil {
			responseChan <- StreamResponse{
				Content: "",
				Done:    true,
				Error:   err,
			}
		}
	}()

	return resp, responseChan, nil
}

func (cs *Anthropic) ChatStream(args ClientArgs, termWidth int, tabWidth int, stream chan<- StreamResponse) (ClientResponse, error) {
	prompt := args.Prompt
	client := cs.Client

	logger.Debug("Anthropic context before conversion", "args.Context", args.Context)
	msgCtx := convertToAnthropicMessages(args.Context)
	msgCtx = append(msgCtx, anthropic.NewUserTextMessage(*prompt))
	logger.Debug("Anthropic context after conversion", "context", msgCtx)

	myInputEstimate := EstimateTokens(*args.Prompt + *args.SystemPrompt)

	// wrapper := linewrap.NewLineWrapper(termWidth, tabWidth, os.Stdout)

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
			// OnContentBlockDelta: func(data anthropic.MessagesEventContentBlockDeltaData) {
			// 	wrapper.Write([]byte(*data.Delta.Text))
			// },
			OnContentBlockDelta: func(data anthropic.MessagesEventContentBlockDeltaData) {
				stream <- StreamResponse{
					Content: *data.Delta.Text,
					Done:    false,
					Error:   nil,
				}
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
