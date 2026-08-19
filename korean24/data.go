// Package korean24 provides the embedded Korean 24 MGZ font.
package korean24

import (
	_ "embed"

	"github.com/rdon-key/modgadget"
)

//go:embed korean24-ksx1001-partial.mgz
var data string

// Font is the embedded Korean 24 font.
var Font modgadget.Font = modgadget.MustOpenMGZ(data)
