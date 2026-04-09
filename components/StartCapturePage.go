package components

import (
	"fmt"
	"image"
	"image/color"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/emunozgutier/Mini-Motorways-Solver/store"
	"github.com/kbinani/screenshot"
)

// handle is a small cyan square that can be dragged
type handle struct {
	widget.BaseWidget
	onDragged func(fyne.Position)
}

func newHandle(onDragged func(fyne.Position)) *handle {
	h := &handle{onDragged: onDragged}
	h.ExtendBaseWidget(h)
	return h
}

func (h *handle) CreateRenderer() fyne.WidgetRenderer {
	rect := canvas.NewRectangle(color.RGBA{R: 0, G: 255, B: 255, A: 255})
	rect.StrokeColor = color.Black
	rect.StrokeWidth = 1
	return widget.NewSimpleRenderer(rect)
}

func (h *handle) MinSize() fyne.Size {
	return fyne.NewSize(20, 20)
}

func (h *handle) Dragged(e *fyne.DragEvent) {
	h.Move(h.Position().Add(e.Dragged))
	if h.onDragged != nil {
		h.onDragged(h.Position())
	}
	h.Refresh()
}

func (h *handle) DragEnd() {}

// StartCapturePage handles the screen capture setup page
type StartCapturePage struct {
	mu sync.Mutex // For UI specific locks if needed

	DisplaySelect *widget.Select
	ModeRadio     *widget.RadioGroup

	PreviewImg *canvas.Image
	CropRect   *canvas.Rectangle
	Handles    [4]*handle // TL, TR, BL, BR

	anim       *fyne.Animation
	stopTicker chan struct{}
}

func CreateStartCapturePage(w fyne.Window) *fyne.Container {
	page := &StartCapturePage{
		stopTicker: make(chan struct{}),
	}

	page.PreviewImg = canvas.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 800, 600)))
	page.PreviewImg.FillMode = canvas.ImageFillContain

	// Red rectangle for crop area
	page.CropRect = canvas.NewRectangle(color.Transparent)
	page.CropRect.StrokeColor = color.RGBA{R: 255, G: 0, B: 0, A: 255}
	page.CropRect.StrokeWidth = 2

	// Setup handles
	updateCrop := func(_ fyne.Position) {
		page.syncCropRect()
	}
	for i := 0; i < 4; i++ {
		page.Handles[i] = newHandle(updateCrop)
		page.Handles[i].Resize(page.Handles[i].MinSize())
	}

	// Default positions from store
	store.Capture.RLock()
	h0 := fyne.NewPos(float32(store.Capture.CropX1), float32(store.Capture.CropY1))
	h1 := fyne.NewPos(float32(store.Capture.CropX2), float32(store.Capture.CropY1))
	h2 := fyne.NewPos(float32(store.Capture.CropX1), float32(store.Capture.CropY2))
	h3 := fyne.NewPos(float32(store.Capture.CropX2), float32(store.Capture.CropY2))
	store.Capture.RUnlock()

	page.Handles[0].Move(h0)
	page.Handles[1].Move(h1)
	page.Handles[2].Move(h2)
	page.Handles[3].Move(h3)

	overlay := container.NewMax(
		container.NewWithoutLayout(
			page.CropRect,
			page.Handles[0], page.Handles[1], page.Handles[2], page.Handles[3],
		),
	)

	page.syncCropRect() // Initialize rect position

	// Max image and overlay together
	previewStack := container.NewMax(page.PreviewImg, overlay)

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
			overlay.Hide()
		} else {
			store.Capture.IsFullscreen = false
			overlay.Show()
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
			page.PreviewImg.Refresh()
			page.CropRect.Refresh()
			lastCapture = time.Now()
		}
	})
	page.anim.RepeatCount = fyne.AnimationRepeatForever
	page.anim.Start()

	return container.NewBorder(
		form,
		startBtn,
		nil, nil,
		previewStack,
	)
}

func (p *StartCapturePage) syncCropRect() {
	tl, tr, bl, br := p.Handles[0].Position(), p.Handles[1].Position(), p.Handles[2].Position(), p.Handles[3].Position()

	minX := minF(tl.X, tr.X, bl.X, br.X)
	minY := minF(tl.Y, tr.Y, bl.Y, br.Y)
	maxX := maxF(tl.X, tr.X, bl.X, br.X)
	maxY := maxF(tl.Y, tr.Y, bl.Y, br.Y)

	p.CropRect.Move(fyne.NewPos(minX, minY))
	p.CropRect.Resize(fyne.NewSize(maxX-minX, maxY-minY))
	p.CropRect.Refresh()

	// Update store coordinates (these are UI coordinates, GetCropPixels handles mapping to real pixels)
	// Actually, let's keep GetCropPixels as the source of truth for "real" pixels.
}

func (p *StartCapturePage) GetCropPixels() (int, int, int, int) {
	store.Capture.RLock()
	baseImg := store.Capture.BaseImage
	store.Capture.RUnlock()

	if baseImg == nil {
		return 0, 0, 0, 0
	}

	bounds := baseImg.Bounds()
	imgW, imgH := float32(bounds.Dx()), float32(bounds.Dy())
	viewW, viewH := p.PreviewImg.Size().Width, p.PreviewImg.Size().Height

	if viewW == 0 || viewH == 0 {
		return 0, 0, 0, 0
	}

	scale := viewW / imgW
	if viewH/imgH < scale {
		scale = viewH / imgH
	}

	drawW := imgW * scale
	drawH := imgH * scale
	offsetX := (viewW - drawW) / 2
	offsetY := (viewH - drawH) / 2

	pixel := func(pos fyne.Position) (int, int) {
		px := int((pos.X - offsetX) / scale)
		py := int((pos.Y - offsetY) / scale)
		if px < 0 { px = 0 }
		if px > bounds.Dx() { px = bounds.Dx() }
		if py < 0 { py = 0 }
		if py > bounds.Dy() { py = bounds.Dy() }
		return px, py
	}

	x1, y1 := pixel(p.CropRect.Position())
	x2, y2 := pixel(p.CropRect.Position().Add(fyne.NewPos(p.CropRect.Size().Width, p.CropRect.Size().Height)))
	
	// Update store with final pixels
	store.Capture.SetCrop(x1, y1, x2, y2)
	
	return x1, y1, x2, y2
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
		
		p.PreviewImg.Image = img
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
