package imgio

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/gesture"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

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
	Pos           f32.Point
	Size          f32.Point
	parent        *WindowManager
	dragStartPos  f32.Point
	dragStartSize f32.Point
	drag          gesture.Drag
	closeButton   widget.Clickable
	closed        bool
	title         string
	im            *Im
}

func (w *Window) Layout(gtx layout.Context, child func(gtx layout.Context) layout.Dimensions) layout.Dimensions {
	if w.closeButton.Clicked(gtx) {
		w.closed = true
	}

	// Apply the window constraints.
	gtx.Constraints.Max = w.Size.Round()

	// Move the window
	defer op.Offset(w.Pos.Round()).Push(gtx.Ops).Pop()

	rect := func(r image.Rectangle, c color.NRGBA) {
		defer clip.Rect(r).Push(gtx.Ops).Pop()
		paint.ColorOp{Color: c}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
	}
	// draw the outline with a full rect and then an inset rect
	//rect(image.Rect(0, 0, int(w.Size.X), int(w.Size.Y)), gTheme.ContrastBg)
	rect(image.Rect(2, 2, int(w.Size.X-2), int(w.Size.Y-2)), gTheme.Bg)
	// clip subsequent draws to the window area
	r := image.Rectangle{Max: w.Size.Round()}
	paint.FillShape(gtx.Ops, gTheme.ContrastBg, clip.Stroke{
		Path:  clip.UniformRRect(r, 0).Path(gtx.Ops),
		Width: float32(gtx.Metric.Dp(2)),
	}.Op())
	defer clip.Rect(image.Rect(2, 2, int(w.Size.X-2), int(w.Size.Y-2))).Push(gtx.Ops).Pop()

	// titlebar
	titlebarHeight := unit.Dp(35)
	func() {
		defer clip.Rect(image.Rect(0, 0, int(w.Size.X), gtx.Metric.Dp(titlebarHeight))).Push(gtx.Ops).Pop()
		paint.ColorOp{Color: gTheme.ContrastBg}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		event.Op(gtx.Ops, w)

		titleGtx := gtx
		titleGtx.Constraints.Max = image.Pt(int(w.Size.X), gtx.Metric.Dp(titlebarHeight))
		layout.UniformInset(unit.Dp(4)).Layout(titleGtx, func(gtx layout.Context) layout.Dimensions {
			p := gTheme.Palette
			p.Bg, p.Fg = p.Fg, p.Bg
			th := gTheme.WithPalette(p)
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(1, material.Body1(&th, w.title).Layout),
				layout.Rigid(material.Button(gTheme, &w.closeButton, "X").Layout),
			)
		})
	}()

	// Layout the child
	layout.Inset{Top: titlebarHeight}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return child(gtx)
	})

	// draw the corner resize triangle
	func() {
		p := clip.Path{}
		p.Begin(gtx.Ops)
		p.MoveTo(w.Size)
		p.Line(f32.Pt(0, -40))
		p.Line(f32.Pt(-40, 40))
		defer clip.Outline{Path: p.End()}.Op().Push(gtx.Ops).Pop()
		paint.ColorOp{Color: gTheme.ContrastBg}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		event.Op(gtx.Ops, &w.Pos)
	}()

	// corner dragging for resize
	forEvent(gtx.Source, pointer.Filter{
		Target: &w.Pos,
		Kinds:  pointer.Drag | pointer.Press,
	}, func(e pointer.Event) bool {
		switch e.Kind {
		case pointer.Press:
			w.dragStartPos = w.Pos
			w.dragStartSize = w.Size
		case pointer.Drag:
			w.Size = w.dragStartSize.Add(w.parent.globalPos.Sub(w.parent.dragStartPos))
			w.Size = truncPt(w.Size)
		}
		return true
	})

	// title bar dragging for position
	forEvent(gtx.Source, pointer.Filter{
		Target: w,
		Kinds:  pointer.Drag | pointer.Press,
	}, func(e pointer.Event) bool {
		switch e.Kind {
		case pointer.Press:
			w.dragStartPos = w.Pos
		case pointer.Drag:
			w.Pos = w.dragStartPos.Add(w.parent.globalPos.Sub(w.parent.dragStartPos))
		}
		//fmt.Printf("local %v %v\n", w, e.Position)
		return true
	})
	return layout.Dimensions{Size: w.Size.Round()}
}
