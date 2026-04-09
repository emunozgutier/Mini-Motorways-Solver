package store

import (
	"image"
	"sync"
	"time"
)

type CaptureState struct {
	sync.RWMutex
	FPS          int
	DisplayIndex int

	// Crop coordinates (pixels)
	CropX1, CropY1 int
	CropX2, CropY2 int

	IsFullscreen bool

	// Shared image data
	BaseImage *image.RGBA
}

var Capture = &CaptureState{
	FPS:          1,
	DisplayIndex: 0,
	CropX1:       100,
	CropY1:       100,
	CropX2:       700,
	CropY2:       500,
	IsFullscreen: true,
}

func (s *CaptureState) GetInterval() time.Duration {
	s.RLock()
	defer s.RUnlock()
	if s.FPS <= 0 {
		return time.Second
	}
	return time.Second / time.Duration(s.FPS)
}

func (s *CaptureState) SetCrop(x1, y1, x2, y2 int) {
	s.Lock()
	defer s.Unlock()
	s.CropX1, s.CropY1 = x1, y1
	s.CropX2, s.CropY2 = x2, y2
}
