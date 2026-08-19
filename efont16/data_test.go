package efont16

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestEmbeddedFont(t *testing.T) {
	if len(data) != 876150 {
		t.Fatalf("data size=%d", len(data))
	}
	if hash := fmt.Sprintf("%x", sha256.Sum256([]byte(data))); hash != "c5fcac999920fafdc6b7d421ece765805c203e634f16a9a4d6f344fae7a2d88d" {
		t.Fatalf("SHA-256=%s", hash)
	}
	if !Font.Valid() || Font.Metrics().LineHeight() != 16 {
		t.Fatalf("valid=%v metrics=%+v", Font.Valid(), Font.Metrics())
	}
	for _, r := range []rune{' ', 'A', '\\', 'あ', 'ア', '日', '￥'} {
		if !Font.HasGlyph(r) {
			t.Fatalf("missing U+%04X", r)
		}
	}
}

func TestAccessAllocations(t *testing.T) {
	if got := testing.AllocsPerRun(100, func() { _ = Font.HasGlyph('あ'); _ = Font.Metrics() }); got != 0 {
		t.Fatalf("allocations=%v", got)
	}
}
