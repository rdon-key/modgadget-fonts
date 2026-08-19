package spleen8x16

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestEmbeddedFont(t *testing.T) {
	if len(data) != 20134 {
		t.Fatalf("data size=%d", len(data))
	}
	if hash := fmt.Sprintf("%x", sha256.Sum256([]byte(data))); hash != "0252670610fb213da5ea8373358f3189a6aeaed2452fe3492866fb142d1e3d02" {
		t.Fatalf("SHA-256=%s", hash)
	}
	if !Font.Valid() || Font.Metrics().LineHeight() != 16 {
		t.Fatalf("valid=%v metrics=%+v", Font.Valid(), Font.Metrics())
	}
	for _, r := range []rune{' ', 'A', 'M', 'a', '¥', 'é', '\ue0b3'} {
		if !Font.HasGlyph(r) {
			t.Fatalf("missing U+%04X", r)
		}
	}
}

func TestAccessAllocations(t *testing.T) {
	if got := testing.AllocsPerRun(100, func() { _ = Font.HasGlyph('A'); _ = Font.Metrics() }); got != 0 {
		t.Fatalf("allocations=%v", got)
	}
}
