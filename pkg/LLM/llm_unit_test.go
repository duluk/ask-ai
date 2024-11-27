package LLM

import (
	"os"
	"testing"
	"time"

	"gopkg.in/yaml.v2"
)

func TestLogChat(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test_log_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	err = LogChat(tempFile, "user", "Hello", "gpt-3.5-turbo", false, 0, 0)
	if err != nil {
		t.Errorf("LogChat failed: %v", err)
	}

	conversations, err := LoadChatLog(tempFile)
	if err != nil {
		t.Errorf("Failed to load chat log: %v", err)
	}

	if len(conversations) != 1 {
		t.Errorf("Expected 1 conversation, got %d", len(conversations))
	}

	if conversations[0].Role != "user" || conversations[0].Content != "Hello" || conversations[0].Model != "gpt-3.5-turbo" {
		t.Errorf("Conversation data mismatch")
	}
}

func TestLoadChatLog(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test_load_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	testData := []LLMConversations{
		{Role: "user", Content: "Hello", Model: "gpt-3.5-turbo", Timestamp: time.Now().Format(time.RFC3339), NewConversation: true},
	}

	yamlData, _ := yaml.Marshal(testData)
	tempFile.Write(yamlData)

	conversations, err := LoadChatLog(tempFile)
	if err != nil {
		t.Errorf("LoadChatLog failed: %v", err)
	}

	if len(conversations) != 1 {
		t.Errorf("Expected 1 conversation, got %d", len(conversations))
	}

	if conversations[0].Role != "user" || conversations[0].Content != "Hello" || conversations[0].Model != "gpt-3.5-turbo" {
		t.Errorf("Conversation data mismatch")
	}
}

func TestLastNChats(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test_last_n_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	testData := []LLMConversations{
		{Role: "user", Content: "Hello", Model: "gpt-3.5-turbo"},
		{Role: "assistant", Content: "Hi there!", Model: "gpt-3.5-turbo"},
		{Role: "user", Content: "How are you?", Model: "gpt-3.5-turbo"},
	}

	yamlData, _ := yaml.Marshal(testData)
	tempFile.Write(yamlData)

	conversations, err := LastNChats(tempFile, 2)
	if err != nil {
		t.Errorf("LastNChats failed: %v", err)
	}

	if len(conversations) != 2 {
		t.Errorf("Expected 2 conversations, got %d", len(conversations))
	}

	if conversations[0].Content != "Hi there!" || conversations[1].Content != "How are you?" {
		t.Errorf("Incorrect last N chats returned")
	}
}

func TestContinueConversation(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test_continue_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	testData := []LLMConversations{
		{Role: "User", Content: "Hello", Model: "gpt-3.5-turbo", NewConversation: true},
		{Role: "Assistant", Content: "Hi there!", Model: "gpt-3.5-turbo", NewConversation: true},
		{Role: "User", Content: "How are you?", Model: "gpt-3.5-turbo", NewConversation: true},
		{Role: "Assistant", Content: "I'm doing well, thanks!", Model: "gpt-3.5-turbo", NewConversation: true},
	}

	yamlData, _ := yaml.Marshal(testData)
	tempFile.Write(yamlData)

	conversations, err := ContinueConversation(tempFile)
	if err != nil {
		t.Errorf("ContinueConversation failed: %v", err)
	}

	if len(conversations) != 2 {
		t.Errorf("Expected 2 conversations, got %d", len(conversations))
	}

	if conversations[0].Content != "How are you?" || conversations[1].Content != "I'm doing well, thanks!" {
		t.Errorf("Incorrect conversation continuation")
	}
}

func TestGetClientKey(t *testing.T) {
	// Test environment variable (first option)
	os.Setenv("TEST_API_KEY", "test-key")
	key := getClientKey("test")
	if key != "test-key" {
		t.Errorf("Expected 'test-key', got '%s'", key)
	}
	os.Unsetenv("TEST_API_KEY")

	home := os.Getenv("HOME")
	os.MkdirAll(home+"/.config/ask-ai", 0755)
	os.WriteFile(home+"/.config/ask-ai/test-api-key", []byte("file-test-key"), 0644)
	defer os.Remove(home + "/.config/ask-ai/test-api-key")

	key = getClientKey("test")
	if key != "file-test-key" {
		t.Errorf("Expected 'file-test-key', got '%s'", key)
	}
}
