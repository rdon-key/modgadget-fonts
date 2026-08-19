// Package japanese16 provides embedded Japanese 16 MGZ fonts.
package japanese16

import (
	_ "embed"

	"github.com/rdon-key/modgadget"
)

//go:embed japanese16-jisx0208-core.mgz
var jisx0208Data string

//go:embed japanese16-cp932-ext-partial.mgz
var cp932ExtData string

// JISX0208 is the embedded JIS X 0208 core font.
var JISX0208 modgadget.Font = modgadget.MustOpenMGZ(jisx0208Data)

// CP932Ext is the embedded partial CP932 extension font.
var CP932Ext modgadget.Font = modgadget.MustOpenMGZ(cp932ExtData)

// Font combines JISX0208 and CP932Ext.
var Font modgadget.Font = mustFontStack(JISX0208, CP932Ext)

func mustFontStack(primary modgadget.Font, fallbacks ...modgadget.Font) modgadget.Font {
	font, err := modgadget.NewFontStack(primary, fallbacks...)
	if err != nil {
		panic(err)
	}
	return font
}
