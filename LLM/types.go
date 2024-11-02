package LLM

import (
	"os"

	"github.com/liushuangls/go-anthropic/v2"
	"github.com/openai/openai-go"
)

// These fields need to have capital lettesr to be exported (ugh)

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

type Client_Args struct {
	Prompt     string
	Context    int
	Max_Tokens int
	Log        *os.File
}
