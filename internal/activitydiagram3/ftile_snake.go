// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/Snake.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/SnakeDirection.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"slices"
	"strings"

	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

// verticalAlignmentNull is a null vertical alignment.
const verticalAlignmentNull geom.VerticalAlignment = -1

// snakeText is a label of a snake.
type snakeText struct {
	textBlock           klimt.TextBlock
	verticalAlignment   geom.VerticalAlignment
	horizontalAlignment geom.HorizontalAlignment
}

func newSnakeText(textBlock klimt.TextBlock, verticalAlignment geom.VerticalAlignment,
	horizontalAlignment geom.HorizontalAlignment) snakeText {
	if textBlock == nil {
		uml.ThrowNPE("")
	}
	return snakeText{textBlock: textBlock, verticalAlignment: verticalAlignment,
		horizontalAlignment: horizontalAlignment}
}

func (t snakeText) hasText(stringBounder font.StringBounder) bool {
	return !klimt.TextBlockUtilsIsEmpty(t.textBlock, stringBounder)
}

// Snake is an arrow of the activity layout: a polyline (Worm) with its
// decorations, colors and labels. Snakes are drawn through UGraphicForSnake,
// which merges the snakes that continue each other.
type Snake struct {
	worm               *Worm
	startDecoration    *klimt.UPolygon
	endDecoration      *klimt.UPolygon
	color              *Rainbow
	texts              *[]snakeText
	mergeable          MergeStrategy
	emphasizeDirection klimt.Direction
	hasEmphasize       bool
	skinParam          klimt.ISkinParam
}

// Move returns the snake moved by (dx, dy); the labels are shared, as
// upstream shares its list.
func (s *Snake) Move(dx, dy float64) *Snake {
	return newSnake(s.skinParam, s.startDecoration, s.color, s.endDecoration, s.worm.Move(dx, dy), s.mergeable,
		s.emphasizeDirection, s.hasEmphasize, s.texts)
}

// IgnoreForCompression marks the snake as ignored by the compression on x;
// it returns the snake.
func (s *Snake) IgnoreForCompression() *Snake {
	s.worm.SetIgnoreForCompression()
	return s
}

// EmphasizeDirection returns the snake with an arrow head drawn on the
// first segment of that direction.
func (s *Snake) EmphasizeDirection(emphasizeDirection klimt.Direction) *Snake {
	return newSnake(s.skinParam, s.startDecoration, s.color, s.endDecoration, s.worm, s.mergeable,
		emphasizeDirection, true, s.texts)
}

// WithoutEndDecoration returns the snake without end decoration.
func (s *Snake) WithoutEndDecoration() *Snake {
	return newSnake(s.skinParam, s.startDecoration, s.color, nil, s.worm, s.mergeable, s.emphasizeDirection,
		s.hasEmphasize, s.texts)
}

// WithMerge returns the snake with another merge strategy.
func (s *Snake) WithMerge(mergeable MergeStrategy) *Snake {
	return newSnake(s.skinParam, s.startDecoration, s.color, s.endDecoration, s.worm, mergeable,
		s.emphasizeDirection, s.hasEmphasize, s.texts)
}

// WithLabel adds a label aligned horizontally; a nil text block adds
// nothing. It returns the snake.
func (s *Snake) WithLabel(textBlock klimt.TextBlock, horizontalAlignment geom.HorizontalAlignment) *Snake {
	if textBlock != nil {
		*s.texts = append(*s.texts, newSnakeText(textBlock, verticalAlignmentNull, horizontalAlignment))
	}
	return s
}

// WithLabelVertical adds a label aligned vertically; a nil or empty text
// block adds nothing. It returns the snake.
func (s *Snake) WithLabelVertical(textBlock klimt.TextBlock, verticalAlignment geom.VerticalAlignment) *Snake {
	if textBlock != nil && textBlock != klimt.TextBlockUtilsEmptyTextBlock() {
		*s.texts = append(*s.texts, newSnakeText(textBlock, verticalAlignment, klimt.HorizontalAlignmentNull))
	}
	return s
}

// SnakeCreate returns an empty snake of color, without decorations.
func SnakeCreate(skinParam klimt.ISkinParam, color *Rainbow) *Snake {
	return SnakeCreateStartEnd(skinParam, nil, color, nil)
}

// SnakeCreateEnd returns an empty snake of color ending with endDecoration
// (nil for none).
func SnakeCreateEnd(skinParam klimt.ISkinParam, color *Rainbow, endDecoration *klimt.UPolygon) *Snake {
	return SnakeCreateStartEnd(skinParam, nil, color, endDecoration)
}

// SnakeCreateStartEnd returns an empty snake of color with decorations at
// both ends (nil for none).
func SnakeCreateStartEnd(skinParam klimt.ISkinParam, startDecoration *klimt.UPolygon, color *Rainbow,
	endDecoration *klimt.UPolygon) *Snake {
	style := klimt.StyleSignatureBasicActivityArrow().MergedStyle(skinParam.CurrentStyleBuilder())
	texts := []snakeText{}
	return newSnake(skinParam, startDecoration, color, endDecoration, NewWorm(style, skinParam.Arrows()), MergeFull,
		0, false, &texts)
}

func newSnake(skinParam klimt.ISkinParam, startDecoration *klimt.UPolygon, color *Rainbow,
	endDecoration *klimt.UPolygon, worm *Worm, mergeable MergeStrategy, emphasizeDirection klimt.Direction,
	hasEmphasize bool, texts *[]snakeText) *Snake {
	if color == nil || texts == nil {
		uml.ThrowNPE("")
	}
	if color.Size() == 0 {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	return &Snake{skinParam: skinParam, worm: worm, texts: texts, emphasizeDirection: emphasizeDirection,
		hasEmphasize: hasEmphasize, startDecoration: startDecoration, endDecoration: endDecoration, color: color,
		mergeable: mergeable}
}

// Translate returns the snake moved by translate.
func (s *Snake) Translate(translate klimt.UTranslate) *Snake {
	return s.Move(translate.Dx(), translate.Dy())
}

// AddPoint appends a point.
func (s *Snake) AddPoint(x, y float64) { s.worm.AddPoint(x, y) }

// AddPointPt appends a point.
func (s *Snake) AddPointPt(p geom.XPoint2D) { s.AddPoint(p.X, p.Y) }

// DrawInternal draws the snake.
func (s *Snake) DrawInternal(ug klimt.UGraphic) {
	if s.color.Size() > 1 {
		s.drawRainbow(ug)
		return
	}
	s.worm.drawInternalOneColor(s.startDecoration, ug, s.color.Colors()[0], 1.5, s.emphasizeDirection,
		s.hasEmphasize, s.endDecoration)
	s.drawInternalLabel(ug)
}

func (s *Snake) drawRainbow(ug klimt.UGraphic) {
	colors := s.color.Colors()
	colorArrowSeparationSpace := s.color.ColorArrowSeparationSpace()
	move := float64(2 + colorArrowSeparationSpace)
	mutation := WormMutationCreate(s.worm, move)
	if mutation.IsDxNegative() {
		colors = slices.Clone(colors)
		slices.Reverse(colors)
	}
	globalMove := -1.0 * float64(len(colors)-1) / 2.0
	current := s.worm.MoveFirstPoint(mutation.First().MultiplyBy(globalMove))
	if mutation.Size() > 2 {
		current = current.MoveLastPoint(mutation.Last().MultiplyBy(globalMove))
	}
	for i := range colors {
		stroke := 1.5
		if colorArrowSeparationSpace == 0 {
			if i == len(colors)-1 {
				stroke = 2.0
			} else {
				stroke = 3.0
			}
		}
		current.drawInternalOneColor(s.startDecoration, ug, colors[i], stroke, s.emphasizeDirection, s.hasEmphasize,
			s.endDecoration)
		current = mutation.Mute(current)
	}
	textTranslate := mutation.TextTranslate(len(colors))
	s.drawInternalLabel(ug.Apply(textTranslate))
}

func (s *Snake) drawInternalLabel(ug klimt.UGraphic) {
	for _, text := range *s.texts {
		if text.hasText(ug.StringBounder()) {
			position := s.textBlockPosition(ug.StringBounder(), text)
			text.textBlock.DrawU(ug.Apply(klimt.UTranslatePoint(position)))
		}
	}
}

// MaxX returns the right of the snake and its labels.
func (s *Snake) MaxX(stringBounder font.StringBounder) float64 {
	result := s.worm.MaxX()
	for _, text := range *s.texts {
		position := s.textBlockPosition(stringBounder, text)
		dim := text.textBlock.CalculateDimension(stringBounder)
		result = max(result, position.X+dim.Width())
	}
	return result
}

func (s *Snake) textBlockPosition(stringBounder font.StringBounder, text snakeText) geom.XPoint2D {
	pt1 := s.worm.Point(0)
	pt2 := s.worm.Point(1)
	dim := text.textBlock.CalculateDimension(stringBounder)
	x := max(pt1.X, pt2.X) + 4
	code := s.worm.DirectionsCode()
	zigzag := strings.HasPrefix(code, "DLD") || strings.HasPrefix(code, "DRD")
	y := (pt1.Y+pt2.Y)/2 - dim.Height()/2
	switch {
	case text.verticalAlignment == geom.VerticalAlignmentBottom:
		x = s.worm.MinX()
		y = s.worm.MaxY()
	case text.verticalAlignment == geom.VerticalAlignmentCenter:
		x = s.worm.MinX()
		y = (s.worm.First().Y+s.worm.Last().Y-10)/2 - dim.Height()/2
	case text.horizontalAlignment == geom.HorizontalAlignmentCenter && zigzag:
		pt3 := s.worm.Point(2)
		x = (pt2.X+pt3.X)/2 - dim.Width()/2
	case text.horizontalAlignment == geom.HorizontalAlignmentRight && zigzag:
		x = max(pt1.X, pt2.X) - dim.Width() - 4
	case code == "RD":
		x = max(pt1.X, pt2.X)
		y = (pt1.Y+s.worm.Point(2).Y)/2 - dim.Height()/2
	case code == "LD":
		x = min(pt1.X, pt2.X)
		y = (pt1.Y+s.worm.Point(2).Y)/2 - dim.Height()/2
	}
	return geom.NewXPoint2D(x, y)
}

// HorizontalLines returns the horizontal segments of the snake.
func (s *Snake) HorizontalLines() []geom.XLine2D {
	var result []geom.XLine2D
	for i := 0; i < s.worm.Size()-1; i++ {
		pt1 := s.worm.Point(i)
		pt2 := s.worm.Point(i + 1)
		if pt1.Y == pt2.Y {
			result = append(result, geom.XLine2DLine(pt1, pt2))
		}
	}
	return result
}

// First returns the first point.
func (s *Snake) First() geom.XPoint2D { return s.worm.Point(0) }

// CannotBeTouched reports whether the snake cannot be the target of a
// touches check.
func (s *Snake) CannotBeTouched() bool { return s.mergeable != MergeFull || s.worm.IsPureHorizontal() }

// Last returns the last point.
func (s *Snake) Last() geom.XPoint2D { return s.worm.Point(s.worm.Size() - 1) }

// snakeSame reports whether two points are the same, within 0.001.
func snakeSame(pt1, pt2 geom.XPoint2D) bool { return pt1.Distance(pt2) < 0.001 }

// Merge returns the snake joining s and other when one continues the
// other, nil otherwise.
func (s *Snake) Merge(other *Snake, stringBounder font.StringBounder) *Snake {
	strategy := s.mergeable.Max(other.mergeable)
	if strategy == MergeNone {
		return nil
	}
	for _, text := range *other.texts {
		if text.hasText(stringBounder) {
			return nil
		}
	}
	if snakeSame(s.Last(), other.First()) {
		oneOf := other.endDecoration
		if oneOf == nil {
			oneOf = s.endDecoration
		}
		if s.startDecoration != nil || other.startDecoration != nil {
			uml.Throw(uml.UnsupportedOperationException, "Not yet coded: to be done")
		}
		mergeTexts := slices.Clone(*s.texts)
		mergeTexts = append(mergeTexts, *other.texts...)
		if mergeTexts == nil {
			mergeTexts = []snakeText{}
		}
		emphasize, hasEmphasize := s.emphasizeDirection, s.hasEmphasize
		if !hasEmphasize {
			emphasize, hasEmphasize = other.emphasizeDirection, other.hasEmphasize
		}
		return newSnake(s.skinParam, nil, s.color, oneOf, s.worm.Merge(other.worm, strategy), strategy, emphasize,
			hasEmphasize, &mergeTexts)
	}
	if snakeSame(s.First(), other.Last()) {
		return other.Merge(s, stringBounder)
	}
	return nil
}

// Touches reports whether other starts where s ends.
func (s *Snake) Touches(other *Snake) bool {
	if other.mergeable != MergeFull {
		return false
	}
	if other.worm.IsPureHorizontal() {
		return false
	}
	return snakeSame(s.Last(), other.First())
}

// DoesHorizontalCross reports whether a horizontal segment crosses minMax.
func (s *Snake) DoesHorizontalCross(minMax geom.MinMax) bool {
	return s.worm.DoesHorizontalCross(minMax)
}

// SnakeDirection is the direction of the first segment of a snake.
type SnakeDirection int

// Snake directions, in upstream's order.
const (
	SnakeVerticalThenHorizontal SnakeDirection = iota
	SnakeHorizontalThenVertical
)

// SnakeDirectionOf returns the direction of the segment from pt1 to pt2;
// a diagonal segment panics, as upstream.
func SnakeDirectionOf(pt1, pt2 geom.XPoint2D) SnakeDirection {
	if pt1.X == pt2.X {
		return SnakeVerticalThenHorizontal
	}
	if pt1.Y == pt2.Y {
		return SnakeHorizontalThenVertical
	}
	uml.Throw(uml.IllegalArgumentException, "")
	return SnakeVerticalThenHorizontal
}
