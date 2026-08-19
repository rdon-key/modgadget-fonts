// Package korean16 provides the embedded Korean 16 MGZ font.
package korean16

import (
	_ "embed"

	"github.com/rdon-key/modgadget"
)

//go:embed korean16-ksx1001-partial.mgz
var data string

// Font is the embedded Korean 16 font.
var Font modgadget.Font = modgadget.MustOpenMGZ(data)
