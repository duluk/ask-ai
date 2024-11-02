package LLM

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func New_Client() *openai.Client {
	api_key := get_openai_key()
	return openai.NewClient(option.WithAPIKey(api_key))
}

func Chat(client *openai.Client, args Client_Args) error {
	log := args.Log

	ctx := context.Background()
	stream := client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
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
	log.WriteString("ChatGPT: ")
	for stream.Next() {
		evt := stream.Current()
		if len(evt.Choices) > 0 {
			data := evt.Choices[0].Delta.Content
			print(data)
			log.WriteString(data)
		}
	}
	println()
	log.WriteString("\n<------>\n")

	if stream.Err() != nil {
		fmt.Printf("Error: %s\n", stream.Err())
		return stream.Err()
	}

	return nil
}

func get_openai_key() string {
	key := os.Getenv("OPENAI_API_KEY")
	home := os.Getenv("HOME")
	if key == "" {
		file, err := os.Open(home + "/.config/ask-ai/openai-api-key")
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		if scanner.Scan() {
			key = scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}
	}
	return key
}
