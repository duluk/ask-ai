package LLM

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func New_OpenAI(max_tokens int) *OpenAI {
	api_key := get_client_key("openai")
	client := openai.NewClient(option.WithAPIKey(api_key))

	return &OpenAI{API_Key: api_key, Tokens: max_tokens, Client: client}
}

func (cs *OpenAI) Chat(args Client_Args) error {
	client := cs.Client
	log := args.Log

	ctx := context.Background()
	stream := client.Chat.Completions.NewStreaming(
		ctx,
		openai.ChatCompletionNewParams{
			Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
				// To use context from previous responses, use AssistantMessage:
				// openai.AssistantMessage(msg_context),
				openai.UserMessage(args.Prompt),
			}),
			Seed:  openai.Int(1),
			Model: openai.F(openai.ChatModelGPT4o),
		})

	// Apparently what happens with stream is that the server chunks the
	// response according to its own internal desires and whims, presenting the
	// result as if it's a stream of responses, which looks more
	// conversational.
	// TODO: I need to line-wrap the responses so they don't go all the way
	// across the entire screen.
	// TODO: make the `log` an aggregate of streams, as in the TODO for the
	// main app; so that it's not just going to stdout by default
	log.WriteString("Assistant: ")
	for stream.Next() {
		evt := stream.Current()
		if len(evt.Choices) > 0 {
			data := evt.Choices[0].Delta.Content
			print(data)
			log.WriteString(data)
		}
	}

	if stream.Err() != nil {
		fmt.Printf("Error: %s\n", stream.Err())
		return stream.Err()
	}

	return nil
}
