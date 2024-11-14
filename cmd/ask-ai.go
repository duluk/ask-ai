package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/duluk/ask-ai/LLM"
)

const version = "0.3.0"

var (
	commit string = "Unknown"
	date   string = "Unknown"
)

// I'm probably writing "Ruby Go"...

func main() {
	var err error
	HOME := os.Getenv("HOME")

	// TODO: I think it's time to take all this out and create an app_args
	// object or structure. Maybe. Maybe it's fine...
	model := pflag.StringP("model", "m", "claude", "Which LLM to use (claude|chatgpt|gemini)")
	context := pflag.IntP("context", "n", 0, "Use previous n messages for context")
	cfg_file := pflag.StringP("config", "C", "", "Configuration file")
	continue_chat := pflag.BoolP("continue", "c", false, "Continue previous conversation")
	show_version := pflag.BoolP("version", "v", false, "Print version and exit")
	show_full_version := pflag.BoolP("full-version", "V", false, "Print full version information and exit")

	pflag.StringP("log", "l", HOME+"/.config/ask-ai/ask-ai.chat.yml", "Chat log file")
	pflag.StringP("system-prompt", "s", "", "System prompt for LLM")
	pflag.IntP("max-tokens", "t", 4096, "Maximum tokens to generate")
	pflag.Float32P("temperature", "T", 0.7, "Temperature for generation")

	pflag.Parse()

	if *show_version {
		fmt.Println("ask-ai version: ", version)
		os.Exit(0)
	}

	if *show_full_version {
		fmt.Println("Version: ", version)
		fmt.Println("Commit:  ", commit)
		fmt.Println("Date:    ", date)
		os.Exit(0)
	}

	if *cfg_file != "" {
		// Validation will happen below with ReadInConfig()
		viper.SetConfigFile(*cfg_file)
	} else {
		viper.SetConfigName("config.yml")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(HOME + "/.config/ask-ai")
		viper.AddConfigPath(".")
	}

	if err = viper.ReadInConfig(); err != nil {
		// TODO: I don't know if panicking in this situation is correct; could
		// just continue with defaults
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	}

	// if viper.ConfigFileUsed() != "" {
	// 	fmt.Println("Using configuration file: ", viper.ConfigFileUsed())
	// }

	viper.BindPFlag("model.max_tokens", pflag.Lookup("max-tokens"))
	viper.BindPFlag("log.file", pflag.Lookup("log"))
	viper.BindPFlag("model.temperature", pflag.Lookup("temperature"))
	viper.BindPFlag("model.system_prompt", pflag.Lookup("system-prompt"))

	// Get configuration values (potentially overridden by flags)
	log_fn := os.ExpandEnv(viper.GetString("log.file"))
	system_prompt := viper.GetString("model.system_prompt")
	max_tokens := viper.GetInt("model.max_tokens")
	temperature := float32(viper.GetFloat64("model.temperature"))

	var log_fd *os.File
	log_fd, err = os.OpenFile(log_fn, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error with chat log file: ", err)
	}
	defer log_fd.Close()

	var prompt_context []LLM.LLM_Conversations
	if *continue_chat {
		prompt_context, err = LLM.Continue_Conversation(log_fd)
		if err != nil {
			fmt.Println("Error reading log for continuing chat: ", err)
		}
	}
	if *context != 0 {
		prompt_context, err = LLM.Last_n_Chats(log_fd, *context)
		if err != nil {
			fmt.Println("Error loading chat context from log: ", err)
		}
	}

	var prompt string
	if pflag.NArg() > 0 {
		prompt = pflag.Arg(0)
	} else {
		// TODO: Make the prompt, `*model+"> "` and remove 'Using model...'
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
		Prompt:        &prompt,
		System_Prompt: &system_prompt,
		Context:       prompt_context,
		Max_Tokens:    &max_tokens,
		Temperature:   &temperature,
		Log:           log_fd,
	}

	chat_with_llm(*model, client_args, *continue_chat)
}

func chat_with_llm(model string, args LLM.Client_Args, continue_chat bool) {
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
	LLM.Log_Chat(log, "User", *args.Prompt, "", continue_chat)

	fmt.Printf("Assistant: ")
	resp, err := client.Chat(args)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
	fmt.Printf("\n\n-%s\n", model)

	LLM.Log_Chat(log, "Assistant", resp, model, continue_chat)
}
