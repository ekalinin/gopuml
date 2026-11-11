# Proposal

## Why

JSON (`@startjson`), YAML (`@startyaml`) and HCL (`@starthcl`) blocks fail today with "... diagrams are not supported yet" (issues #39, #40 and #41). Upstream draws all three with one renderer, `jsondiagram.JsonDiagram`: its YAML and HCL factories only parse their syntax into the `JsonValue` that the JSON factory reads, and that JSON parser is already ported (`internal/json`). The Vega corpus case `vega/nonreg/json/style-highlight-with-style-tags` is reported as not supported yet, `vega/nonreg/yaml/theme-with-empty-style-block` is a known difference for the amiga theme, which is not bundled, and no curated case has these diagram types.

## What Changes

- JSON diagrams under `@startjson`, as upstream's `JsonDiagramFactory`, `StyleExtractor`, `JsonDiagram`, `TextBlockJson`, `SmetanaForJson`, `JsonCurve`, `Arrow` and `Mirror`:
  - the `<style>` blocks and the `title`, `scale`, `skin`, `skinparam`, `!pragma`, `!assume` and `hide` lines before the data are taken out, as `StyleExtractor` does. `#highlight` lines are highlight paths, other lines that start with `#` are dropped, and the remaining lines are the JSON text
  - every object and array is a box of rows, keys in bold beside their values for an object, one column for an array, with `true`, `false` and `null` drawn as `☑ true`, `☐ false` and `␀`. A nested object or array is a box of its own to the right, linked by a dashed arrow from a dot on its row
  - highlight paths (`#highlight "a" / "b"`), with `*` for any key, `**` for any depth and a stereotype that selects a style, drawn with the `highlight` style
  - the `jsonDiagram` styles (`node`, `header`, `highlight`, `separator`, `arrow`), `title`, `scale` and `skin`
  - a root that is a string, a number, a boolean or `null` is drawn as an array of one row, and an empty root object or array as one empty row
  - text that is not JSON, and an empty block, draw upstream's image with the text `Your data does not sound like JSON data`, without an error, as upstream does
  - the boxes are placed as upstream's Smetana port places them: this release gives gogv the same graph, with the tables of empty cells that gogv supports in place of Smetana's record shapes (see design)
- YAML diagrams under `@startyaml`, as upstream's `YamlDiagramFactory` and its `yaml.parser` package: mappings, sequences of values and of mappings, flow sequences (`[a, b]`), `|` block values, plain values continued on the next lines and `#` comments, drawn as JSON diagrams with the `yamlDiagram` styles. Lines that upstream's parser rejects draw the text `Your data does not sound like YAML data`
- HCL diagrams under `@starthcl`, as upstream's `HclDiagramFactory` and `HclParser`: blocks, attributes with `=` or `:`, lists, objects and function calls, drawn as JSON diagrams with the `yamlDiagram` styles, every value as a string. A source of one block is drawn as the body of that block. `<style>` blocks and `title` are ignored, as upstream does. Input that upstream's parser rejects draws the text `Your data does not sound like HCL data`
- Family detection: `json`, `yaml` and `hcl` become supported families, and the syntax check reports them as `JSON (Json)`, `YAML (Yaml)` and `HCL (HCL)`
- Embedded `{{json` and `{{yaml` diagrams in creole text are drawn instead of the "not supported yet" error image
- TXT and UTXT of the three types stay unsupported, as for the other diagram types without a text output
- In the WASM build, the three types render as in the native build
- Conformance: curated cases for every compared scenario, `.go.yaml` cases for the others
- The README and the landing page list the three types in the table of supported diagrams. The README's conformance counts are updated

Out of scope:

- The handwritten drawing of `skinparam handwritten true` (`UGraphicHandwritten`), which no diagram type of this release has: the diagram is drawn without it.
- TXT and UTXT, which upstream draws with `UGraphicTxt`, not ported yet.
- Upstream's Smetana port itself: gogv lays the graph out, within a few points of Smetana's positions.
- `SimpleYamlParser` and `YamlLines`, which upstream does not call.
- The JSON diagrams of upstream's TeaVM build (`PSystemBuilderJs`).

## Capabilities

### New Capabilities

- `json-diagram`: JSON diagrams under `@startjson` and the drawing that YAML and HCL diagrams share: the boxes and their rows, the layout, the highlight paths, the styles, the title and scale, the image for data that cannot be read, and the output formats
- `yaml-diagram`: YAML diagrams under `@startyaml`: the YAML that upstream's parser reads, its errors and its styles
- `hcl-diagram`: HCL diagrams under `@starthcl`: the HCL that upstream's parser reads, its errors, its styles, and the ignored `<style>` and `title`

### Modified Capabilities

- `diagram-source`: `json`, `yaml` and `hcl` become supported families, and the example of a recognized but unsupported family changes from `@startjson` to `@startnwdiag`
- `styling`: embedded `{{json` and `{{yaml` diagrams are drawn, and the example of an embedded diagram of an unsupported type changes from `{{json` to `{{nwdiag`
- `wasm-build`: JSON, YAML and HCL diagrams render in the WASM build, laid out by gogv as in the native build

## Impact

- **Code:**
  - a new `internal/jsondiagram` package: `jsondiagram` and, as the two upstream packages import each other, `yaml.Highlighted` and `yaml.YamlDiagramFactory`
  - a new `internal/yamlparser` package (`yaml.parser`) and a new `internal/hcl` package (`hcl`)
  - `internal/dot`: gogv's output in a format other than SVG, without the merged warnings
  - `internal/core`: the supported families; `internal/engine`: the factory list
- **Runtime dependency:** none. gogv lays the graph out, also when a `dot` is named, as upstream never runs Graphviz for these diagrams.
- **Conformance:** new curated cases under `testdata/conformance/cases/json-diagram/`, `yaml-diagram/`, `hcl-diagram/` and `styling/`, with committed references, and `.go.yaml` cases. The report gets the groups `json`, `yaml` and `hcl`. The Vega JSON case is drawn, so the not-supported-yet count goes from 5 to 4. The embedded-diagram case is an `@startuml` case, so the golden dumps of `internal/sequencediagram`, `internal/classdiagram` and `internal/activitydiagram3` are regenerated.
- **Documentation:** `README.md` and `site/index.html`: rows for JSON, YAML and HCL in the table of supported diagrams, the sentence on the other diagram types, and the README's conformance counts.
- **Compatibility:** inputs that failed with "json diagrams are not supported yet", "yaml diagrams are not supported yet" or "hcl diagrams are not supported yet" now give upstream's drawing.
