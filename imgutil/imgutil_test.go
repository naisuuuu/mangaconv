package imgutil_test

import (
	"fmt"
	"image"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/naisuuuu/mangaconv/imgutil"
)

func TestAdjustGamma(t *testing.T) {
	tests := []struct {
		src   *image.Gray
		gamma float64
		want  *image.Gray
	}{
		{
			&image.Gray{
				Rect:   image.Rect(-1, -1, 2, 2),
				Stride: 3,
				Pix: []uint8{
					0x00, 0x11, 0x22,
					0x33, 0xaa, 0xbb,
					0xcc, 0xff, 0x00,
				},
			},
			0.75,
			&image.Gray{
				Rect:   image.Rect(0, 0, 3, 3),
				Stride: 3,
				Pix: []uint8{
					0x00, 0x07, 0x11,
					0x1e, 0x95, 0xa9,
					0xbd, 0xff, 0x00,
				},
			},
		},
		{
			&image.Gray{
				Rect:   image.Rect(-1, -1, 2, 2),
				Stride: 3,
				Pix: []uint8{
					0x00, 0x11, 0x22,
					0x33, 0xaa, 0xbb,
					0xcc, 0xff, 0x00,
				},
			},
			1.5,
			&image.Gray{
				Rect:   image.Rect(0, 0, 3, 3),
				Stride: 3,
				Pix: []uint8{
					0x00, 0x2a, 0x43,
					0x57, 0xc3, 0xcf,
					0xdc, 0xff, 0x00,
				},
			},
		},
		{
			&image.Gray{
				Rect:   image.Rect(-1, -1, 2, 2),
				Stride: 3,
				Pix: []uint8{
					0x00, 0x11, 0x22,
					0x33, 0xaa, 0xbb,
					0xcc, 0xff, 0x00,
				},
			},
			1.0,
			&image.Gray{
				Rect:   image.Rect(0, 0, 3, 3),
				Stride: 3,
				Pix: []uint8{
					0x00, 0x11, 0x22,
					0x33, 0xaa, 0xbb,
					0xcc, 0xff, 0x00,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%.2f", tt.gamma), func(t *testing.T) {
			got := imgutil.AdjustGamma(tt.src, tt.gamma)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("AdjustGamma() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func BenchmarkAdjustGamma(b *testing.B) {
	img := mustBeGray(mustReadImg("testdata/wikipe-tan-Gray.png"))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		imgutil.AdjustGamma(img, 1.8)
	}
}

func TestHistogram(t *testing.T) {
	tests := []struct {
		name  string
		image *image.Gray
		want  [256]uint
	}{
		{
			name: "basic",
			image: &image.Gray{
				Rect:   image.Rect(-1, -1, 1, 1),
				Stride: 2,
				Pix: []uint8{
					0x00, 0xff,
					0x80, 0x00,
				},
			},
			want: [256]uint{0x00: 2, 0x80: 1, 0xff: 1},
		},
		{
			name:  "empty",
			image: &image.Gray{},
			want:  [256]uint{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := imgutil.Histogram(tt.image)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Histogram() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func BenchmarkHistogram(b *testing.B) {
	img := mustBeGray(mustReadImg("testdata/wikipe-tan-Gray.png"))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		imgutil.Histogram(img)
	}
}

func TestAutoContrast(t *testing.T) {
	tests := []struct {
		name   string
		image  *image.Gray
		cutoff float32
		want   uint8
	}{
		{
			name:   "1 percent cutoff",
			image:  mustBeGray(mustReadImg("testdata/wikipe-tan-Gray.png")),
			cutoff: 1,
			want:   82,
		},
		{
			name:   "0 percent cutoff",
			image:  mustBeGray(mustReadImg("testdata/wikipe-tan-Gray.png")),
			cutoff: 0,
			want:   74,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := imgutil.AutoContrast(tt.image, tt.cutoff)
			got := median(out.Pix)
			if got != tt.want {
				t.Errorf("AutoContrast() median = %d, want %d", got, tt.want)
			}
		})
	}
}

func BenchmarkAutoContrast(b *testing.B) {
	img := mustBeGray(mustReadImg("testdata/wikipe-tan-Gray.png"))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		imgutil.AutoContrast(img, 1)
	}
}
