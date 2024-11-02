package LLM

import "os"

type Claude_Sonnet struct {
	API_Key string
	tokens  int
}

type Client_Args struct {
	prompt     string
	context    int
	max_tokens int
	log        *os.File
}
