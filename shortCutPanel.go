package panelbubble

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	tcell "github.com/gdamore/tcell/v2"
	tcellviews "github.com/gdamore/tcell/v2/views"
)

type PanelStyle struct {
	FocusedBorder   lipgloss.Style
	UnfocusedBorder lipgloss.Style
}

type ShortCutPanelConfig struct {
	GlobalShortcut              KeyBinding
	LocalShortcut               KeyBinding
	ContextualHelp              string
	Title                       string
	TitleStyle                  lipgloss.Style
	PanelStyle                  PanelStyle
	EnableWorkflowFocusMovement bool
	EnableHorizontalMovement    bool
	EnableVerticalMovement      bool
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
	*Panel
	redraw bool
}

type ShortCutPanelOption func(*ShortCutPanel)

func WithGlobalShortcut(shortcut KeyBinding) ShortCutPanelOption {
	return func(config *ShortCutPanel) {
		config.GlobalShortcut = shortcut
	}
}

func WithLocalShortcut(shortcut KeyBinding) ShortCutPanelOption {
	return func(config *ShortCutPanel) {
		config.LocalShortcut = shortcut
	}
}

func WithContextualHelp(help string) ShortCutPanelOption {
	return func(config *ShortCutPanel) {
		config.ContextualHelp = help
	}
}

func WithTitle(title string) ShortCutPanelOption {
	return func(config *ShortCutPanel) {
		config.Title = title
	}
}

func WithTitleStyle(style lipgloss.Style) ShortCutPanelOption {
	return func(config *ShortCutPanel) {
		config.TitleStyle = style
	}
}

func WithPanelStyle(style PanelStyle) ShortCutPanelOption {
	return func(config *ShortCutPanel) {
		config.PanelStyle = style
	}
}

func WithEnableMovement(horizontal, vertical bool) ShortCutPanelOption {
	return func(config *ShortCutPanel) {
		config.EnableHorizontalMovement = horizontal
		config.EnableVerticalMovement = vertical
	}
}

func WithConfig(config ShortCutPanelConfig) ShortCutPanelOption {
	return func(panel *ShortCutPanel) {
		panel.ShortCutPanelConfig = config
	}
}

func NewShortCutPanel(panel *Panel, opts ...ShortCutPanelOption) *ShortCutPanel {
	spanel := &ShortCutPanel{}
	for _, opt := range opts {
		opt(spanel)
	}
	if panel != nil {
		spanel.Panel = panel
	} else {
		spanel.Panel = NewPanel()
	}
	return spanel
}

// var _ tea.Model = &ShortCutPanel{}
var _ IPanel = &ShortCutPanel{}

//var _ CanSendMsgToParent = &ShortCutPanel{}

func (p *ShortCutPanel) SetView(view *tcellviews.ViewPort) {
	p.view = view
}

func (p *ShortCutPanel) Draw(force bool) bool {
	childRedraw := p.Panel.Draw(force)
	if p.redraw || force || childRedraw {
		DebugPrintf("ShortCutPanel.Draw() called for %v. Redraw: %v, force: %v\n", p.GetPath(), p.redraw, force)
		str := p.borderView()
		tcellDrawHelper(str, p.view)
		p.redraw = false
		return true
	}
	return false
}

/* func (p *ShortCutPanel) GetMsgForParent() tea.Msg {
	msg := p.Panel.GetMsgForParent()
	return msg
} */

/* func (p *ShortCutPanel) SetMsgForParent(msg tea.Msg) {
	p.Panel.SetMsgForParent(msg)
} */

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

func (p *ShortCutPanel) borderView() string {
	borderStyle := p.PanelStyle.UnfocusedBorder
	if p.IsFocused() {
		borderStyle = p.PanelStyle.FocusedBorder
	}
	return borderStyle.Render(
		lipgloss.JoinVertical(lipgloss.Top, p.TitleStyle.Render(p.Title)),
	)
}

func (p *ShortCutPanel) HandleMessageFromChild(msg tea.Msg) *UpdateResponse {
	DebugPrintf("ShortCutPanel received message from child: %T %+v\n", msg, msg)
	if msg, ok := msg.(*tcell.EventKey); ok {
		if p.EnableHorizontalMovement {
			switch msg.Key() {
			case tcell.KeyLeft, tcell.KeyRight:
				direction := Left
				if msg.Key() == tcell.KeyRight {
					direction = Right
				}
				p.Panel.SetMsgForParent(GeometricFocusRequestMsg{Direction: direction})
				return nil
			}
		}
		if p.EnableVerticalMovement {
			switch msg.Key() {
			case tcell.KeyUp, tcell.KeyDown:
				direction := Up
				if msg.Key() == tcell.KeyDown {
					direction = Down
				}
				p.Panel.SetMsgForParent(GeometricFocusRequestMsg{Direction: direction})
				return nil
			}
		}
		return &UpdateResponse{
			Cmd: func() tea.Msg {
				return ConsiderForLocalShortcutMsg{EventKey: msg}
			},
			UpPropagateMsg: nil,
		}
	}
	return nil
}

func (p *ShortCutPanel) Update(msg tea.Msg) *UpdateResponse {
	DebugPrintf("ShortCutPanel.Update() called for %v\n", p.GetPath())
	p.redraw = false
	return p.updateHelper(msg)
}

func (p *ShortCutPanel) updateHelper(msg tea.Msg) *UpdateResponse {
	switch msg := msg.(type) {
	case ConsiderForGlobalShortcutMsg, ConsiderForLocalShortcutMsg:
		return p.HandleShortcuts(msg)
	case ResizeMsg:
		return p.HandleSizeMsg(msg)
	default:
		ur1 := p.Panel.Update(msg)
		ur2 := ur1.HandleUpPropagate(p.HandleMessageFromChild)
		return CombineUpdateResponses(ur1, ur2)
	}
}

func (p *ShortCutPanel) HandleShortcuts(msg tea.Msg) *UpdateResponse {
	DebugPrintf("ShortCutPanel received shortcut message: %T %+v\n", msg, msg)
	if msg, ok := msg.(ConsiderForGlobalShortcutMsg); ok {
		if p.GlobalShortcut.IsMatch(msg.EventKey) {
			return &UpdateResponse{
				Cmd: func() tea.Msg {
					return FocusRequestMsg{RequestedPath: p.GetPath(), Relation: Self}
				},
				UpPropagateMsg: nil,
			}
		}
	}
	if msg, ok := msg.(ConsiderForLocalShortcutMsg); ok {
		if p.LocalShortcut.IsMatch(msg.EventKey) {
			return &UpdateResponse{
				Cmd: func() tea.Msg {
					return FocusRequestMsg{RequestedPath: p.GetPath(), Relation: Self}
				},
				UpPropagateMsg: nil,
			}
		}
	}
	return nil
}

func GetStylingSize(s lipgloss.Style) (int, int) {
	return s.GetHorizontalMargins() + s.GetHorizontalPadding() + s.GetHorizontalBorderSize(),
		s.GetVerticalMargins() + s.GetVerticalPadding() + s.GetVerticalBorderSize()
}

func (p *ShortCutPanel) HandleSizeMsg(msg ResizeMsg) *UpdateResponse {
	DebugPrintf("ShortCutPanel received size message: %+v\n", msg)
	p.redraw = true
	if p.view != nil {
		p.view.Resize(msg.X, msg.Y, msg.Width, msg.Height)
	}

	_, tvert := GetStylingSize(p.TitleStyle)
	text_height := 1 + tvert
	horz, vert := GetStylingSize(p.PanelStyle.UnfocusedBorder)
	width := msg.Width - horz
	height := msg.Height - vert - text_height
	//p.TitleStyle = p.TitleStyle.Width(width)
	ur := p.Panel.HandleSizeMsg(ResizeMsg{EventResize: msg.EventResize, X: msg.X, Y: msg.Y, Width: width, Height: height})

	h, w := msg.Height-2, msg.Width-2
	p.PanelStyle.UnfocusedBorder = p.PanelStyle.UnfocusedBorder.Width(w)
	p.PanelStyle.FocusedBorder = p.PanelStyle.FocusedBorder.Width(w)
	p.PanelStyle.UnfocusedBorder = p.PanelStyle.UnfocusedBorder.Height(h)
	p.PanelStyle.FocusedBorder = p.PanelStyle.FocusedBorder.Height(h)
	return ur
}

func (p *ShortCutPanel) SetPath(path []int) {
	//DebugPrintf("ShortCutPanel.SetPath() called for %v\n", path)
	p.Panel.SetPath(path)
}
