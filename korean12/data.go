// Package korean12 provides the embedded Korean 12 MGZ font.
package korean12

import (
	_ "embed"

	"github.com/rdon-key/modgadget"
)

//go:embed korean12.mgz
var data string

// Font is the embedded Korean 12 font.
var Font modgadget.Font = modgadget.MustOpenMGZ(data)
