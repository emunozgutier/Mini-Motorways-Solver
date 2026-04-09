package main

import (
	"image"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/vova616/screenshot"
)

func main() {
	a := app.New()
	w := a.NewWindow("Mini Motorways Solver - Screen Capture")

	// Inputs for coordinates
	xEntry := widget.NewEntry()
	xEntry.SetText("0")
	yEntry := widget.NewEntry()
	yEntry.SetText("0")
	wEntry := widget.NewEntry()
	wEntry.SetText("800")
	hEntry := widget.NewEntry()
	hEntry.SetText("600")

	// The image display
	imgContainer := canvas.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 800, 600)))
	imgContainer.FillMode = canvas.ImageFillContain
	imgContainer.SetMinSize(fyne.NewSize(400, 300))

	var capturing bool
	var ticker *time.Ticker

	startBtn := widget.NewButton("Start Capture", func() {
		if capturing {
			return
		}
		
		x, _ := strconv.Atoi(xEntry.Text)
		y, _ := strconv.Atoi(yEntry.Text)
		width, _ := strconv.Atoi(wEntry.Text)
		height, _ := strconv.Atoi(hEntry.Text)
		
		bounds := image.Rect(x, y, x+width, y+height)

		capturing = true
		ticker = time.NewTicker(100 * time.Millisecond) // ~10 FPS
		go func() {
			for range ticker.C {
				if !capturing {
					return
				}
				img, err := screenshot.CaptureRect(bounds)
				if err == nil {
					// Update the image
					imgContainer.Image = img
					imgContainer.Refresh()
				}
			}
		}()
	})

	stopBtn := widget.NewButton("Stop Capture", func() {
		capturing = false
		if ticker != nil {
			ticker.Stop()
		}
	})

	// Layout
	form := container.NewGridWithColumns(4,
		widget.NewLabel("X:"), xEntry,
		widget.NewLabel("Y:"), yEntry,
		widget.NewLabel("Width:"), wEntry,
		widget.NewLabel("Height:"), hEntry,
	)

	controls := container.NewHBox(startBtn, stopBtn)

	w.SetContent(container.NewBorder(
		container.NewVBox(form, controls),
		nil, nil, nil,
		imgContainer,
	))

	w.Resize(fyne.NewSize(600, 500))
	w.ShowAndRun()
}
