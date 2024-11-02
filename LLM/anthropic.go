package LLM

// Add support for the anthropic model, Claude Sonnet

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/liushuangls/go-anthropic/v2"
)

func New_Claude(max_tokens int) *Claude {
	api_key := get_anthropic_key()
	return &Claude{API_Key: api_key, Tokens: max_tokens}
}

func (cs *Claude) Chat(args Client_Args) (string, error) {
	client := anthropic.NewClient(cs.API_Key)

	prompt := args.Prompt
	// log := args.Log

	resp, err := client.CreateMessagesStream(context.Background(), anthropic.MessagesStreamRequest{
		MessagesRequest: anthropic.MessagesRequest{
			Model: anthropic.ModelClaudeInstant1Dot2,
			Messages: []anthropic.Message{
				anthropic.NewUserTextMessage(prompt),
			},
			MaxTokens: 1000,
		},
		OnContentBlockDelta: func(data anthropic.MessagesEventContentBlockDeltaData) {
			fmt.Printf("Stream Content: %s\n", data.Delta.Text)
			// log.WriteString("Claude: " + data.Delta.Text + "\n\n")
		},
	})
	if err != nil {
		var e *anthropic.APIError
		if errors.As(err, &e) {
			fmt.Printf("Messages stream error, type: %s, message: %s", e.Type, e.Message)
		} else {
			fmt.Printf("Messages stream error: %v\n", err)
		}
		return "", err
	}

	return resp.Content[0].GetText(), nil
}

func (cs *Claude) first_Chat(args Client_Args) (string, error) {
	url := "https://api.openai.com/v1/engines/claude/complete"

	prompt := args.Prompt
	log := args.Log

	payload := strings.NewReader(`{"prompt":"` + prompt + `","max_tokens":` + string(cs.Tokens) + `}`)
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return "", err
	}
	fmt.Println("Using API Key: '" + cs.API_Key + "'")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+cs.API_Key)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", errors.New("Error: " + res.Status)
	}

	var data map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return "", err
	}

	if log != nil {
		log.Write([]byte("Claude: " + data["choices"].([]interface{})[0].(map[string]interface{})["text"].(string) + "\n\n"))
		log.WriteString("<------>\n")
	}

	return data["choices"].([]interface{})[0].(map[string]interface{})["text"].(string), nil
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
