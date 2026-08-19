package chinese16

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestEmbeddedFont(t *testing.T) {
	if len(data) != 296064 {
		t.Fatalf("data size=%d", len(data))
	}
	if hash := fmt.Sprintf("%x", sha256.Sum256([]byte(data))); hash != "4909184d8156362539ad8676d3383e640f12566677f31171b92a36ef9c41789c" {
		t.Fatalf("SHA-256=%s", hash)
	}
	if !Font.Valid() || Font.Metrics().LineHeight() != 16 {
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
