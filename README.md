# gopuml

A Go reimplementation of [PlantUML](https://plantuml.com) v1.2026.9 (MIT
variant). It reads the same diagram language and produces the same SVG, PNG
and text output, and it works as:

- a command line that replaces `java -jar plantuml.jar`
- a Go library
- an HTTP server compatible with PlantUML's built-in server and the main
  plantuml-server routes
- a WebAssembly module for browsers and Node.js

No JVM and no Graphviz are needed: class, description (use case, component,
deployment), state, legacy activity and Chen EER diagrams are laid out by
[gogv](https://github.com/ekalinin/gogv), a Go implementation of Graphviz
`dot`.

A run of `make bench` on an Apple M3 Pro gives these times for the upstream
jar (`java -jar plantuml.jar`, with Graphviz `dot`), gopuml with the same
`dot`, and gopuml with gogv, its default. The
[comparison on the landing page](https://ekalinin.github.io/gopuml/#comparison)
has every diagram type and how the times are measured.

| | PlantUML | gopuml (dot) | gopuml (gogv) |
|---|---:|---:|---:|
| Start time (one small sequence diagram) | 706 ms | 14 ms | 13 ms |
| Rendering: packet (17 cases) | 754 ms (44 ms per case) | 18 ms (1.1 ms per case) | 17 ms (1.0 ms per case) |
| Rendering: mindmap (30 cases) | 777 ms (26 ms per case) | 21 ms (0.7 ms per case) | 18 ms (0.6 ms per case) |
| Rendering: salt (39 cases) | 765 ms (20 ms per case) | 25 ms (0.6 ms per case) | 19 ms (0.5 ms per case) |
| Rendering: chen_eer (31 cases) | 861 ms (28 ms per case) | 95 ms (3.1 ms per case) | 20 ms (0.6 ms per case) |

The [landing page](https://ekalinin.github.io/gopuml/) has a
[playground](https://ekalinin.github.io/gopuml/#playground) that renders
every supported diagram type apart from dot diagrams in the browser with the
WebAssembly module.

## Supported diagrams and formats

| Diagram type | SVG | PNG | TXT / UTXT |
|---|---|---|---|
| Sequence | yes | yes | yes |
| Class and object | yes | yes | no |
| Use case, component and deployment | yes | yes | no |
| State | yes | yes | no |
| Activity (new syntax) | yes | yes | no |
| Activity (legacy syntax) | yes | yes | no |
| Packet | yes | yes | no |
| Gantt | yes | yes | yes |
| Mindmap | yes | yes | no |
| WBS | yes | yes | yes |
| Timing | yes | yes | no |
| Salt | yes | yes | no |
| Chen EER | yes | yes | no |
| Dot | yes, in the `INTERNET_WITH_DOTSVG` and `INSECURE` profiles | with a named `dot` | no |

The preprocessor (`!include`, `!define`, functions, all 74 builtins, JSON),
themes, skin parameters, styles and creole are supported, including diagrams
embedded in text with `{{` ... `}}`, and so is the PREPROC output
(`--preproc`).

The standard library of upstream v1.2026.9 is bundled (`!include
<C4/C4_Container>`, `!include <awslib/AWSCommon>` and the other libraries that
`plantuml -stdlib` lists), apart from nine libraries whose license could not
be verified or does not allow bundling them (see `NOTICE`). Sprites are drawn:
monochrome, color, image and SVG sprites defined in a diagram, and the sprites
and images of the standard library.

Other diagram types (archimate, JSON, and so on) are
reported with an explicit
"... diagrams are not supported yet" error image instead of a wrong drawing.

Compatibility is measured against upstream PlantUML v1.2026.9 with a
conformance suite of 988 cases (upstream's Vega test corpus plus curated
cases): every case passes, apart from 55 documented known differences
(`testdata/conformance/known-differences.yaml`) and 5 cases of diagram types
that are not supported yet. Most SVG outputs are byte-identical to upstream's,
except the encoded source in `<?plantuml-src ...?>`, which uses a different
deflate implementation and decodes to the same text.

## Installation

Each [release](https://github.com/ekalinin/gopuml/releases) has an archive of
the command line for linux, macOS and windows on amd64 and arm64
(`gopuml_<version>_<os>_<arch>.tar.gz`, `.zip` on windows), with `LICENSE`,
`NOTICE` and this README, an archive of the WebAssembly module
(`gopuml_<version>_wasm.tar.gz`, with `LICENSE` and `NOTICE` in the releases
after v0.1.0), and `checksums.txt` with the SHA-256 of every archive.
Download the archive of your system and `checksums.txt` from the release page,
or with the GitHub CLI:

```sh
gh release download v0.4.0 --repo ekalinin/gopuml -p '*_darwin_arm64.tar.gz' -p checksums.txt
```

or with `curl`:

```sh
curl -fLO https://github.com/ekalinin/gopuml/releases/download/v0.4.0/gopuml_0.4.0_darwin_arm64.tar.gz
curl -fLO https://github.com/ekalinin/gopuml/releases/download/v0.4.0/checksums.txt
```

then check and unpack it:

```sh
shasum -a 256 --ignore-missing -c checksums.txt   # sha256sum on linux
tar xzf gopuml_0.4.0_darwin_arm64.tar.gz plantuml
./plantuml --version
```

The macOS binaries are not signed. When a browser downloaded the archive,
macOS blocks the binary until its quarantine attribute is removed:

```sh
xattr -d com.apple.quarantine plantuml
```

With Go 1.26 or later, the command line can also be installed from source:

```sh
go install github.com/ekalinin/gopuml/cmd/plantuml@v0.4.0
```

## Building

Go 1.26 or later is needed. In a checkout of this repository:

```sh
go build -o plantuml ./cmd/plantuml
./plantuml --version
```

## Command line

The command line accepts the flags of `java -jar plantuml.jar`. Flags that
this implementation does not provide (for example `--pdf`) are rejected with
exit status 1 instead of being silently ignored.

```sh
printf '@startuml\nAlice -> Bob : hello\n@enduml\n' > diagram.puml

./plantuml diagram.puml              # writes diagram.png (PNG is the default)
./plantuml --svg diagram.puml        # writes diagram.svg
./plantuml --txt diagram.puml        # writes diagram.atxt
./plantuml --svg -o out diagram.puml # writes out/diagram.svg
./plantuml --check-syntax diagram.puml

printf 'Alice -> Bob\n' | ./plantuml -p --svg > pipe.svg
printf 'Alice -> Bob\n' | ./plantuml -syntax

./plantuml --encode-url diagram.puml
./plantuml --extract-source diagram.svg
./plantuml --help
```

Exit statuses are upstream's: 0 on success, 50 when no input file is found,
100 when no diagram is found, 200 when a diagram has an error.

## Library

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ekalinin/gopuml"
)

func main() {
	src := "@startuml\nAlice -> Bob : hello\n@enduml\n"
	res, err := gopuml.Render(context.Background(), src, &gopuml.Options{Format: gopuml.SVG})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("%dx%d %s\n", res.Width, res.Height, res.Description)
	os.WriteFile("diagram.svg", res.Data, 0o644)
}
```

`Render` returns typed errors (`*gopuml.SyntaxError` with the line and
message, `gopuml.ErrGraphvizNotFound`, `gopuml.ErrUnsupportedDiagram`,
...) and, for a syntax error, the error image together with the error.
`gopuml.Inspect` lists the blocks of a source, and `gopuml.Encode` and
`gopuml.Decode` convert sources to and from the text of server URLs.
Includes can be resolved from an `fs.FS` given in the options.

## HTTP server

```sh
./plantuml --http-server:8080
```

It serves `GET /svg/<encoded>`, `/png/<encoded>`, `/txt/<encoded>`,
`/utxt/<encoded>` (also under `/plantuml/`, and the `d`-prefixed dark mode
routes), `POST /render` with a JSON body, `/coder`, `/check`, `/language` and
`/serverinfo`. It runs in the `INTERNET` security profile, limits request
bodies to 1 MiB and renders for at most 60 seconds.

## WebAssembly

```sh
make wasm
```

This writes `plantuml.wasm`, Go's `wasm_exec.js` and the ES module loader
`plantuml.mjs` into `dist/wasm`:

```js
import { load } from './plantuml.mjs';

const plantuml = await load();
const result = await plantuml.render('@startuml\nAlice -> Bob\n@enduml', { format: 'svg' });
// result.data is a Uint8Array; result.error is null or { kind, message, line }.
```

The module runs in the `SANDBOX` security profile: includes come only from
the `files` option. Class, description, state, legacy activity and Chen EER
diagrams are laid out by gogv, as in the native build.

## Graphviz

Class, object, description, state, legacy activity and Chen EER diagrams are
laid out by gogv, a Go implementation of Graphviz `dot` that is built into the
command line, the library and the WebAssembly module, so no Graphviz is
needed. A Graphviz `dot` named by `--dot-path` (or the library option) or by
the `GRAPHVIZ_DOT` environment variable lays them out instead; a `dot` on
`PATH` is not used. Name one to get upstream's exact layout, or orthogonal
lines with `skinparam linetype ortho`, which gogv does not support yet and
draws as splines. `./plantuml --check-graphviz` reports the engine in use.
The references were made with Graphviz 14.0.1; the cases that gogv lays out
differently are listed as known differences. Sequence diagrams, activity
diagrams in the new syntax, packet diagrams, gantt diagrams, mindmap diagrams,
WBS diagrams, timing diagrams and salt diagrams do not need a graph layout.

Dot diagrams (`@startdot`) are drawn by the layout engine itself, as upstream
does: gogv, or the `dot` named by `--dot-path` or `GRAPHVIZ_DOT`. Their SVG
needs the `INTERNET_WITH_DOTSVG` or `INSECURE` security profile
(`PLANTUML_SECURITY_PROFILE`); in the other profiles, which include the
defaults of the command line, the library, the server and the WebAssembly
module, it is upstream's "This feature has been suppressed" notice. Their PNG
needs a named `dot`, as gogv does not write PNG yet
([ekalinin/gogv#33](https://github.com/ekalinin/gogv/issues/33)): without one,
it is upstream's "GraphViz has crashed" image.

## Conformance suite

`make conformance` renders every case with this implementation and compares
the outputs with the references of upstream PlantUML, which `make reference`
records in a Docker image (`tools/reference/Dockerfile`). The comparison
writes `conformance-report.md`.

CI runs the suite on every pull request and every push to `main`, with every
group required: a failing case that is not listed in
`testdata/conformance/known-differences.yaml` fails the run. The report is
attached to the run and shown in its summary. On a linux/amd64 host, these
commands reproduce the CI run:

```sh
make reference
make conformance REQUIRE=all GO_IN_IMAGE=1
```

`GO_IN_IMAGE=1` renders the outputs of this implementation in the reference
image. They are laid out by gogv, not by the Graphviz of the image, so on any
host, arm64 included, a native run gives the results of CI:

```sh
make reference
make conformance REQUIRE=all
```

## Performance check

CI compares the render benchmark of every pull request with that of the
commit it merges into, in its `benchmark` job (`tools/benchcmp`). Both are
built and run in alternating rounds on the same runner, over the curated
cases of the pull request. The table of the medians per diagram type is
shown in the run summary and attached to the run as `benchmark-report`. The
job fails when the median `B/op` or `allocs/op` of a diagram type is more
than 5% above the base; `ns/op` is only shown, as it varies between runs on
shared runners. A pull request that needs more allocations explains the
increase in its description.

This command gives the same table locally, against `main`:

```sh
make bench-compare BASE=main
```

## License

MIT, see `LICENSE`. This project translates the MIT-licensed variant of
PlantUML v1.2026.9 (`net.sourceforge.plantuml:plantuml-mit`), (C) Copyright
2009-2024, Arnaud Roques. Every Go file names the upstream file it comes from
in its header. `NOTICE` lists the bundled assets (DejaVu fonts, themes, the
standard library) with their licenses, the parts of fdlibm the math functions
come from, the Go code that the binaries link, and the upstream components
that are not included.
