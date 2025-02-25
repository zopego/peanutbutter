package peanutbutter

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TabBar struct {
	ListPanel *ListPanel
	TabNext   KeyBinding
	TabPrev   KeyBinding
	TabStyles
	needsRedraw bool
}

var _ ILeafModel = &TabBar{}
var _ ILeafModelWithView = &TabBar{}

type TabStyles struct {
	SelectedTabStyle   lipgloss.Style
	UnselectedTabStyle lipgloss.Style
	RenderFunc         func(...string) string
}

func NewTabBar(listPanel *ListPanel, opts ...TabBarOption) *TabBar {
	renderfunc := func(views ...string) string {
		return lipgloss.JoinHorizontal(lipgloss.Left, views...)
	}

	t := &TabBar{
		needsRedraw: true,
		ListPanel:   listPanel,
		TabStyles: TabStyles{
			SelectedTabStyle:   lipgloss.NewStyle().Bold(true).Padding(0, 1).Underline(true),
			UnselectedTabStyle: lipgloss.NewStyle().Bold(false).Padding(0, 1),
			RenderFunc:         renderfunc,
		},
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

type TabBarOption func(*TabBar)

func WithTabStyles(tabStyles TabStyles) TabBarOption {
	return func(m *TabBar) {
		m.TabStyles = tabStyles
	}
}

func WithTabNext(tabNext KeyBinding) TabBarOption {
	return func(m *TabBar) {
		m.TabNext = tabNext
	}
}

func WithTabPrev(tabPrev KeyBinding) TabBarOption {
	return func(m *TabBar) {
		m.TabPrev = tabPrev
	}
}

func (m TabBar) Init() tea.Cmd {
	return nil
}

func (m TabBar) Update(msg Msg) tea.Cmd {
	if msg, ok := msg.(KeyMsg); ok {
		if m.TabNext.IsMatch(msg.EventKey) {
			m.needsRedraw = true
			return m.ListPanel.TabNext()
		}
		if m.TabPrev.IsMatch(msg.EventKey) {
			m.needsRedraw = true
			return m.ListPanel.TabPrev()
		}
		msg.SetUnused()
	}
	return nil
}

func (m TabBar) View() string {
	views := []string{}
	for i, tab := range m.ListPanel.Panels {
		style := m.TabStyles.UnselectedTabStyle
		if i == m.ListPanel.Selected {
			style = m.TabStyles.SelectedTabStyle
		}
		views = append(views, style.Render(tab.GetName()))
	}
	m.needsRedraw = false
	return m.TabStyles.RenderFunc(views...)
}

func (m *TabBar) NeedsRedraw() bool {
	return m.needsRedraw
}
