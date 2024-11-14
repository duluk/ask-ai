package LLM

import (
	"context"
	"os"

	"github.com/google/generative-ai-go/genai"
	"github.com/liushuangls/go-anthropic/v2"
	"github.com/openai/openai-go"
)

// These fields need to have capital lettesr to be exported (ugh)

type LLM_Conversations struct {
	Role             string `yaml:"role"`
	Content          string `yaml:"content"`
	Model            string `yaml:"model"`
	Timestamp        string `yaml:"timestamp"`
	New_Conversation bool   `yaml:"new_conversation"`
}

type Client interface {
	Chat(args Client_Args) (string, error)
}

type Anthropic struct {
	API_Key string
	Tokens  int
	Client  *anthropic.Client
}

type OpenAI struct {
	API_Key string
	Tokens  int
	Client  *openai.Client
}

type Google struct {
	API_Key string
	Tokens  int
	Client  *genai.Client
	Context context.Context
}

type Client_Args struct {
	Prompt        *string
	System_Prompt *string
	Context       []LLM_Conversations
	Max_Tokens    *int
	Temperature   *float32
	Log           *os.File
}
