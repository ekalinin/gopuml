// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ftile/TextBlockInterceptorUDrawable.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ftile/UGraphicDispatchFtile.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/UGraphicInterceptorOneSwimlane.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/UGraphicInterceptorAllSwimlanes.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"io"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// textBlockInterceptorUDrawable draws a diagram without lanes through
// UGraphicDispatchFtile.
type textBlockInterceptorUDrawable struct {
	klimt.TextBlockDefaults
	textBlock klimt.TextBlock
	gotoColor color.HColor
	isDebug   bool
}

func newTextBlockInterceptorUDrawable(textBlock klimt.TextBlock, gotoColor color.HColor,
	isDebug bool) *textBlockInterceptorUDrawable {
	return &textBlockInterceptorUDrawable{textBlock: textBlock, gotoColor: gotoColor, isDebug: isDebug}
}

func (t *textBlockInterceptorUDrawable) DrawU(ug klimt.UGraphic) {
	newUGraphicDispatchFtile(ug, map[string]klimt.UTranslate{}, t.gotoColor, t.isDebug).Draw(t.textBlock)
	ug.FlushUg()
}

func (t *textBlockInterceptorUDrawable) CalculateDimension(font.StringBounder) geom.XDimension2D {
	uml.Throw(uml.UnsupportedOperationException, "")
	return geom.XDimension2D{}
}

// atomTextShape is creole's AtomText, which the low-level driver draws.
type atomTextShape = *klimt.AtomText

// UGraphicDispatchFtile draws the tiles on itself, records the position of
// the labels and draws the "goto" lines.
type UGraphicDispatchFtile struct {
	klimt.UGraphicDelegator
	positions map[string]klimt.UTranslate
	gotoColor color.HColor
	isDebug   bool
}

func newUGraphicDispatchFtile(ug klimt.UGraphic, positions map[string]klimt.UTranslate, gotoColor color.HColor,
	isDebug bool) *UGraphicDispatchFtile {
	return &UGraphicDispatchFtile{UGraphicDelegator: klimt.NewUGraphicDelegator(ug), positions: positions,
		gotoColor: gotoColor, isDebug: isDebug}
}

// Draw implements klimt.UGraphic.
func (u *UGraphicDispatchFtile) Draw(shape klimt.UShape) {
	if ftile, ok := shape.(Ftile); ok {
		ftile.DrawU(u)
		if label, ok := ftile.(*FtileLabel); ok {
			u.positions[label.Name()] = u.position()
		}
		if g, ok := ftile.(*FtileGoto); ok {
			u.drawGoto(g)
		}
		return
	}
	_, isImage := shape.(*klimt.UImage)
	_, isAtomText := shape.(atomTextShape)
	if drawable, ok := shape.(klimt.UDrawable); ok && !isImage && !isAtomText {
		// UImage implements TextBlock but must be drawn by the low-level
		// driver; AtomText too, to avoid an infinite recursion with
		// SPECIALTXT
		drawable.DrawU(u)
		return
	}
	u.Ug().Draw(shape)
}

func (u *UGraphicDispatchFtile) position() klimt.UTranslate {
	snake, ok := u.Ug().(*UGraphicForSnake)
	if !ok {
		uml.Throw(uml.ClassCastException, "")
	}
	return snake.Translation()
}

func (u *UGraphicDispatchFtile) drawGoto(ftile *FtileGoto) {
	g := ftile.Geometry(u.StringBounder())
	pt := g.PointIn()
	ugGoto := u.Ug().Apply(u.gotoColor).Apply(klimt.Bg(u.gotoColor))
	ugGoto = ugGoto.Apply(klimt.UTranslatePoint(pt))
	posNow := u.position()
	dest, ok := u.positions[ftile.Name()]
	if !ok {
		return
	}
	dx := dest.Dx() - posNow.Dx()
	dy := dest.Dy() - posNow.Dy()
	if u.isDebug {
		ugGoto.Draw(klimt.UEllipseBuild(3, 3))
		ugGoto.Apply(klimt.NewUTranslate(dx, dy)).Draw(klimt.UEllipseBuild(3, 3))
	}
	ugGoto.Draw(klimt.ULineH(dx))
	ugGoto.Apply(klimt.UTranslateDx(dx)).Draw(klimt.ULineV(dy))
}

// Apply implements klimt.UGraphic.
func (u *UGraphicDispatchFtile) Apply(change klimt.UChange) klimt.UGraphic {
	return newUGraphicDispatchFtile(u.Ug().Apply(change), u.positions, u.gotoColor, u.isDebug)
}

// UGraphicInterceptorOneSwimlane draws the parts of the tiles that lie in
// one lane.
type UGraphicInterceptorOneSwimlane struct {
	klimt.UGraphicDelegator
	swimlane    *Swimlane
	orderedList []*Swimlane
}

func newUGraphicInterceptorOneSwimlane(ug klimt.UGraphic, swimlane *Swimlane,
	orderedList []*Swimlane) *UGraphicInterceptorOneSwimlane {
	return &UGraphicInterceptorOneSwimlane{UGraphicDelegator: klimt.NewUGraphicDelegator(ug), swimlane: swimlane,
		orderedList: orderedList}
}

// Draw implements klimt.UGraphic.
func (u *UGraphicInterceptorOneSwimlane) Draw(shape klimt.UShape) {
	switch s := shape.(type) {
	case Ftile:
		if s.Swimlanes().Contains(u.swimlane) {
			s.DrawU(u)
		}
	case Connection:
		tile1 := s.Ftile1()
		tile2 := s.Ftile2()
		contained1 := tile1 == nil || tile1.SwimlaneOut() == nil || tile1.SwimlaneOut() == u.swimlane
		contained2 := tile2 == nil || tile2.SwimlaneIn() == nil || tile2.SwimlaneIn() == u.swimlane
		if contained1 && contained2 {
			s.DrawU(u)
		}
	default:
		u.Ug().Draw(shape)
	}
}

// Apply implements klimt.UGraphic.
func (u *UGraphicInterceptorOneSwimlane) Apply(change klimt.UChange) klimt.UGraphic {
	return newUGraphicInterceptorOneSwimlane(u.Ug().Apply(change), u.swimlane, u.orderedList)
}

// Swimlane returns the lane.
func (u *UGraphicInterceptorOneSwimlane) Swimlane() *Swimlane { return u.swimlane }

// OrderedListOfAllSwimlanes returns the lanes of the diagram.
func (u *UGraphicInterceptorOneSwimlane) OrderedListOfAllSwimlanes() []*Swimlane {
	return u.orderedList
}

// swimlaneUgMap is a LinkedHashMap from lanes to graphics.
type swimlaneUgMap struct {
	keys []*Swimlane
	ugs  map[*Swimlane]klimt.UGraphic
}

func newSwimlaneUgMap() *swimlaneUgMap { return &swimlaneUgMap{ugs: map[*Swimlane]klimt.UGraphic{}} }

func (m *swimlaneUgMap) put(swimlane *Swimlane, ug klimt.UGraphic) {
	if _, ok := m.ugs[swimlane]; !ok {
		m.keys = append(m.keys, swimlane)
	}
	m.ugs[swimlane] = ug
}

// UGraphicInterceptorAllSwimlanes draws the tiles once on the graphics of
// every lane, each receiving the parts in its lane.
type UGraphicInterceptorAllSwimlanes struct {
	swimlaneToUg    *swimlaneUgMap
	orderedList     []*Swimlane
	activeSwimlanes []*Swimlane
	stringBounder   font.StringBounder
}

var _ klimt.UGraphic = (*UGraphicInterceptorAllSwimlanes)(nil)

func newUGraphicInterceptorAllSwimlanes(swimlaneToUg *swimlaneUgMap, orderedList []*Swimlane,
	stringBounder font.StringBounder) *UGraphicInterceptorAllSwimlanes {
	return &UGraphicInterceptorAllSwimlanes{swimlaneToUg: swimlaneToUg, orderedList: orderedList,
		activeSwimlanes: swimlaneToUg.keys, stringBounder: stringBounder}
}

// Draw implements klimt.UGraphic.
func (u *UGraphicInterceptorAllSwimlanes) Draw(shape klimt.UShape) {
	switch s := shape.(type) {
	case Ftile:
		tileSwimlanes := s.Swimlanes()
		hasMatch := false
		for _, swimlane := range u.activeSwimlanes {
			if tileSwimlanes.Contains(swimlane) {
				hasMatch = true
				break
			}
		}
		if hasMatch {
			s.DrawU(u.withActiveSwimlanes(tileSwimlanes))
		}
	case Connection:
		tile1 := s.Ftile1()
		tile2 := s.Ftile2()
		for _, swimlane := range u.activeSwimlanes {
			contained1 := tile1 == nil || tile1.SwimlaneOut() == nil || tile1.SwimlaneOut() == swimlane
			contained2 := tile2 == nil || tile2.SwimlaneIn() == nil || tile2.SwimlaneIn() == swimlane
			if contained1 && contained2 {
				if ug := u.swimlaneToUg.ugs[swimlane]; ug != nil {
					s.DrawU(ug)
				}
			}
		}
	default:
		// Primitive shape: dispatch only to the active lanes
		for _, swimlane := range u.activeSwimlanes {
			if ug := u.swimlaneToUg.ugs[swimlane]; ug != nil {
				ug.Draw(shape)
			}
		}
	}
}

func (u *UGraphicInterceptorAllSwimlanes) withActiveSwimlanes(tileSwimlanes SwimlaneSet) *UGraphicInterceptorAllSwimlanes {
	containsAll := true
	for _, s := range u.activeSwimlanes {
		if !tileSwimlanes.Contains(s) {
			containsAll = false
			break
		}
	}
	if containsAll {
		return u
	}
	var narrowed []*Swimlane
	for _, s := range u.activeSwimlanes {
		if tileSwimlanes.Contains(s) {
			narrowed = append(narrowed, s)
		}
	}
	return &UGraphicInterceptorAllSwimlanes{swimlaneToUg: u.swimlaneToUg, orderedList: u.orderedList,
		activeSwimlanes: narrowed, stringBounder: u.stringBounder}
}

// Apply implements klimt.UGraphic.
func (u *UGraphicInterceptorAllSwimlanes) Apply(change klimt.UChange) klimt.UGraphic {
	newMap := newSwimlaneUgMap()
	for _, k := range u.swimlaneToUg.keys {
		newMap.put(k, u.swimlaneToUg.ugs[k].Apply(change))
	}
	return &UGraphicInterceptorAllSwimlanes{swimlaneToUg: newMap, orderedList: u.orderedList,
		activeSwimlanes: u.activeSwimlanes, stringBounder: u.stringBounder}
}

func (u *UGraphicInterceptorAllSwimlanes) first() klimt.UGraphic {
	return u.swimlaneToUg.ugs[u.swimlaneToUg.keys[0]]
}

// StringBounder implements klimt.UGraphic.
func (u *UGraphicInterceptorAllSwimlanes) StringBounder() font.StringBounder { return u.stringBounder }

// Param implements klimt.UGraphic.
func (u *UGraphicInterceptorAllSwimlanes) Param() klimt.UParam { return u.first().Param() }

// ColorMapper implements klimt.UGraphic.
func (u *UGraphicInterceptorAllSwimlanes) ColorMapper() *color.ColorMapper {
	return u.first().ColorMapper()
}

// StartURL implements klimt.UGraphic: nothing.
func (u *UGraphicInterceptorAllSwimlanes) StartURL(*klimt.URL) {}

// CloseURL implements klimt.UGraphic: nothing.
func (u *UGraphicInterceptorAllSwimlanes) CloseURL() {}

// StartGroup implements klimt.UGraphic: nothing.
func (u *UGraphicInterceptorAllSwimlanes) StartGroup(*klimt.UGroup) {}

// CloseGroup implements klimt.UGraphic: nothing.
func (u *UGraphicInterceptorAllSwimlanes) CloseGroup() {}

// FlushUg implements klimt.UGraphic.
func (u *UGraphicInterceptorAllSwimlanes) FlushUg() {
	for _, k := range u.swimlaneToUg.keys {
		u.swimlaneToUg.ugs[k].FlushUg()
	}
}

// MatchesProperty implements klimt.UGraphic: false.
func (u *UGraphicInterceptorAllSwimlanes) MatchesProperty(string) bool { return false }

// DefaultBackground implements klimt.UGraphic.
func (u *UGraphicInterceptorAllSwimlanes) DefaultBackground() color.HColor {
	return u.first().DefaultBackground()
}

// WriteToStream implements klimt.UGraphic: it panics, as upstream.
func (u *UGraphicInterceptorAllSwimlanes) WriteToStream(io.Writer, string, int) error {
	uml.Throw(uml.UnsupportedOperationException, "")
	return nil
}
