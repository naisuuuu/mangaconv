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
		f, err := w.Create(fmt.Sprintf("%d.jpg", p.Index))
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

func saveImg(target io.Writer, img image.Image) error {
	if err := jpeg.Encode(target, img, &jpeg.Options{Quality: 75}); err != nil {
		return fmt.Errorf("cannot encode: %w", err)
	}
	return nil
}
