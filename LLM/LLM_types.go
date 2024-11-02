package LLM

import (
	"os"

	"github.com/liushuangls/go-anthropic/v2"
)

// These fields need to have capital lettesr to be exported (ugh)

type Claude struct {
	API_Key string
	Tokens  int
	Client  *anthropic.Client
}

type Client_Args struct {
	Prompt     string
	Context    int
	Max_tokens int
	Log        *os.File
}
