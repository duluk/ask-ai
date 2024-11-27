package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/spf13/pflag"

	"github.com/duluk/ask-ai/pkg/LLM"
	"github.com/duluk/ask-ai/pkg/config"
	"github.com/duluk/ask-ai/pkg/database"
)

// I'm probably writing "Ruby Go"...

func main() {
	var err error

	opts, err := config.Initialize()
	if err != nil {
		fmt.Println("Error initializing config: ", err)
		os.Exit(1)
	}

	if opts.DumpConfig {
		config.DumpConfig(opts)
	}

	var log_fd *os.File
	log_fd, err = os.OpenFile(opts.LogFileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error with chat log file: ", err)
	}
	defer log_fd.Close()

	/* CONTEXT? LOAD IT */
	var promptContext []LLM.LLMConversations
	if opts.ContinueChat {
		promptContext, err = LLM.ContinueConversation(log_fd)
		if err != nil {
			fmt.Println("Error reading log for continuing chat: ", err)
		}
	}
	if opts.Context != 0 {
		promptContext, err = LLM.LastNChats(log_fd, opts.Context)
		if err != nil {
			fmt.Println("Error loading chat context from log: ", err)
		}
	}

	// If DB exists, it just opens it; otherwise, it creates it first
	db, err := database.NewDB(opts.DBFileName, opts.DBTable)
	if err != nil {
		fmt.Println("Error opening database: ", err)
		os.Exit(1)
	}
	defer db.Close()

	clientArgs := LLM.ClientArgs{
		Model:        &opts.Model,
		SystemPrompt: &opts.SystemPrompt,
		Context:      promptContext,
		MaxTokens:    &opts.MaxTokens,
		Temperature:  &opts.Temperature,
		Log:          log_fd,
	}

	/* GET THE PROMPT */
	var prompt string
	if pflag.NArg() > 0 {
		prompt = pflag.Arg(0)
		clientArgs.Prompt = &prompt

		chatWithLLM(clientArgs, opts.ContinueChat, db)
	} else {
		// Gracefully handle CTRL-C interrupt signal
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)
		go func() {
			<-sig
			fmt.Println("\nGoodbye!")
			os.Exit(0)
		}()

		for {
			prompt = getPromptFromUser(opts.Model)
			clientArgs.Prompt = &prompt

			chatWithLLM(clientArgs, opts.ContinueChat, db)

			opts.ContinueChat = true
			promptContext, err = LLM.ContinueConversation(log_fd)
			if err != nil {
				fmt.Println("Error reading log for continuing chat: ", err)
			}
			clientArgs.Context = promptContext
		}
	}
}

func chatWithLLM(args LLM.ClientArgs, continueChat bool, db *database.ChatDB) {
	var client LLM.Client
	log := args.Log
	model := *args.Model

	switch model {
	case "chatgpt":
		api_url := "https://api.openai.com/v1/"
		client = LLM.NewOpenAI("openai", api_url)
	case "claude":
		client = LLM.NewAnthropic()
	case "gemini":
		client = LLM.NewGoogle()
	case "grok":
		api_url := "https://api.x.ai/v1/"
		client = LLM.NewOpenAI("xai", api_url)
	default:
		fmt.Println("Unknown model: ", model)
		os.Exit(1)
	}
	LLM.LogChat(
		log,
		"User",
		*args.Prompt,
		"",
		continueChat,
		LLM.EstimateTokens(*args.Prompt),
		0,
	)

	fmt.Println("Assistant: ")
	resp, err := client.Chat(args)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
	fmt.Printf("\n\n-%s\n", model)

	// If we want the timestamp in the log and in the database to match
	// exactly, we can set it here and pass it in to LogChat and
	// InsertConversation. As it stands, each function uses the current
	// timestamp when the function is executed.

	LLM.LogChat(log, "Assistant", resp.Text, model, continueChat, resp.InputTokens, resp.OutputTokens)
	err = db.InsertConversation(*args.Prompt, resp.Text, model, *args.Temperature)
	if err != nil {
		fmt.Println("error inserting conversation into database: ", err)
	}
}

func getPromptFromUser(model string) string {
	fmt.Printf("%s> ", model)
	reader := bufio.NewReader(os.Stdin)
	prompt, err := reader.ReadString('\n')
	if err != nil {
		if err.Error() == "EOF" {
			fmt.Println("\nGoodbye!")
			os.Exit(0)
		}
		fmt.Println("Error reading prompt: ", err)
		os.Exit(1)
	}

	// Now clean up spaces and remove the newline we just captured
	prompt = strings.TrimSpace(prompt)
	prompt = prompt[:len(prompt)-1]

	return prompt
}
