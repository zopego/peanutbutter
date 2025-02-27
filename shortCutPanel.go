package peanutbutter

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	tcellviews "github.com/gdamore/tcell/v2/views"
)

type PanelStyle struct {
	FocusedBorder   lipgloss.Style
	UnfocusedBorder lipgloss.Style
}

type ShortCutPanelConfig struct {
	ContextualHelp string
	Title          string
	TitleStyle     lipgloss.Style
	PanelStyle     PanelStyle
	KeyBindings    []KeyBinding
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

func WithKeyBindingMaker(keyBindingMaker func(*ShortCutPanel) KeyBinding) ShortCutPanelOption {
	return func(panel *ShortCutPanel) {
		panel.KeyBindings = append(panel.KeyBindings, keyBindingMaker(panel))
	}
}

func NewShortCutPanel(model ILeafModel, opts ...ShortCutPanelOption) *ShortCutPanel {
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

func (p *ShortCutPanel) IsFocused() bool {
	return p.focus
}

func (p *ShortCutPanel) AddKeyBinding(kb KeyBinding) {
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

func (p *ShortCutPanel) DrawModelWithDraw(m ILeafModelWithDraw, force bool) bool {
	childRedraw := m.Draw(force, p.modelView)
	if p.redraw || force || childRedraw {
		DebugPrintf("ShortCutPanel.Draw() called for %v. Redraw: %v, force: %v\n", p.GetPath(), p.redraw, force)
		p.redraw = false
		return true
	}
	return false
}

func (p *ShortCutPanel) DrawModelWithView(m ILeafModelWithView, force bool) bool {
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
	if m, ok := p.Model.(ILeafModelWithDraw); ok {
		childRedraw = p.DrawModelWithDraw(m, force)
		modelOk = true
	}
	if m, ok := p.Model.(ILeafModelWithView); ok {
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

func (p *ShortCutPanel) borderView() string {
	borderStyle := p.PanelStyle.UnfocusedBorder
	if p.IsFocused() {
		borderStyle = p.PanelStyle.FocusedBorder
	}
	return borderStyle.Render(
		lipgloss.JoinVertical(lipgloss.Top, p.TitleStyle.Render(p.Title+" "+fmt.Sprintf("%+v", p.GetPath()))),
	)
}

func (p *ShortCutPanel) FocusRequestCmd(direction Relation) tea.Cmd {
	return func() tea.Msg {
		return FocusRequestMsg{
			RequestedPath: p.GetPath(),
			Relation:      direction,
		}
	}
}

func (p *ShortCutPanel) HandleMessageFromChild(msg Msg, onlyOverrides bool) tea.Cmd {
	DebugPrintf("ShortCutPanel received message from child: %T %+v\n", msg, msg)
	if msg, ok := msg.(KeyMsg); ok {
		for _, keyBinding := range p.KeyBindings {
			isValid := (onlyOverrides && keyBinding.Override) || !onlyOverrides
			if keyBinding.IsMatch(msg.EventKey) && isValid {
				if keyBinding.Enabled {
					if keyBinding.Func != nil {
						return keyBinding.Func()
					}
				}
			}
		}
		msg.SetUnused()
	}
	return nil
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
			cmds = append(cmds, p.HandleMessageFromChild(keyMsg, true))
			if keyMsg.IsUsed() {
				p.cmds <- tea.Batch(cmds...)
				return
			}
		}
		cmds = append(cmds, p.Model.Update(msg))
		if keyMsg, ok := msg.(KeyMsg); ok {
			if !keyMsg.IsUsed() {
				cmds = append(cmds, p.HandleMessageFromChild(keyMsg, false))
			}
		}
		p.cmds <- tea.Batch(cmds...)
	}
	if cmd != nil {
		p.cmds <- p.RoutedCmd(cmd)
	}
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

	text_height := 0
	if p.Title != "" {
		_, _, _, tvert := GetStylingSize(p.TitleStyle)
		text_height = 1 + tvert
	}

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

func (p *ShortCutPanel) GetLeafPanelCenters() []PanelCenter {
	x, y, X, Y := p.view.GetPhysical()
	return []PanelCenter{
		{X: x + (X / 2), Y: y + (Y / 2), Path: p.GetPath()},
	}
}

func (p *ShortCutPanel) IsInHiddenTab() bool {
	return p.tabHidden
}

func (p *ShortCutPanel) SetTabHidden(hidden bool) {
	p.tabHidden = hidden
}
