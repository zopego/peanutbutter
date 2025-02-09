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

type PanelOption func(*Panel)

func WithModel(model tea.Model) PanelOption {
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
	DebugPrintf("Panel.Draw() called for %v. Redraw: %v, force: %v\n", p.path, p.redraw, force)
	str := p.View()
	if p.view != nil {
		tcellDrawHelper(str, p.view)
		return true
	}
	return false
}

func (p *Panel) SetMsgForParent(msg tea.Msg) {
	p.MsgForParent = msg
}

func (p *Panel) GetMsgForParent() tea.Msg {
	msg := p.MsgForParent
	p.MsgForParent = nil
	return msg
}

var _ CanSendMsgToParent = &Panel{}
var _ tea.Model = &Panel{}
var _ Focusable = &Panel{}

// var _ HandlesRecvFocus = &Panel{}
// var _ HandlesFocusRevoke = &Panel{}

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

func (p *Panel) View() string {
	return p.Model.View()
}

func (p *Panel) RoutedCmd(cmd tea.Cmd) tea.Cmd {
	return MakeAutoRoutedCmd(cmd, p.path)
}

func (p *Panel) GetPath() []int {
	return p.path
}

func (p *Panel) HandleMessage(msg tea.Msg) tea.Cmd {
	DebugPrintf("Panel %v received message: %T %+v\n", p.path, msg, msg)
	updatedModel, cmd := p.Model.Update(msg)
	p.Model = updatedModel

	if model, ok := p.Model.(CanSendMsgToParent); ok {
		msgForParent := model.GetMsgForParent()
		p.Model = updatedModel
		if msgForParent != nil {
			p.SetMsgForParent(msgForParent)
		}
	}

	return p.RoutedCmd(cmd)
}

func (p *Panel) HandleFocusGranted(msg FocusGrantMsg) tea.Cmd {
	DebugPrintf("Panel %v received focus grant message: %T %+v\n", p.path, msg, msg)
	p.focus = true
	p.redraw = true
	if p.Model != nil {
		if model, ok := p.Model.(HandlesRecvFocus); ok {
			updatedModel, cmd := model.HandleRecvFocus()
			p.Model = updatedModel
			if cmd != nil {
				return cmd
			}
		}
	}
	return nil
}

func (p *Panel) HandleFocusRevoke() tea.Cmd {
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
				return cmd
			}
		}
	}
	return nil
}

func (p *Panel) HandleFocus(msg tea.Msg) tea.Cmd {
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
			return cmd
		}
	case FocusRevokeMsg:
		//fmt.Printf("Panel received focus revoke\n")
		// Revoke focus from this panel
		return p.HandleFocusRevoke()
	}
	return nil
}

func (p *Panel) HandleSizeMsg(msg ResizeMsg) tea.Cmd {
	if p.view != nil {
		p.view.Resize(msg.X, msg.Y, msg.Width, msg.Height)
	}
	if model, ok := p.Model.(HandlesSizeMsg); ok {
		updatedModel, cmd := model.HandleSizeMsg(msg)
		p.Model = updatedModel
		if cmd != nil {
			return cmd
		}
	}
	return nil
}

func (p *Panel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	DebugPrintf("Panel %v received message: %T %+v\n", p.path, msg, msg)
	switch msg := msg.(type) {
	case FocusGrantMsg, FocusRevokeMsg:
		cmd := p.HandleFocus(msg)
		return p, cmd
	case AutoRoutedMsg:
		cmd := p.HandleMessage(msg.Msg)
		return p, cmd
	case ResizeMsg:
		cmd := p.HandleSizeMsg(msg)
		return p, cmd
	case BroadcastMsg:
		cmd := p.HandleMessage(msg.Msg)
		return p, cmd
	default:
		cmd := p.HandleMessage(msg)
		return p, cmd
	}
}
