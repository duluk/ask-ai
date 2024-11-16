package LLM

import (
	"context"
	"os"

	"github.com/google/generative-ai-go/genai"
	"github.com/liushuangls/go-anthropic/v2"
	"github.com/openai/openai-go"
)

// These fields need to have capital lettesr to be exported (ugh)

type LLMConversations struct {
	Role            string `yaml:"role"`
	Content         string `yaml:"content"`
	Model           string `yaml:"model"`
	Timestamp       string `yaml:"timestamp"`
	NewConversation bool   `yaml:"new_conversation"`
}

type Client interface {
	Chat(args ClientArgs) (string, error)
}

type Anthropic struct {
	APIKey string
	Tokens int
	Client *anthropic.Client
}

type OpenAI struct {
	APIKey string
	Tokens int
	Client *openai.Client
}

type Google struct {
	APIKey  string
	Tokens  int
	Client  *genai.Client
	Context context.Context
}

type ClientArgs struct {
	Model        *string
	Prompt       *string
	SystemPrompt *string
	Context      []LLMConversations
	MaxTokens    *int
	Temperature  *float32
	Log          *os.File
}
