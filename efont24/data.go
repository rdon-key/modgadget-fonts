// Package efont24 provides the embedded Efont Biwidth 24 MGZ font.
package efont24

import (
	_ "embed"

	"github.com/rdon-key/modgadget"
)

//go:embed efont24-full.mgz
var data string

// Font is the embedded full Efont Biwidth 24 font.
var Font modgadget.Font = modgadget.MustOpenMGZ(data)
