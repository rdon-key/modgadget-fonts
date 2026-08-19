package korean16

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestEmbeddedFont(t *testing.T) {
	if len(data) != 276748 {
		t.Fatalf("data size=%d", len(data))
	}
	if hash := fmt.Sprintf("%x", sha256.Sum256([]byte(data))); hash != "09506f9e73c1e327e940fe78045bd3a3c6952bf63306476b62cc9999b8789924" {
		t.Fatalf("SHA-256=%s", hash)
	}
	if !Font.Valid() || Font.Metrics().LineHeight() != 16 {
		t.Fatalf("valid=%v metrics=%+v", Font.Valid(), Font.Metrics())
	}
	for _, r := range []rune{'\ud55c', '\uae00'} {
		if !Font.HasGlyph(r) {
			t.Fatalf("missing U+%04X", r)
		}
	}
}

func TestCachedGlyphAccessAllocations(t *testing.T) {
	Font.HasGlyph('\ud55c')
	if got := testing.AllocsPerRun(100, func() { _ = Font.HasGlyph('\ud55c') }); got != 0 {
		t.Fatalf("allocations=%v", got)
	}
}
