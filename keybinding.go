package peanutbutter

import (
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
	Func      func() tea.Cmd
}

func NewKeyBinding(opts ...KeyBindingOption) *KeyBinding {
	keybinding := &KeyBinding{}
	for _, opt := range opts {
		opt(keybinding)
	}
	return keybinding
}

func SingleRuneBinding(rune rune) KeyBinding {
	return *NewKeyBinding(
		WithKeyDef(KeyDef{Key: tcell.KeyRune, Modifiers: tcell.ModMask(0), Rune: rune}),
		WithEnabled(true),
		WithShortHelp(""),
		WithLongHelp(""),
	)
}

func SingleKeyBinding(key tcell.Key) KeyBinding {
	return *NewKeyBinding(
		WithKeyDef(KeyDef{Key: key, Modifiers: tcell.ModMask(0), Rune: 0}),
		WithEnabled(true),
		WithShortHelp(""),
		WithLongHelp(""),
	)
}

var KeyTabBinding = SingleKeyBinding(tcell.KeyTAB)

func ShiftTabBinding() KeyBinding {
	kb := SingleKeyBinding(tcell.KeyTAB)
	kb.KeyDefs[0].Modifiers = tcell.ModMask(tcell.ModShift)
	return kb
}

var KeyShiftTabBinding = SingleKeyBinding(tcell.KeyBacktab)

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
