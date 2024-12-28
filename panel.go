package panelbubble

import (
	tea "github.com/charmbracelet/bubbletea"
	tcellviews "github.com/gdamore/tcell/v2/views"
)

type Panel struct {
	Model        tea.Model
	focus        bool
	path         []int // Path to uniquely identify this node in the hierarchy
	Name         string
	Workflow     WorkflowHandlerInterface
	MsgForParent tea.Msg
	view         tcellviews.View
	redraw       bool
}

func (p Panel) SetView(view *tcellviews.ViewPort) Focusable {
	p.view = view
	return p
}

func (p Panel) Draw(force bool) (Focusable, bool) {
	DebugPrintf("Panel.Draw() called for %v. Redraw: %v, force: %v\n", p.path, p.redraw, force)
	str := p.View()
	if p.view != nil {
		tcellDrawHelper(str, p.view)
		return p, true
	}
	return p, false
}

func (p *Panel) SetMsgForParent(msg tea.Msg) {
	p.MsgForParent = msg
}

func (p Panel) GetMsgForParent() (tea.Model, tea.Msg) {
	msg := p.MsgForParent
	p.MsgForParent = nil
	return p, msg
}

var _ CanSendMsgToParent = &Panel{}
var _ tea.Model = &Panel{}
var _ Focusable = &Panel{}

// var _ HandlesRecvFocus = &Panel{}
// var _ HandlesFocusRevoke = &Panel{}

func (p Panel) IsFocused() bool {
	return p.focus
}

func (p Panel) SetPath(path []int) Focusable {
	p.path = path
	return p
}

func (p Panel) Init() tea.Cmd {
	DebugPrintf("Panel.Init() called for %v\n", p.GetPath())
	cmd := p.Model.Init()
	if cmd != nil {
		return p.RoutedCmd(cmd)
	}
	return nil
}

func (p Panel) View() string {
	return p.Model.View()
}

func (p *Panel) RoutedCmd(cmd tea.Cmd) tea.Cmd {
	return MakeAutoRoutedCmd(cmd, p.path)
}

func (p Panel) GetPath() []int {
	return p.path
}

func (p Panel) HandleMessage(msg tea.Msg) (Focusable, tea.Cmd) {
	DebugPrintf("Panel %v received message: %T %+v\n", p.path, msg, msg)
	updatedModel, cmd := p.Model.Update(msg)
	p.Model = updatedModel

	if model, ok := p.Model.(CanSendMsgToParent); ok {
		updatedModel, msgForParent := model.GetMsgForParent()
		p.Model = updatedModel
		if msgForParent != nil {
			p.SetMsgForParent(msgForParent)
		}
	}

	return p, p.RoutedCmd(cmd)
}

func (p Panel) HandleFocusGranted(msg FocusGrantMsg) (Focusable, tea.Cmd) {
	DebugPrintf("Panel %v received focus grant message: %T %+v\n", p.path, msg, msg)
	p.focus = true
	p.redraw = true
	if p.Model != nil {
		if model, ok := p.Model.(HandlesRecvFocus); ok {
			updatedModel, cmd := model.HandleRecvFocus()
			p.Model = updatedModel
			if cmd != nil {
				return p, cmd
			}
		}
	}
	return p, nil
}

func (p Panel) HandleFocusRevoke() (Focusable, tea.Cmd) {
	DebugPrintf("Panel %v received focus revoke message\n", p.path)
	if p.focus {
		p.redraw = true
	}
	p.focus = false
	if p.Model != nil {
		if model, ok := p.Model.(HandlesFocusRevoke); ok {
			updatedModel, cmd := model.HandleRecvFocusRevoke()
			p.Model = updatedModel
			if cmd != nil {
				return p, cmd
			}
		}
	}
	return p, nil
}

func (p Panel) HandleFocus(msg tea.Msg) (Focusable, tea.Cmd) {
	DebugPrintf("Panel %v received focus message: %T %+v\n", p.path, msg, msg)
	switch msg := msg.(type) {
	case FocusGrantMsg:
		//fmt.Printf("Panel received focus grant: %v\n", msg)
		switch msg.Relation {
		case Self:
			if IsSamePath(msg.Path, p.path) {
				// Focus grant has reached the target node, now set focus
				return p.HandleFocusGranted(msg)
			}

		case NextWorkflow, PrevWorkflow:
			var cmd tea.Cmd
			if p.Workflow != nil {
				//fmt.Printf("Panel received focus grant for workflow: %v\n", msg.WorkflowName)
				p.Model, cmd = (p.Workflow).HandleFocusGrant(p.Model, msg)
				if cmd != nil {
					return p.HandleFocusGranted(msg)
				}
			}
			return p, cmd
		}
	case FocusRevokeMsg:
		//fmt.Printf("Panel received focus revoke\n")
		// Revoke focus from this panel
		return p.HandleFocusRevoke()
	}
	return p, nil
}

func (p Panel) HandleSizeMsg(msg ResizeMsg) (tea.Model, tea.Cmd) {
	if p.view != nil {
		p.view.Resize(msg.X, msg.Y, msg.Width, msg.Height)
	}
	if model, ok := p.Model.(HandlesSizeMsg); ok {
		updatedModel, cmd := model.HandleSizeMsg(msg)
		p.Model = updatedModel
		if cmd != nil {
			return p, cmd
		}
	}
	return p, nil
}

func (p Panel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	DebugPrintf("Panel %v received message: %T %+v\n", p.path, msg, msg)
	switch msg := (msg).(type) {
	case FocusGrantMsg, FocusRevokeMsg:
		return p.HandleFocus(msg)
	case AutoRoutedMsg:
		return p.HandleMessage(msg.Msg)
	case ResizeMsg:
		return p.HandleSizeMsg(msg)
	case BroadcastMsg:
		return p.HandleMessage(msg.Msg)
	default:
		return p.HandleMessage(msg)
	}
}
