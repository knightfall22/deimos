package main

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type ModelID int

const (
	OLLAMA ModelID = iota
	OPENAI
	CLAUDE
	GEMINI
)

func (mid ModelID) String() string {
	switch mid {
	case OLLAMA:
		return "OLLAMA"
	case OPENAI:
		return "OPENAI"
	case CLAUDE:
		return "CLAUDE"
	case GEMINI:
		return "GEMINI"
	default:
		return "UNKNOWN"
	}
}

func (mid ModelID) FilterValue() string { return mid.String() }

type AIModel struct {
	id    ModelID
	label string
}

type styles struct {
	doc            lipgloss.Style
	highlight      lipgloss.Style
	leftPaneStyle  lipgloss.Style
	rightPaneStyle lipgloss.Style
	activeTab      lipgloss.Style
	selectedItem   lipgloss.Style
	window         lipgloss.Style
	item           lipgloss.Style
}

func newStyles(bgIsDark bool) *styles {
	lightDark := lipgloss.LightDark(bgIsDark)

	highlightColor := lightDark(lipgloss.Color("#874BFD"), lipgloss.Color("#7D56F4"))

	s := new(styles)
	s.window = lipgloss.NewStyle().
		BorderForeground(highlightColor).
		Padding(2, 0).
		Align(lipgloss.Center).
		Border(lipgloss.NormalBorder()).
		UnsetBorderTop()

	//left pane style
	s.leftPaneStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		MarginRight(1)

	// Right pane style
	s.rightPaneStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(1, 2)

	s.selectedItem = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))

	s.item = lipgloss.NewStyle().PaddingLeft(4)

	return s
}

type itemDelegate struct {
	styles *styles
}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(ModelID)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.String())

	fn := d.styles.item.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return d.styles.selectedItem.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type activePane int

const (
	listPane activePane = iota
	textareaPane
)

type model struct {
	list       list.Model
	activePane activePane
	styles     *styles
	width      int
	height     int
	textarea   textarea.Model
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "tab":
			if m.activePane == listPane {
				m.activePane = textareaPane
				cmds = append(cmds, m.textarea.Focus())
			} else {
				m.activePane = listPane
				m.textarea.Blur()
			}
			return m, tea.Batch(cmds...)
		case "esc":
			if m.activePane == textareaPane {
				m.activePane = listPane
				m.textarea.Blur()
			}
			return m, tea.Batch(cmds...)
		}

	case tea.WindowSizeMsg:
		verticalMargin := m.styles.leftPaneStyle.GetVerticalFrameSize()

		m.width = msg.Width
		m.height = msg.Height

		listWidth := (msg.Width * 30) / 100
		m.list.SetSize(listWidth, msg.Height-verticalMargin)

		// Calculate right pane width and update textarea width
		leftPaneWidth := listWidth + m.styles.leftPaneStyle.GetHorizontalFrameSize()
		rightWidth := msg.Width - leftPaneWidth
		taWidth := rightWidth - m.styles.rightPaneStyle.GetHorizontalFrameSize() - 4
		if taWidth > 0 {
			m.textarea.SetWidth(taWidth)
		}
	}

	var cmd tea.Cmd
	// Pass key messages only to the active pane.
	// Pass other messages (like ticks) to both components.
	if _, ok := msg.(tea.KeyMsg); ok {
		if m.activePane == listPane {
			m.list, cmd = m.list.Update(msg)
			cmds = append(cmds, cmd)
		} else {
			m.textarea, cmd = m.textarea.Update(msg)
			cmds = append(cmds, cmd)
		}
	} else {
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)

		m.textarea, cmd = m.textarea.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() tea.View {
	leftWidth := (m.width * 30) / 100

	leftBorderColor := lipgloss.Color("#3A3A8A")  // dimmer blue
	rightBorderColor := lipgloss.Color("#4D1F38") // dimmer pink

	switch m.activePane {
	case listPane:
		leftBorderColor = lipgloss.Color("#8787FF")
	case textareaPane:
		rightBorderColor = lipgloss.Color("#FF87D7")
	}

	leftPane := m.styles.leftPaneStyle.
		BorderForeground(leftBorderColor).
		Width(leftWidth).
		Height(m.height - m.styles.leftPaneStyle.GetVerticalFrameSize()).
		Render(m.list.View())

	rightWidth := m.width - lipgloss.Width(leftPane)

	var rightContent string
	selectedItem := m.list.SelectedItem()
	if selectedRoom, ok := selectedItem.(ModelID); ok {
		header := lipgloss.NewStyle().Bold(true).Render(selectedRoom.String())
		headerBlock := header + "\n\n"

		rightHeight := m.height - m.styles.rightPaneStyle.GetVerticalFrameSize()
		headerBlockHeight := 4
		taHeight := 3

		fillerLines := rightHeight - headerBlockHeight - taHeight
		if fillerLines < 0 {
			fillerLines = 0
		}
		filler := strings.Repeat("\n", fillerLines)

		rightContent = lipgloss.JoinVertical(lipgloss.Left, headerBlock, filler, m.textarea.View())
	} else {
		rightContent = "Select a model..."
	}

	rightPane := m.styles.rightPaneStyle.
		BorderForeground(rightBorderColor).
		Width(rightWidth - m.styles.rightPaneStyle.GetHorizontalFrameSize()).
		Height(m.height - m.styles.rightPaneStyle.GetVerticalFrameSize()).Render(rightContent)

	v := tea.NewView(lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane))
	v.AltScreen = true
	return v
}

func initialModel() model {

	items := []list.Item{
		OLLAMA,
		OPENAI,
		CLAUDE,
		GEMINI,
	}

	l := list.New(items, itemDelegate{styles: newStyles(true)}, 0, 0)

	ta := textarea.New()
	ta.Placeholder = "Ask me anything..."
	ta.ShowLineNumbers = false
	ta.SetHeight(3)
	ta.Blur()

	return model{
		list:       l,
		activePane: listPane,
		styles:     newStyles(true),
		textarea:   ta,
	}
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
