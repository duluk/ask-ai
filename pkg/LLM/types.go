package LLM

import (
	"context"
	"os"

	"github.com/google/generative-ai-go/genai"
	"github.com/liushuangls/go-anthropic/v2"
	"github.com/openai/openai-go"

	"github.com/duluk/ask-ai/pkg/deepseek"
	"github.com/duluk/ask-ai/pkg/ollama"
)

type LLMConversations struct {
	Role            string `yaml:"role"`
	Content         string `yaml:"content"`
	Model           string `yaml:"model"`
	Timestamp       string `yaml:"timestamp"`
	NewConversation bool   `yaml:"new_conversation"`
	InputTokens     int32  `yaml:"input_tokens"`
	OutputTokens    int32  `yaml:"output_tokens"`
	ConvID          int    `yaml:"conv_id"`
}

type ClientResponse struct {
	Text         string
	InputTokens  int32
	OutputTokens int32
	MyEstInput   int32 // May be used at some point
}

type Client interface {
	Chat(args ClientArgs, termWidth int, tabWidth int) (ClientResponse, error)
}

type Anthropic struct {
	APIKey string
	Client *anthropic.Client
}

type OpenAI struct {
	APIKey string
	Client *openai.Client
}

type DeepSeek struct {
	APIKey string
	Client *deepseek.Client
}

type Google struct {
	APIKey  string
	Client  *genai.Client
	Context context.Context
}

type Ollama struct {
	APIKey string
	Client *ollama.Client
}

type ClientArgs struct {
	Model        *string
	Prompt       *string
	SystemPrompt *string
	Context      []LLMConversations
	MaxTokens    *int
	Temperature  *float32
	Log          *os.File
	ConvID       *int
}
