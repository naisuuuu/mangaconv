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

func TestScaler(t *testing.T) {
	tests := []struct {
		name   string
		w      int
		h      int
		image  string
		scaler imgutil.Scaler
	}{
		{
			name:   "CatmullRom-downscale",
			w:      100,
			h:      100,
			image:  "wikipe-tan-100x123",
			scaler: imgutil.CatmullRom,
		},
		{
			name:   "CatmullRom-upscale",
			w:      130,
			h:      150,
			image:  "wikipe-tan-100x123",
			scaler: imgutil.CatmullRom,
		},
		{
			name:   "CacheCatmullRom-downscale",
			w:      100,
			h:      100,
			image:  "wikipe-tan-100x123",
			scaler: imgutil.NewCacheScaler(imgutil.CatmullRom),
		},
		{
			name:   "CacheCatmullRom-upscale",
			w:      130,
			h:      150,
			image:  "wikipe-tan-100x123",
			scaler: imgutil.NewCacheScaler(imgutil.CatmullRom),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := mustBeGray(mustReadImg(fmt.Sprintf("testdata/%s.png", tt.image)))
			goldenFname := fmt.Sprintf("testdata/%s-%s.png", tt.image, tt.name)

			got := image.NewGray(image.Rect(0, 0, tt.w, tt.h))
			tt.scaler.Scale(got, src)

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

// BenchmarkScaler benchmarks scalers against one or more images. The occurences of each image are
// evenly distributed among all benchmark iterations.
func BenchmarkScaler(b *testing.B) {
	benchmarks := []struct {
		name   string
		scaler imgutil.Scaler
		images []string
	}{
		{
			name:   "CatmullRom_singleImage",
			scaler: imgutil.CatmullRom.NewScaler(122, 150, 195, 239),
			images: []string{"wikipe-tan-195x239.png"},
		},
		{
			name:   "CatmullRom_threeImages",
			scaler: imgutil.CatmullRom.NewScaler(122, 150, 195, 239),
			images: []string{"wikipe-tan-195x239.png", "wikipe-tan-100x123.png", "wikipe-tan-82x100.png"},
		},
		{
			name:   "CacheCatmullRom_singleImage",
			scaler: imgutil.NewCacheScaler(imgutil.CatmullRom),
			images: []string{"wikipe-tan-195x239.png"},
		},
		{
			name:   "CacheCatmullRom_threeImages",
			scaler: imgutil.NewCacheScaler(imgutil.CatmullRom),
			images: []string{"wikipe-tan-195x239.png", "wikipe-tan-100x123.png", "wikipe-tan-82x100.png"},
		},
	}
	for _, bb := range benchmarks {
		b.Run(bb.name, func(b *testing.B) {
			var images []*image.Gray
			for _, i := range bb.images {
				images = append(images, mustBeGray(mustReadImg("testdata/"+i)))
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				img := images[i%len(images)]
				dstRect := imgutil.FitRect(img.Bounds(), 150, 150)
				b.StartTimer()
				dst := image.NewGray(dstRect)
				bb.scaler.Scale(dst, img)
			}
		})
	}
	for _, bb := range benchmarks {
		b.Run("Pooled"+bb.name, func(b *testing.B) {
			var images []*image.Gray
			for _, i := range bb.images {
				images = append(images, mustBeGray(mustReadImg("testdata/"+i)))
			}
			pool := imgutil.NewImagePool()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				img := images[i%len(images)]
				dstRect := imgutil.FitRect(img.Bounds(), 150, 150)
				b.StartTimer()
				dst := pool.Get(dstRect.Dx(), dstRect.Dy())
				bb.scaler.Scale(dst, img)
				pool.Put(dst)
			}
		})
	}
}
