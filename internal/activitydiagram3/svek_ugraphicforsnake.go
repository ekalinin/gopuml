// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/UGraphicForSnake.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// UGraphicForSnake collects the snakes drawn on it, merging those that
// continue each other, and draws them on FlushUg. Upstream's class lives in
// svek; it is here because it handles activitydiagram3's Snake.
type UGraphicForSnake struct {
	klimt.UGraphicDelegator
	dx, dy float64
	snakes *[]*pendingSnake
}

var _ klimt.UGraphic = (*UGraphicForSnake)(nil)

type pendingSnake struct {
	snake  *Snake
	ug     klimt.UGraphic
	dx, dy float64
}

func (p *pendingSnake) drawInternal() { p.snake.DrawInternal(p.ug) }

func (p *pendingSnake) removeEndDecorationIfTouches(snakes []*pendingSnake) {
	for _, other := range snakes {
		if p.touchesOther(other) {
			p.snake = p.snake.WithoutEndDecoration()
			return
		}
	}
}

func (p *pendingSnake) touchesOther(other *pendingSnake) bool {
	if other.snake.CannotBeTouched() {
		return false
	}
	thisLast := p.snake.Last()
	otherFirst := other.snake.First()
	return snakeSame(geom.NewXPoint2D(thisLast.X+p.dx, thisLast.Y+p.dy),
		geom.NewXPoint2D(otherFirst.X+other.dx, otherFirst.Y+other.dy))
}

func (p *pendingSnake) merge(newItem *pendingSnake) *pendingSnake {
	s1 := p.snake.Move(p.dx, p.dy)
	s2 := newItem.snake.Move(newItem.dx, newItem.dy)
	merge := s1.Merge(s2, p.ug.StringBounder())
	if merge == nil {
		return nil
	}
	return &pendingSnake{snake: merge.Move(-p.dx, -p.dy), ug: p.ug, dx: p.dx, dy: p.dy}
}

// NewUGraphicForSnake returns the collector over ug.
func NewUGraphicForSnake(ug klimt.UGraphic) *UGraphicForSnake {
	return &UGraphicForSnake{UGraphicDelegator: klimt.NewUGraphicDelegator(ug), snakes: &[]*pendingSnake{}}
}

// Translation returns the translation applied to the collector.
func (u *UGraphicForSnake) Translation() klimt.UTranslate { return klimt.NewUTranslate(u.dx, u.dy) }

// Draw collects snakes and draws the other shapes.
func (u *UGraphicForSnake) Draw(shape klimt.UShape) {
	if snake, ok := shape.(*Snake); ok {
		u.addPendingSnake(snake)
		return
	}
	u.Ug().Draw(shape)
}

func (u *UGraphicForSnake) addPendingSnake(snake *Snake) {
	newItem := &pendingSnake{snake: snake, ug: u.Ug(), dx: u.dx, dy: u.dy}
	for pos, s := range *u.snakes {
		if merge := s.merge(newItem); merge != nil {
			(*u.snakes)[pos] = merge
			return
		}
	}
	*u.snakes = append(*u.snakes, newItem)
}

// FlushUg draws the collected snakes.
func (u *UGraphicForSnake) FlushUg() {
	for _, snake := range *u.snakes {
		snake.removeEndDecorationIfTouches(*u.snakes)
		snake.drawInternal()
	}
	*u.snakes = (*u.snakes)[:0]
}

// Apply returns the collector over the changed graphic, sharing the
// collected snakes.
func (u *UGraphicForSnake) Apply(change klimt.UChange) klimt.UGraphic {
	newdx := u.dx
	newdy := u.dy
	if t, ok := change.(klimt.UTranslate); ok {
		newdx += t.Dx()
		newdy += t.Dy()
	}
	return &UGraphicForSnake{UGraphicDelegator: klimt.NewUGraphicDelegator(u.Ug().Apply(change)), dx: newdx,
		dy: newdy, snakes: u.snakes}
}
