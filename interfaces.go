package peanutbutter

import (
	tea "github.com/charmbracelet/bubbletea"
	tcellviews "github.com/gdamore/tcell/v2/views"
)

func IsSamePath(path1, path2 []int) bool {
	if len(path1) != len(path2) {
		return false
	}
	for i, v := range path1 {
		if v != path2[i] {
			return false
		}
	}
	return true
}

type IRootModel interface {
	Update(msg Msg)
	Init(cmds chan tea.Cmd, view *tcellviews.ViewPort) tea.Cmd
	Draw() bool
}

type ILeafModel interface {
	Update(msg Msg) tea.Cmd
	Init() tea.Cmd
	//SetView(view *tcellviews.ViewPort)
	//Draw(force bool) bool
}

type ILeafModelWithView interface {
	NeedsRedraw() bool
	View() string
}

type ILeafModelWithDraw interface {
	Draw(force bool, view *tcellviews.ViewPort) bool
}

type PanelCenter struct {
	X    int
	Y    int
	Path []int
}

// IPanel is an interface that allows a tea model to be focused.
// It is used to handle focus passing in the UI.
type IPanel interface {
	IsFocused() bool
	GetPath() []int
	SetPath(path []int)
	HandleMessage(msg Msg)
	SetView(view *tcellviews.ViewPort)
	Draw(force bool) bool
	Init(cmds chan tea.Cmd)
	GetName() string
	SetTabHidden(hidden bool)
	IsInHiddenTab() bool
	AddKeyBinding(kb *KeyBinding)
}
