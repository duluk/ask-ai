package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/duluk/ask-ai/LLM"
)

// I'm probably writing "Ruby Go"...

func chat_with_llm(model string, args LLM.Client_Args) {
	var client LLM.Client
	log := args.Log

	switch model {
	case "chatgpt":
		client = LLM.New_OpenAI(args.Max_Tokens)
	case "claude":
		client = LLM.New_Anthropic(args.Max_Tokens)
	case "gemini":
		client = LLM.New_Google(args.Max_Tokens)
	default:
		fmt.Println("Unknown model: ", model)
		os.Exit(1)
	}
	LLM.Log_Chat(log, "User", args.Prompt, "")

	fmt.Printf("Assistant: ")
	resp, err := client.Chat(args)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}

	LLM.Log_Chat(log, "Assistant", resp, model)
}

func main() {
	HOME := os.Getenv("HOME")

	model := flag.String("model", "claude", "Which LLM to use (claude|chatgpt|gemini)")
	log_fn := flag.String("log", HOME+"/.config/ask-ai/ask-ai.chat.yml", "Chat log file")
	context := flag.Int("context", 0, "Use n previous messages for context")
	max_tokens := flag.Int("max-tokens", 4096, "Maximum tokens to generate")

	flag.Parse()

	if _, err := os.Stat(*log_fn); err != nil {
		if os.IsNotExist(err) {
			if err := os.WriteFile(*log_fn, []byte(""), 0644); err != nil {
				fmt.Println("Error opening/creating chat log file: ", err)
			}
		} else {
			fmt.Println("error checking file existence: %w", err)
		}
	}

	// log, err := os.OpenFile(*log_fn, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	// if err != nil {
	// 	fmt.Println("Error opening/creating chat log file: ", err)
	// 	fmt.Println("CHAT WILL NOT BE SAVED (but we're forging on)")
	// }
	// defer log.Close()

	var prompt_context []LLM.LLM_Conversations
	if context != nil {
		prompt_context, _ = LLM.Last_n_Chats(log_fn, *context)
	}

	var prompt string
	var err error
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

	client_args := LLM.Client_Args{
		Prompt:     prompt,
		Context:    prompt_context,
		Max_Tokens: *max_tokens,
		Log:        log_fn,
	}

	chat_with_llm(*model, client_args)
}
