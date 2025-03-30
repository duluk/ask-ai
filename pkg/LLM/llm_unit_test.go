package LLM

import (
	"os"
	"testing"
)

func intPtr(i int) *int {
	return &i
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
	os.MkdirAll(home+"/.config/ask-ai", 0o755)
	os.WriteFile(home+"/.config/ask-ai/test-api-key", []byte("file-test-key"), 0o644)
	defer os.Remove(home + "/.config/ask-ai/test-api-key")

	key = getClientKey("test")
	if key != "file-test-key" {
		t.Errorf("Expected 'file-test-key', got '%s'", key)
	}
}
