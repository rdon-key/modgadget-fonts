package korean24

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestEmbeddedFont(t *testing.T) {
	if len(data) != 405705 {
		t.Fatalf("data size=%d", len(data))
	}
	if hash := fmt.Sprintf("%x", sha256.Sum256([]byte(data))); hash != "2ea4c2a7390b89225d469a90b63f8990704fca6990a522d10a2ba93d40e5ef5f" {
		t.Fatalf("SHA-256=%s", hash)
	}
	if !Font.Valid() || Font.Metrics().LineHeight() != 24 {
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
