package mangaconv_test

import (
	"io"
	"testing"

	"github.com/naisuuuu/mangaconv"
)

func BenchmarkConverter(b *testing.B) {
	benchmarks := []struct {
		name string
		file string
		w, h int
	}{
		{"testdata", "imgutil/testdata", 150, 150},
	}
	for _, bb := range benchmarks {
		b.Run(bb.name, func(b *testing.B) {
			c := mangaconv.New(mangaconv.Params{
				Cutoff: 1,
				Gamma:  0.75,
				Height: bb.h,
				Width:  bb.w,
			})
			for i := 0; i < b.N; i++ {
				c.ConvertToWriter(bb.file, io.Discard)
			}
		})
	}
}
