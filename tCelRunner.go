package peanutbutter

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	tcell "github.com/gdamore/tcell/v2"
	tcellviews "github.com/gdamore/tcell/v2/views"
)

type pbRunModel struct {
	model IRootModel
	s     tcell.Screen
	quit  chan struct{}
	cmds  chan tea.Cmd
}

func (t *pbRunModel) update(ev tcell.Event) {
	switch ev := ev.(type) {
	case *TeaCmdMsgEvent:
		if ev.Msg == nil {
			return
		}
		if batchMsg, ok := ev.Msg.(tea.BatchMsg); ok {
			for _, cmd := range batchMsg {
				if cmd != nil {
					t.cmds <- cmd
				}
			}
			return
		}
		if _, ok := ev.Msg.(tea.QuitMsg); ok {
			t.s.Fini()
			t.quit <- struct{}{}
			return
		}
		t.model.Update(ev.Msg)
		redraw := t.model.Draw()
		if redraw {
			t.s.Show()
		}

	case *tcell.EventKey:
		unused := true
		kmsg := KeyMsg{EventKey: ev, Unused: &unused}
		t.model.Update(kmsg)
		if t.model.Draw() {
			t.s.Show()
		}

	case *tcell.EventResize:
		w, h := ev.Size()
		resizeMsg := ResizeMsg{EventResize: ev, Width: int(w), Height: int(h)}
		t.model.Update(resizeMsg)
		redraw := t.model.Draw()
		if redraw {
			t.s.Sync()
		}
	}
}

type TeaCmdMsgEvent struct {
	Time time.Time
	Msg  tea.Msg
}

var _ tcell.Event = &TeaCmdMsgEvent{}

func (t *TeaCmdMsgEvent) When() time.Time {
	return t.Time
}

func HandleBatchCmds(msg Msg) []tea.Cmd {
	if batchMsg, ok := msg.(tea.BatchMsg); ok {
		cmds := []tea.Cmd{}
		for _, cmd := range batchMsg {
			cmds = append(cmds, cmd)
		}
		return cmds
	}
	if amsg, ok := msg.(AutoRoutedMsg); ok {
		if batchMsg, ok := amsg.Msg.(tea.BatchMsg); ok {
			cmds := []tea.Cmd{}
			for _, cmd := range batchMsg {
				cmds = append(cmds, MakeAutoRoutedCmd(cmd, amsg.RoutePath.Path))
			}
			return cmds
		}
	}
	return nil
}

func Run(model IRootModel, screen tcell.Screen) {

	cmds := make(chan tea.Cmd, 100)
	viewPort := tcellviews.NewViewPort(screen, 0, 0, -1, -1)

	quit := make(chan struct{})
	tmodel := pbRunModel{
		model: model,
		s:     screen,
		quit:  quit,
	}

	cmds <- model.Init(cmds, viewPort)
	evts := make(chan tcell.Event, 100)
	go func() {
		for cmd := range cmds {
			if cmd == nil {
				continue
			}
			go func() {
				msg := cmd()
				if msg != nil {
					bcmds := HandleBatchCmds(msg)
					if bcmds != nil {
						for _, cmd := range bcmds {
							cmds <- cmd
						}
					} else {
						screen.PostEvent(&TeaCmdMsgEvent{Time: time.Now(), Msg: msg})
					}
				}
			}()
		}
		screen.Fini()
	}()

	go func() {
		for evt := range evts {
			tmodel.update(evt)
		}
	}()

	for {
		evt := screen.PollEvent()
		if evt == nil {
			close(evts)
			close(quit)
			break
		}
		evts <- evt
	}

	<-tmodel.quit

}
