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
		w.Option(app.Size(unit.Dp(1200), unit.Dp(1200)))
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
	/*
		var inputFloat2 float64
		var b1 widget.Clickable
		var b2 widget.Clickable
		var f widget.Float
	*/

	imgio.Init()

	wm := &imgio.WindowManager{}
	imgio.TempSetWm(wm)
	win_open := true
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
					//win_open = false
				}
				im.InputText("string", &inputText)
				im.WithSameLine(func(im *imgio.Im) {
					im.Button("A")
					im.Button("B")
				})
				im.Button("C")
				im.SameLine()
				im.Button("D")
				im.SameLine()
				/*

					im.WithSameLine(func(im *imgio.Im) {
						//im.Text("A")
						im.Button("A")
						im.Button("B")
						im.SliderFloat("float2", &inputFloat2, 0, 1)
						im.AddWidget(material.Slider(imgio.GetTheme(), &f).Layout)
						//im.AddWidget(material.Slider(imgio.GetTheme(), &f2).Layout)
						//im.Text("B")
					})
					im.AddWidget(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{}.Layout(gtx,
							layout.Rigid(material.Button(imgio.GetTheme(), &b1, "A").Layout),
							layout.Rigid(material.Button(imgio.GetTheme(), &b2, "B").Layout),
						)
					})
				*/
				im.SliderFloat("float", &inputFloat, -2, 5)
				im.SameLine()
				im.Button("E")
				im.ColorEdit("Color", &imgio.GetTheme().Bg)

			})

			e.Frame(gtx.Ops)

		// this is sent when the application is closed.
		case app.DestroyEvent:
			imgio.DestroyEvent()
			return e.Err
		}
	}
}
