package chinese24

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestEmbeddedFont(t *testing.T) {
	if len(data) != 432248 {
		t.Fatalf("data size=%d", len(data))
	}
	if hash := fmt.Sprintf("%x", sha256.Sum256([]byte(data))); hash != "699ab6c67aa19326d52bba0e325160cf513df89c8d8f9271fb8b7ac9b60232c9" {
		t.Fatalf("SHA-256=%s", hash)
	}
	if !Font.Valid() || Font.Metrics().LineHeight() != 24 {
		t.Fatalf("valid=%v metrics=%+v", Font.Valid(), Font.Metrics())
	}
	for _, r := range []rune{'\u4e2d', '\u56fd'} {
		if !Font.HasGlyph(r) {
			t.Fatalf("missing U+%04X", r)
		}
	}
}

func TestCachedGlyphAccessAllocations(t *testing.T) {
	Font.HasGlyph('\u4e2d')
	if got := testing.AllocsPerRun(100, func() { _ = Font.HasGlyph('\u4e2d') }); got != 0 {
		t.Fatalf("allocations=%v", got)
	}
}
