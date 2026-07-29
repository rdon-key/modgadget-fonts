package bdf

import (
	"fmt"
	"sort"

	"github.com/rdon-key/modgadget-fonts/font"
)

// ConvertedFont contains BDF data converted to the runtime font format.
type ConvertedFont struct {
	Metrics font.Metrics
	Glyphs  []font.GlyphInfo
	Bitmap  string
}

// Convert validates and converts a parsed BDF font to the runtime format.
// Convert does not modify src or any of its glyph bitmaps.
func Convert(src Font) (ConvertedFont, error) {
	ordered := append([]Glyph(nil), src.Glyphs...)
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].Encoding < ordered[j].Encoding
	})

	var totalBitmap uint64
	for i := range ordered {
		glyph := &ordered[i]
		if err := validateEncoding(*glyph); err != nil {
			return ConvertedFont{}, err
		}
		if i > 0 && ordered[i-1].Encoding == glyph.Encoding {
			return ConvertedFont{}, fmt.Errorf("%s: duplicate encoding U+%04X", glyphDescription(*glyph), glyph.Encoding)
		}
		if glyph.Box.Width < 0 || glyph.Box.Height < 0 {
			return ConvertedFont{}, fmt.Errorf("%s: width and height must not be negative", glyphDescription(*glyph))
		}

		rowBytes := (uint64(glyph.Box.Width) + 7) / 8
		bitmapLength := rowBytes * uint64(glyph.Box.Height)
		if uint64(len(glyph.Bitmap)) != bitmapLength {
			return ConvertedFont{}, fmt.Errorf("%s: bitmap has %d bytes; dimensions %dx%d require %d",
				glyphDescription(*glyph), len(glyph.Bitmap), glyph.Box.Width, glyph.Box.Height, bitmapLength)
		}

		bearingY := int32(glyph.Box.Height) + int32(glyph.Box.YOffset)
		if bearingY < -1<<15 || bearingY > 1<<15-1 {
			return ConvertedFont{}, fmt.Errorf("%s: converted BearingY %d does not fit int16", glyphDescription(*glyph), bearingY)
		}

		if totalBitmap > uint64(^uint32(0)) || bitmapLength > uint64(^uint32(0))-totalBitmap {
			return ConvertedFont{}, fmt.Errorf("%s: bitmap offset or length exceeds uint32", glyphDescription(*glyph))
		}
		totalBitmap += bitmapLength
	}

	maxInt := uint64(^uint(0) >> 1)
	if totalBitmap > maxInt {
		return ConvertedFont{}, fmt.Errorf("%s: converted bitmap length %d does not fit int",
			glyphDescription(ordered[len(ordered)-1]), totalBitmap)
	}

	result := ConvertedFont{
		Metrics: font.Metrics{Ascent: src.Ascent, Descent: src.Descent, LineGap: 0},
		Glyphs:  make([]font.GlyphInfo, 0, len(ordered)),
	}
	bitmap := make([]byte, 0, int(totalBitmap))

	for _, glyph := range ordered {
		offset := uint32(len(bitmap))
		rowBytes := (int(glyph.Box.Width) + 7) / 8
		bitmapStart := len(bitmap)
		bitmap = append(bitmap, glyph.Bitmap...)

		if remainder := int(glyph.Box.Width) % 8; remainder != 0 {
			mask := byte(0xff << (8 - remainder))
			for row := 0; row < int(glyph.Box.Height); row++ {
				lastByte := bitmapStart + (row+1)*rowBytes - 1
				bitmap[lastByte] &= mask
			}
		}

		bearingY := int16(int32(glyph.Box.Height) + int32(glyph.Box.YOffset))
		result.Glyphs = append(result.Glyphs, font.GlyphInfo{
			Rune:         glyph.Encoding,
			BitmapOffset: offset,
			Width:        glyph.Box.Width,
			Height:       glyph.Box.Height,
			AdvanceX:     glyph.Advance,
			BearingX:     glyph.Box.XOffset,
			BearingY:     bearingY,
		})
	}

	result.Bitmap = string(bitmap)
	return result, nil
}

func validateEncoding(glyph Glyph) error {
	encoding := int64(glyph.Encoding)
	if encoding < 0 || encoding > 0x10ffff {
		return fmt.Errorf("%s: encoding is outside the Unicode range", glyphDescription(glyph))
	}
	if encoding >= 0xd800 && encoding <= 0xdfff {
		return fmt.Errorf("%s: encoding U+%04X is a Unicode surrogate", glyphDescription(glyph), glyph.Encoding)
	}
	return nil
}

func glyphDescription(glyph Glyph) string {
	if glyph.Name != "" {
		return fmt.Sprintf("glyph %q", glyph.Name)
	}
	if glyph.Encoding >= 0 {
		return fmt.Sprintf("glyph U+%04X", glyph.Encoding)
	}
	return fmt.Sprintf("glyph encoding %d", glyph.Encoding)
}
