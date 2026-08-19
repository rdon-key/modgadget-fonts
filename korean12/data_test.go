package korean12

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestEmbeddedFont(t *testing.T) {
	if len(data) != 234777 {
		t.Fatalf("data size=%d", len(data))
	}
	if hash := fmt.Sprintf("%x", sha256.Sum256([]byte(data))); hash != "c5c29cd8cdeeb282210a4a909dc1e1df8d315936c35f40415907430998eb89d2" {
		t.Fatalf("SHA-256=%s", hash)
	}
	if !Font.Valid() || Font.Metrics().LineHeight() != 13 {
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
