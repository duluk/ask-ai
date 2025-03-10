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

	const minTokens = 32768
	const OllamaModelGemma2 = "gemma2:9b"
	const OllamaModelDeepseekR1_8b = "deepseek-r1:8b"
	const OllamaModelDeepseekR1_14b = "deepseek-r1:14b"

	myInputEstimate := EstimateTokens(msgCtx + *args.Prompt + *args.SystemPrompt)
	adjustedMaxTokens := int(myInputEstimate + int32(*args.MaxTokens))
	req := ollama.ChatCompletionRequest{
		Model: OllamaModelGemma2,
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
		// Since this is a local model, let's give it some room to cook
		MaxTokens:   max(adjustedMaxTokens, minTokens),
		Temperature: float64(*args.Temperature),
	}

	resp, err := client.ChatCompletion(req)
	if err != nil {
		return ClientResponse{}, err
	}

	usage := resp.Usage

	respText := resp.Choices[0].Message.Content
	wrapper := linewrap.NewLineWrapper(termWidth, tabWidth, os.Stdout)

	if _, err := wrapper.Write([]byte(respText)); err != nil {
		return ClientResponse{}, err
	}

	return ClientResponse{
		Text:         respText,
		InputTokens:  usage.PromptTokens,
		OutputTokens: usage.CompletionTokens,
	}, nil
}

func (cs *Ollama) StreamChat(args ClientArgs, termWidth int, tabWidth int, callback func(chunk string)) (ClientResponse, error) {
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
		Model: OllamaModelDeepseekR1_14b,
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

	responseChan, errChan := client.ChatCompletionStream(req)
	var fullText strings.Builder
	var needsSpace bool

	// Helper function to check if a rune is punctuation that shouldn't have spaces before it
	isPunctuation := func(r rune) bool {
		return strings.ContainsRune(",.!?:;\"')", r)
	}

	for {
		select {
		case chunk, ok := <-responseChan:
			if !ok {
				return ClientResponse{
					Text:         fullText.String(),
					InputTokens:  myInputEstimate,
					OutputTokens: EstimateTokens(fullText.String()),
				}, nil
			}

			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				content := chunk.Choices[0].Delta.Content

				// Add space if needed, but not before punctuation
				if needsSpace &&
					!strings.HasPrefix(content, " ") &&
					!strings.HasPrefix(content, "\n") &&
					len(content) > 0 &&
					!isPunctuation([]rune(content)[0]) {
					callback(" ")
					fullText.WriteString(" ")
				}

				callback(content)
				fullText.WriteString(content)

				// Set flag for next chunk if this one ends with a word
				// Don't set needsSpace if we end with punctuation
				lastRune := []rune(content)[len([]rune(content))-1]
				needsSpace = !strings.HasSuffix(content, " ") &&
					!strings.HasSuffix(content, "\n") &&
					!strings.HasSuffix(content, "\t") &&
					!isPunctuation(lastRune) &&
					len(content) > 0
			}

		case err := <-errChan:
			if err != nil {
				return ClientResponse{}, err
			}
		}
	}
}
