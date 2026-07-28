package bdf

import (
	"reflect"
	"strings"
	"testing"
)

func subsetTestFont() Font {
	return Font{
		Name:            "subset-test",
		Size:            Size{PointSize: 16, ResolutionX: 75, ResolutionY: 75},
		BoundingBox:     Box{Width: 8, Height: 1, XOffset: -1, YOffset: -2},
		Ascent:          10,
		Descent:         3,
		CharsetRegistry: "ISO10646",
		CharsetEncoding: "1",
		Glyphs: []Glyph{
			{Name: "zhong", Encoding: '\u4e2d', Box: Box{Width: 8, Height: 1}, Bitmap: []byte{0x01}},
			{Name: "space", Encoding: ' ', Advance: 4},
			{Name: "B", Encoding: 'B', Box: Box{Width: 8, Height: 1}, Bitmap: []byte{0x02}},
			{Name: "hiragana-a", Encoding: '\u3042', Box: Box{Width: 8, Height: 1}, Bitmap: []byte{0x03}},
			{Name: "guo", Encoding: '\u56fd', Box: Box{Width: 8, Height: 1}, Bitmap: []byte{0x04}},
			{Name: "A", Encoding: 'A', Box: Box{Width: 8, Height: 1}, Bitmap: []byte{0x05}},
			{Name: "BOM", Encoding: '\ufeff', Box: Box{Width: 8, Height: 1}, Bitmap: []byte{0x06}},
			{Name: "combining-acute", Encoding: '\u0301', Box: Box{Width: 8, Height: 1}, Bitmap: []byte{0x07}},
			{Name: "variation-selector-16", Encoding: '\ufe0f', Box: Box{Width: 8, Height: 1}, Bitmap: []byte{0x08}},
		},
	}
}

func TestSubsetASCIIOrderAndDuplicates(t *testing.T) {
	src := subsetTestFont()
	got, err := Subset(src, []byte("BAAB"))
	if err != nil {
		t.Fatalf("Subset() error: %v", err)
	}
	assertGlyphEncodings(t, got.Glyphs, []rune{'A', 'B'})
}

func TestSubsetJapaneseAndSimplifiedChinese(t *testing.T) {
	src := subsetTestFont()
	got, err := Subset(src, []byte("\u56fd\u3042\u4e2d"))
	if err != nil {
		t.Fatalf("Subset() error: %v", err)
	}
	assertGlyphEncodings(t, got.Glyphs, []rune{'\u3042', '\u4e2d', '\u56fd'})
}

func TestSubsetBOM(t *testing.T) {
	src := subsetTestFont()

	leading, err := Subset(src, []byte("\ufeffA"))
	if err != nil {
		t.Fatalf("Subset() with leading BOM error: %v", err)
	}
	assertGlyphEncodings(t, leading.Glyphs, []rune{'A'})

	embedded, err := Subset(src, []byte("A\ufeff"))
	if err != nil {
		t.Fatalf("Subset() with embedded U+FEFF error: %v", err)
	}
	assertGlyphEncodings(t, embedded.Glyphs, []rune{'A', '\ufeff'})
}

func TestSubsetIgnoredControlsAndSpace(t *testing.T) {
	src := subsetTestFont()
	got, err := Subset(src, []byte("B\r\nA\t "))
	if err != nil {
		t.Fatalf("Subset() error: %v", err)
	}
	assertGlyphEncodings(t, got.Glyphs, []rune{' ', 'A', 'B'})
}

func TestSubsetCombiningCharacterAndVariationSelector(t *testing.T) {
	src := subsetTestFont()
	got, err := Subset(src, []byte("\ufe0f\u0301"))
	if err != nil {
		t.Fatalf("Subset() error: %v", err)
	}
	assertGlyphEncodings(t, got.Glyphs, []rune{'\u0301', '\ufe0f'})
}

func TestSubsetEmptyTextPreservesFontInformation(t *testing.T) {
	src := subsetTestFont()
	before := append([]Glyph(nil), src.Glyphs...)

	got, err := Subset(src, nil)
	if err != nil {
		t.Fatalf("Subset() error: %v", err)
	}
	if len(got.Glyphs) != 0 {
		t.Fatalf("len(Glyphs) = %d, want 0", len(got.Glyphs))
	}
	got.Glyphs = src.Glyphs
	if !reflect.DeepEqual(got, src) {
		t.Errorf("font-wide information was not preserved:\ngot:  %+v\nsrc: %+v", got, src)
	}
	if !reflect.DeepEqual(src.Glyphs, before) {
		t.Errorf("source Glyphs changed:\ngot:  %+v\nwant: %+v", src.Glyphs, before)
	}
}

func TestSubsetDoesNotChangeSourceAndSharesBitmapRange(t *testing.T) {
	src := subsetTestFont()
	before := append([]Glyph(nil), src.Glyphs...)
	sourceBitmap := src.Glyphs[5].Bitmap

	got, err := Subset(src, []byte("A"))
	if err != nil {
		t.Fatalf("Subset() error: %v", err)
	}
	if !reflect.DeepEqual(src.Glyphs, before) {
		t.Fatalf("source Glyphs changed:\ngot:  %+v\nwant: %+v", src.Glyphs, before)
	}
	if len(got.Glyphs) != 1 || !reflect.DeepEqual(got.Glyphs[0].Bitmap, sourceBitmap) {
		t.Fatalf("selected Bitmap = %x, want %x", got.Glyphs[0].Bitmap, sourceBitmap)
	}
	if len(sourceBitmap) > 0 && &got.Glyphs[0].Bitmap[0] != &sourceBitmap[0] {
		t.Fatal("selected Bitmap does not retain the source reference range")
	}
}

func TestSubsetErrors(t *testing.T) {
	tests := []struct {
		name   string
		font   Font
		text   []byte
		needle string
	}{
		{
			name:   "invalid UTF-8",
			font:   subsetTestFont(),
			text:   []byte{0xff},
			needle: "valid UTF-8",
		},
		{
			name:   "missing glyph",
			font:   subsetTestFont(),
			text:   []byte("\u96ea"),
			needle: "U+96EA",
		},
		{
			name: "duplicate source encoding",
			font: Font{Glyphs: []Glyph{
				{Name: "first", Encoding: 'A'},
				{Name: "second", Encoding: 'A'},
			}},
			needle: "duplicate encoding",
		},
		{
			name:   "negative source encoding",
			font:   Font{Glyphs: []Glyph{{Name: "negative", Encoding: -1}}},
			needle: "outside the Unicode range",
		},
		{
			name:   "source encoding out of range",
			font:   Font{Glyphs: []Glyph{{Name: "outside", Encoding: rune(0x110000)}}},
			needle: "outside the Unicode range",
		},
		{
			name:   "source surrogate",
			font:   Font{Glyphs: []Glyph{{Name: "surrogate", Encoding: rune(0xd800)}}},
			needle: "Unicode surrogate",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Subset(test.font, test.text)
			if err == nil {
				t.Fatal("Subset() succeeded, want error")
			}
			if !strings.Contains(err.Error(), test.needle) {
				t.Errorf("error = %q, want substring %q", err, test.needle)
			}
		})
	}
}

func TestSubsetDeterministicCharacterSet(t *testing.T) {
	src := subsetTestFont()
	inputs := []string{"BAAB", "AB", "B\nA", "A\tB"}

	var first Font
	for i, input := range inputs {
		got, err := Subset(src, []byte(input))
		if err != nil {
			t.Fatalf("Subset(%q) error: %v", input, err)
		}
		if i == 0 {
			first = got
			continue
		}
		if !reflect.DeepEqual(got, first) {
			t.Errorf("Subset(%q) differs from Subset(%q):\ngot:  %+v\nwant: %+v", input, inputs[0], got, first)
		}
	}
}

func assertGlyphEncodings(t *testing.T, glyphs []Glyph, want []rune) {
	t.Helper()
	got := make([]rune, len(glyphs))
	for i, glyph := range glyphs {
		got[i] = glyph.Encoding
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("glyph encodings = %U, want %U", got, want)
	}
}
