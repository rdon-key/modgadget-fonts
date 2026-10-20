// Package efont16zerokb02 embeds a compact Efont Biwidth 16 MGZ subset
// covering the Traditional Chinese Zero-KB02 demo greeting.
//
// The subset preserves original compressed MGZ1 blocks; no font glyphs
// are regenerated or synthesized.
package efont16zerokb02

import (
	_ "embed"

	"github.com/rdon-key/modgadget"
)

//go:embed efont16-zerokb02.mgz
var data string

// Font contains 448 Efont Biwidth 16 glyphs, including 恭喜進軍台灣！.
var Font modgadget.Font = modgadget.MustOpenMGZ(data)
