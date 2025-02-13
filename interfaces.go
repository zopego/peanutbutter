package panelbubble

import (
	tea "github.com/charmbracelet/bubbletea"
	tcellviews "github.com/gdamore/tcell/v2/views"
)

func IsSamePath(path1, path2 []int) bool {
	if len(path1) != len(path2) {
		return false
	}
	for i, v := range path1 {
		if v != path2[i] {
			return false
		}
	}
	return true
}

type IRootModel interface {
	Update(msg Msg)
	Init(cmds chan tea.Cmd, view *tcellviews.ViewPort) tea.Cmd
	Draw() bool
}

type ILeafModel interface {
	Update(msg Msg) tea.Cmd
	Init(MarkMessageNotUsed func(msg *KeyMsg)) tea.Cmd
	//SetView(view *tcellviews.ViewPort)
	//Draw(force bool) bool
}

type ILeafModelWithView interface {
	NeedsRedraw() bool
	View() string
}

type ILeafModelWithDraw interface {
	Draw(force bool, view *tcellviews.ViewPort) bool
}

type PanelCenter struct {
	X    int
	Y    int
	Path []int
}

// IPanel is an interface that allows a tea model to be focused.
// It is used to handle focus passing in the UI.
type IPanel interface {
	IsFocused() bool
	GetPath() []int
	SetPath(path []int)
	HandleMessage(msg Msg)
	SetView(view *tcellviews.ViewPort)
	Draw(force bool) bool
	Init(cmds chan tea.Cmd, MarkMessageNotUsed func(msg *KeyMsg))
	GetLeafPanelCenters() []PanelCenter
}

// WorkflowHandlerInterface is an interface that allows a tea model to be a workflow handler.
// A workflow handler is a panel that is part of a workflow.
/* type WorkflowHandlerInterface interface {
	HandleFocusGrant(model tea.Model, msg FocusGrantMsg) (tea.Model, tea.Cmd)
	GetNumber() int
	GetWorkflowName() string
	IsFirst() bool
	IsLast() bool
} */

/*
type HandlesSizeMsg interface {
	HandleSizeMsg(msg ResizeMsg) (tea.Model, tea.Cmd)
}
*/

// Initiable is an interface that allows calling Init() on
// things that implement the Init() method but might not be a tea.Model
/* type DoesInit interface {
	Init() tea.Cmd
} */

// A Tea Model that can handle focus grants can implement this interface
// The parent panel will call this method when it receives a focus grant
/* type HandlesRecvFocus interface {
	HandleRecvFocus() (tea.Model, tea.Cmd)
}*/

// A Tea Model that can handle focus revokes can implement this interface
// The parent panel will call this method when it receives a focus revoke
/* type HandlesFocusRevoke interface {
	HandleRecvFocusRevoke() (tea.Model, tea.Cmd)
} */

// This interfaces allows a tea model to directly pass a message to its
// parent panel. This mechanism is used to handle focus passing for example -
// when the down arrow key is pressed on the last item of a list, focus is passed
// to the next panel in the hierarchy.
/* type CanSendMsgToParent interface {
	GetMsgForParent() tea.Msg
} */ // => this is not needed anymore, we can just use HandlesUpdates
