package tui

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/duluk/ask-ai/pkg/LLM"
	"github.com/duluk/ask-ai/pkg/config"
	"github.com/duluk/ask-ai/pkg/database"
)

// Layout constants
const (
	inputHeight    = 3 // Input box height
	statusHeight   = 1 // Status line height
	borderWidth    = 2 // Border width (left + right)
	borderHeight   = 2 // Border height (top + bottom)
	contentPadding = 1 // Padding inside components
	contentMargin  = 4 // Total horizontal margin for content (borders + padding)
)

// Color constants
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

// Styles
var (
	outputBorderColor = lipgloss.Color(lipColorDarkBlue)
	inputBorderColor  = lipgloss.Color(lipColorGreen)

	// If using a system that doesn't have great support for fancy borders, just use plain ASCII
	borderStyle = func() lipgloss.Border {
		term := os.Getenv("TERM")
		// if strings.Contains(term, "windows") || strings.Contains(term, "xterm") || strings.Contains(term, "tmux-256color") {
		// tmux is not displaying the rounded corners very well
		if strings.Contains(term, "tmux-256color") {
			return lipgloss.NormalBorder()
		}
		return lipgloss.RoundedBorder()
	}()

	outputStyle = lipgloss.NewStyle().
			BorderStyle(borderStyle).
			BorderForeground(outputBorderColor).
			Padding(1)

	statusStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(lipColorLightBlue)).
			Foreground(lipgloss.Color(lipColorWhite)).
			Padding(0, 1)

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
	viewport   viewport.Model
	textInput  textinput.Model
	content    string
	opts       *config.Options
	clientArgs LLM.ClientArgs
	db         *database.ChatDB
	width      int
	height     int
	ready      bool
	processing bool
	statusMsg  string
}

// Create a new TUI model ^
func Initialize(opts *config.Options, clientArgs LLM.ClientArgs, db *database.ChatDB) Model {
	// log.Printf("Initializing with screen size: %dx%d", opts.ScreenWidth, opts.ScreenHeight)

	// Note: textinput already has built-in Emacs-style keybindings
	// Ctrl+A: Move cursor to beginning of line
	// Ctrl+E: Move cursor to end of line
	// Ctrl+W: Delete word backward
	// Ctrl+K: Delete to end of line
	// Ctrl+U: Delete to beginning of line

	ti := textinput.New()
	ti.Placeholder = "Ask a question..."
	ti.Focus()
	ti.CharLimit = 0

	ti.Width = opts.ScreenWidth

	// Scrollable viewport for chat history
	// TODO: remove this? I can't figure out what this affects.
	inputHeight := 3  // Input box height
	statusHeight := 1 // Status line height
	borderHeight := 2 // Account for borders
	adjustedHeight := inputHeight + statusHeight + borderHeight

	vp := viewport.New(opts.ScreenWidth, opts.ScreenHeight-adjustedHeight)
	vp.SetContent("")
	vp.YPosition = 0

	return Model{
		viewport:   vp,
		textInput:  ti,
		content:    "",
		opts:       opts,
		clientArgs: clientArgs,
		db:         db,
		statusMsg:  fmt.Sprintf("Model: %s | ConvID: %d | Ctrl+C: Exit | /help: Commands", *clientArgs.Model, *clientArgs.ConvID),
		width:      opts.ScreenWidth,
		height:     opts.ScreenHeight - adjustedHeight,
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model
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

			// Add user message to content
			userMsg := userStyle.Render("User: ") + strings.Trim(prompt, " \t\n") + "\n\n"
			m.content += userMsg
			m.viewport.SetContent(m.content)
			m.viewport.GotoBottom()

			// Set processing state
			// TODO: add spinner
			m.processing = true
			m.statusMsg = "Processing..."

			// Create command to process the prompt
			return m, tea.Batch(append(cmds, func() tea.Msg {
				return promptMsg{prompt: prompt}
			})...)
		}

	case tea.WindowSizeMsg:
		// log.Printf("Window size message: %dx%d", msg.Width, msg.Height)
		m.height = msg.Height
		m.width = msg.Width

		if !m.ready {
			m.ready = true
		}

		// log.Printf("Window size: %dx%d", m.width, m.height)

		// Adjust viewport and input sizes
		// NOTE: This affects the viewport height/top border
		inputHeight := 3  // Input box height
		statusHeight := 1 // Status line height
		// borderHeight := 2   // Account for borders
		adjustedHeight := 4 // padding to show the top border, I don't know why
		// adjustedHeight := inputHeight + statusHeight + borderHeight - 6

		viewportHeight := m.height - inputHeight - statusHeight - adjustedHeight
		// viewportHeight := m.height - adjustedHeight
		adjustedWidth := m.width - 2 // Account for left and right borders

		// NOTE: this is the inside width, for the text itself - so
		// it affects things like wrapping. If no adjustment is made,
		// the text doesn't wrap correctly. Why 4? /shrug
		// adjustedWidth := m.width - 4
		m.viewport.Width = adjustedWidth
		m.viewport.Height = viewportHeight
		m.textInput.Width = adjustedWidth

		m.viewport.SetContent(m.content)
		m.viewport.GotoBottom()

		return m, nil

	case promptMsg:
		// log.Printf("Prompt message received")
		// Process the prompt with LLM
		// Make sure clientArgs.Prompt is not nil
		if m.clientArgs.Prompt == nil {
			m.clientArgs.Prompt = new(string)
		}
		*m.clientArgs.Prompt = msg.prompt

		// Create command to get LLM response
		return m, func() tea.Msg {
			resp, err := m.processPrompt()
			return responseMsg{response: resp, err: err}
		}

	// Handle streaming chunks
	case streamChunkMsg:
		// Update viewport with new content
		m.viewport.SetContent(m.content)
		m.viewport.GotoBottom()

		// Force immediate render
		return m, tea.Batch(
			func() tea.Msg {
				return tea.WindowSizeMsg{
					Width:  m.width,
					Height: m.height,
				}
			},
		)

	case responseMsg:
		// log.Printf("Response message received")
		m.processing = false

		if msg.err != nil {
			m.statusMsg = lipgloss.NewStyle().Foreground(lipgloss.Color(lipColorRed)).Render("Error: " + msg.err.Error())
			// m.statusMsg = fmt.Sprintf("Error: %s", msg.err)
			return m, nil
		}

		// Add assistant response to content
		assistantMsg := assistantStyle.Render("Assistant: ") + msg.response + "\n\n"
		m.content += assistantMsg

		// Make sure to set the content and scroll to bottom
		m.viewport.SetContent(m.content)
		m.viewport.GotoBottom()

		// TODO: commented out because the redraw looked awkward and it isn't
		// needed AFAICT; I think this was an attempt to fix the border issue,
		// which wasn't an issue. Probably can be removed at some point.
		// Force a redraw of the viewport to ensure proper rendering
		cmds = []tea.Cmd{func() tea.Msg {
			return tea.WindowSizeMsg{
				Width:  m.width,
				Height: m.height,
			}
		}}

		// Update status message
		m.statusMsg = fmt.Sprintf("Model: %s | ConvID: %d | /help for commands", *m.clientArgs.Model, *m.clientArgs.ConvID)

		return m, tea.Batch(cmds...)
	}

	// Handle viewport updates
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	// Handle text input updates
	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View implements tea.Model
func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// NOTE: This accounts for the border on viewport and input, as in it makes
	// them line up correctly. Why 2? /shrug
	contentWidth := m.width - 2

	// Layout the components
	outputBox := outputStyle.
		Width(contentWidth).
		Height(m.viewport.Height).
		Render(m.viewport.View())

	statusLine := statusStyle.
		// Reduce width by 2 to account for left and right spacing; this allows
		// it to line up with the borders better
		Width(m.width - 2).
		// Align(lipgloss.Center).  // this centers the text
		Render(m.statusMsg)

	inputBox := inputStyle.
		Width(contentWidth).
		Render(m.textInput.View())

	// Ensure proper vertical layout with fixed heights
	return fmt.Sprintf(
		"%s\n%s%s%s\n%s",
		outputBox,
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

// Process the prompt with LLM
// Add these types to your custom message types section
type streamChunkMsg struct {
	chunk string
	done  bool
	err   error
}

// Update the processPrompt method
func (m *Model) processPrompt() (string, error) {
	var client LLM.Client

	// Ensure we have valid pointers
	if m.clientArgs.Model == nil {
		m.clientArgs.Model = new(string)
		*m.clientArgs.Model = "chatgpt" // Default model
	}

	if m.clientArgs.Prompt == nil {
		return "", fmt.Errorf("prompt is nil")
	}

	if m.clientArgs.ConvID == nil {
		m.clientArgs.ConvID = new(int)
		*m.clientArgs.ConvID = 1
	}

	model := *m.clientArgs.Model

	switch model {
	// case "chatgpt":
	// 	api_url := "https://api.openai.com/v1/"
	// 	client = LLM.NewOpenAI("openai", api_url)
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

	// Start a goroutine to process the stream
	go func() {
		err := client.ChatStream(m.clientArgs, m.opts.ScreenWidth, m.opts.TabWidth, streamChan)
		if err != nil {
			// If there's an error starting the stream, it will be sent through the channel
			// by the ChatStream implementation
		}
	}()

	// Collect the full response for logging
	fullResponse := ""

	// Process the stream
	for resp := range streamChan {
		if resp.Error != nil {
			m.opts.NoOutput = originalNoOutput
			return "", resp.Error
		}

		fullResponse += resp.Content

		// Send the chunk to the Bubble Tea program
		m.content += resp.Content
		// m.viewport.SetContent(m.content)
		// m.viewport.GotoBottom()

		// Create a command to update the UI
		cmds := []tea.Cmd{
			func() tea.Msg {
				return streamChunkMsg{
					chunk: resp.Content,
					done:  resp.Done,
				}
			},
		}

		// Execute the commands immediately
		for _, cmd := range cmds {
			if msg := cmd(); msg != nil {
				m.Update(msg)
			}
		}

		// TODO: this may look ugly...
		// Force a screen refresh after each chunk
		// tea.Batch(func() tea.Msg {
		// 	return tea.WindowSizeMsg{
		// 		Width:  m.width,
		// 		Height: m.height,
		// 	}
		// })

		// Add a small delay to ensure proper rendering
		// time.Sleep(10 * time.Millisecond)

		if resp.Done {
			m.viewport.GotoBottom()
			break
		}
	}

	// Restore original setting
	m.opts.NoOutput = originalNoOutput

	if !m.opts.NoRecord {
		// Estimate tokens for now
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
		m.viewport.GotoBottom()
		m.textInput.SetValue("")

	case "/model":
		if len(parts) > 1 && parts[1] != "" {
			newModel := strings.TrimSpace(parts[1])
			*m.clientArgs.Model = newModel
			m.statusMsg = fmt.Sprintf("Model changed to: %s | ConvID: %d", newModel, *m.clientArgs.ConvID)
		} else {
			m.content += fmt.Sprintf("Current model: %s\n\n", *m.clientArgs.Model)
			m.viewport.SetContent(m.content)
			m.viewport.GotoBottom()
		}
		m.textInput.SetValue("")

	case "/id":
		m.content += fmt.Sprintf("Conversation ID: %d\n\n", *m.clientArgs.ConvID)
		m.viewport.SetContent(m.content)
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
		m.viewport.SetContent(m.content)
		m.viewport.GotoBottom()
		m.textInput.SetValue("")

	default:
		m.content += fmt.Sprintf("Unknown command: %s\n\n", command)
		m.viewport.SetContent(m.content)
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
		tea.WithMouseCellMotion(),
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

// Add a new command to handle streaming chunks
func streamChunkCmd(chunkChan <-chan streamChunkMsg) tea.Cmd {
	return func() tea.Msg {
		chunk, ok := <-chunkChan
		if !ok {
			// Channel closed, no more chunks
			return nil
		}
		return chunk
	}
}
