package peanutbutter

import (
	tea "github.com/charmbracelet/bubbletea"
	tcellviews "github.com/gdamore/tcell/v2/views"
)

const (
	titleOffset = 2
)

type ShortCutPanelConfig struct {
	ContextualHelp string
	Title          string
	TitleStyle     TitleStyle
	PanelStyle     PanelStyle
	KeyBindings    []*KeyBinding
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
	Model              ILeafModel
	focus              bool
	path               []int
	Name               string
	view               *tcellviews.ViewPort
	MarkMessageNotUsed func(msg *KeyMsg)
	modelView          *tcellviews.ViewPort
	tabHidden          bool
}

type ShortCutPanelOption func(*ShortCutPanel)

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

func WithShortCutPanelStyle(style PanelStyle) ShortCutPanelOption {
	return func(config *ShortCutPanel) {
		config.PanelStyle = style
	}
}

func WithShortCutPanelTitleStyle(style TitleStyle) ShortCutPanelOption {
	return func(config *ShortCutPanel) {
		config.TitleStyle = style
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

func WithKeyBindingMaker(keyBindingMaker func(*ShortCutPanel) *KeyBinding) ShortCutPanelOption {
	return func(panel *ShortCutPanel) {
		panel.KeyBindings = append(panel.KeyBindings, keyBindingMaker(panel))
	}
}

func NewShortCutPanel(model ILeafModel, opts ...ShortCutPanelOption) *ShortCutPanel {
	spanel := &ShortCutPanel{}
	spanel.ShortCutPanelConfig = DefaultPanelConfig
	for _, opt := range opts {
		opt(spanel)
	}
	if model != nil {
		spanel.Model = model
	}
	return spanel
}

var _ IPanel = &ShortCutPanel{}

func (p *ShortCutPanel) IsFocused() bool {
	return p.focus
}

func (p *ShortCutPanel) AddKeyBinding(kb *KeyBinding) {
	p.KeyBindings = append(p.KeyBindings, kb)
}

func (p *ShortCutPanel) SetPath(path []int) {
	p.path = make([]int, len(path))
	copy(p.path, path)
}

func (p *ShortCutPanel) RoutedCmd(cmd tea.Cmd) tea.Cmd {
	return MakeAutoRoutedCmd(cmd, p.path)
}

func (p *ShortCutPanel) GetPath() []int {
	return p.path
}

func (p *ShortCutPanel) GetName() string {
	return p.Name
}

func (p *ShortCutPanel) SetView(view *tcellviews.ViewPort) {
	p.view = view
	p.modelView = tcellviews.NewViewPort(p.view, 0, 0, -1, -1)
}

func (p *ShortCutPanel) drawModelWithDraw(m ILeafModelWithDraw, force bool) bool {
	childRedraw := m.Draw(force, p.modelView)
	if p.redraw || force || childRedraw {
		DebugPrintf("ShortCutPanel.Draw() called for %v. Redraw: %v, force: %v\n", p.GetPath(), p.redraw, force)
		p.redraw = false
		return true
	}
	return false
}

func (p *ShortCutPanel) drawModelWithView(m ILeafModelWithView, force bool) bool {
	if !m.NeedsRedraw() && !force {
		return false
	}
	DebugPrintf("ShortCutPanel.DrawModelWithView() called for %v. force: %v\n", p.GetPath(), force)
	str := m.View()
	TcellDrawHelper(str, p.modelView, []*tcellviews.ViewPort{})
	return true
}

func (p *ShortCutPanel) renderTitle() {
	renderTextOnBorder(
		p.TitleStyle.RenderTitle(p.Title, p.IsFocused()),
		renderOnTopEdge,
		offsetFromLeftSide,
		titleOffset,
		p.view,
	)
}

func (p *ShortCutPanel) renderBorder() {
	renderBorder(
		p.IsFocused(),
		p.PanelStyle,
		p.view,
	)
}

func (p *ShortCutPanel) Draw(force bool) bool {
	childRedraw := false
	modelOk := false
	if m, ok := p.Model.(ILeafModelWithDraw); ok {
		childRedraw = p.drawModelWithDraw(m, force)
		modelOk = true
	}
	if m, ok := p.Model.(ILeafModelWithView); ok {
		childRedraw = p.drawModelWithView(m, force)
		modelOk = true
	}
	if childRedraw || p.redraw || force {
		p.renderBorder()
		p.renderTitle()
		p.redraw = false
		return true
	}

	if !modelOk {
		DebugPrintf("ShortCutPanel.Draw() called for %v. Model is neither IModelWithDraw nor IModelWithView", p.GetPath())
	}
	return false
}

func (p *ShortCutPanel) Init(cmds chan tea.Cmd) {
	DebugPrintf("ShortCutPanel.Init() called for %v\n", p.GetPath())
	p.cmds = cmds
	var batchCmds []tea.Cmd
	cmd := p.Model.Init()
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

func (p *ShortCutPanel) FocusRequestCmd(direction Relation) tea.Cmd {
	return func() tea.Msg {
		return FocusRequestMsg{
			RequestedPath: p.GetPath(),
			Relation:      direction,
		}
	}
}

func (p *ShortCutPanel) HandleKeybindings(msg KeyMsg, onlyOverrides bool) tea.Cmd {
	DebugPrintf("ShortCutPanel received message from child: %T %+v\n", msg, msg)
	return KeyBindingsHandler(p.KeyBindings, msg, onlyOverrides)
}

func (p *ShortCutPanel) HandleMessage(msg Msg) {
	DebugPrintf("ShortCutPanel %v received message: %T %+v\n", p.GetPath(), msg, msg)
	var cmd tea.Cmd = nil
	switch msg := msg.(type) {
	case ResizeMsg:
		p.HandleSizeMsg(msg)
	case AutoRoutedMsg:
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
		cmds := []tea.Cmd{}
		if keyMsg, ok := msg.(KeyMsg); ok {
			cmds = append(cmds, p.HandleKeybindings(keyMsg, true))
			if keyMsg.IsUsed() {
				p.cmds <- tea.Batch(cmds...)
				return
			}
		}
		cmds = append(cmds, p.Model.Update(msg))
		if keyMsg, ok := msg.(KeyMsg); ok {
			if !keyMsg.IsUsed() {
				cmds = append(cmds, p.HandleKeybindings(keyMsg, false))
			}
		}
		p.cmds <- tea.Batch(cmds...)
	}
	if cmd != nil {
		p.cmds <- p.RoutedCmd(cmd)
	}
}

func (p *ShortCutPanel) HandleSizeMsg(msg ResizeMsg) tea.Cmd {
	DebugPrintf("ShortCutPanel received size message: %+v\n", msg)
	p.redraw = true
	SetSize(&p.PanelStyle, p.view, msg.X, msg.Y, msg.Width, msg.Height)

	start_x, start_y, horz, vert := GetStylingMargins(&p.PanelStyle)
	DebugPrintf("ShortCutPanel start_x start_y horz vert %v %v %v %v\n", start_x, start_y, horz, vert)
	width := msg.Width - horz
	height := msg.Height - vert
	p.modelView.Resize(start_x, start_y, width, height)
	cmd := p.Model.Update(ResizeMsg{EventResize: msg.EventResize, X: start_x, Y: start_y, Width: width, Height: height})
	return cmd
}

func (p *ShortCutPanel) IsInHiddenTab() bool {
	return p.tabHidden
}

func (p *ShortCutPanel) GetView() *tcellviews.ViewPort {
	return p.view
}

func (p *ShortCutPanel) SetTabHidden(hidden bool) {
	p.tabHidden = hidden
}
