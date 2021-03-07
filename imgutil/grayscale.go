// Package imgutil implements a set of image processing utilities.
// No funcs in this package perform modifications in-place, copies are returned instead.
package imgutil

import (
	"image"
	"image/draw"
)

// Grayscale returns an image converted to grayscale.
func Grayscale(img image.Image) *image.Gray {
	switch i := img.(type) {
	case *image.Gray:
		return cp(i)
	case *image.RGBA:
		return rgbaToGray(i)
	case *image.RGBA64:
		return rgba64ToGray(i)
	case *image.NRGBA:
		return nrgbaToGray(i)
	case *image.NRGBA64:
		return nrgba64ToGray(i)
	case *image.YCbCr:
		return ycbcrToGray(i)
	default:
		return drawGray(img)
	}
}

// drawGray uses draw.Draw as a slow conversion fallback for unsupported image types.
func drawGray(img image.Image) *image.Gray {
	b := img.Bounds()
	dst := image.NewGray(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Bounds(), img, b.Min, draw.Src)
	return dst
}

// cp returns a copy of a grayscale image.
func cp(src *image.Gray) *image.Gray {
	b := src.Bounds()
	dst := image.NewGray(image.Rect(0, 0, b.Dx(), b.Dy()))
	copy(dst.Pix, src.Pix)
	return dst
}

// rgbToGray returns a grayscale value from alpha premultiplied red, green and blue values.
func rgbToGray(r, g, b uint32) uint8 {
	// From https://golang.org/src/image/color/color.go#L244:
	// "These coefficients (the fractions 0.299, 0.587 and 0.114) are the same
	// as those given by the JFIF specification and used by func RGBToYCbCr in
	// ycbcr.go.
	//
	// Note that 19595 + 38470 + 7471 equals 65536.
	//
	// The 24 is 16 + 8. The 16 is the same as used in RGBToYCbCr. The 8 is
	// because the return value is 8 bit color, not 16 bit color."
	t := (19595*r + 38470*g + 7471*b + 1<<15) >> 24
	return uint8(t)
}

// rgbaToGray converts an RGBA image to grayscale.
func rgbaToGray(src *image.RGBA) *image.Gray {
	b := src.Bounds()
	dst := image.NewGray(image.Rect(0, 0, b.Dx(), b.Dy()))
	concurrentIterate(b.Dy(), func(y int) {
		for x := 0; x < dst.Stride; x++ {
			i := y*src.Stride + x*4
			s := src.Pix[i : i+3 : i+3]
			var (
				r = uint32(s[0])
				g = uint32(s[1])
				b = uint32(s[2])
			)
			r |= r << 8
			g |= g << 8
			b |= b << 8
			dst.Pix[y*dst.Stride+x] = rgbToGray(r, g, b)
		}
	})
	return dst
}

// rgba64ToGray lossily converts a 64 bit RGBA image to grayscale.
func rgba64ToGray(src *image.RGBA64) *image.Gray {
	b := src.Bounds()
	dst := image.NewGray(image.Rect(0, 0, b.Dx(), b.Dy()))
	concurrentIterate(b.Dy(), func(y int) {
		for x := 0; x < dst.Stride; x++ {
			i := y*src.Stride + x*8
			s := src.Pix[i : i+6 : i+6]
			var (
				r = uint32(s[0])<<8 | uint32(s[1])
				g = uint32(s[2])<<8 | uint32(s[3])
				b = uint32(s[4])<<8 | uint32(s[5])
			)
			dst.Pix[y*dst.Stride+x] = rgbToGray(r, g, b)
		}
	})
	return dst
}

// nrgbaToGray converts an NRGBA image to grayscale.
func nrgbaToGray(src *image.NRGBA) *image.Gray {
	b := src.Bounds()
	dst := image.NewGray(image.Rect(0, 0, b.Dx(), b.Dy()))
	concurrentIterate(b.Dy(), func(y int) {
		for x := 0; x < dst.Stride; x++ {
			i := y*src.Stride + x*4
			s := src.Pix[i : i+4 : i+4]
			var (
				r = uint32(s[0])
				g = uint32(s[1])
				b = uint32(s[2])
				a = uint32(s[3])
			)
			r |= r << 8
			g |= g << 8
			b |= b << 8
			a |= a << 8
			r = r * a / 0xffff
			g = g * a / 0xffff
			b = b * a / 0xffff
			dst.Pix[y*dst.Stride+x] = rgbToGray(r, g, b)
		}
	})
	return dst
}

// nrgba64ToGray converts an NRGBA64 image to grayscale.
func nrgba64ToGray(src *image.NRGBA64) *image.Gray {
	b := src.Bounds()
	dst := image.NewGray(image.Rect(0, 0, b.Dx(), b.Dy()))
	concurrentIterate(b.Dy(), func(y int) {
		for x := 0; x < dst.Stride; x++ {
			i := y*src.Stride + x*8
			s := src.Pix[i : i+8 : i+8]
			var (
				r = uint32(s[0])<<8 | uint32(s[1])
				g = uint32(s[2])<<8 | uint32(s[3])
				b = uint32(s[4])<<8 | uint32(s[5])
				a = uint32(s[6])<<8 | uint32(s[7])
			)
			r = r * a / 0xffff
			g = g * a / 0xffff
			b = b * a / 0xffff
			dst.Pix[y*dst.Stride+x] = rgbToGray(r, g, b)
		}
	})
	return dst
}

// ycbcrToGray converts a YCbCr image to grayscale.
func ycbcrToGray(src *image.YCbCr) *image.Gray {
	b := src.Bounds()
	dst := image.NewGray(image.Rect(0, 0, b.Dx(), b.Dy()))
	for i := 0; i < b.Dy(); i++ {
		dstH := i * dst.Stride
		srcH := i * src.YStride
		copy(dst.Pix[dstH:dstH+dst.Stride], src.Y[srcH:srcH+dst.Stride])
	}
	return dst
}
