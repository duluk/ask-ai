package LLM

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

func Log_Chat(log_fd *os.File, role string, content string, model string, continue_chat bool) error {
	// TODO: is it necessary to load the file every time? I suppose it's not
	// the worst since this is a run-once program. But if the log is very
	// large, it seems inefficient to read it all in, append to it, then
	// re-write it. (only plus is that when changing the YML structure, it
	// automarically re-writes the entire log and applies the new tags, though
	// they may be empty)
	chat, err := Load_Chat_Log(log_fd)
	if err != nil {
		return err
	}

	timestamp := time.Now().Format(time.RFC3339)

	chat = append(chat, LLM_Conversations{
		Role:             role,
		Content:          content,
		Model:            model,
		Timestamp:        timestamp,
		New_Conversation: !continue_chat,
	})

	data, err := yaml.Marshal(chat)
	if err != nil {
		return err
	}

	// Remove all data and re-write from scratch to avoid corruption or
	// duplication. Until I decide on a better way to handle this that doens't
	// involve reading the entire log and re-writing.
	err = log_fd.Truncate(0)
	if err != nil {
		return err
	}
	_, err = log_fd.Seek(0, 0)
	if err != nil {
		return err
	}

	_, err = log_fd.Write(data)
	return err
}

// TODO: should this be loaded into some memory structure? I think it's
// probably only called twice in one run, but things could change.
func Load_Chat_Log(log_fd *os.File) ([]LLM_Conversations, error) {
	_, err := log_fd.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	content, err := io.ReadAll(log_fd)
	if err != nil {
		return nil, err
	}

	var conversations []LLM_Conversations
	err = yaml.Unmarshal(content, &conversations)
	if err != nil {
		return nil, err
	}

	return conversations, nil
}

func Last_n_Chats(log_fd *os.File, n int) ([]LLM_Conversations, error) {
	chat, err := Load_Chat_Log(log_fd)
	if err != nil {
		return nil, err
	}

	total_turns := len(chat)
	if n <= 0 || n >= total_turns {
		return nil, fmt.Errorf("Context value is invalid (either <= 0 or too large): %d", n)
	}

	return chat[total_turns-n:], nil
}

func Continue_Conversation(log_fd *os.File) ([]LLM_Conversations, error) {
	chat, err := Load_Chat_Log(log_fd)
	if err != nil {
		return nil, err
	}

	if len(chat) == 0 {
		return nil, fmt.Errorf("No chat history to continue")
	}

	last_conversation := 0
	for i := len(chat) - 1; i >= 0; i-- {
		if chat[i].New_Conversation {
			last_conversation = i
			break
		}
	}

	// -1 to get the first user prompt for the conversation
	return chat[last_conversation-1:], nil
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
