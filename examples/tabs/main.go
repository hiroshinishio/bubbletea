package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	Tabs       []string
	TabContent []string
	activeTab  int
	styles     *styles
}

type styles struct {
	inactiveTabBorder lipgloss.Border
	activeTabBorder   lipgloss.Border
	highlightColor    lipgloss.AdaptiveColor
	docStyle          lipgloss.Style
	inactiveTabStyle  lipgloss.Style
	activeTabStyle    lipgloss.Style
	windowStyle       lipgloss.Style
}

func (m model) Init(ctx tea.Context) (tea.Model, tea.Cmd) {
	var (
		inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
		activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
		highlightColor    = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	)
	m.styles = &styles{
		inactiveTabBorder: inactiveTabBorder,
		activeTabBorder:   activeTabBorder,
		highlightColor:    highlightColor,
		docStyle:          ctx.NewStyle().Padding(1, 2, 1, 2),
		inactiveTabStyle:  ctx.NewStyle().Border(inactiveTabBorder, true).BorderForeground(highlightColor).Padding(0, 1),
		activeTabStyle:    ctx.NewStyle().Border(activeTabBorder, true).BorderForeground(highlightColor).Padding(0, 1),
		windowStyle:       ctx.NewStyle().BorderForeground(highlightColor).Padding(2, 0).Align(lipgloss.Center).Border(lipgloss.NormalBorder()).UnsetBorderTop(),
	}
	return m, nil
}

func (m model) Update(ctx tea.Context, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "right", "l", "n", "tab":
			m.activeTab = min(m.activeTab+1, len(m.Tabs)-1)
			return m, nil
		case "left", "h", "p", "shift+tab":
			m.activeTab = max(m.activeTab-1, 0)
			return m, nil
		}
	}

	return m, nil
}

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

func (m model) View(ctx tea.Context) string {
	doc := strings.Builder{}

	var renderedTabs []string

	for i, t := range m.Tabs {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(m.Tabs)-1, i == m.activeTab
		if isActive {
			style = m.styles.activeTabStyle.Copy()
		} else {
			style = m.styles.inactiveTabStyle.Copy()
		}
		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		} else if isLast && isActive {
			border.BottomRight = "│"
		} else if isLast && !isActive {
			border.BottomRight = "┤"
		}
		style = style.Border(border)
		renderedTabs = append(renderedTabs, style.Render(t))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	doc.WriteString("\n")
	doc.WriteString(m.styles.windowStyle.Width((lipgloss.Width(row) - m.styles.windowStyle.GetHorizontalFrameSize())).Render(m.TabContent[m.activeTab]))
	return m.styles.docStyle.Render(doc.String())
}

func main() {
	tabs := []string{"Lip Gloss", "Blush", "Eye Shadow", "Mascara", "Foundation"}
	tabContent := []string{"Lip Gloss Tab", "Blush Tab", "Eye Shadow Tab", "Mascara Tab", "Foundation Tab"}
	m := model{Tabs: tabs, TabContent: tabContent}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
