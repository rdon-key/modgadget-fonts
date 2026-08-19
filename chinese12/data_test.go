package chinese12

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestEmbeddedFont(t *testing.T) {
	if len(data) != 225107 {
		t.Fatalf("data size=%d", len(data))
	}
	if hash := fmt.Sprintf("%x", sha256.Sum256([]byte(data))); hash != "5b1f51a7e6567c0a12e7b25969e561d89d91d30f570644807fe62044dce134ef" {
		t.Fatalf("SHA-256=%s", hash)
	}
	if !Font.Valid() || Font.Metrics().LineHeight() != 15 {
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
