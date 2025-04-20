package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/spf13/pflag"

	"github.com/duluk/ask-ai/pkg/LLM"
	"github.com/duluk/ask-ai/pkg/config"
	"github.com/duluk/ask-ai/pkg/database"
	"github.com/duluk/ask-ai/pkg/linewrap"
	"github.com/duluk/ask-ai/pkg/logger"
	"github.com/duluk/ask-ai/pkg/tui"
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

	err = logger.Initialize(logger.Config{
		// Level: slog.LevelDebug,
		Level:      slog.LevelInfo,
		Format:     "json",
		FilePath:   opts.LogFileName,
		MaxSize:    10, // MB
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
		UseConsole: false,
	})
	if err != nil {
		fmt.Println("Error initializing logger: ", err)
		os.Exit(1)
	}

	// If DB exists, just opens it; otherwise, creates it first
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
		logger.Debug("Conversation ID provided", "convID", convID)
	} else if opts.ContinueChat {
		convID, err = db.GetLastConversationID()
		if err != nil {
			// convID will remain 0
			fmt.Println("Error getting last conversation ID: ", err)
		}
		logger.Debug("Continuing last conversation", "convID", convID)
	}

	// Make sure we are using the correct conversation id when not provided
	if convID == 0 {
		// This is a 'new' chat, not continued from previous
		convID, _ = db.GetLastConversationID()
		convID++
		logger.Debug("New conversation ID", "convID", convID)
	} else {
		// Either opts.id or opts.continue was used (determined above); we need
		// to load the context from the convID
		promptContext, err = db.LoadConversationFromDB(convID)
		logger.Debug("Loaded conversation from DB", "convID", convID)
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
		Thinking:     &opts.Thinking,
		// Log:          log_fd,
	}

	// If TUI mode is enabled, start the TUI
	if opts.UseTUI {
		// Ensure we don't output directly to terminal when in TUI mode
		opts.NoOutput = true
		logger.Info("Starting TUI")
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
	model := *args.Model
	// Parse model spec of form "provider/modelKey" or just "modelKey"
	spec := model
	provider := opts.Provider
	modelKey := spec
	// Check for explicit provider in spec
	if idx := strings.Index(spec, "/"); idx >= 0 {
		provider = spec[:idx]
		modelKey = spec[idx+1:]
		opts.Provider = provider
		// Use only the model key for recording
		model = modelKey
	} else {
		// Infer provider from modelKey if unambiguous
		var matches []string
		for provName, prov := range opts.Config.Models {
			// Direct model key match
			if _, ok := prov.Models[modelKey]; ok {
				matches = append(matches, provName)
			} else {
				// Alias match
				for _, mConf := range prov.Models {
					if slices.Contains(mConf.Aliases, modelKey) {
						matches = append(matches, provName)
					}
				}
			}
		}
		if len(matches) == 1 {
			provider = matches[0]
			opts.Provider = provider
		} else if len(matches) > 1 {
			panic(fmt.Sprintf("model %q is ambiguous across providers: %v", modelKey, matches))
		}
	}

	modelConf, err := config.GetModelConfig(opts.Config, provider, modelKey)
	if err != nil {
		fmt.Printf("Model %q not found for provider %q\n", modelKey, provider)
		os.Exit(1)
	}

	// Override args with API-specific values
	apiModel := modelConf.ModelName
	args.Model = &apiModel
	apiTemp := float32(modelConf.Temperature)
	args.Temperature = &apiTemp
	apiMax := modelConf.MaxTokens
	args.MaxTokens = &apiMax
	logger.Info("Processing prompt", "->", *args.Prompt, "convID", *args.ConvID)
	logger.Info("Using model", "provider", provider, "model", model, "temperature", apiTemp, "maxTokens", apiMax)

	var client LLM.Client

	switch opts.Provider {
	case "openai":
		apiURL := "https://api.openai.com/v1/"
		client = LLM.NewOpenAI("openai", apiURL)
	case "claude", "anthropic":
		client = LLM.NewAnthropic()
	case "gemini", "google":
		client = LLM.NewGoogle()
	case "ollama":
		client = LLM.NewOllama()
	case "grok", "xai":
		apiURL := "https://api.x.ai/v1/"
		client = LLM.NewOpenAI("xai", apiURL)
	default:
		fmt.Printf("Unknown provider: %q\n", provider)
		os.Exit(1)
	}

	if !opts.Quiet {
		fmt.Println("Assistant: ")
	}

	// Send the chat request and start streaming responses
	resp, streamChan, err := client.Chat(args, opts.ScreenTextWidth, opts.TabWidth)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
	// Spinner while waiting for the model to respond
	spinnerActive := !opts.Quiet && !opts.NoOutput
	var spinnerDone chan struct{}
	var spinnerAck chan struct{}
	if spinnerActive {
		spinnerDone = make(chan struct{})
		spinnerAck = make(chan struct{})
		go func() {
			chars := []rune{'|', '/', '-', '\\'}
			i := 0
			for {
				select {
				case <-spinnerDone:
					// clear spinner line
					fmt.Printf("\r\033[K")
					close(spinnerAck)
					return
				default:
					fmt.Printf("\rWaiting for response... %c", chars[i%len(chars)])
					time.Sleep(100 * time.Millisecond)
					i++
				}
			}
		}()
	}

	// Prepare line wrapper for streaming output
	lw := linewrap.NewLineWrapper(opts.ScreenTextWidth, opts.TabWidth, os.Stdout)

	// Collect the full response while printing chunks
	fullResponse := ""
	// Stop spinner on first chunk and wait for it to clear the line
	spinnerStopped := false
	for chunk := range streamChan {
		if spinnerActive && !spinnerStopped {
			// signal spinner to stop and await its acknowledgment
			close(spinnerDone)
			<-spinnerAck
			spinnerStopped = true
		}
		if chunk.Error != nil {
			fmt.Println("Error: ", chunk.Error)
			os.Exit(1)
		}
		lw.Write([]byte(chunk.Content))
		fullResponse += chunk.Content
	}

	if !opts.Quiet {
		fmt.Printf("\n\n-%s (convID: %d)\n", model, *args.ConvID)
	}

	if !opts.NoRecord {
		err = db.InsertConversation(
			*args.Prompt,
			fullResponse,
			model,
			*args.Temperature,
			resp.InputTokens,
			resp.OutputTokens,
			*args.ConvID,
		)
		if err != nil {
			fmt.Println("error inserting conversation into database: ", err)
		}
		logger.Debug("Inserted conversation into database", "convID", *args.ConvID)
		logger.Debug("Usage stats from model", "inputTokens", resp.InputTokens, "outputTokens", resp.OutputTokens)
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
