// Package imgutil implements a set of image processing utilities.
// No funcs in this package perform modifications in-place, copies are returned instead.
package imgutil

import (
	"image"
	"math"
	"runtime"
	"sync"

	"golang.org/x/image/draw"
)

// applyLookup returns an image with specified lookup table applied.
func applyLookup(src *image.Gray, lut *[256]uint8) *image.Gray {
	b := src.Bounds()
	dst := image.NewGray(image.Rect(0, 0, b.Dx(), b.Dy()))
	concurrentIterate(b.Dy(), func(y int) {
		for x := 0; x < dst.Stride; x++ {
			// This is safe to do concurrently because we never write to the same index twice.
			//
			// Note that we use dst.Stride to interact with the destination image and
			// src.Stride for the source. This is because even though both images are
			// grayscale, sometimes the source may be encoded with extra dangling bits and
			// longer stride.
			dst.Pix[y*dst.Stride+x] = lut[src.Pix[y*src.Stride+x]]
		}
	})
	return dst
}

// AdjustGamma returns an image with gamma modifications applied.
//
// Gamma value of 1 doesn't change the image, < 1 darkens and >1 brightens it.
func AdjustGamma(img *image.Gray, gamma float64) *image.Gray {
	if gamma == 1 {
		return clone(img)
	}
	var lut [256]uint8
	for i := 0; i < 256; i++ {
		lut[i] = clamp(math.Pow(float64(i)/255, 1/gamma) * 255)
	}
	return applyLookup(img, &lut)
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

// AutoContrast returns a grayscale image with histogram normalization applied, ignoring specified
// cutoff % highest and lowest values.
//
// This implementation is taken from Pillow's ImageOps.autocontrast method. See:
// https://pillow.readthedocs.io/en/stable/_modules/PIL/ImageOps.html#autocontrast
func AutoContrast(img *image.Gray, cutoff float32) *image.Gray {
	hist := Histogram(img)

	// Cutoff % of lowest/highest samples.
	if cutoff > 0 {
		cutl := uint(float32(len(img.Pix)) * cutoff / 100)
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
		return img
	}

	// Generate lookup table.
	var lut [256]uint8
	scale := 255 / float64(hi-lo)
	offset := float64(-lo) * scale
	for i := 0; i < 256; i++ {
		lut[i] = clamp(float64(i)*scale + offset)
	}

	return applyLookup(img, &lut)
}

// Fit returns an image scaled to fit the specified bounding box without changing the aspect ratio.
func Fit(img *image.Gray, x, y int) *image.Gray {
	bounds := img.Bounds()
	if x <= 0 {
		x = bounds.Dx()
	}
	if y <= 0 {
		y = bounds.Dy()
	}
	if x == bounds.Dx() && y == bounds.Dy() {
		return clone(img)
	}
	width, height := float64(bounds.Dx()), float64(bounds.Dy())
	scale := math.Min(float64(x)/width, float64(y)/height)
	rect := image.Rect(0, 0, int(math.Round(scale*width)), int(math.Round(scale*height)))
	// This is hardly optimal, but since there are no fast paths for grayscale destination images in
	// x/image/draw, it winds up being faster to scale to RGBA dst and then convert to grayscale
	// using our own optimized implementation.
	// TODO: re-implement resampling logic to directly handle grayscale images.
	dst := image.NewRGBA(rect)
	draw.CatmullRom.Scale(dst, rect, img, img.Bounds(), draw.Over, nil)
	return Grayscale(dst)
}
