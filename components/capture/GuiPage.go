package capture

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/kbinani/screenshot"
)

// interactiveImage is a custom widget that handles dragging to select a crop area
type interactiveImage struct {
	widget.BaseWidget
	img          *canvas.Image
	page         *GuiPage
	activeCorner int // 1:TL, 2:TR, 3:BL, 4:BR, 0:none
}

func newInteractiveImage(page *GuiPage) *interactiveImage {
	i := &interactiveImage{page: page}
	i.img = canvas.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 800, 600)))
	i.img.FillMode = canvas.ImageFillContain
	i.ExtendBaseWidget(i)
	return i
}

func (i *interactiveImage) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(i.img)
}

func (i *interactiveImage) MinSize() fyne.Size {
	return fyne.NewSize(600, 350)
}

func (i *interactiveImage) pixelFromUI(pos fyne.Position) (int, int) {
	i.page.mu.Lock()
	baseImg := i.page.BaseImage
	i.page.mu.Unlock()

	if baseImg == nil {
		return 0, 0
	}
	bounds := baseImg.Bounds()
	imgW, imgH := float32(bounds.Dx()), float32(bounds.Dy())
	viewW, viewH := i.Size().Width, i.Size().Height

	if viewW == 0 || viewH == 0 {
		return 0, 0
	}

	scale := viewW / imgW
	if viewH/imgH < scale {
		scale = viewH / imgH
	}

	drawW := imgW * scale
	drawH := imgH * scale
	offsetX := (viewW - drawW) / 2
	offsetY := (viewH - drawH) / 2

	px := int((pos.X - offsetX) / scale)
	py := int((pos.Y - offsetY) / scale)

	if px < 0 {
		px = 0
	}
	if px > bounds.Dx() {
		px = bounds.Dx()
	}
	if py < 0 {
		py = 0
	}
	if py > bounds.Dy() {
		py = bounds.Dy()
	}

	return px, py
}

func (i *interactiveImage) Dragged(e *fyne.DragEvent) {
	if i.page.ModeRadio.Selected != "Windowed" {
		return
	}

	px, py := i.pixelFromUI(e.Position)

	i.page.mu.Lock()
	if i.activeCorner == 0 {
		// Identify closest corner on first drag event
		x1, y1 := i.page.CropX1, i.page.CropY1
		x2, y2 := i.page.CropX2, i.page.CropY2

		d1 := dist(px, py, x1, y1) // TL
		d2 := dist(px, py, x2, y1) // TR
		d3 := dist(px, py, x1, y2) // BL
		d4 := dist(px, py, x2, y2) // BR

		minD := d1
		i.activeCorner = 1
		if d2 < minD {
			minD = d2
			i.activeCorner = 2
		}
		if d3 < minD {
			minD = d3
			i.activeCorner = 3
		}
		if d4 < minD {
			minD = d4
			i.activeCorner = 4
		}
	}

	switch i.activeCorner {
	case 1:
		i.page.CropX1, i.page.CropY1 = px, py
	case 2:
		i.page.CropX2, i.page.CropY1 = px, py
	case 3:
		i.page.CropX1, i.page.CropY2 = px, py
	case 4:
		i.page.CropX2, i.page.CropY2 = px, py
	}
	i.page.mu.Unlock()

	i.page.updatePreview()
}

func (i *interactiveImage) DragEnd() {
	i.activeCorner = 0
}

func dist(x1, y1, x2, y2 int) float64 {
	dx := float64(x1 - x2)
	dy := float64(y1 - y2)
	return math.Sqrt(dx*dx + dy*dy)
}

// GuiPage handles the screen capture setup page
type GuiPage struct {
	mu            sync.Mutex
	BaseImage     *image.RGBA
	PreviewWidget *interactiveImage

	DisplaySelect *widget.Select
	ModeRadio     *widget.RadioGroup

	CropX1, CropY1 int
	CropX2, CropY2 int

	ticker     *time.Ticker
	stopTicker chan struct{}
}

func CreateGuiPage(w fyne.Window) *fyne.Container {
	page := &GuiPage{
		stopTicker: make(chan struct{}),
	}

	// Default crop (center-ish)
	page.CropX1, page.CropY1 = 100, 100
	page.CropX2, page.CropY2 = 700, 500

	page.PreviewWidget = newInteractiveImage(page)

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

	page.ModeRadio = widget.NewRadioGroup([]string{"Fullscreen", "Windowed"}, func(s string) {
		page.updatePreview()
	})
	page.ModeRadio.Horizontal = true
	page.ModeRadio.SetSelected("Fullscreen")

	startBtn := widget.NewButton("Start Analysis", func() {
		fmt.Printf("Capturing region: (%d, %d) to (%d, %d)\n", page.CropX1, page.CropY1, page.CropX2, page.CropY2)
	})

	form := container.NewVBox(
		container.NewHBox(widget.NewLabel("Select Display:"), page.DisplaySelect),
		page.ModeRadio,
	)

	if len(displayOptions) > 0 {
		page.DisplaySelect.SetSelectedIndex(0)
	}

	// Start 30 FPS ticker
	page.ticker = time.NewTicker(33 * time.Millisecond)
	go func() {
		for {
			select {
			case <-page.ticker.C:
				page.updateBaseImage()
				page.updatePreview()
			case <-page.stopTicker:
				return
			}
		}
	}()

	return container.NewBorder(
		form,
		startBtn,
		nil, nil,
		page.PreviewWidget,
	)
}

func (p *GuiPage) updateBaseImage() {
	if p.DisplaySelect == nil {
		return
	}
	idx := p.DisplaySelect.SelectedIndex()
	if idx < 0 || idx >= screenshot.NumActiveDisplays() {
		return
	}
	bounds := screenshot.GetDisplayBounds(idx)
	img, err := screenshot.CaptureRect(bounds)
	if err == nil {
		p.mu.Lock()
		p.BaseImage = img
		p.mu.Unlock()
	}
}

func (p *GuiPage) updatePreview() {
	p.mu.Lock()
	baseImg := p.BaseImage
	x1, x2 := p.CropX1, p.CropX2
	y1, y2 := p.CropY1, p.CropY2
	mode := p.ModeRadio.Selected
	p.mu.Unlock()

	if baseImg == nil {
		return
	}

	bounds := baseImg.Bounds()
	imgCopy := image.NewRGBA(bounds)
	draw.Draw(imgCopy, bounds, baseImg, bounds.Min, draw.Src)

	if mode == "Windowed" {
		// Normalize coordinates for drawing
		realX1, realX2 := x1, x2
		if realX1 > realX2 {
			realX1, realX2 = realX2, realX1
		}
		realY1, realY2 := y1, y2
		if realY1 > realY2 {
			realY1, realY2 = realY2, realY1
		}

		w := realX2 - realX1
		h := realY2 - realY1

		drawRedBox(imgCopy, realX1, realY1, w, h, 4)

		// Draw handles
		handleSize := 20
		drawHandle(imgCopy, x1, y1, handleSize)
		drawHandle(imgCopy, x2, y1, handleSize)
		drawHandle(imgCopy, x1, y2, handleSize)
		drawHandle(imgCopy, x2, y2, handleSize)
	}

	p.PreviewWidget.img.Image = imgCopy
	p.PreviewWidget.Refresh()
}

func drawRedBox(img *image.RGBA, x, y, w, h, thickness int) {
	red := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	bounds := img.Bounds()

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

func drawHandle(img *image.RGBA, x, y, size int) {
	cyan := color.RGBA{R: 0, G: 255, B: 255, A: 255}
	bounds := img.Bounds()
	half := size / 2
	for i := x - half; i < x+half; i++ {
		for j := y - half; j < y+half; j++ {
			if i >= 0 && i < bounds.Dx() && j >= 0 && j < bounds.Dy() {
				img.Set(i, j, cyan)
			}
		}
	}
}
