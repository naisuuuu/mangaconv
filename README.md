[![CI](https://github.com/naisuuuu/mangaconv/actions/workflows/ci.yml/badge.svg?branch=main)
](https://github.com/naisuuuu/mangaconv/actions?query=branch%3Amain)

# mangaconv

mangaconv is a portable cli tool to convert comic and manga files/folders for reading on e-ink
devices.

Currently supported input formats are zip/cbz or a folder of images. Output is a cbz archive.

This project is heavily inspired by [KCC](https://github.com/ciromattia/kcc). Unlike KCC, it does
not require any runtime dependencies and does not attempt to make any internet connections.

## Usage

Simple usage:

```sh
mangaconv path/to/my/manga.zip another/path/to/my/manga/dir
```

Configure with flags:

```sh
mangaconv -height 1080 -width 1920 path/to/my/manga.zip another/path/to/my/manga/dir
```

To learn about provided flags:

```sh
mangaconv -help
```

## TODOS

This project is still a work in progress.

Notable missing features include:

- epub input / output support
- pdf input support
- memory usage optimization

## License

MIT. See LICENSE.

Wikipe-tan image used for testing by
[Kasuga~enwiki](https://en.wikipedia.org/wiki/User:Kasuga~enwiki),
borrowed from
[wikimedia](https://commons.wikimedia.org/wiki/File:Wikipe-tan_at_Mother%27s_day.png)
under
[CC BY-SA 3.0](https://creativecommons.org/licenses/by-sa/3.0/deed.en).
