// Package chinese24 provides the embedded Chinese 24 MGZ font.
package chinese24

import (
	_ "embed"

	"github.com/rdon-key/modgadget"
)

//go:embed chinese24.mgz
var data string

// Font is the embedded Chinese 24 font.
var Font modgadget.Font = modgadget.MustOpenMGZ(data)
