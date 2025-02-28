package peanutbutter

import (
	"github.com/charmbracelet/lipgloss"
	tcellviews "github.com/gdamore/tcell/v2/views"
	"github.com/mattn/go-runewidth"
)

func renderBorder(focus bool, p PanelStyle, view *tcellviews.ViewPort) {
	borderStyle := p.UnfocusedBorder
	if focus {
		borderStyle = p.FocusedBorder
	}

	str := borderStyle.Render(
		lipgloss.JoinVertical(lipgloss.Top, ""),
	)

	X, Y := view.Size()
	sx, sy, horz, vert := GetStylingMargins(&p)
	innerView := tcellviews.NewViewPort(view, sx, sy, X-horz, Y-vert)
	TcellDrawHelper(str, view, []*tcellviews.ViewPort{innerView})
}

type offSetSide int

const (
	offsetFromLeftSide offSetSide = iota
	offsetFromRightSide
)

type renderEdge int

const (
	renderOnTopEdge renderEdge = iota
	renderOnBottomEdge
)

func renderTextOnBorder(text string, edge renderEdge, side offSetSide, offset int, view *tcellviews.ViewPort) {
	numCells := runewidth.StringWidth(text)
	X, Y := view.Size()
	y := 0
	if edge == renderOnBottomEdge {
		y = Y - 1
	}
	x := offset
	if side == offsetFromRightSide {
		x = X - numCells - offset
		if x < 0 {
			x = 0
		}
	} else {
		x = offset
	}

	miniView := tcellviews.NewViewPort(view, x, y, numCells, 1)
	TcellDrawHelper(text, miniView, []*tcellviews.ViewPort{})
}
