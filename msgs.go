package peanutbutter

import (
	"fmt"

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
	Up
	Down
	Left
	Right
	//StartWorkflow
	//NextWorkflow
	//PrevWorkflow
)

type Msg interface {
}

// Fundamental handling types

type RequestMsgType struct {
	Msg Msg
}

type FocusPropagatedMsgType struct {
	Msg Msg
}

type RoutedMsgType struct {
	Msg Msg
	RoutePath
}

type BroadcastMsgType struct {
	Msg Msg
}

type UntypedMsgType struct {
	Msg Msg
}

// End of fundamental handling types

type FocusRequestMsg struct {
	RequestedPath []int // Path to identify the focus request, e.g., [0, 2] means first panel's second child
	Relation      Relation
}

type ContextualHelpTextMsg struct {
	Text string
	Line int
}

type RoutePath struct {
	Path []int
}

func (routedPath *RoutePath) GetRoutePath() *RoutePath {
	return routedPath
}

type FocusGrantMsg struct {
	RoutePath
	Relation Relation
}

type FocusRevokeMsg struct{}

type PropagationDirection int

const (
	DownwardPropagation PropagationDirection = iota
	UpwardPropagation
	AnyPropagation
)

type KeyMsg struct {
	*tcell.EventKey
	Unused    *bool
	Direction *PropagationDirection
}

func (keyMsg KeyMsg) String() string {
	kdf := KeyDef{Key: keyMsg.EventKey.Key(), Modifiers: keyMsg.EventKey.Modifiers(), Rune: keyMsg.EventKey.Rune()}
	return fmt.Sprintf("KeyMsg: %s, Unused: %v, Direction: %v\n", kdf.String(), *keyMsg.Unused, *keyMsg.Direction)
}

// IsUsed returns true if the keyMsg has not been used
// We make the assumption that before a keyMsg is given to a
// IModel it is used up, making it unavailable for subsequent key
// bindings to match.
// However, it is up to the IModel to set the keyMsg to be unused
// allowing it to have control over the keyMsg's lifecycle.
func (keyMsg *KeyMsg) IsUsed() bool {
	return !*keyMsg.Unused
}

// SetUnused sets the keyMsg to unused
func (keyMsg *KeyMsg) SetUnused() {
	*keyMsg.Unused = true
}

// SetUsed sets the keyMsg to used
func (keyMsg *KeyMsg) SetUsed() {
	*keyMsg.Unused = false
}

func (keyMsg *KeyMsg) SetDirection(direction PropagationDirection) {
	keyMsg.Direction = &direction
}

func (keyMsg *KeyMsg) Matches(keyDef KeyDef) bool {
	return keyDef.Matches(keyMsg.EventKey)
}

type ConsiderForLocalShortcutMsg struct {
	KeyMsg
	*RoutePath
}

type ConsiderForGlobalShortcutMsg struct {
	KeyMsg
}

type AutoRoutedMsg struct {
	Msg
	RoutePath
}

type BroadcastMsg struct {
	Msg
}

type SelectedTabIndexMsg struct {
	Index         int
	ListPanelName string
}

type SelectTabIndexMsg struct {
	Index         int
	ListPanelName string
}

type IMessageWithRoutePath interface {
	GetRoutePath() RoutePath
}

/*
func GetHandlingForMessageWithRoutePath(msg IMessageWithRoutePath) func(msg Msg) Msg {
	routePath := msg.GetRoutePath()

	if routePath == nil || len(routePath.Path) == 0 {
		return func(msg Msg) Msg {
			return BroadcastMsgType{Msg: msg}
		}
	}

	return func(msg Msg) Msg {
		return RoutedMsgType{Msg: msg, RoutePath: routePath}
	}

}
*/

func GetMessageHandlingType(msg Msg) Msg {
	switch msg := msg.(type) {
	case FocusGrantMsg:
		return RoutedMsgType{Msg: msg, RoutePath: msg.RoutePath}
	case FocusRevokeMsg:
		return BroadcastMsgType{Msg: msg}
	case FocusRequestMsg:
		return RequestMsgType{Msg: msg}
	case ContextualHelpTextMsg:
		return RequestMsgType{Msg: msg}
	case AutoRoutedMsg:
		return RoutedMsgType{Msg: msg, RoutePath: msg.RoutePath}
	case KeyMsg:
		return FocusPropagatedMsgType{Msg: msg}
	case ResizeMsg:
		return msg
	default:
		return UntypedMsgType{Msg: msg}
	}
}
