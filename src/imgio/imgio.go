package imgio

import (
	"fmt"
	"image"
	"image/color"
	"regexp"
	"strings"

	"gioui.org/f32"
	"gioui.org/gesture"
	"gioui.org/io/event"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
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
	gtx   layout.Context
}

func NewIm(theme *material.Theme) *Im {
	im := &Im{
		widgets: map[string]any{},
		theme:   theme,
	}

	return im
}

func fromCache[T any](i *Im, key string, makeValue func() T) T {
	item, exists := i.widgets[key]
	if !exists {
		item = makeValue()
		i.widgets[key] = item
	}
	return item.(T)
}

var findHashes = regexp.MustCompile("(.*?)(##.*)").FindStringSubmatch

// From a string, extract the id that should be used uniquely
// Follows imgui's label conventions of ## and ###
func (i *Im) getId(str string) (label, id string) {
	parts := findHashes(str)
	if len(parts) == 0 {
		return str, str
	}
	if strings.HasPrefix(parts[2], "###") {
		return parts[1], parts[2]
	}
	return parts[1], parts[0]
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
	return layout.Inset{
		Left:  unit.Dp(5),
		Right: unit.Dp(5),
	}.Layout(gtx,
		func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:    layout.Vertical,
				Spacing: layout.SpaceEnd,
			}.Layout(
				gtx,
				children...)
		})
}

func (i *Im) Button(title string) bool {
	label, id := i.getId(title)
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
	label, id := i.getId(label)
	lineEditor := fromCache(i, id, func() *widget.Editor {
		lineEditor := &widget.Editor{
			SingleLine: true,
			Submit:     true,
		}
		i.AddUpdater(func() {
			for {
				_, more := lineEditor.Update(i.gtx)
				if more {
					*textVariable = lineEditor.Text()
				} else {
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

func (i *Im) SliderFloat(label string, float *float64, min, max float64) {
	label, id := i.getId(label)
	w := fromCache(i, id, func() layout.Widget {
		scale := max - min
		f := widget.Float{}
		return func(gtx layout.Context) layout.Dimensions {
			*float = float64(f.Value)*scale + min
			inset := layout.UniformInset(unit.Dp(8))
			return layout.Flex{}.Layout(gtx,
				layout.Flexed(1, material.Slider(i.theme, &f).Layout),
				rigid(&inset, i.text("%s % 7.3f", label, *float)))
		}
	})
	i.AddWidget(w)
}

type WindowManager struct {
	globalPos    f32.Point
	dragStartPos f32.Point
}

func (w *WindowManager) Layout(gtx layout.Context) {
	event.Op(gtx.Ops, w)
	forEvent(gtx.Source, pointer.Filter{
		Target: w,
		Kinds:  pointer.Press | pointer.Drag,
	}, func(e pointer.Event) bool {
		switch e.Kind {
		case pointer.Press:
			w.dragStartPos = e.Position
			w.globalPos = e.Position
			//fmt.Printf("wm pressed %v %v\n\n", w, e.Position)
		case pointer.Drag:
			w.globalPos = e.Position
			//fmt.Printf("wm global %v %v\n", w, e.Position)
		}
		return true
	})
}

type Window struct {
	Parent         *WindowManager
	Pos            f32.Point
	Size           f32.Point
	windowStartPos f32.Point
	drag           gesture.Drag
}

func (w *Window) Layout(gtx layout.Context, child func(gtx layout.Context) layout.Dimensions) layout.Dimensions {
	// Apply the window constraints.
	gtx.Constraints.Max = w.Size.Round()

	// Move the window
	trans := op.Offset(w.Pos.Round()).Push(gtx.Ops)
	defer trans.Pop()

	// restrict to the window
	area := clip.Rect(image.Rect(0, 0, int(w.Size.X), int(w.Size.Y))).Push(gtx.Ops)
	// paint the window background
	paint.ColorOp{Color: color.NRGBA{G: 0x80, A: 0xFF}}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	titlebarHeight := unit.Dp(25)
	titlebar := clip.Rect(image.Rect(0, 0, int(w.Size.X), gtx.Metric.Dp(titlebarHeight))).Push(gtx.Ops)
	paint.ColorOp{Color: color.NRGBA{R: 0x80, A: 0xFF}}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	// register for titlebar events.  Using a pointer to a member just for uniqueness
	event.Op(gtx.Ops, w)
	titlebar.Pop()
	layout.Inset{Top: titlebarHeight}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return child(gtx)
	})

	area.Pop()

	forEvent(gtx.Source, pointer.Filter{
		Target: w,
		Kinds:  pointer.Drag | pointer.Press,
	}, func(e pointer.Event) bool {
		switch e.Kind {
		case pointer.Press:
			w.windowStartPos = w.Pos
		case pointer.Drag:
			w.Pos = w.windowStartPos.Add(w.Parent.globalPos.Sub(w.Parent.dragStartPos))
		}
		//fmt.Printf("local %v %v\n", w, e.Position)
		return true
	})
	return layout.Dimensions{Size: w.Size.Round()}
}

// forEvent will filter s according to filter and try to convert any matching events
// to type T.  body will be called for all events that pass both the filter and the cast.
// To exit early, body can return false
func forEvent[T any](s input.Source, filter event.Filter, body func(evt T) bool) {
	for {
		ev, ok := s.Event(filter)
		if !ok {
			return
		}
		e, ok := ev.(T)
		if !ok {
			continue
		}
		if !body(e) {
			return
		}
	}
}

var (
	gGtx      layout.Context
	gToLayout []*Im
	gTempWm   *WindowManager
)

func SetContext(gtx layout.Context) {
	gGtx = gtx
}

func Layout() {
	for _, im := range gToLayout {
		im.Layout(gGtx)
	}
	gToLayout = nil
}

func TempSetWm(wm *WindowManager) {
	gTempWm = wm
}

var windows = make(map[string]*Window)

func Begin(im *Im, name string, open *bool, body func(im *Im)) {
	if open != nil && *open {
		win, ok := windows[name]
		if !ok {
			win = &Window{Parent: gTempWm, Size: f32.Pt(500, 400)}
			windows[name] = win
		}
		body(im)
		win.Layout(gGtx, im.Layout)
	}
}
