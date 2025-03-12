package tui

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/duluk/ask-ai/pkg/LLM"
	"github.com/duluk/ask-ai/pkg/config"
	"github.com/duluk/ask-ai/pkg/database"
)

const (
	inputHeight    = 3 // Input box height
	statusHeight   = 1 // Status line height
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
		// tmux is not displaying the rounded corners very well
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
}

// TickMsg is used for periodic updates
type TickMsg time.Time

func Initialize(opts *config.Options, clientArgs LLM.ClientArgs, db *database.ChatDB) Model {
	// log.Printf("Initializing with screen size: %dx%d", opts.ScreenWidth, opts.ScreenHeight)

	// Note: textinput already has built-in Emacs-style keybindings
	// Ctrl+A: Move cursor to beginning of line
	// Ctrl+E: Move cursor to end of line
	// Ctrl+W: Delete word backward
	// Ctrl+K: Delete to end of line
	// Ctrl+U: Delete to beginning of line

	/*
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

	vp := viewport.New(opts.ScreenWidth, viewportHeight)
	vp.SetContent("")
	vp.YPosition = 0 // Bubble Tea/Lipgloss will adjust for the border

	return Model{
		viewport:     vp,
		textInput:    ti,
		content:      "",
		opts:         opts,
		clientArgs:   clientArgs,
		db:           db,
		statusMsg:    fmt.Sprintf("Model: %s | ConvID: %d | Ctrl+C: Exit | /help: Commands", *clientArgs.Model, *clientArgs.ConvID),
		windowWidth:  opts.ScreenWidth,
		windowHeight: viewportHeight,
		ready:        false,
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		tick(),
	)
}

// TODO: this probably isn't needed anymore; it was trying to fix the scrolling issue
// tick sends a message every second
func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	// log.Printf("Received message type: %T", msg)

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

			// Handle slash commands
			if strings.HasPrefix(prompt, "/") {
				return m.handleSlashCommand(prompt)
			}

			// Clear input
			m.textInput.SetValue("")

			userMsg := userStyle.Render("User: ") + strings.Trim(prompt, " \t\n") + "\n\n"
			m.content += userMsg

			// Pre-wrap content using viewport width to handle Charm's wrapping problems
			wrappedContent := lipgloss.NewStyle().Width(m.viewport.Width).Render(m.content)
			m.viewport.SetContent(wrappedContent)
			m.viewport.GotoBottom()

			// TODO: add spinner
			m.processing = true
			m.statusMsg = "Processing..."

			// Create command to process the prompt
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
		contentWidth := m.windowWidth - contentMargin

		// NOTE: this is the inside width, for the text itself - so
		// it affects things like wrapping. If no adjustment is made,
		// the text doesn't wrap correctly. Why 4? /shrug
		// Update viewport dimensions
		m.viewport.Width = contentWidth
		m.viewport.Height = viewportHeight
		m.textInput.Width = contentWidth

		if !m.ready {
			m.ready = true
			// Perform initial setup like setting initial content (already handled below)
			// and potentially focusing input, starting cursor blink if needed here.
		} else {
			// Re-wrap and set content on resize after initial setup (this is
			// to deal with Charm's wrapping problems)
			wrappedContent := lipgloss.NewStyle().Width(m.viewport.Width).Render(m.content)
			m.viewport.SetContent(wrappedContent)
			m.viewport.GotoBottom()
		}

		return m, nil

	case promptMsg:
		if m.clientArgs.Prompt == nil {
			m.clientArgs.Prompt = new(string)
		}
		*m.clientArgs.Prompt = msg.prompt

		return m, func() tea.Msg {
			resp, err := m.processPrompt()
			return responseMsg{response: resp, err: err}
		}

	// Handle streaming chunks (this may be broken)
	case streamChunkMsg:
		// TODO: remove this probably (duplicated by processPrompt?)
		// Update viewport with new content
		// wrappedContent := lipgloss.NewStyle().Width(m.viewport.Width).Render(m.content)
		// m.viewport.SetContent(wrappedContent)

		// m.viewport, cmd = m.viewport.Update(msg)
		// cmds = append(cmds, cmd)

		// // Update viewport with new content
		// // Force a proper viewport refresh
		// cmds = append(cmds, func() tea.Msg {
		// 	return tea.WindowSizeMsg{
		// 		Width:  m.windowWidth,
		// 		Height: m.windowHeight,
		// 	}
		// })

		// m.viewport.GotoBottom()
		return m, tea.Batch(cmds...)

	// TODO: again this is probably not needed
	case TickMsg:
		m.viewport.GotoBottom()
		cmds = append(cmds, tick())

	case responseMsg:
		// log.Printf("Response message received")
		m.processing = false

		if msg.err != nil {
			m.statusMsg = lipgloss.NewStyle().Foreground(lipgloss.Color(lipColorRed)).Render("Error: " + msg.err.Error())
			return m, nil
		}

		assistantMsg := assistantStyle.Render("Assistant: ") + msg.response + "\n\n"
		m.content += assistantMsg

		// Deal with Charm's wrapping problems
		wrappedContent := lipgloss.NewStyle().Width(m.viewport.Width).Render(m.content)
		m.viewport.SetContent(wrappedContent)
		m.viewport.GotoBottom()

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

	// Handle text input updates
	var tiCmd tea.Cmd
	m.textInput, tiCmd = m.textInput.Update(msg)
	if tiCmd != nil {
		cmds = append(cmds, tiCmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// NOTE: This accounts for the border on viewport and input, as in it makes
	// them line up correctly. Why 2? /shrug
	contentWidth := m.windowWidth - 2

	viewportBox := viewportStyle.
		Width(contentWidth).
		Height(m.viewport.Height).
		Render(m.viewport.View())

	statusLine := statusStyle.
		// Reduce width by 2 to account for left and right spacing; this allows
		// it to line up with the borders better
		Width(m.windowWidth - 2).
		Render(m.statusMsg)

	inputBox := inputStyle.
		Width(contentWidth).
		Render(m.textInput.View())

	// Ensure proper vertical layout with fixed heights
	// TODO: try lipgloss.JoinVertical again
	return fmt.Sprintf(
		"%s\n%s%s%s\n%s",
		viewportBox,
		" ", statusLine, " ",
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

func (m *Model) processPrompt() (string, error) {
	var client LLM.Client

	if m.clientArgs.Model == nil {
		m.clientArgs.Model = new(string)
		*m.clientArgs.Model = "chatgpt" // Default model
	}

	if m.clientArgs.Prompt == nil {
		return "", fmt.Errorf("prompt is nil")
	}

	// TODO: I think this is broken
	if m.clientArgs.ConvID == nil {
		m.clientArgs.ConvID = new(int)
		*m.clientArgs.ConvID = 1
	}

	model := *m.clientArgs.Model

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
	case "ollama":
		client = LLM.NewOllama()
	default:
		return "", fmt.Errorf("unknown model: %s", model)
	}

	if !m.opts.NoRecord {
		LLM.LogChat(
			m.clientArgs.Log,
			"User",
			*m.clientArgs.Prompt,
			"",
			m.opts.ContinueChat,
			LLM.EstimateTokens(*m.clientArgs.Prompt),
			0,
			*m.clientArgs.ConvID,
		)
	}

	// TODO: I don't know if this NoOutput stuff is still needed
	// Make sure we're not printing to stdout
	originalNoOutput := m.opts.NoOutput
	m.opts.NoOutput = true

	// Create a channel for streaming responses
	streamChan := make(chan LLM.StreamResponse)

	go func() {
		err := client.ChatStream(m.clientArgs, m.opts.ScreenWidth, m.opts.TabWidth, streamChan)
		if err != nil {
			// If there's an error starting the stream, it will be sent through
			// the channel by the ChatStream implementation
		}
	}()

	fullResponse := ""
	for resp := range streamChan {
		if resp.Error != nil {
			m.opts.NoOutput = originalNoOutput
			return "", resp.Error
		}

		fullResponse += resp.Content

		// Send the chunk to the Bubble Tea program
		m.content += resp.Content

		// Deal with Charm's wrapping problems
		wrappedContent := lipgloss.NewStyle().Width(m.viewport.Width).Render(m.content)
		m.viewport.SetContent(wrappedContent)
		m.viewport.GotoBottom()

		if resp.Done {
			m.viewport.GotoBottom()
			break
		}
	}

	m.opts.NoOutput = originalNoOutput

	if !m.opts.NoRecord {
		inputTokens := LLM.EstimateTokens(*m.clientArgs.Prompt)
		outputTokens := LLM.EstimateTokens(fullResponse)

		LLM.LogChat(
			m.clientArgs.Log,
			"Assistant",
			fullResponse,
			model,
			m.opts.ContinueChat,
			inputTokens,
			outputTokens,
			*m.clientArgs.ConvID,
		)

		err := m.db.InsertConversation(
			*m.clientArgs.Prompt,
			fullResponse,
			model,
			*m.clientArgs.Temperature,
			inputTokens,
			outputTokens,
			*m.clientArgs.ConvID,
		)
		if err != nil {
			return fullResponse, fmt.Errorf("error inserting conversation into database: %v", err)
		}
	}

	// Update context for next conversation
	m.opts.ContinueChat = true
	promptContext, err := LLM.ContinueConversation(m.clientArgs.Log)
	if err == nil {
		m.clientArgs.Context = promptContext
	}

	return fullResponse, nil
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
			wrappedContent := lipgloss.NewStyle().Width(m.viewport.Width).Render(m.content)
			m.viewport.SetContent(wrappedContent)
			m.viewport.GotoBottom()
		}
		m.textInput.SetValue("")

	case "/id":
		m.content += fmt.Sprintf("Conversation ID: %d\n\n", *m.clientArgs.ConvID)
		// Deal with Charm's wrapping problems
		wrappedContent := lipgloss.NewStyle().Width(m.viewport.Width).Render(m.content)
		m.viewport.SetContent(wrappedContent)
		m.viewport.GotoBottom()
		m.textInput.SetValue("")

	case "/clear":
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
		wrappedContent := lipgloss.NewStyle().Width(m.viewport.Width).Render(m.content)
		m.viewport.SetContent(wrappedContent)
		m.viewport.GotoBottom()
		m.textInput.SetValue("")

	default:
		m.content += fmt.Sprintf("Unknown command: %s\n\n", command)
		// Deal with Charm's wrapping problems
		wrappedContent := lipgloss.NewStyle().Width(m.viewport.Width).Render(m.content)
		m.viewport.SetContent(wrappedContent)
		m.viewport.GotoBottom()
		m.textInput.SetValue("")
	}

	return m, nil
}

func Run(opts *config.Options, clientArgs LLM.ClientArgs, db *database.ChatDB) error {
	m := Initialize(opts, clientArgs, db)

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

// TODO: should this just be removed?
// func streamChunkCmd(chunkChan <-chan streamChunkMsg) tea.Cmd {
// 	return func() tea.Msg {
// 		chunk, ok := <-chunkChan
// 		if !ok {
// 			// Channel closed, no more chunks
// 			return nil
// 		}
// 		return chunk
// 	}
// }
