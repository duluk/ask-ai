package main

// TODO:
// - Add a flag to specify the chat log file
// - Read the chat log for context possibilities
//   - A flag exists for this, but it's not implemented yet
// - Create an output class/struct or something that can receive different
//   'stream' objects so that one output function can be called, then it will
//   send the output to all attached streams. (eg, stdout, log file, etc)

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/duluk/ask-ai/LLM"
)

func chat_with_openai(args LLM.Client_Args) {
	client := LLM.New_OpenAI(args.Max_Tokens)
	err := client.Chat(args)
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func chat_with_sonnet(args LLM.Client_Args) {
	client := LLM.New_Anthropic(args.Max_Tokens)
	err := client.Chat(args)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
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
	// NArg is for positional arguments, so it can accept a prompt as a
	// positional string argument
	if flag.NArg() > 0 {
		prompt = flag.Arg(0)
	} else {
		fmt.Println("Using model:", *model)
		fmt.Print("> ")
		reader := bufio.NewReader(os.Stdin)
		prompt, err = reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading prompt: ", err)
			os.Exit(1)
		}
		fmt.Println()
	}
	log.WriteString("User: " + prompt + "\n\n")

	client_args := LLM.Client_Args{
		Prompt:     prompt,
		Context:    *context,
		Max_Tokens: *max_tokens,
		Log:        log,
	}

	switch *model {
	case "sonnet":
		chat_with_sonnet(client_args)
	case "chatgpt":
		chat_with_openai(client_args)
	default:
		fmt.Println("Unknown model: ", *model)
	}
}
