package mangaconv

import (
	"archive/zip"
	"fmt"
	"image"
	"image/jpeg"

	// for image decoding.
	_ "image/png"
	"io"
)

func (c *Converter) writeZip(writer io.Writer, deflate bool, pages <-chan page) error {
	method := zip.Store
	if deflate {
		method = zip.Deflate
	}

	w := zip.NewWriter(writer)
	defer w.Close()
	for p := range pages {
		f, err := w.CreateHeader(&zip.FileHeader{
			Name:   fmt.Sprintf("%09d.jpg", p.Index),
			Method: method,
		})
		if err != nil {
			return err
		}
		err = saveImg(f, p.Image)
		if v, ok := p.Image.(*image.Gray); ok {
			c.pool.Put(v)
		}
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
