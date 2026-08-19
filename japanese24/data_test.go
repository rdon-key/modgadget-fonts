package japanese24

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestEmbeddedFonts(t *testing.T) {
	tests := []struct {
		name string
		data string
		size int
		hash string
	}{
		{"JISX0208", jisx0208Data, 401902, "549a214c5e002a410616b3c5929056d910d67e865de72193a92fe0a6d4425810"},
		{"CP932Ext", cp932ExtData, 5096, "c469eef4ad7411b64b53dcbde5e338779d70dc6353eb292c639ef3c1175c111b"},
	}
	for _, test := range tests {
		if len(test.data) != test.size {
			t.Errorf("%s data size=%d", test.name, len(test.data))
		}
		if hash := fmt.Sprintf("%x", sha256.Sum256([]byte(test.data))); hash != test.hash {
			t.Errorf("%s SHA-256=%s", test.name, hash)
		}
	}
	for name, font := range map[string]interface{ Valid() bool }{
		"JISX0208": JISX0208, "CP932Ext": CP932Ext, "Font": Font,
	} {
		if !font.Valid() {
			t.Errorf("%s is invalid", name)
		}
	}
	if got := Font.Metrics().LineHeight(); got != 25 {
		t.Fatalf("line height=%d", got)
	}
	for _, r := range []rune{'\u65e5', '\u4f03'} {
		if !Font.HasGlyph(r) {
			t.Fatalf("missing U+%04X", r)
		}
	}
	if !CP932Ext.HasGlyph('\u4f03') {
		t.Fatalf("CP932 extension missing U+%04X", '\u4f03')
	}
}

func TestCachedGlyphAccessAllocations(t *testing.T) {
	Font.HasGlyph('\u65e5')
	if got := testing.AllocsPerRun(100, func() { _ = Font.HasGlyph('\u65e5') }); got != 0 {
		t.Fatalf("allocations=%v", got)
	}
}
