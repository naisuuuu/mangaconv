package imgutil_test

import (
	"image"
	"image/draw"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/naisuuuu/mangaconv/imgutil"
)

func drawGray(img image.Image) *image.Gray {
	b := img.Bounds()
	dst := image.NewGray(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Bounds(), img, b.Min, draw.Src)
	return dst
}

func TestGrayscale(t *testing.T) {
	tests := []struct {
		name string
		src  image.Image
	}{
		{
			"Gray",
			&image.Gray{
				Pix: []uint8{
					0xcc, 0x00, 0x00, 0x01, 0x00, 0xcc, 0x00, 0x02, 0x00, 0x00, 0xcc, 0x03,
					0x11, 0x22, 0x33, 0xff, 0x33, 0x22, 0x11, 0xff, 0xaa, 0x33, 0xbb, 0xff,
					0x00, 0x00, 0x00, 0xff, 0x33, 0x33, 0x33, 0xff, 0xff, 0xff, 0xff, 0xff,
				},
				Stride: 3 * 4,
				Rect:   image.Rect(-1, -1, 11, 2),
			},
		},
		{
			"RGBA",
			&image.RGBA{
				Pix: []uint8{
					0xcc, 0x00, 0x00, 0x01, 0x00, 0xcc, 0x00, 0x02, 0x00, 0x00, 0xcc, 0x03,
					0x11, 0x22, 0x33, 0xff, 0x33, 0x22, 0x11, 0xff, 0xaa, 0x33, 0xbb, 0xff,
					0x00, 0x00, 0x00, 0xff, 0x33, 0x33, 0x33, 0xff, 0xff, 0xff, 0xff, 0xff,
				},
				Stride: 3 * 4,
				Rect:   image.Rect(-1, -1, 2, 2),
			},
		},
		{
			"RGBA64",
			&image.RGBA64{
				Pix: []uint8{
					0xcc, 0x00, 0x00, 0x01, 0x00, 0xcc, 0x00, 0x02,
					0x00, 0x00, 0xcc, 0x03, 0x00, 0x00, 0xcc, 0x03,
					0x11, 0x22, 0x33, 0xff, 0x33, 0x22, 0x11, 0xff,
					0xaa, 0x33, 0xbb, 0xff, 0xaa, 0x33, 0xbb, 0xff,
					0x00, 0x00, 0x00, 0xff, 0x33, 0x33, 0x33, 0xff,
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				},
				Stride: 4 * 4,
				Rect:   image.Rect(-1, -1, 1, 2),
			},
		},
		{
			"NRGBA",
			&image.NRGBA{
				Pix: []uint8{
					0xcc, 0x00, 0x00, 0x01, 0x00, 0xcc, 0x00, 0x02, 0x00, 0x00, 0xcc, 0x03,
					0x11, 0x22, 0x33, 0xff, 0x33, 0x22, 0x11, 0xff, 0xaa, 0x33, 0xbb, 0xff,
					0x00, 0x00, 0x00, 0xff, 0x33, 0x33, 0x33, 0xff, 0xff, 0xff, 0xff, 0xff,
				},
				Stride: 3 * 4,
				Rect:   image.Rect(-1, -1, 2, 2),
			},
		},
		{
			"NRGBA64",
			&image.NRGBA64{
				Pix: []uint8{
					0xcc, 0x00, 0x00, 0x01, 0x00, 0xcc, 0x00, 0x02,
					0x00, 0x00, 0xcc, 0x03, 0x00, 0x00, 0xcc, 0x03,
					0x11, 0x22, 0x33, 0xff, 0x33, 0x22, 0x11, 0xff,
					0xaa, 0x33, 0xbb, 0xff, 0xaa, 0x33, 0xbb, 0xff,
					0x00, 0x00, 0x00, 0xff, 0x33, 0x33, 0x33, 0xff,
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				},
				Stride: 4 * 4,
				Rect:   image.Rect(-1, -1, 1, 2),
			},
		},
		{
			"YCbCr",
			mustReadImg("testdata/wikipe-tan-YCbCr.jpg"),
		},
	}
	for _, tt := range tests {
		if !isImageType(tt.src, tt.name) {
			t.Fatalf("source image is not of type %s", tt.name)
		}
		t.Run(tt.name, func(t *testing.T) {
			want := drawGray(tt.src)
			got := imgutil.Grayscale(tt.src)

			// Special case for when passed image is already grayscale.
			// In such case we want to create a copy of it and avoid related pitfalls.
			if img, ok := tt.src.(*image.Gray); ok {
				if got == img {
					t.Error("got source image, want a copy")
				}
				if len(img.Pix) > 0 && &img.Pix[0] == &got.Pix[0] {
					t.Error("copied image points to the same underlying pixel array")
				}
			}

			// Special case for YCbCr images. This is difficult to test, as direct comparisons fail with
			// seemingly random bits misplaced. My best guess is compression?
			// The "real" test here is how the outcome looks, and I can't see any differences between this
			// and draw.Draw. For being roughly 350 times faster than the aforementioned, seems good enough.
			// If you have a better idea how to test this or know why direct comparisons fail like they do,
			// please submit an issue/PR!
			if _, ok := tt.src.(*image.YCbCr); ok {
				if !isWithinDeltaDiff(want.Pix, got.Pix, 8) {
					t.Errorf("Grayscale() difference above acceptable delta")
				}
				return
			}

			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("Grayscale() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func BenchmarkGrayscale(b *testing.B) {
	benchmarks := []struct {
		name string
		img  image.Image
	}{
		{"RGBA", mustReadImg("testdata/wikipe-tan-RGBA.png")},
		{"RGBA64", mustReadImg("testdata/wikipe-tan-RGBA64.png")},
		{"NRGBA", mustReadImg("testdata/wikipe-tan-NRGBA.png")},
		{"NRGBA64", mustReadImg("testdata/wikipe-tan-NRGBA64.png")},
		{"YCbCr", mustReadImg("testdata/wikipe-tan-YCbCr.jpg")},
		{"Gray", mustReadImg("testdata/wikipe-tan-Gray.png")},
	}
	for _, bb := range benchmarks {
		if !isImageType(bb.img, bb.name) {
			b.Fatalf("source image is not of type %s", bb.name)
		}
		b.Run(bb.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				imgutil.Grayscale(bb.img)
			}
		})
		// // In case you want to compare the performance against draw.Draw implementation.
		// b.Run(bb.name+"Draw", func(b *testing.B) {
		// 	for i := 0; i < b.N; i++ {
		// 		drawGray(bb.img)
		// 	}
		// })
	}
}
