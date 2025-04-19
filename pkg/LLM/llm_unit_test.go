package LLM

import (
	"os"
	"path/filepath"
	"testing"
)

// func intPtr(i int) *int {
// 	return &i
// }

func TestGetClientKey(t *testing.T) {
	// Test environment variable (first option)
	os.Setenv("TEST_API_KEY", "test-key")
	key, err := getClientKey("test")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if key != "test-key" {
		t.Errorf("Expected 'test-key', got '%s'", key)
	}
	os.Unsetenv("TEST_API_KEY")

	home := os.Getenv("HOME")
	// Ensure we write into the same config dir used by getClientKey
	cfgDir := filepath.Join(home, ".config", "ask-ai")
	os.MkdirAll(cfgDir, 0o755)
	keyPath := filepath.Join(cfgDir, "test-api-key")
	os.WriteFile(keyPath, []byte("file-test-key"), 0o644)
	defer os.Remove(keyPath)

	key, err = getClientKey("test")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if key != "file-test-key" {
		t.Errorf("Expected 'file-test-key', got '%s'", key)
	}
}
