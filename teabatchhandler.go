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

func BatchCmdHandler(msg tea.BatchMsg) []Msg {
	msgs := []Msg{}
	for _, m := range msg {
		if m != nil {
			msg := m()
			if msg, ok := msg.(tea.BatchMsg); ok {
				msgs = append(msgs, BatchCmdHandler(msg)...)
			}
		}
	}
	return msgs
}

// For any tea.Model that uses auto-routed messages,
// the top-level model should use this function to handle
// auto-routed batch messages.
// Auto-routed batch messages need to be converted to auto-routed cmds
// and then added to the batch to be executed.
func HandleAutoRoutedBatchMsg(msg tea.BatchMsg, path []int) []tea.Cmd {
	DebugPrintf("HandleAutoRoutedBatchMsg: %+v\n", msg)
	cmds := []tea.Cmd{}
	for _, m := range msg {
		if m != nil {
			cmds = append(cmds, MakeAutoRoutedCmd(m, path))
		}
	}
	return cmds
}

// When a bunch of tea-bubbles are put together in multiple panels,
// this function embeds the cmd in an auto-routed cmd, so that
// when the cmd is executed, the message is auto-routed to the
// generator of the cmd
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
