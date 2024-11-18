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

func LogChat(logFd *os.File, role string, content string, model string, continueChat bool) error {
	// TODO: is it necessary to load the file every time? I suppose it's not
	// the worst since this is a run-once program. But if the log is very
	// large, it seems inefficient to read it all in, append to it, then
	// re-write it. (only plus is that when changing the YML structure, it
	// automarically re-writes the entire log and applies the new tags, though
	// they may be empty)
	chat, err := LoadChatLog(logFd)
	if err != nil {
		return err
	}

	timestamp := time.Now().Format(time.RFC3339)

	chat = append(chat, LLMConversations{
		Role:            role,
		Content:         content,
		Model:           model,
		Timestamp:       timestamp,
		NewConversation: !continueChat,
	})

	data, err := yaml.Marshal(chat)
	if err != nil {
		return err
	}

	// Remove all data and re-write from scratch to avoid corruption or
	// duplication. Until I decide on a better way to handle this that doens't
	// involve reading the entire log and re-writing.
	err = logFd.Truncate(0)
	if err != nil {
		return err
	}
	_, err = logFd.Seek(0, 0)
	if err != nil {
		return err
	}

	_, err = logFd.Write(data)
	return err
}

// TODO: should this be loaded into some memory structure? I think it's
// probably only called twice in one run, but things could change.
func LoadChatLog(logFd *os.File) ([]LLMConversations, error) {
	_, err := logFd.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	content, err := io.ReadAll(logFd)
	if err != nil {
		return nil, err
	}

	var conversations []LLMConversations
	err = yaml.Unmarshal(content, &conversations)
	if err != nil {
		return nil, err
	}

	return conversations, nil
}

func LastNChats(log_fd *os.File, n int) ([]LLMConversations, error) {
	chat, err := LoadChatLog(log_fd)
	if err != nil {
		return nil, err
	}

	totalTurns := len(chat)
	if n <= 0 || n >= totalTurns {
		return nil, fmt.Errorf("Context value is invalid (either <= 0 or too large): %d", n)
	}

	return chat[totalTurns-n:], nil
}

func ContinueConversation(logFd *os.File) ([]LLMConversations, error) {
	chat, err := LoadChatLog(logFd)
	if err != nil {
		return nil, err
	}

	if len(chat) == 0 {
		return nil, fmt.Errorf("No chat history to continue")
	}

	lastConv := 0
	for i := len(chat) - 1; i >= 0; i-- {
		if chat[i].NewConversation {
			lastConv = i
			break
		}
	}

	// -1 to get the first user prompt for the conversation
	return chat[lastConv-1:], nil
}

func getClientKey(llm string) string {
	keyUpper := strings.ToUpper(llm) + "_API_KEY"
	keyLower := strings.ToLower(llm) + "-api-key"

	key := os.Getenv(keyUpper)
	// TODO this should attempt XDG_CONFIG_HOME first, then HOME
	home := os.Getenv("HOME")
	if key == "" {
		file, err := os.Open(home + "/.config/ask-ai/" + keyLower)
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
