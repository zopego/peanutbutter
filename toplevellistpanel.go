package panelbubble

import (
	tea "github.com/charmbracelet/bubbletea"
	tcell "github.com/gdamore/tcell/v2"
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
}

var _ tea.Model = &TopLevelListPanel{}
var _ Focusable = &TopLevelListPanel{}

func (m *TopLevelListPanel) Init() tea.Cmd {
	m.ListPanel.SetPath([]int{})
	return m.ListPanel.Init()
}

func (m *TopLevelListPanel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	DebugPrintf("TopLevelListPanel received message: %T %+v\n", msg, msg)
	switch msg := msg.(type) {
	case FocusRequestMsg:
		cmds := []tea.Cmd{}
		_, cmd := m.ListPanel.Update(FocusRevokeMsg{})

		if cmd != nil {
			cmds = append(cmds, cmd)
		}

		// This ensures that a complete update is done before we send a focus grant message
		newCmd := func() tea.Msg {
			return FocusGrantMsg{RoutePath: RoutePath{Path: msg.RequestedPath}, Relation: msg.Relation, WorkflowName: msg.WorkflowName}
		}
		cmds = append(cmds, newCmd)
		return m, tea.Batch(cmds...)

	case *tcell.EventKey:
		// We propagate key messages as global shortcuts initially
		// if no panel consumes it, we will send it out as a propagate key message
		globalShortcutMsg := ConsiderForGlobalShortcutMsg{EventKey: msg}
		_, cmd := m.ListPanel.Update(globalShortcutMsg)
		if cmd != nil {
			return m, cmd
		} else {
			return m, func() tea.Msg {
				return PropagateKeyMsg{EventKey: msg}
			}
		}

	case PropagateKeyMsg:
		_, cmd := m.ListPanel.Update(msg.EventKey)
		return m, cmd

	default:
		_, cmd := m.ListPanel.Update(msg)
		return m, cmd
	}

	return m, nil
}
