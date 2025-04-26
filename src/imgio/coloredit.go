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
	r int64
	g int64
	b int64
	a int64
	w layout.Widget
}

// returns true if the values changed
func (im *Im) ColorEdit(label string, col *color.NRGBA) bool {
	label, id := getId(label, "coloredit")
	c := fromCache(im, id, func() *colorEditContext {
		ret := &colorEditContext{
			r: int64(col.R),
			g: int64(col.G),
			b: int64(col.B),
			a: int64(col.A),
		}
		ret.w = func(gtx layout.Context) layout.Dimensions {
			h := LineHeight(gtx)
			gtx.Constraints.Min.Y = h
			ops := im.gtx.Ops
			defer clip.Rect(image.Rect(0, 0, h, h)).Push(ops).Pop()
			r, g, b, a := uint8(ret.r), uint8(ret.g), uint8(ret.b), uint8(ret.a)
			*col = color.NRGBA{R: r, G: g, B: b, A: a}
			paint.ColorOp{Color: *col}.Add(ops)
			paint.PaintOp{}.Add(ops)
			return layout.Dimensions{Size: image.Pt(h, h)}
		}
		return ret
	})

	im.WithSameLine(func(im *Im) {
		im.WithFlexMode(FlexModeRigid, func(im *Im) {
			im.WithMinConstraints(layout.Constraints{Min: image.Pt(120, LineHeight(im.gtx))}, func(im *Im) {
				im.DragInt("G##"+label, &c.g, 1, 0, 255, "G:%d")
				im.DragInt("B##"+label, &c.b, 1, 0, 255, "B:%d")
				im.DragInt("R##"+label, &c.r, 1, 0, 255, "R:%d")
				im.DragInt("A##"+label, &c.a, 1, 0, 255, "A:%d")
				im.AddWidget(c.w)
			})
		})
		im.Text(label)
	})

	return true
}
