package LLM

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

func Log_Chat(log_file *string, role string, content string, model string) error {
	// TODO: is it necessary to load the file every time? I suppose it's not
	// the worst since this is a run-once program. But if the log is very
	// large, it seems inefficient to read it all in, append to it, then
	// re-write it.
	chat, _ := Load_Chat_Log(log_file)
	timestamp := time.Now().Format(time.RFC3339)

	chat = append(chat, LLM_Conversations{Role: role, Content: content, Model: model, Timestamp: timestamp})

	data, err := yaml.Marshal(chat)
	if err != nil {
		return err
	}

	return os.WriteFile(*log_file, data, 0644)
}

func Load_Chat_Log(log_file *string) ([]LLM_Conversations, error) {
	chat, err := os.ReadFile(*log_file)
	if err != nil {
		return nil, err
	}

	var conversations []LLM_Conversations
	err = yaml.Unmarshal(chat, &conversations)
	if err != nil {
		return nil, err
	}

	return conversations, nil
}

func Last_n_Chats(log_file *string, n int) ([]LLM_Conversations, error) {
	chat, err := Load_Chat_Log(log_file)
	if err != nil {
		return nil, err
	}

	total_turns := len(chat)
	if n <= 0 || n >= total_turns {
		return nil, fmt.Errorf("Context value is invalid (either <= 0 or too large): %d", n)
	}

	return chat[total_turns-n:], nil
}

func get_client_key(llm string) string {
	key_upper := strings.ToUpper(llm) + "_API_KEY"
	key_lower := strings.ToLower(llm) + "-api-key"

	key := os.Getenv(key_upper)
	// TODO this should attempt XDG_CONFIG_HOME first, then HOME
	home := os.Getenv("HOME")
	if key == "" {
		file, err := os.Open(home + "/.config/ask-ai/" + key_lower)
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
