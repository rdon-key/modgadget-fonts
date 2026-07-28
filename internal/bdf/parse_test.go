package bdf

import (
	"os"
	"strings"
	"testing"
)

const validBDF = `STARTFONT 2.1
FONT test-font
SIZE 16 75 75
FONTBOUNDINGBOX 9 3 -1 -2
STARTPROPERTIES 6
FONT_ASCENT 10
FONT_DESCENT 3
CHARSET_REGISTRY "iso10646"
CHARSET_ENCODING 1
COPYRIGHT "Example"
DEFAULT_CHAR 32
ENDPROPERTIES
COMMENT ignored global comment
CHARS 4
STARTCHAR Z
ENCODING 90
SWIDTH 500 0
DWIDTH 8 0
BBX 8 2 0 -1
BITMAP
81
7e
ENDCHAR

STARTCHAR A
ENCODING 65
DWIDTH 9 0
BBX 9 1 -1 0
BITMAP
aa80
ENDCHAR
STARTCHAR UNENCODED
ENCODING -1
DWIDTH 8 0
BBX 8 1 0 0
BITMAP
00
ENDCHAR
STARTCHAR SPACE
ENCODING 32
DWIDTH 4 0
BBX 0 0 0 0
BITMAP
ENDCHAR
ENDFONT
`

func TestParseMinimalTestdata(t *testing.T) {
	f, err := os.Open("testdata/minimal.bdf")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	got, err := Parse("minimal.bdf", f)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if got.Name == "" || len(got.Glyphs) != 1 || got.Glyphs[0].Encoding != 'A' {
		t.Fatalf("Parse() = %+v, want one encoded A glyph", got)
	}
	if string(got.Glyphs[0].Bitmap) != "\x81\x7e" {
		t.Fatalf("Bitmap = %x, want 817e", got.Glyphs[0].Bitmap)
	}
}

func TestParseValidFont(t *testing.T) {
	got, err := Parse("input.bdf", strings.NewReader(validBDF))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if got.Name != "test-font" {
		t.Errorf("Name = %q, want test-font", got.Name)
	}
	if got.Size != (Size{PointSize: 16, ResolutionX: 75, ResolutionY: 75}) {
		t.Errorf("Size = %+v", got.Size)
	}
	if got.BoundingBox != (Box{Width: 9, Height: 3, XOffset: -1, YOffset: -2}) {
		t.Errorf("BoundingBox = %+v", got.BoundingBox)
	}
	if got.Ascent != 10 || got.Descent != 3 {
		t.Errorf("metrics = ascent %d, descent %d", got.Ascent, got.Descent)
	}
	if got.CharsetRegistry != "iso10646" || got.CharsetEncoding != "1" {
		t.Errorf("charset = %q-%q", got.CharsetRegistry, got.CharsetEncoding)
	}

	// The unencoded glyph is omitted, while BDF order is retained rather than
	// sorted into Unicode order.
	if len(got.Glyphs) != 3 {
		t.Fatalf("len(Glyphs) = %d, want 3", len(got.Glyphs))
	}
	if got.Glyphs[0].Encoding != 'Z' || got.Glyphs[1].Encoding != 'A' || got.Glyphs[2].Encoding != ' ' {
		t.Fatalf("glyph order = %q, %q, %q; want Z, A, space",
			got.Glyphs[0].Encoding, got.Glyphs[1].Encoding, got.Glyphs[2].Encoding)
	}
	if string(got.Glyphs[0].Bitmap) != "\x81\x7e" {
		t.Errorf("width-8, multi-row Bitmap = %x", got.Glyphs[0].Bitmap)
	}
	if string(got.Glyphs[1].Bitmap) != "\xaa\x80" {
		t.Errorf("width-9 lowercase-hex Bitmap = %x", got.Glyphs[1].Bitmap)
	}
	if got.Glyphs[1].Box.YOffset != 0 {
		t.Errorf("glyph YOffset = %d, want original BDF value 0", got.Glyphs[1].Box.YOffset)
	}
	if len(got.Glyphs[2].Bitmap) != 0 || got.Glyphs[2].Box.Width != 0 || got.Glyphs[2].Box.Height != 0 {
		t.Errorf("empty glyph = %+v", got.Glyphs[2])
	}
}

func TestParseCRLF(t *testing.T) {
	input := strings.ReplaceAll(validBDF, "\n", "\r\n")
	if _, err := Parse("crlf.bdf", strings.NewReader(input)); err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
}

func TestParseZeroWidthPositiveHeight(t *testing.T) {
	input := strings.Replace(validBDF,
		"BBX 0 0 0 0\nBITMAP\nENDCHAR",
		"BBX 0 2 0 0\nBITMAP\n\n\nENDCHAR", 1)

	got, err := Parse("zero-width.bdf", strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	space := got.Glyphs[len(got.Glyphs)-1]
	if space.Box.Width != 0 || space.Box.Height != 2 || len(space.Bitmap) != 0 {
		t.Fatalf("zero-width glyph = %+v, want width 0, height 2, empty bitmap", space)
	}
}

func TestParseWidthPositiveBlankBitmapLineIsNotRow(t *testing.T) {
	input := strings.Replace(validBDF, "81\n7e\nENDCHAR", "81\n\nENDCHAR", 1)
	_, err := Parse("blank-row.bdf", strings.NewReader(input))
	if err == nil || !strings.Contains(err.Error(), "BITMAP has 1 rows") {
		t.Fatalf("Parse() error = %v, want one-row BITMAP error", err)
	}
}

func TestParsePropertyCount(t *testing.T) {
	tests := []struct {
		name string
		bdf  string
		want string
	}{
		{
			name: "blank lines not counted",
			bdf: strings.Replace(
				validBDF,
				"FONT_ASCENT 10\nFONT_DESCENT 3",
				"FONT_ASCENT 10\n\n   \nFONT_DESCENT 3", 1),
		},
		{
			name: "unknown property counted",
			bdf: strings.Replace(
				strings.Replace(validBDF, "STARTPROPERTIES 6", "STARTPROPERTIES 7", 1),
				"ENDPROPERTIES", "UNUSED_PROPERTY value\nENDPROPERTIES", 1),
		},
		{
			name: "declared count too large",
			bdf:  strings.Replace(validBDF, "STARTPROPERTIES 6", "STARTPROPERTIES 7", 1),
			want: "declares 7 properties but found 6",
		},
		{
			name: "declared count too small",
			bdf:  strings.Replace(validBDF, "STARTPROPERTIES 6", "STARTPROPERTIES 5", 1),
			want: "declares 5 properties but found 6",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Parse("properties.bdf", strings.NewReader(test.bdf))
			if test.want == "" {
				if err != nil {
					t.Fatalf("Parse() error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Parse() error = %v, want substring %q", err, test.want)
			}
		})
	}
}

func TestParseRejectsMisplacedStructuralDirectives(t *testing.T) {
	tests := []struct {
		name string
		bdf  string
		want string
	}{
		{
			name: "glyph directive outside glyph",
			bdf:  strings.Replace(validBDF, "CHARS 4", "ENCODING 65\nCHARS 4", 1),
			want: "ENCODING is not allowed outside a glyph",
		},
		{
			name: "font directive inside glyph",
			bdf:  strings.Replace(validBDF, "STARTCHAR Z", "STARTCHAR Z\nFONT misplaced", 1),
			want: "FONT is not allowed inside glyph",
		},
		{
			name: "structural directive inside properties",
			bdf:  strings.Replace(validBDF, "ENDPROPERTIES", "CHARS 4\nENDPROPERTIES", 1),
			want: "CHARS is not allowed inside STARTPROPERTIES",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Parse("misplaced.bdf", strings.NewReader(test.bdf))
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Parse() error = %v, want substring %q", err, test.want)
			}
		})
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name string
		bdf  string
		want string
	}{
		{
			name: "STARTFONT missing",
			bdf:  strings.Replace(validBDF, "STARTFONT 2.1\n", "", 1),
			want: "STARTFONT",
		},
		{
			name: "FONT_ASCENT missing",
			bdf: strings.Replace(
				strings.Replace(validBDF, "STARTPROPERTIES 6", "STARTPROPERTIES 5", 1),
				"FONT_ASCENT 10\n", "", 1),
			want: "FONT_ASCENT missing",
		},
		{
			name: "FONT_DESCENT missing",
			bdf: strings.Replace(
				strings.Replace(validBDF, "STARTPROPERTIES 6", "STARTPROPERTIES 5", 1),
				"FONT_DESCENT 3\n", "", 1),
			want: "FONT_DESCENT missing",
		},
		{
			name: "CHARSET_REGISTRY missing",
			bdf: strings.Replace(
				strings.Replace(validBDF, "STARTPROPERTIES 6", "STARTPROPERTIES 5", 1),
				"CHARSET_REGISTRY \"iso10646\"\n", "", 1),
			want: "CHARSET_REGISTRY missing",
		},
		{
			name: "non-Unicode registry",
			bdf:  strings.Replace(validBDF, "\"iso10646\"", "\"JISX0208\"", 1),
			want: "only ISO10646 is supported",
		},
		{
			name: "CHARS mismatch",
			bdf:  strings.Replace(validBDF, "CHARS 4", "CHARS 5", 1),
			want: "CHARS declares 5 glyphs but found 4",
		},
		{
			name: "duplicate encoding",
			bdf:  strings.Replace(validBDF, "ENCODING 65", "ENCODING 90", 1),
			want: "duplicate ENCODING",
		},
		{
			name: "Unicode out of range",
			bdf:  strings.Replace(validBDF, "ENCODING 90", "ENCODING 1114112", 1),
			want: "outside the Unicode range",
		},
		{
			name: "surrogate",
			bdf:  strings.Replace(validBDF, "ENCODING 90", "ENCODING 55296", 1),
			want: "Unicode surrogate",
		},
		{
			name: "two-value encoding",
			bdf:  strings.Replace(validBDF, "ENCODING 90", "ENCODING -1 90", 1),
			want: "two-value ENCODING",
		},
		{
			name: "DWIDTH Y",
			bdf:  strings.Replace(validBDF, "DWIDTH 8 0", "DWIDTH 8 1", 1),
			want: "DWIDTH Y value must be 0",
		},
		{
			name: "negative BBX width",
			bdf:  strings.Replace(validBDF, "BBX 8 2 0 -1", "BBX -1 2 0 -1", 1),
			want: "must not be negative",
		},
		{
			name: "negative BBX height",
			bdf:  strings.Replace(validBDF, "BBX 8 2 0 -1", "BBX 8 -1 0 -1", 1),
			want: "must not be negative",
		},
		{
			name: "BITMAP rows short",
			bdf:  strings.Replace(validBDF, "81\n7e\nENDCHAR", "81\nENDCHAR", 1),
			want: "BITMAP has 1 rows",
		},
		{
			name: "BITMAP rows excessive",
			bdf:  strings.Replace(validBDF, "81\n7e\nENDCHAR", "81\n7e\n00\nENDCHAR", 1),
			want: "BITMAP has 3 rows",
		},
		{
			name: "BITMAP row byte count",
			bdf:  strings.Replace(validBDF, "aa80", "aa", 1),
			want: "BITMAP row has 1 byte; glyph width 9 requires 2",
		},
		{
			name: "invalid hexadecimal",
			bdf:  strings.Replace(validBDF, "aa80", "gggg", 1),
			want: "invalid hexadecimal BITMAP row",
		},
		{
			name: "ENDCHAR missing",
			bdf:  strings.Replace(validBDF, "ENDCHAR\n\nSTARTCHAR A", "\nSTARTCHAR A", 1),
			want: "ENDCHAR missing",
		},
		{
			name: "ENDFONT missing",
			bdf:  strings.Replace(validBDF, "ENDFONT\n", "", 1),
			want: "ENDFONT missing",
		},
		{
			name: "duplicate required property",
			bdf:  strings.Replace(validBDF, "FONT_ASCENT 10\n", "FONT_ASCENT 10\nFONT_ASCENT 10\n", 1),
			want: "duplicate FONT_ASCENT",
		},
		{
			name: "numeric overflow",
			bdf:  strings.Replace(validBDF, "SIZE 16 75 75", "SIZE 32768 75 75", 1),
			want: "does not fit int16",
		},
		{
			name: "content after ENDFONT",
			bdf:  validBDF + "FONT trailing\n",
			want: "content after ENDFONT",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Parse("input.bdf", strings.NewReader(test.bdf))
			if err == nil {
				t.Fatal("Parse() succeeded, want error")
			}
			message := err.Error()
			if !strings.Contains(message, "input.bdf:") {
				t.Errorf("error %q does not contain filename and line prefix", message)
			}
			if !strings.Contains(message, test.want) {
				t.Errorf("error = %q, want substring %q", message, test.want)
			}
		})
	}
}
