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
	redraw             bool
	cmds               chan tea.Cmd
	Model              IModel
	focus              bool
	path               []int
	Name               string
	view               *tcellviews.ViewPort
	MarkMessageNotUsed func(msg *KeyMsg)
	modelView          *tcellviews.ViewPort
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

func WithName(name string) ShortCutPanelOption {
	return func(panel *ShortCutPanel) {
		panel.Name = name
	}
}

func NewShortCutPanel(model IModel, opts ...ShortCutPanelOption) *ShortCutPanel {
	spanel := &ShortCutPanel{}
	for _, opt := range opts {
		opt(spanel)
	}
	if model != nil {
		spanel.Model = model
	}
	return spanel
}

var _ IPanel = &ShortCutPanel{}

//var _ CanSendMsgToParent = &ShortCutPanel{}

func (p *ShortCutPanel) IsFocused() bool {
	return p.focus
}

func (p *ShortCutPanel) SetPath(path []int) {
	p.path = path
}

func (p *ShortCutPanel) RoutedCmd(cmd tea.Cmd) tea.Cmd {
	return MakeAutoRoutedCmd(cmd, p.path)
}

func (p *ShortCutPanel) GetPath() []int {
	return p.path
}

func (p *ShortCutPanel) SetView(view *tcellviews.ViewPort) {
	p.view = view
	p.modelView = tcellviews.NewViewPort(p.view, 0, 0, -1, -1)
}

func (p *ShortCutPanel) DrawModelWithDraw(m IModelWithDraw, force bool) bool {
	childRedraw := m.Draw(force, p.modelView)
	if p.redraw || force || childRedraw {
		DebugPrintf("ShortCutPanel.Draw() called for %v. Redraw: %v, force: %v\n", p.GetPath(), p.redraw, force)
		p.redraw = false
		return true
	}
	return false
}

func (p *ShortCutPanel) DrawModelWithView(m IModelWithView, force bool) bool {
	if !m.NeedsRedraw() && !force {
		return false
	}
	DebugPrintf("ShortCutPanel.DrawModelWithView() called for %v. force: %v\n", p.GetPath(), force)
	str := m.View()
	TcellDrawHelper(str, p.modelView, []*tcellviews.ViewPort{})
	return true
}

func (p *ShortCutPanel) Draw(force bool) bool {
	childRedraw := false
	modelOk := false
	if m, ok := p.Model.(IModelWithDraw); ok {
		childRedraw = p.DrawModelWithDraw(m, force)
		modelOk = true
	}
	if m, ok := p.Model.(IModelWithView); ok {
		childRedraw = p.DrawModelWithView(m, force)
		modelOk = true
	}
	if childRedraw || p.redraw || force {
		str := p.borderView()
		TcellDrawHelper(str, p.view, []*tcellviews.ViewPort{p.modelView})
		p.redraw = false
		return true
	}

	if !modelOk {
		DebugPrintf("ShortCutPanel.Draw() called for %v. Model is neither IModelWithDraw nor IModelWithView", p.GetPath())
	}
	return false
}

func (p *ShortCutPanel) Init(cmds chan tea.Cmd, MarkMessageNotUsed func(msg *KeyMsg)) {
	DebugPrintf("ShortCutPanel.Init() called for %v\n", p.GetPath())
	p.cmds = cmds
	p.MarkMessageNotUsed = MarkMessageNotUsed
	var batchCmds []tea.Cmd
	cmd := p.Model.Init(func(msg *KeyMsg) {
		p.MarkMessageNotUsedInternal(*msg)
	})
	if cmd != nil {
		batchCmds = append(batchCmds, cmd)
	}
	if p.IsFocused() {
		batchCmds = append(batchCmds, func() tea.Msg {
			return ContextualHelpTextMsg{Text: p.ContextualHelp}
		})
	}
	cmds <- tea.Batch(batchCmds...)
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

func (p *ShortCutPanel) HandleMessageFromChild(msg Msg) tea.Cmd {
	DebugPrintf("ShortCutPanel received message from child: %T %+v\n", msg, msg)
	if msg, ok := msg.(*tcell.EventKey); ok {
		if p.EnableHorizontalMovement {
			switch msg.Key() {
			case tcell.KeyLeft, tcell.KeyRight:
				direction := Left
				if msg.Key() == tcell.KeyRight {
					direction = Right
				}
				return func() tea.Msg {
					return FocusRequestMsg{RequestedPath: p.GetPath(), Relation: direction}
				}
			}
		}
		if p.EnableVerticalMovement {
			switch msg.Key() {
			case tcell.KeyUp, tcell.KeyDown:
				direction := Up
				if msg.Key() == tcell.KeyDown {
					direction = Down
				}
				return func() tea.Msg {
					return FocusRequestMsg{RequestedPath: p.GetPath(), Relation: direction}
				}
			}
		}
		return func() tea.Msg {
			return ConsiderForLocalShortcutMsg{KeyMsg: KeyMsg{EventKey: msg}}
		}
	}
	return nil
}

func (p *ShortCutPanel) HandleMessage(msg Msg) {
	DebugPrintf("ShortCutPanel %v received message: %T %+v\n", p.GetPath(), msg, msg)
	var cmd tea.Cmd = nil
	switch msg := msg.(type) {
	case ConsiderForGlobalShortcutMsg, ConsiderForLocalShortcutMsg:
		p.HandleShortcuts(msg)
	case ResizeMsg:
		p.HandleSizeMsg(msg)
	case AutoRoutedMsg:
		cmd = p.Model.Update(msg.Msg)
	case BroadcastMsg:
		cmd = p.Model.Update(msg.Msg)
	case FocusGrantMsg:
		p.focus = true
		p.redraw = true
		cmd = p.Model.Update(msg)
	case FocusRevokeMsg:
		p.focus = false
		p.redraw = true
		cmd = p.Model.Update(msg)
	default:
		cmd = p.Model.Update(msg)
	}
	if cmd != nil {
		p.cmds <- p.RoutedCmd(cmd)
	}
}

func (p *ShortCutPanel) MarkMessageNotUsedInternal(msg KeyMsg) {
	cmd := p.HandleMessageFromChild(msg)
	if cmd != nil {
		p.cmds <- cmd
	} else {
		p.MarkMessageNotUsed(&msg)
	}
}

func (p *ShortCutPanel) HandleShortcuts(msg tea.Msg) tea.Cmd {
	DebugPrintf("ShortCutPanel received shortcut message: %T %+v\n", msg, msg)
	if msg, ok := msg.(ConsiderForGlobalShortcutMsg); ok {
		if p.GlobalShortcut.IsMatch(msg.EventKey) {
			return func() tea.Msg {
				return FocusRequestMsg{RequestedPath: p.GetPath(), Relation: Self}
			}
		}
	}
	if msg, ok := msg.(ConsiderForLocalShortcutMsg); ok {
		if p.LocalShortcut.IsMatch(msg.EventKey) {
			return func() tea.Msg {
				return FocusRequestMsg{RequestedPath: p.GetPath(), Relation: Self}
			}
		}
	}
	return nil
}

func GetStylingSize(s lipgloss.Style) (int, int, int, int) {

	horz_left := s.GetBorderLeftSize() + s.GetMarginLeft() + s.GetPaddingLeft()
	vert_top := s.GetBorderTopSize() + s.GetMarginTop() + s.GetPaddingTop()

	horz_total := s.GetHorizontalMargins() + s.GetHorizontalPadding() + s.GetHorizontalBorderSize()
	vert_total := s.GetVerticalMargins() + s.GetVerticalPadding() + s.GetVerticalBorderSize()
	return horz_left, vert_top, horz_total, vert_total
}

func (p *ShortCutPanel) HandleSizeMsg(msg ResizeMsg) tea.Cmd {
	DebugPrintf("ShortCutPanel received size message: %+v\n", msg)
	p.redraw = true
	if p.view != nil {
		p.view.Resize(msg.X, msg.Y, msg.Width, msg.Height)
	}

	_, _, _, tvert := GetStylingSize(p.TitleStyle)
	text_height := 1 + tvert
	start_x, start_y, horz, vert := GetStylingSize(p.PanelStyle.UnfocusedBorder)
	width := msg.Width - horz
	height := msg.Height - vert - text_height
	//p.TitleStyle = p.TitleStyle.Width(width)
	p.modelView.Resize(start_x, start_y+text_height, width, height)
	cmd := p.Model.Update(ResizeMsg{EventResize: msg.EventResize, X: msg.X, Y: msg.Y, Width: width, Height: height})

	h, w := msg.Height-2, msg.Width-2
	p.PanelStyle.UnfocusedBorder = p.PanelStyle.UnfocusedBorder.Width(w)
	p.PanelStyle.FocusedBorder = p.PanelStyle.FocusedBorder.Width(w)
	p.PanelStyle.UnfocusedBorder = p.PanelStyle.UnfocusedBorder.Height(h)
	p.PanelStyle.FocusedBorder = p.PanelStyle.FocusedBorder.Height(h)
	return cmd
}
