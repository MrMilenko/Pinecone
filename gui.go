package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

type GUIOptions struct {
	DataFolder   string
	JSONFilePath string
	JSONUrl      string
}

func startGUI(options GUIOptions) {
	// TODO: Show screen, do fun stuff

	a := app.New()
	w := a.NewWindow("Hello World")

	w.SetContent(widget.NewLabel("Hello World!"))
	w.ShowAndRun()
}