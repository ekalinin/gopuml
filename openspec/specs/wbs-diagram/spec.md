# wbs-diagram Specification

## Purpose
Supports PlantUML v1.2026.9 WBS diagrams under `@startwbs`, with the same syntax, layout, links, styles, errors and text output as upstream.

## Requirements

### Requirement: WBS diagram blocks

The system SHALL render `@startwbs` ... `@endwbs` blocks as upstream's WBS diagrams. The diagram type SHALL be wbs and the description `Work Breakdown Structure`. A WBS diagram SHALL be drawn as upstream draws it: each node as a box with its label, the root at the top, and straight lines from each node to its children. The common commands of upstream's WBS diagrams SHALL be supported: title, header, footer, caption, legend, mainframe, `skin`, `skinparam`, `<style>`, sprites, `scale`, pragmas, and the other commands that upstream adds to WBS diagrams. WBS diagrams SHALL NOT need Graphviz. A WBS diagram without nodes SHALL fail to draw, as upstream's drawing crashes on it: rendering it SHALL return an error and no drawing, in every format.

#### Scenario: WBS tree

- **WHEN** a `@startwbs` block contains `* Project`, `** Planning`, `*** Scope` and `** Build`
- **THEN** `Project` is drawn at the top, `Planning` and `Build` below it, side by side, `Scope` below `Planning`, each parent is linked to each of its children by straight lines, the diagram type is wbs and the description is `Work Breakdown Structure`

#### Scenario: Title and style

- **WHEN** a WBS diagram contains `title Plan` and a `<style>` block with `wbsDiagram { node { BackGroundColor #33b2e2 } arrow { LineColor red } }`, and `* root` and `** a`
- **THEN** the title `Plan` is drawn above the diagram, the boxes are filled with `#33b2e2` and the lines are red

#### Scenario: Scaled WBS diagram

- **WHEN** a WBS diagram contains `scale 2`, `* root`, `**(x) a`, `**(y) b` and `x -> y`
- **THEN** the drawing, the link from `a` to `b` included, is twice as large as without `scale 2`, as upstream draws it

#### Scenario: WBS diagram without Graphviz

- **WHEN** a WBS diagram is rendered and `dot` cannot be found
- **THEN** the diagram renders without an error

#### Scenario: WBS diagram without nodes

- **WHEN** a WBS diagram contains only `title Plan`, and is rendered as SVG and as TXT
- **THEN** rendering returns an error and no drawing, in both formats

### Requirement: Nodes

The system SHALL read a node from a line that starts with one or more `*`, `+` or `-` markers, optionally after spaces and tabs, as upstream does. The markers MAY be followed by `_` for a node without a box, then by `<` or `>` for its side (`>_` is not read, as upstream does not read it), then by a background color in brackets (`[#Orange]`, `[#FFA500]`) and an alias in parentheses (`(code)`), in either order, then by a space and the label. A node MAY also be written as markers, `_`, `<` or `>` and a color, followed by a quoted label and `as` and an alias (`** "First node" as f`). A node with `_` and no label SHALL be a pseudo node: it SHALL NOT be drawn, and its children SHALL be drawn and linked through it. Any other node without a label SHALL fail the line with the message `Missing label for WBS node. (Assumed diagram type: wbs)`. A stereotype at the end of a label (`<<name>>`) SHALL be removed from the label and SHALL select the styles of that stereotype. `\n` in a label SHALL start a new line. An unknown color SHALL fail the line with the message `No such color (Assumed diagram type: wbs)`.

#### Scenario: Indented markers

- **WHEN** a WBS diagram contains `* root`, ` * a` (one space before the marker) and `  * b` (two spaces before the marker)
- **THEN** `a` is a child of `root` and `b` is a child of `a`

#### Scenario: Background color

- **WHEN** a WBS diagram contains `* root` and `**[#Orange] a`
- **THEN** the box of `a` is filled with orange (`#FFA500`) and the box of `root` keeps the default color

#### Scenario: Unknown color

- **WHEN** a WBS diagram contains `* root`, `** a` and `**[#ZZZ] b`
- **THEN** the diagram fails at the line `**[#ZZZ] b` with the message `No such color (Assumed diagram type: wbs)`

#### Scenario: Boxless node

- **WHEN** a WBS diagram contains `* root`, `**_ a` and `*** b`
- **THEN** `a` is drawn as text without a box, linked to `root` and to `b`, and `root` and `b` are drawn in boxes

#### Scenario: Pseudo node

- **WHEN** a WBS diagram contains `* root`, `**_`, `*** a`, `*** b` and `** c`
- **THEN** no box and no text are drawn for the pseudo node, `a` and `b` are drawn below the place of the pseudo node and linked to `root` through it, and `c` is drawn next to them as a child of `root`

#### Scenario: Missing label

- **WHEN** a WBS diagram contains `* root`, `** a` and `**`
- **THEN** the diagram fails at the line `**` with the message `Missing label for WBS node. (Assumed diagram type: wbs)`

#### Scenario: Stereotype at the end of a label

- **WHEN** a WBS diagram has a `<style>` block with `.blue { BackGroundColor #33b2e2 }`, and contains `* root` and `** a <<blue>>`
- **THEN** the node `a` is drawn with the label `a`, without the stereotype, in a box filled with `#33b2e2`

#### Scenario: Line break in a label

- **WHEN** a WBS diagram contains `* root` and `** first\nsecond`
- **THEN** the box of the child has two lines, `first` and `second`

### Requirement: Multiline nodes

The system SHALL read a node over several lines when its markers, with the optional `_`, `<` or `>`, color and alias, are followed by `:`, as upstream does: the label runs from the text after `:` to a line that ends with `;`, which MAY be followed by a stereotype (`; <<name>>`). Each line of the block SHALL be a line of the label.

#### Scenario: Multiline node

- **WHEN** a WBS diagram contains `* root`, `**:first line` and `second line;`
- **THEN** the child of `root` is drawn as one box with the two lines `first line` and `second line`

#### Scenario: Multiline node with a stereotype

- **WHEN** a WBS diagram has a `<style>` block with `.blue { BackGroundColor #33b2e2 }`, and contains `* root`, `**:first line` and `second line; <<blue>>`
- **THEN** the child of `root` has the two lines `first line` and `second line`, without the stereotype, in a box filled with `#33b2e2`

### Requirement: Sides

The system SHALL draw a node on the left side of its parent when its markers contain `-` or are followed by `<`, and on the right side otherwise or when they are followed by `>`, as upstream does. The children of a node below the root SHALL be drawn one below the other, the left ones to the left of a vertical line that goes down from the middle of the node and the right ones to its right. The children of the root SHALL be drawn side by side below it, and a child of the root on the left side SHALL be drawn before the other children of the root, as upstream does.

#### Scenario: Side markers

- **WHEN** a WBS diagram contains `* root`, `** a`, `***< b` and `***> c`
- **THEN** `b` is drawn to the left of the vertical line below `a` and `c` to its right

#### Scenario: Plus and minus markers

- **WHEN** a WBS diagram contains `+ root`, `++ a`, `+++ b` and `++- c`
- **THEN** `a` is a child of `root`, `b` is drawn to the right of the vertical line below `a` and `c` to its left

#### Scenario: Left child of the root

- **WHEN** a WBS diagram contains `* root`, `** a` and `**< b`
- **THEN** `b` is drawn to the left of `a`, both below `root`

### Requirement: Tree structure

The system SHALL build the tree as upstream does. The first node SHALL be the root, whatever its markers. The level of every later node SHALL be computed as upstream does: the number of markers minus one when no space or tab comes before them, and otherwise from the spaces, the tabs and the markers of the line and the markers of the first node, as upstream computes it. A node SHALL be a child of the previous node when its level is one more than the level of that node, and SHALL otherwise be a child of the ancestor of the previous node one level above it. A level more than one below the previous node SHALL fail the line with the message `Bad tree structure (Assumed diagram type: wbs)`. A second node at level zero SHALL fail the line with the message `WBS diagram only accepts one root first declared. (Assumed diagram type: wbs)`. Markers that upstream cannot read, such as `  ++` after a first node written with `*`, SHALL fail the line with upstream's crash message, `You should send a mail to plantuml@gmail.com with this log (V1.2026.9) java.lang.UnsupportedOperationException: type=<  ++>[*] (Assumed diagram type: wbs)`.

#### Scenario: Root with several markers

- **WHEN** a WBS diagram contains `** root` and `** a`
- **THEN** `root` is drawn as the root and `a` as its child

#### Scenario: Skipped level

- **WHEN** a WBS diagram contains `* root`, `** a` and `**** b`
- **THEN** the diagram fails at the line `**** b` with the message `Bad tree structure (Assumed diagram type: wbs)`

#### Scenario: Second root

- **WHEN** a WBS diagram contains `* a`, `** b` and `* c`
- **THEN** the diagram fails at the line `* c` with the message `WBS diagram only accepts one root first declared. (Assumed diagram type: wbs)`

#### Scenario: Unreadable markers

- **WHEN** a WBS diagram contains `* root`, `** a` and `  ++ b`
- **THEN** the diagram fails at the line `  ++ b` with the message `You should send a mail to plantuml@gmail.com with this log (V1.2026.9) java.lang.UnsupportedOperationException: type=<  ++>[*] (Assumed diagram type: wbs)`

### Requirement: Old order

The system SHALL read a node whose color and alias come before `_`, `<` and `>`, on one line or in multiline form, as upstream does, and SHALL draw upstream's warning above the diagram for each such line: `Shape (_) and direction (< >) must come before color and id: start the line with '<prefix>'`, where `<prefix>` is the markers of the line, without the spaces and tabs before them, followed by its `_`, its `<` or `>`, its color in brackets and, unless its label is quoted, its alias in parentheses.

#### Scenario: Old order

- **WHEN** a WBS diagram contains `* root`, `**[#red](x)_ a`, `**(y) b` and `x -> y`
- **THEN** `a` is drawn as text without a box, the warning `Shape (_) and direction (< >) must come before color and id: start the line with '**_[#red](x)'` is drawn above the diagram, and the arrow from `a` to `b` is drawn at their place below the warning

#### Scenario: Old order in a multiline node

- **WHEN** a WBS diagram contains `* root`, `**[#red]>:first` and `second;`
- **THEN** the child of `root` has the two lines `first` and `second`, and the warning `Shape (_) and direction (< >) must come before color and id: start the line with '**>[#red]'` is drawn above the diagram

### Requirement: Links

The system SHALL read a line `code1 -> code2`, with any number of `-` and `.` before `>`, as a link from the node with the alias `code1` to the node with the alias `code2`, as upstream does. The link SHALL be drawn over the tree as a straight arrow along the line between the centers of the two nodes, from the border of the first node to the border of the second, with an arrow head at the second. A color after the aliases (`#red`) SHALL be the color of the link. Without a color, the link SHALL take the line color of the `arrow` style of `wbsDiagram`, with the stereotype of the link (`<<name>>`). A label after `:` SHALL be read and not drawn, as upstream does. An alias that no node has SHALL fail the line with the message `No such node <alias> (Assumed diagram type: wbs)`.

#### Scenario: Link between aliased nodes

- **WHEN** a WBS diagram contains `* root`, `**(x) a`, `**(y) b` and `x -> y`
- **THEN** an arrow is drawn from the box of `a` to the box of `b`, with its head at `b`

#### Scenario: Quoted label with an alias

- **WHEN** a WBS diagram contains `* root`, `** "First node" as f`, `**(s) second` and `f -> s`
- **THEN** the first child of `root` is drawn with the label `First node`, and an arrow is drawn from it to `second`

#### Scenario: Link color

- **WHEN** a WBS diagram contains `* root`, `**(x) a`, `**(y) b` and `x -> y #red`
- **THEN** the arrow from `a` to `b` is red

#### Scenario: Link stereotype

- **WHEN** a WBS diagram has a `<style>` block with `wbsDiagram { .hot { LineColor red } }`, and contains `* root`, `**(x) a`, `**(y) b` and `x -> y <<hot>>`
- **THEN** the arrow from `a` to `b` is red

#### Scenario: Link label

- **WHEN** a WBS diagram contains `* root`, `**(x) a`, `**(y) b` and `x -> y : hello`
- **THEN** the arrow from `a` to `b` is drawn and `hello` is not drawn

#### Scenario: Unknown alias

- **WHEN** a WBS diagram contains `* root`, `**(x) a` and `x -> y`
- **THEN** the diagram fails at the line `x -> y` with the message `No such node y (Assumed diagram type: wbs)`

### Requirement: Layout

The system SHALL lay out the nodes as upstream does: the children of the root side by side below it, joined to it by a horizontal line, and the children of a deeper node one below the other below it, next to the vertical line that goes down from its middle. A `Width auto` in the style of a node SHALL give the boxes of its children the width of the widest of them. The boxes SHALL follow the `MinimumWidth`, `HorizontalAlignment`, `RoundCorner`, `Padding`, `Margin` and `Shadowing` of their style, as upstream draws them.

#### Scenario: Columns

- **WHEN** a WBS diagram contains `* root`, `** a`, `*** b`, `*** c` and `**** d`
- **THEN** `b` and `c` are drawn below `a`, one below the other, to the right of the vertical line below `a`, and `d` below `c`, to the right of the vertical line below `c`

#### Scenario: Auto width

- **WHEN** a WBS diagram has a `<style>` block with `wbsDiagram { node { Width auto } }`, and contains `* root`, `** a`, `** longer node`, `*** b` and `*** much longer node`
- **THEN** the boxes of `a` and `longer node` have the same width, and the boxes of `b` and `much longer node` have the same width

#### Scenario: Minimum width

- **WHEN** a WBS diagram has a `<style>` block with `wbsDiagram { node { MinimumWidth 100 } }`, and contains `* root` and `** a`
- **THEN** the boxes of `root` and `a` are 100 wide

### Requirement: WBS styles

The system SHALL style the nodes with the `wbsDiagram` styles, as upstream does: `node` for every node, `rootNode` for the root, `leafNode` for the nodes without children, `boxless` for the nodes without a box, `arrow` for the lines, `:depth(n)` for the nodes at depth `n`, and `.name` for the nodes with the stereotype `name`. A style declared under `.name *` SHALL apply to the node with the stereotype `name` and to its descendants. When several ancestors of a node have such a style, the nearest one SHALL win.

#### Scenario: Root, leaf and boxless nodes

- **WHEN** a WBS diagram has a `<style>` block with `wbsDiagram { node { BackGroundColor lightgray } rootNode { BackGroundColor gold } leafNode { BackGroundColor lightblue } boxless { FontColor red } }`, and contains `* root`, `** a`, `*** b` and `**_ c`
- **THEN** `root` is filled with gold, `a` with light gray, `b` with light blue, and `c` is drawn without a box in red

#### Scenario: Depth style

- **WHEN** a WBS diagram has a `<style>` block with `wbsDiagram { :depth(1) { BackGroundColor pink } }`, and contains `* root`, `** a` and `*** b`
- **THEN** `a` is filled with pink, and `root` and `b` keep the default color

#### Scenario: Style of the descendants

- **WHEN** a WBS diagram has a `<style>` block with `wbsDiagram { .deep * { node { FontColor blue } } }`, and contains `* root`, `** a <<deep>>`, `*** b`, `**** c` and `** d`
- **THEN** `a`, `b` and `c` are written in blue, and `root` and `d` in the default color

#### Scenario: Nearest ancestor

- **WHEN** a WBS diagram has a `<style>` block with `wbsDiagram { .outer * { node { FontColor blue } } .inner * { node { FontColor red } } }`, and contains `* root`, `** a <<outer>>`, `*** b <<inner>>`, `**** c` and `*** e`
- **THEN** `a` and `e` are written in blue, and `b` and `c` in red

### Requirement: WBS output formats

WBS diagrams SHALL render to SVG and PNG. TXT and UTXT SHALL render upstream's text output in ASCII, without the title and the other annotations: the label of the root in a box, then the label of each child of the root in a box, each box 4 columns to the right of and 4 lines below the previous one, without the deeper nodes. The box of the root SHALL have 1 space on each side of its label, the box of a child of the root with children of its own 7 spaces, and the box of a child of the root without children 3 spaces.

#### Scenario: PNG WBS diagram

- **WHEN** a WBS diagram with `* root`, `** a` and `*** b` is rendered as PNG
- **THEN** the PNG image shows the three nodes and their lines, as upstream's PNG does

#### Scenario: UTXT WBS diagram

- **WHEN** a WBS diagram with `* Project`, `** Planning`, `*** Scope` and `** Build` is rendered as UTXT
- **THEN** the output is the same ASCII text as for TXT, with the boxes of `Project`, `Planning` and `Build`

#### Scenario: Text of the root's children

- **WHEN** a WBS diagram with `title T`, `* root`, `** a`, `*** b` and `** c` is rendered as TXT
- **THEN** the output has the boxes `| root |`, `|       a       |` and `|   c   |`, without `T` and `b`

### Requirement: Embedded WBS diagrams

A `{{wbs` ... `}}` block in creole text SHALL be drawn as the WBS diagram it contains, as upstream draws it.

#### Scenario: WBS diagram in a note

- **WHEN** a sequence diagram has a note that contains `{{wbs`, `* root`, `** a` and `}}`
- **THEN** the note contains the drawing of the WBS diagram, and the rest of the diagram renders normally
