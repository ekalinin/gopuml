# Spec Delta

## Purpose

Supports PlantUML v1.2026.9 mindmap diagrams under `@startmindmap`, with the same syntax, layout, styles and errors as upstream.

## ADDED Requirements

### Requirement: Mindmap diagram blocks

The system SHALL render `@startmindmap` ... `@endmindmap` blocks as upstream's mindmap diagrams. The diagram type SHALL be mindmap and the description `MindMap`. A mindmap diagram SHALL be drawn as upstream draws it: each node as a rounded box with its label, the root with its children next to it, and a curved link from each node to each of its children. The common commands of upstream's mindmap diagrams SHALL be supported: title, header, footer, caption, legend, mainframe, `skin`, `skinparam`, `<style>`, sprites, `scale`, pragmas, and the other commands that upstream adds to mindmap diagrams. Mindmap diagrams SHALL NOT need Graphviz.

#### Scenario: Org-mode mindmap

- **WHEN** a `@startmindmap` block contains `* root`, `** a`, `*** a1`, `*** a2`, `** b` and `*** b1`
- **THEN** `root` is drawn on the left, `a` and `b` to its right, one below the other, `a1` and `a2` to the right of `a`, `b1` to the right of `b`, each parent is linked to each of its children by a curve, the diagram type is mindmap and the description is `MindMap`

#### Scenario: Title and style

- **WHEN** a mindmap diagram contains `title Ideas` and a `<style>` block with `mindmapDiagram { node { BackGroundColor #33b2e2 } arrow { LineColor red } }`, and `* root` and `** a`
- **THEN** the title `Ideas` is drawn above the diagram, the boxes are filled with `#33b2e2` and the link is red

#### Scenario: Scaled mindmap diagram

- **WHEN** a mindmap diagram contains `scale 2`
- **THEN** the drawing is twice as large as without `scale 2`, as upstream draws it

#### Scenario: Mindmap diagram without Graphviz

- **WHEN** a mindmap diagram is rendered and `dot` cannot be found
- **THEN** the diagram renders without an error

### Requirement: Nodes

The system SHALL read a node from a line that starts with one or more `*` or `#` markers, optionally after spaces and tabs, as upstream does. The markers MAY be followed by a background color in brackets (`[#Orange]`, `[#FFA500]`), then by `_` for a node without a box, then by the label. A stereotype at the end of a label (`<<name>>`) SHALL be removed from the label and SHALL select the styles of that stereotype. `\n` in a label SHALL start a new line. An unknown color SHALL fail the line with the message `No such color (Assumed diagram type: mindmap)`.

#### Scenario: Hash markers

- **WHEN** a mindmap diagram contains `# root`, `## a` and `### b`
- **THEN** the diagram is drawn as the same lines with `*` markers

#### Scenario: Indented markers

- **WHEN** a mindmap diagram contains `* root`, ` * a` (one space before the marker) and `  * b` (two spaces before the marker)
- **THEN** `a` is a child of `root` and `b` is a child of `a`

#### Scenario: Background color

- **WHEN** a mindmap diagram contains `* root` and `**[#Orange] a`
- **THEN** the box of `a` is filled with orange (`#FFA500`) and the box of `root` keeps the default color

#### Scenario: Unknown color

- **WHEN** a mindmap diagram contains `* root` and `**[#ZZZ] a`
- **THEN** the diagram fails at that line with the message `No such color (Assumed diagram type: mindmap)`

#### Scenario: Boxless node

- **WHEN** a mindmap diagram contains `* root`, `**_ a` and `*** b`
- **THEN** `a` is drawn as text without a box, linked to `root` and to `b`, and `root` and `b` are drawn in boxes

#### Scenario: Stereotype at the end of a label

- **WHEN** a mindmap diagram has a `<style>` block with `.blue { BackGroundColor #33b2e2 }`, and contains `* root` and `** a <<blue>>`
- **THEN** the node `a` is drawn with the label `a`, without the stereotype, in a box filled with `#33b2e2`

#### Scenario: Line break in a label

- **WHEN** a mindmap diagram contains `* root` and `** first\nsecond`
- **THEN** the box of the child has two lines, `first` and `second`

### Requirement: Multiline nodes

The system SHALL read a node over several lines when its markers are followed by `:`, as upstream does: the label runs from the text after `:` to a line that ends with `;`, which MAY be followed by a stereotype (`; <<name>>`). Each line of the block SHALL be a line of the label. Spaces and tabs before the markers SHALL be ignored, and the depth SHALL be the number of markers minus one.

#### Scenario: Multiline node

- **WHEN** a mindmap diagram contains `* root`, `**:first line` and `second line;`
- **THEN** the child of `root` is drawn as one box with the two lines `first line` and `second line`

#### Scenario: Multiline node with a stereotype

- **WHEN** a mindmap diagram has a `<style>` block with `.blue { BackGroundColor #33b2e2 }`, and contains `* root`, `**:first line` and `second line; <<blue>>`
- **THEN** the child of `root` has the two lines `first line` and `second line`, without the stereotype, in a box filled with `#33b2e2`

### Requirement: Sides

The system SHALL draw the nodes of the right side to the right of the root and the nodes of the left side to its left, as upstream does. A line that starts with `+` markers SHALL add a node on the right side and a line that starts with `-` markers a node on the left side, with the same color, `_` and label syntax as the `*` markers. A line that names `left`, `right`, `top` or `bottom` followed by `side` or `direction` (for example `left side`) SHALL set the side of the next `*` and `#` nodes: `left` and `top` the left side, `right` and `bottom` the right side. A line that starts with `0` SHALL set the root to the text after it. When both sides have nodes, the root SHALL be drawn once, in the middle.

#### Scenario: Left and right branches

- **WHEN** a mindmap diagram contains `+ root`, `++ right` and `-- left`
- **THEN** `root` is drawn once in the middle, `right` to its right and `left` to its left, each linked to `root`

#### Scenario: Left side command

- **WHEN** a mindmap diagram contains `* root`, `** a`, `left side` and `** b`
- **THEN** `a` is drawn to the right of `root` and `b` to its left

#### Scenario: Root command

- **WHEN** a mindmap diagram contains `0 Root` and `** a`
- **THEN** `Root` is drawn as the root and `a` as its child

### Requirement: Indentation

The system SHALL compute the depth of a node as upstream does. The first node of a mindmap SHALL be its root, whatever its depth. The depth of the first node after the root SHALL be the step of the indentation, and the depth of every later node SHALL be divided by that step. A depth that is not a multiple of the step SHALL fail the line with the message `Bad indentation (Assumed diagram type: mindmap)`. A depth more than one level below the previous node SHALL fail the line with the message `error42L (Assumed diagram type: mindmap)`. A node at depth zero after the root SHALL start a new mindmap, drawn below the previous one. Markers that upstream cannot read, such as `  ##` after a first node written with `*`, SHALL fail the line with upstream's crash message, `You should send a mail to plantuml@gmail.com with this log (V1.2026.9) java.lang.UnsupportedOperationException: type=<  ##>[*] (Assumed diagram type: mindmap)`.

#### Scenario: Root at any depth

- **WHEN** a mindmap diagram contains `** root` and `*** a`
- **THEN** `root` is drawn as the root and `a` as its child

#### Scenario: Indentation step

- **WHEN** a mindmap diagram contains `* root`, `*** a` and `***** b`
- **THEN** `a` is a child of `root` and `b` is a child of `a`

#### Scenario: Bad indentation

- **WHEN** a mindmap diagram contains `** root`, `*** a` and `**** b`
- **THEN** the diagram fails at the line `**** b` with the message `Bad indentation (Assumed diagram type: mindmap)`

#### Scenario: Skipped level

- **WHEN** a mindmap diagram contains `* root`, `** a` and `**** b`
- **THEN** the diagram fails at the line `**** b` with the message `error42L (Assumed diagram type: mindmap)`

#### Scenario: Several mindmaps

- **WHEN** a mindmap diagram contains `* a`, `** b`, `* c` and `** d`
- **THEN** two mindmaps are drawn, `a` with its child `b`, and below it `c` with its child `d`

#### Scenario: Unreadable markers

- **WHEN** a mindmap diagram contains `* root`, `** a` and `  ## x`
- **THEN** the diagram fails at the line `  ## x` with the message `You should send a mail to plantuml@gmail.com with this log (V1.2026.9) java.lang.UnsupportedOperationException: type=<  ##>[*] (Assumed diagram type: mindmap)`

### Requirement: Layout

The system SHALL lay out the nodes as upstream does: the children of a node are placed next to it, one below the other, as close as their subtrees allow, and centered on the node. `top to bottom direction` SHALL draw the mindmap from top to bottom, with the children of a node below it, side by side. A link color that is transparent SHALL leave the links undrawn. The boxes SHALL follow the `MinimumWidth`, `HorizontalAlignment`, `RoundCorner`, `Padding`, `Margin` and `Shadowing` of their style, as upstream draws them.

#### Scenario: Top to bottom direction

- **WHEN** a mindmap diagram contains `top to bottom direction`, `* root`, `** a` and `** b`
- **THEN** `root` is drawn at the top, `a` and `b` below it, side by side, each linked to `root` by a vertical curve

#### Scenario: Transparent links

- **WHEN** a mindmap diagram has a `<style>` block with `mindmapDiagram { arrow { LineColor transparent } }`, and contains `* root` and `** a`
- **THEN** `root` and `a` are drawn without a link between them

#### Scenario: Minimum width and alignment

- **WHEN** a mindmap diagram has a `<style>` block with `node { MinimumWidth 100 }`, `.right { HorizontalAlignment Right }` and `.center { HorizontalAlignment Center }`, and contains `* Root`, ` * Node 1 <<right>>` and ` * Node 2 <<center>>`
- **THEN** the boxes are 100 wide, and the labels are aligned as upstream aligns them: `Root` on the left of its box, `Node 1` on the right and `Node 2` in the middle

### Requirement: Mindmap styles

The system SHALL style the nodes with the `mindmapDiagram` styles, as upstream does: `node` for every node, `rootNode` for the root, `leafNode` for the nodes without children, `boxless` for the nodes without a box, `arrow` for the links, `:depth(n)` for the nodes at depth `n`, and `.name` for the nodes with the stereotype `name`. A style declared under `.name *` SHALL apply to the node with the stereotype `name` and to its descendants. When several ancestors of a node have such a style, the nearest one SHALL win.

#### Scenario: Root, leaf and boxless nodes

- **WHEN** a mindmap diagram has a `<style>` block with `mindmapDiagram { node { BackGroundColor lightgray } rootNode { BackGroundColor gold } leafNode { BackGroundColor lightblue } boxless { FontColor red } }`, and contains `* root`, `** a`, `*** b` and `**_ c`
- **THEN** `root` is filled with gold, `a` with light gray, `b` with light blue, and `c` is drawn without a box in red

#### Scenario: Depth style

- **WHEN** a mindmap diagram has a `<style>` block with `mindmapDiagram { :depth(1) { BackGroundColor pink } }`, and contains `* root`, `** a` and `*** b`
- **THEN** `a` is filled with pink, and `root` and `b` keep the default color

#### Scenario: Style of the descendants

- **WHEN** a mindmap diagram has a `<style>` block with `mindmapDiagram { .deep * { node { FontColor blue } } }`, and contains `* root`, `** a <<deep>>`, `*** b`, `**** c` and `** d`
- **THEN** `a`, `b` and `c` are written in blue, and `root` and `d` in the default color

#### Scenario: Nearest ancestor

- **WHEN** a mindmap diagram has a `<style>` block with `mindmapDiagram { .outer * { node { FontColor blue } } .inner * { node { FontColor red } } }`, and contains `* root`, `** a <<outer>>`, `*** b <<inner>>`, `**** c` and `*** e`
- **THEN** `a` and `e` are written in blue, and `b` and `c` in red

### Requirement: Mindmap output formats

Mindmap diagrams SHALL render to SVG and PNG. TXT and UTXT output SHALL be rejected with the error that names the format and the diagram type, without output, as for the other diagram types that have no text output (every type except sequence and gantt).

#### Scenario: PNG mindmap diagram

- **WHEN** a mindmap diagram with `* root`, `** a` and `*** b` is rendered as PNG
- **THEN** the PNG image shows the three nodes and their links, as upstream's PNG does

#### Scenario: TXT for a mindmap diagram

- **WHEN** a mindmap diagram is rendered as TXT
- **THEN** the system reports that TXT output is not supported for mindmap diagrams and produces no output

### Requirement: Embedded mindmap diagrams

A `{{mindmap` ... `}}` block in creole text SHALL be drawn as the mindmap diagram it contains, as upstream draws it.

#### Scenario: Mindmap diagram in a note

- **WHEN** a sequence diagram has a note that contains `{{mindmap`, `* root`, `** a` and `}}`
- **THEN** the note contains the drawing of the mindmap diagram, and the rest of the diagram renders normally
