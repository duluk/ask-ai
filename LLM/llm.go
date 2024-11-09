package LLM

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

func Log_Chat(log_file *string, role string, content string) error {
	// TODO: is it necessary to load the file every time? I suppose it's not
	// the worst since this is a run-once program.
	chat, _ := Load_Chat_Log(log_file)

	chat = append(chat, LLM_Conversations{Role: role, Content: content})

	data, err := yaml.Marshal(chat)
	if err != nil {
		return err
	}

	return os.WriteFile(*log_file, data, 0644)
}

func Load_Chat_Log(log_file *string) ([]LLM_Conversations, error) {
	// Read the YAML data from the file
	data, err := os.ReadFile(*log_file)
	if err != nil {
		return nil, err
	}

	// Unmarshal the YAML data into a slice of ConversationTurn
	var conversations []LLM_Conversations
	err = yaml.Unmarshal(data, &conversations)
	if err != nil {
		return nil, err
	}

	return conversations, nil
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
