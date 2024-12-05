package panelbubble

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type PanelStyle struct {
	FocusedBorder   lipgloss.Style
	UnfocusedBorder lipgloss.Style
}

type ShortCutPanelConfig struct {
	GlobalShortcut              string
	LocalShortcut               string
	ContextualHelp              string
	Title                       string
	TitleStyle                  lipgloss.Style
	PanelStyle                  PanelStyle
	EnableWorkflowFocusMovement bool
}

// ShortCutPanel is an extension of Panel that can handle
// -- global and local shortcuts
// -- focus movement on key strokes -- up, down, backspace and enter
// For focus movement to work, the child panel must send the key-strokes
// to this panel via GetMsgForParent()
// It listens for ConsiderForGlobalShortcutMsg and ConsiderForLocalShortcutMsg
// for local and global shortcuts respectively
// When focused, it sends a ContextualHelpTextMsg to the parent
// to set ContextualHelp. The parent can decide to display it
// or not
type ShortCutPanel struct {
	ShortCutPanelConfig
	Panel
}

var _ tea.Model = &ShortCutPanel{}
var _ Focusable = &ShortCutPanel{}
var _ CanSendMsgToParent = &ShortCutPanel{}

func (p ShortCutPanel) Init() tea.Cmd {
	DebugPrintf("ShortCutPanel.Init() called for %v\n", p.GetPath())
	var batchCmds []tea.Cmd
	cmd := p.Panel.Init()
	if cmd != nil {
		batchCmds = append(batchCmds, cmd)
	}
	if p.IsFocused() {
		batchCmds = append(batchCmds, func() tea.Msg {
			return ContextualHelpTextMsg{Text: p.ContextualHelp}
		})
	}
	return tea.Batch(batchCmds...)
}

func (p ShortCutPanel) View() string {
	borderStyle := p.PanelStyle.UnfocusedBorder
	if p.IsFocused() {
		borderStyle = p.PanelStyle.FocusedBorder
	}
	return borderStyle.Render(
		lipgloss.JoinVertical(lipgloss.Top, p.TitleStyle.Render(p.Title), p.Panel.View()),
	)
}

func (p *ShortCutPanel) HandleMessageFromChild(msg tea.Msg) tea.Cmd {
	DebugPrintf("ShortCutPanel received message from child: %T %+v\n", msg, msg)
	if msg, ok := msg.(tea.KeyMsg); ok {
		if p.Panel.Workflow != nil {
			switch msg.String() {
			case "enter", "down":
				if !p.Workflow.IsLast() {
					path := []int{p.Workflow.GetNumber()}
					return func() tea.Msg {
						return FocusRequestMsg{
							RequestedPath: path,
							Relation:      NextWorkflow,
							WorkflowName:  p.Workflow.GetWorkflowName(),
						}
					}
				}
			case "backspace", "up":
				if !p.Workflow.IsFirst() {
					path := []int{p.Workflow.GetNumber()}
					return func() tea.Msg {
						return FocusRequestMsg{
							RequestedPath: path,
							Relation:      PrevWorkflow,
							WorkflowName:  p.Workflow.GetWorkflowName(),
						}
					}
				}
			}
		}
		return func() tea.Msg {
			return ConsiderForLocalShortcutMsg{Msg: msg}
		}
	}
	return nil
}

func (p ShortCutPanel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	DebugPrintf("ShortCutPanel.Update() called for %v\n", p.GetPath())
	switch msg := msg.(type) {
	case ConsiderForGlobalShortcutMsg, ConsiderForLocalShortcutMsg:
		return p.HandleShortcuts(msg)
	case ResizeMsg:
		return p.HandleSizeMsg(msg)
	default:
		cmds := []tea.Cmd{}
		updatedPanel, cmd := p.Panel.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		p.Panel = updatedPanel.(Panel)
		updatedPanel, msg = p.Panel.GetMsgForParent()
		p.Panel = updatedPanel.(Panel)
		if msg != nil {
			cmd2 := p.HandleMessageFromChild(msg)
			if cmd2 != nil {
				cmds = append(cmds, cmd2)
			}
		}
		if len(cmds) > 0 {
			return p, tea.Batch(cmds...)
		}
		return p, nil
	}
}

func (p ShortCutPanel) HandleShortcuts(msg tea.Msg) (tea.Model, tea.Cmd) {
	DebugPrintf("ShortCutPanel received shortcut message: %T %+v\n", msg, msg)
	if msg, ok := msg.(ConsiderForGlobalShortcutMsg); ok {
		if p.GlobalShortcut != "" {
			if msg.Msg.String() == p.GlobalShortcut {
				return p, func() tea.Msg {
					return FocusRequestMsg{RequestedPath: p.GetPath(), Relation: Self}
				}
			}
		}
	}
	if msg, ok := msg.(ConsiderForLocalShortcutMsg); ok {
		if p.LocalShortcut != "" && msg.Msg.String() == p.LocalShortcut {
			return p, func() tea.Msg {
				return FocusRequestMsg{RequestedPath: p.GetPath(), Relation: Self}
			}
		}
	}

	return p, nil
}

func GetStylingSize(s lipgloss.Style) (int, int) {
	return s.GetHorizontalMargins() + s.GetHorizontalPadding() + s.GetHorizontalBorderSize(),
		s.GetVerticalMargins() + s.GetVerticalPadding() + s.GetVerticalBorderSize()
}

func (p ShortCutPanel) HandleSizeMsg(msg ResizeMsg) (tea.Model, tea.Cmd) {
	DebugPrintf("ShortCutPanel received size message: %+v\n", msg)
	if model, ok := p.Panel.Model.(HandlesSizeMsg); ok {
		_, tvert := GetStylingSize(p.TitleStyle)
		text_height := 1 + tvert
		horz, vert := GetStylingSize(p.PanelStyle.UnfocusedBorder)
		width := msg.Width - horz
		height := msg.Height - vert - text_height
		//p.TitleStyle = p.TitleStyle.Width(width)
		updatedModel, cmd := model.HandleSizeMsg(ResizeMsg{Msg: msg, Width: width, Height: height})
		p.Panel.Model = updatedModel
		if cmd != nil {
			return p, cmd
		}
	}
	h, w := msg.Height-2, msg.Width-2
	p.PanelStyle.UnfocusedBorder = p.PanelStyle.UnfocusedBorder.Width(w)
	p.PanelStyle.FocusedBorder = p.PanelStyle.FocusedBorder.Width(w)
	p.PanelStyle.UnfocusedBorder = p.PanelStyle.UnfocusedBorder.Height(h)
	p.PanelStyle.FocusedBorder = p.PanelStyle.FocusedBorder.Height(h)
	return p, nil
}

func (p ShortCutPanel) SetPath(path []int) Focusable {
	//DebugPrintf("ShortCutPanel.SetPath() called for %v\n", path)
	p.Panel = p.Panel.SetPath(path).(Panel)
	return p
}
