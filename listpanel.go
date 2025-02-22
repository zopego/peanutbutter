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
	Panels             []IPanel
	path               []int // Path to uniquely identify this node in the hierarchy
	MsgForParent       tea.Msg
	Layout             Layout
	Selected           int // Index of the selected panel, only used if the orientation is Vertical
	Name               string
	view               *tcellviews.ViewPort
	redraw             bool
	cmds               chan tea.Cmd
	MarkMessageNotUsed func(msg *KeyMsg)
}

var _ IPanel = &ListPanel{}

func (p *ListPanel) SetView(view *tcellviews.ViewPort) {
	p.view = view
	for _, panel := range p.Panels {
		newView := tcellviews.NewViewPort(p.view, 0, 0, -1, -1)
		panel.SetView(newView)
	}
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
	p.redraw = false
	return redrawn
}

func NewListPanel(models []IPanel, layout Layout) ListPanel {
	panels := make([]IPanel, len(models))

	for i, model := range models {
		panels[i] = model
	}

	return ListPanel{
		Panels: panels,
		Layout: layout,
	}
}

func (m *ListPanel) GetMsgForParent() tea.Msg {
	msg := m.MsgForParent
	m.MsgForParent = nil
	return msg
}

func (m *ListPanel) SetMsgForParent(msg tea.Msg) {
	m.MsgForParent = msg
}

func (m *ListPanel) Init(cmds chan tea.Cmd, MarkMessageNotUsed func(msg *KeyMsg)) {
	m.cmds = cmds
	m.MarkMessageNotUsed = MarkMessageNotUsed
	DebugPrintf("ListPanel.Init() called for %v\n", m.path)
	for _, panel := range m.Panels {
		panel.Init(cmds, MarkMessageNotUsed)
	}
	if !m.IsLayoutValid() {
		fmt.Printf("Invalid layout: %+v -- \n", m.path)
		m.AreDimensionsValid(true)
		panic("-- Invalid layout")
	}
}

func (m *ListPanel) SetPath(path []int) {
	m.path = path
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

// Type enforces
type HasView interface {
	View() string
}

/*
func (m *ListPanel) View() string {
	if m.Layout.Orientation == ZStacked {
		return m.Panels[m.Selected].View()
	}
	return m.ListView()
}*/

func (m *ListPanel) ListView() string {
	var views []string
	for _, panel := range m.Panels {
		if model, ok := panel.(HasView); ok {
			views = append(views, model.View())
		}
	}

	if m.Layout.Orientation == Horizontal {
		return lipgloss.JoinHorizontal(lipgloss.Top, views...)
	}
	return lipgloss.JoinVertical(lipgloss.Left, views...)
}

func (m *ListPanel) HandleZStackedMsg(msg tea.Msg) {
	if msg, ok := msg.(ConsiderForLocalShortcutMsg); ok {
		m.Panels[m.Selected].HandleMessage(msg)
	}
	if msg, ok := msg.(SelectTabIndexMsg); ok {
		if msg.ListPanelName == m.Name {
			m.SetSelected(msg.Index)
			m.redraw = true
		}
	}
}

func (m *ListPanel) HandleMessage(msg Msg) {
	p := GetMessageHandlingType(msg)
	DebugPrintf("ListPanel %v received message: %T %+v %T\n", m.path, msg, msg, p)

	if m.Layout.Orientation == ZStacked {
		m.HandleZStackedMsg(msg)
	}

	switch msg := p.(type) {
	case ResizeMsg:
		m.HandleSizeMsg(msg)

	case FocusPropagatedMsgType:
		for _, panel := range m.Panels {
			if panel.IsFocused() {
				panel.HandleMessage(msg.Msg)
			}
		}

	case RoutedMsgType:
		l_mypath := len(m.path)
		r_path := msg.GetRoutePath().Path
		l_msgpath := len(r_path)
		if l_mypath == l_msgpath {
			// This message is destined for this listpanel
			// return m.HandleRoutedMessage(msg.Msg)
		} else {
			nextIdx := r_path[l_mypath]
			if nextIdx < 0 || nextIdx > len(m.Panels) {
				return
			}
			m.Panels[nextIdx].HandleMessage(msg.Msg)
		}

	case BroadcastMsgType:
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

/*
func (m *ListPanel) HandleRoutedMessage(msg tea.Msg) {
	DebugPrintf("ListPanel %v received message from child: %T %+v\n", m.path, msg, msg)
	if msg, ok := msg.(FocusGrantMsg); ok {
		focusIndex := msg.RequesterPath
		direction := msg.Relation
		var path *[]int = nil
		if m.Layout.Orientation == Horizontal && (msg.Direction == Left || msg.Direction == Right) {
			// first, lets find the currently focused panel
			focusIndex = handleFocusIndex(focusIndex, msg.Direction, len(m.Panels))
			path = &[]int{focusIndex}
		}
		if m.Layout.Orientation == Vertical && (msg.Direction == Up || msg.Direction == Down) {
			focusIndex = handleFocusIndex(focusIndex, msg.Direction, len(m.Panels))
			path = &[]int{focusIndex}
		}
		if path != nil {
			return &UpdateResponse{
				Cmd: func() tea.Msg {
					return FocusRequestMsg{Relation: Self, RequestedPath: *path}
				},
				UpPropagateMsg: nil,
			}
		}
	}
	m.SetMsgForParent(msg)
	return nil
}
*/

/* func (m *ListPanel) HandleRoutedMessage(msg tea.Msg) *UpdateResponse {
	DebugPrintf("ListPanel %v received routed message: %T %+v\n", m.path, msg, msg)
	if msg, ok := msg.(FocusGrantMsg); ok {
		// we'll add a path segment to the first panel in the list
		if len(m.Panels) > 0 {
			newmsg := msg
			newmsg.RoutePath.Path = append(newmsg.RoutePath.Path, 0)
			return m.Panels[0].Update(newmsg)
		}
	}
	return nil
}*/

func (m ListPanel) GetLayout() Layout {
	return m.Layout
}

func (m *ListPanel) GetLeafPanelCenters() []PanelCenter {
	centers := []PanelCenter{}
	for _, panel := range m.Panels {
		centers = append(centers, panel.GetLeafPanelCenters()...)
	}
	return centers
}

func (m *ListPanel) HandleZStackedSizeMsg(msg ResizeMsg) {
	for _, panel := range m.Panels {
		newMsg := ResizeMsg{
			X:      0,
			Y:      0,
			Width:  msg.Width,
			Height: msg.Height,
		}
		panel.HandleMessage(newMsg)
	}
}

func (m *ListPanel) HandleSizeMsg(msg ResizeMsg) {
	DebugPrintf("ListPanel %v received size message: %+v\n", m.path, msg)
	if m.view != nil {
		m.view.Resize(msg.X, msg.Y, msg.Width, msg.Height)
	}

	if m.Layout.Orientation == ZStacked {
		m.HandleZStackedSizeMsg(msg)
		return
	}

	if m.Layout.Orientation == Horizontal {
		widths := m.Layout.CalculateDims(msg.Width)
		X := 0
		for i, panel := range m.Panels {
			w := widths[i]
			h := msg.Height
			newMsg := ResizeMsg{
				X:      X,
				Y:      0,
				Width:  w,
				Height: h,
			}
			X += w
			panel.HandleMessage(newMsg)
		}
	} else {
		heights := m.Layout.CalculateDims(msg.Height)
		Y := 0
		for i, panel := range m.Panels {
			w := msg.Width
			h := heights[i]
			newMsg := ResizeMsg{
				X:      0,
				Y:      Y,
				Width:  w,
				Height: h,
			}
			Y += h
			panel.HandleMessage(newMsg)
		}
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
	return func() tea.Msg {
		return SelectedTabIndexMsg{Index: i, ListPanelName: m.Name}
	}
}
