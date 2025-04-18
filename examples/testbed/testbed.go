package main

import (
	"fmt"
	"log"
	"os"

	"gioui.org/app"
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
	imgio.Init(th)

	wm := &imgio.WindowManager{}
	imgio.TempSetWm(wm)
	win_open := true
	second := true
	var col [3]float64 = [3]float64{1.0, 0.2, 0.4}
	for {
		// listen for events in the window.
		switch e := w.Event().(type) {

		// this is sent when the application should re-render.
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			imgio.SetContext(gtx)
			wm.Layout(gtx)

			imgio.Begin("debug", &win_open, func(im *imgio.Im) {
				im.Text("Hello world %v", 123)
				if im.Button("Close This") {
					fmt.Println("Saved")
					fmt.Println(inputText)
					inputText = "foo"
					*&win_open = false
				}

				im.InputText("string", &inputText)
				im.SliderFloat("float", &inputFloat, -2, 5)
				im.ColorEdit3("Color", &col)
				im.Text("%f %f %f", col[0], col[1], col[2])
			})

			imgio.Begin("second", &second, func(im *imgio.Im) {
				if im.Button("Open Other") {
					*&win_open = true
				}
			})

			e.Frame(gtx.Ops)

		// this is sent when the application is closed.
		case app.DestroyEvent:
			imgio.DestroyEvent()
			return e.Err
		}
	}
}
