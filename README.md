# ModGadget Fonts

ModGadget Fonts distributes generated MGF bitmap fonts as importable Go
packages for ModGadget applications.

The related repositories have separate responsibilities:

* `modgadget-font-assets` owns the canonical font sources, provenance,
  generation, and validation pipeline.
* `modgadget-fonts` distributes generated MGF assets as importable Go packages.
* `modgadget` reads MGF data and provides font measurement, layout, and drawing.

## Packages

* `efont16`: Efont Biwidth 16
  License and notice materials: `LICENSES/efont-unicode-bdf-0.4.2/`

* `efont24`: Efont Biwidth 24
  License and notice materials: `LICENSES/efont-unicode-bdf-0.4.2/`

* `shinonome12`: Shinonome 12
  Font data: Public Domain
  Canonical source: Debian `xfonts-shinonome` version `1:0.9.11-7`, file
  `shnmk12.pcf.gz`
  Source/license information: `LICENSES/xfonts-shinonome-debian-copyright`
  The `Files: *` Public-Domain stanza applies to the upstream font data. The
  separate GPL-2+ stanza applies only to the Debian `debian/*` packaging files,
  not to the generated MGF font data.

* `spleen8x16`: Spleen 8x16
  License: `LICENSES/spleen-BSD-2-Clause.txt`

Each package exports a `Font` value of type `modgadget.Font`. The raw MGF data
is embedded privately and is not part of the package API. Applications do not
need to load or manage font files at runtime.

## Usage

```go
import (
        "github.com/rdon-key/modgadget"
        "github.com/rdon-key/modgadget-fonts/efont24"
)

func configure() {
        styles := modgadget.StyleSet{
                Default: modgadget.Style{
                        Font: efont24.Font,
                },
        }

        _ = styles
}
```

## Using a custom font package

An application or another module can distribute a generated MGF with the same
small wrapper pattern:

```go
package customfont

import (
        _ "embed"

        "github.com/rdon-key/modgadget"
)

//go:embed custom.mgf
var data string

var Font = modgadget.MustOpenMGF(data)
```

Generate and validate MGF assets in `modgadget-font-assets` or an equivalent
asset pipeline. Applications only need the generated `.mgf` data and the small
wrapper shown above.

## License

The source code in this repository is licensed under the BSD 3-Clause License.
See `LICENSE`.

Embedded fonts remain subject to their original licenses, copyrights, and
notice requirements. The applicable license and notice materials for each
distributed font are preserved under `LICENSES/`.
