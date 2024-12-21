package panelbubble

import (
	tea "github.com/charmbracelet/bubbletea"
	tcellviews "github.com/gdamore/tcell/v2/views"
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
	ListPanel
}

var _ tea.Model = &TopLevelListPanel{}
var _ Focusable = &TopLevelListPanel{}

func (m TopLevelListPanel) SetView(viewport *tcellviews.ViewPort) Focusable {
	newListPanel := m.ListPanel.SetView(viewport)
	m.ListPanel = newListPanel.(ListPanel)
	return m
}

func (m TopLevelListPanel) Draw(force bool) Focusable {
	newListPanel := m.ListPanel.Draw(force)
	m.ListPanel = newListPanel.(ListPanel)
	return m
}

func (m TopLevelListPanel) Init() tea.Cmd {
	m.ListPanel.SetPath([]int{})
	return m.ListPanel.Init()
}

func (m TopLevelListPanel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	DebugPrintf("TopLevelListPanel received message: %T %+v\n", msg, msg)
	switch msg := msg.(type) {
	case FocusRequestMsg:
		cmds := []tea.Cmd{}
		updatedModel, cmd := m.ListPanel.Update(FocusRevokeMsg{})
		m.ListPanel = updatedModel.(ListPanel)

		if cmd != nil {
			cmds = append(cmds, cmd)
		}

		updatedModel, cmd = m.ListPanel.Update(FocusGrantMsg{RoutePath: RoutePath{Path: msg.RequestedPath}, Relation: msg.Relation, WorkflowName: msg.WorkflowName})
		m.ListPanel = updatedModel.(ListPanel)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		// We propagate key messages as global shortcuts initially
		// if no panel consumes it, we will send it out as a propagate key message
		globalShortcutMsg := ConsiderForGlobalShortcutMsg{Msg: msg}
		updatedModel, cmd := m.ListPanel.Update(globalShortcutMsg)
		m.ListPanel = updatedModel.(ListPanel)
		if cmd != nil {
			return m, cmd
		} else {
			return m, func() tea.Msg {
				return PropagateKeyMsg{KeyMsg: msg}
			}
		}

	case PropagateKeyMsg:
		updatedModel, cmd := m.ListPanel.Update(msg.KeyMsg)
		m.ListPanel = updatedModel.(ListPanel)
		if cmd != nil {
			return m, cmd
		}

	default:
		updatedModel, cmd := m.ListPanel.Update(msg)
		m.ListPanel = updatedModel.(ListPanel)
		return m, cmd
	}

	return m, nil
}
