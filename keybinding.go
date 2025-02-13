package panelbubble

import (
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
