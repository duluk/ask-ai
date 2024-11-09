package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"

	"github.com/duluk/ask-ai/LLM"
)

// I'm probably writing "Ruby Go"...

func chat_with_llm(model string, args LLM.Client_Args) {
	var client LLM.Client

	switch model {
	case "chatgpt":
		client = LLM.New_OpenAI(args.Max_Tokens)
	case "claude":
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

	model := flag.String("model", "claude", "Which LLM to use (claude|chatgpt|gemini)")
	log_fn := flag.String("log", HOME+"/.config/ask-ai/ask-ai.chat.log", "Chat log file")
	context := flag.Int("context", 0, "Use n previous messages for context")
	max_tokens := flag.Int("max-tokens", 4096, "Maximum tokens to generate")

	flag.Parse()

	log, err := os.OpenFile(*log_fn, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error opening/creating chat log file: ", err)
		fmt.Println("CHAT WILL NOT BE SAVED (but we're forging on)")
	}
	defer log.Close()

	var ostr *LLM.Output_Stream
	ostr = LLM.New_Output_Stream()
	ostr.Add_Stream("stdout", os.Stdout)
	if log != nil {
		ostr.Add_Stream("log", log)
	}

	var prompt_context []string
	if context != nil {
		data, err := os.ReadFile(*log_fn)
		if err != nil {
			fmt.Println("File reading error", err)
			return
		}

		rx := regexp.MustCompile(`(?m)^\s*<------>\s*$`)
		records := rx.Split(string(data), -1)

		prompt_context = records[len(records)-*context-1:]
	}

	var prompt string
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
	ostr.Printf("User: " + prompt + "\n\n")

	client_args := LLM.Client_Args{
		Prompt:     prompt,
		Context:    prompt_context,
		Max_Tokens: *max_tokens,
		Out:        ostr,
	}

	chat_with_llm(*model, client_args)

	ostr.Printf("\n\n<model - %s>\n<------>\n", *model)
}
