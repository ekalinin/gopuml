// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/asciiverse/AsciiBlock.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/asciiverse/AsciiBlockMarginLR.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/asciiverse/AsciiBlockStackTB.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package asciiverse

// AsciiBlock is something that draws itself on the character grid: its
// size, and its drawing at the position of the plan it is given (the top
// left corner of the block).
//
// Upstream's default methods throw UnsupportedOperationException with the
// class name; types without a text drawing do the same (see
// ThrowUnsupported).
type AsciiBlock interface {
	AsciiDimension() ADimension2D
	AsciiDraw(plan *InfinitePlan)
}

// ThrowUnsupported panics as upstream's AsciiBlock default methods:
// UnsupportedOperationException with the Java class name of the block.
func ThrowUnsupported(className string) { throw(unsupportedOperation, className) }

// MarginLR returns block with blank columns on its left and right
// (upstream's default method marginLR).
func MarginLR(block AsciiBlock, marginLeft, marginRight int) AsciiBlock {
	return newAsciiBlockMarginLR(block, marginLeft, marginRight)
}

// AsciiBlockMarginLR pads another block with blank columns on the left and
// the right; an empty block (no width) stays empty.
type AsciiBlockMarginLR struct {
	inner       AsciiBlock
	marginLeft  int
	marginRight int
}

func newAsciiBlockMarginLR(inner AsciiBlock, marginLeft, marginRight int) *AsciiBlockMarginLR {
	if marginLeft < 0 || marginRight < 0 {
		throw(illegalArgument, "")
	}
	return &AsciiBlockMarginLR{inner: inner, marginLeft: marginLeft, marginRight: marginRight}
}

// AsciiDimension returns the size of the inner block with the margins, or
// the size of an empty inner block unchanged.
func (b *AsciiBlockMarginLR) AsciiDimension() ADimension2D {
	dim := b.inner.AsciiDimension()
	if dim.Width() == 0 {
		// An empty inner block (an unlabeled message) stays empty: no
		// margin is reserved
		return dim
	}
	return dim.DeltaWH(b.marginLeft+b.marginRight, 0)
}

// AsciiDraw draws the inner block after the left margin.
func (b *AsciiBlockMarginLR) AsciiDraw(plan *InfinitePlan) {
	b.inner.AsciiDraw(plan.Move(b.marginLeft, 0))
}

// AsciiBlockStackTB stacks a block above another one, both centered on the
// wider of the two, with a blank row between them (the text counterpart of
// DecorateEntityImage.addTop).
type AsciiBlockStackTB struct {
	top    AsciiBlock
	bottom AsciiBlock
}

// NewAsciiBlockStackTB returns top above bottom.
func NewAsciiBlockStackTB(top, bottom AsciiBlock) *AsciiBlockStackTB {
	return &AsciiBlockStackTB{top: top, bottom: bottom}
}

// AsciiDimension returns the size of the stack; an empty top block
// reserves nothing, not even the blank row.
func (b *AsciiBlockStackTB) AsciiDimension() ADimension2D {
	dimTop := b.top.AsciiDimension()
	if dimTop.Width() == 0 {
		return b.bottom.AsciiDimension()
	}
	dimBottom := b.bottom.AsciiDimension()
	width := max(dimTop.Width(), dimBottom.Width())
	height := dimTop.Height() + 1 + dimBottom.Height()
	return NewADimension2D(width, height)
}

// AsciiDraw draws both blocks centered on the total width.
func (b *AsciiBlockStackTB) AsciiDraw(plan *InfinitePlan) {
	dimTop := b.top.AsciiDimension()
	if dimTop.Width() == 0 {
		b.bottom.AsciiDraw(plan)
		return
	}
	dimBottom := b.bottom.AsciiDimension()
	totalWidth := max(dimTop.Width(), dimBottom.Width())
	xTop := (totalWidth - dimTop.Width()) / 2
	b.top.AsciiDraw(plan.Move(xTop, 0))
	xBottom := (totalWidth - dimBottom.Width()) / 2
	b.bottom.AsciiDraw(plan.Move(xBottom, dimTop.Height()+1))
}
