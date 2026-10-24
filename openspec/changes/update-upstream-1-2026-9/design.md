# Design

## Context

See proposal.md for the motivation. The current state that shapes the approach:

- The upstream version is pinned in `tools/reference/Dockerfile` (the jar, twice), `tools/upstream/fetch.sh` and `checksums.sha256` (the binary and sources jars), `tools/corpus/fetch-vega.sh` (the Vega tag and a minimum of 305 cases), `tools/refgen/main.go` (the plantuml-server image), `tools/provenance/main.go` (the header prefix), `internal/version` (`Upstream`, used by `--version`, the SVG `<?plantuml ...?>` line, error images and `%version()`), `internal/tim/preproc_defines.go` (`Version`, the `PLANTUML_VERSION` define) and `internal/server` (`UpstreamVersion`, the `Server` header). CI runs `make provenance` with the sources jar, so every header path must exist in it.
- 641 Go files name 1,497 upstream files in 1,703 `Provenance` lines. Between the two sources jars, 502 of these files changed in code (ignoring imports and line endings), and 22 were removed or moved. Most of the changes do not change output: `StyleSignatureBasic.of(...)` became `StyleQueries` constants at about 220 call sites, commands declare their first tokens for an index, `Collections.unmodifiable*` became `MyCollections`, split regexes are precompiled, and `formatDecimal` replaced `String.format` ("identical results" per upstream).
- The output-changing commits are those listed in the proposal. Three of them are inside refactor commits: `Specificity` replaces the int priority of style values (c9d62097, "fixes a real bug" in the wbs/mindmap cascade), `StyleQuery.withStereotype` puts all labels of a stereotype in one query where `withTOBECHANGED` made one signature per label, and `StringUtils.trin`/`trim2` trim U+00A0 (0c54c010). A fourth, `-D` with no value defining an empty value instead of `"null"` in pipe mode (bd49ca55), needs no change: gopuml already defines an empty value.
- The removed or moved files that gopuml ports are the style engine (`StyleKey`, `StyleSignature`, `StyleSignatureBasic`, `StyleSignatures`, `StyleStorage`, `StringTrie`, `Value*`, `DarkString`, the `CommandStyle*` commands and the whole `style.parser` package), `preproc/Define.java` and `command/CommandMultilinesComment.java`. Their Go ports are 9 files: `internal/klimt/style_{stylesignature,stylebuilder,parser,value,commands}.go`, `internal/tim/preproc_defines.go` and `internal/uml/commands_skinparam.go`.
- Every golden file recorded from the upstream jar embeds the version: SVG goldens of `internal/klimt/svg`, `internal/export`, `internal/errdiagram`, `internal/svek`, the layout references of `internal/activitydiagram3`, the model dumps of `internal/{sequencediagram,classdiagram,activitydiagram3}` and `internal/tim`, and the other `testdata/golden` sets. Their Java harnesses compile against the jar. Five of them use removed classes: `KlimtGolden`, `KlimtGolden2`, `CreoleGolden` (`internal/klimt`), `RoseGolden` (`internal/klimt/rose`) and `BodyGolden` (`internal/abel`).
- The conformance comparison checks outcome, texts, size within 5% and SSIM. It does not see SVG structure, so the sequence group attributes are invisible to it.
- `TestScenarioCoverage` reads the delta specs of active changes, so every new scenario needs a case before `make test` passes.
- Upstream's own image `plantuml/plantuml:1.2026.9` still builds Graphviz 14.0.1.
- gopuml has no zip support: `!import` only checks that the file exists. A file that includes itself crashes the gopuml process with a Go stack overflow (upstream throws `StackOverflowError`).

## Goals / Non-Goals

**Goals:**

- Every reference, golden and conformance case recorded with the 1.2026.9 jar, and gopuml matching them as it matched 1.2026.8: no new failure, and known differences only for features that are not ported.
- Every provenance header naming a 1.2026.9 file whose behavior the Go code reproduces.
- Keep the Go names that the diagram packages use (user decision), so the diff stays in the code that changes behavior.

**Non-Goals:**

- Porting refactors that do not change output, beyond the replacements of removed files (D3).
- Changing the landing page: it describes v0.1.0 and runs its WebAssembly module, built against 1.2026.8. The release that ships this change updates it.

## Decisions

### D1. Pins

- `tools/reference/Dockerfile` takes `plantuml-mit-1.2026.9.jar` (SHA-256 `e84e45b37ad3030cf802e7be18a687d209eca55733a5679b1f1f1a515f7f70dd`) in both stages. The Graphviz stage keeps its current pin, `plantuml/plantuml:1.2026.8`, and its comment says why: 1.2026.9's image builds the same Graphviz 14.0.1, and rebuilt binaries could move references for reasons unrelated to the jar.
- `tools/upstream/fetch.sh` and `checksums.sha256` take the 1.2026.9 jars (sources `88e30e43639555a6963c270bc6245906b81e8cc65140645f416795d0e0bc71c3`, binary `e84e45b37ad3030cf802e7be18a687d209eca55733a5679b1f1f1a515f7f70dd`).
- `tools/corpus/fetch-vega.sh` takes `v1.2026.9`, expects at least 425 cases, and also copies `*.zip` (only `nonreg/group2896/archive.zip`). `refgen` already copies the whole corpus tree to a scratch directory and renders each case from its file, so relative imports resolve. The CI cache key follows the script's hash.
- `tools/refgen` takes `plantuml/plantuml-server:v1.2026.9@sha256:065c66adc6728ba6dd353a29d4ada73f8358b68549a0f710c4bf46b43c103618`.
- `version.Upstream`, `tim.Version` and `server.UpstreamVersion` become `1.2026.9`, with the tests that spell the version.
- Alternative: take Graphviz from the 1.2026.9 image. Rejected: same Graphviz version, more reference noise.

### D2. Provenance headers

- `tools/provenance` checks the prefix `plantuml-mit 1.2026.9`, and every header line changes to it.
- A header names the 1.2026.9 file whose behavior the Go code reproduces. When upstream changed a file only in structure (D5), the Go file keeps its structure and the header names the same path.
- The 9 Go files of the removed or moved files name their replacements:
  - `command/CommandMultilinesComment.java`: dropped from `internal/uml/commands_skinparam.go`, whose `commentSingleLine` now comes from `CommandSkinParamMultilines.java`, already listed
  - `preproc/Define.java`: becomes `preproc/Environment.java` in `internal/tim/preproc_defines.go`
  - `style/CommandStyle*.java`: `command/CommandStyle*.java`
  - `style/{Value,ValueAbstract,ValueColor,ValueImpl,ValueNull,DarkString}.java`: `style/value/*`, with `style/value/Specificity.java`
  - `style/StringTrie.java`: `utils/StringTrie.java`
  - `style/{StyleKey,StyleSignature,StyleSignatureBasic,StyleSignatures}.java`: `style/{StyleQuery,AtomArray,StyleAtom,LevelConstraint}.java`
  - `style/StyleStorage.java`: `style/{StyleIndex,StyleAtomTrie}.java`
  - `style/parser/*`: `style/parser/{RawStyleParser,RawStyleSheet,RawStyleRule,RawSelector,MergedStyleNode,PrioritizedValue}.java` and `style/LegacyStyleFlattener.java`
- `make provenance` with the 1.2026.9 sources checks that every path exists.

### D3. Style engine replacements, with the Go names kept

The style files are ported to the 1.2026.9 engine. The Go names and methods that the diagram packages call stay:

- `StyleSignatureBasic` stays the Go type of a query and becomes the port of `StyleQuery`: a sorted atom array (SName atoms before stereotype atoms) with the level and star constraint of `LevelConstraint`. `StyleSignatureBasicOf`, `AddSName`, `AddStereotype`, `AddLevel`, `AddStar`, `ForStereotypeItself`, `MergedStyle` and the other methods keep their names. `WithTOBECHANGED` follows `StyleQuery.withStereotype`: one query with all the labels of the stereotype. The Go `StyleKey` is folded into the query, as upstream removed it.
- `StyleStorage` becomes the port of `StyleIndex`: matching rules come from a `StyleAtomTrie`, the merged-style cache is shared for queries without a stereotype and kept per `StyleBuilder` for the others (7e92400a).
- The parser becomes `RawStyleParser`, `RawStyleSheet`, `RawStyleRule`, `RawSelector`, `MergedStyleNode` and `PrioritizedValue`, and `LegacyStyleFlattener` turns the merged tree into the `[]*Style` that `StyleBuilder` and `muteStyle` already take (`StyleLoader.parseStyleText`).
- Values carry a `Specificity` (forced override, ancestor cascade, ancestor rank, stereotype count, file order) instead of an int priority. `addPriority` becomes `withAncestorRank`.
- `StyleQueries` is not ported: its constants serve call sites, which keep their `StyleSignatureBasicOf(...)` form.
- Checked by the regenerated `KlimtGolden` style goldens (matching, merging, values), the style Vega fixtures of 1.2026.9 (`nonreg/activity`, `nonreg/stereogroup`, `nonreg/stereotype`) and the whole conformance run.
- Alternatives: keeping the 1.2026.8 engine with headers pointed at the nearest 1.2026.9 files (rejected: the headers would not name the code, and the `Specificity` and `withStereotype` differences would be missed); renaming to `StyleQuery` and porting `StyleQueries` at every call site (rejected with the user: a diff across the 71 files that build queries, for no output change).

### D4. Other removed files

- `internal/tim/preproc_defines.go` keeps its `Defines`, which already holds the environment apart from the values, and names `Defines.java` and `Environment.java`.
- `internal/uml/commands_skinparam.go` drops the `CommandMultilinesComment.java` line (D2).

### D5. Refactors not ported

`StyleQueries` call sites, the first-token command index (`mandatoryFirstTokens*`, `FirstTokens`, `StringLocated.getFirstToken`), `MyCollections`, precompiled split patterns, `formatDecimal`, the per-diagram allocation changes and the shared standard-function registry of the preprocessor. Any output drift they cause shows in the regenerated goldens or the conformance run. A drift found there is ported, with its file, or explained in the task.

### D6. Behavior ports

The specs list the behavior. The design points that are not visible in them:

- **Sequence groups:** `SequenceDiagram.AddMessage` numbers messages (`msg1`, ...) after the parallel-brother step, as upstream does. `Participant` gets `groupTypeHead`, `groupTypeTail` and `groupTypeLifeline`, and `MessageBase` gets `groupTypeMessage`. Only the drawing of the component is wrapped: created participants' heads and multicast arrows stay outside, as upstream does. `klimt/svg` already writes `UGroup` attributes for class diagrams, so it does not change.
- **Stacked notes:** `YGauge` carries the set of note anchors of a run, `NoteTile` stacks a note whose participant and position are already in the set, and `PlayingSpace.findAnchorLivingSpace` resolves notes. The ASCII copy in `PlayingSpaceWithParticipants` stays as upstream left it.
- **Group margin:** `ensureFollowingParticipantClearsFrame` measures to the next lifeline (`posC`) instead of the next box (`posA`).
- **Role labels:** `internal/svek/rolelabels.go` ports `RoleLabels`. `GraphvizImageBuilder.buildImage` reserves the margins of classes with roles before any `SvekEdge` is built, because building an edge caches the class's shield. A class at either end of a role then becomes an HTML table node in the dot input, as upstream's does. `RoleLabels` calls `Math.hypot`, which Java computes with fdlibm, so `internal/jmath` gets a port of fdlibm's `e_hypot.c` with the fdlibm provenance.
- **Class stereotype label:** `EntityImageClassHeader` takes the merged style of `stereotype` under the class with `ForStereotypeItself`. `FromSkinparamToStyle` maps the four `classStereotypeFont*` parameters, and `internal/klimt/skin/plantuml.skin` gets upstream's `classDiagram { class { stereotype { FontSize 12 } } }`.
- **Junction:** `LeafType.JUNCTION` goes right after `STATE_TRANSITION_LABEL`, as in upstream's enum, so that names and orders match. `CircleStart` takes a size of 12 for a junction.
- **Activity colors:** `CommandActivity3` accepts a leading color, so a one-line action with a color no longer reaches the multi-line command, which today swallows the following lines. `CommandIf4` and `CommandElseIf3` use the color of the trailing stereotype. The warning shared today by `Activity3`, `ActivityLong3` and `Repeat3` is split, because only `ActivityLong3` keeps the old text.
- **Include registry:** keyed by the canonical path of a local file and the selector. It is filled after the reader is built and before the lines run, so a file that includes itself with `!include` stops there. Stdlib, URL and archive entries are not recorded, as upstream does.
- **Include selectors:** the selector reaches the block extractor of local files. An index that does not parse, an id that is not a valid pattern, and a selector that selects no block give `cannot include <file>!<selector>`. URL includes keep their current behavior, as upstream's do.
- **Zip import:** `PathSystem` keeps a list of imported archives shared by the path systems derived from it, and `getInputFile` falls back to the first archive that holds the path, from its root. Archives are read with `archive/zip` through the existing file access, so the security profile and the WebAssembly `files` option apply to them. A missing entry reports `cannot include <path>`, as any missing file does. Upstream's message for an entry missing inside an archive embeds a Java exception text with the absolute archive path. No corpus case reaches it.
- **Missing local theme:** `loadFileTheme` returns no theme instead of panicking, and `eater_theme.go` already builds the message.
- **Source compression:** `internal/code` ports `stripStartEnd` and `removeFromToken`: case-sensitive tags, blanks including U+00A0, and the rule for U+0085, U+2028 and U+2029 before the end.
- **Trimming:** the Go ports of `StringUtils.trin` and `trim2` trim U+00A0 too. The ports of Java's `String.trim` keep trimming the characters up to the space. `text.Trin`, documented today as both, follows `StringUtils.trin`. Each trimming call site is checked against the Java code it ports.

### D7. Goldens and harnesses

- Every harness is compiled against the 1.2026.9 jar and run in the reference image, as its header documents. The five harnesses that use removed classes are ported to `StyleQuery`, `StyleLoader.parseStyleText` and `style.value`.
- Every regenerated file is reviewed. Apart from the version line, a difference must come from a listed change. Anything else is a drift (D5).

### D8. Conformance

- `make reference` renders the references with the new image. The committed curated references change.
- A curated case is added for each new scenario. Two scenarios are not compared with Java:
  - "sequence-diagram: Groups of a message": the comparison does not see SVG structure. `TestSVGElementGroups` compares the `<g>` start tags of gopuml's output with those of the committed reference of `sequence-diagram/svg-element-groups`.
  - "url-encoding: Start and end lines in another case": encoded strings are not byte-identical. `TestEncodeCaseVariantDecodedByUpstream` checks the round trip through the upstream jar, as `TestEncodeDecodedByUpstream` does.
- `known-differences.yaml` gets the 15 Vega cases that set `!pragma layout smetana` (`nonreg/group2903`, `group2904`, `group2917`), with the existing Smetana reason, unless they pass, and `vega/nonreg/warnings/skinparam_handwritten`, because the handwritten drawing is not ported. The reasons and the `why` of the `.go.yaml` cases that cite v1.2026.8 behavior are checked against 1.2026.9 and reworded.
- The new cases of unsupported types (salt, gantt, wbs, json, yaml) are reported as not supported yet.

### D9. Order of work

The pins, the version constants, the headers and the regenerated goldens and references come first. The new curated cases come with them, because `TestScenarioCoverage` reads the delta specs. From then until the ports are done, `make test` and `make conformance` fail exactly where the ports are missing. Each port task ends with its packages passing. Nothing is merged before the last task, so main never sees the intermediate state.

### D10. Documentation and specs

- `README.md` (the version, the license line and the conformance counts), `NOTICE` and `doc.go` name v1.2026.9.
- `docs/releases/v0.1.0.md` stays as published. `TestReleaseNotes` accepts any upstream version in the header row of the comparison table, instead of `1.2026.8` only.
- The `Purpose` paragraphs of the main specs that name v1.2026.8 are edited directly, because delta specs do not carry them.
- `add-packet-diagrams` also modifies "Diagram family detection". This change's delta is written on top of it, so `add-packet-diagrams` is archived first.

## Risks / Trade-offs

- **[The style engine port changes merged styles]** → The `KlimtGolden` goldens are regenerated from 1.2026.9 and compared value by value. The style Vega fixtures and the whole conformance run cover the diagram types.
- **[A refactor not ported changes output]** → Every golden and reference is regenerated, and each difference is traced to a listed change (D7).
- **[Role labels change class layouts in user diagrams]** → Upstream's change. The class Vega cases with roles all use Smetana, so the curated "Roles with visibility" case is the only Graphviz comparison. The port follows `RoleLabels` line by line. A prototype on `all_visibilities` without Smetana, with a local Graphviz, came within 1% of upstream's size (255x186 against 253x186).
- **[Include deduplication changes the meaning of existing diagrams]** → Upstream's change. An `!include` inside a procedure that is called twice now runs once.
- **[An id selector that is valid in Java's regex but not in Go's]** → It reports `cannot include` where upstream would match. Ids are plain names in practice.
- **[A self-recursive `!include_many` still crashes the gopuml process]** → Out of scope. It needs an include depth limit, which upstream does not have either.
- **[Large regenerated diffs]** → The version line changes everywhere. The review filters it out before reading the other differences.

## Migration Plan

Nothing to migrate. Outputs change as upstream's changed between the two versions (proposal, Impact). Rollback means reverting the change.
