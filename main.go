package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/duluk/ask-ai/LLM"
)

// I'm probably writing "Ruby Go"...

func chat_with_llm(model string, args LLM.Client_Args) {
	var client LLM.Client
	log := args.Log

	switch model {
	case "chatgpt":
		client = LLM.New_OpenAI(*args.Max_Tokens)
	case "claude":
		client = LLM.New_Anthropic(*args.Max_Tokens)
	case "gemini":
		client = LLM.New_Google(*args.Max_Tokens)
	default:
		fmt.Println("Unknown model: ", model)
		os.Exit(1)
	}
	LLM.Log_Chat(log, "User", *args.Prompt, "")

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

	model := pflag.StringP("model", "m", "claude", "Which LLM to use (claude|chatgpt|gemini)")
	context := pflag.IntP("context", "n", 0, "Use previous n messages for context")
	pflag.StringP("log", "l", HOME+"/.config/ask-ai/ask-ai.chat.yml", "Chat log file")
	pflag.IntP("max-tokens", "t", 4096, "Maximum tokens to generate")
	pflag.StringP("config", "c", HOME+"/.config/ask-ai/config", "Configuration file")
	pflag.Float64P("temperature", "T", 0.7, "Temperature for generation")

	pflag.Parse()

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath(HOME + "/.config/ask-ai")
	viper.ReadInConfig()

	viper.BindPFlag("model.max_tokens", pflag.Lookup("max-tokens"))
	viper.BindPFlag("log.file", pflag.Lookup("log"))
	viper.BindPFlag("model.temperature", pflag.Lookup("temperature"))

	// Get configuration values (potentially overridden by flags)
	log_fn := viper.GetString("log.file")
	max_tokens := viper.GetInt("model.max_tokens")
	temperature := viper.GetFloat64("model.temperature")

	if _, err := os.Stat(log_fn); err != nil {
		if os.IsNotExist(err) {
			if err := os.WriteFile(log_fn, []byte(""), 0644); err != nil {
				fmt.Println("Error opening/creating chat log file: ", err)
			}
		} else {
			fmt.Println("error checking file existence: %w", err)
		}
	}

	var prompt_context []LLM.LLM_Conversations
	if context != nil {
		prompt_context, _ = LLM.Last_n_Chats(&log_fn, *context)
	}

	var prompt string
	var err error
	if pflag.NArg() > 0 {
		prompt = pflag.Arg(0)
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
		Prompt:      &prompt,
		Context:     prompt_context,
		Max_Tokens:  &max_tokens,
		Temperature: &temperature,
		Log:         &log_fn,
	}

	chat_with_llm(*model, client_args)
}
