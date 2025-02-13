package panelbubble

import (
	tea "github.com/charmbracelet/bubbletea"
)

type AutoRoutedCmd struct {
	tea.Cmd
	OriginPath []int
}

func (msg AutoRoutedCmd) AsRouteTypedMsg() Msg {
	return RequestMsgType{Msg: msg}
}

func MakeAutoRoutedCmd(cmd tea.Cmd, path []int) tea.Cmd {
	return func() tea.Msg {
		if cmd != nil {
			msg := cmd()
			switch msg := GetMessageHandlingType(msg).(type) {
			case UntypedMsgType:
				return AutoRoutedMsg{Msg: msg.Msg, RoutePath: &RoutePath{Path: path}}
			}
			return msg
		}
		return nil
	}
}
