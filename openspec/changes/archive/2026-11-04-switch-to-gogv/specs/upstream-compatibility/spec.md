# Spec Delta

## MODIFIED Requirements

### Requirement: Continuous conformance

The conformance suite SHALL run on linux/amd64 for every pull request and every push to the main branch. In that run:

- The outputs of this implementation SHALL be produced in the pinned reference container, laid out by its built-in layout engine (gogv) and not by the Graphviz of the container, unless a case names a `dot`. The references keep the pinned Graphviz version.
- Every conformance group SHALL be required: a failing case that is not listed in the known-differences file SHALL fail the run.
- When the run renders references, it SHALL fail if the committed references differ from the new render.
- The run SHALL publish the conformance report, whether it passes or fails.

#### Scenario: Regression fails the run

- **WHEN** a pull request changes the implementation so that a sequence case no longer matches its reference
- **THEN** the conformance run of the pull request fails and names the case

#### Scenario: Drawing an unsupported type fails the run

- **WHEN** a change makes a diagram type that is not supported yet produce a drawing instead of the "not supported yet" error
- **THEN** the conformance run fails

#### Scenario: Known differences do not fail the run

- **WHEN** the only failing cases are listed in the known-differences file
- **THEN** the conformance run passes, and the report lists those cases as known differences

#### Scenario: Stale committed reference fails the run

- **WHEN** the run renders the references and a committed curated reference differs from the new render
- **THEN** the conformance run fails and names the differing reference files

#### Scenario: Report published on failure

- **WHEN** the conformance run fails
- **THEN** the conformance report of that run is still available from the run

#### Scenario: Graph cases laid out by the built-in engine

- **WHEN** the run renders the class case `graph-layout/class-layout` with this implementation in the reference container, whose environment sets `GRAPHVIZ_DOT`
- **THEN** this implementation lays the case out with the built-in engine, and the container's `dot` runs only for the reference
