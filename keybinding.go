package peanutbutter

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gdamore/tcell/v2"
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
	Override  bool
	Func      func() tea.Cmd
}

func NewKeyBinding(opts ...KeyBindingOption) *KeyBinding {
	keybinding := &KeyBinding{}
	for _, opt := range opts {
		opt(keybinding)
	}
	return keybinding
}

func SingleRuneBinding(rune rune) *KeyBinding {
	return NewKeyBinding(
		WithKeyDef(KeyDef{Key: tcell.KeyRune, Modifiers: tcell.ModMask(0), Rune: rune}),
		WithEnabled(true),
		WithShortHelp(""),
		WithLongHelp(""),
	)
}

func SingleKeyBinding(key tcell.Key) *KeyBinding {
	return NewKeyBinding(
		WithKeyDef(KeyDef{Key: key, Modifiers: tcell.ModMask(0), Rune: 0}),
		WithEnabled(true),
		WithShortHelp(""),
		WithLongHelp(""),
	)
}

var KeyTabBinding = *SingleKeyBinding(tcell.KeyTAB)

func ShiftTabBinding() *KeyBinding {
	kb := SingleKeyBinding(tcell.KeyTAB)
	kb.KeyDefs[0].Modifiers = tcell.ModMask(tcell.ModShift)
	return kb
}

var KeyShiftTabBinding = *SingleKeyBinding(tcell.KeyBacktab)

type KeyBindingOption func(*KeyBinding)

func WithFunc(fn func() tea.Cmd) KeyBindingOption {
	return func(keybinding *KeyBinding) {
		keybinding.Func = fn
	}
}

func WithEnabled(enabled bool) KeyBindingOption {
	return func(keybinding *KeyBinding) {
		keybinding.Enabled = enabled
	}
}

func WithKeyDefs(keyDefs []KeyDef) KeyBindingOption {
	return func(keybinding *KeyBinding) {
		keybinding.KeyDefs = keyDefs
	}
}

func WithShortHelp(shortHelp string) KeyBindingOption {
	return func(keybinding *KeyBinding) {
		keybinding.ShortHelp = shortHelp
	}
}

func WithLongHelp(longHelp string) KeyBindingOption {
	return func(keybinding *KeyBinding) {
		keybinding.LongHelp = longHelp
	}
}

func WithKeyDef(keyDef KeyDef) KeyBindingOption {
	return func(keybinding *KeyBinding) {
		keybinding.KeyDefs = append(keybinding.KeyDefs, keyDef)
	}
}

func (keybinding *KeyBinding) IsEnabled() bool {
	return keybinding.Enabled && len(keybinding.KeyDefs) > 0
}

func (keyDef *KeyDef) Matches(eventKey *tcell.EventKey) bool {
	if eventKey.Key() == tcell.KeyRune {
		return eventKey.Rune() == keyDef.Rune && eventKey.Modifiers() == keyDef.Modifiers
	}
	return eventKey.Key() == keyDef.Key && eventKey.Modifiers() == keyDef.Modifiers
}

func (keybinding *KeyBinding) IsMatch(eventKey *tcell.EventKey) bool {
	for _, keyDef := range keybinding.KeyDefs {
		if keyDef.Matches(eventKey) {
			return true
		}
	}
	return false
}

func KeyBindingsHandler(keyBindings []*KeyBinding, msg KeyMsg, onlyOverrides bool) tea.Cmd {
	DebugPrintf("KeyBindingsHandler received message from child: %T %+v\n", msg, msg)
	for _, keyBinding := range keyBindings {
		isValid := (onlyOverrides && keyBinding.Override) || !onlyOverrides
		if keyBinding.IsMatch(msg.EventKey) && isValid {
			if keyBinding.Enabled {
				if keyBinding.Func != nil {
					cmd := keyBinding.Func()
					msg.SetUsed()
					return cmd
				}
			}
		}
	}
	//msg.SetUnused()
	return nil
}

func (keyBinding *KeyBinding) SetFunc(fn func() tea.Cmd) *KeyBinding {
	keyBinding.Func = fn
	return keyBinding
}

func (keyBinding *KeyBinding) SetShortHelp(shortHelp string) *KeyBinding {
	keyBinding.ShortHelp = shortHelp
	return keyBinding
}

func (keyBinding *KeyBinding) SetLongHelp(longHelp string) *KeyBinding {
	keyBinding.LongHelp = longHelp
	return keyBinding
}

func (ev KeyDef) String() string {
	s := ""
	m := []string{}
	if ev.Modifiers&tcell.ModShift != 0 {
		m = append(m, "Shift")
	}
	if ev.Modifiers&tcell.ModAlt != 0 {
		m = append(m, "Alt")
	}
	if ev.Modifiers&tcell.ModMeta != 0 {
		m = append(m, "Meta")
	}
	if ev.Modifiers&tcell.ModCtrl != 0 {
		m = append(m, "Ctrl")
	}

	ok := false
	if s, ok = tcell.KeyNames[ev.Key]; !ok {
		if ev.Key == tcell.KeyRune {
			s = string(ev.Rune)
		} else {
			s = fmt.Sprintf("Key[%d,%d]", ev.Key, int(ev.Rune))
		}
	}
	if len(m) != 0 {
		if ev.Modifiers&tcell.ModCtrl != 0 && strings.HasPrefix(s, "Ctrl-") {
			s = s[5:]
		}
		return fmt.Sprintf("%s+%s", strings.Join(m, "+"), s)
	}
	return s
}

func renderKeyDefs(keyDefs []KeyDef) string {
	keys := []string{}
	for _, keyDef := range keyDefs {
		keys = append(keys, keyDef.String())
	}
	return strings.Join(keys, "/")
}

func ShortHelpTexts(keybindings []*KeyBinding) []string {
	helpTexts := []string{}
	for _, keybinding := range keybindings {
		if keybinding.ShortHelp != "" && keybinding.Enabled {
			helptext := fmt.Sprintf("%s: %s", renderKeyDefs(keybinding.KeyDefs), keybinding.ShortHelp)
			helpTexts = append(helpTexts, helptext)
		}
	}
	return helpTexts
}
