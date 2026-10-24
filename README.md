# gopuml

A Go reimplementation of [PlantUML](https://plantuml.com) v1.2026.8 (MIT
variant). It reads the same diagram language and produces the same SVG, PNG
and text output, and it works as:

- a command line that replaces `java -jar plantuml.jar`
- a Go library
- an HTTP server compatible with PlantUML's built-in server and the main
  plantuml-server routes
- a WebAssembly module for browsers and Node.js

No JVM is needed. Graphviz `dot` is needed only for class, description
(use case, component, deployment), state and legacy activity diagrams.

The [landing page](https://ekalinin.github.io/gopuml/) has a
[playground](https://ekalinin.github.io/gopuml/#playground) that renders
sequence and activity diagrams in the browser with the WebAssembly module.

## Supported diagrams and formats

| Diagram type | SVG | PNG | TXT / UTXT |
|---|---|---|---|
| Sequence | yes | yes | yes |
| Class and object | yes (needs `dot`) | yes (needs `dot`) | no |
| Use case, component and deployment | yes (needs `dot`) | yes (needs `dot`) | no |
| State | yes (needs `dot`) | yes (needs `dot`) | no |
| Activity (new syntax) | yes | yes | no |
| Activity (legacy syntax) | yes (needs `dot`) | yes (needs `dot`) | no |
| Packet | yes | yes | no |

The preprocessor (`!include`, `!define`, functions, all 74 builtins, JSON),
themes, skin parameters, styles and creole are supported, including diagrams
embedded in text with `{{` ... `}}`, and so is the PREPROC output
(`--preproc`).

Other diagram types (archimate, timing, gantt, mindmap, and so on) are
reported with an explicit
"... diagrams are not supported yet" error image instead of a wrong drawing.

Compatibility is measured against upstream PlantUML v1.2026.8 with a
conformance suite of 520 cases (upstream's Vega test corpus plus curated
cases): every case passes, apart from 25 documented known differences
(`testdata/conformance/known-differences.yaml`) and 43 cases of diagram types
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
gh release download v0.1.0 --repo ekalinin/gopuml -p '*_darwin_arm64.tar.gz' -p checksums.txt
```

or with `curl`:

```sh
curl -fLO https://github.com/ekalinin/gopuml/releases/download/v0.1.0/gopuml_0.1.0_darwin_arm64.tar.gz
curl -fLO https://github.com/ekalinin/gopuml/releases/download/v0.1.0/checksums.txt
```

then check and unpack it:

```sh
shasum -a 256 --ignore-missing -c checksums.txt   # sha256sum on linux
tar xzf gopuml_0.1.0_darwin_arm64.tar.gz plantuml
./plantuml --version
```

The macOS binaries are not signed. When a browser downloaded the archive,
macOS blocks the binary until its quarantine attribute is removed:

```sh
xattr -d com.apple.quarantine plantuml
```

With Go 1.26 or later, the command line can also be installed from source:

```sh
go install github.com/ekalinin/gopuml/cmd/plantuml@v0.1.0
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
the `files` option. Class, description, state and legacy activity diagrams
report that graph layout is not available in this build.

## Graphviz

Class, object, description, state and legacy activity diagrams are laid out
by Graphviz `dot`, as upstream does.
`dot` is looked up from `--dot-path` (or the library option), then the
`GRAPHVIZ_DOT` environment variable, then `PATH` and the usual install
locations. `./plantuml --check-graphviz` reports the version found. The
references were made with Graphviz 14.0.1. Sequence diagrams, activity
diagrams in the new syntax and packet diagrams do not need Graphviz.

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
image, so that both implementations use the same Graphviz. On arm64 hosts,
`make conformance` renders them natively with the host's Graphviz, so the
results there are indicative only.

## License

MIT, see `LICENSE`. This project translates the MIT-licensed variant of
PlantUML v1.2026.8 (`net.sourceforge.plantuml:plantuml-mit`), (C) Copyright
2009-2024, Arnaud Roques. Every Go file names the upstream file it comes from
in its header. `NOTICE` lists the bundled assets (DejaVu fonts, themes) with
their licenses, the parts of fdlibm the math functions come from, the Go code
that the binaries link, and the upstream components that are not included.
