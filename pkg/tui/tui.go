package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/duluk/ask-ai/pkg/LLM"
	"github.com/duluk/ask-ai/pkg/config"
	"github.com/duluk/ask-ai/pkg/database"
)

// Styles
var (
	borderColor = lipgloss.Color("63")

	outputStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(1)

	statusStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("12")).
			Foreground(lipgloss.Color("7")).
			Padding(0, 1)

	inputStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(0, 1)

	userStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("5")).
			Bold(true)

	assistantStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")).
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

// Initialize creates a new TUI model
func Initialize(opts *config.Options, clientArgs LLM.ClientArgs, db *database.ChatDB) Model {
	ti := textinput.New()
	ti.Placeholder = "Ask a question..."
	ti.Focus()
	ti.CharLimit = 0
	ti.Width = 80

	// Enable Emacs-style keybindings
	// Note: textinput already has built-in Emacs-style keybindings
	// Ctrl+A: Move cursor to beginning of line
	// Ctrl+E: Move cursor to end of line
	// Ctrl+W: Delete word backward
	// Ctrl+K: Delete to end of line
	// Ctrl+U: Delete to beginning of line
	// These are already implemented in the textinput component

	// Scrollable viewport for chat history
	vp := viewport.New(80, 20)
	vp.SetContent("")

	return Model{
		viewport:   vp,
		textInput:  ti,
		content:    "",
		opts:       opts,
		clientArgs: clientArgs,
		db:         db,
		statusMsg:  fmt.Sprintf("Model: %s | ConvID: %d | Ctrl+C: Exit | /help: Commands", *clientArgs.Model, *clientArgs.ConvID),
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
			userMsg := userStyle.Render("User: ") + prompt + "\n\n"
			m.content += userMsg
			m.viewport.SetContent(m.content)
			m.viewport.GotoBottom()

			// Set processing state
			m.processing = true
			m.statusMsg = "Processing..."

			// Create command to process the prompt
			return m, tea.Batch(append(cmds, func() tea.Msg {
				return promptMsg{prompt: prompt}
			})...)
		}

	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width

		if !m.ready {
			m.ready = true
		}

		// Adjust viewport and input sizes
		inputHeight := 3 // Input + border
		statusHeight := 1
		viewportHeight := m.height - inputHeight - statusHeight - 2 // Extra padding

		m.viewport.Width = m.width - 4 // Account for padding
		m.viewport.Height = viewportHeight
		m.textInput.Width = m.width - 6 // Account for padding and border

		m.viewport.SetContent(m.content)
		return m, nil

	case promptMsg:
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

	case responseMsg:
		m.processing = false

		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Error: %s", msg.err)
			return m, nil
		}

		// Add assistant response to content
		assistantMsg := assistantStyle.Render("Assistant: ") + msg.response + "\n\n"
		m.content += assistantMsg

		// Make sure to set the content and scroll to bottom
		m.viewport.SetContent(m.content)
		m.viewport.GotoBottom()

		// Force a redraw of the viewport to ensure proper rendering
		cmds := []tea.Cmd{func() tea.Msg {
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

	// Layout the components
	outputBox := outputStyle.
		Width(m.width - 4).
		Height(m.viewport.Height).
		Render(m.viewport.View())

	statusLine := statusStyle.
		Width(m.width).
		Render(m.statusMsg)

	inputBox := inputStyle.
		Width(m.width - 4).
		Render(m.textInput.View())

	// Ensure proper vertical layout with fixed heights
	return fmt.Sprintf(
		"%s\n%s\n%s",
		outputBox,
		statusLine,
		inputBox,
	)
}

// Custom message types
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
		m.viewport.SetContent(m.content)
		m.viewport.GotoBottom()

		if resp.Done {
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

// Handle slash commands
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

// Run starts the TUI
func Run(opts *config.Options, clientArgs LLM.ClientArgs, db *database.ChatDB) error {
	m := Initialize(opts, clientArgs, db)
	p := tea.NewProgram(m, tea.WithAltScreen())

	_, err := p.Run()
	return err
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
