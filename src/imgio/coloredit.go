package imgio

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

// returns true if the values changed
func (im *Im) ColorEdit3(label string, col *[3]float64) bool {
	label, id := getId(label, "sliderfloat")
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
