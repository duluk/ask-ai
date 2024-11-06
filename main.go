package main

// TODO:
// - Read the chat log for context possibilities
//   - A flag exists for this, but it's not implemented yet
// - Create an output class/struct or something that can receive different
//   'stream' objects so that one output function can be called, then it will
//   send the output to all attached streams. (eg, stdout, log file, etc)
// - Consider something similar to the above for the backend model itself. This
//   would allow the --compare flag mentioned below to be implemented.
// - Add configuration file for:
//   * providing the option for storing chat results in a DB
//   * storing model and system prompt information
// - How would this app be tested?
// - Add --compare flag to use mulitple models and compare the results
// - Add --system-prompt flag to allow the creation of a system prompt
// - Add support for --image and --file attachments for multi-modal models
// - Add model flags like `--chatgpt`, `--sonnet`, etc instead of having to use
//   `--model chatgpt`

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/duluk/ask-ai/LLM"
)

func chat_with_llm(model string, args LLM.Client_Args) {
	var client LLM.Client

	switch model {
	case "chatgpt":
		client = LLM.New_OpenAI(args.Max_Tokens)
	case "sonnet":
		client = LLM.New_Anthropic(args.Max_Tokens)
	case "gemini":
		client = LLM.New_Google(args.Max_Tokens)
	default:
		fmt.Println("Unknown model: ", model)
	}

	err := client.Chat(args)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
}

func main() {
	HOME := os.Getenv("HOME")

	model := flag.String("model", "sonnet", "Which LLM to use (sonnet|chatgpt|gemini)")
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

	chat_with_llm(*model, client_args)

	footer := fmt.Sprintf("\n\n<model - %s>\n<------>\n", *model)
	log.WriteString(footer)
}
