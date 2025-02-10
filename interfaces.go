package panelbubble

import (
	tea "github.com/charmbracelet/bubbletea"
	tcellviews "github.com/gdamore/tcell/v2/views"
)

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

type UpdateResponse struct {
	//NeedToRedraw   bool
	UpPropagateMsg tea.Msg
	Cmd            tea.Cmd
}

// CombineUpdateResponses combines a list of UpdateResponse objects into a single UpdateResponse object.
// It returns nil if all the UpdateResponse objects are nil.
// It returns the last UpPropagateMsg if present.
func CombineUpdateResponses(urs ...*UpdateResponse) *UpdateResponse {
	cmds := []tea.Cmd{}
	for _, ur := range urs {
		if ur.Cmd != nil {
			cmds = append(cmds, ur.Cmd)
		}
	}
	batchCmd := tea.Batch(cmds...)
	lastPropagateMsg := urs[len(urs)-1].UpPropagateMsg

	if batchCmd != nil || lastPropagateMsg != nil {
		return &UpdateResponse{
			Cmd:            batchCmd,
			UpPropagateMsg: lastPropagateMsg,
		}
	}
	return nil
}

func (ur *UpdateResponse) CombineHandleUpPropagate(handler func(msg tea.Msg) *UpdateResponse) *UpdateResponse {
	urs := []*UpdateResponse{ur}
	if ur != nil && ur.UpPropagateMsg != nil {
		urs = append(urs, handler(ur.UpPropagateMsg))
	}
	return CombineUpdateResponses(urs...)
}

type IModel interface {
	HandleFocusGranted() *UpdateResponse
	HandleFocusRevoked() *UpdateResponse
	HandleSizeMsg(msg ResizeMsg) *UpdateResponse
	Update(msg tea.Msg) *UpdateResponse
	Init() tea.Cmd
	SetView(view *tcellviews.ViewPort)
	Draw(force bool) bool
}

// IPanel is an interface that allows a tea model to be focused.
// It is used to handle focus passing in the UI.
type IPanel interface {
	IsFocused() bool
	GetPath() []int
	SetPath(path []int)
	IModel
}

// WorkflowHandlerInterface is an interface that allows a tea model to be a workflow handler.
// A workflow handler is a panel that is part of a workflow.
type WorkflowHandlerInterface interface {
	HandleFocusGrant(model tea.Model, msg FocusGrantMsg) (tea.Model, tea.Cmd)
	GetNumber() int
	GetWorkflowName() string
	IsFirst() bool
	IsLast() bool
}

/*
type HandlesSizeMsg interface {
	HandleSizeMsg(msg ResizeMsg) (tea.Model, tea.Cmd)
}
*/
