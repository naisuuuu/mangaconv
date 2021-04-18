package imgutil

import (
	"image"
	"sync"
)

// NewImagePool creates an ImagePool.
func NewImagePool() *ImagePool {
	return &ImagePool{
		cache: make(map[int]*sync.Pool),
	}
}

// ImagePool maintains a sync.Pool of pixel arrays for each image resolution gotten from it.
type ImagePool struct {
	cache map[int]*sync.Pool
	mu    sync.Mutex
}

// GetFromImage converts an image into a grayscale image with pixel slice taken from the pool.
func (p *ImagePool) GetFromImage(img image.Image) *image.Gray {
	dst := p.Get(img.Bounds().Dx(), img.Bounds().Dy())
	grayscale(dst, img)
	return dst
}

func (p *ImagePool) getPool(pixLen int) *sync.Pool {
	p.mu.Lock()
	pool, ok := p.cache[pixLen]
	if !ok {
		pool = &sync.Pool{
			New: func() interface{} {
				tmp := make([]uint8, pixLen)
				return &tmp
			},
		}
		p.cache[pixLen] = pool
	}
	p.mu.Unlock()
	return pool
}

// Get gets a grayscale image of specified width and height with pixel slice taken from the pool.
func (p *ImagePool) Get(width, height int) *image.Gray {
	tmp := p.getPool(width * height).Get().(*[]uint8)
	return &image.Gray{
		Pix:    *tmp,
		Stride: width,
		Rect:   image.Rect(0, 0, width, height),
	}
}

// Put puts an images pixel slice back into the pool.
func (p *ImagePool) Put(img *image.Gray) {
	p.getPool(len(img.Pix)).Put(&img.Pix)
}
