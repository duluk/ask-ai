package LLM

import (
	"context"
	"fmt"

	"github.com/duluk/ask-ai/pkg/linewrap"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func NewOpenAI(apiLLC string, apiURL string) *OpenAI {
	apiKey := getClientKey(apiLLC)
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(apiURL),
	)

	return &OpenAI{APIKey: apiKey, Client: client}
}

func (cs *OpenAI) Chat(args ClientArgs) (string, error) {
	client := cs.Client

	var msgCtx string
	for _, msg := range args.Context {
		switch msg.Role {
		case "User":
			msgCtx += "User: " + msg.Content + "\n"
		case "Assistant":
			msgCtx += "Assistant: " + msg.Content + "\n"
		}
	}

	const ChatModelGrokBeta openai.ChatModel = "grok-beta"
	var model openai.ChatModel
	switch *args.Model {
	case "chatgpt":
		// model = ChatModelO1Preview2024_09_12            ChatModel = "o1-preview-2024-09-12"
		model = openai.ChatModelGPT4o
	case "grok":
		model = ChatModelGrokBeta
	}

	ctx := context.Background()
	stream := client.Chat.Completions.NewStreaming(
		ctx,
		openai.ChatCompletionNewParams{
			Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
				openai.AssistantMessage(msgCtx),
				openai.UserMessage(*args.Prompt),
			}),
			// Seed:        openai.Int(1), // Same seed/parameters will attempt to return the same results
			Model:               openai.F(model),
			MaxCompletionTokens: openai.Int(int64(*args.MaxTokens)),
			Temperature:         openai.Float(float64(*args.Temperature)), // Controls randomness (0.0 to 2.0)
			// TopP:             openai.Float(1.0),                // Controls diversity via nucleus sampling; alter this or Temperature but not both
			// N:                openai.Int(1),                    // Number of completions to generate
			// ResponseFormat:   openai.ChatResponseFormatDefault, // Format of the response
		},
	)

	// Apparently what happens with stream is that the server chunks the
	// response according to its own internal desires and whims, presenting the
	// result as if it's a stream of responses, which looks more
	// conversational.
	var resp string
	wrapper := linewrap.NewLineWrapper(TermWidth, TabWidth)
	for stream.Next() {
		evt := stream.Current()
		if len(evt.Choices) > 0 {
			data := evt.Choices[0].Delta.Content
			if _, err := wrapper.Write([]byte(data)); err != nil {
				fmt.Printf("Error writing to wrapper: %s\n", err)
				return "", err
			}
			resp += data
			// fmt.Printf(data)
		}
	}

	if stream.Err() != nil {
		fmt.Printf("Error: %s\n", stream.Err())
		return "", stream.Err()
	}

	return resp, nil
}
