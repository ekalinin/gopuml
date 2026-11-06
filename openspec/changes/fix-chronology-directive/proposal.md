# Proposal

## Why

An `@startchronology` block is drawn as an error image with "chronology diagrams are not supported yet" (issue #58), while upstream draws its "Diagram not supported by this release of PlantUML" diagram, which says that the directive `@startchronology` is not recognized. Upstream recognizes the start tag (`DiagramType.CHRONOLOGY`), but its MIT variant registers no parser for it (`ChronologyDiagramFactory` is commented out in `PSystemBuilder`), so no factory runs and upstream builds `PSystemUnsupported`. Chronology is the only start tag that upstream recognizes without a parser. No conformance case has an `@startchronology` block today, so the report shows no failure for it.

## What Changes

- `@startchronology` gives upstream's "Diagram not supported by this release of PlantUML" diagram, which names the directive, as `@startfoo` does, instead of the "not supported yet" error.
- The syntax check prints nothing for such a block and exits with status 200, as upstream, instead of reporting an error.
- The family detection requirement of `diagram-source` names `@startchronology` among the start tags that give upstream's unsupported diagram, so that "recognizes" in the requirement means a start tag that upstream also parses.

Out of scope:

- Chronology diagrams themselves: upstream's MIT variant does not draw them.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `diagram-source`: the family detection requirement gives `@startchronology` upstream's unsupported diagram, with a new scenario

## Impact

- **Code:** `internal/core` (`CheckFamily`).
- **Conformance:** a curated case `diagram-source/chronology-directive` with its reference. The report's counts of cases grow by one; the known differences and the cases not supported yet do not change. No `@startuml` case is added, so the golden dumps stay as they are.
- **Documentation:** the README's case count.
- **Compatibility:** an `@startchronology` block, and a `{{chronology` ... `}}` diagram embedded in creole text, which goes through the same family check, give upstream's unsupported diagram instead of the "not supported yet" error. The exit status of the command line stays 200.
