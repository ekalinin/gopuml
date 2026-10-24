// SPDX-License-Identifier: MIT
// Provenance: original

package packetdiag

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/jfmt"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// rec is a UGraphic that records the paths, rectangles and texts drawn,
// with the translation they are drawn at.
type rec struct {
	klimt.UGraphic
	log    *[]string
	dx, dy float64
}

func newRec() *rec { return &rec{UGraphic: klimt.NewUGraphicNull(), log: new([]string)} }

func (r *rec) StringBounder() font.StringBounder { return font.NewStringBounderSvg(nil) }

func (r *rec) Apply(change klimt.UChange) klimt.UGraphic {
	c := *r
	if t, ok := change.(klimt.UTranslate); ok {
		c.dx += t.Dx()
		c.dy += t.Dy()
	}
	return &c
}

func d2s(d float64) string { return jfmt.DoubleToString(d) }

func (r *rec) Draw(shape klimt.UShape) {
	at := d2s(r.dx) + "," + d2s(r.dy) + " "
	switch s := shape.(type) {
	case *klimt.UPath:
		var segs []string
		for _, seg := range s.Segments() {
			segs = append(segs, seg.String())
		}
		*r.log = append(*r.log, at+"path "+strings.Join(segs, " "))
	case *klimt.URectangle:
		*r.log = append(*r.log, at+"rect "+d2s(s.Width())+"x"+d2s(s.Height()))
	case *klimt.UText:
		*r.log = append(*r.log, at+"text "+s.Text())
	}
}
