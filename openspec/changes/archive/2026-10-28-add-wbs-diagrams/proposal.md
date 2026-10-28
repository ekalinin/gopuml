# Proposal

## Why

WBS diagrams (`@startwbs`) fail today with "wbs diagrams are not supported yet" (issue #30). Seven Vega cases of the conformance report are not supported yet because of it: the five that the issue lists, plus `vega/nonreg/warnings/wbs_old_order` and `vega/nonreg/wbs/star-and-plain-same-stereotype`, which came with the update to 1.2026.9. Upstream's `wbs` package (14 classes) needs no Graphviz: it lays the nodes out itself. Most of its parts outside the package are already ported: `mindmap.IdeaShape` (exported by the mindmap change for this one), `FtileBoxOld`, the style engine with levels, stereotypes and the cascade of starred styles, `SkinParamColors`, `ExtremityArrow`, the warnings and the ATable and InfinitePlan of `asciiverse`. What is missing outside the package is `FtileBoxOld.createWbs` and `setMinimumWidth`, and `Direction.getWBSDirection`.

## What Changes

- WBS diagrams under `@startwbs`, as upstream's `WBSDiagram`, with the whole `wbs` package ported:
  - the nodes (`*`, `+` and `-` markers, with leading spaces and tabs), the multiline form (`**:` ... `;`), the side markers (`<` and `>`), the boxless (`_`) and pseudo (`_` without a label) nodes, background colors (`[#color]`), aliases (`(code)` and `"label" as code`), stereotypes at the end of a label, and `\n` in labels
  - upstream's old order (the color and alias before `_`, `<` and `>`), which still draws the node and adds upstream's warning above the diagram
  - links between aliased nodes (`a -> b`), drawn as arrows on top of the tree, with their color and stereotype. A link label (`a -> b : label`) is read and not drawn, as upstream does
  - upstream's errors: `WBS diagram only accepts one root first declared.`, `Bad tree structure`, `Missing label for WBS node.`, `No such node`, `No such color`, and the crash message of markers that upstream cannot read
  - upstream's layout: the root at the top, its children side by side below it, the deeper nodes in columns, on the left and right of a vertical line, and `Width auto`, which gives the children of a node the width of the widest one
  - the styles of `wbsDiagram` (`node`, `rootNode`, `leafNode`, `boxless`, `arrow`, `:depth(n)`, stereotypes and the cascade of `.name *` to the descendants)
  - the common commands (title, header, footer, caption, legend, mainframe, `skin`, `skinparam`, `<style>`, sprites, `scale`, ...)
- `FtileBoxOld.createWbs` and `setMinimumWidth` in `activitydiagram3`, and `getWBSDirection`.
- Family detection: `wbs` becomes a supported family. The "not supported yet" error for `@startwbs` goes away. `@startjson` becomes the example of a recognized family that is not supported yet.
- WBS diagrams render in SVG and PNG, and in TXT and UTXT as upstream's text output: the root and its children as ASCII boxes, drawn in ASCII for both formats.
- The WASM build renders WBS diagrams fully, because they need no graph layout.
- WBS diagrams can be embedded in creole text with `{{wbs` ... `}}`.
- Conformance: curated cases for every new scenario. The 7 Vega WBS cases are compared instead of being reported as not supported yet.
- The README and the landing page list WBS diagrams in the table of supported diagrams and among the diagrams that do not need Graphviz. The README's conformance counts are updated.

## Capabilities

### New Capabilities

- `wbs-diagram`: WBS diagrams under `@startwbs`, with upstream's syntax, layout, links, styles, errors and text output

### Modified Capabilities

- `diagram-source`: `wbs` becomes a supported family next to `uml`, `packetdiag`, `gantt` and `mindmap`, and the example of a recognized family that is not supported yet becomes `@startjson`
- `output-formats`: WBS diagrams join sequence and gantt diagrams as diagrams with a text output
- `wasm-build`: WBS diagrams join sequence, activity, packet, gantt and mindmap diagrams as diagrams with a built-in layout that render fully in the WASM build

## Impact

- **Code:**
  - a new `internal/wbs` package: the diagram, the factory, the commands, the model (`WElement`, `WBSLink`), the drawing (`Fork`, `ITFComposed`, `ITFLeaf`, `WBSTextBlock`) and the text output
  - `internal/activitydiagram3`: `FtileBoxOld` with `createWbs` and `setMinimumWidth`
  - `internal/engine`: the factory list and the check of the text formats
  - `internal/core`: the supported families
  - tests that use `@startwbs` as their example of a family that is not supported yet (`internal/core`, `internal/errdiagram`, the root package) switch to `@startjson`
- **Runtime dependency:** none. WBS diagrams do not need Graphviz.
- **Conformance:** new curated cases and committed references under `testdata/conformance/`. The not-supported-yet count of the report drops from 29 to 22, unless some WBS cases end up in `known-differences.yaml`.
- **Documentation:** `README.md` and `site/index.html` (the table of supported diagrams and the notes on the diagrams that do not need Graphviz).
- **Ordering:** the delta specs of `diagram-source` and `wasm-build` start from the requirements of `add-mindmap-diagrams`, so that change is archived first.
- **Compatibility:** inputs that failed with "wbs diagrams are not supported yet" now render.
