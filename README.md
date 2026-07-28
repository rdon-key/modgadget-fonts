# ModGadget Fonts

> 🚧 Work in progress. Formats and APIs may change.

Font conversion tools and compact bitmap font data for ModGadget and TinyGo.

The project is intended for memory-constrained embedded devices and will
initially support 1-bit bitmap fonts converted from BDF files.

## Planned

- BDF to Go source conversion
- Unicode glyph lookup
- Font subset generation
- English, Japanese, and Simplified Chinese fonts
- Allocation-free runtime access

## Repository layout

- `font`: runtime font types and glyph lookup
- `cmd/bdf2mgfont`: BDF conversion tool
- `fonts`: generated font packages
- `testdata`: test BDF files and expected output
- `LICENSES`: licenses for bundled or derived font data

## Status

Experimental. No stable API is available yet.

## License

Source code developed for ModGadget Fonts is licensed under the BSD 3-Clause
License. See `LICENSE`.

Font files and generated font data may be subject to separate licenses.
When a font directory contains its own `LICENSE` file, that license applies
to the font files and generated data in that directory.

See `LICENSES/README.md` and the documentation in each font directory for
details.
