package peanutbutter

import (
	tea "github.com/charmbracelet/bubbletea"
	tcell "github.com/gdamore/tcell/v2"
)

func SingleRuneBinding(rune rune) KeyBinding {
	return *NewKeyBinding(
		WithKeyDef(KeyDef{Key: tcell.KeyRune, Modifiers: tcell.ModMask(0), Rune: rune}),
		WithEnabled(true),
		WithShortHelp(""),
		WithLongHelp(""),
	)
}

func SingleKeyBinding(key tcell.Key) KeyBinding {
	return *NewKeyBinding(
		WithKeyDef(KeyDef{Key: key, Modifiers: tcell.ModMask(0), Rune: 0}),
		WithEnabled(true),
		WithShortHelp(""),
		WithLongHelp(""),
	)
}

type StructTabFocus struct {
	Index  int
	Panels []*ShortCutPanel
}

func (m *StructTabFocus) HandleTabMsg(n int) tea.Msg {
	n = (n + 1) % len(m.Panels)
	path := m.Panels[n].GetPath()
	return FocusRequestMsg{
		Relation:      Self,
		RequestedPath: path,
	}
}

func (m *StructTabFocus) AddPanel(panel *ShortCutPanel) KeyBinding {
	m.Panels = append(m.Panels, panel)
	n := len(m.Panels) - 1
	kb := SingleKeyBinding(tcell.KeyTAB)
	kb.Func = func() tea.Cmd {
		return func() tea.Msg {
			return m.HandleTabMsg(n)
		}
	}
	return kb
}
