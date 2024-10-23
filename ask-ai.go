package main

// TODO:
// - Add a flag to specify the model to use
// - Read the chat log for context possibilities
//   - That is, could add a flag to read the last n messages for context
// - Add a flag to specify the chat log file

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

	log, err := os.OpenFile(home+"/.config/ask-ai/ask-ai.chat.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening/creating chat log file: ", err)
		fmt.Println("Chat will not be saved")
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
	log.WriteString("Prompt: " + prompt + "\n\n")

	client := openai.NewClient(option.WithAPIKey(key))
	ctx := context.Background()
	stream := client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		}),
		Seed:  openai.Int(1),
		Model: openai.F(openai.ChatModelGPT4o),
	})

	for stream.Next() {
		evt := stream.Current()
		if len(evt.Choices) > 0 {
			data := evt.Choices[0].Delta.Content
			print(data)
			_, err := log.WriteString(data)
			if err != nil {
				fmt.Println("Error writing chat to chat log file: ", err)
			}
		}
	}
	println()
	log.WriteString("\n------\n")

	if stream.Err() != nil {
		fmt.Printf("Error: %s\n", stream.Err())
	}
}
