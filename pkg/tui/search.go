package tui

import (
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/duluk/ask-ai/pkg/LLM"
	"github.com/duluk/ask-ai/pkg/config"
	"github.com/duluk/ask-ai/pkg/database"
)

type searchItem struct {
	title string
	desc  string
	id    int
}

func (i searchItem) Title() string       { return i.title }
func (i searchItem) Description() string { return i.desc }
func (i searchItem) FilterValue() string { return i.title + " " + i.desc }

type listModel struct {
	list       list.Model
	selectedID int
}

type inlineDelegate struct {
	list.DefaultDelegate
}

func (d inlineDelegate) Height() int {
	// only one line per item
	return 1
}

func (d inlineDelegate) Spacing() int {
	// no blank lines
	return 0
}

// Render renders a single item inline: title, separator, and description on one line
func (d inlineDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	// Assert the correct item type
	item, ok := listItem.(searchItem)
	if !ok {
		return
	}
	// Choose styles based on selection
	var titleStyle, descStyle lipgloss.Style
	if index == m.Index() {
		titleStyle = d.Styles.SelectedTitle
		descStyle = d.Styles.SelectedDesc
	} else {
		titleStyle = d.Styles.NormalTitle
		descStyle = d.Styles.NormalDesc
	}
	// Join title, separator, and description horizontally
	joined := lipgloss.JoinHorizontal(
		lipgloss.Top,
		titleStyle.Render(item.title),
		lipgloss.NewStyle().Foreground(lipgloss.Color("238")).Render(" – "),
		descStyle.Render(item.desc),
	)
	// Render inline, truncating to available width
	style := lipgloss.NewStyle().Inline(true).MaxWidth(m.Width())
	fmt.Fprint(w, style.Render(joined))
}

func (m listModel) Init() tea.Cmd { return nil }

func (m listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			// If user is not currently editing the filter, select item
			if !m.list.SettingFilter() {
				if item, ok := m.list.SelectedItem().(searchItem); ok {
					m.selectedID = item.id
				}
				return m, tea.Quit
			}
		}
	}

	// Delegate all other messages (including filter input) to the list
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m listModel) View() string {
	return m.list.View() + lipgloss.NewStyle().Padding(1, 0).Render("Press ENTER to select a conversation, ESC to cancel.")
}

// RunSearch launches an interactive list to select a conversation matching the keyword
// Returns the selected conversation ID, or 0 if none selected or no matches.
func RunSearch(opts *config.Options, db *database.ChatDB) (int, error) {
	width := int(math.Max(float64(opts.ScreenWidth-10), 20))

	ids, err := db.SearchForConversation(opts.SearchKeyword)
	if err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}

	var items []list.Item
	for _, id := range ids {
		convs, err := db.LoadConversationFromDB(id)
		if err != nil {
			return 0, err
		}

		excerpt := excerptFromConvs(convs, opts.SearchKeyword, width)
		title := fmt.Sprintf("%04d", id)
		items = append(items, searchItem{title: title, desc: excerpt, id: id})
	}

	// limit height to a reasonable size
	height := len(items) + 4
	if max := opts.ScreenHeight - 5; height > max {
		height = max
	}

	delegate := inlineDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
	}
	lst := list.New(items, delegate /*list.NewDefaultDelegate(),*/, width, height)
	lst.Title = fmt.Sprintf("Search results for '%s'", opts.SearchKeyword)

	m := listModel{list: lst}
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return 0, err
	}
	result := finalModel.(listModel).selectedID
	return result, nil
}

// excerptFromConvs returns a snippet of the conversation messages
// If keyword is non-empty, show context around its first occurrence; otherwise show leading text
func excerptFromConvs(convs []LLM.LLMConversations, keyword string, maxLen int) string {
	if len(convs) == 0 {
		return ""
	}
	content := convs[0].Content

	// If searching for keyword, show snippet around keyword
	if keyword != "" {
		lowerKW := strings.ToLower(keyword)
		lowerContent := strings.ToLower(content)
		if idx := strings.Index(lowerContent, lowerKW); idx >= 0 {
			// Determine context window
			half := maxLen / 2
			start := idx - half
			if start < 0 {
				start = 0
			}
			end := idx + len(keyword) + half
			if end > len(content) {
				end = len(content)
			}
			snippet := content[start:end]

			if start > 0 {
				snippet = "..." + snippet
			}
			if end < len(content) {
				snippet = snippet + "..."
			}

			if len(snippet) > maxLen {
				snippet = snippet[:maxLen-3] + "..."
			}
			return snippet
		}
	}
	if len(content) > maxLen {
		return content[:maxLen-3] + "..."
	}
	return content
}

// RunList launches an interactive list to select any conversation
// Returns the selected conversation ID, or 0 if none selected or no conversations.
func RunList(opts *config.Options, db *database.ChatDB) (int, error) {
	ids, err := db.ListConversationIDs()
	if err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}

	width := int(math.Max(float64(opts.ScreenWidth-10), 20))

	var items []list.Item
	for _, id := range ids {
		convs, err := db.LoadConversationFromDB(id)
		if err != nil {
			return 0, err
		}

		excerpt := excerptFromConvs(convs, "", width)
		title := fmt.Sprintf("%04d", id)
		items = append(items, searchItem{title: title, desc: excerpt, id: id})
	}

	height := len(items) + 4
	if max := opts.ScreenHeight - 5; height > max {
		height = max
	}

	delegate := inlineDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
	}
	lst := list.New(items, delegate /*list.NewDefaultDelegate(),*/, width, height)
	lst.Title = "Conversations"

	m := listModel{list: lst}
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return 0, err
	}
	selected := finalModel.(listModel).selectedID
	return selected, nil
}
