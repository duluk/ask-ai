package logger

import (
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// resetLogger clears the logger state for testing
func resetLogger() {
	defaultLogger = nil
	logWriter = nil
	once = sync.Once{}
}

func TestInitializeAndInfoTextFormat(t *testing.T) {
	resetLogger()
	// Prepare temp file
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.log")
	cfg := Config{
		Level:      slog.LevelInfo,
		Format:     "text",
		FilePath:   filePath,
		MaxSize:    1,
		MaxBackups: 1,
		MaxAge:     1,
		Compress:   false,
		UseConsole: false,
	}
	// Initialize logger
	err := Initialize(cfg)
	require.NoError(t, err)
	// Log a message with key/value
	Info("hello world", "key", "value")
	// Close writer
	err = Close()
	require.NoError(t, err)
	// Read file and verify content
	data, err := os.ReadFile(filePath)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "hello world")
	assert.Contains(t, content, "key=value")
}

func TestClose_NoLogWriter(t *testing.T) {
	resetLogger()
	// No logger initialized; Close should not error
	err := Close()
	assert.NoError(t, err)
}
