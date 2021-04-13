package imgutil_test

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"os"
	"strings"
)

func mustReadImg(path string) image.Image {
	f, err := os.Open(path)
	if err != nil {
		panic(fmt.Sprintf("cannot open %s: %s", path, err))
	}
	defer f.Close()
	i, _, err := image.Decode(f)
	if err != nil {
		panic(fmt.Sprintf("cannot decode %s: %s", path, err))
	}
	return i
}

func writeImg(path string, i image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("cannot create: %v", err)
	}
	defer f.Close()
	if err := png.Encode(f, i); err != nil {
		return fmt.Errorf("cannot encode: %v", err)
	}
	return nil
}

func mustBeGray(i image.Image) *image.Gray {
	v, ok := i.(*image.Gray)
	if !ok {
		panic("image is not grayscale")
	}
	return v
}

// isImageType checks whether image is of type t. This is kind of a hack, but prevents easy to
// overlook testing errors. t is case sensitive.
func isImageType(img image.Image, t string) bool {
	c := color.RGBA{}
	return strings.HasSuffix(fmt.Sprintf("%T", img.ColorModel().Convert(c)), t)
}

func isWithinDeltaDiff(a, b []uint8, delta uint) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if abs(int(a[i])-int(b[i])) > delta {
			return false
		}
	}
	return true
}

func abs(i int) uint {
	if i < 0 {
		return uint(-i)
	}
	return uint(i)
}

func median(s []uint8) uint8 {
	total := 0
	for i := 0; i < len(s); i++ {
		total += int(s[i])
	}
	return uint8(total / len(s))
}

func cloneSlice(b []uint8) []uint8 {
	c := make([]uint8, len(b))
	copy(c, b)
	return c
}

func cloneImg(src image.Image) image.Image {
	switch s := src.(type) {
	case *image.Gray:
		clone := *s
		clone.Pix = cloneSlice(s.Pix)
		return &clone
	case *image.NRGBA:
		clone := *s
		clone.Pix = cloneSlice(s.Pix)
		return &clone
	case *image.NRGBA64:
		clone := *s
		clone.Pix = cloneSlice(s.Pix)
		return &clone
	case *image.RGBA:
		clone := *s
		clone.Pix = cloneSlice(s.Pix)
		return &clone
	case *image.RGBA64:
		clone := *s
		clone.Pix = cloneSlice(s.Pix)
		return &clone
	case *image.YCbCr:
		clone := *s
		clone.Y = cloneSlice(s.Y)
		clone.Cb = cloneSlice(s.Cb)
		clone.Cr = cloneSlice(s.Cr)
		return &clone
	}
	return nil
}

func cloneGray(s *image.Gray) *image.Gray {
	c := *s
	c.Pix = cloneSlice(s.Pix)
	return &c
}
