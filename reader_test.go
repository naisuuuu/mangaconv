package mangaconv

import (
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/sync/errgroup"
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

func readHelper(path string) ([]page, error) {
	read, err := selectReader(path)
	if err != nil {
		return nil, err
	}

	errg, ctx := errgroup.WithContext(context.Background())
	pages := make(chan page, 100)
	errg.Go(func() error {
		defer close(pages)
		return read(ctx, pages, path)
	})

	// since pages channel is buffered, this will run as soon as the processing is done.
	if err := errg.Wait(); err != nil {
		return nil, err
	}

	// at this point pages channel is already closed and safe to operate on synchronously.
	out := make([]page, len(pages))
	for p := range pages {
		out[p.Index] = p
	}
	return out, nil
}

func TestReader(t *testing.T) {
	tests := []struct {
		name string
		path string
		want []page
		err  error
	}{
		{
			name: "directory reader",
			path: "testdata/",
			want: []page{
				{mustReadImg("testdata/wikipe-tan-0.png"), 0},
				{mustReadImg("testdata/wikipe-tan-1.png"), 1},
			},
		},
		{
			name: "zip reader",
			path: "testdata/wikipe-tan.zip",
			want: []page{
				{mustReadImg("testdata/wikipe-tan-0.png"), 0},
				{mustReadImg("testdata/wikipe-tan-1.png"), 1},
			},
		},
		{
			name: "file without extension",
			path: "testdata/file",
			err:  ErrUnsupportedFormat,
		},
		{
			name: "unsupported file format",
			path: "testdata/file.unsupported",
			err:  ErrUnsupportedFormat,
		},
		{
			name: "nonexistant file",
			path: "testdata/nothinghere",
			err:  ErrCannotReadPath,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readHelper(tt.path)
			if err != tt.err {
				t.Errorf("reader error %v, want %v", got, tt.want)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("reader mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
