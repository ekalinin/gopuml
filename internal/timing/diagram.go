// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/TimingDiagram.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/Clocks.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package timing

import (
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/jmath"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/tim"
	"github.com/ekalinin/gopuml/internal/uml"
)

// marginX1 and marginX2 are the margins on the left and the right of the
// lanes.
const (
	marginX1 = 5.0
	marginX2 = 5.0
)

// TimingDiagram is upstream's TimingDiagram: the players in the order of
// their declarations, the named times, the messages and the highlights, on
// one ruler.
type TimingDiagram struct {
	uml.Base

	// codes are the named times; codeOrder holds their names in the order
	// they were first given, for the iteration order of upstream's HashMap.
	codes     map[string]*TimeTick
	codeOrder []string
	// players are kept by code; playerOrder holds the codes in the order of
	// their first declaration, as upstream's LinkedHashMap.
	players     map[string]player
	playerOrder []string
	clocks      map[string]*playerClock
	messages    []*timeMessage
	highlights  []*highlight
	ruler       *TimingRuler
	now         *TimeTick
	lastPlayer  player

	timeAxisStategy  TimeAxisStategy
	compactByDefault bool
	sdf              *tim.DateFormat
}

// NewTimingDiagram returns an empty timing diagram for src; md5map holds
// the base64 images of the source.
func NewTimingDiagram(ctx *core.RenderContext, src *core.UmlSource, md5map map[string]string) *TimingDiagram {
	d := &TimingDiagram{Base: uml.NewBase(ctx, src, core.Timing, md5map), codes: map[string]*TimeTick{},
		players: map[string]player{}, clocks: map[string]*playerClock{}}
	d.ruler = newTimingRuler(d.SkinParam())
	return d
}

var _ core.TypedDiagram = (*TimingDiagram)(nil)

// Description implements core.Diagram: "(Timing Diagram)".
func (d *TimingDiagram) Description() string { return "(Timing Diagram)" }

// timingStyle is upstream's StyleQueries.TIMINGDIAG with the SNames that
// follow timingDiagram.
func timingStyle(names ...klimt.SName) klimt.StyleSignatureBasic {
	return klimt.StyleSignatureBasicOf(append([]klimt.SName{klimt.SNameRoot, klimt.SNameElement,
		klimt.SNameTimingDiagram}, names...)...)
}

// namedTime is a named time, in the iteration order of upstream's codes.
type namedTime struct {
	name string
	tick *TimeTick
}

// namedTimes returns the named times in the iteration order of upstream's
// HashMap.
func (d *TimingDiagram) namedTimes() []namedTime {
	var result []namedTime
	for _, name := range klimt.JavaHashMapOrder(d.codeOrder) {
		result = append(result, namedTime{name: name, tick: d.codes[name]})
	}
	return result
}

// playersInOrder returns the players in the order of their declarations.
func (d *TimingDiagram) playersInOrder() []player {
	result := make([]player, len(d.playerOrder))
	for i, code := range d.playerOrder {
		result[i] = d.players[code]
	}
	return result
}

// putPlayer adds a player; a second player with the same code replaces the
// first one at its place.
func (d *TimingDiagram) putPlayer(code string, p player) {
	if _, ok := d.players[code]; !ok {
		d.playerOrder = append(d.playerOrder, code)
	}
	d.players[code] = p
}

func (d *TimingDiagram) createPlayerRobust(code, full string, compact bool, stereotype *klimt.Stereotype,
	backColor color.HColor) player {
	p := newPlayerRobust(full, d.SkinParam(), d.ruler, d.compactByDefault || compact, stereotype, backColor)
	d.putPlayer(code, p)
	d.lastPlayer = p
	return p
}

func (d *TimingDiagram) createPlayerConcise(code, full string, compact bool, stereotype *klimt.Stereotype,
	backColor color.HColor) player {
	p := newPlayerConcise(full, d.SkinParam(), d.ruler, d.compactByDefault || compact, stereotype, backColor)
	d.putPlayer(code, p)
	d.lastPlayer = p
	return p
}

func (d *TimingDiagram) createPlayerRectangle(code, full string, compact bool, stereotype *klimt.Stereotype,
	backColor color.HColor) player {
	p := newPlayerRectangle(full, d.SkinParam(), d.ruler, d.compactByDefault || compact, stereotype, backColor)
	d.putPlayer(code, p)
	d.lastPlayer = p
	return p
}

// createPlayerClock adds a clock, whose period becomes a time of the axis.
// It ignores compact, as upstream does.
func (d *TimingDiagram) createPlayerClock(code, full string, period, pulse, offset jmath.BigDecimal,
	stereotype *klimt.Stereotype) *playerClock {
	p := newPlayerClock(full, d.SkinParam(), d.ruler, period, pulse, offset, d.compactByDefault, stereotype)
	d.putPlayer(code, p)
	d.clocks[code] = p
	d.ruler.addTime(newTimeTick(period, formatDecimal))
	return p
}

func (d *TimingDiagram) createPlayerAnalog(code, full string, stereotype *klimt.Stereotype) *playerAnalog {
	p := newPlayerAnalog(full, d.SkinParam(), d.ruler, d.compactByDefault, stereotype)
	d.putPlayer(code, p)
	return p
}

func (d *TimingDiagram) createPlayerBinary(code, full string, stereotype *klimt.Stereotype) player {
	p := newPlayerBinary(full, d.SkinParam(), d.ruler, d.compactByDefault, stereotype)
	d.putPlayer(code, p)
	return p
}

func (d *TimingDiagram) createTimeMessage(player1 player, time1 *TimeTick, player2 player, time2 *TimeTick,
	label string, hasLabel bool) *timeMessage {
	message := newTimeMessage(tickInPlayer{player1, time1}, tickInPlayer{player2, time2}, label, hasLabel,
		d.SkinParam())
	d.messages = append(d.messages, message)
	return message
}

// addTime sets the current time and adds it to the axis; a code names it.
func (d *TimingDiagram) addTime(t *TimeTick, code string) {
	d.now = t
	d.ruler.addTime(t)
	if code != "" {
		if _, ok := d.codes[code]; !ok {
			d.codeOrder = append(d.codeOrder, code)
		}
		d.codes[code] = t
	}
}

// codeValue is the tick of a named time, nil for an unknown name.
func (d *TimingDiagram) codeValue(code string) *TimeTick { return d.codes[code] }

func (d *TimingDiagram) updateNow(t *TimeTick) { d.now = t }

// player returns the player of a code, nil for none.
func (d *TimingDiagram) player(code string) player { return d.players[code] }

func (d *TimingDiagram) getNow() *TimeTick { return d.now }

// clockValue is nb periods of a clock, without its offset; nil for an
// unknown clock.
func (d *TimingDiagram) clockValue(clockName string, nb int) *TimeTick {
	clock := d.clocks[clockName]
	if clock == nil {
		return nil
	}
	return newTimeTick(clock.period.Multiply(jmath.BigDecimalOfInt64(int64(nb))), formatDecimal)
}

func (d *TimingDiagram) setLastPlayer(p player) { d.lastPlayer = p }

func (d *TimingDiagram) getLastPlayer() player { return d.lastPlayer }

func (d *TimingDiagram) scaleInPixels(tick, pixel int64) { d.ruler.scaleInPixels(tick, pixel) }

func (d *TimingDiagram) setTimeAxisStategy(newStrategy TimeAxisStategy) command.Result {
	d.timeAxisStategy = newStrategy
	return command.Ok()
}

func (d *TimingDiagram) highlight(tickFrom, tickTo *TimeTick, caption *klimt.Display,
	colors *color.Colors) command.Result {
	d.highlights = append(d.highlights, newHighlight(d.SkinParam(), tickFrom, tickTo, caption, colors))
	return command.Ok()
}

func (d *TimingDiagram) goCompactMode() { d.compactByDefault = true }

// useDateFormat is the date format of the dates that follow; a pattern
// that SimpleDateFormat rejects fails with "Bad date format".
func (d *TimingDiagram) useDateFormat(dateFormat string) command.Result {
	sdf, err := tim.NewDateFormat(dateFormat)
	if err != nil {
		return command.Error("Bad date format")
	}
	d.sdf = sdf
	return command.Ok()
}

// timingFormatDate is the format of the dates: DATE, or the date format.
func (d *TimingDiagram) timingFormatDate() *TimingFormat {
	if d.sdf == nil {
		return formatDate
	}
	return newTimingFormat(d.sdf)
}
