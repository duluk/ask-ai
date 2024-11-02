package main

// TODO:
// - Add a flag to specify the model to use
// - Add a flag to specify the chat log file
// - Read the chat log for context possibilities
//   - That is, could add a flag to read the last n messages for context
// - Create an output class/struct or something that can receive different
//   'stream' objects so that one output functoin can be called, then it will
//   send the output to all attached streams. (eg, stdout, log file, etc)

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/duluk/ask-ai/LLM"
)

// func chat_with_openai(prompt string, context int, max_tokens int, log *os.File) {
func chat_with_openai(args LLM.Client_Args) {
	client := LLM.New_Client()
	err := LLM.Chat(client, args)
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

// func chat_with_sonnet(prompt string, context int, max_tokens int, log *os.File) {
func chat_with_sonnet(args LLM.Client_Args) {
	cs := LLM.New_Claude_Sonnet(args.Max_tokens)
	resp, err := cs.Chat(args)
	if err != nil {
		fmt.Println("Error: ", err)
	}
	fmt.Println("Claude: ", resp)
}

func main() {
	HOME := os.Getenv("HOME")

	model := flag.String("model", "sonnet", "Which LLM to use (sonnet|chatgpt)")
	log_fn := flag.String("log", HOME+"/.config/ask-ai/ask-ai.chat.log", "Chat log file")
	context := flag.Int("context", 0, "Use n previous messages for context")
	max_tokens := flag.Int("max-tokens", 1024, "Maximum tokens to generate")

	flag.Parse()

	log, err := os.OpenFile(*log_fn, os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("Error opening/creating chat log file: ", err)
		fmt.Println("CHAT WILL NOT BE SAVED (but we're forging on)")
	}
	defer log.Close()

	var prompt string
	if flag.NArg() > 0 {
		prompt = flag.Arg(0)
	} else {
		fmt.Print("> ")
		reader := bufio.NewReader(os.Stdin)
		prompt, _ = reader.ReadString('\n')
		fmt.Println()
	}
	log.WriteString("User: " + prompt + "\n\n")

	client_args := LLM.Client_Args{
		Prompt:     prompt,
		Context:    *context,
		Max_tokens: *max_tokens,
		Log:        log,
	}

	switch *model {
	case "sonnet":
		// chat_with_sonnet(prompt, *context, *max_tokens, log)
		chat_with_sonnet(client_args)
	case "chatgpt":
		// chat_with_openai(prompt, *context, *max_tokens, log)
		chat_with_openai(client_args)
	default:
		fmt.Println("Unknown model: ", *model)
	}
}
