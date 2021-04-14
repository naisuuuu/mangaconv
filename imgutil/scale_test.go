package imgutil_test

import (
	"flag"
	"fmt"
	"image"
	"reflect"
	"testing"

	"github.com/naisuuuu/mangaconv/imgutil"
)

var genGoldenFiles = flag.Bool("gen_golden_files", false, "whether to generate the TestXxx golden files.")

func TestKernelScaler(t *testing.T) {
	tests := []struct {
		name  string
		w     int
		h     int
		image string
	}{
		{
			name:  "downscale",
			w:     100,
			h:     100,
			image: "wikipe-tan-100x123",
		},
		{
			name:  "upscale",
			w:     130,
			h:     150,
			image: "wikipe-tan-100x123",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := mustBeGray(mustReadImg(fmt.Sprintf("testdata/%s.png", tt.image)))
			goldenFname := fmt.Sprintf("testdata/%s-%s.png", tt.image, tt.name)

			got := image.NewGray(image.Rect(0, 0, tt.w, tt.h))
			imgutil.CatmullRom.Scale(got, src)

			if *genGoldenFiles {
				if err := writeImg(goldenFname, got); err != nil {
					t.Error(err)
					return
				}
			}

			want := mustBeGray(mustReadImg(goldenFname))
			if !reflect.DeepEqual(got, want) {
				t.Errorf("%s: actual image differs from golden image", goldenFname)
			}
		})
	}
}

func BenchmarkKernelScaler(b *testing.B) {
	src := mustBeGray(mustReadImg("testdata/wikipe-tan-Gray.png"))
	dstRect := imgutil.FitRect(src.Bounds(), 150, 150)
	scaler := imgutil.CatmullRom.NewScaler(dstRect.Dx(), dstRect.Dy(), src.Rect.Dx(), src.Rect.Dy())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dst := image.NewGray(dstRect)
		scaler.Scale(dst, src)
	}
}
