# Proposal

## Why

Salt diagrams (`@startsalt`) fail today with "salt diagrams are not supported yet" (issue #29). Ten Vega cases of the conformance report are not supported yet because of it: the seven that the issue lists, plus the three cases of `vega/nonreg/group2882` (open droplists), which came with the update to 1.2026.9. Upstream's `salt` package (62 files) needs no Graphviz and almost no styling: it lays the widgets out itself, in black and grays with a fixed font. Its parts outside the package are already ported (`uml.Base`, the common commands, the creole engine, sprites, `LimitFinder`, `UPixel`, `WithDark`, `JavaHashOrder` and the `Skeleton2` of creole trees), apart from `Splitter.purgeAllTag`, which salt uses to count the characters of a widget.

## What Changes

- Salt diagrams under `@startsalt`, as upstream's `PSystemSalt`, with the whole `salt` package ported:
  - the widgets: plain text, buttons (`[OK]`), checkboxes (`[X]`, `[]`), radio buttons (`(X)`, `()`), text fields (`"text"`), droplists (`^item^`, open with `^head^^item 1^^item 2^`), separators (`..`, `==`, `~~`, `--`) and empty cells (`.`), with creole in their labels (bold, icons, sprites, links)
  - the groups: `{` ... `}` with cells split by `|`, nested groups, the table lines of `{#`, `{!`, `{-` and `{+`, the group box with a title (`{^"title"`), and cells that span the next column (`*`)
  - the escapes `\|`, `\{` and `\}` in cells
  - trees and tree tables (`{T`, `{T#`, `+` levels), tabs (`{/`), menus (`{*`), scroll areas (`{S`, `{SI`, `{S-`), the border layout (`N=`, `S=`, `E=`, `W=`, `C=`) and pixel images (`<<name` ... `>>`, reused with `<<name>>`)
  - upstream's two drawing passes, so that open droplists and menus float over what follows, and the larger canvas of an open droplist (issue #2882)
  - the common commands (title, header, footer, caption, legend, mainframe, `skin`, `skinparam`, `<style>`, sprites, `scale`, ...), with sprites kept for the widgets as upstream does, and upstream's quirks: a row that starts with a common command is read as that command, `skinparam` changes only the background and the annotations
  - upstream's errors: `This is not needed anymore.` for a first line `salt`, and the drawing failures (content outside a group, an unclosed group, an unknown image or menu), which return an error and no drawing
- `Splitter.purgeAllTag` with its `htmlTag` pattern in `internal/klimt`.
- The sprite commands keep the sprites of a diagram that stores them itself, as upstream's `PSystemSalt.addSprite` override does.
- Family detection: `salt` becomes a supported family. The "not supported yet" error for `@startsalt` goes away.
- Salt diagrams render in SVG and PNG, and with `--dark-mode` in upstream's dark colors. TXT and UTXT stay unsupported, as for class and timing diagrams.
- The WASM build renders salt diagrams fully, because they need no graph layout.
- Salt diagrams can be embedded in creole text with `{{salt` ... `}}`, for example in notes and activity labels.
- Conformance: curated cases for every new scenario. The 10 Vega salt cases are compared instead of being reported as not supported yet.
- The README and the landing page list salt diagrams in the table of supported diagrams and among the diagrams that do not need Graphviz, and stop naming salt among the types that are not supported yet. The README's conformance counts are updated.

Out of scope:

- The crash image that upstream draws when the drawing of a salt diagram throws. It has a random quote and a Java stack trace, so it cannot be reproduced. The rendering fails with an error instead, as for gantt, WBS and timing diagrams.
- Upstream's generic text output, which TXT and UTXT of salt diagrams fall back to, as class and timing diagrams do.
- Upstream's order of the open lists of several menus, which follows Java's identity hash codes. This change draws them in the order of the menus.

## Capabilities

### New Capabilities

- `salt-diagram`: salt diagrams under `@startsalt`, with upstream's widgets, groups, tables, trees, tabs, menus, scroll areas, border layout, images, sprites, errors and dark colors

### Modified Capabilities

- `diagram-source`: `salt` becomes a supported family next to `uml`, `packetdiag`, `gantt`, `mindmap` and `wbs`, and the examples of recognized families that are not supported yet stop naming `@startsalt`
- `wasm-build`: salt diagrams join sequence, activity, packet, gantt, mindmap, WBS and timing diagrams as diagrams with a built-in layout that render fully in the WASM build

## Impact

- **Code:**
  - a new `internal/salt` package: the diagram and its dictionary, the factory and `CommandAnything`, the tokenizer, the element factories, the elements and their layout
  - `internal/klimt`: `Splitter.purgeAllTag`, and an exported form of the Java `StringTokenizer` with delimiters that `klimt` already has
  - `internal/uml`: the sprite commands call the diagram's own sprite store when it has one
  - `internal/engine`: the factory list
  - `internal/core`: the supported families
- **Runtime dependency:** none. Salt diagrams do not need Graphviz.
- **Conformance:** new curated cases and committed references under `testdata/conformance/`. The not-supported-yet count of the report drops from 19 to 9, unless some salt cases end up in `known-differences.yaml`.
- **Documentation:** `README.md` and `site/index.html` (the table of supported diagrams, the Graphviz notes and the sentence on the types that are not supported yet).
- **Compatibility:** inputs that failed with "salt diagrams are not supported yet" now render, or fail with upstream's error.
