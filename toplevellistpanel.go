package panelbubble

import (
	"time"

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
	cmds               chan tea.Cmd
	MarkMessageNotUsed func(msg *KeyMsg)
}

var _ IPanel = &TopLevelListPanel{}

func (m *TopLevelListPanel) Init(cmds chan tea.Cmd, MarkMessageNotUsed func(msg *KeyMsg)) {
	m.ListPanel.SetPath([]int{})
	m.MarkMessageNotUsed = MarkMessageNotUsed
	m.cmds = cmds
	m.ListPanel.Init(cmds, m.MessageNotUsedInternal)
}

func (m *TopLevelListPanel) MessageNotUsedInternal(msg *KeyMsg) {
	m.MarkMessageNotUsed(msg)
}

func (m *TopLevelListPanel) FigureOutFocusGrant(msg FocusRequestMsg) *FocusGrantMsg {
	switch msg.Relation {
	case Self:
		return &FocusGrantMsg{RoutePath: &RoutePath{Path: msg.RequestedPath}, Relation: msg.Relation}
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

	case *tcell.EventKey:
		k := KeyMsg{EventKey: msg, Id: time.Now().UnixNano()}
		m.ListPanel.HandleMessage(k)

	default:
		m.ListPanel.HandleMessage(msg)
	}
}
