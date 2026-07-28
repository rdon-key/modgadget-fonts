# ModGadget Fonts

ModGadget Fonts provides a command-line converter and a small runtime package
for using BDF bitmap fonts with TinyGo and embedded applications. The converter
turns a Unicode/ISO10646 BDF font into immutable Go data containing only the
glyphs requested by a UTF-8 text file.

The current format has the following properties:

- Unicode/ISO10646 BDF input
- Glyph subsetting from UTF-8 text
- Framebuffer-independent 1-bit bitmap data
- Glyph metadata ordered by Unicode code point
- One continuous immutable `string` containing all bitmap data
- Binary-search glyph lookup at runtime
- No external dependencies

The project is experimental, and its formats and APIs may still change.

## Installation

Install the command from the module:

```sh
go install github.com/rdon-key/modgadget-fonts/cmd/modgadget-fonts@latest
```

From a repository checkout, run the command without installing it:

```sh
go run ./cmd/modgadget-fonts \
    -bdf fonts/example/example.bdf \
    -subset characters.txt \
    -package assets \
    -var HeadlineFont \
    -o assets/headline_font.go
```

## Subset file

The subset file is UTF-8 text. Each distinct Unicode code point in the file is
included once in the generated font. For example, one file can contain English,
Japanese, and Simplified Chinese text:

```text
Hello
こんにちは
新闻
```

The subset processing rules are:

- Duplicate characters are included only once.
- Carriage returns, line feeds, and tabs are ignored.
- An ASCII space is a normal character and is included when present.
- A UTF-8 BOM is ignored only at the beginning of the file.
- Unicode normalization is not performed.
- Conversion fails if a requested character is absent from the BDF font.

Line breaks in the example above do not add glyphs. If the generated font must
contain an ASCII space, include an actual space somewhere in the subset file.

## Generating a font

Create the output directory first; the command does not create directories.
Then run:

```sh
modgadget-fonts \
    -bdf fonts/example/example.bdf \
    -subset characters.txt \
    -package assets \
    -var HeadlineFont \
    -o assets/headline_font.go
```

The flags are all required:

- `-bdf`: input Unicode/ISO10646 BDF file
- `-subset`: UTF-8 text file specifying the characters to include
- `-package`: Go package name written to the generated source
- `-var`: name of the generated `font.Font` variable
- `-o`: output Go source file

The command writes the output only after parsing, subsetting, conversion, and
source generation have all succeeded. An existing output file is overwritten.

## Runtime use

Import the generated package and look up glyphs by rune:

```go
import "example.com/project/assets"

glyph, ok := assets.HeadlineFont.Lookup('A')
if !ok {
    // glyph is not included
}
```

A returned glyph contains:

- `Bitmap`: an immutable substring of the font's shared bitmap data
- `Width` and `Height`: bitmap dimensions in pixels
- `XOffset` and `YOffset`: placement relative to the pen and line top
- `Advance`: horizontal pen advance

`YOffset` has already been converted from the BDF coordinate system to a value
measured from the top of the runtime line, so callers do not need to interpret
the BDF baseline-relative Y offset.

Bitmap data is 1-bit, row-major, top-to-bottom, and MSB first. Each row occupies
`(Width + 7) / 8` bytes. No antialiasing, compression, rendering surface, or
framebuffer is part of the runtime font data.

## Generation pipeline

The command uses the following pipeline:

```text
BDF
  -> Parse
  -> Subset
  -> Convert
  -> Generate Go source
```

The parser, subsetter, converter, and generator are implementation details in
`internal` packages. Normal users should use the `modgadget-fonts` command to
generate data and the public `font` package through the generated font value at
runtime.

## License

The conversion tool and runtime source code developed in this repository are
licensed under the BSD 3-Clause License. See [LICENSE](LICENSE).

BDF font files, generated bitmap data, and other data derived from a font remain
subject to the original font's license. Generating a Go file does not
automatically relicense that font data under the repository's BSD 3-Clause
License. If a font directory contains its own `LICENSE`, review that file and
the directory documentation before using or redistributing the font or generated
data.

See [LICENSES/README.md](LICENSES/README.md) for the repository's font-license
policy.
