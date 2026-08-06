# ModGadget Fonts

ModGadget Fonts distributes generated MGF bitmap fonts as importable Go
packages for ModGadget applications. It does not generate, convert, or modify
font assets.

The related repositories have separate responsibilities:

- `modgadget-font-assets` generates and validates MGF assets.
- `modgadget-fonts` embeds generated MGF assets in Go packages for applications.
- `modgadget` reads MGF data and provides font measurement, layout, and drawing.

## Packages

- `efont16`: Efont Biwidth 16
- `efont24`: Efont Biwidth 24
- `shinonome12`: Shinonome 12
- `spleen8x16`: Spleen 8x16

Each package exports a `Font` value of type `modgadget.Font`. The raw MGF data
is embedded privately and is not part of the package API.

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

## Creating another font package

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

Generate and validate MGF assets in `modgadget-font-assets` (or an equivalent
asset pipeline); this runtime distribution module deliberately contains no
generator.

## License

Repository source code is licensed under the BSD 3-Clause License in
[LICENSE](LICENSE). Each embedded font remains subject to its original font
license, copyright, and notice terms. The applicable original materials are
preserved under [LICENSES](LICENSES).
