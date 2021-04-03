package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/naisuuuu/mangaconv"
)

func main() {
	cutoff := flag.Float64("cutoff", 1, `Autocontrast cutoff.
This value is the percentage of brightest and darkest pixels ignored when normalizing the histogram.
Applying a cutoff nets a more perceivable contrast improvement.`)
	gamma := flag.Float64("gamma", 0.7, `Gamma correction value.
Values < 1 darken the image, > 1 brighten it and 1 disables gamma correction.
The default will look too dark on your computer screen, but much richer than before on e-ink.`)
	height := flag.Int("height", 1920, "Maximum height of the image.")
	width := flag.Int("width", 1920, "Maximum width of the image.")
	outdir := flag.String("outdir", "", "Path to output directory. (default input dir)")

	flag.Parse()

	for _, in := range flag.Args() {
		out := filepath.Dir(in)
		if *outdir != "" {
			out = *outdir
		}
		out = filepath.Join(out, fname(in))

		if err := mangaconv.Convert(in, out, mangaconv.Params{
			Cutoff: *cutoff,
			Gamma:  *gamma,
			Height: *height,
			Width:  *width,
		}); err != nil {
			fmt.Println("Failed to convert", filepath.Base(in), err)
			continue
		}
		fmt.Println("Converted", filepath.Base(in))
	}
}

func fname(in string) string {
	return strings.TrimSuffix(filepath.Base(in), filepath.Ext(in)) + ".mc.cbz"
}
