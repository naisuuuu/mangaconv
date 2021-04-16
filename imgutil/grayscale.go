package imgutil

import (
	"image"
	"image/draw"
)

// Grayscale returns an image converted to grayscale.
// It always returns a copy.
func Grayscale(img image.Image) *image.Gray {
	b := img.Bounds()
	dst := image.NewGray(image.Rect(0, 0, b.Dx(), b.Dy()))
	grayscale(dst, img)
	return dst
}

func grayscale(dst *image.Gray, src image.Image) {
	switch i := src.(type) {
	case *image.Gray:
		clone(dst, i)
	case *image.RGBA:
		rgbaToGray(dst, i)
	case *image.RGBA64:
		rgba64ToGray(dst, i)
	case *image.NRGBA:
		nrgbaToGray(dst, i)
	case *image.NRGBA64:
		nrgba64ToGray(dst, i)
	case *image.YCbCr:
		ycbcrToGray(dst, i)
	default:
		drawGray(dst, src)
	}
}

// drawGray uses draw.Draw as a slow conversion fallback for unsupported image types.
func drawGray(dst *image.Gray, src image.Image) {
	draw.Draw(dst, dst.Bounds(), src, src.Bounds().Min, draw.Src)
}

// clone returns a copy of a grayscale image. It additionally corrects negative bounds and removes
// dangling bytes from the underlaying pixel slice, if any are present.
func clone(dst, src *image.Gray) {
	if dst.Stride == src.Stride {
		// no need to correct stride, simply copy pixels.
		copy(dst.Pix, src.Pix)
		return
	}
	// need to correct stride.
	for i := 0; i < src.Rect.Dy(); i++ {
		dstH := i * dst.Stride
		srcH := i * src.Stride
		copy(dst.Pix[dstH:dstH+dst.Stride], src.Pix[srcH:srcH+dst.Stride])
	}
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
func rgbaToGray(dst *image.Gray, src *image.RGBA) {
	concurrentIterate(src.Rect.Dy(), func(y int) {
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
}

// rgba64ToGray lossily converts a 64 bit RGBA image to grayscale.
func rgba64ToGray(dst *image.Gray, src *image.RGBA64) {
	concurrentIterate(src.Rect.Dy(), func(y int) {
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
}

// nrgbaToGray converts an NRGBA image to grayscale.
func nrgbaToGray(dst *image.Gray, src *image.NRGBA) {
	concurrentIterate(src.Rect.Dy(), func(y int) {
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
}

// nrgba64ToGray lossily converts an NRGBA64 image to grayscale.
func nrgba64ToGray(dst *image.Gray, src *image.NRGBA64) {
	concurrentIterate(src.Rect.Dy(), func(y int) {
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
}

// ycbcrToGray converts a YCbCr image to grayscale.
func ycbcrToGray(dst *image.Gray, src *image.YCbCr) {
	for i := 0; i < src.Rect.Dy(); i++ {
		dstH := i * dst.Stride
		srcH := i * src.YStride
		copy(dst.Pix[dstH:dstH+dst.Stride], src.Y[srcH:srcH+dst.Stride])
	}
}
