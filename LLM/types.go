package LLM

import (
	"context"

	"github.com/google/generative-ai-go/genai"
	"github.com/liushuangls/go-anthropic/v2"
	"github.com/openai/openai-go"
)

// These fields need to have capital lettesr to be exported (ugh)

type Client interface {
	Chat(args Client_Args) error
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
	Prompt     string
	Context    int
	Max_Tokens int
	Out        *Output_Stream
}
