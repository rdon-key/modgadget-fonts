package japanese16

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
		{"JISX0208", jisx0208Data, 260897, "9d81089d08490a92a5f76dc82f3b3ccd982b20d55d8b7daeaa61930f5b9f191f"},
		{"CP932Ext", cp932ExtData, 1063, "54b649a499758aa3511ce301e6f7b4633e5176794279a61dc878ae76d63d2a62"},
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
	if got := Font.Metrics().LineHeight(); got != 16 {
		t.Fatalf("line height=%d", got)
	}
	for _, r := range []rune{'\u65e5', '\u4f39'} {
		if !Font.HasGlyph(r) {
			t.Fatalf("missing U+%04X", r)
		}
	}
	if !CP932Ext.HasGlyph('\u4f39') {
		t.Fatalf("CP932 extension missing U+%04X", '\u4f39')
	}
}

func TestCachedGlyphAccessAllocations(t *testing.T) {
	Font.HasGlyph('\u65e5')
	if got := testing.AllocsPerRun(100, func() { _ = Font.HasGlyph('\u65e5') }); got != 0 {
		t.Fatalf("allocations=%v", got)
	}
}
