package panelbubble

import (
	tea "github.com/charmbracelet/bubbletea"
	tcell "github.com/gdamore/tcell/v2"
)

// There are 4 types of messages:
// 1. Request messages: messages that are sent to the top level panel
// These are messages like FocusRequestMsg, FocusRevokeMsg, ContextualHelpTextMsg,
// that are responded to by the top level panel
// 2. Focus Propagated messages: messages that are sent along the current focus path
// 3. Routed messages: messages that are sent to a specific panel identified by path
// 4. Broadcast messages: messages that are sent to all panels
// Relation is an enum that describes the relation of a panel to its requested focus panel
type Relation int

const (
	Self Relation = iota
	Parent
	//	StartWorkflow
	NextWorkflow
	PrevWorkflow
)

type KeyDef struct {
	Key       tcell.Key
	Modifiers tcell.ModMask
	Rune      rune
}

type KeyBinding struct {
	KeyDefs   []KeyDef
	ShortHelp string
	LongHelp  string
	Enabled   bool
}

func NewKeyBinding(keyDefs []KeyDef, enabled bool, shortHelp string, longHelp string) *KeyBinding {
	return &KeyBinding{KeyDefs: keyDefs, ShortHelp: shortHelp, LongHelp: longHelp, Enabled: enabled}
}

func (keybinding *KeyBinding) IsEnabled() bool {
	return keybinding.Enabled && len(keybinding.KeyDefs) > 0
}

func (keybinding *KeyBinding) IsMatch(eventKey *tcell.EventKey) bool {
	for _, keyDef := range keybinding.KeyDefs {
		if eventKey.Key() == keyDef.Key &&
			eventKey.Modifiers() == keyDef.Modifiers &&
			eventKey.Rune() == keyDef.Rune {
			return true
		}
	}
	return false
}

type RequestMsgType struct {
	Msg tea.Msg
}

type FocusPropagatedMsgType struct {
	Msg tea.Msg
}

type RoutedMsgType struct {
	Msg tea.Msg
	*RoutePath
}

func (routedMsg RoutedMsgType) GetRoutePath() []int {
	return routedMsg.RoutePath.Path
}

type BroadcastMsgType struct {
	Msg tea.Msg
}

type UntypedMsgType struct {
	Msg tea.Msg
}

type FocusRequestMsg struct {
	RequestedPath []int // Path to identify the focus request, e.g., [0, 2] means first panel's second child
	Relation      Relation
	WorkflowName  string
}

type Direction int

const (
	Up Direction = iota
	Down
	Left
	Right
)

type GeometricFocusRequestMsg struct {
	Direction Direction
}

func (msg FocusRequestMsg) AsRouteTypedMsg() tea.Msg {
	return RequestMsgType{Msg: msg}
}

type ContextualHelpTextMsg struct {
	Text string
	Line int
}

type RoutePath struct {
	Path []int
}

func (routedPath RoutePath) GetPath() []int {
	return routedPath.Path
}

type FocusGrantMsg struct {
	RoutePath
	Relation     Relation
	WorkflowName string
}

func (msg FocusGrantMsg) AsRouteTypedMsg() tea.Msg {
	if msg.RoutePath.Path == nil || len(msg.RoutePath.Path) == 0 {
		return BroadcastMsgType{Msg: msg}
	}
	if msg.Relation == Self {
		return RoutedMsgType{Msg: msg, RoutePath: &msg.RoutePath}
	}
	return BroadcastMsgType{Msg: msg}
}

type FocusRevokeMsg struct{}

func (msg FocusRevokeMsg) AsRouteTypedMsg() tea.Msg {
	return BroadcastMsgType{Msg: msg}
}

type MsgUsed struct{}

func (msg MsgUsed) AsRouteTypedMsg() tea.Msg {
	return RequestMsgType{Msg: msg}
}

func MsgUsedCmd() tea.Cmd {
	return func() tea.Msg {
		return MsgUsed{}
	}
}

type ConsiderForLocalShortcutMsg struct {
	EventKey *tcell.EventKey
	RoutePath
}

func (msg ConsiderForLocalShortcutMsg) AsRouteTypedMsg() tea.Msg {
	if msg.RoutePath.Path == nil || len(msg.RoutePath.Path) == 0 {
		return BroadcastMsgType{Msg: msg}
	}
	return RoutedMsgType{Msg: msg, RoutePath: &msg.RoutePath}
}

type ConsiderForGlobalShortcutMsg struct {
	EventKey *tcell.EventKey
}

func (msg ConsiderForGlobalShortcutMsg) AsRouteTypedMsg() tea.Msg {
	return BroadcastMsgType{Msg: msg}
}

type AutoRoutedMsg struct {
	tea.Msg
	RoutePath
}

func (msg AutoRoutedMsg) AsRouteTypedMsg() tea.Msg {
	return RoutedMsgType{Msg: msg, RoutePath: &msg.RoutePath}
}

type AutoRoutedCmd struct {
	tea.Cmd
	OriginPath []int
}

func (msg AutoRoutedCmd) AsRouteTypedMsg() tea.Msg {
	return RequestMsgType{Msg: msg}
}

type BroadcastMsg struct {
	tea.Msg
}

func (msg BroadcastMsg) AsRouteTypedMsg() tea.Msg {
	return BroadcastMsgType{Msg: msg}
}

type PropagateKeyMsg struct {
	EventKey *tcell.EventKey
}

type SelectedTabIndexMsg struct {
	Index         int
	ListPanelName string
}

type SelectTabIndexMsg struct {
	Index         int
	ListPanelName string
}

func (msg PropagateKeyMsg) AsRouteTypedMsg() tea.Msg {
	return FocusPropagatedMsgType{Msg: msg}
}

type ImplementsAsRouteTypedMsg interface {
	AsRouteTypedMsg() tea.Msg
}

func GetMessageHandlingType(msg tea.Msg) tea.Msg {
	if msg, ok := msg.(ImplementsAsRouteTypedMsg); ok {
		return msg.AsRouteTypedMsg()
	}
	if msg, ok := msg.(tea.KeyMsg); ok {
		return FocusPropagatedMsgType{Msg: msg}
	}
	if msg, ok := msg.(*tcell.EventKey); ok {
		return FocusPropagatedMsgType{Msg: msg}
	}
	return UntypedMsgType{Msg: msg}
}

func BatchCmdHandler(msg tea.BatchMsg) []tea.Msg {
	msgs := []tea.Msg{}
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
func HandleAutoRoutedBatchMsg(msg tea.BatchMsg, path []int) tea.Cmd {
	DebugPrintf("HandleAutoRoutedBatchMsg: %+v\n", msg)
	cmds := []tea.Cmd{}
	for _, m := range msg {
		if m != nil {
			cmds = append(cmds, MakeAutoRoutedCmd(m, path))
		}
	}
	return tea.Batch(cmds...)
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
				return AutoRoutedMsg{Msg: msg.Msg, RoutePath: RoutePath{Path: path}}
			}
			return msg
		}
		return nil
	}
}
