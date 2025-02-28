package peanutbutter

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	tcellviews "github.com/gdamore/tcell/v2/views"
)

// ListPanel can house a list of panels
// that can be displayed in a horizontal, vertical, or stacked layout
// List panels also support handling focus propagation
type ListPanel struct {
	Panels       []IPanel
	path         []int // Path to uniquely identify this node in the hierarchy
	MsgForParent tea.Msg
	Layout       Layout
	Selected     int // Index of the selected panel, only used if the orientation is ZStacked
	Name         string
	view         *tcellviews.ViewPort
	redraw       bool
	cmds         chan tea.Cmd
	tabHidden    bool
	topLevel     bool
	panelStyle   PanelStyle
	titleStyle   TitleStyle
	iAmInFocus   bool
	KeyBindings  []*KeyBinding
}

var _ IPanel = &ListPanel{}

func (p *ListPanel) SetView(view *tcellviews.ViewPort) {
	p.view = view
	for _, panel := range p.Panels {
		newView := tcellviews.NewViewPort(p.view, 0, 0, -1, -1)
		panel.SetView(newView)
	}
}

func (p *ListPanel) renderBorder() {
	renderBorder(
		p.IAmInFocus(),
		p.panelStyle,
		p.view,
	)
}

func (p *ListPanel) IAmInFocus() bool {
	return p.iAmInFocus
}

func (p *ListPanel) GetView() *tcellviews.ViewPort {
	return p.view
}

func (p *ListPanel) AddKeyBinding(kb *KeyBinding) {
	p.KeyBindings = append(p.KeyBindings, kb)
}

func (p *ListPanel) HandleKeybindings(msg KeyMsg, onlyOverrides bool) tea.Cmd {
	DebugPrintf("ListPanel received message from child: %T %+v\n", msg, msg)
	return KeyBindingsHandler(p.KeyBindings, msg, onlyOverrides)
}

func (p *ListPanel) Draw(force bool) bool {
	redrawn := false
	DebugPrintf("ListPanel.Draw() called for %v. Redraw: %v, force: %v\n", p.path, p.redraw, force)
	continueForce := p.redraw || force
	if p.Layout.Orientation == ZStacked {
		redrawn = p.Panels[p.Selected].Draw(continueForce)
	} else {
		for _, panel := range p.Panels {
			panelDrawn := panel.Draw(continueForce)
			if panelDrawn {
				redrawn = true
			}
		}
	}

	if redrawn || continueForce {
		p.renderBorder()
		if p.Layout.Orientation == ZStacked {
			p.renderTabs()
		}
	}
	p.redraw = false
	return redrawn
}

func (p *ListPanel) renderTabs() {
	tabs := make([]string, len(p.Panels))
	for i, panel := range p.Panels {
		selected := i == p.Selected
		txt := panel.GetName()
		if selected {
			txt = "[" + txt + "]"
		}
		tabs[i] = p.titleStyle.RenderTitle(txt, selected)
	}
	renderTextOnBorder(
		lipgloss.JoinHorizontal(lipgloss.Top, tabs...),
		renderOnTopEdge,
		offsetFromLeftSide,
		2,
		p.view,
	)
}

func NewListPanel(models []IPanel, layout Layout, options ...ListPanelOption) *ListPanel {
	panels := make([]IPanel, len(models))

	for i, model := range models {
		panels[i] = model
	}

	listPanel := ListPanel{
		Panels:     panels,
		Layout:     layout,
		topLevel:   false,
		iAmInFocus: false,
	}

	if listPanel.Layout.Orientation == ZStacked {
		listPanel.Selected = 0
		listPanel.panelStyle = DefaultPanelConfig.PanelStyle
	}

	for _, option := range options {
		option(&listPanel)
	}

	return &listPanel
}

type ListPanelOption func(*ListPanel)

func WithTopLevel(topLevel bool) ListPanelOption {
	return func(m *ListPanel) {
		m.topLevel = topLevel
	}
}

func WithListPanelName(name string) ListPanelOption {
	return func(m *ListPanel) {
		m.Name = name
	}
}

func WithListPanelBorderStyle(panelStyle PanelStyle) ListPanelOption {
	return func(m *ListPanel) {
		m.panelStyle = panelStyle
	}
}

func WithListPanelTitleStyle(titleStyle TitleStyle) ListPanelOption {
	return func(m *ListPanel) {
		m.titleStyle = titleStyle
	}
}

func WithTabAdvanceKeyBindings(keyBindings KeyBinding) ListPanelOption {
	newKb := keyBindings
	return func(m *ListPanel) {
		newKb.Func = func() tea.Cmd {
			m.TabNext()
			return nil
		}
		m.AddKeyBinding(&newKb)
	}
}

func WithTabReverseKeyBindings(keyBindings KeyBinding) ListPanelOption {
	newKb := keyBindings
	return func(m *ListPanel) {
		newKb.Func = func() tea.Cmd {
			m.TabPrev()
			return nil
		}
		m.AddKeyBinding(&newKb)
	}
}

func (m *ListPanel) Init(cmds chan tea.Cmd) {
	m.cmds = cmds
	if m.topLevel {
		m.SetPath([]int{})
	}

	DebugPrintf("ListPanel.Init() called for %v\n", m.path)
	for _, panel := range m.Panels {
		panel.Init(cmds)
	}
	if !m.IsLayoutValid() {
		fmt.Printf("Invalid layout: %+v -- \n", m.path)
		m.AreDimensionsValid(true)
		panic("-- Invalid layout")
	}
}

func (m *ListPanel) SetPath(path []int) {
	m.path = make([]int, len(path))
	copy(m.path, path)
	for i, panel := range m.Panels {
		panel.SetPath(append(m.path, i))
	}
}

func (m *ListPanel) IsFocused() bool {
	// A ListPanel is focused if any of its children are focused
	for _, panel := range m.Panels {
		if panel.IsFocused() {
			return true
		}
	}
	return false
}

func (m *ListPanel) GetPath() []int {
	return m.path
}

func (m *ListPanel) HandleRequestMsg(msg RequestMsgType) {
	if msg, ok := msg.Msg.(FocusRequestMsg); ok {
		m.HandleMessage(FocusRevokeMsg{})
		m.cmds <- func() tea.Msg {
			return FocusGrantMsg{Relation: msg.Relation, RoutePath: &RoutePath{Path: msg.RequestedPath}}
		}
	}
}

func (m *ListPanel) HandleMyMessage(msg Msg) tea.Cmd {
	DebugPrintf("ListPanel %v received message to itself: %T %+v\n", m.path, msg, msg)
	if m.Layout.Orientation != ZStacked {
		return nil
	}
	switch msg := msg.(type) {
	case FocusRevokeMsg:
		m.iAmInFocus = false
		m.redraw = true
	case FocusGrantMsg:
		m.iAmInFocus = true
		m.redraw = true
	case KeyMsg:
		return m.HandleKeybindings(msg, false)
	}
	return nil
}

func (m *ListPanel) HandleMessage(msg Msg) {
	p := GetMessageHandlingType(msg)
	DebugPrintf("ListPanel %v received message: %T %+v %T\n", m.path, msg, msg, p)

	switch msg := p.(type) {
	case RequestMsgType:
		if m.topLevel {
			m.HandleRequestMsg(msg)
			return
		}

	case ResizeMsg:
		m.HandleSizeMsg(msg)

	case FocusPropagatedMsgType:
		if m.iAmInFocus {
			m.cmds <- m.HandleMyMessage(msg.Msg)
		} else {
			for _, panel := range m.Panels {
				if panel.IsFocused() {
					panel.HandleMessage(msg.Msg)
				}
			}
		}

	case RoutedMsgType:
		l_mypath := len(m.path)
		r_path := msg.GetRoutePath().Path
		l_msgpath := len(r_path)
		if l_mypath == l_msgpath {
			m.HandleMyMessage(msg.Msg)
			return
		} else {
			nextIdx := r_path[l_mypath]
			if nextIdx < 0 || nextIdx > len(m.Panels) {
				return
			}
			m.Panels[nextIdx].HandleMessage(msg.Msg)
		}

	case BroadcastMsgType:
		m.HandleMyMessage(msg.Msg)
		for _, panel := range m.Panels {
			//DebugPrintf("ListPanel %v broadcasting message to child %v\n", m.path, i)
			//DebugPrintf("panel: %T\n", panel)
			panel.HandleMessage(msg.Msg)
		}
	}
}

func (m *ListPanel) GetFocusIndex() int {
	for i, panel := range m.Panels {
		if panel.IsFocused() {
			return i
		}
	}
	return -1
}

func (m *ListPanel) handleFocusIndex(direction Relation) int {
	focusIndex := m.GetFocusIndex()
	len := len(m.Panels)
	if direction == Up {
		focusIndex--
	}
	if direction == Down {
		focusIndex++
	}
	if direction == Left {
		focusIndex--
	}
	if direction == Right {
		focusIndex++
	}

	return ((focusIndex % len) + len) % len
}

func (m *ListPanel) HandleFocusRequestMsg(msg FocusRequestMsg) *FocusGrantMsg {
	newFocusIndex := m.handleFocusIndex(msg.Relation)
	m.SetSelected(newFocusIndex)
	return &FocusGrantMsg{RoutePath: &RoutePath{Path: m.GetPath()}, Relation: msg.Relation}
}

func (m ListPanel) GetLayout() Layout {
	return m.Layout
}

func (m *ListPanel) HandleZStackedSizeMsg(x int, y int, width int, height int) {
	for _, panel := range m.Panels {
		newMsg := ResizeMsg{
			X:      x,
			Y:      y,
			Width:  width,
			Height: height,
		}
		panel.HandleMessage(newMsg)
	}
}

func (m *ListPanel) HandleSizeMsg(msg ResizeMsg) {
	DebugPrintf("ListPanel %v received size message: %+v\n", m.path, msg)

	width := msg.Width
	height := msg.Height
	if m.Layout.Width > 0 {
		width = m.Layout.Width
	}
	if m.Layout.Height > 0 {
		height = m.Layout.Height
	}

	DebugPrintf("ListPanel x y w h %v %v %v %v\n", msg.X, msg.Y, width, height)
	SetSize(&m.panelStyle, m.view, msg.X, msg.Y, width, height)

	start_x, start_y, horz, vert := GetStylingMargins(&m.panelStyle)
	DebugPrintf("ListPanel start_x start_y horz vert %v %v %v %v\n", start_x, start_y, horz, vert)

	switch m.Layout.Orientation {
	case ZStacked:
		m.HandleZStackedSizeMsg(start_x, start_y, width-horz, height-vert)
		return
	case Horizontal:
		m.HandleHorzSizeMsg(start_x, start_y, width-horz, height-vert)
		return
	case Vertical:
		m.HandleVertSizeMsg(start_x, start_y, width-horz, height-vert)
		return
	}
}

func (m *ListPanel) HandleHorzSizeMsg(X int, Y int, width int, height int) {
	widths := m.Layout.CalculateDims(width)
	for i, panel := range m.Panels {
		w := widths[i]
		h := height
		newMsg := ResizeMsg{
			X:      X,
			Y:      Y,
			Width:  w,
			Height: h,
		}
		X += w
		panel.HandleMessage(newMsg)
	}
}

func (m *ListPanel) HandleVertSizeMsg(X int, Y int, width int, height int) {
	heights := m.Layout.CalculateDims(height)
	for i, panel := range m.Panels {
		w := width
		h := heights[i]
		newMsg := ResizeMsg{
			X:      X,
			Y:      Y,
			Width:  w,
			Height: h,
		}
		Y += h
		panel.HandleMessage(newMsg)
	}
}

func (m *ListPanel) GetSelected() IPanel {
	return m.Panels[m.Selected]
}

func (m *ListPanel) GetSelectedIndex() int {
	return m.Selected
}

func (m *ListPanel) setSelectedModel(model IPanel) {
	m.Panels[m.Selected] = model
}

func (m *ListPanel) SetSelected(i int) tea.Cmd {
	DebugPrintf("ListPanel %v setting selected to %v\n", m.path, i)
	m.Selected = i
	m.redraw = true
	for i, panel := range m.Panels {
		if i == m.Selected {
			panel.SetTabHidden(false)
		} else {
			panel.SetTabHidden(true)
		}
	}
	return nil
}

func (m *ListPanel) GetName() string {
	return m.Name
}

func (m *ListPanel) TabNext() tea.Cmd {
	return m.SetSelected((m.Selected + 1) % len(m.Panels))
}

func (m *ListPanel) TabPrev() tea.Cmd {
	return m.SetSelected((m.Selected - 1 + len(m.Panels)) % len(m.Panels))
}

func (m *ListPanel) SetTabHidden(hidden bool) {
	m.tabHidden = hidden
	for _, panel := range m.Panels {
		panel.SetTabHidden(hidden)
	}
}

func (m *ListPanel) IsInHiddenTab() bool {
	return m.tabHidden
}
