package panelbubble

import (
	"fmt"

	tcell "github.com/gdamore/tcell/v2"
	tcellviews "github.com/gdamore/tcell/v2/views"
	ansi "github.com/leaanthony/go-ansi-parser"
	"github.com/mattn/go-runewidth"
)

func translateAnsiStyleToTcellStyle(s ansi.TextStyle) tcell.AttrMask {
	attrs := tcell.AttrMask(0)
	if s&ansi.Bold != 0 {
		attrs |= tcell.AttrBold
	}
	if s&ansi.Faint != 0 {
		attrs |= tcell.AttrDim
	}
	if s&ansi.Italic != 0 {
		attrs |= tcell.AttrItalic
	}
	if s&ansi.Blinking != 0 {
		attrs |= tcell.AttrBlink
	}
	if s&ansi.Inversed != 0 {
		attrs |= tcell.AttrReverse
	}
	if s&ansi.Underlined != 0 {
		attrs |= tcell.AttrUnderline
	}
	if s&ansi.Strikethrough != 0 {
		attrs |= tcell.AttrStrikeThrough
	}
	if s&ansi.Bright != 0 {
		attrs |= tcell.AttrBold
	}
	return attrs
}

func TcellDrawHelper(j string, s tcellviews.View, preserveViews []*tcellviews.ViewPort) {
	view, err := ansi.Parse(j)
	if err != nil {
		fmt.Printf("Error parsing view: %v\n", err)
		return
	}
	x := 0
	y := 0
	for _, block := range view {
		for _, r := range block.Label {
			var fgc tcell.Color = tcell.ColorDefault
			var bgc tcell.Color = tcell.ColorDefault
			if block.FgCol != nil {
				fg := block.FgCol.Rgb
				fgc = tcell.NewRGBColor(int32(fg.R), int32(fg.G), int32(fg.B))
			}
			if block.BgCol != nil {
				bg := block.BgCol.Rgb
				bgc = tcell.NewRGBColor(int32(bg.R), int32(bg.G), int32(bg.B))
			}
			style := tcell.StyleDefault.Foreground(fgc).Background(bgc).Attributes(translateAnsiStyleToTcellStyle(block.Style))
			if r == '\n' {
				x = 0
				y++
			} else {
				// if width of rune is more than 1, then we need to split it
				rw := runewidth.RuneWidth(r)
				// check if x, y happen to be in the preserveViews list
				skipRune := false
				for _, v := range preserveViews {
					px, py, pX, pY := v.GetPhysical()
					if x >= px && x <= pX && y >= py && y <= pY {
						skipRune = true
						break
					}
				}
				if !skipRune {
					s.SetContent(x, y, r, nil, style)
				}
				x += rw
			}
		}
	}
}
