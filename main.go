package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/emunozgutier/Mini-Motorways-Solver/components/capture"
)

func main() {
	a := app.New()
	w := a.NewWindow("Mini Motorways Solver - Screen Capture SETUP")

	guiContainer := capture.CreateGuiPage(w)
	w.SetContent(guiContainer)

	w.Resize(fyne.NewSize(800, 600))
	w.ShowAndRun()
}
