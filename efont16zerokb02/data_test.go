package efont16zerokb02

import "testing"

func TestGreetingGlyphs(t *testing.T) {
	font, err := Open()
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if !font.Valid() {
		t.Fatal("font is invalid")
	}
	const greeting = "恭喜進軍台灣！"
	for _, r := range greeting {
		if !font.HasGlyph(r) {
			t.Errorf("missing glyph %q U+%04X", r, r)
		}
	}
	if got, want := len(data), 17001; got != want {
		t.Errorf("MGZ size = %d; want %d", got, want)
	}
}
