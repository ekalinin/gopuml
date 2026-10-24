// SPDX-License-Identifier: MIT
// Provenance: original

package klimt

import (
	"sync"
	"testing"

	"github.com/ekalinin/gopuml/internal/klimt/color"
)

// The standard library keeps its SVG sprites for the process, so one parser
// draws for concurrent renders; run with -race.
func TestSvgSpriteParserConcurrent(t *testing.T) {
	svg := `<svg viewBox="0 0 10 10"><path d="M0 0 L5 5" fill="#336699"/><circle cx="5" cy="5" r="2" stroke="red"/></svg>`
	for _, parser := range []ISvgSpriteParser{NewSvgNanoParser(svg), NewSvgSaxParser(svg)} {
		forced := color.Instance().GetColorOrWhite("#FF8000")
		var want []string
		parser.DrawU(newSpriteRec(&want, false, color.IdentityMapper), 2, nil, forced)
		var wg sync.WaitGroup
		for range 8 {
			wg.Go(func() {
				var got []string
				parser.DrawU(newSpriteRec(&got, false, color.IdentityMapper), 2, nil, forced)
				if len(got) != len(want) {
					t.Errorf("%T: %d shapes, want %d", parser, len(got), len(want))
				}
			})
		}
		wg.Wait()
	}
}

// A document that is not well-formed draws nothing with the SAX parser: the
// first pass fails before the second draws, and DrawU panics with upstream's
// RuntimeException. The Nano parser draws the elements it finds.
func TestSvgSaxParserMalformed(t *testing.T) {
	svg := `<svg viewBox="0 0 10 10"><circle cx="5" cy="5" r="5" fill="red"></svg>`
	var log []string
	tb := NewSvgSaxParser(svg).AsTextBlock(nil, nil, 1, nil)
	if got := spriteTry(func() { tb.DrawU(newSpriteRec(&log, false, color.IdentityMapper)) }); got !=
		"ERR RuntimeException: Failed to parse SVG" {
		t.Errorf("SAX parser: %q, want the RuntimeException", got)
	}
	if len(log) != 0 {
		t.Errorf("SAX parser drew %q, want nothing", log)
	}
	NewSvgNanoParser(svg).AsTextBlock(nil, nil, 1, nil).DrawU(newSpriteRec(&log, false, color.IdentityMapper))
	if len(log) != 1 {
		t.Errorf("Nano parser drew %q, want the circle", log)
	}
}
