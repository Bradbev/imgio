package main

import (
	"fmt"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/bradbev/imgio/src/imgio"
)

func main() {
	go func() {
		// create new window
		w := new(app.Window)
		w.Option(app.Title("Egg timer"))
		w.Option(app.Size(unit.Dp(600), unit.Dp(600)))
		if err := draw(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()

	app.Main()
}

type C = layout.Context
type D = layout.Dimensions

func box(gtx layout.Context) {
	//ops := gtx.Ops
	//defer op.Offset(image.Pt(100, 100)).Push(ops).Pop()
	//defer clip.Rect{Max: image.Pt(220, 220)}.Push(ops).Pop()
	//paint.ColorOp{Color: color.NRGBA{R: 0x80, A: 0xFF}}.Add(ops)
	//paint.PaintOp{}.Add(ops)
	b := new(widget.Clickable)
	layout.UniformInset(unit.Dp(80)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return material.Button(material.NewTheme(), b, "hi").Layout(gtx)
	})
}

func draw(w *app.Window) error {
	var ops op.Ops

	var inputText string
	var inputFloat float64

	th := material.NewTheme()
	th.Face = "monospace"

	im := imgio.NewIm(th)

	var button2 widget.Clickable
	wm := &imgio.WindowManager{}
	imgio.TempSetWm(wm)
	//win1 := &imgio.Window{Parent: wm, Size: f32.Pt(500, 400)}
	win2 := &imgio.Window{Parent: wm, Size: f32.Pt(200, 200), Pos: f32.Pt(200, 200)}
	win_open := true
	for {
		// listen for events in the window.
		switch e := w.Event().(type) {

		// this is sent when the application should re-render.
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			imgio.SetContext(gtx)
			im.Reset(gtx)
			wm.Layout(gtx)

			imgio.Begin(im, "debug", &win_open, func(im *imgio.Im) {
				im.Text("Hello world %v", 123)
				if im.Button("Save") {
					fmt.Println("Saved")
					fmt.Println(inputText)
					inputText = "foo"
					*&win_open = false
				}

				im.InputText("string", &inputText)
				im.SliderFloat("float", &inputFloat, -2, 5)
			})
			//im.Layout(gtx)
			//win1.Layout(gtx, im.Layout)

			if button2.Clicked(gtx) {
				*&win_open = true
			}
			/*
				win1.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return im.Layout(gtx)
						return layout.Flex{
							Axis:    layout.Vertical,
							Spacing: layout.SpaceEnd,
						}.Layout(gtx, layout.Rigid(func(gtx C) D {
							return layout.UniformInset(unit.Dp(5)).Layout(gtx, func(gtx C) D {
								return material.Button(th, &button, "First").Layout(gtx)
							})
						}))
				})
			*/
			win2.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:    layout.Vertical,
					Spacing: layout.SpaceEnd,
				}.Layout(gtx, layout.Rigid(func(gtx C) D {
					return layout.UniformInset(unit.Dp(5)).Layout(gtx, func(gtx C) D {
						return material.Button(th, &button2, "Second").Layout(gtx)
					})
				}))
			})

			e.Frame(gtx.Ops)

		// this is sent when the application is closed.
		case app.DestroyEvent:
			return e.Err
		}
	}
}
