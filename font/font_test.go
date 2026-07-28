package font

import (
	"math"
	"testing"
)

func TestLookupEmptyFont(t *testing.T) {
	f := New(Metrics{}, nil, "")

	if got := f.GlyphCount(); got != 0 {
		t.Fatalf("GlyphCount() = %d, want 0", got)
	}
	if _, ok := f.Lookup('A'); ok {
		t.Fatal("Lookup('A') succeeded for an empty font")
	}
}

func TestNewAndMetrics(t *testing.T) {
	metrics := Metrics{Ascent: 9, Descent: 3}
	glyphs := []GlyphInfo{{Rune: 'A', Width: 8, Height: 1}}
	bitmap := "\x81"
	f := New(metrics, glyphs, bitmap)

	if got := f.Metrics(); got != metrics {
		t.Fatalf("Metrics() = %+v, want %+v", got, metrics)
	}
	if got := f.GlyphCount(); got != 1 {
		t.Fatalf("GlyphCount() = %d, want 1", got)
	}

	// Mutating the input demonstrates that New retained the supplied metadata
	// slice.
	glyphs[0].Rune = 'B'
	got, ok := f.Lookup('B')
	if !ok {
		t.Fatal("Lookup('B') failed after modifying the input glyph slice")
	}
	if got.Bitmap != "\x81" {
		t.Fatalf("Bitmap = %q, want %q", got.Bitmap, "\x81")
	}
}

func TestLookupPositionsAndMissingRune(t *testing.T) {
	glyphs := []GlyphInfo{
		{Rune: 'A', BitmapOffset: 0, Width: 8, Height: 1},
		{Rune: 'M', BitmapOffset: 1, Width: 8, Height: 1},
		{Rune: 'Z', BitmapOffset: 2, Width: 8, Height: 1},
	}
	f := New(Metrics{}, glyphs, "\xa1\xb2\xc3")

	for _, test := range []struct {
		r    rune
		want byte
	}{
		{r: 'A', want: 0xa1},
		{r: 'M', want: 0xb2},
		{r: 'Z', want: 0xc3},
	} {
		got, ok := f.Lookup(test.r)
		if !ok {
			t.Fatalf("Lookup(%q) failed", test.r)
		}
		if len(got.Bitmap) != 1 || got.Bitmap[0] != test.want {
			t.Errorf("Lookup(%q).Bitmap = %q, want %q", test.r, got.Bitmap, string(test.want))
		}
	}

	if _, ok := f.Lookup('N'); ok {
		t.Fatal("Lookup('N') succeeded for a missing rune")
	}
}

func TestLookupBitmapDimensions(t *testing.T) {
	tests := []struct {
		name       string
		info       GlyphInfo
		bitmap     string
		wantBitmap string
	}{
		{
			name:       "width 8",
			info:       GlyphInfo{Rune: 'A', BitmapOffset: 1, Width: 8, Height: 1},
			bitmap:     "\xff\x81",
			wantBitmap: "\x81",
		},
		{
			name:       "width 9",
			info:       GlyphInfo{Rune: 'B', BitmapOffset: 1, Width: 9, Height: 1},
			bitmap:     "\xff\x80\x00",
			wantBitmap: "\x80\x00",
		},
		{
			name:       "multiple rows",
			info:       GlyphInfo{Rune: 'C', BitmapOffset: 1, Width: 9, Height: 3},
			bitmap:     "\xff\x80\x00\x40\x00\x20\x00",
			wantBitmap: "\x80\x00\x40\x00\x20\x00",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f := New(Metrics{}, []GlyphInfo{test.info}, test.bitmap)
			got, ok := f.Lookup(test.info.Rune)
			if !ok {
				t.Fatal("Lookup failed")
			}
			if len(got.Bitmap) != len(test.wantBitmap) {
				t.Fatalf("len(Bitmap) = %d, want %d", len(got.Bitmap), len(test.wantBitmap))
			}
			if got.Bitmap != test.wantBitmap {
				t.Fatalf("Bitmap = %q, want %q", got.Bitmap, test.wantBitmap)
			}
		})
	}
}

func TestLookupRejectsInvalidMetadata(t *testing.T) {
	tests := []struct {
		name   string
		info   GlyphInfo
		bitmap string
	}{
		{
			name:   "negative width",
			info:   GlyphInfo{Rune: 'A', Width: -1, Height: 1},
			bitmap: "\x00",
		},
		{
			name:   "negative height",
			info:   GlyphInfo{Rune: 'A', Width: 1, Height: -1},
			bitmap: "\x00",
		},
		{
			name:   "offset out of range",
			info:   GlyphInfo{Rune: 'A', BitmapOffset: 2, Width: 0, Height: 0},
			bitmap: "\x00",
		},
		{
			name:   "range past bitmap end",
			info:   GlyphInfo{Rune: 'A', BitmapOffset: 1, Width: 9, Height: 1},
			bitmap: "\x00\x00",
		},
		{
			name:   "offset plus length exceeds uint32",
			info:   GlyphInfo{Rune: 'A', BitmapOffset: math.MaxUint32, Width: 8, Height: 1},
			bitmap: "\x00",
		},
		{
			name:   "largest dimensions",
			info:   GlyphInfo{Rune: 'A', Width: math.MaxInt16, Height: math.MaxInt16},
			bitmap: "\x00",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f := New(Metrics{}, []GlyphInfo{test.info}, test.bitmap)
			if _, ok := f.Lookup(test.info.Rune); ok {
				t.Fatal("Lookup succeeded for invalid metadata")
			}
		})
	}
}

func TestLookupEmptyGlyph(t *testing.T) {
	bitmap := "\xaa\xbb"
	glyphs := []GlyphInfo{
		{
			Rune:         ' ',
			BitmapOffset: uint32(len(bitmap)),
			Width:        0,
			Height:       0,
			Advance:      4,
		},
	}
	f := New(Metrics{}, glyphs, bitmap)

	got, ok := f.Lookup(' ')
	if !ok {
		t.Fatal("Lookup(' ') failed")
	}
	if len(got.Bitmap) != 0 {
		t.Fatalf("len(Bitmap) = %d, want 0", len(got.Bitmap))
	}
	if got.Advance != 4 {
		t.Fatalf("Advance = %d, want 4", got.Advance)
	}
}

func TestLookupBitmapSubstring(t *testing.T) {
	bitmap := "\x11\x22\x33"
	f := New(Metrics{}, []GlyphInfo{
		{Rune: 'A', BitmapOffset: 1, Width: 8, Height: 1},
	}, bitmap)

	got, ok := f.Lookup('A')
	if !ok {
		t.Fatal("Lookup('A') failed")
	}
	if got.Bitmap != bitmap[1:2] {
		t.Fatalf("Bitmap = %q, want substring %q", got.Bitmap, bitmap[1:2])
	}
}
