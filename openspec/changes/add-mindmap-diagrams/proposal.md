# Proposal

## Why

Mindmap diagrams (`@startmindmap`) fail today with "mindmap diagrams are not supported yet" (issue #31). Four Vega cases of the conformance report are not supported yet because of it. Upstream's `mindmap` package (18 classes) is small and needs no Graphviz: it lays the nodes out itself. Its parts outside the package are already ported (`uml.Base`, the common commands, `CommandRankDir`, the style engine with levels, stereotypes and the cascade of starred styles, `SkinParamColors`, the sheet blocks), apart from `FtileBoxOld` of `activitydiagram3`, which draws its boxes.

The issue says that `mindmap` uses `wbs.WElement`. In upstream's sources, `mindmap` does not import `wbs`: `WElement` is only named in a `@DuplicateCode` annotation of `Idea`. The dependency goes the other way (`wbs` uses `mindmap.IdeaShape`), so mindmap diagrams can be ported without WBS diagrams (#30).

## What Changes

- Mindmap diagrams under `@startmindmap`, as upstream's `MindMapDiagram`, with the whole `mindmap` package ported:
  - the nodes in org-mode syntax (`*` or `#` markers, with leading spaces and tabs), in the `+` and `-` syntax (right and left sides), in multiline form (`**:` ... `;`), and the root command `0`
  - background colors (`[#color]`), boxless nodes (`_`), stereotypes at the end of a label, and `\n` in labels
  - the side of the next nodes (`left side`, `right side`, ...) and the direction of the drawing (`top to bottom direction`)
  - upstream's indentation rules and errors (`Bad indentation`, `error42L`, `No such color`, and the crash message of markers that upstream cannot read), and several mindmaps in one diagram when a new root follows the first one
  - upstream's layout: the root in the middle, the right and left branches, the children of a node packed and centered next to it, the curved links
  - the styles of `mindmapDiagram` (`node`, `rootNode`, `leafNode`, `boxless`, `arrow`, `:depth(n)`, stereotypes and the cascade of `.name *` to the descendants)
  - the common commands (title, header, footer, caption, legend, mainframe, `skin`, `skinparam`, `<style>`, sprites, `scale`, ...)
- `FtileBoxOld` of upstream's `activitydiagram3`, which draws the boxes of the nodes.
- Family detection: `mindmap` becomes a supported family. The "not supported yet" error for `@startmindmap` goes away. `@startwbs` becomes the example of a recognized family that is not supported yet.
- Mindmap diagrams render in SVG and PNG. TXT and UTXT are rejected, as for the other diagram types except sequence and gantt.
- The WASM build renders mindmap diagrams fully, because they need no graph layout.
- Mindmap diagrams can be embedded in creole text with `{{mindmap` ... `}}`.
- Conformance: curated cases for every new scenario. The 4 Vega mindmap cases are compared instead of being reported as not supported yet.
- The README and the landing page list mindmap diagrams in the table of supported diagrams and stop naming them among the types that are not supported yet. The README's conformance counts are updated.

Out of scope:

- WBS diagrams (#30). They can reuse `IdeaShape` and `FtileBoxOld` from this change later.
- The TXT and UTXT output of mindmap diagrams, which upstream has. The engine keeps rejecting these formats for every type except sequence and gantt.

## Capabilities

### New Capabilities

- `mindmap-diagram`: mindmap diagrams under `@startmindmap`, with upstream's syntax, layout, styles and errors

### Modified Capabilities

- `diagram-source`: `mindmap` becomes a supported family next to `uml`, `packetdiag` and `gantt`, and the example of a recognized family that is not supported yet becomes `@startwbs`
- `wasm-build`: mindmap diagrams join sequence, activity, packet and gantt diagrams as diagrams with a built-in layout that render fully in the WASM build

## Impact

- **Code:**
  - a new `internal/mindmap` package: the diagram, the factory, the commands, the model (`Idea`, `Branch`, `MindMap`) and the layout (`Finger`, `Tetris`, `StripeFrontier`, `SymetricalTee`)
  - `internal/activitydiagram3`: `FtileBoxOld` with `createMindMap`
  - `internal/engine`: the factory list
  - `internal/core`: the supported families
  - tests that use `@startmindmap` as their example of a family that is not supported yet (`internal/core`, `internal/errdiagram`, the root package) switch to `@startwbs`
- **Runtime dependency:** none. Mindmap diagrams do not need Graphviz.
- **Conformance:** new curated cases and committed references under `testdata/conformance/`. The not-supported-yet count of the report drops from 33 to 29, unless some mindmap cases end up in `known-differences.yaml`.
- **Documentation:** `README.md` and `site/index.html` (the table of supported diagrams, the sentence on the types that are not supported yet and the notes on the diagrams that do not need Graphviz).
- **Compatibility:** inputs that failed with "mindmap diagrams are not supported yet" now render.
