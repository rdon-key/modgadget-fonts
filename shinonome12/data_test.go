package shinonome12

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestEmbeddedFont(t *testing.T) {
	if len(data) != 218927 {
		t.Fatalf("data size=%d", len(data))
	}
	if hash := fmt.Sprintf("%x", sha256.Sum256([]byte(data))); hash != "e269ebf7aa12b6757a9b176f68b16a5ade6aeef615d75884093fa8c4a0bb58a1" {
		t.Fatalf("SHA-256=%s", hash)
	}
	if !Font.Valid() || Font.Metrics().LineHeight() != 12 {
		t.Fatalf("valid=%v metrics=%+v", Font.Valid(), Font.Metrics())
	}
	for _, r := range []rune{'\\', '々', 'あ', 'ア', '日', '￥'} {
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
