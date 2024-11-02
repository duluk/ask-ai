package LLM

// Add support for the anthropic model, Claude Sonnet

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/liushuangls/go-anthropic/v2"
)

func New_Anthropic(max_tokens int) *Anthropic {
	api_key := get_anthropic_key()
	client := anthropic.NewClient(api_key)

	return &Anthropic{API_Key: api_key, Tokens: max_tokens, Client: client}
}

func (cs *Anthropic) Chat(args Client_Args) error {
	prompt := args.Prompt
	log := args.Log
	client := cs.Client

	resp, err := client.CreateMessagesStream(context.Background(), anthropic.MessagesStreamRequest{
		MessagesRequest: anthropic.MessagesRequest{
			Model: anthropic.ModelClaudeInstant1Dot2,
			Messages: []anthropic.Message{
				anthropic.NewUserTextMessage(prompt),
			},
			MaxTokens: 1000,
		},
		OnContentBlockDelta: func(data anthropic.MessagesEventContentBlockDeltaData) {
			print(*data.Delta.Text)
		},
	})
	if err != nil {
		var e *anthropic.APIError
		if errors.As(err, &e) {
			fmt.Printf("Messages stream error, type: %s, message: %s", e.Type, e.Message)
		} else {
			fmt.Printf("Messages stream error: %v\n", err)
		}
		return err
	}
	println()
	log.WriteString("Assistant: " + resp.Content[0].GetText())
	log.WriteString("\n<------>\n")

	return nil
}

func get_anthropic_key() string {
	key := os.Getenv("ANTHROPIC_API_KEY")
	home := os.Getenv("HOME")
	if key == "" {
		file, err := os.Open(home + "/.config/ask-ai/anthropic-api-key")
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
