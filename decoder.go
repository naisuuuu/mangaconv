package mangaconv

import (
	"context"
	"fmt"
	"image"
	"io"
	"runtime"

	// This adds webp support.
	_ "golang.org/x/image/webp"
	"golang.org/x/sync/errgroup"
)

// decode reads a channel of raw pages and emits decoded pages.
func decode(ctx context.Context, pages chan<- page, raws <-chan rawPage) error {
	errg, ctx := errgroup.WithContext(ctx)
	for i := 0; i < runtime.NumCPU(); i++ {
		errg.Go(func() error {
			for raw := range raws {
				img, err := decodeImage(raw.File)
				if err != nil {
					return fmt.Errorf("cannot decode image number %d: %w", raw.Index, err)
				}
				select {
				case pages <- page{img, raw.Index}:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return nil
		})
	}
	return errg.Wait()
}

func decodeImage(f io.ReadCloser) (image.Image, error) {
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	return img, nil
}
