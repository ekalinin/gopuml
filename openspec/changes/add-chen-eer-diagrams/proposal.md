# Proposal

## Why

Chen EER diagrams (`@startchen`) fail today with "chen_eer diagrams are not supported yet" (issue #32), and the four Vega cases `vega/migration/ChenMovie_Test`, `ChenMovieExtended_Test`, `ChenMovieAlias_Test` and `ChenRankDir_Test` are reported as not supported yet. Upstream's `cheneer` package is small (the diagram, its factory and six commands), and most of what it needs is already ported: `AbstractEntityDiagram`, the Chen leaf types, the bold links and the middle subset and superset signs, the svek layout with Graphviz `dot`, the style names of Chen diagrams and the smallest ellipse around a text. The missing parts are the package itself and the four entity images that svek reports as not ported.

## What Changes

- Chen EER diagrams under `@startchen`, as upstream's `ChenEerDiagram`, laid out with Graphviz `dot` as the class and description diagrams:
  - entities and relationships (`entity CODE {` ... `}`, `relationship CODE {` ... `}`), with a display name (`entity "Name" as CODE {`), a stereotype and colors, drawn as a rectangle and a diamond; `<<weak>>` draws a double rectangle and `<<identifying>>` a double diamond
  - attributes inside an entity, a relationship or a composite attribute (`Name {` ... `}`), with a display name, a stereotype and colors, drawn as an ellipse linked to its owner; `<<key>>` underlines the name, `<<multi>>` draws a double ellipse and `<<derived>>` a dashed one
  - associations with a cardinality (`A -1- B`, `A -N- B`, `A -(N,M)- B`), drawn bold with a `=` on the left (`A =N= B`)
  - subclasses (`A ->- B`, `A -<- B`) with the superset or subset sign in the middle of the link, and subclass groups (`A =>= d { B, C }`, with `d`, `o` or `U`) drawn through a circle with the symbol
  - `left to right direction` and the common commands of upstream's factory (title, header, footer, caption, legend, mainframe, `skin`, `skinparam`, `<style>`, sprites, `scale`, pragmas and the common hide commands)
  - upstream's errors: `Attribute must be inside an entity, relationship or another attribute`, `Attribute already exists`, `No such entity: <name>`, `Unbalanced brackets` and `Bad name`
  - upstream's quirks: the attribute command matches any other line, so `Born : DATE` is one attribute and an unknown command inside a block is an attribute too; only the `=` on the left of an association or a subclass makes it bold; the main color of attributes, associations and subclasses is the line color, and that of entities, relationships and subclass groups is the background color; title and legend styles come from `activityDiagram`, as upstream maps no style name to Chen diagrams
- The four entity images in `internal/svek`: `EntityImageChenEntity`, `EntityImageChenRelationship`, `EntityImageChenAttribute` and `EntityImageChenCircle`, instead of the "not ported" panic of `GeneralImageBuilder`.
- **Removal:** the stub `klimt.TextBlockInEllipse`, which panics with "not ported", and its test. The Chen attribute image uses `klimt.TextBlockInContainingEllipse`, the working port of upstream's `TextBlockInEllipse` that use cases and requirements already use. Nothing else calls the stub.
- Family detection: `chen` becomes a supported family. The "not supported yet" error for `@startchen` goes away.
- Chen diagrams render in SVG and PNG, and with `--dark-mode` in upstream's dark colors. TXT and UTXT stay unsupported, as for class and description diagrams.
- In the WASM build, Chen diagrams return the Graphviz-not-available error, as the other graph-based diagrams.
- Chen diagrams can be embedded in creole text with `{{chen` ... `}}`, for example in notes.
- Conformance: curated cases for every new scenario. The 4 Vega Chen cases are compared instead of being reported as not supported yet.
- The README and the landing page list Chen diagrams in the table of supported diagrams and among the diagrams that need Graphviz. The README's conformance counts are updated.

Out of scope:

- Upstream's generic text output, which TXT and UTXT of Chen diagrams fall back to, as for class and description diagrams.
- A style name of its own for the title and the legend of Chen diagrams, which upstream does not have.

## Capabilities

### New Capabilities

- `chen-eer-diagram`: Chen EER diagrams under `@startchen`, with upstream's entities, relationships, attributes, associations, subclasses, subclass groups, stereotypes, colors, direction, errors and styles, laid out with Graphviz `dot`

### Modified Capabilities

- `diagram-source`: `chen` becomes a supported family next to `uml`, `packetdiag`, `gantt`, `mindmap`, `wbs` and `salt`
- `wasm-build`: Chen diagrams join class, description, state and legacy activity diagrams as graph-based diagrams that return the Graphviz-not-available error in the WASM build

## Impact

- **Code:**
  - a new `internal/cheneer` package: the diagram with its stack of owners, the factory and the six commands
  - `internal/svek`: the four Chen entity images and their cases in `GeneralImageBuilder`
  - `internal/klimt`: the stub `TextBlockInEllipse` and its panic test are removed
  - `internal/engine`: the factory list
  - `internal/core`: the supported families
- **Runtime dependency:** Graphviz `dot`, as for class and description diagrams.
- **Conformance:** new curated cases and committed references under `testdata/conformance/`. The not-supported-yet count of the report drops from 9 to 5, unless some Chen cases end up in `known-differences.yaml`. The embedded case is an `@startuml` case, so the golden dumps of the `@startuml` parsers are regenerated.
- **Documentation:** `README.md` and `site/index.html` (the table of supported diagrams and the Graphviz notes), and the README's WASM note on the diagrams that report that graph layout is not available.
- **Compatibility:** inputs that failed with "chen_eer diagrams are not supported yet" now render, or fail with upstream's error.
