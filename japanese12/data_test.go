package japanese12

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
		{"JISX0208", jisx0208Data, 218961, "8cb265cbd6d4e40bd6b256a7949e5fd63954ba99575e58bf5bf673970b686133"},
		{"CP932Ext", cp932ExtData, 3516, "ad1b1001f55f2fddf5da58de58f5627387937dcde6f01ba5cd7969586189ef75"},
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
	if got := Font.Metrics().LineHeight(); got != 12 {
		t.Fatalf("line height=%d", got)
	}
	for _, r := range []rune{'A', '\u65e5'} {
		if !Font.HasGlyph(r) {
			t.Fatalf("missing U+%04X", r)
		}
	}
	if !CP932Ext.HasGlyph('A') {
		t.Fatalf("CP932 extension missing U+%04X", 'A')
	}
}

func TestCachedGlyphAccessAllocations(t *testing.T) {
	Font.HasGlyph('\u65e5')
	if got := testing.AllocsPerRun(100, func() { _ = Font.HasGlyph('\u65e5') }); got != 0 {
		t.Fatalf("allocations=%v", got)
	}
}
