package imgio

import (
	"encoding/json"
	"fmt"
	"image/color"
	"os"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// Needs to hold a shadow dom
// If the layout is the same each loop, reuse the widgets.  If it changes, reuse where possible
// otherwise rebuild
// Implies that items need to have unique ids and that, ie changing a button title or duping a button title
// needs to have an escape hatch
type Im struct {
	widgets      map[string]any
	widgetsOrder []layout.Widget
	updaters     []func()

	theme *material.Theme
	axis  layout.Axis
	gtx   layout.Context
}

func NewIm(theme *material.Theme) *Im {
	im := &Im{
		widgets: map[string]any{},
		theme:   theme,
		axis:    layout.Vertical,
	}
	return im
}

func (i *Im) WithSameLine(body func(im *Im)) {

	newIm := fromCache(i, "sameline", func() *Im {
		im := NewIm(i.theme)
		im.axis = layout.Horizontal
		return im
	})
	newIm.Reset(i.gtx)
	body(newIm)
	i.AddWidget(newIm.Layout)
	/*
	   var children []layout.FlexChild

	   	for _, d := range newIm.widgetsOrder {
	   		children = append(children, layout.Rigid(d))
	   	}

	   	i.AddWidget(func(gtx layout.Context) layout.Dimensions {
	   		return layout.Flex{}.Layout(gtx, children...)
	   	})
	*/
}

func (i *Im) Reset(gtx layout.Context) {
	i.widgetsOrder = i.widgetsOrder[:0]
	i.gtx = gtx

	for _, u := range i.updaters {
		u()
	}
}

func (i *Im) AddWidget(w layout.Widget) {
	withInset := func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top:    unit.Dp(2),
			Bottom: unit.Dp(2),
		}.Layout(gtx, w)
	}
	i.widgetsOrder = append(i.widgetsOrder, withInset)
}

func (i *Im) AddUpdater(updater func()) {
	i.updaters = append(i.updaters, updater)
}

func (i *Im) Layout(gtx layout.Context) layout.Dimensions {
	var children []layout.FlexChild
	for _, d := range i.widgetsOrder {
		children = append(children, layout.Rigid(d))
	}
	spacing := layout.SpaceEnd
	if i.axis == layout.Horizontal {
		return layout.Flex{Spacing: layout.SpaceEvenly}.Layout(gtx, children...)
	}
	return layout.Inset{
		Left:  unit.Dp(5),
		Right: unit.Dp(5),
	}.Layout(gtx,
		func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:    i.axis,
				Spacing: spacing,
			}.Layout(
				gtx,
				children...)
		})
}

func (i *Im) Button(label string) bool {
	label, id := getId(label, "button")
	btn := fromCache(i, id, func() *widget.Clickable {
		return new(widget.Clickable)
	})
	i.AddWidget(material.Button(i.theme, btn, label).Layout)
	return btn.Clicked(i.gtx)
}

func (i *Im) Text(s string, args ...any) {
	i.AddWidget(i.text(s, args...))
}

func (i *Im) text(s string, args ...any) func(gtx layout.Context) layout.Dimensions {
	s = fmt.Sprintf(s, args...)
	return material.Body1(i.theme, s).Layout
}

func (i *Im) InputText(label string, textVariable *string) {
	label, id := getId(label, "inputtext")
	lineEditor := fromCache(i, id, func() *widget.Editor {
		lineEditor := &widget.Editor{
			SingleLine: true,
			Submit:     true,
		}
		i.AddUpdater(func() {
			for {
				_, keepGoing := lineEditor.Update(i.gtx)
				if keepGoing {
					*textVariable = lineEditor.Text()
				} else {
					// the backing string might have changed, update the editor text
					if *textVariable != lineEditor.Text() {
						_, col := lineEditor.CaretPos()
						lineEditor.SetText(*textVariable)
						lineEditor.SetCaret(col, col)
					}
					break
				}
			}
		})
		return lineEditor
	})

	inset := layout.UniformInset(unit.Dp(8))
	editBox := layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
		e := material.Editor(i.theme, lineEditor, "")
		border := widget.Border{Color: color.NRGBA{A: 0xff}, CornerRadius: unit.Dp(2), Width: unit.Dp(2)}
		return border.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return inset.Layout(gtx, e.Layout)
		})
	})
	text := rigid(&inset, material.Body1(i.theme, label).Layout)
	widget := func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{}.Layout(gtx,
			text,
			editBox,
		)
	}

	i.AddWidget(widget)
}

func rigid(inset *layout.Inset, w layout.Widget) layout.FlexChild {
	return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return inset.Layout(gtx, w)
	})
}

// returns true if value changed
func (i *Im) SliderFloat(label string, float *float64, min, max float64) bool {
	label, id := getId(label, "sliderfloat")
	w := fromCache(i, id, func() layout.Widget {
		scale := max - min
		f := widget.Float{Value: float32(*float)}
		return func(gtx layout.Context) layout.Dimensions {
			*float = float64(f.Value)*scale + min
			inset := layout.UniformInset(unit.Dp(8))
			return layout.Flex{}.Layout(gtx,
				layout.Flexed(1, material.Slider(i.theme, &f).Layout),
				rigid(&inset, i.text("%s % 7.3f", label, *float)))
		}
	})
	i.AddWidget(w)
	return true
}

var (
	gGtx        layout.Context
	gTempWm     *WindowManager
	gWindows    = make(map[string]*Window)
	gSavedState = make(map[string]json.RawMessage)
	gTheme      *material.Theme
)

const saveFileName = "imgio.json"
const themeFileName = "theme.json"

func Init() {
	gTheme = material.NewTheme()
	gTheme.Face = "monospace"

	toLoad, err := os.ReadFile(saveFileName)
	if err == nil {
		json.Unmarshal(toLoad, &gSavedState)
	}

	toLoad, err = os.ReadFile(themeFileName)
	if err == nil {
		err = json.Unmarshal(toLoad, &gTheme.Palette)
		println(err)
	}
}

func GetTheme() *material.Theme {
	return gTheme
}

func SetContext(gtx layout.Context) {
	gGtx = gtx
}

func Layout() {
}

func TempSetWm(wm *WindowManager) {
	gTempWm = wm
}

func Begin(title string, open *bool, body func(im *Im)) {
	if open != nil && *open {
		win, ok := gWindows[title]
		if !ok {
			win = &Window{parent: gTempWm,
				Size:  f32.Pt(500, 400),
				title: title,
			}
			if val, ok := gSavedState[title]; ok {
				json.Unmarshal(val, &win)
			}
			win.im = NewIm(gTheme)
			gWindows[title] = win
		}
		win.closed = false
		win.im.Reset(gGtx)
		body(win.im)
		win.Layout(gGtx, win.im.Layout)
		if win.closed {
			*open = false
		}
	}
}

func DestroyEvent() {
	toSave, _ := json.MarshalIndent(gWindows, "", " ")
	os.WriteFile(saveFileName, toSave, os.ModePerm)

	toSave, _ = json.MarshalIndent(gTheme.Palette, "", " ")
	os.WriteFile(themeFileName, toSave, os.ModePerm)
}
