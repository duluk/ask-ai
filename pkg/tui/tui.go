package tui

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/duluk/ask-ai/pkg/LLM"
	"github.com/duluk/ask-ai/pkg/config"
	"github.com/duluk/ask-ai/pkg/database"
	"github.com/duluk/ask-ai/pkg/linewrap"
	"github.com/duluk/ask-ai/pkg/logger"
)

const (
	inputHeight    = 3 // Input box height in lines
	statusHeight   = 1 // Status line height in lines
	borderWidth    = 2 // Border width (left + right)
	borderHeight   = 2 // Border height (top + bottom)
	contentPadding = 2 // Padding inside components
	contentMargin  = 4 // Total horizontal margin for content (borders + padding)
	testPadding    = 0 // Padding for testing
)

const (
	lipColorBlack        = "0"
	lipColorDarkRed      = "1"
	lipColorRed          = "1"
	lipColorDarkGreen    = "2"
	lipColorGreen        = "2"
	lipColorDarkYellow   = "3"
	lipColorYellow       = "3"
	lipColorDarkBlue     = "4"
	lipColorDarkMagenta  = "5"
	lipColorMagenta      = "5"
	lipColorDarkCyan     = "6"
	lipColorCyan         = "6"
	lipColorLightGray    = "7"
	lipColorWhite        = "7"
	lipColorDarkGray     = "8"
	lipColorGray         = "8"
	lipColorLightRed     = "9"
	lipColorLightGreen   = "10"
	lipColorLightYellow  = "11"
	lipColorLightBlue    = "12"
	lipColorLightMagenta = "13"
	lipColorLightCyan    = "14"
	lipColorLightWhite   = "15"
	lipColorBlue         = "63"
)

var (
	outputBorderColor = lipgloss.Color(lipColorDarkBlue)
	inputBorderColor  = lipgloss.Color(lipColorGreen)

	// If using a system that doesn't have great support for fancy borders,
	// just use plain ASCII
	borderStyle = func() lipgloss.Border {
		term := os.Getenv("TERM")
		// tmux doesn't display fancy code points well
		if strings.Contains(term, "tmux-256color") {
			return lipgloss.NormalBorder()
		}
		return lipgloss.RoundedBorder()
	}()

	viewportStyle = lipgloss.NewStyle().
			BorderStyle(borderStyle).
			BorderForeground(outputBorderColor).
			Padding(1) // Both vertical and horizontal

	statusStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(lipColorLightBlue)).
			Foreground(lipgloss.Color(lipColorWhite)).
			Padding(0, 1) // Vertical, Horizontal

	inputStyle = lipgloss.NewStyle().
			BorderStyle(borderStyle).
			BorderForeground(inputBorderColor).
			Padding(0, 1)

	userStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(lipColorMagenta)).
			Bold(true)

	assistantStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(lipColorGreen)).
			Bold(true)
)

// Model represents the TUI state (this has nothing to do with LLMs)
type Model struct {
	viewport     viewport.Model
	textInput    textinput.Model
	content      string
	opts         *config.Options
	clientArgs   LLM.ClientArgs
	db           *database.ChatDB
	windowWidth  int
	windowHeight int
	ready        bool
	processing   bool
	statusMsg    string
	streamChan   <-chan LLM.StreamResponse
	fullResponse string
	lineWrapper  *linewrap.LineWrapper
}

func Initialize(opts *config.Options, clientArgs LLM.ClientArgs, db *database.ChatDB) Model {
	// log.Printf("Initializing with screen size: %dx%d", opts.ScreenWidth, opts.ScreenHeight)

	// Note: textinput already has built-in Emacs-style keybindings
	// Ctrl+A: Move cursor to beginning of line
	// Ctrl+E: Move cursor to end of line
	// Ctrl+W: Delete word backward
	// Ctrl+K: Delete to end of line
	// Ctrl+U: Delete to beginning of line

	/*
		// textinput.Model definition for reference
		type TextInput struct {
		    Value             string         // Current input value
		    Prompt            string         // Prompt displayed before input
		    Placeholder       string         // Placeholder when empty
		    Width             int            // Width of the input field
		    Cursor            cursor.Model   // Cursor model
		    CharLimit         int            // Character limit
		    Style             lipgloss.Style // Text styling
		    PromptStyle       lipgloss.Style // Prompt styling
		    PlaceholderStyle  lipgloss.Style // Placeholder styling
		}
	*/

	ti := textinput.New()
	ti.Placeholder = "Ask a question..."
	ti.Focus()
	ti.CharLimit = 0

	ti.Width = opts.ScreenWidth

	// Account for tmux which doesn't show fancy code points
	term := os.Getenv("TERM")
	if strings.Contains(term, "tmux-256color") {
		ti.Prompt = "> "
	} else {
		ti.Prompt = "❯ "
	}

	/*
		type Viewport struct {
		    Width      int           // Total width including borders
		    Height     int           // Total height including borders
		    Style      lipgloss.Style // Style for the viewport (borders, padding)
		    KeyMap     KeyMap        // Key mappings for scrolling
		    YPosition  int           // Current Y scroll position
		    YOffset    int           // Y offset for content
		    HighOffset int           // Highest scrollable offset
		    content    strings.Builder // Content inside viewport
		}
	*/

	// Scrollable viewport for chat history, with small padding
	totalFixedHeight := inputHeight + statusHeight + borderHeight + contentPadding + testPadding
	viewportHeight := opts.ScreenHeight - totalFixedHeight

	// TODO: Avante made this change
	// Calculate initial content width, accounting for viewport padding and borders
	contentWidth := opts.ScreenTextWidth - contentMargin

	vp := viewport.New(contentWidth, viewportHeight)
	vp.SetContent("")
	vp.YPosition = 0 // Bubble Tea/Lipgloss will adjust for the border

	// Wrap it up
	lw := linewrap.NewLineWrapper(contentWidth, opts.TabWidth, linewrap.NilWriter)

	return Model{
		viewport:     vp,
		textInput:    ti,
		content:      "",
		opts:         opts,
		clientArgs:   clientArgs,
		db:           db,
		fullResponse: "",
		statusMsg:    fmt.Sprintf("Model: %s | ConvID: %d | Ctrl+C: Exit | /help: Commands", *clientArgs.Model, *clientArgs.ConvID),
		windowWidth:  opts.ScreenWidth,
		// windowHeight is the viewport height, not the total window height
		windowHeight: viewportHeight,
		ready:        false,
		lineWrapper:  lw,
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	// textinput.Blink is the command to make the cursor blink.
	return tea.Batch(
		textinput.Blink,
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.processing {
				return m, nil
			}

			prompt := m.textInput.Value()
			if prompt == "" {
				return m, nil
			}

			if strings.HasPrefix(prompt, "/") {
				return m.handleSlashCommand(prompt)
			}

			m.textInput.SetValue("")

			userMsg := userStyle.Render("User: ") + strings.Trim(prompt, " \t\n") + "\n\n"
			m.content += userMsg

			// Pre-wrap content using viewport width to handle Charm's wrapping problems
			m.updateViewportContent()

			// Add Assistant prefix before streaming starts
			m.content += assistantStyle.Render("Assistant: ")
			m.lineWrapper.SetCurrWidth(len("Assistant: "))

			// TODO: add spinner
			m.processing = true
			m.statusMsg = "Processing..."

			return m, tea.Batch(append(cmds, func() tea.Msg {
				return promptMsg{prompt: prompt}
			})...)
		}

	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height

		// NOTE: This affects the viewport height/top border
		// Use consistent height calculations across the application
		totalFixedHeight := inputHeight + statusHeight + borderHeight + contentPadding + testPadding
		// TODO: this can possible be removed as we figured out the scrolling issue
		// Add a small buffer to prevent content from being hidden
		viewportHeight := m.windowHeight - totalFixedHeight
		viewportWidth := m.windowWidth - contentMargin

		textWidth := min(viewportWidth, m.opts.ScreenTextWidth)
		m.lineWrapper.SetMaxWidth(textWidth)

		m.viewport.Height = viewportHeight
		m.viewport.Width = viewportWidth
		m.textInput.Width = viewportWidth

		if !m.ready {
			m.ready = true
		} else {
			// Re-wrap and set content on resize after initial setup (this is
			// to deal with Charm's wrapping problems)
			m.updateViewportContent()
		}

		return m, nil

	case promptMsg:
		if m.clientArgs.Prompt == nil {
			m.clientArgs.Prompt = new(string)
		}
		*m.clientArgs.Prompt = msg.prompt

		// Start the stream processing in a goroutine and return a command
		// that waits for stream messages.
		return m, m.startStreaming()

	case streamChunkMsg:
		if msg.err != nil {
			m.content += "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color(lipColorRed)).Render("Error: "+msg.err.Error()) + "\n\n"
			m.processing = false
			m.statusMsg = fmt.Sprintf("Error | Model: %s | ConvID: %d", *m.clientArgs.Model, *m.clientArgs.ConvID)
		} else {
			// m.content is for the viewport and contains everything that has
			// been displayed so far; m.fullResponse is for the current response only
			m.content += msg.chunk
			m.fullResponse += msg.chunk // Don't store the wrapped chunk in DB
			if msg.done {
				m.content += "\n\n"
				m.processing = false
				m.lineWrapper.Reset()
				m.statusMsg = fmt.Sprintf("Model: %s | ConvID: %d | /help for commands", *m.clientArgs.Model, *m.clientArgs.ConvID)
				m.saveConversation()
				m.updateContext()
			}
		}
		m.updateViewportContent()

		if !msg.done {
			cmds = append(cmds, waitForStreamChunk(&m, m.streamChan))
		}
		return m, tea.Batch(cmds...)

	case responseMsg:
		m.processing = false

		if msg.err != nil {
			m.statusMsg = lipgloss.NewStyle().Foreground(lipgloss.Color(lipColorRed)).Render("Error: " + msg.err.Error())
			return m, nil
		}

		// This case might not be needed anymore if streamChunkMsg handles everything,
		// but keeping it for potential non-streaming errors or scenarios.
		m.content += assistantStyle.Render("Assistant: ") + msg.response + "\n\n"
		m.updateViewportContent()

		m.statusMsg = fmt.Sprintf("Model: %s | ConvID: %d | /help for commands", *m.clientArgs.Model, *m.clientArgs.ConvID)

		return m, tea.Batch(cmds...)
	}

	if m.textInput.Focused() {
		// When text input is focused, only handle specific keys and mouse
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyPgUp:
				m.viewport.LineUp(m.viewport.Height)
			case tea.KeyPgDown:
				m.viewport.LineDown(m.viewport.Height)
			case tea.KeyUp:
				m.viewport.LineUp(1)
			case tea.KeyDown:
				m.viewport.LineDown(1)
			}
		case tea.MouseMsg:
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}
	} else {
		// When not focused on text input, handle all viewport updates
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	var tiCmd tea.Cmd
	m.textInput, tiCmd = m.textInput.Update(msg)
	if tiCmd != nil {
		cmds = append(cmds, tiCmd)
	}

	return m, tea.Batch(cmds...)
}

// Deal with Charm's wrapping problems by pre-wrapping content with lipgloss
func (m *Model) updateViewportContent() {
	wrappedContent := lipgloss.NewStyle().Width(m.viewport.Width).Render(m.content)
	m.viewport.SetContent(wrappedContent)
	m.viewport.GotoBottom()
}

func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// NOTE: this accounts for issues with the right border not appearing; for
	// whatever reason, I have to reduce the width by 2 (which happens to be
	// the same as 'contentPadding', so I'm going to use that and see if anything
	// changes, they stay related)
	contentWidth := m.windowWidth - contentPadding

	viewportBox := viewportStyle.
		Width(contentWidth).
		Height(m.viewport.Height).
		Render(m.viewport.View())

	statusLine := statusStyle.
		Width(contentWidth).
		Padding(0, 1).
		Render(m.statusMsg)

	inputBox := inputStyle.
		Width(contentWidth).
		Render(m.textInput.View())

	return lipgloss.JoinVertical(lipgloss.Center,
		viewportBox,
		statusLine,
		inputBox,
	)
}

type promptMsg struct {
	prompt string
}

type responseMsg struct {
	response string
	err      error
}

// TODO: Does this need to be in types.go?
type streamChunkMsg struct {
	chunk string
	done  bool
	err   error
}

func (m *Model) startStreaming() tea.Cmd {
	// Determine provider and model key (support provider/model syntax)
	spec := *m.clientArgs.Model
	provider := m.opts.Provider
	modelKey := spec
	if idx := strings.Index(spec, "/"); idx >= 0 {
		provider = spec[:idx]
		modelKey = spec[idx+1:]
		m.opts.Provider = provider
		// update clientArgs.Model to the pure model key for recording
		*m.clientArgs.Model = modelKey
	}
	// Load model configuration from config file
	modelConf, err := config.GetModelConfig(m.opts.Config, provider, modelKey)
	if err != nil {
		return func() tea.Msg {
			return streamChunkMsg{err: fmt.Errorf("model %q not found for provider %q", modelKey, provider), done: true}
		}
	}
	// Override args with API-specific configuration
	apiModel := modelConf.ModelName
	m.clientArgs.Model = &apiModel
	apiTemp := float32(modelConf.Temperature)
	m.clientArgs.Temperature = &apiTemp
	apiMax := modelConf.MaxTokens
	m.clientArgs.MaxTokens = &apiMax
	// Initialize the LLM client based on provider
	var client LLM.Client
	switch provider {
	case "openai":
		client = LLM.NewOpenAI("openai", "https://api.openai.com/v1/")
	case "claude", "anthropic":
		client = LLM.NewAnthropic()
	case "gemini", "google":
		client = LLM.NewGoogle()
	case "ollama":
		client = LLM.NewOllama()
	case "grok", "xai":
		client = LLM.NewOpenAI("xai", "https://api.x.ai/v1/")
	default:
		return func() tea.Msg {
			return streamChunkMsg{err: fmt.Errorf("unknown provider: %q", provider), done: true}
		}
	}
	// Start the chat stream
	_, streamChan, err := client.Chat(m.clientArgs, m.opts.ScreenTextWidth, m.opts.TabWidth)
	if err != nil {
		return func() tea.Msg {
			return streamChunkMsg{err: err, done: true}
		}
	}
	m.streamChan = streamChan
	m.fullResponse = ""
	return waitForStreamChunk(m, m.streamChan)
}

func (m *Model) saveConversation() {
	if m.opts.NoRecord {
		return
	}

	// TODO: get actual counts if the API provides them
	inputTokens := LLM.EstimateTokens(*m.clientArgs.Prompt)
	outputTokens := LLM.EstimateTokens(m.fullResponse)

	// Remove the ANSI escape sequences from the response using a regex to cover all
	// possible escape sequences.
	ansiEscapeRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	m.fullResponse = ansiEscapeRegex.ReplaceAllString(m.fullResponse, "")

	// Save to the database
	dbErr := m.db.InsertConversation(
		*m.clientArgs.Prompt,
		m.fullResponse,
		*m.clientArgs.Model,
		*m.clientArgs.Temperature,
		inputTokens,
		outputTokens,
		*m.clientArgs.ConvID,
	)
	if dbErr != nil {
		// TODO: Log the error
		m.statusMsg = fmt.Sprintf("Error saving to DB: %v", dbErr)
	}
	logger.Debug("Inserted conversation into database", "convID", *m.clientArgs.ConvID, "prompt", *m.clientArgs.Prompt, "response", m.fullResponse)
}

func (m *Model) updateContext() {
	m.opts.ContinueChat = true
	promptContext, err := m.db.LoadConversationFromDB(*m.clientArgs.ConvID)
	if err == nil {
		m.clientArgs.Context = promptContext
	} else {
		m.clientArgs.Context = nil
		m.statusMsg = fmt.Sprintf("Error loading context: %v", err)
	}
}

func (m Model) handleSlashCommand(cmd string) (tea.Model, tea.Cmd) {
	parts := strings.SplitN(cmd, " ", 2)
	command := parts[0]

	switch command {
	case "/exit", "/quit":
		return m, tea.Quit

	case "/help", "/?":
		helpText := `
Available commands:
  /exit, /quit - Exit the application
  /help, /?    - Show this help message
  /model       - Show current model
  /model NAME  - Change model to NAME
  /id          - Show conversation ID
  /clear       - Clear the conversation history
  /context     - Show the current context
`
		m.content += helpText + "\n"
		m.viewport.SetContent(m.content)
		// Deal with Charm's wrapping problems
		wrappedContent := lipgloss.NewStyle().Width(m.viewport.Width).Render(m.content)
		m.viewport.SetContent(wrappedContent)
		m.viewport.GotoBottom()
		m.textInput.SetValue("")

	case "/model":
		if len(parts) > 1 && parts[1] != "" {
			newModel := strings.TrimSpace(parts[1])
			*m.clientArgs.Model = newModel
			m.statusMsg = fmt.Sprintf("Model changed to: %s | ConvID: %d", newModel, *m.clientArgs.ConvID)
		} else {
			m.content += fmt.Sprintf("Current model: %s\n\n", *m.clientArgs.Model)
			// Deal with Charm's wrapping problems
			m.updateViewportContent()
		}
		m.textInput.SetValue("")

	case "/id":
		m.content += fmt.Sprintf("Conversation ID: %d\n\n", *m.clientArgs.ConvID)
		// Deal with Charm's wrapping problems
		m.updateViewportContent()
		m.textInput.SetValue("")

	case "/clear":
		m.clientArgs.Context = nil
		m.content = ""
		m.viewport.SetContent(m.content)
		m.textInput.SetValue("")

	case "/context":
		if len(m.clientArgs.Context) == 0 {
			m.content += "No conversation context available.\n\n"
		} else {
			m.content += "Current context:\n"
			for i, ctx := range m.clientArgs.Context {
				m.content += fmt.Sprintf("%d. %s: %s\n", i+1, ctx.Role, ctx.Content[:min(50, len(ctx.Content))])
				if len(ctx.Content) > 50 {
					m.content += "...\n"
				}
			}
			m.content += "\n"
		}
		// Deal with Charm's wrapping problems
		m.updateViewportContent()
		m.textInput.SetValue("")

	default:
		m.content += fmt.Sprintf("Unknown command: %s\n\n", command)
		// Deal with Charm's wrapping problems
		m.updateViewportContent()
		m.textInput.SetValue("")
	}

	return m, nil
}

func Run(opts *config.Options, clientArgs LLM.ClientArgs, db *database.ChatDB) error {
	m := Initialize(opts, clientArgs, db)
	logger.Debug("Starting TUI program", "opts", opts, "clientArgs", clientArgs)

	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(), // Enable mouse support
		// tea.WithInputTTY(),         // Enable input from TTY
		// tea.WithOutput(os.Stderr),  // Redirect output to stderr
		// tea.WithInput(os.Stdin),    // Redirect input from stdin
		// tea.WithInputTTY(),         // Enable input from TTY
	)

	if _, err := p.Run(); err != nil {
		log.Printf("Error running program: %v", err)
		return err
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// returns a command that waits for the next message on the stream channel
func waitForStreamChunk(m *Model, sub <-chan LLM.StreamResponse) tea.Cmd {
	return func() tea.Msg {
		resp, ok := <-sub
		if !ok {
			return streamChunkMsg{done: true}
		}

		wrappedChunk := m.lineWrapper.Wrap([]byte(resp.Content))
		return streamChunkMsg{chunk: wrappedChunk, done: resp.Done, err: resp.Error}
	}
}
