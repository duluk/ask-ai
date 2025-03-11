package LLM

import (
	"strings"

	"github.com/duluk/ask-ai/pkg/ollama"
)

func NewOllama() *Ollama {
	apiKey := ""
	client := ollama.NewClient(apiKey)

	return &Ollama{APIKey: apiKey, Client: client}
}

func (cs *Ollama) Chat(args ClientArgs, termWidth int, tabWidth int) (ClientResponse, <-chan StreamResponse, error) {
	// For non-streaming requests, we'll collect the full response
	responseChan := make(chan StreamResponse)

	// Start a goroutine to collect the full response
	// go func() {
	// 	for resp := range responseChan {
	// 		if resp.Error != nil {
	// 			// Handle error
	// 			break
	// 		}
	// 		fullResponse += resp.Content
	// 	}
	// }()

	go func() {
		defer close(responseChan)

		// Use the streaming implementation
		err := cs.ChatStream(args, termWidth, tabWidth, responseChan)
		if err != nil {
			responseChan <- StreamResponse{
				Error: err,
			}
		}
	}()

	return ClientResponse{}, responseChan, nil

	// // Return the full response
	// // Estimate tokens since we don't have exact counts in this implementation
	// inputTokens := EstimateTokens(*args.Prompt)
	// outputTokens := EstimateTokens(fullResponse)
	//
	// return ClientResponse{
	// 	Text:         fullResponse,
	// 	InputTokens:  inputTokens,
	// 	OutputTokens: outputTokens,
	// }, nil
}

// Add this method to the Ollama struct
func (cs *Ollama) ChatStream(args ClientArgs, termWidth int, tabWidth int, stream chan<- StreamResponse) error {
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

	const minTokens = 32768
	const OllamaModelGemma2 = "gemma2:9b"
	const OllamaModelDeepseekR1_8b = "deepseek-r1:8b"
	const OllamaModelDeepseekR1_14b = "deepseek-r1:14b"

	myInputEstimate := EstimateTokens(msgCtx + *args.Prompt + *args.SystemPrompt)
	adjustedMaxTokens := int(myInputEstimate + int32(*args.MaxTokens))

	req := ollama.ChatCompletionRequest{
		Model: OllamaModelGemma2, // or whatever model you're using
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
		MaxTokens:   max(adjustedMaxTokens, minTokens),
		Temperature: float64(*args.Temperature),
		Stream:      true,
	}

	// Use the streaming API from ollama client
	err := client.ChatCompletionStream(req, func(chunk ollama.ChatCompletionChunk) {
		if len(chunk.Choices) > 0 {
			content := chunk.Choices[0].Delta.Content

			// Only send non-empty content
			if content != "" {
				stream <- StreamResponse{
					Content: content,
					Done:    false,
					Error:   nil,
				}
			}
		}
	})
	if err != nil {
		stream <- StreamResponse{
			Content: "",
			Done:    true,
			Error:   err,
		}
		return err
	}

	// Signal completion
	stream <- StreamResponse{
		Content: "",
		Done:    true,
		Error:   nil,
	}

	return nil
}
