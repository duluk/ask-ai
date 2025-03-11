package LLM

import (
	"context"
	"os"

	"github.com/google/generative-ai-go/genai"
	"github.com/liushuangls/go-anthropic/v2"

	// "github.com/anthropics/anthropic-sdk-go"
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

// StreamResponse represents a chunk of streaming response
type StreamResponse struct {
	Content string
	Done    bool
	Error   error
}

type Client interface {
	Chat(args ClientArgs, termWidth int, tabWidth int) (ClientResponse, <-chan StreamResponse, error)
	ChatStream(args ClientArgs, termWidth int, tabWidth int, stream chan<- StreamResponse) error
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
	Model         *string
	Prompt        *string
	SystemPrompt  *string
	Context       []LLMConversations
	MaxTokens     *int
	Temperature   *float32
	Log           *os.File
	ConvID        *int
	DisableOutput bool
}
