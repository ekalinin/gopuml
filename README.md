# plantuml-in-go

A Go reimplementation of [PlantUML](https://plantuml.com) v1.2026.8 (MIT
variant). It reads the same diagram language and produces the same SVG, PNG
and text output, and it works as:

- a command line that replaces `java -jar plantuml.jar`
- a Go library
- an HTTP server compatible with PlantUML's built-in server and the main
  plantuml-server routes
- a WebAssembly module for browsers and Node.js

No JVM is needed. Graphviz `dot` is needed only for class diagrams.

## Supported diagrams and formats

| Diagram type | SVG | PNG | TXT / UTXT |
|---|---|---|---|
| Sequence | yes | yes | yes |
| Class and object | yes (needs `dot`) | yes (needs `dot`) | no |
| Activity (new syntax) | yes | yes | no |

The preprocessor (`!include`, `!define`, functions, all 74 builtins, JSON),
themes, skin parameters, styles and creole are supported, and so is the
PREPROC output (`--preproc`).

Other diagram types (state, use case, component, deployment, timing, gantt,
mindmap, the legacy activity syntax, and so on) are reported with an explicit
"... diagrams are not supported yet" error image instead of a wrong drawing.

Compatibility is measured against upstream PlantUML v1.2026.8 with a
conformance suite of 455 cases (upstream's Vega test corpus plus curated
cases): every case passes, apart from 11 documented known differences
(`testdata/conformance/known-differences.yaml`) and 154 cases of diagram types
that are not supported yet. Most SVG outputs are byte-identical to upstream's,
except the encoded source in `<?plantuml-src ...?>`, which uses a different
deflate implementation and decodes to the same text.

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

	"github.com/ekalinin/plantuml-in-go"
)

func main() {
	src := "@startuml\nAlice -> Bob : hello\n@enduml\n"
	res, err := plantuml.Render(context.Background(), src, &plantuml.Options{Format: plantuml.SVG})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("%dx%d %s\n", res.Width, res.Height, res.Description)
	os.WriteFile("diagram.svg", res.Data, 0o644)
}
```

`Render` returns typed errors (`*plantuml.SyntaxError` with the line and
message, `plantuml.ErrGraphvizNotFound`, `plantuml.ErrUnsupportedDiagram`,
...) and, for a syntax error, the error image together with the error.
`plantuml.Inspect` lists the blocks of a source, and `plantuml.Encode` and
`plantuml.Decode` convert sources to and from the text of server URLs.
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
the `files` option. Class diagrams report that graph layout is not available
in this build.

## Graphviz

Class and object diagrams are laid out by Graphviz `dot`, as upstream does.
`dot` is looked up from `--dot-path` (or the library option), then the
`GRAPHVIZ_DOT` environment variable, then `PATH` and the usual install
locations. `./plantuml --check-graphviz` reports the version found. The
references were made with Graphviz 14.0.1. Sequence and activity diagrams do
not need Graphviz.

## Conformance suite

`make conformance` renders every case with this implementation and compares
the outputs with the references of upstream PlantUML, which `make reference`
records in a Docker image (`tools/reference/Dockerfile`). The comparison
writes `conformance-report.md`.

## License

MIT, see `LICENSE`. This project translates the MIT-licensed variant of
PlantUML v1.2026.8 (`net.sourceforge.plantuml:plantuml-mit`), (C) Copyright
2009-2024, Arnaud Roques. Every Go file names the upstream file it comes from
in its header. `NOTICE` lists the bundled assets (DejaVu fonts, themes) with
their licenses, the parts of fdlibm the math functions come from, and the
upstream components that are not included.
