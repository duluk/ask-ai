package tui

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/duluk/ask-ai/pkg/LLM"
	"github.com/duluk/ask-ai/pkg/config"
	"github.com/duluk/ask-ai/pkg/database"
	"github.com/duluk/ask-ai/pkg/linewrap"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TUI represents the terminal user interface for ask-ai
type TUI struct {
	app            *tview.Application
	outputBox      *tview.TextView
	inputField     *tview.TextArea
	statusBar      *tview.TextView
	opts           *config.Options
	db             *database.ChatDB
	log            *os.File
	model          string
	conversationID int
	client         LLM.Client
	messages       []LLM.LLMConversations
	uiMutex        sync.Mutex
	buffer         strings.Builder
	responseLines  int
}

// NewTUI creates a new TUI instance
func NewTUI(opts *config.Options, db *database.ChatDB, log *os.File) *TUI {
	app := tview.NewApplication()
	outputBox := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		// SetWrap(true).
		// SetWordWrap(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			app.Draw()
		})
	outputBox.SetBorder(true).
		SetTitle("Conversation").
		SetBorderColor(tcell.ColorBlue)

	inputField := tview.NewTextArea().
		SetPlaceholder("Type your message here...").
		SetPlaceholderStyle(tcell.StyleDefault.Foreground(tcell.ColorGray)).
		SetWordWrap(true)
	inputField.SetBorder(true).
		SetTitle("Your message").
		SetBorderColor(tcell.ColorGreen)
	inputField.SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorLightCyan))

	// We'll set the change handler after tui is initialized

	statusBar := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	statusBar.SetBackgroundColor(tcell.ColorBlue)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(outputBox, 0, 1, false).
		AddItem(statusBar, 1, 0, false).
		AddItem(inputField, 3, 0, true)

	tui := &TUI{
		app:            app,
		outputBox:      outputBox,
		inputField:     inputField,
		statusBar:      statusBar,
		opts:           opts,
		db:             db,
		log:            log,
		model:          opts.Model,
		conversationID: 0,
		buffer:         strings.Builder{},
		responseLines:  0,
	}

	// Set up key bindings
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc || (event.Key() == tcell.KeyCtrlC) {
			app.Stop()
			return nil
		}

		if event.Key() == tcell.KeyEnter {
			if inputField.HasFocus() {
				text := inputField.GetText()
				if text == "" {
					return nil
				}

				if event.Modifiers()&tcell.ModShift != 0 {
					// Shift+Enter - add newline
					inputField.SetText(text+"\n", true)
					return nil
				} else {
					queryText := text
					inputField.SetText("", false)

					go func(t *TUI) {
						time.Sleep(50 * time.Millisecond)
						defer func() {
							if r := recover(); r != nil {
								t.app.QueueUpdateDraw(func() {
									t.appendText(fmt.Sprintf("\nError: Recovered from panic: %v\n", r))
									t.updateStatusBar()
									t.app.SetFocus(t.inputField)
								})
							}
						}()

						t.updateStatusBar("Processing your request...")
						t.sendQuery(queryText)
					}(tui)
					return nil
				}
			}
		}
		return event
	})

	// Set the root primitive
	app.SetRoot(flex, true)

	return tui
}

// Initialize sets up the TUI, creates a client based on the model, and displays the welcome message
func (t *TUI) Initialize() error {
	// Initialize the client based on the model
	switch t.model {
	case "chatgpt":
		apiURL := "https://api.openai.com/v1/"
		t.client = LLM.NewOpenAI("openai", apiURL)
	case "claude":
		t.client = LLM.NewAnthropic()
	case "gemini":
		t.client = LLM.NewGoogle()
	case "grok":
		apiURL := "https://api.x.ai/v1/"
		t.client = LLM.NewOpenAI("xai", apiURL)
	case "deepseek":
		apiURL := "https://api.deepseek.com/v1/"
		t.client = LLM.NewOpenAI("deepseek", apiURL)
	case "ollama":
		t.client = LLM.NewOllama()
	default:
		return fmt.Errorf("unknown model: %s", t.model)
	}

	// Initialize a new conversation
	convID := LLM.FindLastConversationID(t.log)
	if convID == nil {
		t.conversationID = 1
	} else {
		t.conversationID = *convID + 1
	}

	// Set up a simple welcome message
	t.outputBox.SetText("Welcome to Ask AI Terminal UI\nType your question below and press Enter\n\n")

	// Set a simple status bar
	t.statusBar.SetText(fmt.Sprintf(" Model: %s | Conversation ID: %d | Press ESC to quit", t.model, t.conversationID))

	// Make sure input field has focus but wait until UI starts
	go func() {
		time.Sleep(100 * time.Millisecond) // Wait for UI to be ready
		t.app.QueueUpdateDraw(func() {
			t.app.SetFocus(t.inputField)
		})
	}()

	return nil
}

// Run starts the TUI application with proper signal handling
func (t *TUI) Run() error {
	// Create channel for signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Go back to the original implementation with goroutine
	// This was working before our changes
	doneChan := make(chan struct{}, 1)

	// Run the application in a goroutine
	go func() {
		if err := t.app.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error in TUI: %v\n", err)
		}
		doneChan <- struct{}{}
	}()

	// Handle signals
	select {
	case <-doneChan:
		return nil
	case <-sigChan:
		t.app.Stop()
		return nil
	}
}

// handleQuery is a wrapper that calls sendQuery but with additional error handling
func (t *TUI) handleQuery(query string) {
	defer func() {
		if r := recover(); r != nil {
			t.app.QueueUpdateDraw(func() {
				t.appendText(fmt.Sprintf("\nError: Fatal error processing query: %v\n", r))
				t.updateStatusBar("Ready")
				t.app.SetFocus(t.inputField)
			})
		}
	}()

	// Call the actual query processing function
	t.sendQuery(query)
}

func (t *TUI) Cleanup() {
	t.uiMutex.Lock()
	defer t.uiMutex.Unlock()

	if t.app != nil {
		t.app.Stop()
	}
	if t.db != nil {
		t.db.Close()
	}
	if t.log != nil {
		t.log.Close()
	}
}

func (t *TUI) sendQuery(query string) {
	// Handle commands starting with /
	if strings.HasPrefix(query, "/") {
		cmd := strings.Split(query, " ")[0]
		switch cmd {
		case "/model":
			if len(query) > 6 && query[7:] != "" {
				newModel := query[7:]
				if !t.validateModel(newModel) {
					t.appendText(fmt.Sprintf("\nError: Invalid model '%s'. Valid models are: chatgpt, claude, gemini, grok, deepseek, ollama\n", newModel))
					t.updateStatusBar()
					return
				}
				t.model = newModel
				// Reinitialize client with new model
				switch t.model {
				case "chatgpt":
					apiURL := "https://api.openai.com/v1/"
					t.client = LLM.NewOpenAI("openai", apiURL)
				case "claude":
					t.client = LLM.NewAnthropic()
				case "gemini":
					t.client = LLM.NewGoogle()
				case "grok":
					apiURL := "https://api.x.ai/v1/"
					t.client = LLM.NewOpenAI("xai", apiURL)
				case "deepseek":
					apiURL := "https://api.deepseek.com/v1/"
					t.client = LLM.NewOpenAI("deepseek", apiURL)
				case "ollama":
					t.client = LLM.NewOllama()
				}
				t.appendText(fmt.Sprintf("\nSwitched to model: %s\n", t.model))
				t.updateStatusBar()
			} else {
				t.appendText(fmt.Sprintf("\nCurrent model: %s\n", t.model))
			}
			return
		}
	}

	// Prepare args for LLM
	args := LLM.ClientArgs{
		Model:        &t.model,
		Prompt:       &query,
		SystemPrompt: &t.opts.SystemPrompt,
		Context:      t.messages,
		MaxTokens:    &t.opts.MaxTokens,
		Temperature:  &t.opts.Temperature,
		Log:          t.log,
		ConvID:       &t.conversationID,
	}

	// Show initial AI prompt
	t.appendText("AI:\n")

	// Create context for streaming
	ctx, cancelStream := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancelStream()

	var fullResponse string
	var streamErr error

	// Create buffered channel for stream completion
	streamDone := make(chan bool, 1)

	// Define callback for chunks
	chunkCallback := func(chunk string) {
		fullResponse += chunk
		t.appendTextNoNewline(chunk)
	}

	// Start streaming in goroutine
	go func() {
		defer func() {
			streamDone <- true
		}()

		resp, err := t.client.StreamChat(args, t.opts.ScreenWidth, t.opts.TabWidth, chunkCallback)
		if err != nil {
			streamErr = err
			return
		}

		// Record response in database and log
		cleanResponse := fullResponse
		LLM.LogChat(
			t.log,
			"Assistant",
			cleanResponse,
			t.model,
			true,
			resp.InputTokens,
			resp.OutputTokens,
			t.conversationID,
		)

		_ = t.db.InsertConversation(
			query,
			cleanResponse,
			t.model,
			*args.Temperature,
			resp.InputTokens,
			resp.OutputTokens,
			t.conversationID,
		)
	}()

	// Wait for completion or timeout
	select {
	case <-streamDone:
		if streamErr != nil {
			t.appendText(fmt.Sprintf("\nError: %s\n", streamErr))
		} else {
			t.appendText("\n\n")
		}
	case <-ctx.Done():
		t.appendText("\nRequest timed out or was cancelled\n")
	}

	// Reset UI state
	t.updateStatusBar()
	t.inputField.SetText("", false)
	t.app.SetFocus(t.inputField)
}

func (t *TUI) validateModel(model string) bool {
	validModels := []string{"chatgpt", "claude", "gemini", "grok", "deepseek", "ollama"}

	return slices.Contains(validModels, model)
}

// appendText adds text to the output box and scrolls to the bottom
func (t *TUI) appendText(text string) {
	t.app.QueueUpdateDraw(func() {
		t.outputBox.Write([]byte(text))
		t.outputBox.ScrollToEnd()
	})
}

// appendTextNoNewline adds text without adding a newline and scrolls to the bottom
func (t *TUI) appendTextNoNewline(text string) {
	if strings.Contains(text, "\n") {
		// Handle newlines by splitting the text
		lines := strings.Split(text, "\n")

		// Handle first part with buffer
		t.buffer.WriteString(lines[0])

		t.app.QueueUpdateDraw(func() {
			if t.responseLines == 0 {
				t.outputBox.Write([]byte(t.buffer.String()))
				t.responseLines++
			} else {
				// Update last line with buffer content
				currentContent := t.outputBox.GetText(true)
				lines := strings.Split(currentContent, "\n")
				lines[len(lines)-1] = lines[len(lines)-1] + t.buffer.String()
				t.outputBox.SetText(strings.Join(lines, "\n"))
			}
			t.outputBox.ScrollToEnd()
		})

		// Reset buffer
		t.buffer.Reset()

		// Handle remaining lines
		for i := 1; i < len(lines); i++ {
			t.outputBox.Write([]byte(lines[i]))
			t.responseLines++
		}
	} else {
		// Add to buffer
		t.buffer.WriteString(text)

		// Check if we should flush buffer (on sentence endings or length)
		if strings.ContainsAny(t.buffer.String(), ".!?") || t.buffer.Len() >= 30 {
			var wrappedBuffer bytes.Buffer
			wrapper := linewrap.NewLineWrapper(t.opts.ScreenWidth, t.opts.TabWidth, &wrappedBuffer)
			wrapper.Write([]byte(t.buffer.String()))
			wrappedText := wrappedBuffer.String()

			t.app.QueueUpdateDraw(func() {
				if t.responseLines == 0 {
					t.outputBox.Write([]byte(wrappedText))
					t.responseLines++
				} else {
					// Update last line
					currentContent := t.outputBox.GetText(true)
					lines := strings.Split(currentContent, "\n")
					lastLine := lines[len(lines)-1]

					if len(lastLine)+len(wrappedText) > t.opts.ScreenWidth {
						// Start new line
						t.outputBox.Write([]byte("\n" + wrappedText))
						t.responseLines++
					} else {
						// Append to current line
						lines[len(lines)-1] = lastLine + wrappedText
						t.outputBox.SetText(strings.Join(lines, "\n"))
					}
				}
				t.outputBox.ScrollToEnd()
			})

			// Reset buffer
			t.buffer.Reset()
		}
	}
}

// updateStatusBar updates the status bar with current information
func (t *TUI) updateStatusBar(status ...string) {
	t.app.QueueUpdateDraw(func() {
		// Status bar update logic
		text := fmt.Sprintf(" Model: %s | Conversation ID: %d", t.model, t.conversationID)
		if len(status) > 0 {
			text = status[0]
		}
		t.statusBar.SetText(text)
	})
}

// SetDatabase sets the database for the TUI
func (t *TUI) SetDatabase(db *database.ChatDB) {
	t.db = db
}

// SetLogFile sets the log file for the TUI
func (t *TUI) SetLogFile(log *os.File) {
	t.log = log
}
