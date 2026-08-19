// Package chinese16 provides the embedded Chinese 16 MGZ font.
package chinese16

import (
	_ "embed"

	"github.com/rdon-key/modgadget"
)

//go:embed chinese16.mgz
var data string

// Font is the embedded Chinese 16 font.
var Font modgadget.Font = modgadget.MustOpenMGZ(data)
