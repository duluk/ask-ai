package LLM

import "os"

// These fields need to have capital lettesr to be exported (ugh)

type Claude_Sonnet struct {
	API_Key string
	Tokens  int
}

type Client_Args struct {
	Prompt     string
	Context    int
	Max_tokens int
	Log        *os.File
}
