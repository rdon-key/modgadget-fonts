// Package font provides compact bitmap font data and allocation-free glyph
// lookup for memory-constrained systems.
package font

// Metrics contains the font-wide metrics needed at runtime.
type Metrics struct {
	Ascent  int16
	Descent int16
}

// GlyphInfo contains the stored metadata for one glyph.
//
// GlyphInfo values in a Font must be sorted by Rune in ascending order. The
// bitmap length is derived from Width and Height.
type GlyphInfo struct {
	Rune         rune
	BitmapOffset uint32

	Width   int16
	Height  int16
	XOffset int16
	YOffset int16
	Advance int16
}

// Glyph is a view of one glyph returned by Font.Lookup.
//
// Bitmap is an immutable substring of the Font's shared bitmap data.
type Glyph struct {
	Rune   rune
	Bitmap string

	Width   int16
	Height  int16
	XOffset int16
	YOffset int16
	Advance int16
}

// Font contains runtime metrics, glyph metadata, and shared bitmap data.
type Font struct {
	metrics Metrics
	glyphs  []GlyphInfo
	bitmap  string
}

// New constructs a Font without copying glyphs or bitmap.
//
// glyphs must be sorted by Rune in ascending order and contain no duplicate
// runes. Its bitmap ranges must refer to bitmap. New does not validate these
// invariants.
func New(metrics Metrics, glyphs []GlyphInfo, bitmap string) Font {
	return Font{
		metrics: metrics,
		glyphs:  glyphs,
		bitmap:  bitmap,
	}
}

// Metrics returns the font-wide runtime metrics.
func (f *Font) Metrics() Metrics {
	return f.metrics
}

// GlyphCount returns the number of glyph metadata entries in the font.
func (f *Font) GlyphCount() int {
	return len(f.glyphs)
}

// Lookup returns the glyph for r. The returned immutable Bitmap shares the
// Font's bitmap data and is not copied.
//
// Lookup returns false if r is absent or its metadata describes an invalid
// bitmap range.
func (f *Font) Lookup(r rune) (Glyph, bool) {
	low, high := 0, len(f.glyphs)
	for low < high {
		middle := low + (high-low)/2
		if f.glyphs[middle].Rune < r {
			low = middle + 1
		} else {
			high = middle
		}
	}

	if low == len(f.glyphs) || f.glyphs[low].Rune != r {
		return Glyph{}, false
	}

	info := f.glyphs[low]
	if info.Width < 0 || info.Height < 0 {
		return Glyph{}, false
	}

	width := uint64(info.Width)
	height := uint64(info.Height)
	rowBytes := (width + 7) / 8
	length := rowBytes * height
	start := uint64(info.BitmapOffset)
	end := start + length

	// All arithmetic above is performed in uint64. Current field widths make
	// rowBytes and length unable to overflow uint64; checking end against start
	// also keeps this correct if the metadata fields are widened in the future.
	if end < start {
		return Glyph{}, false
	}
	if end > uint64(^uint32(0)) {
		return Glyph{}, false
	}

	maxInt := uint64(^uint(0) >> 1)
	if start > maxInt || end > maxInt || end > uint64(len(f.bitmap)) {
		return Glyph{}, false
	}

	startIndex := int(start)
	endIndex := int(end)
	return Glyph{
		Rune:    info.Rune,
		Bitmap:  f.bitmap[startIndex:endIndex],
		Width:   info.Width,
		Height:  info.Height,
		XOffset: info.XOffset,
		YOffset: info.YOffset,
		Advance: info.Advance,
	}, true
}
