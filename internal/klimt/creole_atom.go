// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/creole/atom/AtomSpace.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/creole/atom/AtomHorizontalTexts.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/creole/atom/AtomVerticalTexts.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/creole/atom/AtomWithMargin.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/creole/atom/Bullet.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/creole/atom/AtomTable.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/creole/atom/AtomTree.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/creole/atom/AtomSprite.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/creole/atom/AtomImg.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/creole/atom/AtomImgSvg.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/creole/CreoleHorizontalLine.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/element/Skeleton2.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/utils/Base64Coder.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"bytes"
	"errors"
	"image"
	"image/draw"
	_ "image/gif"  // GIF decoding for <img>, as Java's ImageIO
	_ "image/jpeg" // JPEG decoding for <img>, as Java's ImageIO
	_ "image/png"  // PNG decoding for <img>, as Java's ImageIO
	"math"
	"strings"

	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// AtomSpace is an empty space of a given width ("<space:n>").
type AtomSpace struct {
	AtomBase
	width float64
}

// AtomSpaceCreate returns a space of width.
func AtomSpaceCreate(width float64) Atom { return &AtomSpace{width: width} }

// CalculateDimension returns the width and a height of 1.
func (a *AtomSpace) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return a.Dimension(stringBounder, func(font.StringBounder) geom.XDimension2D {
		return geom.NewXDimension2D(a.width, 1)
	})
}

// StartingAltitude returns 0.
func (a *AtomSpace) StartingAltitude(font.StringBounder) float64 { return 0 }

// DrawU draws nothing.
func (a *AtomSpace) DrawU(UGraphic) {}

// Neutrons returns the atom as one neutron.
func (a *AtomSpace) Neutrons() []*Neutron { return atomNeutrons(a) }

// AtomHorizontalTexts puts atoms side by side, on the baseline of the
// first.
type AtomHorizontalTexts struct {
	AtomBase
	all []Atom
}

// NewAtomHorizontalTexts returns the atoms side by side.
func NewAtomHorizontalTexts(texts []Atom) *AtomHorizontalTexts {
	return &AtomHorizontalTexts{all: texts}
}

// CalculateDimension returns the sum of the widths and the largest height.
func (a *AtomHorizontalTexts) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return a.Dimension(stringBounder, func(sb font.StringBounder) geom.XDimension2D {
		width, height := 0.0, 0.0
		for _, text := range a.all {
			dim := text.CalculateDimension(sb)
			height = math.Max(height, dim.Height())
			width += dim.Width()
		}
		return geom.NewXDimension2D(width, height)
	})
}

// StartingAltitude returns the altitude of the first atom, 0 without atoms.
func (a *AtomHorizontalTexts) StartingAltitude(stringBounder font.StringBounder) float64 {
	if len(a.all) == 0 {
		return 0
	}
	return a.all[0].StartingAltitude(stringBounder)
}

// DrawU draws the atoms side by side.
func (a *AtomHorizontalTexts) DrawU(ug UGraphic) {
	stringBounder := ug.StringBounder()
	startingAltitude := a.StartingAltitude(stringBounder)
	x := 0.0
	for _, text := range a.all {
		dim := text.CalculateDimension(stringBounder)
		localStartingAltitude := text.StartingAltitude(stringBounder)
		text.DrawU(ug.Apply(NewUTranslate(x, startingAltitude-localStartingAltitude)))
		x += dim.Width()
	}
}

// Neutrons returns the atom as one neutron.
func (a *AtomHorizontalTexts) Neutrons() []*Neutron { return atomNeutrons(a) }

// AtomVerticalTexts stacks atoms.
type AtomVerticalTexts struct {
	AtomBase
	all []Atom
}

// NewAtomVerticalTexts returns the atoms stacked.
func NewAtomVerticalTexts(texts []Atom) *AtomVerticalTexts { return &AtomVerticalTexts{all: texts} }

// CalculateDimension returns the largest width and the sum of the heights.
func (a *AtomVerticalTexts) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return a.Dimension(stringBounder, func(sb font.StringBounder) geom.XDimension2D {
		width, height := 0.0, 0.0
		for _, text := range a.all {
			dim := text.CalculateDimension(sb)
			width = math.Max(width, dim.Width())
			height += dim.Height()
		}
		return geom.NewXDimension2D(width, height)
	})
}

// StartingAltitude returns the altitude of the first atom; without atoms it
// panics, as upstream's List.get.
func (a *AtomVerticalTexts) StartingAltitude(stringBounder font.StringBounder) float64 {
	if len(a.all) == 0 {
		throw(indexOutOfBounds, "Index 0 out of bounds for length 0")
	}
	return a.all[0].StartingAltitude(stringBounder)
}

// DrawU draws the atoms stacked.
func (a *AtomVerticalTexts) DrawU(ug UGraphic) {
	y := 0.0
	for _, text := range a.all {
		dim := text.CalculateDimension(ug.StringBounder())
		text.DrawU(ug.Apply(UTranslateDy(y)))
		y += dim.Height()
	}
}

// Neutrons returns the atom as one neutron.
func (a *AtomVerticalTexts) Neutrons() []*Neutron { return atomNeutrons(a) }

// AtomWithMargin adds vertical margins to an atom.
type AtomWithMargin struct {
	AtomBase
	marginY1, marginY2 float64
	atom               Atom
}

// NewAtomWithMargin returns atom with marginY1 above and marginY2 below.
func NewAtomWithMargin(atom Atom, marginY1, marginY2 float64) *AtomWithMargin {
	return &AtomWithMargin{atom: atom, marginY1: marginY1, marginY2: marginY2}
}

// CalculateDimension returns the size of the atom with the margins.
func (a *AtomWithMargin) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return a.Dimension(stringBounder, func(sb font.StringBounder) geom.XDimension2D {
		return a.atom.CalculateDimension(sb).DeltaWH(0, a.marginY1+a.marginY2)
	})
}

// StartingAltitude returns the altitude of the atom.
func (a *AtomWithMargin) StartingAltitude(stringBounder font.StringBounder) float64 {
	return a.atom.StartingAltitude(stringBounder)
}

// DrawU draws the atom below the top margin.
func (a *AtomWithMargin) DrawU(ug UGraphic) { a.atom.DrawU(ug.Apply(UTranslateDy(a.marginY1))) }

// Neutrons returns the atom as one neutron.
func (a *AtomWithMargin) Neutrons() []*Neutron { return atomNeutrons(a) }

// Bullet is the bullet of a list item: a disc at the first level, a small
// square below.
type Bullet struct {
	AtomBase
	fontConfiguration *FontConfiguration
	order             int
}

// NewBullet returns the bullet of a level.
func NewBullet(fontConfiguration *FontConfiguration, order int) *Bullet {
	return &Bullet{fontConfiguration: fontConfiguration, order: order}
}

// DrawU draws the bullet in the font color.
func (a *Bullet) DrawU(ug UGraphic) {
	c := a.fontConfiguration.Color()
	ug = ug.Apply(c).Apply(Bg(c)).Apply(UStrokeWithThickness(0))
	if a.order == 0 {
		ug = ug.Apply(UTranslateDx(3))
		ug.Draw(UEllipseBuild(5, 5))
	} else {
		ug = ug.Apply(UTranslateDx(float64(1 + 8*a.order)))
		ug.Draw(URectangleBuild(3.5, 3.5))
	}
}

// CalculateDimension returns the size of the bullet with its indentation.
func (a *Bullet) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return a.Dimension(stringBounder, func(font.StringBounder) geom.XDimension2D {
		if a.order == 0 {
			return geom.NewXDimension2D(12, 5)
		}
		return geom.NewXDimension2D(float64(8+8*a.order), 3)
	})
}

// StartingAltitude returns the baseline offset of the bullet.
func (a *Bullet) StartingAltitude(font.StringBounder) float64 {
	if a.order == 0 {
		return -5
	}
	return -7
}

// Neutrons returns the atom as one neutron.
func (a *Bullet) Neutrons() []*Neutron { return atomNeutrons(a) }

// atomTableLine is a row of an AtomTable.
type atomTableLine struct {
	cells          []Atom
	cellsBackColor []color.HColor
	lineBackColor  color.HColor
}

// AtomTable is a creole table: rows of cells with optional background
// colors, and the grid lines.
type AtomTable struct {
	AtomBase
	lines     []*atomTableLine
	positions map[Atom]Position
	lineColor color.HColor
}

// NewAtomTable returns an empty table whose grid has lineColor.
func NewAtomTable(lineColor color.HColor) *AtomTable {
	return &AtomTable{lineColor: lineColor, positions: map[Atom]Position{}}
}

// CalculateDimension returns the size of the grid.
func (a *AtomTable) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return a.Dimension(stringBounder, func(sb font.StringBounder) geom.XDimension2D {
		a.initMap(sb)
		return geom.NewXDimension2D(a.endingX(a.nbCols()-1), a.endingY(a.nbLines()-1))
	})
}

// StartingAltitude returns 0.
func (a *AtomTable) StartingAltitude(font.StringBounder) float64 { return 0 }

// DrawU draws the backgrounds, the cells and the grid.
func (a *AtomTable) DrawU(ug UGraphic) {
	a.initMap(ug.StringBounder())
	for i := 0; i < a.nbLines(); i++ {
		line := a.lines[i]
		if line.lineBackColor != nil {
			y1 := a.startingY(i)
			y2 := a.startingY(i + 1)
			x1 := a.startingX(0)
			x2 := a.startingX(a.nbCols())
			ug.Apply(color.None()).Apply(Bg(line.lineBackColor)).Apply(NewUTranslate(x1, y1)).
				Draw(URectangleBuild(x2-x1, y2-y1))
		}
		for j := 0; j < a.nbCols(); j++ {
			if j >= len(line.cells) {
				continue
			}
			cell := line.cells[j]
			align := geom.HorizontalAlignmentLeft
			if sb, ok := cell.(*SheetBlock1); ok {
				align = sb.CellAlignment()
			}
			cellBackColor := line.cellsBackColor[j]
			x1 := a.startingX(j)
			x2 := a.startingX(j + 1)
			cellWidth := x2 - x1
			if cellBackColor != nil {
				y1 := a.startingY(i)
				y2 := a.startingY(i + 1)
				ug.Apply(color.None()).Apply(Bg(cellBackColor)).Apply(NewUTranslate(x1, y1)).
					Draw(URectangleBuild(x2-x1, y2-y1))
			}
			pos := a.positions[cell]
			dimCell := cell.CalculateDimension(ug.StringBounder())
			dx := 0.0
			if align == geom.HorizontalAlignmentRight {
				dx = cellWidth - dimCell.Width()
			}
			if cellBackColor == nil {
				cell.DrawU(ug.Apply(pos.UTranslate().Compose(UTranslateDx(dx))))
			} else {
				cell.DrawU(ug.Apply(Bg(cellBackColor)).Apply(pos.UTranslate().Compose(UTranslateDx(dx))))
			}
		}
	}
	ug = ug.Apply(a.currentLineColor(ug))
	hline := ULineH(a.endingX(a.nbCols() - 1))
	for i := 0; i <= a.nbLines(); i++ {
		ug.Apply(UTranslateDy(a.startingY(i))).Draw(hline)
	}
	vline := ULineV(a.endingY(a.nbLines() - 1))
	for i := 0; i <= a.nbCols(); i++ {
		ug.Apply(UTranslateDx(a.startingX(i))).Draw(vline)
	}
}

func (a *AtomTable) currentLineColor(ug UGraphic) color.HColor {
	if scheme, ok := a.lineColor.(*color.HColorScheme); ok {
		backcolor := ug.Param().Backcolor()
		if backcolor.IsTransparent() {
			backcolor = ug.DefaultBackground()
		}
		return scheme.AppropriateColor(backcolor)
	}
	return a.lineColor
}

func (a *AtomTable) initMap(stringBounder font.StringBounder) {
	if len(a.positions) > 0 {
		return
	}
	for _, line := range a.lines {
		for _, cell := range line.cells {
			a.positions[cell] = NewPosition(0, 0, cell.CalculateDimension(stringBounder))
		}
	}
	for i, line := range a.lines {
		for j, cell := range line.cells {
			dim := cell.CalculateDimension(stringBounder)
			a.positions[cell] = NewPosition(a.startingX(j), a.startingY(i), dim)
		}
	}
}

func (a *AtomTable) startingX(col int) float64 {
	result := 0.0
	for i := 0; i < col; i++ {
		result += a.colWidth(i)
	}
	return result
}

func (a *AtomTable) endingX(col int) float64 {
	result := 0.0
	for i := 0; i <= col; i++ {
		result += a.colWidth(i)
	}
	return result
}

func (a *AtomTable) startingY(line int) float64 {
	result := 0.0
	for i := 0; i < line; i++ {
		result += a.lineHeight(i)
	}
	return result
}

func (a *AtomTable) endingY(line int) float64 {
	result := 0.0
	for i := 0; i <= line; i++ {
		result += a.lineHeight(i)
	}
	return result
}

func (a *AtomTable) colWidth(col int) float64 {
	result := 0.0
	for i := 0; i < a.nbLines(); i++ {
		if position, ok := a.position(i, col); ok {
			result = math.Max(result, position.Width())
		}
	}
	return result
}

func (a *AtomTable) lineHeight(line int) float64 {
	result := 0.0
	for i := 0; i < a.nbCols(); i++ {
		if position, ok := a.position(line, i); ok {
			result = math.Max(result, position.Height())
		}
	}
	return result
}

func (a *AtomTable) position(line, col int) (Position, bool) {
	if line >= len(a.lines) {
		return Position{}, false
	}
	l := a.lines[line]
	if col >= len(l.cells) {
		return Position{}, false
	}
	p, ok := a.positions[l.cells[col]]
	return p, ok
}

func (a *AtomTable) nbCols() int {
	if len(a.lines) == 0 {
		throw(indexOutOfBounds, "Index 0 out of bounds for length 0")
	}
	result := len(a.lines[0].cells)
	for _, line := range a.lines[1:] {
		result = max(result, len(line.cells))
	}
	return result
}

func (a *AtomTable) nbLines() int { return len(a.lines) }

// AddCell adds a cell to the last row; cellBackColor may be nil.
func (a *AtomTable) AddCell(cell Atom, cellBackColor color.HColor) {
	last := a.lines[len(a.lines)-1]
	last.cells = append(last.cells, cell)
	last.cellsBackColor = append(last.cellsBackColor, cellBackColor)
	clear(a.positions)
}

// NewLine starts a row; lineBackColor may be nil.
func (a *AtomTable) NewLine(lineBackColor color.HColor) {
	a.lines = append(a.lines, &atomTableLine{lineBackColor: lineBackColor})
}

// Neutrons returns the atom as one neutron.
func (a *AtomTable) Neutrons() []*Neutron { return atomNeutrons(a) }

// AtomTree is a creole tree ("|_" lines): cells indented by level and
// joined by lines.
type AtomTree struct {
	AtomBase
	lineColor color.HColor
	cells     []Atom
	levels    map[Atom]int
}

const atomTreeMargin = 2

// NewAtomTree returns an empty tree whose lines have lineColor.
func NewAtomTree(lineColor color.HColor) *AtomTree {
	return &AtomTree{lineColor: lineColor, levels: map[Atom]int{}}
}

// CalculateDimension returns the size of the tree.
func (a *AtomTree) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return a.Dimension(stringBounder, func(sb font.StringBounder) geom.XDimension2D {
		skeleton := &skeleton2{}
		width, height := 0.0, 0.0
		for _, cell := range a.cells {
			dim := cell.CalculateDimension(sb)
			height += dim.Height()
			level := a.level(cell)
			width = math.Max(width, skeleton.xEndForLevel(level)+atomTreeMargin+dim.Width())
		}
		return geom.NewXDimension2D(width, height)
	})
}

// StartingAltitude returns 0.
func (a *AtomTree) StartingAltitude(font.StringBounder) float64 { return 0 }

// DrawU draws the cells and the tree lines.
func (a *AtomTree) DrawU(ugInit UGraphic) {
	skeleton := &skeleton2{}
	y := 0.0
	ug := ugInit
	for _, cell := range a.cells {
		level := a.level(cell)
		cell.DrawU(ug.Apply(UTranslateDx(atomTreeMargin + skeleton.xEndForLevel(level))))
		dim := cell.CalculateDimension(ug.StringBounder())
		skeleton.add(level, y+dim.Height()/2)
		ug = ug.Apply(UTranslateDy(dim.Height()))
		y += dim.Height()
	}
	skeleton.draw(ugInit.Apply(a.lineColor))
}

func (a *AtomTree) level(atom Atom) int {
	level, ok := a.levels[atom]
	if !ok {
		throw(nullPointer, "")
	}
	return level
}

// AddCell adds a cell at a level.
func (a *AtomTree) AddCell(cell Atom, level int) {
	a.cells = append(a.cells, cell)
	a.levels[cell] = level
}

// Neutrons returns the atom as one neutron.
func (a *AtomTree) Neutrons() []*Neutron { return atomNeutrons(a) }

// skeleton2 is upstream's salt Skeleton2: the lines of a tree.
type skeleton2 struct {
	entries []skeleton2Entry
}

const skeleton2SizeX = 8

type skeleton2Entry struct {
	level int
	ypos  float64
}

func (e skeleton2Entry) drawHline(ug UGraphic) {
	xpos := skeleton2XStartForLevel(e.level)
	ug.Apply(NewUTranslate(xpos+skeleton2SizeX-1, e.ypos-1)).Draw(URectangleBuild(2, 2))
	ug.Apply(NewUTranslate(xpos, e.ypos)).Draw(ULineH(skeleton2SizeX))
}

func (e skeleton2Entry) drawVline(ug UGraphic, lastY float64) {
	xpos := skeleton2XStartForLevel(e.level)
	ug.Apply(NewUTranslate(xpos, lastY)).Draw(ULineV(e.ypos - lastY))
}

func (s *skeleton2) add(level int, y float64) {
	s.entries = append(s.entries, skeleton2Entry{level, y})
}

func (s *skeleton2) draw(ug UGraphic) {
	for i, en := range s.entries {
		en.drawHline(ug)
		lastY := 0.0
		if up, ok := s.motherOrSister(i); ok {
			lastY = up.ypos
		}
		en.drawVline(ug, lastY)
	}
}

func (s *skeleton2) motherOrSister(idx int) (skeleton2Entry, bool) {
	currentLevel := s.entries[idx].level
	for i := idx - 1; i >= 0; i-- {
		otherLevel := s.entries[i].level
		if otherLevel == currentLevel || otherLevel == currentLevel-1 {
			return s.entries[i], true
		}
	}
	return skeleton2Entry{}, false
}

func skeleton2XStartForLevel(level int) float64 { return float64(level) * skeleton2SizeX }

func (s *skeleton2) xEndForLevel(level int) float64 {
	return skeleton2XStartForLevel(level) + skeleton2SizeX
}

// AtomSprite is a sprite in a text ("<$name>").
type AtomSprite struct {
	AtomBase
	sprite      Sprite
	scale       float64
	url         *URL
	fontColor   color.HColor
	forcedColor color.HColor
	backColor   color.HColor
}

// NewAtomSprite returns the atom of a sprite; the colors and url may be
// nil.
func NewAtomSprite(fontColor, forcedColor color.HColor, scale float64, sprite Sprite, url *URL,
	backColor color.HColor) *AtomSprite {
	return &AtomSprite{sprite: sprite, scale: scale, url: url, fontColor: fontColor, forcedColor: forcedColor,
		backColor: backColor}
}

// CalculateDimension returns the size of the sprite.
func (a *AtomSprite) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return a.Dimension(stringBounder, func(sb font.StringBounder) geom.XDimension2D {
		return a.sprite.AsTextBlock(a.fontColor, a.forcedColor, a.scale, nil).CalculateDimension(sb)
	})
}

// StartingAltitude returns 0.
func (a *AtomSprite) StartingAltitude(font.StringBounder) float64 { return 0 }

// DrawU draws the sprite, inside its link if any.
func (a *AtomSprite) DrawU(ug UGraphic) {
	if a.url != nil {
		ug.StartURL(a.url)
	}
	a.sprite.AsTextBlock(a.fontColor, a.forcedColor, a.scale, a.backColor).DrawU(ug)
	if a.url != nil {
		ug.CloseURL()
	}
}

// Neutrons returns the atom as one neutron.
func (a *AtomSprite) Neutrons() []*Neutron { return atomNeutrons(a) }

// The data URL prefixes of images in texts.
const (
	AtomImgDataImagePngBase64 = "data:image/png;base64,"
	AtomImgDataImagePngSpm    = "data:image/png;spm "
	atomImgDataImageSvgBase64 = "data:image/svg+xml;base64,"
	// umlSourceBase64TagReplacement is upstream's
	// UmlSource.BASE64_TAG_REPLACEMENT: the preprocessor replaces long base64
	// images by their MD5.
	umlSourceBase64TagReplacement = "data:image/png;md5,"
)

// AtomImg is a raster image in a text ("<img:...>").
type AtomImg struct {
	AtomBase
	image       PortableImage
	scale       float64
	url         *URL
	rawFileName string
}

// AtomImgCreate returns the atom of an image source: a data URL (PNG or
// SVG in base64), or the MD5 of a base64 image the preprocessor set aside.
// Undecodable data gives the error text upstream draws. Files, URLs and the
// standard library images ("data:image/png;spm") are not ported yet: they
// panic.
func AtomImgCreate(skinParam ISkinSimple, src string, valign geom.ImgValign, vspace int, scale float64,
	url *URL) Atom {
	fc := FontConfigurationBlackBlueTrue(font.Monospace(14))
	if base64, ok := strings.CutPrefix(src, umlSourceBase64TagReplacement); ok {
		md5 := base64
		data, found := skinParam.FromMd5(md5)
		if !found {
			return AtomTextUtilsCreateLegacy("[md5:"+md5+"]", fc)
		}
		bytes, err := base64Decode(data)
		if err != nil {
			return AtomTextUtilsCreateLegacy("ERROR "+err.Error(), fc)
		}
		return atomImgFromRasterData(AtomImgDataImagePngBase64+data, fc, bytes, scale, url)
	}
	if data, ok := strings.CutPrefix(src, AtomImgDataImagePngBase64); ok {
		bytes, err := base64Decode(data)
		if err != nil {
			return AtomTextUtilsCreateLegacy("ERROR "+err.Error(), fc)
		}
		return atomImgFromRasterData(src, fc, bytes, scale, url)
	}
	if strings.HasPrefix(src, AtomImgDataImagePngSpm) {
		notPorted("klimt/creole/atom/AtomImg", "image of the standard library "+src)
	}
	if data, ok := strings.CutPrefix(src, atomImgDataImageSvgBase64); ok {
		bytes, err := base64Decode(data)
		if err != nil {
			return AtomTextUtilsCreateLegacy("ERROR "+err.Error(), fc)
		}
		return NewAtomImgSvg(NewTileImageSvg(string(bytes), scale))
	}
	notPorted("klimt/creole/atom/AtomImg", "image file or URL "+src)
	return nil
}

// atomImgFromRasterData is upstream's buildRasterFromData.
func atomImgFromRasterData(source string, fc *FontConfiguration, data []byte, scale float64, url *URL) Atom {
	read, err := readRasterImage(data)
	if err != nil {
		return AtomTextUtilsCreateLegacy("ERROR "+err.Error(), fc)
	}
	if read == nil {
		return AtomTextUtilsCreateLegacy("(Cannot decode: "+source+")", fc)
	}
	return &AtomImg{image: read, scale: scale, url: url}
}

// readRasterImage is upstream's SImageIO.read: the image of PNG, JPEG or
// GIF data, nil (without error) when the format is unknown. Java's ImageIO
// also reads BMP, WBMP and TIFF.
func readRasterImage(data []byte) (PortableImage, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if errors.Is(err, image.ErrFormat) {
		return nil, nil
	}
	if err != nil {
		return nil, errors.New("javax.imageio.IIOException: " + err.Error())
	}
	b := img.Bounds()
	nrgba := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(nrgba, nrgba.Bounds(), img, b.Min, draw.Src)
	result := NewPortableImageARGB(b.Dx(), b.Dy())
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			o := nrgba.PixOffset(x, y)
			p := nrgba.Pix[o : o+4]
			result.SetRGB(x, y, int32(uint32(p[3])<<24|uint32(p[0])<<16|uint32(p[1])<<8|uint32(p[2])))
		}
	}
	return result, nil
}

// CalculateDimension returns the scaled size of the image.
func (a *AtomImg) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return a.Dimension(stringBounder, func(font.StringBounder) geom.XDimension2D {
		return geom.NewXDimension2D(float64(a.image.Width())*a.scale, float64(a.image.Height())*a.scale)
	})
}

// StartingAltitude returns 0.
func (a *AtomImg) StartingAltitude(font.StringBounder) float64 { return 0 }

// DrawU draws the image, inside its link if any.
func (a *AtomImg) DrawU(ug UGraphic) {
	if a.url != nil {
		ug.StartURL(a.url)
	}
	ug.Draw(NewUImage(NewPixelImage(a.image, AffineTransformBilinear)).WithRawFileName(a.rawFileName).
		Scale(a.scale))
	if a.url != nil {
		ug.CloseURL()
	}
}

// Neutrons returns the atom as one neutron.
func (a *AtomImg) Neutrons() []*Neutron { return atomNeutrons(a) }

// AtomImgSvg is an SVG image in a text.
type AtomImgSvg struct {
	AtomBase
	tileImageSvg *TileImageSvg
}

// NewAtomImgSvg returns the atom of an SVG tile.
func NewAtomImgSvg(tileImageSvg *TileImageSvg) *AtomImgSvg {
	return &AtomImgSvg{tileImageSvg: tileImageSvg}
}

// CalculateDimension returns the size of the SVG.
func (a *AtomImgSvg) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return a.Dimension(stringBounder, a.tileImageSvg.CalculateDimension)
}

// StartingAltitude returns 0.
func (a *AtomImgSvg) StartingAltitude(font.StringBounder) float64 { return 0 }

// DrawU draws the SVG.
func (a *AtomImgSvg) DrawU(ug UGraphic) { a.tileImageSvg.DrawU(ug) }

// Neutrons returns the atom as one neutron.
func (a *AtomImgSvg) Neutrons() []*Neutron { return atomNeutrons(a) }

// base64Map2 is upstream's Base64Coder.map2: the value of each base64
// character, -1 for the others.
var base64Map2 = func() [128]int8 {
	var m [128]int8
	for i := range m {
		m[i] = -1
	}
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	for i := 0; i < len(alphabet); i++ {
		m[alphabet[i]] = int8(i)
	}
	return m
}()

// base64Decode is upstream's Base64Coder.decode(String), with its error
// messages (as Java's Throwable.toString).
func base64Decode(s string) ([]byte, error) {
	in := utf16Of(s)
	iLen := len(in)
	if iLen%4 != 0 {
		return nil, errors.New("java.lang.IllegalArgumentException: " +
			"Length of Base64 encoded input string is not a multiple of 4.")
	}
	for iLen > 0 && in[iLen-1] == '=' {
		iLen--
	}
	oLen := iLen * 3 / 4
	out := make([]byte, oLen)
	illegal := errors.New("java.lang.IllegalArgumentException: Illegal character in Base64 encoded data.")
	ip, op := 0, 0
	for ip < iLen {
		i0 := in[ip]
		i1 := in[ip+1]
		ip += 2
		i2, i3 := uint16('A'), uint16('A')
		if ip < iLen {
			i2 = in[ip]
			ip++
		}
		if ip < iLen {
			i3 = in[ip]
			ip++
		}
		if i0 > 127 || i1 > 127 || i2 > 127 || i3 > 127 {
			return nil, illegal
		}
		b0, b1, b2, b3 := int(base64Map2[i0]), int(base64Map2[i1]), int(base64Map2[i2]), int(base64Map2[i3])
		if b0 < 0 || b1 < 0 || b2 < 0 || b3 < 0 {
			return nil, illegal
		}
		out[op] = byte(b0<<2 | b1>>4)
		op++
		if op < oLen {
			out[op] = byte((b1&0xf)<<4 | b2>>2)
			op++
		}
		if op < oLen {
			out[op] = byte((b2&3)<<6 | b3)
			op++
		}
	}
	return out, nil
}

// CreoleHorizontalLine is a horizontal rule of a text ("----", "====",
// "....", or a title between the marks), drawn across the stencil.
type CreoleHorizontalLine struct {
	AtomBase
	fontConfiguration *FontConfiguration
	line              string
	style             rune
	skinParam         ISkinSimple
}

const creoleHorizontalLineDefaultThickness = 1

// CreoleHorizontalLineCreate returns a horizontal rule with an optional
// title (line, "" for none) and its style character.
func CreoleHorizontalLineCreate(fontConfiguration *FontConfiguration, line string, style rune,
	skinParam ISkinSimple) *CreoleHorizontalLine {
	return &CreoleHorizontalLine{fontConfiguration: fontConfiguration, line: line, style: style,
		skinParam: skinParam}
}

func (a *CreoleHorizontalLine) horizontalLine() *UHorizontalLine {
	if a.line == "" {
		return UHorizontalLineInfinite(creoleHorizontalLineDefaultThickness, 0, 0, a.style)
	}
	return UHorizontalLineInfiniteTitle(creoleHorizontalLineDefaultThickness, 0, 0, a.title(), a.style)
}

func (a *CreoleHorizontalLine) title() TextBlock {
	if a.line == "" {
		return TextBlockUtilsEmpty(0, 0)
	}
	parser := a.skinParam.Sheet(a.fontConfiguration, geom.HorizontalAlignmentLeft, CreoleModeFull)
	sheet := parser.CreateSheet(DisplayWithNewlines(a.skinParam.Pragma(), a.line))
	return NewSheetBlock1Padding(sheet, LineBreakStrategyNone, a.skinParam.Padding())
}

// DrawU draws the rule at half the height.
func (a *CreoleHorizontalLine) DrawU(ug UGraphic) {
	dim := a.CalculateDimension(ug.StringBounder())
	ug = ug.Apply(UTranslateDy(dim.Height() / 2))
	ug.Draw(a.horizontalLine())
}

// CalculateDimension returns 10x10 without title, else the size of the
// title.
func (a *CreoleHorizontalLine) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return a.Dimension(stringBounder, func(sb font.StringBounder) geom.XDimension2D {
		if a.line == "" {
			return geom.NewXDimension2D(10, 10)
		}
		return a.title().CalculateDimension(sb)
	})
}

// StartingAltitude returns 0.
func (a *CreoleHorizontalLine) StartingAltitude(font.StringBounder) float64 { return 0 }

// Neutrons returns the atom as one neutron.
func (a *CreoleHorizontalLine) Neutrons() []*Neutron { return atomNeutrons(a) }
