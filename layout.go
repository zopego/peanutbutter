package panelbubble

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

func (l Layout) IsLayoutValid() bool {
	return l.AreDimensionsValid(false)
}

func (l Layout) AreDimensionsValid(printErrors bool) bool {
	numConstrained := 0
	numUnspecified := 0
	for i, d := range l.Dimensions {
		if !d.IsValid() {
			if printErrors {
				fmt.Printf("Invalid dimension at index %d: %+v\n", i, d)
			}
			return false
		}
		if d.IsConstrained() {
			numConstrained++
		}
		if d.IsUnspecified() {
			numUnspecified++
		}
	}
	if numConstrained > 1 {
		if printErrors && numUnspecified != 1 {
			fmt.Printf("There needs to be exactly one unspecified dimension, found %d unspecified & %d constrained\n", numUnspecified, numConstrained)
			fmt.Printf("Layout: %+v\n", l)
		}
		return numUnspecified == 1
	} else {
		if printErrors && numUnspecified > 0 {
			fmt.Printf("There should be no unspecified dimensions, found %d unspecified & %d constrained\n", numUnspecified, numConstrained)
			fmt.Printf("Layout: %+v\n", l)
		}
		return numUnspecified == 0
	}
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
	return l.Layout.AreDimensionsValid(printErrors)
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
