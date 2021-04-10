// Package imgutil implements a set of image processing utilities.
// Funcs in this package perform modifications in-place, except where otherwise noted.
package imgutil

import (
	"image"
	"math"
	"runtime"
	"sync"

	"golang.org/x/image/draw"
)

// applyLookup applies a lookup table to an image.
func applyLookup(img *image.Gray, lut *[256]uint8) {
	for i := 0; i < len(img.Pix); i++ {
		img.Pix[i] = lut[img.Pix[i]]
	}
}

// AdjustGamma applies gamma adjustments.
//
// Gamma value of 1 doesn't change the image, < 1 darkens and >1 brightens it.
func AdjustGamma(img *image.Gray, gamma float64) {
	if gamma == 1 {
		return
	}
	var lut [256]uint8
	for i := 0; i < 256; i++ {
		lut[i] = clamp(math.Pow(float64(i)/255, 1/gamma) * 255)
	}
	applyLookup(img, &lut)
}

// Histogram returns a histogram of a grayscale image.
//
// Resulting histogram is represented as a fixed length array of 256 unsigned integers,
// where histogram[i] is the amount of pixels of value i present in the image.
func Histogram(img *image.Gray) [256]uint {
	var hist [256]uint
	var mu sync.Mutex
	cpus := runtime.NumCPU()
	height := img.Bounds().Dy()
	m := 1
	if height > cpus {
		m = height / cpus
	}
	concurrentIterate(height/m+1, func(y int) {
		var tmp [256]uint
		for x := 0; x < img.Stride*m; x++ {
			index := y*img.Stride*m + x
			if index == len(img.Pix) {
				break
			}
			tmp[img.Pix[index]]++
		}
		mu.Lock()
		for i := 0; i < 256; i++ {
			hist[i] += tmp[i]
		}
		mu.Unlock()
	})
	return hist
}

// AutoContrast applies histogram normalization to the image, ignoring specified cutoff % highest
// and lowest values.
//
// This implementation is taken from Pillow's ImageOps.autocontrast method. See:
// https://pillow.readthedocs.io/en/stable/_modules/PIL/ImageOps.html#autocontrast
func AutoContrast(img *image.Gray, cutoff float64) {
	hist := Histogram(img)

	// Cutoff % of lowest/highest samples.
	if cutoff > 0 {
		cutl := uint(float64(len(img.Pix)) * cutoff / 100)
		cuth := cutl
		for i := 0; i < 256; i++ {
			if hist[i] >= cutl {
				hist[i] -= cutl
				break
			}
			cutl -= hist[i]
			hist[i] = 0
		}
		for i := 255; i >= 0; i-- {
			if hist[i] >= cuth {
				hist[i] -= cuth
				break
			}
			cuth -= hist[i]
			hist[i] = 0
		}
	}

	// Find lowest/highest samples.
	var hi, lo int
	for i := 0; i < 256; i++ {
		if hist[i] > 0 {
			lo = i
			break
		}
	}
	for i := 255; i >= 0; i-- {
		if hist[i] > 0 {
			hi = i
			break
		}
	}

	if hi <= lo {
		return
	}

	// Generate lookup table.
	var lut [256]uint8
	scale := 255 / float64(hi-lo)
	offset := float64(-lo) * scale
	for i := 0; i < 256; i++ {
		lut[i] = clamp(float64(i)*scale + offset)
	}

	applyLookup(img, &lut)
}

// Fit returns an image scaled to fit the specified bounding box without changing the aspect ratio.
// It returns a copy if the image needs resampling, otherwise returns the source image.
func Fit(img *image.Gray, x, y int) *image.Gray {
	bounds := img.Bounds()
	if x <= 0 {
		x = bounds.Dx()
	}
	if y <= 0 {
		y = bounds.Dy()
	}
	width, height := float64(bounds.Dx()), float64(bounds.Dy())
	scale := math.Min(float64(x)/width, float64(y)/height)
	if scale == 1 {
		return img
	}
	rect := image.Rect(0, 0, int(math.Round(scale*width)), int(math.Round(scale*height)))
	// This is hardly optimal, but since there are no fast paths for grayscale destination images in
	// x/image/draw, it winds up being faster to scale to RGBA dst and then convert to grayscale
	// using our own optimized implementation.
	// TODO: re-implement resampling logic to directly handle grayscale images.
	dst := image.NewRGBA(rect)
	draw.CatmullRom.Scale(dst, rect, img, img.Bounds(), draw.Over, nil)
	return Grayscale(dst)
}
