package panelbubble

import (
	"fmt"

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

// If all fields are 0, it is assumed that the panel should take up the remaining space
type Dimension struct {
	Min   int     //valid if > 0, assumed unspecified otherwise
	Max   int     //valid if > 0, assumed unspecified otherwise
	Fixed int     //valid if > 0, assumed unspecified otherwise
	Ratio float64 //valid if > 0, assumed unspecified otherwise
}

func (d Dimension) IsFlexible() bool {
	return d.Ratio <= 0.0 && d.Min <= 0 && d.Max <= 0 && d.Fixed <= 0
}

func (d Dimension) IsValid() bool {
	if d.IsFlexible() {
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

func (l Layout) AreDimensionsValid() bool {
	for _, d := range l.Dimensions {
		if !d.IsValid() {
			return false
		}
	}
	numFlexible := 0
	for _, d := range l.Dimensions {
		if d.IsFlexible() {
			numFlexible++
		}
	}
	return numFlexible <= 1
}

type ResizeMsg struct {
	tea.Msg
	Width  int
	Height int
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
		if d.IsFlexible() {
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
			if ratioSize < d.Min {
				ratioSize = d.Min
			}
			if ratioSize > d.Max {
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
