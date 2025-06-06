package main

import (
	"fmt"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/bradbev/imgio/src/imgio"
)

func main() {
	go func() {
		// create new window
		w := new(app.Window)
		w.Option(app.Title("Testbed"))
		w.Option(app.Size(unit.Dp(1200), unit.Dp(800)))
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

	var (
		inputText string
	)

	/*
		var inputFloat2 float64
		var b1 widget.Clickable
		var b2 widget.Clickable
		var f widget.Float
	*/

	imgio.Init(w)

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
					im.Button("C")
					im.Button("D")
				})
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
				//im.SliderFloat("float", &inputFloat, -2, 5)
				im.SameLine()
				im.Button("E")
				im.ColorEdit("Bg", &imgio.GetTheme().Bg)
				im.ColorEdit("ContrastBg", &imgio.GetTheme().ContrastBg)
				im.ColorEdit("Fg", &imgio.GetTheme().Fg)
				im.ColorEdit("ContrastFg", &imgio.GetTheme().ContrastFg)
				im.Text("Test text")
			})
			imgio.ThemeEdit(&win_open)

			e.Frame(gtx.Ops)

		// this is sent when the application is closed.
		case app.DestroyEvent:
			imgio.DestroyEvent()
			return e.Err
		}
	}
}
