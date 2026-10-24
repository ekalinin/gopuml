# Proposal

## Why

gopuml targets upstream PlantUML v1.2026.8 (MIT variant). [v1.2026.9](https://github.com/plantuml/plantuml/releases/tag/v1.2026.9) was released on 2026-10-17 and changes the output of supported diagram types: rendered with both jars, 130 of the 306 SVG outputs of the v1.2026.8 Vega corpus differ. Its Vega corpus grows from 305 to 425 cases. Issue #59 asks to move the target to v1.2026.9.

## What Changes

- Pin v1.2026.9: the reference image's jar, the sources jar of `tools/upstream/fetch.sh`, the Vega tag of `tools/corpus/fetch-vega.sh`, the plantuml-server image of `tools/refgen`, and the version that gopuml reports (`--version`, `%version()`, the `<?plantuml ...?>` line of SVG output, the built-in server's `Server` header). Graphviz stays at 14.0.1.
- Import the 120 new Vega cases. `fetch-vega.sh` also copies the zip archive that the `nonreg/group2896` cases import.
- Port the upstream changes that change the output of supported diagram types:
  - sequence: the `class` and `data-*` attributes of lifeline, head, tail and message groups in SVG; a group's right margin equal to its left margin; notes of one participant joined with `&` stacked instead of overlapping; the `/` note merge syntax deprecated with a warning; the participant name aligned when `MinimumWidth` widens the box; new texts for the note position warnings
  - class: role labels with visibility icons, placed away from the line and the classes; the stereotype label of a class styled by `<style>` and the dark theme (with the new default skin rule); a single backslash as namespace separator; `package code as "Display"`
  - state: the `<<junction>>` pseudostate; with node-style transition labels, `hide` and `remove` of a state also drop the label node; stylable history pseudo-states
  - activity: a leading color (`#red:X;`, `#red:if`, `#red:repeat`, ...) is parsed, ignored and reported with upstream's new warnings, and a stereotype color after `if`/`elseif` applies; new texts for the `partition` and `end group` warnings
  - preprocessor: `!include` skips a file already included and `!include_once` reports it; the selector of a local include (`file!1`, `file!ID`) is applied, and an unusable one is reported; a missing `!theme NAME from DIR` reports "Cannot load theme NAME in DIR"; `!import` of a zip archive, which gopuml does not implement today, makes its entries includable; new texts for the `!option` warnings
  - skinparam: new texts for the `handwritten` and `ParticipantPadding` warnings, no warning for `padding`
  - source compression: only lowercase `@startuml` and `@enduml` lines are removed from the encoded source
  - line trimming: `StringUtils.trin` also trims the non-breaking space
- Port the replacements of the upstream files that v1.2026.9 removed and that gopuml translates: the style engine (`StyleQuery`, `StyleIndex`, the new style parser, `Specificity`), `Defines` without `Define`, and `CommandSkinParamMultilines` without `CommandMultilinesComment`. The Go names used across the diagram packages stay.
- Every provenance header names `plantuml-mit 1.2026.9`. Upstream changes that do not change output (`StyleQueries` at every call site, `MyCollections`, precompiled patterns) are not ported. Three refactors that turned out to change output are ported: the order of the groups of a regex result, `formatDecimal` (no `-0`), and the first-token command index, which tries a command only on lines that start with a token it declares.
- Regenerate every reference and golden file recorded from the upstream jar, and update the Java harnesses that use removed upstream classes.
- Update the version in `README.md`, `NOTICE`, `doc.go`, the specs and the known-difference reasons that cite upstream behavior.

Out of scope:

- The changes to diagram types gopuml does not support (salt, gantt, wbs, json, yaml, `@startdot`), C4-PlantUML v2.14.0 in the standard library (#60), the TeaVM and JS engine changes (`PLANTUML_FILE_LOADER`, the C build), and the Smetana fixes (plantuml/plantuml#2903, plantuml/plantuml#2917). The Vega cases that need Smetana are listed as known differences.
- The handwritten drawing, which `skinparam handwritten` and `!option handwritten` need. The Vega case `nonreg/warnings/skinparam_handwritten` is listed as a known difference.
- A depth limit for recursive includes. A file that includes itself with `!include_many` still overflows the stack, in upstream as in gopuml.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `upstream-compatibility`: the pinned version and the Vega corpus move to v1.2026.9, and the corpus includes the files its cases import
- `sequence-diagram`: SVG element groups, group margins, stacked notes and the `/` deprecation, participant text alignment, note position warnings, and the version of the Teoz requirement
- `class-diagram`: role labels, the single backslash separator and `package code as "Display"`
- `state-diagram`: the junction pseudostate, node-style labels on hide and remove, and stylable history pseudo-states
- `activity-diagram`: leading colors and stereotype colors of actions, conditionals and loops, and the grouping warnings
- `preprocessor`: include deduplication, include selectors, zip imports, `!option` warnings, and the version of the builtin requirement
- `styling`: the class stereotype label style, the skinparam warnings, the missing local theme error, and the version of the theme requirement
- `url-encoding`: only lowercase start and end lines are removed, and the version of the round-trip scenario
- `diagram-source`: the version of the family detection requirement
- `cli`: the version of the flag and `--version` requirements
- `release`: the version printed by a release build

## Impact

- **Code:** `internal/sequencediagram`, `internal/klimt/rose`, `internal/svek` (with a new `rolelabels.go`), `internal/abel`, `internal/statediagram`, `internal/activitydiagram3`, `internal/uml`, `internal/klimt` (style engine, skin, warnings), `internal/tim` (includes, selectors, zip import, themes, options, `Defines`), `internal/code`, `internal/text` and the other ports of `StringUtils.trin`, `internal/version`, `internal/server`. The provenance headers of 641 Go files change (1,703 header lines).
- **Tooling:** `tools/reference/Dockerfile`, `tools/upstream/fetch.sh` and `checksums.sha256`, `tools/corpus/fetch-vega.sh`, `tools/refgen`, `tools/provenance`, and the Java golden harnesses under `internal/*/testdata`.
- **Conformance:** all references are regenerated. 120 new Vega cases, of which 15 Smetana cases and the handwritten case are expected as known differences, and the cases of unsupported diagram types are reported as not supported yet. New curated cases for the new scenarios.
- **Compatibility:** output changes as upstream's does: sequence SVG gains group attributes and some layouts move, role labels and class layouts with roles change, `!include` of the same file twice includes it once, `!include_once` of a file already included is an error, deprecated activity colors are ignored with a warning.
- **Dependencies:** none new. The zip import uses Go's `archive/zip`.
