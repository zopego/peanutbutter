package peanutbutter

import (
	tea "github.com/charmbracelet/bubbletea"
)

// TopLevelListPanel is a special case of ListPanel
// that is the topmost panel in the hierarchy
// It handles responding to focus-request messages
// with focus-grant messages
// It also initializes the path of all its children
// And gives all children panels a chance to request
// focus for key-strokes by passing them a ConsiderForGlobalShortcutMsg
// before passing a key-stroke as a regular key-stroke message
type TopLevelListPanel struct {
	*ListPanel
	cmds chan tea.Cmd
}

var _ IPanel = &TopLevelListPanel{}

func (m *TopLevelListPanel) Init(cmds chan tea.Cmd) {
	m.ListPanel.SetPath([]int{})
	m.cmds = cmds
	m.ListPanel.Init(cmds)
}

func (m *TopLevelListPanel) FigureOutFocusGrant(msg FocusRequestMsg) *FocusGrantMsg {
	switch msg.Relation {
	case Self:
		return &FocusGrantMsg{RoutePath: RoutePath{Path: msg.RequestedPath}, Relation: msg.Relation}
	case Left, Right, Up, Down:
		return nil
	default:
		return nil
	}
}

func (m *TopLevelListPanel) HandleMessage(msg Msg) {
	DebugPrintf("TopLevelListPanel received message: %T %+v\n", msg, msg)
	switch msg := msg.(type) {
	case FocusRequestMsg:
		m.ListPanel.HandleMessage(FocusRevokeMsg{})
		focusGrantMsg := m.FigureOutFocusGrant(msg)
		if focusGrantMsg != nil {
			newCmd := func() tea.Msg {
				return *focusGrantMsg
			}
			m.cmds <- newCmd
		}

	default:
		m.ListPanel.HandleMessage(msg)
	}
}
