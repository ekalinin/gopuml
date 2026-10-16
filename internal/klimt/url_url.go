// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/url/Url.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/url/UrlMode.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/url/Check.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/url/CMapData.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"slices"
	"strings"

	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// URL is a link of a diagram element: a target, a tooltip and a label. It
// records the area it covers as it is drawn (EnsureVisible).
//
// Upstream's Check.isJunit, a process-wide flag that makes getCoords throw
// on empty areas in upstream's own tests, is not ported.
type URL struct {
	url     string
	tooltip string
	label   string
	member  bool
	visible *geom.BasicEnsureVisible
}

var _ geom.EnsureVisible = (*URL)(nil)

// NewURL returns a link. The target loses its enclosing double quotes; a
// missing tooltip (ok false) is the target, and a missing or empty label is
// the target.
func NewURL(url string, tooltip string, tooltipOK bool, label string, labelOK bool) *URL {
	url = eventuallyRemoveStartingAndEndingDoubleQuoteFormat(url, "\"")
	u := &URL{url: url, visible: geom.NewBasicEnsureVisible()}
	if !tooltipOK {
		u.tooltip = url
	} else {
		u.tooltip = tooltip
	}
	if !labelOK || label == "" {
		u.label = url
	} else {
		u.label = label
	}
	return u
}

// URLIsLatex reports whether a link target is a "latex://" link.
func URLIsLatex(pendingURL string) bool { return strings.HasPrefix(pendingURL, "latex://") }

// IsLatex reports whether the target is a "latex://" link.
func (u *URL) IsLatex() bool { return URLIsLatex(u.url) }

// URL returns the target.
func (u *URL) URL() string { return u.url }

// Tooltip returns the tooltip.
func (u *URL) Tooltip() string { return u.tooltip }

// Label returns the label.
func (u *URL) Label() string { return u.label }

// String is upstream's toString, without Java's identity hash.
func (u *URL) String() string { return "Url " + u.url + " " + u.visible.Coords(1.0) }

// Coords returns the covered area as "x1,y1,x2,y2" at a scale.
func (u *URL) Coords(scale float64) string { return u.visible.Coords(scale) }

// SetMember marks the link as the link of a class member.
func (u *URL) SetMember(member bool) { u.member = member }

// IsMember reports whether the link is the link of a class member.
func (u *URL) IsMember() bool { return u.member }

// EnsureVisible extends the covered area to (x, y).
func (u *URL) EnsureVisible(x, y float64) { u.visible.EnsureVisible(x, y) }

// HasData reports whether the link covers an area.
func (u *URL) HasData() bool { return u.visible.HasData() }

// URLSurfaceCompare is upstream's SURFACE_COMPARATOR: links with smaller
// areas first.
func URLSurfaceCompare(url1, url2 *URL) int {
	surface1 := url1.visible.Surface()
	surface2 := url2.visible.Surface()
	if surface1 > surface2 {
		return 1
	} else if surface1 < surface2 {
		return -1
	}
	return 0
}

// URLMode tells whether a link must be the whole text or may be anywhere in
// it.
type URLMode int

// URL modes, in upstream's order.
const (
	URLModeStrict URLMode = iota
	URLModeAnywhere
)

// CMapData is an HTML image map of the links of a PNG image.
type CMapData struct {
	sb strings.Builder
}

// AsString returns the <map> element.
func (c *CMapData) AsString(nameID string) string {
	return "<map id=\"" + nameID + "_map\" name=\"" + nameID + "_map\">\n" + c.sb.String() + "</map>\n"
}

// ContainsData reports whether the map has areas.
func (c *CMapData) ContainsData() bool { return c.sb.Len() > 0 }

// AppendString appends text.
func (c *CMapData) AppendString(s string) { c.sb.WriteString(s) }

// AppendLong appends a number.
func (c *CMapData) AppendLong(v int64) { c.sb.WriteString(itoa(int(v))) }

func (c *CMapData) appendURL(seq int, url *URL, scale float64) {
	c.AppendString("<area shape=\"rect\" id=\"id")
	c.AppendLong(int64(seq))
	c.AppendString("\" href=\"")
	c.AppendString(url.URL())
	c.AppendString("\" title=\"")
	tooltip := strings.ReplaceAll(url.Tooltip(), "\\n", backSlashNewline)
	tooltip = strings.ReplaceAll(tooltip, "&", "&#38;")
	tooltip = strings.ReplaceAll(tooltip, "\"", "&#34;")
	tooltip = strings.ReplaceAll(tooltip, "'", "&#39;")
	c.AppendString(tooltip)
	c.AppendString("\" alt=\"\" coords=\"")
	c.AppendString(url.Coords(scale))
	c.AppendString("\"/>")
	c.AppendString(backSlashNewline)
}

// CMapString returns the map of the links that cover an area, smaller areas
// first. Upstream takes a Set, whose iteration order decides between links
// of equal area; here the caller's order does.
func CMapString(allURLEncountered []*URL, scale float64) *CMapData {
	cmapdata := &CMapData{}
	all := slices.Clone(allURLEncountered)
	slices.SortStableFunc(all, URLSurfaceCompare)
	seq := 1
	for _, u := range all {
		if !u.HasData() {
			continue
		}
		cmapdata.appendURL(seq, u, scale)
		seq++
	}
	return cmapdata
}
