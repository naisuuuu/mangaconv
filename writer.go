package mangaconv

import (
	"archive/zip"
	"fmt"
	"image"
	"image/jpeg"

	// for image decoding.
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func writeZip(path string, pages <-chan page) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := zip.NewWriter(f)
	defer w.Close()
	for p := range pages {
		f, err := w.Create(jpgFname(p.Name))
		if err != nil {
			return err
		}
		err = saveImg(f, p.Image)
		if err != nil {
			return err
		}
	}
	return nil
}

func jpgFname(n string) string {
	return fmt.Sprintf("%s.jpg", strings.TrimSuffix(filepath.Base(n), filepath.Ext(n)))
}

func saveImg(target io.Writer, img image.Image) error {
	if err := jpeg.Encode(target, img, &jpeg.Options{Quality: 75}); err != nil {
		return fmt.Errorf("cannot encode: %w", err)
	}
	return nil
}
