package peanutbutter

import (
	tea "github.com/charmbracelet/bubbletea"
	tcell "github.com/gdamore/tcell/v2"
)

var tCellKeyToTeaMsg = map[tcell.Key]tea.KeyMsg{
	tcell.KeyEnter:     tea.KeyMsg{Type: tea.KeyEnter},
	tcell.KeyBackspace: tea.KeyMsg{Type: tea.KeyBackspace},
	tcell.KeyTab:       tea.KeyMsg{Type: tea.KeyTab},
	//	tcell.KeyBacktab:    tea.KeyMsg{Type: tea.KeyBacktab},
	tcell.KeyEsc:        tea.KeyMsg{Type: tea.KeyEsc},
	tcell.KeyBackspace2: tea.KeyMsg{Type: tea.KeyBackspace},
	tcell.KeyDelete:     tea.KeyMsg{Type: tea.KeyDelete},
	tcell.KeyInsert:     tea.KeyMsg{Type: tea.KeyInsert},
	tcell.KeyUp:         tea.KeyMsg{Type: tea.KeyUp},
	tcell.KeyDown:       tea.KeyMsg{Type: tea.KeyDown},
	tcell.KeyLeft:       tea.KeyMsg{Type: tea.KeyLeft},
	tcell.KeyRight:      tea.KeyMsg{Type: tea.KeyRight},
	tcell.KeyHome:       tea.KeyMsg{Type: tea.KeyHome},
	tcell.KeyEnd:        tea.KeyMsg{Type: tea.KeyEnd},
	tcell.KeyUpLeft:     tea.KeyMsg{Type: tea.KeyUp},
	tcell.KeyUpRight:    tea.KeyMsg{Type: tea.KeyUp},
	tcell.KeyDownLeft:   tea.KeyMsg{Type: tea.KeyDown},
	tcell.KeyDownRight:  tea.KeyMsg{Type: tea.KeyDown},
	//	tcell.KeyCenter:         tea.KeyMsg{Type: tea.KeyCenter},
	tcell.KeyPgDn: tea.KeyMsg{Type: tea.KeyPgDown},
	tcell.KeyPgUp: tea.KeyMsg{Type: tea.KeyPgUp},
	//	tcell.KeyClear:          tea.KeyMsg{Type: tea.KeyClear},
	//	tcell.KeyExit:           tea.KeyMsg{Type: tea.KeyExit},
	//	tcell.KeyCancel:         tea.KeyMsg{Type: tea.KeyCancel},
	//	tcell.KeyPause:          tea.KeyMsg{Type: tea.KeyPause},
	//	tcell.KeyPrint:          tea.KeyMsg{Type: tea.KeyPrint},
	tcell.KeyF1:    tea.KeyMsg{Type: tea.KeyF1},
	tcell.KeyF2:    tea.KeyMsg{Type: tea.KeyF2},
	tcell.KeyF3:    tea.KeyMsg{Type: tea.KeyF3},
	tcell.KeyF4:    tea.KeyMsg{Type: tea.KeyF4},
	tcell.KeyF5:    tea.KeyMsg{Type: tea.KeyF5},
	tcell.KeyF6:    tea.KeyMsg{Type: tea.KeyF6},
	tcell.KeyF7:    tea.KeyMsg{Type: tea.KeyF7},
	tcell.KeyF8:    tea.KeyMsg{Type: tea.KeyF8},
	tcell.KeyF9:    tea.KeyMsg{Type: tea.KeyF9},
	tcell.KeyF10:   tea.KeyMsg{Type: tea.KeyF10},
	tcell.KeyF11:   tea.KeyMsg{Type: tea.KeyF11},
	tcell.KeyF12:   tea.KeyMsg{Type: tea.KeyF12},
	tcell.KeyF13:   tea.KeyMsg{Type: tea.KeyF13},
	tcell.KeyF14:   tea.KeyMsg{Type: tea.KeyF14},
	tcell.KeyF15:   tea.KeyMsg{Type: tea.KeyF15},
	tcell.KeyF16:   tea.KeyMsg{Type: tea.KeyF16},
	tcell.KeyF17:   tea.KeyMsg{Type: tea.KeyF17},
	tcell.KeyF18:   tea.KeyMsg{Type: tea.KeyF18},
	tcell.KeyF19:   tea.KeyMsg{Type: tea.KeyF19},
	tcell.KeyF20:   tea.KeyMsg{Type: tea.KeyF20},
	tcell.KeyCtrlA: tea.KeyMsg{Type: tea.KeyCtrlA},
	tcell.KeyCtrlB: tea.KeyMsg{Type: tea.KeyCtrlB},
	tcell.KeyCtrlC: tea.KeyMsg{Type: tea.KeyCtrlC},
	tcell.KeyCtrlD: tea.KeyMsg{Type: tea.KeyCtrlD},
	tcell.KeyCtrlE: tea.KeyMsg{Type: tea.KeyCtrlE},
	tcell.KeyCtrlF: tea.KeyMsg{Type: tea.KeyCtrlF},
	tcell.KeyCtrlG: tea.KeyMsg{Type: tea.KeyCtrlG},
	tcell.KeyCtrlJ: tea.KeyMsg{Type: tea.KeyCtrlJ},
	tcell.KeyCtrlK: tea.KeyMsg{Type: tea.KeyCtrlK},
	tcell.KeyCtrlL: tea.KeyMsg{Type: tea.KeyCtrlL},
	tcell.KeyCtrlN: tea.KeyMsg{Type: tea.KeyCtrlN},
	tcell.KeyCtrlO: tea.KeyMsg{Type: tea.KeyCtrlO},
	tcell.KeyCtrlP: tea.KeyMsg{Type: tea.KeyCtrlP},
	tcell.KeyCtrlQ: tea.KeyMsg{Type: tea.KeyCtrlQ},
	tcell.KeyCtrlR: tea.KeyMsg{Type: tea.KeyCtrlR},
	tcell.KeyCtrlS: tea.KeyMsg{Type: tea.KeyCtrlS},
	tcell.KeyCtrlT: tea.KeyMsg{Type: tea.KeyCtrlT},
	tcell.KeyCtrlU: tea.KeyMsg{Type: tea.KeyCtrlU},
	tcell.KeyCtrlV: tea.KeyMsg{Type: tea.KeyCtrlV},
	tcell.KeyCtrlW: tea.KeyMsg{Type: tea.KeyCtrlW},
	tcell.KeyCtrlX: tea.KeyMsg{Type: tea.KeyCtrlX},
	tcell.KeyCtrlY: tea.KeyMsg{Type: tea.KeyCtrlY},
	tcell.KeyCtrlZ: tea.KeyMsg{Type: tea.KeyCtrlZ},
	//	tcell.KeyCtrlSpace:      tea.KeyMsg{Type: tea.KeyCtrlS},
	tcell.KeyCtrlUnderscore: tea.KeyMsg{Type: tea.KeyCtrlUnderscore},
	//	tcell.KeyCtrlRightSq:    tea.KeyMsg{Type: tea.KeyCtr},
	tcell.KeyCtrlBackslash: tea.KeyMsg{Type: tea.KeyCtrlBackslash},
	tcell.KeyCtrlCarat:     tea.KeyMsg{Type: tea.KeyCtrlCaret},
}

type TcellKeyToTeaMsgResult struct {
	Msg tea.KeyMsg
	Ok  bool
}

func MapKeyMsg(msg KeyMsg) (tea.KeyMsg, bool) {
	a := MapTCellKeyToTeaMsg(*msg.EventKey)
	return a.Msg, a.Ok
}

func MapTCellKeyToTeaMsg(ev tcell.EventKey) TcellKeyToTeaMsgResult {
	if msg, ok := tCellKeyToTeaMsg[ev.Key()]; ok {
		return TcellKeyToTeaMsgResult{Msg: msg, Ok: true}
	}
	if ev.Rune() != 0 {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ev.Rune()}}
		if ev.Modifiers() == tcell.ModAlt {
			msg.Alt = true
		}
		return TcellKeyToTeaMsgResult{Msg: msg, Ok: true}
	}
	return TcellKeyToTeaMsgResult{Ok: false}
}
