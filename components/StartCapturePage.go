package components

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
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
	id        int
	page      *StartCapturePage
	onDragged func(int, fyne.Position)
}

func newHandle(id int, page *StartCapturePage, onDragged func(int, fyne.Position)) *handle {
	h := &handle{id: id, page: page, onDragged: onDragged}
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
	// Calculate absolute position in parent
	newPos := h.Position().Add(e.Dragged).Add(fyne.NewPos(h.Size().Width/2, h.Size().Height/2))
	if h.onDragged != nil {
		h.onDragged(h.id, newPos)
	}
	h.Refresh()
}

func (h *handle) DragEnd() {
	h.page.ZoomContainer.Hide()
}

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

	// Normalized coordinates (0.0 to 1.0) for responsive layout
	PctX1, PctY1 float32
	PctX2, PctY2 float32

	ZoomContainer *fyne.Container
	ZoomImg       *canvas.Image
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
	updateCrop := func(id int, pos fyne.Position) {
		page.syncHandles(id, pos)
	}
	for i := 0; i < 4; i++ {
		page.Handles[i] = newHandle(i, page, updateCrop)
		page.Handles[i].Resize(page.Handles[i].MinSize())
	}

	// Default positions from store (convert pixels to percentages based on first capture)
	page.PctX1, page.PctY1 = 0.1, 0.1
	page.PctX2, page.PctY2 = 0.3, 0.3

	overlay := container.New(&cropLayout{page: page},
		page.CropRect,
		page.Handles[0], page.Handles[1], page.Handles[2], page.Handles[3],
	)

	// Zoom view (magnifier)
	page.ZoomImg = canvas.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 100, 100)))
	page.ZoomImg.FillMode = canvas.ImageFillStretch
	page.ZoomContainer = container.NewStack(
		canvas.NewRectangle(color.Black), // Background
		page.ZoomImg,
		canvas.NewRectangle(color.Transparent), // Border
	)
	page.ZoomContainer.Hide()
	page.ZoomContainer.Resize(fyne.NewSize(120, 120))
	
	// Add a small red crosshair to the zoom view center
	crosshairH := canvas.NewLine(color.RGBA{R: 255, G: 0, B: 0, A: 128})
	crosshairV := canvas.NewLine(color.RGBA{R: 255, G: 0, B: 0, A: 128})
	page.ZoomContainer.Objects = append(page.ZoomContainer.Objects, crosshairH, crosshairV)


	// Max image and overlay together
	previewStack := container.NewMax(page.PreviewImg, overlay, container.NewWithoutLayout(page.ZoomContainer))

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

func (p *StartCapturePage) syncHandles(id int, pos fyne.Position) {
	size := p.PreviewImg.Size()
	if size.Width == 0 || size.Height == 0 {
		return
	}

	// Clamp pos to container bounds
	if pos.X < 0 { pos.X = 0 }
	if pos.X > size.Width { pos.X = size.Width }
	if pos.Y < 0 { pos.Y = 0 }
	if pos.Y > size.Height { pos.Y = size.Height }

	// Update our internal percentages
	switch id {
	case 0: // TL
		p.PctX1 = pos.X / size.Width
		p.PctY1 = pos.Y / size.Height
	case 1: // TR
		p.PctX2 = pos.X / size.Width
		p.PctY1 = pos.Y / size.Height
	case 2: // BL
		p.PctX1 = pos.X / size.Width
		p.PctY2 = pos.Y / size.Height
	case 3: // BR
		p.PctX2 = pos.X / size.Width
		p.PctY2 = pos.Y / size.Height
	}

	p.updateZoom(pos)
	p.syncCropRect()
}

func (p *StartCapturePage) updateZoom(pos fyne.Position) {
	p.mu.Lock()
	baseImg := store.Capture.BaseImage
	p.mu.Unlock()

	if baseImg == nil {
		return
	}

	// Map UI pos to image pixels
	x, y := p.pixelAt(pos)
	
	// Create a zoom crop (50x50 pixels around mouse)
	zoomSize := 40
	rect := image.Rect(x-zoomSize/2, y-zoomSize/2, x+zoomSize/2, y+zoomSize/2)
	
	// Ensure rect is within image
	bounds := baseImg.Bounds()
	if rect.Min.X < 0 { rect = rect.Add(image.Pt(-rect.Min.X, 0)) }
	if rect.Max.X > bounds.Max.X { rect = rect.Add(image.Pt(bounds.Max.X-rect.Max.X, 0)) }
	if rect.Min.Y < 0 { rect = rect.Add(image.Pt(0, -rect.Min.Y)) }
	if rect.Max.Y > bounds.Max.Y { rect = rect.Add(image.Pt(0, bounds.Max.Y-rect.Max.Y)) }

	zoomCrop := image.NewRGBA(image.Rect(0, 0, zoomSize, zoomSize))
	draw.Draw(zoomCrop, zoomCrop.Bounds(), baseImg, rect.Min, draw.Src)
	
	p.ZoomImg.Image = zoomCrop
	p.ZoomImg.Refresh()

	// Position ZoomContainer near mouse
	offset := float32(20)
	zoomPos := pos.Add(fyne.NewPos(offset, offset))
	// If too close to edge, move to other side
	if zoomPos.X + 120 > p.PreviewImg.Size().Width {
		zoomPos.X = pos.X - 120 - offset
	}
	if zoomPos.Y + 120 > p.PreviewImg.Size().Height {
		zoomPos.Y = pos.Y - 120 - offset
	}
	p.ZoomContainer.Move(zoomPos)
	p.ZoomContainer.Show()
}

func (p *StartCapturePage) syncCropRect() {
	// Re-layout handles based on percentages and redraw rect
	// The layout handles it, but we need to refresh the red box
	p.CropRect.Refresh()
}

func (p *StartCapturePage) GetCropPixels() (int, int, int, int) {
	p.mu.Lock()
	baseImg := store.Capture.BaseImage
	p.mu.Unlock()

	if baseImg == nil {
		return 0, 0, 0, 0
	}

	size := p.PreviewImg.Size()
	x1, y1 := p.pixelAt(fyne.NewPos(p.PctX1 * size.Width, p.PctY1 * size.Height))
	x2, y2 := p.pixelAt(fyne.NewPos(p.PctX2 * size.Width, p.PctY2 * size.Height))
	
	// Update store with final pixels
	store.Capture.SetCrop(x1, y1, x2, y2)
	
	return x1, y1, x2, y2
}

func (p *StartCapturePage) pixelAt(pos fyne.Position) (int, int) {
	store.Capture.RLock()
	baseImg := store.Capture.BaseImage
	store.Capture.RUnlock()
	if baseImg == nil { return 0, 0 }

	bounds := baseImg.Bounds()
	imgW, imgH := float32(bounds.Dx()), float32(bounds.Dy())
	viewW, viewH := p.PreviewImg.Size().Width, p.PreviewImg.Size().Height

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
	if px < 0 { px = 0 }
	if px > bounds.Dx() { px = bounds.Dx() }
	if py < 0 { py = 0 }
	if py > bounds.Dy() { py = bounds.Dy() }
	return px, py
}

type cropLayout struct {
	page *StartCapturePage
}

func (l *cropLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if size.Width <= 0 || size.Height <= 0 {
		return
	}
	p := l.page
	hSize := float32(20)

	// Position handles
	pos := [4]fyne.Position{
		{X: p.PctX1 * size.Width, Y: p.PctY1 * size.Height}, // TL
		{X: p.PctX2 * size.Width, Y: p.PctY1 * size.Height}, // TR
		{X: p.PctX1 * size.Width, Y: p.PctY2 * size.Height}, // BL
		{X: p.PctX2 * size.Width, Y: p.PctY2 * size.Height}, // BR
	}

	for i := 0; i < 4; i++ {
		// Move handle such that it is CENTERED on the corner
		p.Handles[i].Move(pos[i].Subtract(fyne.NewPos(hSize/2, hSize/2)))
	}

	// Position red box
	minX := minF(pos[0].X, pos[1].X, pos[2].X, pos[3].X)
	minY := minF(pos[0].Y, pos[1].Y, pos[2].Y, pos[3].Y)
	maxX := maxF(pos[0].X, pos[1].X, pos[2].X, pos[3].X)
	maxY := maxF(pos[0].Y, pos[1].Y, pos[2].Y, pos[3].Y)

	p.CropRect.Move(fyne.NewPos(minX, minY))
	p.CropRect.Resize(fyne.NewSize(maxX-minX, maxY-minY))
}

func (l *cropLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(100, 100)
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
