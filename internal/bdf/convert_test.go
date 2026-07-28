package bdf

import (
	"reflect"
	"strings"
	"testing"

	"github.com/rdon-key/modgadget-fonts/font"
)

func TestConvertMetricsOrderCoordinatesAndBitmap(t *testing.T) {
	bBitmap := []byte{0x81, 0x42}
	aBitmap := []byte{0xff, 0xff}
	src := Font{
		Ascent:  10,
		Descent: 3,
		Glyphs: []Glyph{
			{
				Name:     "B",
				Encoding: 'B',
				Advance:  8,
				Box:      Box{Width: 8, Height: 2, XOffset: 1, YOffset: 2},
				Bitmap:   bBitmap,
			},
			{
				Name:     "A",
				Encoding: 'A',
				Advance:  9,
				Box:      Box{Width: 9, Height: 1, XOffset: -1, YOffset: -2},
				Bitmap:   aBitmap,
			},
		},
	}

	got, err := Convert(src)
	if err != nil {
		t.Fatalf("Convert() error: %v", err)
	}
	if got.Metrics != (font.Metrics{Ascent: 10, Descent: 3}) {
		t.Errorf("Metrics = %+v, want ascent 10, descent 3", got.Metrics)
	}
	if len(got.Glyphs) != 2 || got.Glyphs[0].Rune != 'A' || got.Glyphs[1].Rune != 'B' {
		t.Fatalf("Glyphs = %+v, want Unicode order A, B", got.Glyphs)
	}
	if src.Glyphs[0].Encoding != 'B' || src.Glyphs[1].Encoding != 'A' {
		t.Fatalf("Convert changed input glyph order: %+v", src.Glyphs)
	}

	a := got.Glyphs[0]
	if a.BitmapOffset != 0 || a.XOffset != -1 || a.Advance != 9 || a.YOffset != 11 {
		t.Errorf("converted A metadata = %+v", a)
	}
	b := got.Glyphs[1]
	if b.BitmapOffset != 2 || b.XOffset != 1 || b.Advance != 8 || b.YOffset != 6 {
		t.Errorf("converted B metadata = %+v", b)
	}
	if got.Bitmap != "\xff\x80\x81\x42" {
		t.Errorf("Bitmap = %x, want ff808142", got.Bitmap)
	}
	if !reflect.DeepEqual(aBitmap, []byte{0xff, 0xff}) || !reflect.DeepEqual(bBitmap, []byte{0x81, 0x42}) {
		t.Errorf("Convert changed input bitmaps: A=%x B=%x", aBitmap, bBitmap)
	}
}

func TestConvertBitmapMasks(t *testing.T) {
	tests := []struct {
		name   string
		width  int16
		height int16
		bitmap []byte
		want   string
	}{
		{name: "width 8", width: 8, height: 1, bitmap: []byte{0xff}, want: "\xff"},
		{name: "width 9", width: 9, height: 1, bitmap: []byte{0xff, 0xff}, want: "\xff\x80"},
		{name: "width 10", width: 10, height: 1, bitmap: []byte{0xff, 0xff}, want: "\xff\xc0"},
		{name: "width 15", width: 15, height: 1, bitmap: []byte{0xff, 0xff}, want: "\xff\xfe"},
		{
			name:   "multiple rows",
			width:  10,
			height: 2,
			bitmap: []byte{0xaa, 0xff, 0x55, 0xff},
			want:   "\xaa\xc0\x55\xc0",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			original := append([]byte(nil), test.bitmap...)
			got, err := Convert(Font{Glyphs: []Glyph{{
				Name: "mask", Encoding: 'A', Box: Box{Width: test.width, Height: test.height}, Bitmap: test.bitmap,
			}}})
			if err != nil {
				t.Fatalf("Convert() error: %v", err)
			}
			if got.Bitmap != test.want {
				t.Errorf("Bitmap = %x, want %x", got.Bitmap, test.want)
			}
			if !reflect.DeepEqual(test.bitmap, original) {
				t.Errorf("input Bitmap changed from %x to %x", original, test.bitmap)
			}
		})
	}
}

func TestConvertEmptyGlyphAndOffsets(t *testing.T) {
	src := Font{Glyphs: []Glyph{
		{Name: "A", Encoding: 'A', Box: Box{Width: 8, Height: 1}, Bitmap: []byte{0x80}},
		{Name: "empty", Encoding: 'Z', Advance: 4, Box: Box{Width: 0, Height: 0}},
	}}

	got, err := Convert(src)
	if err != nil {
		t.Fatalf("Convert() error: %v", err)
	}
	if got.Glyphs[0].Rune != 'A' || got.Glyphs[0].BitmapOffset != 0 {
		t.Errorf("A glyph = %+v, want offset 0", got.Glyphs[0])
	}
	if got.Glyphs[1].Rune != 'Z' || got.Glyphs[1].BitmapOffset != 1 || got.Glyphs[1].Advance != 4 {
		t.Errorf("empty glyph = %+v, want current-end offset 1 and advance 4", got.Glyphs[1])
	}
	if got.Bitmap != "\x80" {
		t.Errorf("Bitmap = %x, want 80", got.Bitmap)
	}
}

func TestConvertErrors(t *testing.T) {
	tests := []struct {
		name   string
		font   Font
		needle string
	}{
		{
			name: "duplicate encoding",
			font: Font{Glyphs: []Glyph{
				{Name: "first", Encoding: 'A'},
				{Name: "second", Encoding: 'A'},
			}},
			needle: "duplicate encoding",
		},
		{
			name:   "surrogate",
			font:   Font{Glyphs: []Glyph{{Name: "surrogate", Encoding: rune(0xd800)}}},
			needle: "Unicode surrogate",
		},
		{
			name:   "Unicode out of range",
			font:   Font{Glyphs: []Glyph{{Name: "outside", Encoding: rune(0x110000)}}},
			needle: "outside the Unicode range",
		},
		{
			name:   "negative width",
			font:   Font{Glyphs: []Glyph{{Name: "negative-width", Encoding: 'A', Box: Box{Width: -1}}}},
			needle: "must not be negative",
		},
		{
			name:   "negative height",
			font:   Font{Glyphs: []Glyph{{Name: "negative-height", Encoding: 'A', Box: Box{Height: -1}}}},
			needle: "must not be negative",
		},
		{
			name: "bitmap short",
			font: Font{Glyphs: []Glyph{{
				Name: "short", Encoding: 'A', Box: Box{Width: 9, Height: 1}, Bitmap: []byte{0xff},
			}}},
			needle: "require 2",
		},
		{
			name: "bitmap excessive",
			font: Font{Glyphs: []Glyph{{
				Name: "long", Encoding: 'A', Box: Box{Width: 8, Height: 1}, Bitmap: []byte{0xff, 0xff},
			}}},
			needle: "require 1",
		},
		{
			name: "YOffset overflow",
			font: Font{Ascent: 32767, Glyphs: []Glyph{{
				Name: "too-low", Encoding: 'A', Box: Box{YOffset: -32768},
			}}},
			needle: "does not fit int16",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Convert(test.font)
			if err == nil {
				t.Fatal("Convert() succeeded, want error")
			}
			if !strings.Contains(err.Error(), test.needle) {
				t.Errorf("error = %q, want substring %q", err, test.needle)
			}
			if !strings.Contains(err.Error(), "glyph") {
				t.Errorf("error = %q, want glyph identification", err)
			}
		})
	}
}

func TestConvertDeterministic(t *testing.T) {
	src := Font{Ascent: 8, Descent: 2, Glyphs: []Glyph{
		{Name: "B", Encoding: 'B', Box: Box{Width: 8, Height: 1}, Bitmap: []byte{0x42}},
		{Name: "A", Encoding: 'A', Box: Box{Width: 8, Height: 1}, Bitmap: []byte{0x81}},
	}}

	first, err := Convert(src)
	if err != nil {
		t.Fatalf("first Convert() error: %v", err)
	}
	second, err := Convert(src)
	if err != nil {
		t.Fatalf("second Convert() error: %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("Convert results differ:\nfirst:  %+v\nsecond: %+v", first, second)
	}
}
