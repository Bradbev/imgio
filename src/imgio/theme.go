package imgio

import (
	"sync"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

type Theme struct {
	Palette     *material.Palette
	ButtonInset layout.Inset
	WidgetInset layout.Inset
}

type floatInset struct {
	Top    float64
	Bottom float64
	Left   float64
	Right  float64
}

func fromInset(i layout.Inset) *floatInset {
	return &floatInset{
		Top:    float64(i.Top),
		Bottom: float64(i.Bottom),
		Left:   float64(i.Left),
		Right:  float64(i.Right),
	}
}
func (f *floatInset) toInset(i *layout.Inset) {
	i.Top = unit.Dp(f.Top)
	i.Bottom = unit.Dp(f.Bottom)
	i.Left = unit.Dp(f.Left)
	i.Right = unit.Dp(f.Right)
}

var (
	once        sync.Once
	buttonInset = &floatInset{}
	widgetInset = &floatInset{}
)

func ThemeEdit(open *bool) {
	once.Do(func() {
		buttonInset = fromInset(gImTheme.ButtonInset)
		widgetInset = fromInset(gImTheme.WidgetInset)
	})
	Begin("Theme Edit", open, func(im *Im) {
		im.SliderFloat("Button Top/Bottom", &buttonInset.Top, 0, 20)
		buttonInset.Bottom = buttonInset.Top
		im.SliderFloat("Widget Top/Bottom", &widgetInset.Top, 0, 20)
		widgetInset.Bottom = widgetInset.Top
		im.SliderFloat("Widget Left/Right", &widgetInset.Left, 0, 20)
		widgetInset.Right = widgetInset.Left
	})
	buttonInset.toInset(&gImTheme.ButtonInset)
	widgetInset.toInset(&gImTheme.WidgetInset)

}
