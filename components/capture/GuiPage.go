package capture

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/kbinani/screenshot"
)

// GuiPage handles the screen capture setup page
type GuiPage struct {
	BaseImage        *image.RGBA
	PreviewContainer *canvas.Image

	DisplaySelect    *widget.Select
	ModeRadio        *widget.RadioGroup
	XEntry, YEntry   *widget.Entry
	WEntry, HEntry   *widget.Entry

	ControlsContainer *fyne.Container
}

func CreateGuiPage(w fyne.Window) *fyne.Container {
	page := &GuiPage{}

	// Screen preview
	page.PreviewContainer = canvas.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 800, 600)))
	page.PreviewContainer.FillMode = canvas.ImageFillContain
	page.PreviewContainer.SetMinSize(fyne.NewSize(600, 350))

	// Get displays
	numDisplays := screenshot.NumActiveDisplays()
	var displayOptions []string
	for i := 0; i < numDisplays; i++ {
		bounds := screenshot.GetDisplayBounds(i)
		displayOptions = append(displayOptions, fmt.Sprintf("Display %d (%dx%d)", i+1, bounds.Dx(), bounds.Dy()))
	}

	page.DisplaySelect = widget.NewSelect(displayOptions, func(s string) {
		page.updateBaseImage()
		page.updatePreview()
	})
	if len(displayOptions) > 0 {
		page.DisplaySelect.SetSelectedIndex(0)
	}

	page.XEntry = widget.NewEntry()
	page.XEntry.SetText("0")
	page.YEntry = widget.NewEntry()
	page.YEntry.SetText("0")
	page.WEntry = widget.NewEntry()
	page.WEntry.SetText("800")
	page.HEntry = widget.NewEntry()
	page.HEntry.SetText("600")

	// Trigger preview update on entry change
	updateFunc := func(_ string) { page.updatePreview() }
	page.XEntry.OnChanged = updateFunc
	page.YEntry.OnChanged = updateFunc
	page.WEntry.OnChanged = updateFunc
	page.HEntry.OnChanged = updateFunc

	page.ModeRadio = widget.NewRadioGroup([]string{"Fullscreen", "Windowed"}, func(s string) {
		if s == "Fullscreen" {
			page.ControlsContainer.Hide()
		} else {
			page.ControlsContainer.Show()
		}
		page.updatePreview()
	})
	page.ModeRadio.Horizontal = true
	page.ModeRadio.SetSelected("Fullscreen")

	page.ControlsContainer = container.NewGridWithColumns(4,
		widget.NewLabel("X:"), page.XEntry,
		widget.NewLabel("Y:"), page.YEntry,
		widget.NewLabel("Width:"), page.WEntry,
		widget.NewLabel("Height:"), page.HEntry,
	)
	page.ControlsContainer.Hide() // hide by default

	startBtn := widget.NewButton("Start Analysis", func() {
		// Placeholder for starting actual analysis of the game state
		fmt.Println("Start capturing...")
	})

	form := container.NewVBox(
		container.NewHBox(widget.NewLabel("Select Display:"), page.DisplaySelect),
		page.ModeRadio,
		page.ControlsContainer,
	)

	return container.NewBorder(
		form,
		startBtn,
		nil, nil,
		page.PreviewContainer,
	)
}

func (p *GuiPage) updateBaseImage() {
	idx := p.DisplaySelect.SelectedIndex()
	if idx < 0 || idx >= screenshot.NumActiveDisplays() {
		return
	}
	bounds := screenshot.GetDisplayBounds(idx)
	img, err := screenshot.CaptureRect(bounds)
	if err == nil {
		p.BaseImage = img
	}
}

func (p *GuiPage) updatePreview() {
	if p.BaseImage == nil {
		return
	}

	bounds := p.BaseImage.Bounds()
	imgCopy := image.NewRGBA(bounds)
	draw.Draw(imgCopy, bounds, p.BaseImage, bounds.Min, draw.Src)

	if p.ModeRadio.Selected == "Windowed" {
		x, _ := strconv.Atoi(p.XEntry.Text)
		y, _ := strconv.Atoi(p.YEntry.Text)
		w, _ := strconv.Atoi(p.WEntry.Text)
		h, _ := strconv.Atoi(p.HEntry.Text)

		drawRedBox(imgCopy, x, y, w, h, 4)
	}

	p.PreviewContainer.Image = imgCopy
	p.PreviewContainer.Refresh()
}

func drawRedBox(img *image.RGBA, x, y, w, h, thickness int) {
	red := color.RGBA{R: 255, G: 0, B: 0, A: 255}

	// Make sure bounds are within image to prevent panic
	bounds := img.Bounds()
	
	// Top and bottom lines
	for i := x; i < x+w; i++ {
		for t := 0; t < thickness; t++ {
			if i >= 0 && i < bounds.Dx() {
				if y+t >= 0 && y+t < bounds.Dy() {
					img.Set(i, y+t, red)
				}
				if y+h-1-t >= 0 && y+h-1-t < bounds.Dy() {
					img.Set(i, y+h-1-t, red)
				}
			}
		}
	}

	// Left and right lines
	for j := y; j < y+h; j++ {
		for t := 0; t < thickness; t++ {
			if j >= 0 && j < bounds.Dy() {
				if x+t >= 0 && x+t < bounds.Dx() {
					img.Set(x+t, j, red)
				}
				if x+w-1-t >= 0 && x+w-1-t < bounds.Dx() {
					img.Set(x+w-1-t, j, red)
				}
			}
		}
	}
}
