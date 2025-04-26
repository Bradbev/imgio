package imgio

import (
	"fmt"
	"image"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/gesture"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type sliderFloatCtx struct {
	changed bool
	w       layout.Widget
}

// returns true if value changed
func (i *Im) SliderFloat(label string, float *float64, min, max float64) bool {
	label, id := getId(label, "sliderfloat")
	sliderCtx := fromCache(i, id, func() *sliderFloatCtx {
		scale := max - min
		f := widget.Float{Value: float32((*float - min) / scale)}
		s := &sliderFloatCtx{}
		s.w = func(gtx layout.Context) layout.Dimensions {
			s.changed = false
			newVal := float64(f.Value)*scale + min
			if *float != newVal {
				*float = newVal
				gApp.Invalidate()
				s.changed = true
			}
			return material.Slider(i.theme, &f).Layout(gtx)
		}
		return s
	})
	i.WithSameLine(func(im *Im) {
		i.AddWidget(sliderCtx.w)
		i.WithFlexMode(FlexModeRigid, func(im *Im) {
			i.Text("%s % 7.3f", label, *float)
		})
	})
	return sliderCtx.changed
}

type DragFloatCtx struct {
	w       layout.Widget
	drag    Drag
	Changed bool
	Value   float64
	Min     float64
	Max     float64
}

func (i *Im) DragFloat(label string, value *float64, speed, minv, maxv float64, format string) bool {
	label, id := getId(label, "dragint")
	ctx := fromCache(i, id, func() *DragFloatCtx {
		return MakeDragFloatContext(*value, speed, minv, maxv, func(delta f32.Point, ctx *DragFloatCtx) string {
			*value = ctx.Value
			return fmt.Sprintf(format, *value)
		})
	})
	i.AddWidget(ctx.w)
	i.SameLine()
	i.Text(label)
	return ctx.Changed
}

func (i *Im) DragFloat32(label string, value *float32, speed, minv, maxv float64, format string) bool {
	label, id := getId(label, "dragint")
	ctx := fromCache(i, id, func() *DragFloatCtx {
		return MakeDragFloatContext(float64(*value), speed, minv, maxv, func(delta f32.Point, ctx *DragFloatCtx) string {
			*value = float32(ctx.Value)
			return fmt.Sprintf(format, ctx.Value)
		})
	})
	i.AddWidget(ctx.w)
	i.SameLine()
	i.Text(label)
	return ctx.Changed
}
func (i *Im) DragInt(label string, value *int64, speed float64, minv, maxv int64, format string) bool {
	label, id := getId(label, "dragint")
	ctx := fromCache(i, id, func() *DragFloatCtx {
		vf, minf, maxf := float64(*value), float64(minv), float64(maxv)
		return MakeDragFloatContext(vf, speed, minf, maxf, func(delta f32.Point, ctx *DragFloatCtx) string {
			*value = int64(ctx.Value)
			return fmt.Sprintf(format, *value)
		})
	})
	i.AddWidget(ctx.w)
	return ctx.Changed
}

func MakeDragFloatContext(value, speed, minValue, maxValue float64, callback func(delta f32.Point, ctx *DragFloatCtx) string) *DragFloatCtx {
	ctx := &DragFloatCtx{
		Value: value,
	}
	fnt := font.Font{}
	size := gTheme.TextSize * 14.0 / 16.0
	fnt.Typeface = gTheme.Face
	ctx.w = func(gtx layout.Context) layout.Dimensions {

		if ctx.Changed {
			gApp.Invalidate()
		}
		delta := ctx.drag.Update(gtx.Metric, gtx.Source, gesture.Horizontal)
		ctx.Changed = delta.X != 0 || delta.Y != 0
		ctx.Value = clamp(ctx.Value+float64(delta.X)*speed, minValue, maxValue)
		str := callback(delta, ctx)

		return layout.Background{}.Layout(gtx,
			func(gtx layout.Context) layout.Dimensions {
				defer clip.UniformRRect(image.Rectangle{Max: gtx.Constraints.Min}, 0).Push(gtx.Ops).Pop()
				paint.Fill(gtx.Ops, gTheme.ContrastBg)
				ctx.drag.Add(gtx.Ops)
				return layout.Dimensions{Size: gtx.Constraints.Min}
			},
			func(gtx layout.Context) layout.Dimensions {
				colMacro := op.Record(gtx.Ops)
				paint.ColorOp{Color: gTheme.ContrastFg}.Add(gtx.Ops)
				return widget.Label{Alignment: text.Middle}.Layout(gtx, gTheme.Shaper, fnt, size, str, colMacro.Stop())
			},
		)
	}
	return ctx
}

type Drag struct {
	drag       gesture.Drag
	startPos   f32.Point
	currentPos f32.Point
}

func (d *Drag) Add(ops *op.Ops) {
	d.drag.Add(ops)
}

func (d *Drag) Update(m unit.Metric, q input.Source, axis gesture.Axis) f32.Point {
	var delta f32.Point
	for {
		e, ok := d.drag.Update(m, q, axis)
		if !ok {
			break
		}
		lastPos := d.currentPos
		switch e.Kind {
		case pointer.Press:
			d.startPos = e.Position
			d.currentPos = e.Position
		case pointer.Drag:
			d.currentPos = e.Position
			delta = delta.Add(d.currentPos.Sub(lastPos))
		}
	}
	return delta
}
