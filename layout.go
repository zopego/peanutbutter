package peanutbutter

import (
	"fmt"

	tcell "github.com/gdamore/tcell/v2"
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
	Width       int
	Height      int
}

// If all fields are 0, it is assumed that the panel should take up the remaining space
type Dimension struct {
	Min   int     //valid if > 0, assumed unspecified otherwise
	Max   int     //valid if > 0, assumed unspecified otherwise
	Fixed int     //valid if > 0, assumed unspecified otherwise
	Ratio float64 //valid if > 0, assumed unspecified otherwise
}

func (d Dimension) IsUnspecified() bool {
	return d.Ratio <= 0.0 && d.Min <= 0 && d.Max <= 0 && d.Fixed <= 0
}

func (d Dimension) IsPureRatio() bool {
	return d.Ratio > 0.0 && d.Min <= 0 && d.Max <= 0 && d.Fixed <= 0
}

func (d Dimension) IsPureFixed() bool {
	return d.Fixed > 0 && d.Ratio <= 0.0 && d.Min <= 0 && d.Max <= 0
}

func (d Dimension) IsConstrained() bool {
	return d.IsPureFixed() || (d.Ratio > 0.0 && (d.Max > 0 || d.Min > 0))
}

func (d Dimension) IsValid() bool {
	if d.IsUnspecified() {
		return true
	}

	if d.Ratio > 0.0 {
		return true
	}

	if d.Fixed > 0 {
		return true
	}

	return false
}

func (l ListPanel) IsLayoutValid() bool {
	return l.AreDimensionsValid(false)
}

func (l ListPanel) AreDimensionsValid(printErrors bool) bool {
	if l.Layout.Orientation == ZStacked {
		return true
	}
	if len(l.Panels) != len(l.Layout.Dimensions) {
		if printErrors {
			fmt.Printf("Number of panels (%d) does not match number of dimensions (%d)\n", len(l.Panels), len(l.Layout.Dimensions))
		}
		return false
	}
	return true
}

type ResizeMsg struct {
	EventResize *tcell.EventResize
	X           int
	Y           int
	Width       int
	Height      int
}

// Here we assume that the layout is valid
func CalculateDimensions(dimensions []Dimension, total int) []int {
	sizes := make([]int, len(dimensions))
	adjustIndex := -1
	allocated := 0
	for i, d := range dimensions {
		if !d.IsValid() {
			fmt.Println("Invalid dimension", d)
			sizes[i] = 0
			continue
		}
		if d.IsUnspecified() {
			adjustIndex = i
			continue
		}
		if d.Fixed > 0 {
			sizes[i] = d.Fixed
			allocated += d.Fixed
			continue
		}
		if d.Ratio > 0.0 {
			ratioSize := int(float64(total) * d.Ratio)
			if d.Min > 0 && ratioSize < d.Min {
				ratioSize = d.Min
			}
			if d.Max > 0 && ratioSize > d.Max {
				ratioSize = d.Max
			}
			sizes[i] = ratioSize
			allocated += ratioSize
		}
	}
	if adjustIndex == -1 {
		return sizes
	} else {
		sizes[adjustIndex] = total - allocated
	}
	return sizes
}

// Preconditions: The layout is Vertical or Horizontal
// Total > 0
func (l Layout) CalculateDims(total int) []int {
	return CalculateDimensions(l.Dimensions, total)
}
