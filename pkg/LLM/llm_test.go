package LLM

import (
   "os"
   "path/filepath"
   "testing"

   "github.com/liushuangls/go-anthropic/v2"
   "github.com/spf13/viper"
   "github.com/stretchr/testify/assert"
)

// Tokenization tests
func TestTokenizeWord_EdgeCases(t *testing.T) {
   tests := []struct {
       word string
       want int32
       desc string
   }{
       {"(hello)", 3, "wrapped punctuation"},
       {"foo_bar", 3, "underscore punctuation"},
       {"你好", 1, "multi-byte letters"},
       {"👋🌍", 2, "emoji tokens"},
       {"", 0, "empty word"},
   }
   for _, tt := range tests {
       got := tokenizeWord(tt.word)
       assert.Equal(t, tt.want, got, tt.desc)
   }
}

func TestEstimateTokens(t *testing.T) {
   tests := []struct {
       input    string
       expected int32
   }{
       {"hello world", 2},
       {"hello, world!", 4},
       {"foo-bar baz", 4},
       {"123 4567", 2},
       {"", 0},
       {"a b c", 3},
       {"!@#", 3},
   }
   for _, tt := range tests {
       got := EstimateTokens(tt.input)
       assert.Equal(t, tt.expected, got, "input %q", tt.input)
   }
}

func TestEstimateTokens_Whitespace(t *testing.T) {
   input := "a\tb\nc  d"
   got := EstimateTokens(input)
   assert.Equal(t, int32(4), got)
}

// getClientKey tests
func TestGetClientKey(t *testing.T) {
   os.Setenv("TEST_API_KEY", "test-key")
   key, err := getClientKey("test")
   assert.NoError(t, err)
   assert.Equal(t, "test-key", key)
   os.Unsetenv("TEST_API_KEY")

   home := os.Getenv("HOME")
   cfgDir := filepath.Join(home, ".config", "ask-ai")
   os.MkdirAll(cfgDir, 0o755)
   keyPath := filepath.Join(cfgDir, "test-api-key")
   os.WriteFile(keyPath, []byte("file-test-key"), 0o644)
   defer os.Remove(keyPath)

   key, err = getClientKey("test")
   assert.NoError(t, err)
   assert.Equal(t, "file-test-key", key)
}

func TestGetClientKey_PriorityEnvViper(t *testing.T) {
   viper.Reset()
   os.Unsetenv("TESTPRE_API_KEY")
   defer os.Unsetenv("TESTPRE_API_KEY")
   os.Setenv("TESTPRE_API_KEY", "envKey")
   viper.Set("models.testpre.api_key", "cfgKey")

   key, err := getClientKey("testpre")
   assert.NoError(t, err)
   assert.Equal(t, "envKey", key)
}

func TestGetClientKey_ViperNoWhitespace(t *testing.T) {
   viper.Reset()
   os.Unsetenv("TEST_API_KEY")
   cfgHome := t.TempDir()
   os.Setenv("XDG_CONFIG_HOME", cfgHome)

   viper.Set("models.test.api_key", "cfg-key")
   key, err := getClientKey("test")
   assert.NoError(t, err)
   assert.Equal(t, "cfg-key", key)
}

func TestGetClientKey_ViperShellCommand(t *testing.T) {
   viper.Reset()
   os.Unsetenv("TEST_API_KEY")
   cfgHome := t.TempDir()
   os.Setenv("XDG_CONFIG_HOME", cfgHome)

   viper.Set("models.test.api_key", "echo shell-key")
   key, err := getClientKey("test")
   assert.NoError(t, err)
   assert.Equal(t, "shell-key", key)
}

func TestGetClientKey_FileMultiline(t *testing.T) {
   viper.Reset()
   os.Unsetenv("MULTI_API_KEY")
   tmp := t.TempDir()
   os.Setenv("XDG_CONFIG_HOME", tmp)
   defer os.Unsetenv("XDG_CONFIG_HOME")

   dir := filepath.Join(tmp, "ask-ai")
   os.MkdirAll(dir, 0o755)
   fname := filepath.Join(dir, "multi-api-key")
   os.WriteFile(fname, []byte("line1\nline2"), 0o644)

   key, err := getClientKey("multi")
   assert.NoError(t, err)
   assert.Equal(t, "line1", key)
}

// buildPrompt tests
func TestBuildPrompt(t *testing.T) {
   ctx := []LLMConversations{
       {Role: "user", Content: "hello"},
       {Role: "assistant", Content: "hi there"},
   }
   result := buildPrompt(ctx, "new message")
   expected := "user: hello\nassistant: hi there\nuser: new message"
   assert.Equal(t, expected, result)
}

func TestBuildPrompt_EmptyContext(t *testing.T) {
   got := buildPrompt(nil, "ask me")
   assert.Equal(t, "user: ask me", got)
}

// client constructors tests
func TestNewOpenAI_SetsAPIKeyAndClient(t *testing.T) {
   os.Setenv("OPENAI_API_KEY", "oapi")
   defer os.Unsetenv("OPENAI_API_KEY")
   cli := NewOpenAI("openai", "http://example.com")
   assert.Equal(t, "oapi", cli.APIKey)
   assert.NotNil(t, cli.Client)
}

func TestNewGoogle_SetsAPIKeyAndClient(t *testing.T) {
   os.Setenv("GOOGLE_API_KEY", "gkey")
   defer os.Unsetenv("GOOGLE_API_KEY")
   g := NewGoogle()
   assert.Equal(t, "gkey", g.APIKey)
   assert.NotNil(t, g.Client)
}

// conversion tests
func TestConvertToAnthropicMessages(t *testing.T) {
   hist := []LLMConversations{
       {Role: "user", Content: "u1"},
       {Role: "assistant", Content: "a1"},
       {Role: "unknown", Content: "x"},
   }
   msgs := convertToAnthropicMessages(hist)
   assert.Len(t, msgs, 3)
   assert.Equal(t, anthropic.RoleUser, msgs[0].Role)
   assert.Equal(t, "u1", *msgs[0].Content[0].Text)
   assert.Equal(t, anthropic.RoleAssistant, msgs[1].Role)
   assert.Equal(t, "a1", *msgs[1].Content[0].Text)
   assert.Equal(t, anthropic.ChatRole(""), msgs[2].Role)
   assert.Equal(t, "x", *msgs[2].Content[0].Text)
}
