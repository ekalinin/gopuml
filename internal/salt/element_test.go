// SPDX-License-Identifier: MIT
// Provenance: original

package salt

import (
	"fmt"
	"testing"

	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

var stringBounder = font.NewStringBounderSvg(nil)

func testDictionary() *saltDictionary { return newSaltDictionary(nil, nil) }

func TestSaltColors(t *testing.T) {
	tests := []struct {
		name        string
		c           color.HColor
		light, dark string
	}{
		{"black", saltBlack(), "#000000", "#FFFFFF"},
		{"white", saltWhite(), "#FFFFFF", "#000000"},
		{"88", saltColor88(), "#888888", "#888888"},
		{"AA", saltColorAA(), "#AAAAAA", "#666666"},
		{"BB", saltColorBB(), "#BBBBBB", "#555555"},
		{"DD", saltColorDD(), "#DDDDDD", "#333333"},
		{"EE", saltColorEE(), "#EEEEEE", "#222222"},
	}
	for _, tt := range tests {
		if got := tt.c.ToRGB(color.IdentityMapper); got != tt.light {
			t.Errorf("%s: %s, want %s", tt.name, got, tt.light)
		}
		if got := tt.c.DarkSchemeTheme().ToRGB(color.IdentityMapper); got != tt.dark {
			t.Errorf("%s in dark mode: %s, want %s", tt.name, got, tt.dark)
		}
	}
}

func TestCharNumber(t *testing.T) {
	tests := []struct {
		in   string
		want int
	}{
		{"Enter text here   ", 18},
		// The icon counts as "00".
		{"<&person> a", 4},
		{"<b>ab</b>", 2},
		// Java's length counts UTF-16 code units: the emoji counts as two.
		{"a\U0001F600", 3},
	}
	for _, tt := range tests {
		if got := charNumber(tt.in); got != tt.want {
			t.Errorf("charNumber(%q) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

// TestWidgetDimensions checks the preferred dimensions against the ones of
// upstream's elements, measured with the SVG string bounder of the
// reference image.
func TestWidgetDimensions(t *testing.T) {
	f := font.ByDefault(12)
	d := testDictionary()
	tests := []struct {
		name   string
		e      element
		want   string
		beyond bool
	}{
		{"button OK", newElementButton("OK", f, d), "[26.314453125,22.96875]", false},
		{"button '  OK   '", newElementButton("  OK   ", f, d), "[65,22.96875]", false},
		{"text field 'Enter text here   '", newElementTextField("Enter text here   ", f, d), "[150,15.96875]", false},
		{"text field ''", newElementTextField("", f, d), "[9.814453125,15.96875]", false},
		{"droplist 'This is a droplist'", newElementDroplist("This is a droplist", f, d), "[160,17.96875]", false},
		{"droplist 'head^^a^^b'", newElementDroplist("head^^a^^b", f, d), "[48,17.96875]", true},
		{"droplist 'head^ a^ b'", newElementDroplist("head^ a^ b", f, d), "[48,17.96875]", true},
		{"radio 'public'", newElementRadioCheckbox([]string{"public"}, f, true, true, d), "[56.10546875,13.96875]", false},
		{"text 'Just plain text'", newElementText([]string{"Just plain text"}, f, d), "[82.869140625,13.96875]", false},
		{"text '.'", newElementText([]string{"."}, f, d), "[3.814453125,13.96875]", false},
		{"separator", &elementLine{separator: '-'}, "[10,6]", false},
		{"image", &elementImage{img: []string{"XX.XX", ".....", "X...X", ".XXX."}}, "[5,4]", false},
	}
	for _, tt := range tests {
		dim := tt.e.preferredDimension(stringBounder, 0, 0)
		if got := fmt.Sprintf("[%v,%v]", dim.Width(), dim.Height()); got != tt.want {
			t.Errorf("%s: %s, want %s", tt.name, got, tt.want)
		}
		if got := tt.e.mayDrawBeyondPreferredDimension(); got != tt.beyond {
			t.Errorf("%s: mayDrawBeyondPreferredDimension %v, want %v", tt.name, got, tt.beyond)
		}
	}
}

// TestDroplistItems checks the text and the items of open droplists.
func TestDroplistItems(t *testing.T) {
	if got := droplistExtract("head^^a^^b"); got != "head" {
		t.Errorf("text %q, want head", got)
	}
	closed := newElementDroplist("Pick", font.ByDefault(12), testDictionary())
	if closed.openDrop != nil {
		t.Error("a droplist without items is open")
	}
}
