// Package spleen8x16 provides the embedded Spleen 8x16 MGZ font.
package spleen8x16

import (
	_ "embed"

	"github.com/rdon-key/modgadget"
)

//go:embed spleen-8x16-full.mgz
var data string

// Font is the embedded full Spleen 8x16 font.
var Font modgadget.Font = modgadget.MustOpenMGZ(data)
