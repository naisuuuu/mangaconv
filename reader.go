package mangaconv

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"golang.org/x/sync/errgroup"
)

var (
	ErrUnsupportedFormat = errors.New("unsupported format")
	ErrCannotReadPath    = errors.New("cannot read path")
)

// reader reads provided path's contents and emits a page for each image in it.
type reader func(ctx context.Context, pages chan<- page, path string) error

// rawPage represents a page before decoding.
type rawPage struct {
	File  io.ReadCloser
	Index int
}

// selectReader returns an appropriate reader for the file format at path, or error if path cannot
// be read or the file format is not supported.
func selectReader(path string) (reader, error) {
	f, err := os.Stat(path)
	if err != nil {
		return nil, ErrCannotReadPath
	}

	switch filepath.Ext(path) {
	case "":
		if f.IsDir() {
			return readDir, nil
		}
	case ".zip", ".cbz":
		return readZip, nil
	}

	return nil, ErrUnsupportedFormat
}

// readDir reads a directory and emits a page for each image in it.
func readDir(ctx context.Context, pages chan<- page, path string) error {
	errg, ctx := errgroup.WithContext(ctx)
	raw := make(chan rawPage)
	errg.Go(func() error {
		defer close(raw)
		return readDirFiles(ctx, raw, path)
	})

	errg.Go(func() error {
		return decode(ctx, pages, raw)
	})

	return errg.Wait()
}

func readDirFiles(ctx context.Context, pages chan<- rawPage, root string) error {
	i := 0
	return filepath.WalkDir(root, func(path string, e fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("cannot walk %s: %w", root, err)
		}
		if !isImage(path) {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("cannot open %s: %w", path, err)
		}
		select {
		case pages <- rawPage{file, i}:
		case <-ctx.Done():
			// Don't forget to close any open files.
			file.Close()
			return ctx.Err()
		}
		i++
		return nil
	})
}

// readZip reads a zip file and emtis a page for each image in it.
func readZip(ctx context.Context, pages chan<- page, path string) error {
	r, err := zip.OpenReader(path)
	if err != nil {
		return fmt.Errorf("cannot open %s: %w", path, err)
	}
	defer r.Close()

	errg, ctx := errgroup.WithContext(ctx)
	raw := make(chan rawPage)
	errg.Go(func() error {
		defer close(raw)
		return readZipFiles(ctx, raw, r)
	})

	errg.Go(func() error {
		return decode(ctx, pages, raw)
	})

	return errg.Wait()
}

func readZipFiles(ctx context.Context, pages chan<- rawPage, r *zip.ReadCloser) error {
	i := 0
	for _, f := range r.File {
		if !isImage(f.Name) {
			continue
		}
		file, err := f.Open()
		if err != nil {
			return fmt.Errorf("cannot open %s: %w", f.Name, err)
		}
		select {
		case pages <- rawPage{file, i}:
		case <-ctx.Done():
			// Don't forget to close any open files.
			file.Close()
			return ctx.Err()
		}
		i++
	}
	return nil
}

func isImage(fname string) bool {
	switch filepath.Ext(fname) {
	case ".png", ".jpg", ".webp":
		return true
	default:
		return false
	}
}
