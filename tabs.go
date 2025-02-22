package peanutbutter

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TabBar struct {
	TabNames []string
	TabStyles
	ListPanelName string
	selected      int
}

var _ tea.Model = &TabBar{}

type TabStyles struct {
	SelectedTabStyle   lipgloss.Style
	UnselectedTabStyle lipgloss.Style
	RenderFunc         func(...string) string
}

func (m TabBar) Init() tea.Cmd {
	return nil
}

func (m TabBar) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(SelectedTabIndexMsg); ok {
		if msg.ListPanelName == m.ListPanelName {
			m.selected = msg.Index
			return m, nil
		}
	}
	return m, nil
}

func (m TabBar) View() string {
	views := []string{}

	for i, tab := range m.TabNames {
		style := m.TabStyles.UnselectedTabStyle
		if i == m.selected {
			style = m.TabStyles.SelectedTabStyle
		}
		views = append(views, style.Render(tab))
	}
	return m.TabStyles.RenderFunc(views...)
}

func (m TabBar) Selected() int {
	return m.selected
}
