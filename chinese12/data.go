// Package chinese12 provides the embedded Chinese 12 MGZ font.
package chinese12

import (
	_ "embed"

	"github.com/rdon-key/modgadget"
)

//go:embed chinese12-gb2312-partial.mgz
var data string

// Font is the embedded Chinese 12 font.
var Font modgadget.Font = modgadget.MustOpenMGZ(data)
