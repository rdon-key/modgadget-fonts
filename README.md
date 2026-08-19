# ModGadget Fonts

ModGadget Fonts distributes generated MGZ bitmap fonts as importable Go
packages for ModGadget applications.

The related repositories have separate responsibilities:

* `modgadget-font-assets` owns the canonical font sources, provenance,
  generation, and validation pipeline.
* `modgadget-fonts` distributes generated MGZ assets as importable Go packages.
* `modgadget` reads MGF and MGZ data and provides font measurement, layout, and drawing.

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
  not to the generated MGZ font data.

* `spleen8x16`: Spleen 8x16
  License: `LICENSES/spleen-BSD-2-Clause.txt`

* `chinese12`: WenQuanYi 9pt, GB2312 partial
  Source: Debian `xfonts-wqy` version `1.0.0~rc1-8`, file `wenquanyi_9pt.pcf`
  License and notice: `LICENSES/xfonts-wqy-debian-copyright`

* `chinese16`, `chinese24`: X.Org Chinese bitmap fonts
  Source: Debian `xfonts-base` version `1:1.0.5+nmu1`, files `gb16st.pcf.gz`
  and `gb24st.pcf.gz`
  License and notice: `LICENSES/xfonts-base-debian-copyright`

* `japanese12`: Shinonome 12 JIS X 0208 core with a Shinonome Roman/kana
  partial CP932 extension
  Source: Debian `xfonts-shinonome` version `1:0.9.11-7`, files
  `shnmk12.pcf.gz` and `shnm6x12r.pcf.gz`
  License and notice: `LICENSES/xfonts-shinonome-debian-copyright`

* `japanese16`, `japanese24`: X.Org JIS X 0208 core with a Japanese
  supplemental partial CP932 extension
  Sources: Debian `xfonts-base` version `1:1.0.5+nmu1`, files
  `jiskan16.pcf.gz` and `jiskan24.pcf.gz`; Debian `xfonts-intl-japanese`
  version `1.4.2-2`, files `jksp16.pcf.gz` and `jksp24.pcf.gz`
  License and notices: `LICENSES/xfonts-base-debian-copyright` and
  `LICENSES/xfonts-intl-japanese-debian-copyright`

* `korean12`: Baekmuk Dotum 12
  Source: Debian `xfonts-baekmuk` version `2.2-9`, file `dotum12.pcf.gz`
  License and notice: `LICENSES/xfonts-baekmuk-debian-copyright`

* `korean16`, `korean24`: X.Org Korean bitmap fonts, KS X 1001 partial
  Source: Debian `xfonts-base` version `1:1.0.5+nmu1`, files
  `hanglm16.pcf.gz` and `hanglm24.pcf.gz`
  License and notice: `LICENSES/xfonts-base-debian-copyright`

Each package exports a `Font` value of type `modgadget.Font`. The raw MGZ data
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

An application or another module can distribute a generated MGZ with the same
small wrapper pattern:

```go
package customfont

import (
        _ "embed"

        "github.com/rdon-key/modgadget"
)

//go:embed custom.mgz
var data string

var Font = modgadget.MustOpenMGZ(data)
```

Generate and validate MGZ assets in `modgadget-font-assets` or an equivalent
asset pipeline. Applications only need the generated `.mgz` data and the small
wrapper shown above.

## License

The source code in this repository is licensed under the BSD 3-Clause License.
See `LICENSE`.

Embedded fonts remain subject to their original licenses, copyrights, and
notice requirements. The applicable license and notice materials for each
distributed font are preserved under `LICENSES/`.
