package bdf

// Size describes the BDF SIZE directive.
type Size struct {
	PointSize   int16
	ResolutionX int16
	ResolutionY int16
}

// Box describes a BDF bounding box in the original BDF coordinate system.
type Box struct {
	Width   int16
	Height  int16
	XOffset int16
	YOffset int16
}

// Font contains the parsed BDF data needed by later conversion stages.
type Font struct {
	Name            string
	Size            Size
	BoundingBox     Box
	Ascent          int16
	Descent         int16
	CharsetRegistry string
	CharsetEncoding string
	Glyphs          []Glyph
}

// Glyph contains one encoded BDF glyph in the original BDF coordinate system.
type Glyph struct {
	Name     string
	Encoding rune
	Advance  int16
	Box      Box
	Bitmap   []byte
}
