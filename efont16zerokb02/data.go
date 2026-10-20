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

// Open validates and opens the embedded MGZ font when the caller requests it.
//
// This deliberately avoids a package-level MustOpenMGZ initializer, allowing
// microcontroller code to initialize USB serial / OLED before parsing the
// font. The file is embedded in firmware as read-only data.
func Open() (modgadget.Font, error) {
	return modgadget.OpenMGZ(data)
}
