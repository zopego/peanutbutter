package panelbubble

import (
	O "github.com/IBM/fp-go/option"
	tea "github.com/charmbracelet/bubbletea"
)

type Orientation int

const (
	Horizontal Orientation = iota
	Vertical
	ZStacked
)

// Layout is a struct that describes the layout of a panel
type Layout struct {
	Orientation Orientation
	Dimensions  []Dimension
}

// Atleast one of the fields should be present
type Dimension struct {
	Min   O.Option[int]
	Max   O.Option[int]
	Fixed O.Option[int]
	Ratio float64
}

// In this order of precedence:
// If Fixed is specified, it is always fixed
// If Fixed is not specified, Min is respected but can be ratiometrically scaled
// If Max is specified, it is always respected but can be ratiometrically scaled
// If Ratio is specified, it is always respected but can be ratiometrically scaled

func (d Dimension) GetMin() int {
	min, ok := O.Unwrap(d.Min)
	if ok {
		return min
	}
	fixed, ok := O.Unwrap(d.Fixed)
	if ok {
		return fixed
	}
	return 0
}

func (d Dimension) GetMax() int {
	max, ok := O.Unwrap(d.Max)
	if ok {
		return max
	}
	fixed, ok := O.Unwrap(d.Fixed)
	if ok {
		return fixed
	}
	return 9999
}

type ResizeMsg struct {
	tea.Msg
	Width  int
	Height int
}

// Layout can not magically find a size that works for all panels,
// but when possible, it will try to respect the constraints.
// If available size is less than minimum size, all panels will be set to minimum size
// If available size is more than maximum size, all panels will be set to maximum size
// If available size is in between min and max, all panels will be set to ratio-metric size
