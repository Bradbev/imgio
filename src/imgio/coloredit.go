package imgio

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

// returns true if the values changed
func (im *Im) ColorEditf3(label string, col *[3]float64) bool {
	label, id := getId(label, "coloreditf3")
	w := fromCache(im, id, func() layout.Widget {
		return func(gtx layout.Context) layout.Dimensions {
			ops := im.gtx.Ops
			defer clip.Rect(image.Rect(0, 0, 40, 40)).Push(ops).Pop()
			u8 := func(f float64) uint8 { return uint8(f * 255) }
			r, g, b := u8(col[0]), u8(col[1]), u8(col[2])
			paint.ColorOp{Color: color.NRGBA{R: r, G: g, B: b, A: 0xFF}}.Add(ops)
			paint.PaintOp{}.Add(ops)
			return layout.Dimensions{Size: image.Pt(40, 40)}
		}
	})

	im.SliderFloat("R", &col[0], 0, 1)
	im.SliderFloat("G", &col[1], 0, 1)
	im.SliderFloat("B", &col[2], 0, 1)
	im.Text(label)
	im.AddWidget(w)

	return true
}

type colorEditContext struct {
	r float64
	g float64
	b float64
	a float64
	w layout.Widget
}

// returns true if the values changed
func (im *Im) ColorEdit(label string, col *color.NRGBA) bool {
	label, id := getId(label, "coloredit")
	c := fromCache(im, id, func() *colorEditContext {
		ret := &colorEditContext{
			r: float64(col.R) / 255.0,
			g: float64(col.G) / 255.0,
			b: float64(col.B) / 255.0,
			a: float64(col.A) / 255.0,
		}
		ret.w = func(gtx layout.Context) layout.Dimensions {
			ops := im.gtx.Ops
			defer clip.Rect(image.Rect(0, 0, 40, 40)).Push(ops).Pop()
			u8 := func(f float64) uint8 { return uint8(f * 255) }
			r, g, b, a := u8(ret.r), u8(ret.g), u8(ret.b), u8(ret.a)
			*col = color.NRGBA{R: r, G: g, B: b, A: a}
			paint.ColorOp{Color: *col}.Add(ops)
			paint.PaintOp{}.Add(ops)
			return layout.Dimensions{Size: image.Pt(40, 40)}
		}
		return ret
	})

	im.SliderFloat("R", &c.r, 0, 1)
	im.SliderFloat("G", &c.g, 0, 1)
	im.SliderFloat("B", &c.b, 0, 1)
	im.SliderFloat("A", &c.a, 0, 1)
	im.Text(label)
	im.AddWidget(c.w)

	return true
}
