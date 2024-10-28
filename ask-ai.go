package main

// TODO:
// - Add a flag to specify the model to use
// - Add a flag to specify the chat log file
// - Read the chat log for context possibilities
//   - That is, could add a flag to read the last n messages for context
// - Create an output class/struct or something that can receive different
//   'stream' objects so that one output functoin can be called, then it will
//   send the output to all attached streams. (eg, stdout, log file, etc)

import (
	"bufio"
	"context"
	"fmt"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"os"
)

func main() {
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

	log, err := os.OpenFile(home+"/.config/ask-ai/ask-ai.chat.log", os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("Error opening/creating chat log file: ", err)
		fmt.Println("CHAT WILL NOT BE SAVED (but we're forging on)")
	}

	var prompt string
	if len(os.Args) > 1 {
		prompt = os.Args[1]
	} else {
		fmt.Print("> ")
		reader := bufio.NewReader(os.Stdin)
		prompt, _ = reader.ReadString('\n')
		fmt.Println()
	}
	log.WriteString("User: " + prompt + "\n\n")

	client := openai.NewClient(option.WithAPIKey(key))
	ctx := context.Background()
	stream := client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			// To use context from previous responses, use AssistantMessage:
			// openai.AssistantMessage(msg_context),
			openai.UserMessage(prompt),
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
	log.WriteString("Assistant: ")
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
	}
}
