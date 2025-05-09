package LLM

import (
	"context"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
)

func NewOpenAI(apiLLC string, apiURL string) *OpenAI {
	apiKey, err := getClientKey(apiLLC)
	if err != nil {
		panic(err)
	}
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(apiURL),
	)

	return &OpenAI{APIKey: apiKey, Client: &client}
}

func (cs *OpenAI) Chat(args ClientArgs, termWidth int, tabWidth int) (ClientResponse, <-chan StreamResponse, error) {
	// Create a channel for streaming responses
	responseChan := make(chan StreamResponse)

	// Start a goroutine to handle streaming
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

	// Return an empty ClientResponse since the actual response will be streamed
	// The full response can be collected by the caller if needed
	return resp, responseChan, nil
}

func (cs *OpenAI) ChatStream(args ClientArgs, termWidth int, tabWidth int, stream chan<- StreamResponse) (ClientResponse, error) {
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

	model := openai.ChatModel(*args.Model)

	// myInputEstimate := EstimateTokens(msgCtx + *args.Prompt + *args.SystemPrompt)
	ctx := context.Background()
	openaiStream := client.Chat.Completions.NewStreaming(
		ctx,
		openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(*args.SystemPrompt),
				openai.AssistantMessage(msgCtx),
				openai.UserMessage(*args.Prompt),
			},
			Model:               model, // Directly use the model string or value
			MaxCompletionTokens: openai.Int(int64(*args.MaxTokens)),
			Temperature:         openai.Float(float64(*args.Temperature)), // Controls randomness (0.0 to 2.0)
			StreamOptions: openai.ChatCompletionStreamOptionsParam{
				IncludeUsage: openai.Bool(true),
			},
			ReasoningEffort: shared.ReasoningEffort(*args.Thinking),
			TopP:            openai.Float(1.0), // Controls diversity via nucleus sampling
			N:               openai.Int(1),     // Number of completions to generate
			// These are not correct but possible I think:
			// ResponseFormat:   openai.ChatResponseFormatDefault, // Format of the response
			// Stop:             openai.String("\n"),              // Stop completion at this token
		},
	)

	// Process the stream in chunks
	for openaiStream.Next() {
		evt := openaiStream.Current()
		if len(evt.Choices) > 0 {
			data := evt.Choices[0].Delta.Content

			// Only send non-empty content
			if data != "" {
				// // Write to terminal if not disabled
				// if !args.DisableOutput {
				// 	if _, err := wrapper.Write([]byte(data)); err != nil {
				// 		err := fmt.Errorf("error writing to wrapper: %s", err)
				// 		stream <- StreamResponse{
				// 			Content: "",
				// 			Done:    true,
				// 			Error:   err,
				// 		}
				// 		return err
				// 	}
				// }

				// Send data to the stream channel
				stream <- StreamResponse{
					Content: data,
					Done:    false,
					Error:   nil,
				}
			}
		}
	}

	// Check for errors
	if openaiStream.Err() != nil {
		err := openaiStream.Err()
		fmt.Printf("Error: %s\n", err)
		stream <- StreamResponse{
			Content: "",
			Done:    true,
			Error:   err,
		}
		return ClientResponse{}, err
	}

	// Signal completion
	stream <- StreamResponse{
		Content: "",
		Done:    true,
		Error:   nil,
	}

	// TODO: populate this?
	return ClientResponse{}, nil
}
