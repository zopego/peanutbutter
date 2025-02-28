package peanutbutter

import (
	catppuccin "github.com/catppuccin/go"
	"github.com/charmbracelet/lipgloss"
	tcellviews "github.com/gdamore/tcell/v2/views"
)

type PanelStyle struct {
	FocusedBorder   lipgloss.Style
	UnfocusedBorder lipgloss.Style
}

type TitleStyle struct {
	FocusedTitle   lipgloss.Style
	UnfocusedTitle lipgloss.Style
}

func (t *TitleStyle) RenderTitle(title string, focus bool) string {
	if focus {
		return t.FocusedTitle.Render(title)
	}
	return t.UnfocusedTitle.Render(title)
}

var Plt = catppuccin.Latte

func LgColor(color catppuccin.Color) lipgloss.Color {
	return lipgloss.Color(color.Hex)
}

func Fg(color catppuccin.Color) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(LgColor(color))
}

func Bg(color catppuccin.Color) lipgloss.Style {
	return lipgloss.NewStyle().Background(LgColor(color))
}

func FgBg(fg catppuccin.Color, bg catppuccin.Color) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(LgColor(fg)).Background(LgColor(bg))
}

var DefaultPanelConfig = ShortCutPanelConfig{
	PanelStyle: PanelStyle{
		FocusedBorder: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(LgColor(Plt.Lavender())),

		UnfocusedBorder: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(LgColor(Plt.Text())),
	},
	TitleStyle: TitleStyle{
		FocusedTitle:   lipgloss.NewStyle().Bold(true).Padding(0, 1),
		UnfocusedTitle: lipgloss.NewStyle().Padding(0, 1),
	},
}

var NoBorderPanelStyle = PanelStyle{
	FocusedBorder:   lipgloss.NewStyle(),
	UnfocusedBorder: lipgloss.NewStyle(),
}

func GetStylingMargins(p *PanelStyle) (int, int, int, int) {
	if p == nil {
		return 0, 0, 0, 0
	}
	s := p.UnfocusedBorder

	horz_left := s.GetBorderLeftSize() + s.GetMarginLeft() + s.GetPaddingLeft()
	vert_top := s.GetBorderTopSize() + s.GetMarginTop() + s.GetPaddingTop()

	horz_total := s.GetHorizontalMargins() + s.GetHorizontalPadding() + s.GetHorizontalBorderSize()
	vert_total := s.GetVerticalMargins() + s.GetVerticalPadding() + s.GetVerticalBorderSize()
	return horz_left, vert_top, horz_total, vert_total
}

func SetSize(p *PanelStyle, view *tcellviews.ViewPort, x int, y int, horz_total int, vert_total int) {
	if view != nil {
		view.Resize(x, y, horz_total, vert_total)
	}
	if p == nil {
		return
	}

	_, _, h, v := GetStylingMargins(p)

	p.UnfocusedBorder = p.UnfocusedBorder.Width(horz_total - h)
	p.FocusedBorder = p.FocusedBorder.Width(horz_total - h)
	p.UnfocusedBorder = p.UnfocusedBorder.Height(vert_total - v)
	p.FocusedBorder = p.FocusedBorder.Height(vert_total - v)
}
