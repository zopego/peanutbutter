package panelbubble

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
}

var _ tea.Model = &ListPanel{}
var _ Focusable = &ListPanel{}

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

func (m ListPanel) Init() tea.Cmd {
	DebugPrintf("ListPanel.Init() called for %v\n", m.path)
	var cmds []tea.Cmd
	for _, panel := range m.Panels {
		if model, ok := panel.(Focusable); ok {
			cmd := model.Init()
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

func (m ListPanel) SetPath(path []int) Focusable {
	m.path = path
	for i, panel := range m.Panels {
		if focusable, ok := panel.(Focusable); ok {
			m.Panels[i] = focusable.SetPath(append(m.path, i))
		}
	}
	return m
}

func (m ListPanel) IsFocused() bool {
	// A ListPanel is focused if any of its children are focused
	for _, panel := range m.Panels {
		if panel.IsFocused() {
			return true
		}
	}
	return false
}

func (m ListPanel) GetPath() []int {
	return m.path
}

// Type enforces
type HasView interface {
	View() string
}

func (m ListPanel) View() string {
	if m.Layout.Orientation == ZStacked {
		return m.Panels[m.Selected].View()
	}
	return m.ListView()
}

func (m ListPanel) ListView() string {
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

func (m ListPanel) HandleZStackedMsg(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	if msg, ok := msg.(ConsiderForLocalShortcutMsg); ok {
		updatedPanel, cmd := m.Panels[m.Selected].Update(msg)
		m.Panels[m.Selected] = updatedPanel.(Focusable)
		return m, cmd, true
	}
	if msg, ok := msg.(SelectTabIndexMsg); ok {
		if msg.ListPanelName == m.Name {
			m, cmd := m.SetSelected(msg.Index)
			return m, cmd, true
		}
	}
	return m, nil, false
}

func (m ListPanel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(ResizeMsg); ok {
		return m.HandleSizeMsg(msg)
	}

	if m.Layout.Orientation == ZStacked {
		updatedModel, cmd, handled := m.HandleZStackedMsg(msg)
		m = updatedModel.(ListPanel)
		if cmd != nil {
			return m, cmd
		}
		if handled {
			return m, nil
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
					return m, cmd
				}
			}
		}
	case UntypedMsgType:
		// this really should not happen
		return m, nil

	case RoutedMsgType:
		l := len(m.path)

		nextIdx := msg.GetRoutePath()[l]
		if nextIdx < 0 || nextIdx > len(m.Panels) {
			return m, nil
		}
		if nextIdx == len(m.Panels) {
			// This message is destined for this panel
			updatedModel, cmd := m.HandleMessage(msg.Msg)
			m.Panels[nextIdx] = updatedModel.(Focusable)
			if cmd != nil {
				return m, cmd
			}
		} else {
			updatedModel, cmd := m.Panels[nextIdx].Update(msg.Msg)
			m.Panels[nextIdx] = updatedModel.(Focusable)
			if cmd != nil {
				return m, cmd
			}
		}
		return m, nil

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
			return m, tea.Batch(cmds...)
		}
		return m, nil

	case RequestMsgType:
		return m, nil
	}
	return m, nil
}

func (m ListPanel) HandleMessage(msg tea.Msg) (Focusable, tea.Cmd) {
	return m, nil
}

func (m ListPanel) GetLayout() Layout {
	return m.Layout
}

func (m ListPanel) HandleZStackedSizeMsg(msg ResizeMsg) (tea.Model, tea.Cmd) {
	cmds := []tea.Cmd{}
	for i, panel := range m.Panels {
		newMsg := ResizeMsg{
			Width:  msg.Width,
			Height: msg.Height,
		}
		updatedModel, cmd := panel.Update(newMsg)
		m.Panels[i] = updatedModel.(Focusable)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

/*
func (m ListPanel) makeDims(total int) []int {
	total_ratio := A.Reduce(func(acc float64, p Dimension) float64 {
		return acc + p.Ratio
	}, 0.0)(m.Layout.Dimensions)
	dims := A.Map(func(p Dimension) int {
		return int(float64(total) * p.Ratio)
	})(m.Layout.Dimensions)
	total_sum := A.Reduce(func(acc int, p int) int {
		return acc + p
	}, 0)(dims)
	total_by_ratio := int(float64(total) * total_ratio)
	if total_sum != total_by_ratio {
		diff := total_by_ratio - total_sum
		_ = diff
		//dims[len(dims)-1] += diff
	}
	return dims
}
*/

func (m ListPanel) HandleSizeMsg(msg ResizeMsg) (tea.Model, tea.Cmd) {
	DebugPrintf("ListPanel %v received size message: %+v\n", m.path, msg)
	cmds := []tea.Cmd{}
	if m.Layout.Orientation == ZStacked {
		return m.HandleZStackedSizeMsg(msg)
	}

	if m.Layout.Orientation == Horizontal {
		widths := m.Layout.CalculateDims(msg.Width)
		for i, panel := range m.Panels {
			newMsg := ResizeMsg{
				Width:  widths[i],
				Height: msg.Height,
			}
			updatedModel, cmd := panel.Update(newMsg)
			m.Panels[i] = updatedModel.(Focusable)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	} else {
		heights := m.Layout.CalculateDims(msg.Height)
		for i, panel := range m.Panels {
			newMsg := ResizeMsg{
				Width:  msg.Width,
				Height: heights[i],
			}
			updatedModel, cmd := panel.Update(newMsg)
			m.Panels[i] = updatedModel.(Focusable)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}
	return m, nil
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

func (m ListPanel) SetSelected(i int) (ListPanel, tea.Cmd) {
	DebugPrintf("ListPanel %v setting selected to %v\n", m.path, i)
	m.Selected = i
	return m, func() tea.Msg {
		return SelectedTabIndexMsg{Index: i, ListPanelName: m.Name}
	}
}
