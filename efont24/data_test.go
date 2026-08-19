package efont24

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestEmbeddedFont(t *testing.T) {
	if len(data) != 1515662 {
		t.Fatalf("data size=%d", len(data))
	}
	if hash := fmt.Sprintf("%x", sha256.Sum256([]byte(data))); hash != "87aae3c259f37b48b574fb1ff2942c04736b84b942bec6783d9d70425a852952" {
		t.Fatalf("SHA-256=%s", hash)
	}
	if !Font.Valid() || Font.Metrics().LineHeight() != 24 {
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
