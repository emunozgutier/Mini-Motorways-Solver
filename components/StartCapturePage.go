package components

import (
	"fmt"
	"image"
	"image/color"
	"sync"
	"time"

	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/emunozgutier/Mini-Motorways-Solver/store"
	"github.com/kbinani/screenshot"
)

// InteractiveCapture is a unified widget that handles screenshot display,
// rectangular selection area, and draggable corners in a single layer.
type InteractiveCapture struct {
	widget.BaseWidget
	page *StartCapturePage

	// Internal state for the selection (Normalized 0.0 - 1.0)
	px1, py1 float32
	px2, py2 float32

	activeHandle int // -1 for none, 0-3 for corners
}

func newInteractiveCapture(page *StartCapturePage) *InteractiveCapture {
	c := &InteractiveCapture{
		page:         page,
		px1:          0.1, py1: 0.1,
		px2:          0.3, py2: 0.3,
		activeHandle: -1,
	}
	c.ExtendBaseWidget(c)
	return c
}

type captureRenderer struct {
	ic *InteractiveCapture

	img      *canvas.Image
	cropRect *canvas.Rectangle
	handles  [4]*canvas.Rectangle
}

func (c *InteractiveCapture) CreateRenderer() fyne.WidgetRenderer {
	r := &captureRenderer{ic: c}
	r.img = canvas.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 800, 600)))
	r.img.FillMode = canvas.ImageFillContain

	r.cropRect = canvas.NewRectangle(color.Transparent)
	r.cropRect.StrokeColor = color.RGBA{R: 255, G: 0, B: 0, A: 255}
	r.cropRect.StrokeWidth = 2

	for i := 0; i < 4; i++ {
		r.handles[i] = canvas.NewRectangle(color.RGBA{R: 0, G: 255, B: 255, A: 255})
		r.handles[i].StrokeColor = color.Black
		r.handles[i].StrokeWidth = 1
	}

	return r
}

func (r *captureRenderer) Layout(size fyne.Size) {
	r.img.Resize(size)

	// Calculate corner positions
	hSize := float32(20)
	pos := [4]fyne.Position{
		{X: r.ic.px1 * size.Width, Y: r.ic.py1 * size.Height},
		{X: r.ic.px2 * size.Width, Y: r.ic.py1 * size.Height},
		{X: r.ic.px1 * size.Width, Y: r.ic.py2 * size.Height},
		{X: r.ic.px2 * size.Width, Y: r.ic.py2 * size.Height},
	}

	for i := 0; i < 4; i++ {
		r.handles[i].Move(pos[i].Subtract(fyne.NewPos(hSize/2, hSize/2)))
		r.handles[i].Resize(fyne.NewSize(hSize, hSize))
	}

	minX := minF(pos[0].X, pos[1].X, pos[2].X, pos[3].X)
	minY := minF(pos[0].Y, pos[1].Y, pos[2].Y, pos[3].Y)
	maxX := maxF(pos[0].X, pos[1].X, pos[2].X, pos[3].X)
	maxY := maxF(pos[0].Y, pos[1].Y, pos[2].Y, pos[3].Y)

	r.cropRect.Move(fyne.NewPos(minX, minY))
	r.cropRect.Resize(fyne.NewSize(maxX-minX, maxY-minY))
}

func (r *captureRenderer) MinSize() fyne.Size {
	return fyne.NewSize(100, 100)
}

func (r *captureRenderer) Refresh() {
	store.Capture.RLock()
	if store.Capture.BaseImage != nil {
		r.img.Image = store.Capture.BaseImage
	}
	store.Capture.RUnlock()
	r.img.Refresh()
	canvas.Refresh(r.ic)
}

func (r *captureRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.img, r.cropRect, r.handles[0], r.handles[1], r.handles[2], r.handles[3]}
}

func (r *captureRenderer) Destroy() {}

func (c *InteractiveCapture) Dragged(e *fyne.DragEvent) {
	size := c.Size()
	if size.Width <= 0 || size.Height <= 0 {
		return
	}

	// On first drag, find the best handle
	if c.activeHandle == -1 {
		dist := func(x, y float32) float32 {
			dx := x - e.PointEvent.Position.X
			dy := y - e.PointEvent.Position.Y
			return dx*dx + dy*dy
		}
		
		dists := []float32{
			dist(c.px1*size.Width, c.py1*size.Height),
			dist(c.px2*size.Width, c.py1*size.Height),
			dist(c.px1*size.Width, c.py2*size.Height),
			dist(c.px2*size.Width, c.py2*size.Height),
		}

		closest := 0
		for i := 1; i < 4; i++ {
			if dists[i] < dists[closest] {
				closest = i
			}
		}
		
		// Only activate if within range (say 40 pixels)
		if dists[closest] < 1600 {
			c.activeHandle = closest
		}
	}

	if c.activeHandle != -1 {
		dx := e.Dragged.DX / size.Width
		dy := e.Dragged.DY / size.Height

		clamp := func(v float32) float32 {
			if v < 0 { return 0 }
			if v > 1 { return 1 }
			return v
		}

		switch c.activeHandle {
		case 0: // TL
			c.px1 = clamp(c.px1 + dx)
			c.py1 = clamp(c.py1 + dy)
		case 1: // TR
			c.px2 = clamp(c.px2 + dx)
			c.py1 = clamp(c.py1 + dy)
		case 2: // BL
			c.px1 = clamp(c.px1 + dx)
			c.py2 = clamp(c.py2 + dy)
		case 3: // BR
			c.px2 = clamp(c.px2 + dx)
			c.py2 = clamp(c.py2 + dy)
		}
		c.Refresh()
	}
}

func (c *InteractiveCapture) DragEnd() {
	c.activeHandle = -1
}

/*
func (c *InteractiveCapture) Cursor() fyne.Cursor {
	return fyne.CursorCrosshair
}
*/

// StartCapturePage handles the screen capture setup page
type StartCapturePage struct {
	mu sync.Mutex // For UI specific locks if needed

	DisplaySelect *widget.Select
	ModeRadio     *widget.RadioGroup

	anim       *fyne.Animation
	stopTicker chan struct{}

	CaptureView *InteractiveCapture
}

func CreateStartCapturePage(w fyne.Window) *fyne.Container {
	page := &StartCapturePage{
		stopTicker: make(chan struct{}),
	}

	page.CaptureView = newInteractiveCapture(page)

	numDisplays := screenshot.NumActiveDisplays()
	var displayOptions []string
	for i := 0; i < numDisplays; i++ {
		bounds := screenshot.GetDisplayBounds(i)
		displayOptions = append(displayOptions, fmt.Sprintf("Display %d (%dx%d)", i+1, bounds.Dx(), bounds.Dy()))
	}

	page.DisplaySelect = widget.NewSelect(displayOptions, func(s string) {
		page.updateBaseImage()
	})

	page.ModeRadio = widget.NewRadioGroup([]string{"Fullscreen", "Windowed"}, func(s string) {
		store.Capture.Lock()
		if s == "Fullscreen" {
			store.Capture.IsFullscreen = true
		} else {
			store.Capture.IsFullscreen = false
		}
		store.Capture.Unlock()
	})
	page.ModeRadio.Horizontal = true
	// We'll set the initial mode selection after creation or based on store
	page.ModeRadio.SetSelected("Fullscreen")

	startBtn := widget.NewButton("Start Analysis", func() {
		x1, y1, x2, y2 := page.GetCropPixels()
		fmt.Printf("Capturing region: (%d, %d) to (%d, %d)\n", x1, y1, x2, y2)
	})

	form := container.NewVBox(
		container.NewHBox(widget.NewLabel("Select Display:"), page.DisplaySelect),
		page.ModeRadio,
	)

	if len(displayOptions) > 0 {
		page.DisplaySelect.SetSelectedIndex(0)
	}

	// Use Animation instead of Ticker to ensure updates happen on Main Thread (safe for macOS)
	// Animation duration is irrelevant for our use case since we use the tick function.
	page.stopTicker = make(chan struct{})
	lastCapture := time.Now()

	page.anim = fyne.NewAnimation(time.Hour, func(f float32) {
		interval := store.Capture.GetInterval()
		if time.Since(lastCapture) >= interval {
			page.updateBaseImage()
			page.CaptureView.Refresh()
			lastCapture = time.Now()
		}
	})
	page.anim.RepeatCount = fyne.AnimationRepeatForever
	page.anim.Start()

	return container.NewBorder(
		form,
		startBtn,
		nil, nil,
		page.CaptureView,
	)
}

func (p *StartCapturePage) GetCropPixels() (int, int, int, int) {
	p.mu.Lock()
	baseImg := store.Capture.BaseImage
	p.mu.Unlock()

	if baseImg == nil {
		return 0, 0, 0, 0
	}

	size := p.CaptureView.Size()
	x1, y1 := p.pixelAt(fyne.NewPos(p.CaptureView.px1*size.Width, p.CaptureView.py1*size.Height))
	x2, y2 := p.pixelAt(fyne.NewPos(p.CaptureView.px2*size.Width, p.CaptureView.py2*size.Height))

	// Update store with final pixels
	store.Capture.SetCrop(x1, y1, x2, y2)

	return x1, y1, x2, y2
}

func (p *StartCapturePage) pixelAt(pos fyne.Position) (int, int) {
	store.Capture.RLock()
	baseImg := store.Capture.BaseImage
	store.Capture.RUnlock()
	if baseImg == nil {
		return 0, 0
	}

	bounds := baseImg.Bounds()
	imgW, imgH := float32(bounds.Dx()), float32(bounds.Dy())
	viewW, viewH := p.CaptureView.Size().Width, p.CaptureView.Size().Height

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

func (p *StartCapturePage) updateBaseImage() {
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
		// Update store safely
		store.Capture.Lock()
		store.Capture.BaseImage = img
		store.Capture.DisplayIndex = idx
		store.Capture.Unlock()
	}
}

func minF(vals ...float32) float32 {
	m := vals[0]
	for _, v := range vals {
		if v < m { m = v }
	}
	return m
}

func maxF(vals ...float32) float32 {
	m := vals[0]
	for _, v := range vals {
		if v > m { m = v }
	}
	return m
}
