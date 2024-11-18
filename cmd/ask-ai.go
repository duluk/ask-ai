package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/duluk/ask-ai/pkg/LLM"
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

	/* HANDLE OPTIONS */
	// TODO: I think it's time to take all this out and create an app_args
	// object or structure. Maybe. Maybe it's fine...
	model := pflag.StringP("model", "m", "claude", "Which LLM to use (claude|chatgpt|gemini|grok)")
	context := pflag.IntP("context", "n", 0, "Use previous n messages for context")
	cfgFile := pflag.StringP("config", "C", "", "Configuration file")
	continueChat := pflag.BoolP("continue", "c", false, "Continue previous conversation")
	showVersion := pflag.BoolP("version", "v", false, "Print version and exit")
	showFullVersion := pflag.BoolP("full-version", "V", false, "Print full version information and exit")

	pflag.StringP("log", "l", HOME+"/.config/ask-ai/ask-ai.chat.yml", "Chat log file")
	pflag.StringP("system-prompt", "s", "", "System prompt for LLM")
	pflag.IntP("max-tokens", "t", 4096, "Maximum tokens to generate")
	pflag.Float32P("temperature", "T", 0.7, "Temperature for generation")

	pflag.Parse()

	if *showVersion {
		fmt.Println("ask-ai version: ", version)
		os.Exit(0)
	}

	if *showFullVersion {
		fmt.Println("Version: ", version)
		fmt.Println("Commit:  ", commit)
		fmt.Println("Date:    ", date)
		os.Exit(0)
	}

	if *cfgFile != "" {
		// Validation will happen below with ReadInConfig()
		viper.SetConfigFile(*cfgFile)
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
	logFileName := os.ExpandEnv(viper.GetString("log.file"))
	systemPrompt := viper.GetString("model.system_prompt")
	maxTokens := viper.GetInt("model.max_tokens")
	temperature := float32(viper.GetFloat64("model.temperature"))

	var log_fd *os.File
	log_fd, err = os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error with chat log file: ", err)
	}
	defer log_fd.Close()

	/* CONTEXT? LOAD IT */
	var promptContext []LLM.LLMConversations
	if *continueChat {
		promptContext, err = LLM.ContinueConversation(log_fd)
		if err != nil {
			fmt.Println("Error reading log for continuing chat: ", err)
		}
	}
	if *context != 0 {
		promptContext, err = LLM.LastNChats(log_fd, *context)
		if err != nil {
			fmt.Println("Error loading chat context from log: ", err)
		}
	}

	/* GET THE PROMPT */
	var prompt string
	if pflag.NArg() > 0 {
		prompt = pflag.Arg(0)
	} else {
		fmt.Printf("%s> ", *model)
		reader := bufio.NewReader(os.Stdin)
		prompt, err = reader.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				os.Exit(0)
			}
			fmt.Println("Error reading prompt: ", err)
			os.Exit(1)
		}
		fmt.Println()
	}

	clientArgs := LLM.ClientArgs{
		Model:        model,
		Prompt:       &prompt,
		SystemPrompt: &systemPrompt,
		Context:      promptContext,
		MaxTokens:    &maxTokens,
		Temperature:  &temperature,
		Log:          log_fd,
	}

	chatWithLLM(clientArgs, *continueChat)
}

func chatWithLLM(args LLM.ClientArgs, continueChat bool) {
	var client LLM.Client
	log := args.Log
	model := *args.Model

	switch model {
	case "chatgpt":
		api_url := "https://api.openai.com/v1/"
		client = LLM.NewOpenAI(*args.MaxTokens, "openai", api_url)
	case "claude":
		client = LLM.NewAnthropic(*args.MaxTokens)
	case "gemini":
		client = LLM.NewGoogle(*args.MaxTokens)
	case "grok":
		api_url := "https://api.x.ai/v1/"
		client = LLM.NewOpenAI(*args.MaxTokens, "xai", api_url)
	default:
		fmt.Println("Unknown model: ", model)
		os.Exit(1)
	}
	LLM.LogChat(log, "User", *args.Prompt, "", continueChat)

	fmt.Println("Assistant: ")
	resp, err := client.Chat(args)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
	fmt.Printf("\n\n-%s\n", model)

	LLM.LogChat(log, "Assistant", resp, model, continueChat)
}
