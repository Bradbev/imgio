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
	widgetsOrder []layout.FlexChild
	widgetsHoriz []layout.FlexChild
	updaters     []func()

	samelineActive  bool
	singleSameLine  bool
	lastAddedWidget layout.Widget
	theme           *material.Theme
	axis            layout.Axis
	flex            FlexMode
	gtx             layout.Context
}

type FlexMode uint8

const (
	FlexModeDefault FlexMode = iota
	FlexModeFlex
	FlexModeRigid
)

func NewIm(theme *material.Theme) *Im {
	im := &Im{
		widgets: map[string]any{},
		theme:   theme,
		axis:    layout.Vertical,
	}
	return im
}

func (i *Im) WithSameLine(body func(im *Im)) {
	if i.widgetsHoriz != nil && i.singleSameLine == false && i.samelineActive == false {
		i.EndLine()
	}
	samelineActive := i.samelineActive
	i.samelineActive = true
	body(i)
	i.samelineActive = samelineActive
}

func (i *Im) SameLine() {
	if i.widgetsHoriz == nil {
		// Re-add the last widget because the flex changes
		i.singleSameLine = true
		i.AddWidget(i.lastAddedWidget)
		// remove that last widget from the vertical list
		wo := i.widgetsOrder
		i.widgetsOrder = wo[1 : len(wo)-1]
	}
	i.singleSameLine = true
}

func (i *Im) WithFlexMode(mode FlexMode, body func(im *Im)) {
	current := i.flex
	i.flex = mode
	body(i)
	i.flex = current
}

func (i *Im) Reset(gtx layout.Context) {
	i.widgetsOrder = i.widgetsOrder[:0]
	i.gtx = gtx

	for _, u := range i.updaters {
		u()
	}
}

func (i *Im) AddWidget(widget layout.Widget) {
	withInset := func(gtx layout.Context) layout.Dimensions {
		return gImTheme.WidgetInset.Layout(gtx, widget)
	}
	w := withInset

	var flexchild layout.FlexChild

	switch i.flex {
	case FlexModeFlex:
		flexchild = layout.Flexed(1, w)
	case FlexModeRigid:
		flexchild = layout.Rigid(w)
	case FlexModeDefault:
		if i.samelineActive || i.singleSameLine {
			flexchild = layout.Flexed(1, w)
		} else {
			flexchild = layout.Rigid(w)
		}
	}
	if i.samelineActive || i.singleSameLine {
		i.widgetsHoriz = append(i.widgetsHoriz, flexchild)
		i.singleSameLine = false
	} else {
		i.EndLine()
		i.widgetsOrder = append(i.widgetsOrder, flexchild)
	}
	i.lastAddedWidget = widget
}

func (i *Im) EndLine() {
	if i.widgetsHoriz != nil {
		horiz := i.widgetsHoriz
		w := func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, horiz...)
		}
		i.widgetsOrder = append(i.widgetsOrder, layout.Rigid(w))
		i.widgetsHoriz = nil
	}
}

func (i *Im) AddUpdater(updater func()) {
	i.updaters = append(i.updaters, updater)
}

func (i *Im) Layout(gtx layout.Context) layout.Dimensions {
	i.EndLine()
	spacing := layout.SpaceEnd
	if i.axis == layout.Horizontal {
		return layout.Flex{Spacing: layout.SpaceEvenly}.Layout(gtx, i.widgetsOrder...)
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
				i.widgetsOrder...)
		})
}

func (i *Im) Button(label string) bool {
	label, id := getId(label, "button")
	btn := fromCache(i, id, func() *widget.Clickable {
		return new(widget.Clickable)
	})
	b := material.Button(i.theme, btn, label)
	b.Inset = gImTheme.ButtonInset
	i.AddWidget(b.Layout)
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
		// TODO not needed.  Fold the update into the layout func
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

	inset := layout.UniformInset(unit.Dp(6))
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

var (
	gGtx        layout.Context
	gTempWm     *WindowManager
	gWindows    = make(map[string]*Window)
	gSavedState = make(map[string]json.RawMessage)
	gTheme      *material.Theme
	gImTheme    Theme
	gApp        App
)

const saveFileName = "imgio.json"
const themeFileName = "theme.json"

type App interface {
	Invalidate()
}

func Init(a App) {
	gApp = a
	gTheme = material.NewTheme()
	gTheme.Face = "monospace"
	gImTheme.Palette = &gTheme.Palette

	toLoad, err := os.ReadFile(saveFileName)
	if err == nil {
		json.Unmarshal(toLoad, &gSavedState)
	}

	toLoad, err = os.ReadFile(themeFileName)
	if err == nil {
		json.Unmarshal(toLoad, &gImTheme)
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

	toSave, _ = json.MarshalIndent(gImTheme, "", " ")
	os.WriteFile(themeFileName, toSave, os.ModePerm)
}
