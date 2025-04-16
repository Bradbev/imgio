package windowexample

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"os"

	"gioui.org/app"
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

func Main() {
	go func() {
		// create new window
		w := new(app.Window)
		w.Option(app.Size(unit.Dp(400), unit.Dp(600)))
		if err := draw(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()

	app.Main()
}

type C = layout.Context
type D = layout.Dimensions

func draw(w *app.Window) error {
	var ops op.Ops

	th := material.NewTheme()

	var button widget.Clickable
	var button2 widget.Clickable
	wm := &WindowManager{}
	win1 := &Window{parent: wm, size: f32.Pt(300, 300)}
	win2 := &Window{parent: wm, size: f32.Pt(200, 200), pos: f32.Pt(200, 400)}

	for {
		switch e := w.Event().(type) {
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			wm.Layout(gtx)

			winLayout := func(win *Window, b *widget.Clickable, name string) {
				win.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis:    layout.Vertical,
						Spacing: layout.SpaceEnd,
					}.Layout(gtx, layout.Rigid(func(gtx C) D {
						return layout.UniformInset(unit.Dp(5)).Layout(gtx, func(gtx C) D {
							return material.Button(th, b, name).Layout(gtx)
						})
					}))
				})

			}
			winLayout(win1, &button, "First")
			winLayout(win2, &button2, "Second")

			e.Frame(gtx.Ops)

		// this is sent when the application is closed.
		case app.DestroyEvent:
			return e.Err
		}
	}
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
			fmt.Printf("wm pressed %v %v\n\n", w, e.Position)
		case pointer.Drag:
			w.globalPos = e.Position
			fmt.Printf("wm global %v %v\n", w, e.Position)
		}
		return true
	})
}

type Window struct {
	parent         *WindowManager
	pos            f32.Point
	windowStartPos f32.Point
	size           f32.Point
	drag           gesture.Drag
}

func (w *Window) Layout(gtx layout.Context, child func(gtx layout.Context) layout.Dimensions) layout.Dimensions {
	// Apply the window constraints.
	gtx.Constraints.Max = w.size.Round()

	// Move the window
	trans := op.Offset(w.pos.Round()).Push(gtx.Ops)
	defer trans.Pop()

	// restrict to the window
	area := clip.Rect(image.Rect(0, 0, int(w.size.X), int(w.size.Y))).Push(gtx.Ops)

	// register for window-bound events.  Using a pointer to a member just for uniqueness
	event.Op(gtx.Ops, w)

	// paint the window area
	paint.ColorOp{Color: color.NRGBA{G: 0x80, A: 0xFF}}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)

	child(gtx)
	area.Pop()

	forEvent(gtx.Source, pointer.Filter{
		Target: w,
		Kinds:  pointer.Drag | pointer.Press,
	}, func(e pointer.Event) bool {
		switch e.Kind {
		case pointer.Press:
			w.windowStartPos = w.pos
		case pointer.Drag:
			w.pos = w.windowStartPos.Add(w.parent.globalPos.Sub(w.parent.dragStartPos))
		}
		fmt.Printf("local %v %v\n", w, e.Position)
		return true
	})
	return layout.Dimensions{Size: w.size.Round()}
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
