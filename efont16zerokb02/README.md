# Efont16 subset for Zero-KB02

This package embeds a 17,001-byte MGZ1 bitmap font derived from the
[full Efont Biwidth 16 MGZ](../efont16/efont16-full.mgz).
It covers the complete greeting **恭喜進軍台灣！**.

This is a **compressed-block subset**, not a recompressed or redrawn font.
It keeps seven original MGZ1 blocks (64 glyphs each) at indices
89, 95, 129, 182, 283, 287 and 381, for a total of 448 glyphs. Unchanged
glyph records and compressed block bytes are copied from the source,
with MGZ1 offsets, counts, metadata and file size updated accordingly.

No glyphs are synthesized or rescaled. The input asset is derived from
Efont Unicode BDF 0.4.2 and remains subject to its original font license
and notice, available under [LICENSES](../LICENSES/efont-unicode-bdf-0.4.2/).

It is intentionally smaller than the complete Efont16 MGZ (876,150
bytes). Its `Open()` function validates 448 glyph metadata records rather
than 24,618. Unlike the original Efont16 package, this diagnostic subset
does **not** parse the font at package initialization time. Applications
can initialize serial and the display before calling `Open()`.

Usage:

```go
import "github.com/rdon-key/modgadget-fonts/efont16zerokb02"

font, err := efont16zerokb02.Open()
if err != nil {
    panic(err)
}
```

The package provides a diagnostic workaround for the Zero-KB02 greeting
example. It does not establish the cause of the full Efont16 firmware's
failure to reach `main()` on the RP2040. Device testing is still needed.
