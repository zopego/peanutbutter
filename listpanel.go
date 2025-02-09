package panelbubble

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
	Panels       []Focusable
	path         []int // Path to uniquely identify this node in the hierarchy
	MsgForParent tea.Msg
	Layout       Layout
	Selected     int // Index of the selected panel, only used if the orientation is Vertical
	Name         string
	view         *tcellviews.ViewPort
	redraw       bool
}

var _ tea.Model = &ListPanel{}
var _ Focusable = &ListPanel{}
var _ CanSendMsgToParent = &ListPanel{}

func (p *ListPanel) SetView(view *tcellviews.ViewPort) {
	p.view = view
	for i, panel := range p.Panels {
		newView := tcellviews.NewViewPort(p.view, 0, 0, -1, -1)
		panel.SetView(newView)
		p.Panels[i] = panel.(Focusable)
	}
}

func (p *ListPanel) Draw(force bool) bool {
	redrawn := false
	DebugPrintf("ListPanel.Draw() called for %v. Redraw: %v, force: %v\n", p.path, p.redraw, force)
	continueForce := p.redraw || force
	if p.Layout.Orientation == ZStacked {
		redrawn = p.Panels[p.Selected].Draw(continueForce)
	} else {
		for i, panel := range p.Panels {
			redrawn = redrawn || panel.Draw(continueForce)
			p.Panels[i] = panel.(Focusable)
		}
	}
	p.redraw = false
	return redrawn
}

func NewListPanel(models []Focusable, layout Layout) ListPanel {
	panels := make([]Focusable, len(models))

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

func (m *ListPanel) Init() tea.Cmd {
	DebugPrintf("ListPanel.Init() called for %v\n", m.path)
	var cmds []tea.Cmd
	for _, panel := range m.Panels {
		if panel, ok := panel.(Focusable); ok {
			cmd := panel.Init()
			cmds = append(cmds, cmd)
		}
	}
	if !m.IsLayoutValid() {
		fmt.Printf("Invalid layout: %+v -- \n", m.path)
		m.AreDimensionsValid(true)
		panic("-- Invalid layout")
	}
	return tea.Batch(cmds...)
}

func (m *ListPanel) SetPath(path []int) {
	m.path = path
	for i, panel := range m.Panels {
		if focusable, ok := panel.(Focusable); ok {
			focusable.SetPath(append(m.path, i))
		}
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

func (m *ListPanel) View() string {
	if m.Layout.Orientation == ZStacked {
		return m.Panels[m.Selected].View()
	}
	return m.ListView()
}

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

func (m *ListPanel) HandleZStackedMsg(msg tea.Msg) (tea.Cmd, bool) {
	if msg, ok := msg.(ConsiderForLocalShortcutMsg); ok {
		_, cmd := m.Panels[m.Selected].Update(msg)
		return cmd, true
	}
	if msg, ok := msg.(SelectTabIndexMsg); ok {
		if msg.ListPanelName == m.Name {
			m.SetSelected(msg.Index)
			m.redraw = true
			return nil, true
		}
	}
	return nil, false
}

func (m *ListPanel) UpdateInner(msg tea.Msg) tea.Cmd {
	if msg, ok := msg.(ResizeMsg); ok {
		return m.HandleSizeMsg(msg)
	}

	if m.Layout.Orientation == ZStacked {
		cmd, handled := m.HandleZStackedMsg(msg)
		if handled {
			return cmd
		}
		if handled {
			return nil
		}
	}
	DebugPrintf("ListPanel %v received message: %T %+v\n", m.path, msg, msg)
	p := GetMessageHandlingType(msg)
	switch msg := p.(type) {
	case FocusPropagatedMsgType:
		for i, panel := range m.Panels {
			if panel.IsFocused() {
				updatedModel, cmd := panel.Update(msg.Msg)
				m.Panels[i] = updatedModel.(Focusable)
				if cmd != nil {
					return cmd
				}
			}
		}
	case UntypedMsgType:
		// this really should not happen
		return nil

	case RoutedMsgType:
		l_mypath := len(m.path)
		l_msgpath := len(msg.GetRoutePath())
		if l_mypath == l_msgpath {
			// This message is destined for this listpanel
			return m.HandleRoutedMessage(msg.Msg)
		} else {
			nextIdx := msg.GetRoutePath()[l_mypath]
			if nextIdx < 0 || nextIdx > len(m.Panels) {
				return nil
			}
			_, cmd := m.Panels[nextIdx].Update(msg.Msg)
			if cmd != nil {
				return cmd
			}
		}
		return nil

	case BroadcastMsgType:
		cmds := []tea.Cmd{}
		for i, panel := range m.Panels {
			//DebugPrintf("ListPanel %v broadcasting message to child %v\n", m.path, i)
			//DebugPrintf("panel: %T\n", panel)
			updatedModel, cmd := panel.Update(msg.Msg)
			m.Panels[i] = updatedModel.(Focusable)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		if len(cmds) > 0 {
			return tea.Batch(cmds...)
		}
		return nil

	case RequestMsgType:
		return nil
	}

	return nil
}

func (m *ListPanel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := []tea.Cmd{}
	cmd := m.UpdateInner(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	for i, panel := range m.Panels {
		if panel, ok := panel.(CanSendMsgToParent); ok {
			msg := panel.GetMsgForParent()
			m.Panels[i] = panel.(Focusable)
			if msg != nil {
				return m, m.HandleMessageFromChild(msg)
			}
		}
	}
	return m, tea.Batch(cmds...)
}

func (m *ListPanel) GetFocusIndex() int {
	for i, panel := range m.Panels {
		if panel.IsFocused() {
			return i
		}
	}
	return -1
}

func handleFocusIndex(focusIndex int, direction Direction, len int) int {
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

func (m *ListPanel) HandleMessageFromChild(msg tea.Msg) tea.Cmd {
	DebugPrintf("ListPanel %v received message from child: %T %+v\n", m.path, msg, msg)
	if msg, ok := msg.(GeometricFocusRequestMsg); ok {
		focusIndex := m.GetFocusIndex()
		if focusIndex == -1 {
			return nil
		}
		if m.Layout.Orientation == Horizontal && (msg.Direction == Left || msg.Direction == Right) {
			// first, lets find the currently focused panel
			focusIndex = handleFocusIndex(focusIndex, msg.Direction, len(m.Panels))
			path := append(m.path, focusIndex)
			return func() tea.Msg {
				return FocusRequestMsg{Relation: Self, RequestedPath: path}
			}
		}
		if m.Layout.Orientation == Vertical && (msg.Direction == Up || msg.Direction == Down) {
			focusIndex = handleFocusIndex(focusIndex, msg.Direction, len(m.Panels))
			path := append(m.path, focusIndex)
			return func() tea.Msg {
				return FocusRequestMsg{Relation: Self, RequestedPath: path}
			}
		}
	}
	m.SetMsgForParent(msg)
	return nil
}

func (m *ListPanel) HandleRoutedMessage(msg tea.Msg) tea.Cmd {
	DebugPrintf("ListPanel %v received routed message: %T %+v\n", m.path, msg, msg)
	if msg, ok := msg.(FocusGrantMsg); ok {
		// we'll add a path segment to the first panel in the list
		if len(m.Panels) > 0 {
			newmsg := msg
			newmsg.RoutePath.Path = append(newmsg.RoutePath.Path, 0)
			_, cmd := m.Panels[0].Update(newmsg)
			if cmd != nil {
				return cmd
			}
		}
	}
	return nil
}

func (m ListPanel) GetLayout() Layout {
	return m.Layout
}

func (m *ListPanel) HandleZStackedSizeMsg(msg ResizeMsg) tea.Cmd {
	cmds := []tea.Cmd{}
	for i, panel := range m.Panels {
		newMsg := ResizeMsg{
			X:      0,
			Y:      0,
			Width:  msg.Width,
			Height: msg.Height,
		}
		_, cmd := panel.Update(newMsg)
		m.Panels[i] = panel.(Focusable)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	if len(cmds) > 0 {
		return tea.Batch(cmds...)
	}
	return nil
}

func (m *ListPanel) HandleSizeMsg(msg ResizeMsg) tea.Cmd {
	DebugPrintf("ListPanel %v received size message: %+v\n", m.path, msg)
	if m.view != nil {
		m.view.Resize(msg.X, msg.Y, msg.Width, msg.Height)
	}
	cmds := []tea.Cmd{}
	if m.Layout.Orientation == ZStacked {
		return m.HandleZStackedSizeMsg(msg)
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
			updatedModel, cmd := panel.Update(newMsg)
			m.Panels[i] = updatedModel.(Focusable)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
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
			updatedModel, cmd := panel.Update(newMsg)
			m.Panels[i] = updatedModel.(Focusable)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	if len(cmds) > 0 {
		return tea.Batch(cmds...)
	}
	return nil
}

func (m *ListPanel) GetSelected() tea.Model {
	return m.Panels[m.Selected]
}

func (m *ListPanel) GetSelectedIndex() int {
	return m.Selected
}

func (m *ListPanel) setSelectedModel(model Focusable) {
	m.Panels[m.Selected] = model
}

func (m *ListPanel) SetSelected(i int) tea.Cmd {
	DebugPrintf("ListPanel %v setting selected to %v\n", m.path, i)
	m.Selected = i
	return func() tea.Msg {
		return SelectedTabIndexMsg{Index: i, ListPanelName: m.Name}
	}
}
