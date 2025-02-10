package panelbubble

import (
	tea "github.com/charmbracelet/bubbletea"
	tcellviews "github.com/gdamore/tcell/v2/views"
)

type Panel struct {
	Model        IModel
	focus        bool
	path         []int // Path to uniquely identify this node in the hierarchy
	Name         string
	Workflow     WorkflowHandlerInterface
	MsgForParent tea.Msg
	view         tcellviews.View
	//redraw       bool
}

type PanelOption func(*Panel)

func WithModel(model IModel) PanelOption {
	return func(panel *Panel) {
		panel.Model = model
	}
}

func WithName(name string) PanelOption {
	return func(panel *Panel) {
		panel.Name = name
	}
}

func NewPanel(opts ...PanelOption) *Panel {
	panel := &Panel{}
	for _, opt := range opts {
		opt(panel)
	}
	return panel
}

func (p *Panel) SetView(view *tcellviews.ViewPort) {
	p.view = view
}

func (p *Panel) Draw(force bool) bool {
	redraw := p.Model.Draw(force)
	DebugPrintf("Panel.Draw() called for %v. Redraw: %v, force: %v\n", p.path, redraw, force)
	return redraw
}

func (p *Panel) SetMsgForParent(msg tea.Msg) {
	p.MsgForParent = msg
}

func (p *Panel) GetMsgForParent() tea.Msg {
	msg := p.MsgForParent
	p.MsgForParent = nil
	return msg
}

var _ IPanel = &Panel{}

func (p *Panel) IsFocused() bool {
	return p.focus
}

func (p *Panel) SetPath(path []int) {
	p.path = path
}

func (p *Panel) Init() tea.Cmd {
	DebugPrintf("Panel.Init() called for %v\n", p.GetPath())
	cmd := p.Model.Init()
	if cmd != nil {
		return p.RoutedCmd(cmd)
	}
	return nil
}

/* func (p *Panel) View() string {
	return p.Model.View()
} */

func (p *Panel) RoutedCmd(cmd tea.Cmd) tea.Cmd {
	return MakeAutoRoutedCmd(cmd, p.path)
}

func (p *Panel) GetPath() []int {
	return p.path
}

func (p *Panel) HandleMessage(msg tea.Msg) *UpdateResponse {
	DebugPrintf("Panel %v received message: %T %+v\n", p.path, msg, msg)
	ur := p.Model.Update(msg)
	//p.redraw = ur.NeedToRedraw
	return &UpdateResponse{
		//NeedToRedraw:   ur.NeedToRedraw,
		UpPropagateMsg: ur.UpPropagateMsg,
		Cmd:            p.RoutedCmd(ur.Cmd),
	}
}

func (p *Panel) HandleFocusGranted() *UpdateResponse {
	DebugPrintf("Panel %v received focus grant message\n", p.path)
	p.focus = true
	//p.redraw = true
	if p.Model != nil {
		return p.Model.HandleFocusGranted()
	}
	return nil
}

func (p *Panel) HandleFocusRevoked() *UpdateResponse {
	DebugPrintf("Panel %v received focus revoke message\n", p.path)
	//if p.focus {
	//	p.redraw = true
	//}
	p.focus = false
	if p.Model != nil {
		return p.Model.HandleFocusRevoked()
	}
	return nil
}

func (p *Panel) HandleFocus(msg tea.Msg) *UpdateResponse {
	DebugPrintf("Panel %v received focus message: %T %+v\n", p.path, msg, msg)
	switch msg := msg.(type) {
	case FocusGrantMsg:
		//fmt.Printf("Panel received focus grant: %v\n", msg)
		switch msg.Relation {
		case Self:
			if IsSamePath(msg.Path, p.path) {
				// Focus grant has reached the target node, now set focus
				return p.HandleFocusGranted()
			}
		}
	case FocusRevokeMsg:
		//fmt.Printf("Panel received focus revoke\n")
		// Revoke focus from this panel
		return p.HandleFocusRevoked()
	}
	return nil
}

func (p *Panel) HandleSizeMsg(msg ResizeMsg) *UpdateResponse {
	/*if p.view != nil {
		p.view.Resize(msg.X, msg.Y, msg.Width, msg.Height)
	}*/
	return p.Model.HandleSizeMsg(msg)
}

func (p *Panel) Update(msg tea.Msg) *UpdateResponse {
	DebugPrintf("Panel %v received message: %T %+v\n", p.path, msg, msg)
	switch msg := msg.(type) {
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
