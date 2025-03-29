package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"

	"github.com/duluk/ask-ai/pkg/LLM"
	"github.com/duluk/ask-ai/pkg/config"
	"github.com/duluk/ask-ai/pkg/database"
	"github.com/duluk/ask-ai/pkg/tui" // Add this import
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

	err = os.MkdirAll(filepath.Dir(opts.LogFileName), 0o755)
	if err != nil {
		fmt.Println("Error creating log directory: ", err)
		os.Exit(1)
	}

	var log_fd *os.File
	log_fd, err = os.OpenFile(opts.LogFileName, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		fmt.Println("Error with chat log file: ", err)
	}
	defer log_fd.Close()

	err = logger.Initialize(logger.Config{
		Level:      slog.LevelInfo,
		Format:     "json",
		FilePath:   opts.LogFileName,
		MaxSize:    10, // MB
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
		UseConsole: false,
	})

	// If DB exists, it just opens it; otherwise, it creates it first
	db, err := database.InitializeDB(opts.DBFileName, opts.DBTable)
	if err != nil {
		fmt.Println("Error opening database: ", err)
		os.Exit(1)
	}
	defer db.Close()

	model := opts.Model

	/* CONTEXT? LOAD IT */
	var convID int
	var promptContext []LLM.LLMConversations
	if opts.ConversationID != 0 {
		convID = opts.ConversationID
	} else if opts.ContinueChat {
		convID, err = db.GetLastConversationID()
		if err != nil {
			// convID will remain 0
			fmt.Println("Error getting last conversation ID: ", err)
		}
	}

	// Make sure we are using the correct conversation id when not provided
	if convID == 0 {
		// This is a 'new' chat, not continued from previous
		convID, _ = db.GetLastConversationID()
		convID++
	} else {
		// Either opts.id or opts.continue was used (determined above); we need
		// to load the context from the convID
		promptContext, err = db.LoadConversationFromDB(convID)
		if err != nil {
			fmt.Printf("error loading conversation from DB: %v", err)
		}

		if !pflag.CommandLine.Changed("model") && len(promptContext) > 0 {
			model = promptContext[len(promptContext)-1].Model
		}
	}

	clientArgs := LLM.ClientArgs{
		Model:        &model,
		SystemPrompt: &opts.SystemPrompt,
		ConvID:       &convID,
		Context:      promptContext,
		MaxTokens:    &opts.MaxTokens,
		Temperature:  &opts.Temperature,
		Log:          log_fd,
	}

	// If TUI mode is enabled, start the TUI
	if opts.UseTUI {
		// Ensure we don't output directly to terminal when in TUI mode
		opts.NoOutput = true
		err = tui.Run(opts, clientArgs, db)
		if err != nil {
			fmt.Println("Error running TUI: ", err)
			os.Exit(1)
		}
		return
	}

	/* GET THE PROMPT */
	var prompt string
	if pflag.NArg() > 0 {
		prompt = pflag.Arg(0)
		clientArgs.Prompt = &prompt

		chatWithLLM(opts, clientArgs, db)
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
			prompt = getPromptFromUser(model)
			if prompt[0] == '/' {
				cmd := strings.Split(prompt, " ")[0]
				switch cmd {
				case "/help", "/?":
					fmt.Println("Special commands:")
					fmt.Println("  /exit: Exit the program")
					fmt.Println("  /context: Show the current context")
					fmt.Println("  /model <model>: Show the current model")
					fmt.Println("  /id: Show the current conversation ID")
					continue
				case "/exit", "/quit":
					fmt.Println("Goodbye!")
					os.Exit(0)
				case "/context":
					fmt.Println("Context: ", promptContext)
					continue
				case "/model":
					if len(prompt) > 6 && prompt[7:] != "" {
						model = prompt[7:]
						clientArgs.Model = &model
					}
					continue
				case "/id":
					fmt.Println("Conversation ID: ", *clientArgs.ConvID)
					continue
				}
			}
			clientArgs.Prompt = &prompt

			chatWithLLM(opts, clientArgs, db)

			opts.ContinueChat = true
			// TODO: this needs to come from the DB
			// promptContext, err = LLM.ContinueConversation(log_fd)
			promptContext, err = db.LoadConversationFromDB(*clientArgs.ConvID)
			if err != nil {
				fmt.Println("Error reading log for continuing chat: ", err)
			}
			// TODO: promptContext will be nil if err != nil above. That's
			// probably what we want. Would write a test but not sure how to
			// test the LLM functions without using tokens.
			clientArgs.Context = promptContext
		}
	}
}

func chatWithLLM(opts *config.Options, args LLM.ClientArgs, db *database.ChatDB) {
	var client LLM.Client
	log := args.Log
	model := *args.Model
	continueChat := opts.ContinueChat

	switch model {
	case "chatgpt":
		api_url := "https://api.openai.com/v1/"
		client = LLM.NewOpenAI("openai", api_url)
	// case "claude":
	// 	client = LLM.NewAnthropic()
	// case "gemini":
	// 	client = LLM.NewGoogle()
	// case "grok":
	// 	api_url := "https://api.x.ai/v1/"
	// 	client = LLM.NewOpenAI("xai", api_url)
	// case "deepseek":
	// 	api_url := "https://api.deepseek.com/v1/"
	// 	client = LLM.NewOpenAI("deepseek", api_url)
	// 	// client = LLM.NewDeepSeek()
	case "ollama":
		client = LLM.NewOllama()
	default:
		fmt.Println("Unknown model: ", model)
		os.Exit(1)
	}

	if !opts.NoRecord {
		LLM.LogChat(
			log,
			"User",
			*args.Prompt,
			"",
			continueChat,
			LLM.EstimateTokens(*args.Prompt),
			0,
			*args.ConvID,
		)
	}

	if !opts.Quiet {
		fmt.Println("Assistant: ")
	}

	resp, streamChan, err := client.Chat(args, opts.ScreenTextWidth, opts.TabWidth)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}

	// Collect the full response while printing chunks
	fullResponse := ""
	for chunk := range streamChan {
		if chunk.Error != nil {
			fmt.Println("Error: ", chunk.Error)
			os.Exit(1)
		}
		if !opts.Quiet {
			fmt.Print(chunk.Content)
		}
		fullResponse += chunk.Content
	}

	if !opts.Quiet {
		fmt.Printf("\n\n-%s (convID: %d)\n", model, *args.ConvID)
	}

	// If we want the timestamp in the log and in the database to match
	// exactly, we can set it here and pass it in to LogChat and
	// InsertConversation. As it stands, each function uses the current
	// timestamp when the function is executed.

	if !opts.NoRecord {
		LLM.LogChat(
			log,
			"Assistant",
			resp.Text,
			model,
			continueChat,
			resp.InputTokens,
			resp.OutputTokens,
			*args.ConvID,
		)

		err = db.InsertConversation(
			*args.Prompt,
			resp.Text,
			model,
			*args.Temperature,
			resp.InputTokens,
			resp.OutputTokens,
			*args.ConvID,
		)
		if err != nil {
			fmt.Println("error inserting conversation into database: ", err)
		}
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
	if prompt[len(prompt)-1] == '\n' {
		prompt = prompt[:len(prompt)-1]
	}

	return prompt
}
