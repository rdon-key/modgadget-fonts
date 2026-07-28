package bdf

import (
	"fmt"
	"sort"
	"unicode/utf8"
)

// Subset selects the glyphs used by UTF-8 text and returns them in Unicode
// code point order. Subset does not modify src or copy glyph bitmap data.
func Subset(src Font, text []byte) (Font, error) {
	if !utf8.Valid(text) {
		return Font{}, fmt.Errorf("subset text is not valid UTF-8")
	}
	if len(text) >= 3 && text[0] == 0xef && text[1] == 0xbb && text[2] == 0xbf {
		text = text[3:]
	}

	requestedSet := make(map[rune]struct{})
	for _, r := range string(text) {
		if r == '\r' || r == '\n' || r == '\t' {
			continue
		}
		requestedSet[r] = struct{}{}
	}

	requested := make([]rune, 0, len(requestedSet))
	for r := range requestedSet {
		requested = append(requested, r)
	}
	sort.Slice(requested, func(i, j int) bool {
		return requested[i] < requested[j]
	})

	byEncoding := make(map[rune]Glyph, len(src.Glyphs))
	for _, glyph := range src.Glyphs {
		if err := validateEncoding(glyph); err != nil {
			return Font{}, err
		}
		if _, exists := byEncoding[glyph.Encoding]; exists {
			return Font{}, fmt.Errorf("%s: duplicate encoding U+%04X", glyphDescription(glyph), glyph.Encoding)
		}
		byEncoding[glyph.Encoding] = glyph
	}

	result := src
	result.Glyphs = make([]Glyph, 0, len(requested))
	for _, r := range requested {
		glyph, exists := byEncoding[r]
		if !exists {
			return Font{}, fmt.Errorf("missing glyph U+%04X (%q)", r, r)
		}
		result.Glyphs = append(result.Glyphs, glyph)
	}
	return result, nil
}
