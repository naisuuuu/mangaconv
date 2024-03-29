// package mangaconv provides utilities to convert manga and other comics for reading on an
// e-reader.
package mangaconv

import (
	"context"
	"fmt"
	"image"
	"io"
	"os"
	"runtime"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/naisuuuu/mangaconv/imgutil"
)

// Params adjust how each page of a manga is transformed. For sane defaults, see cmd/mangaconv.
//
// Cutoff is the % of brightest and darkest pixels ignored when applying histogram normalization.
// Deflate controls whether or not an image should be additionally compressed when saved to a cbz
// file.
// Gamma is the multiplier by which an image is darkened or brightened. Values > 1 brighten and
// values < 1 darken it, with 1 leaving the image as is.
// Height and Width describe a bounding box in which the output image will be fit.
type Params struct {
	Cutoff  float64
	Deflate bool
	Gamma   float64
	Height  int
	Width   int
}

// New creates a new Converter with the provided Params.
func New(p Params) *Converter {
	return &Converter{
		params: p,
		scaler: imgutil.NewCacheScaler(imgutil.CatmullRom),
		pool:   imgutil.NewImagePool(),
	}
}

// Converter converts manga for reading on an e-reader. It's safe to use concurrently.
type Converter struct {
	params Params
	scaler imgutil.Scaler
	pool   *imgutil.ImagePool
}

// Convert reads a file from in, converts it, and writes to out.
func (c *Converter) Convert(in, out string) error {
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()

	return c.ConvertToWriter(in, f)
}

// Convert reads a file from in, converts it, and writes to an io.Writer.
func (c *Converter) ConvertToWriter(in string, out io.Writer) error {
	read, err := selectReader(in)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", in, err)
	}

	errg, ctx := errgroup.WithContext(context.Background())
	pages := make(chan page)
	errg.Go(func() error {
		defer close(pages)
		return read(ctx, pages, in)
	})

	converted := make(chan page)
	errg.Go(func() error {
		defer close(converted)
		c.convert(ctx, converted, pages)
		return nil
	})

	errg.Go(func() error {
		return c.writeZip(out, c.params.Deflate, converted)
	})

	return errg.Wait()
}

// page represents a single manga page.
type page struct {
	Image image.Image
	Index int
}

// convert reads a channel of pages, applies modifications as adjusted by params and emits converted
// pages.
func (c *Converter) convert(ctx context.Context, converted chan<- page, pages <-chan page) {
	var wg sync.WaitGroup
	wg.Add(runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			defer wg.Done()
			for pg := range pages {
				src := c.pool.GetFromImage(pg.Image)
				r := imgutil.FitRect(src.Bounds(), c.params.Width, c.params.Height)
				dst := c.pool.Get(r.Dx(), r.Dy())
				c.scaler.Scale(dst, src)
				c.pool.Put(src)
				imgutil.AutoContrast(dst, c.params.Cutoff)
				imgutil.AdjustGamma(dst, c.params.Gamma)
				select {
				case converted <- page{dst, pg.Index}:
				case <-ctx.Done():
					return
				}
			}
		}()
	}
	wg.Wait()
}
