// Package bdf parses the subset of the Glyph Bitmap Distribution Format used
// by ModGadget Fonts.
package bdf

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Parse parses a Unicode/ISO10646 BDF font. filename is used only in errors.
func Parse(filename string, r io.Reader) (Font, error) {
	if r == nil {
		return Font{}, fmt.Errorf("%s:1: reader is nil", filename)
	}
	p := parser{
		filename:  filename,
		encodings: make(map[rune]struct{}),
	}

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1024), 1024*1024)
	for scanner.Scan() {
		p.line++
		line := strings.TrimSuffix(scanner.Text(), "\r")
		if err := p.parseLine(line); err != nil {
			return Font{}, err
		}
	}
	if err := scanner.Err(); err != nil {
		return Font{}, p.errorf("read BDF: %v", err)
	}
	if err := p.finish(); err != nil {
		return Font{}, err
	}
	return p.font, nil
}

type parser struct {
	filename string
	line     int

	font Font

	started        bool
	ended          bool
	fontSeen       bool
	sizeSeen       bool
	boxSeen        bool
	properties     bool
	propertiesSeen bool
	propertiesEnd  bool
	propertyCount  int
	propertyActual int
	ascentSeen     bool
	descentSeen    bool
	registrySeen   bool
	encodingSeen   bool
	charsSeen      bool
	declaredChars  int
	blockCount     int

	glyph     *glyphBuilder
	encodings map[rune]struct{}
}

type glyphBuilder struct {
	glyph Glyph

	encodingValue int64
	encodingSeen  bool
	advanceSeen   bool
	boxSeen       bool
	bitmapSeen    bool
	bitmapRows    int
}

func (p *parser) parseLine(raw string) error {
	trimmed := strings.TrimSpace(raw)
	if p.ended {
		if trimmed != "" {
			return p.errorf("content after ENDFONT")
		}
		return nil
	}

	if p.glyph != nil {
		return p.parseGlyphLine(trimmed)
	}
	if trimmed == "" {
		// Blank lines are permitted between properties and do not contribute to
		// the STARTPROPERTIES count.
		return nil
	}

	keyword, value := splitDirective(trimmed)
	if !p.started {
		if keyword != "STARTFONT" {
			return p.errorf("STARTFONT is required before %s", keyword)
		}
		if value == "" {
			return p.errorf("STARTFONT requires a version")
		}
		p.started = true
		return nil
	}

	if p.properties {
		if keyword == "ENDPROPERTIES" {
			if p.propertyActual != p.propertyCount {
				return p.errorf("STARTPROPERTIES declares %d properties but found %d", p.propertyCount, p.propertyActual)
			}
			p.properties = false
			p.propertiesEnd = true
			return nil
		}
		if isStructuralDirective(keyword) {
			return p.errorf("%s is not allowed inside STARTPROPERTIES", keyword)
		}
		p.propertyActual++
		return p.parseProperty(keyword, value)
	}

	switch keyword {
	case "STARTFONT":
		return p.errorf("duplicate STARTFONT")
	case "FONT":
		if p.fontSeen {
			return p.errorf("duplicate FONT")
		}
		if value == "" {
			return p.errorf("FONT requires a name")
		}
		p.fontSeen = true
		p.font.Name = value
	case "SIZE":
		if p.sizeSeen {
			return p.errorf("duplicate SIZE")
		}
		values, err := p.int16Values("SIZE", value, 3)
		if err != nil {
			return err
		}
		p.sizeSeen = true
		p.font.Size = Size{PointSize: values[0], ResolutionX: values[1], ResolutionY: values[2]}
	case "FONTBOUNDINGBOX":
		if p.boxSeen {
			return p.errorf("duplicate FONTBOUNDINGBOX")
		}
		box, err := p.parseBox("FONTBOUNDINGBOX", value)
		if err != nil {
			return err
		}
		if box.Width < 0 || box.Height < 0 {
			return p.errorf("FONTBOUNDINGBOX width and height must not be negative")
		}
		p.boxSeen = true
		p.font.BoundingBox = box
	case "STARTPROPERTIES":
		if p.propertiesSeen {
			return p.errorf("duplicate STARTPROPERTIES")
		}
		count, err := p.nonNegativeInt("STARTPROPERTIES", value)
		if err != nil {
			return err
		}
		p.propertiesSeen = true
		p.properties = true
		p.propertyCount = count
	case "ENDPROPERTIES":
		return p.errorf("ENDPROPERTIES without STARTPROPERTIES")
	case "CHARS":
		if p.charsSeen {
			return p.errorf("duplicate CHARS")
		}
		count, err := p.nonNegativeInt("CHARS", value)
		if err != nil {
			return err
		}
		p.charsSeen = true
		p.declaredChars = count
	case "STARTCHAR":
		if !p.charsSeen {
			return p.errorf("STARTCHAR before CHARS")
		}
		if value == "" {
			return p.errorf("STARTCHAR requires a name")
		}
		p.blockCount++
		p.glyph = &glyphBuilder{glyph: Glyph{Name: value}, encodingValue: -2}
	case "ENDFONT":
		if err := p.validateFont(); err != nil {
			return err
		}
		p.ended = true
	case "COMMENT", "COPYRIGHT", "DEFAULT_CHAR", "SWIDTH":
		// Ignored metadata.
	case "ENCODING", "DWIDTH", "BBX", "BITMAP", "ENDCHAR":
		return p.errorf("%s is not allowed outside a glyph", keyword)
	default:
		// Unknown global metadata is ignored. Structural requirements are checked
		// when ENDFONT is reached.
	}
	return nil
}

func (p *parser) parseProperty(keyword, value string) error {
	switch keyword {
	case "FONT_ASCENT":
		if p.ascentSeen {
			return p.errorf("duplicate FONT_ASCENT")
		}
		v, err := p.oneInt16("FONT_ASCENT", value)
		if err != nil {
			return err
		}
		if v < 0 {
			return p.errorf("FONT_ASCENT must not be negative")
		}
		p.ascentSeen = true
		p.font.Ascent = v
	case "FONT_DESCENT":
		if p.descentSeen {
			return p.errorf("duplicate FONT_DESCENT")
		}
		v, err := p.oneInt16("FONT_DESCENT", value)
		if err != nil {
			return err
		}
		if v < 0 {
			return p.errorf("FONT_DESCENT must not be negative")
		}
		p.descentSeen = true
		p.font.Descent = v
	case "CHARSET_REGISTRY":
		if p.registrySeen {
			return p.errorf("duplicate CHARSET_REGISTRY")
		}
		v, err := p.propertyString("CHARSET_REGISTRY", value)
		if err != nil {
			return err
		}
		if !strings.EqualFold(v, "ISO10646") {
			return p.errorf("unsupported CHARSET_REGISTRY %q; only ISO10646 is supported", v)
		}
		p.registrySeen = true
		p.font.CharsetRegistry = v
	case "CHARSET_ENCODING":
		if p.encodingSeen {
			return p.errorf("duplicate CHARSET_ENCODING")
		}
		v, err := p.propertyString("CHARSET_ENCODING", value)
		if err != nil {
			return err
		}
		p.encodingSeen = true
		p.font.CharsetEncoding = v
	default:
		// Unused properties are intentionally ignored.
	}
	return nil
}

func (p *parser) parseGlyphLine(line string) error {
	g := p.glyph
	if g.bitmapSeen {
		if line == "ENDCHAR" {
			return p.finishGlyph()
		}
		keyword, _ := splitDirective(line)
		if keyword == "STARTCHAR" || keyword == "ENDFONT" {
			return p.errorf("ENDCHAR missing for glyph %q", g.glyph.Name)
		}
		if isStructuralDirective(keyword) {
			return p.errorf("%s is not allowed in BITMAP for glyph %q", keyword, g.glyph.Name)
		}
		if line == "" {
			if g.glyph.Box.Width == 0 && g.bitmapRows < int(g.glyph.Box.Height) {
				g.bitmapRows++
			}
			return nil
		}
		return p.parseBitmapRow(line)
	}

	if line == "" {
		return nil
	}
	keyword, value := splitDirective(line)
	switch keyword {
	case "ENCODING":
		if g.encodingSeen {
			return p.errorf("duplicate ENCODING in glyph %q", g.glyph.Name)
		}
		fields := strings.Fields(value)
		if len(fields) != 1 {
			if len(fields) == 2 {
				return p.errorf("two-value ENCODING is not supported")
			}
			return p.errorf("ENCODING requires one value")
		}
		v, err := strconv.ParseInt(fields[0], 10, 64)
		if err != nil {
			return p.errorf("invalid ENCODING %q", fields[0])
		}
		if v < -1 {
			return p.errorf("ENCODING must be -1 or a Unicode code point")
		}
		if v > 0x10ffff {
			return p.errorf("ENCODING U+%X is outside the Unicode range", v)
		}
		if v >= 0xd800 && v <= 0xdfff {
			return p.errorf("ENCODING U+%04X is a Unicode surrogate", v)
		}
		if v >= 0 {
			r := rune(v)
			if _, exists := p.encodings[r]; exists {
				return p.errorf("duplicate ENCODING U+%04X", r)
			}
			p.encodings[r] = struct{}{}
		}
		g.encodingSeen = true
		g.encodingValue = v
	case "DWIDTH":
		if g.advanceSeen {
			return p.errorf("duplicate DWIDTH in glyph %q", g.glyph.Name)
		}
		values, err := p.int16Values("DWIDTH", value, 2)
		if err != nil {
			return err
		}
		if values[1] != 0 {
			return p.errorf("DWIDTH Y value must be 0")
		}
		g.advanceSeen = true
		g.glyph.Advance = values[0]
	case "BBX":
		if g.boxSeen {
			return p.errorf("duplicate BBX in glyph %q", g.glyph.Name)
		}
		box, err := p.parseBox("BBX", value)
		if err != nil {
			return err
		}
		if box.Width < 0 || box.Height < 0 {
			return p.errorf("BBX width and height must not be negative")
		}
		g.boxSeen = true
		g.glyph.Box = box
	case "BITMAP":
		if g.bitmapSeen {
			return p.errorf("duplicate BITMAP in glyph %q", g.glyph.Name)
		}
		if !g.boxSeen {
			return p.errorf("BITMAP before BBX in glyph %q", g.glyph.Name)
		}
		if value != "" {
			return p.errorf("BITMAP does not accept a value")
		}
		g.bitmapSeen = true
	case "ENDCHAR":
		return p.errorf("BITMAP missing in glyph %q", g.glyph.Name)
	case "STARTFONT", "FONT", "SIZE", "FONTBOUNDINGBOX", "STARTPROPERTIES", "ENDPROPERTIES", "CHARS", "STARTCHAR", "ENDFONT":
		return p.errorf("%s is not allowed inside glyph %q", keyword, g.glyph.Name)
	case "COMMENT", "SWIDTH":
		// Ignored glyph metadata.
	default:
		// Unknown non-structural glyph metadata is ignored.
	}
	return nil
}

func (p *parser) parseBitmapRow(line string) error {
	g := p.glyph
	rowBytes := (int(g.glyph.Box.Width) + 7) / 8
	if len(line)%2 != 0 {
		return p.errorf("invalid hexadecimal BITMAP row %q", line)
	}
	actualBytes := len(line) / 2
	if actualBytes != rowBytes {
		return p.errorf("BITMAP row has %d byte; glyph width %d requires %d", actualBytes, g.glyph.Box.Width, rowBytes)
	}
	decoded := make([]byte, actualBytes)
	if _, err := hex.Decode(decoded, []byte(line)); err != nil {
		return p.errorf("invalid hexadecimal BITMAP row %q", line)
	}
	g.glyph.Bitmap = append(g.glyph.Bitmap, decoded...)
	g.bitmapRows++
	return nil
}

func (p *parser) finishGlyph() error {
	g := p.glyph
	if !g.encodingSeen {
		return p.errorf("ENCODING missing in glyph %q", g.glyph.Name)
	}
	if !g.advanceSeen {
		return p.errorf("DWIDTH missing in glyph %q", g.glyph.Name)
	}
	if !g.boxSeen {
		return p.errorf("BBX missing in glyph %q", g.glyph.Name)
	}
	if !g.bitmapSeen {
		return p.errorf("BITMAP missing in glyph %q", g.glyph.Name)
	}
	if g.bitmapRows != int(g.glyph.Box.Height) {
		return p.errorf("BITMAP has %d rows; glyph height %d requires %d", g.bitmapRows, g.glyph.Box.Height, g.glyph.Box.Height)
	}
	if g.encodingValue >= 0 {
		g.glyph.Encoding = rune(g.encodingValue)
		p.font.Glyphs = append(p.font.Glyphs, g.glyph)
	}
	p.glyph = nil
	return nil
}

func (p *parser) validateFont() error {
	checks := []struct {
		ok   bool
		name string
	}{
		{p.fontSeen, "FONT"},
		{p.sizeSeen, "SIZE"},
		{p.boxSeen, "FONTBOUNDINGBOX"},
		{p.propertiesSeen, "STARTPROPERTIES"},
		{p.propertiesEnd, "ENDPROPERTIES"},
		{p.ascentSeen, "FONT_ASCENT"},
		{p.descentSeen, "FONT_DESCENT"},
		{p.registrySeen, "CHARSET_REGISTRY"},
		{p.charsSeen, "CHARS"},
	}
	for _, check := range checks {
		if !check.ok {
			return p.errorf("%s missing", check.name)
		}
	}
	if p.blockCount != p.declaredChars {
		return p.errorf("CHARS declares %d glyphs but found %d STARTCHAR blocks", p.declaredChars, p.blockCount)
	}
	return nil
}

func (p *parser) finish() error {
	if !p.started {
		return p.errorf("STARTFONT missing")
	}
	if p.glyph != nil {
		return p.errorf("ENDCHAR missing for glyph %q", p.glyph.glyph.Name)
	}
	if p.properties {
		return p.errorf("ENDPROPERTIES missing")
	}
	if !p.ended {
		return p.errorf("ENDFONT missing")
	}
	return nil
}

func (p *parser) parseBox(name, value string) (Box, error) {
	values, err := p.int16Values(name, value, 4)
	if err != nil {
		return Box{}, err
	}
	return Box{Width: values[0], Height: values[1], XOffset: values[2], YOffset: values[3]}, nil
}

func (p *parser) oneInt16(name, value string) (int16, error) {
	values, err := p.int16Values(name, value, 1)
	if err != nil {
		return 0, err
	}
	return values[0], nil
}

func (p *parser) int16Values(name, value string, count int) ([]int16, error) {
	fields := strings.Fields(value)
	if len(fields) != count {
		return nil, p.errorf("%s requires %d values", name, count)
	}
	values := make([]int16, count)
	for i, field := range fields {
		v, err := strconv.ParseInt(field, 10, 16)
		if err != nil {
			return nil, p.errorf("%s value %q does not fit int16", name, field)
		}
		values[i] = int16(v)
	}
	return values, nil
}

func (p *parser) nonNegativeInt(name, value string) (int, error) {
	fields := strings.Fields(value)
	if len(fields) != 1 {
		return 0, p.errorf("%s requires one value", name)
	}
	v, err := strconv.ParseInt(fields[0], 10, 32)
	if err != nil || v < 0 || uint64(v) > uint64(^uint(0)>>1) {
		return 0, p.errorf("%s requires a non-negative integer", name)
	}
	return int(v), nil
}

func (p *parser) propertyString(name, value string) (string, error) {
	v := strings.TrimSpace(value)
	if len(v) >= 2 && v[0] == '"' && v[len(v)-1] == '"' {
		v = v[1 : len(v)-1]
	}
	if v == "" {
		return "", p.errorf("%s requires a value", name)
	}
	return v, nil
}

func (p *parser) errorf(format string, args ...any) error {
	line := p.line
	if line == 0 {
		line = 1
	}
	return fmt.Errorf("%s:%d: %s", p.filename, line, fmt.Sprintf(format, args...))
}

func splitDirective(line string) (string, string) {
	separator := strings.IndexAny(line, " \t")
	if separator < 0 {
		return line, ""
	}
	return line[:separator], strings.TrimSpace(line[separator+1:])
}

func isStructuralDirective(keyword string) bool {
	switch keyword {
	case "STARTFONT", "FONT", "SIZE", "FONTBOUNDINGBOX",
		"STARTPROPERTIES", "ENDPROPERTIES", "CHARS", "STARTCHAR",
		"ENCODING", "DWIDTH", "BBX", "BITMAP", "ENDCHAR", "ENDFONT":
		return true
	default:
		return false
	}
}
