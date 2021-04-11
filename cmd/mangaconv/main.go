package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/naisuuuu/mangaconv"
)

var (
	version = "dev"
	date    = "unknown"
)

func main() {
	cutoff := flag.Float64("cutoff", 1, `Autocontrast cutoff.
This value is the percentage of brightest and darkest pixels ignored when normalizing the histogram.
Applying a cutoff nets a more perceivable contrast improvement.`)
	gamma := flag.Float64("gamma", 0.75, `Gamma correction value.
Values < 1 darken the image, > 1 brighten it and 1 disables gamma correction.
The default will look too dark on your computer screen, but much richer than before on e-ink.`)
	height := flag.Int("height", 1920, "Maximum height of the image.")
	width := flag.Int("width", 1920, "Maximum width of the image.")
	outdir := flag.String("outdir", "", `Path to output directory.
If provided directory does not exist, mangaconv will attempt to create it. (default input dir)`)
	ver := flag.Bool("version", false, "Print version information.")

	flag.Parse()

	if *ver {
		fmt.Printf("mangaconv version %s, built at %s\n", version, date)
	}

	// Create outdir if it doesn't exist.
	if *outdir != "" {
		if err := os.MkdirAll(*outdir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Could not create outdir: %v\n", err)
			os.Exit(1)
		}
	}

	targets := make(chan target, len(flag.Args()))
	go func() {
		defer close(targets)
		for _, in := range flag.Args() {
			out := filepath.Dir(in)
			if *outdir != "" {
				out = *outdir
			}
			out = filepath.Join(out, fname(in))
			targets <- target{in, out}
		}
	}()

	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range targets {
				if err := mangaconv.Convert(t.in, t.out, mangaconv.Params{
					Cutoff: *cutoff,
					Gamma:  *gamma,
					Height: *height,
					Width:  *width,
				}); err != nil {
					fmt.Println("Failed to convert", filepath.Base(t.in), err)
					return
				}
				fmt.Println("Converted", filepath.Base(t.in))
			}
		}()
	}
	wg.Wait()
}

type target struct {
	in  string
	out string
}

func fname(in string) string {
	return strings.TrimSuffix(filepath.Base(in), filepath.Ext(in)) + ".mc.cbz"
}
