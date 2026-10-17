// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/cucadiagram/HideOrShow.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/cucadiagram/LinkConstraint.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/cucadiagram/PortionShower.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/cucadiagram/GroupHierarchy.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package abel

import (
	"strings"
	"time"

	"github.com/dlclark/regexp2"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// HideOrShow is one "hide", "show", "remove" or "restore" rule on
// elements: a name pattern ('*' matches anything), "<<stereotype>>",
// "$tag" or "@unlinked".
type HideOrShow struct {
	what string
	show bool
}

// NewHideOrShow returns a rule.
func NewHideOrShow(what string, show bool) *HideOrShow { return &HideOrShow{what: what, show: show} }

// String is upstream's toString, for example "Foo* (false)".
func (h *HideOrShow) String() string {
	if h.show {
		return h.what + " (true)"
	}
	return h.what + " (false)"
}

// What returns the pattern.
func (h *HideOrShow) What() string { return h.what }

// Show reports whether the rule shows (or restores) its elements.
func (h *HideOrShow) Show() bool { return h.show }

func (h *HideOrShow) isApplyable(leaf *Entity) bool {
	if leaf.IsRoot() {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	if tag, ok := strings.CutPrefix(h.what, "$"); ok {
		return isApplyableTag(leaf, tag)
	}
	if strings.HasPrefix(h.what, "<<") && strings.HasSuffix(h.what, ">>") {
		return isApplyableStereotype(leaf.Stereotype(), stereotypePattern(h.what))
	}
	if h.IsAboutUnlinked() {
		return leaf.IsAloneAndUnlinked()
	}
	return hideMatch(leaf.Quark().QualifiedName(), h.what)
}

// stereotypePattern is what.substring(2, what.length() - 2).trim(): for
// "<<>" the indexes cross and upstream throws
// StringIndexOutOfBoundsException.
func stereotypePattern(what string) string {
	if len(what) < 4 {
		uml.Throw(uml.StringIndexOutOfBoundsException, "begin 2, end "+itoa(len(what)-2)+", length "+
			itoa(len(what)))
	}
	return strings.TrimFunc(what[2:len(what)-2], func(r rune) bool { return r <= ' ' })
}

func (h *HideOrShow) isApplyableStereo(stereotype *klimt.Stereotype) bool {
	if strings.HasPrefix(h.what, "<<") && strings.HasSuffix(h.what, ">>") {
		return isApplyableStereotype(stereotype, stereotypePattern(h.what))
	}
	return false
}

// IsAboutUnlinked reports whether the rule is "@unlinked".
func (h *HideOrShow) IsAboutUnlinked() bool { return uml.EqualsIgnoreCase(h.what, "@unlinked") }

func isApplyableStereotype(stereotype *klimt.Stereotype, pattern string) bool {
	if stereotype == nil {
		return false
	}
	for _, label := range stereotype.MultipleLabels() {
		if hideMatch(label, pattern) {
			return true
		}
	}
	return false
}

func isApplyableTag(leaf *Entity, pattern string) bool {
	for _, tag := range leaf.Stereotags() {
		if hideMatch(tag.Name(), pattern) {
			return true
		}
	}
	return false
}

// hideMatchTimeout bounds the match of a user pattern (design D4).
const hideMatchTimeout = 2 * time.Second

// hideMatch is upstream's match: the name after the last magic separator
// equals the pattern, or matches it as the Java regular expression
// "^" + pattern + "$" with '*' turned into ".*" when it has a '*'. As other
// user regular expressions, the pattern runs on regexp2 (design D4); its
// "." excludes Java's line terminators. An invalid pattern panics, as
// upstream throws PatternSyntaxException.
func hideMatch(name, pattern string) bool {
	if idx := strings.LastIndex(name, MagicSeparator); idx != -1 {
		name = name[idx+1:]
	}
	if !strings.Contains(pattern, "*") {
		return name == pattern
	}
	reg := "^" + strings.ReplaceAll(pattern, "*", ".*") + "$"
	re, err := regexp2.Compile(javaDotsRegexp2(reg), regexp2.None)
	if err != nil {
		uml.Throw("java.util.regex.PatternSyntaxException", err.Error())
	}
	re.MatchTimeout = hideMatchTimeout
	ok, err := re.MatchString(name)
	if err != nil {
		uml.Throw("java.lang.RuntimeException", err.Error())
	}
	return ok
}

// javaDotsRegexp2 rewrites "." outside character classes so that it does
// not match Java's line terminators, and "$" so that it only matches at the
// end, as String.matches requires a full match.
func javaDotsRegexp2(pattern string) string {
	var sb strings.Builder
	inClass := false
	for i := 0; i < len(pattern); i++ {
		c := pattern[i]
		switch {
		case c == '\\' && i+1 < len(pattern):
			sb.WriteByte(c)
			i++
			sb.WriteByte(pattern[i])
			continue
		case c == '[' && !inClass:
			inClass = true
		case c == ']' && inClass:
			inClass = false
		case c == '.' && !inClass:
			sb.WriteString(`[^\n\r\u0085\u2028\u2029]`)
			continue
		case c == '$' && !inClass && i == len(pattern)-1:
			sb.WriteString(`\z`)
			continue
		}
		sb.WriteByte(c)
	}
	return sb.String()
}

// Apply returns the hidden (or removed) state of an entity after the rule.
func (h *HideOrShow) Apply(hidden bool, leaf *Entity) bool {
	if h.isApplyable(leaf) {
		return !h.show
	}
	return hidden
}

// ApplyStereotype returns the removed state of a stereotype after the
// rule.
func (h *HideOrShow) ApplyStereotype(hidden bool, stereotype *klimt.Stereotype) bool {
	if h.isApplyableStereo(stereotype) {
		return !h.show
	}
	return hidden
}

// LinkConstraint is a "constraint on links" label between the two last
// links. Its positions are set when the links are drawn.
type LinkConstraint struct {
	link1, link2   *Link
	display        *klimt.Display
	x1, y1, x2, y2 float64
}

// NewLinkConstraint returns a constraint.
func NewLinkConstraint(link1, link2 *Link, display *klimt.Display) *LinkConstraint {
	return &LinkConstraint{link1: link1, link2: link2, display: display}
}

// SetPosition records where a link of the constraint is drawn; another
// link panics, as upstream throws IllegalArgumentException.
func (c *LinkConstraint) SetPosition(link *Link, pt geom.XPoint2D) {
	switch link {
	case c.link1:
		c.x1, c.y1 = pt.X, pt.Y
	case c.link2:
		c.x2, c.y2 = pt.X, pt.Y
	default:
		uml.Throw(uml.IllegalArgumentException, "")
	}
}

// Links returns the two links.
func (c *LinkConstraint) Links() (*Link, *Link) { return c.link1, c.link2 }

// Display returns the label.
func (c *LinkConstraint) Display() *klimt.Display { return c.display }

// PortionShower tells which portions of an entity are shown.
type PortionShower interface {
	ShowPortion(portion uml.EntityPortion, entity *Entity) bool
	// VisibleStereotypeLabels returns the stereotype labels to draw, nil
	// when the entity has no stereotype.
	VisibleStereotypeLabels(entity *Entity) []string
}

type portionShowerAll struct{}

func (portionShowerAll) ShowPortion(uml.EntityPortion, *Entity) bool { return true }

func (portionShowerAll) VisibleStereotypeLabels(*Entity) []string { return []string{} }

// PortionShowerAll is upstream's PortionShower.ALL: everything is shown.
var PortionShowerAll PortionShower = portionShowerAll{}

// GroupHierarchy gives the groups of a diagram.
type GroupHierarchy interface {
	RootGroup() *Entity
	ChildrenGroups(parent *Entity) []*Entity
	IsEmpty(g *Entity) bool
}

func itoa(i int) string {
	if i < 0 {
		return "-" + itoa(-i)
	}
	if i < 10 {
		return string(rune('0' + i))
	}
	return itoa(i/10) + string(rune('0'+i%10))
}
