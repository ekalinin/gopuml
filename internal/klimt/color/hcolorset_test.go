// SPDX-License-Identifier: MIT
// Provenance: original

package color

import (
	"errors"
	"slices"
	"testing"
)

// The expected values below were checked against upstream's classes run on
// the reference JVM.

func TestParseSimpleColors(t *testing.T) {
	tests := []struct {
		in, asString, svg string
	}{
		// Named colors ignore the case of ASCII letters; '#' is optional.
		{"LightBlue", "#ADD8E6", "#ADD8E6"},
		{"lightblue", "#ADD8E6", "#ADD8E6"},
		{"LIGHTBLUE", "#ADD8E6", "#ADD8E6"},
		{"#LightBlue", "#ADD8E6", "#ADD8E6"},
		{"red", "#FF0000", "#FF0000"},
		{"##red", "#FF0000", "#FF0000"}, // parseColor and parseSimpleColor each strip one '#'
		{"Darkorange", "#FF8C00", "#FF8C00"},
		{"DarkOrange", "#FF8C00", "#FF8C00"},
		{"orange", "#FFA500", "#FFA500"}, // six letters, not hex: a name
		{"Tan", "#D2B48C", "#D2B48C"},    // three letters, not hex: a name
		{"business", "#FFFFCC", "#FFFFCC"},
		{"APPLICATION", "#C2F0FF", "#C2F0FF"},
		// Hex forms: 1 digit is a gray, 3 digits #RGB, 6 digits, 8 digits with alpha last.
		{"8", "#888888", "#888888"},
		{"#F", "#FFFFFF", "#FFFFFF"},
		{"abc", "#AABBCC", "#AABBCC"},
		{"#123456", "#123456", "#123456"},
		{"#12345678", "#78123456", "#12345678"},
		// asString puts the alpha first in lower case, SVG puts it last.
		{"#FF000080", "#80ff0000", "#FF000080"},
		// Alpha 0 is transparent whatever the channels.
		{"#80FF0000", "transparent", "#00000000"},
	}
	set := Instance()
	for _, tt := range tests {
		c := set.GetColorOrNull(tt.in)
		if _, ok := c.(*HColorSimple); !ok {
			t.Errorf("%q: got %T, want *HColorSimple", tt.in, c)
			continue
		}
		if got := c.AsString(); got != tt.asString {
			t.Errorf("%q: AsString = %q, want %q", tt.in, got, tt.asString)
		}
		if got := c.ToSvg(IdentityMapper); got != tt.svg {
			t.Errorf("%q: ToSvg = %q, want %q", tt.in, got, tt.svg)
		}
	}
}

func TestParseInvalidColors(t *testing.T) {
	set := Instance()
	for _, in := range []string{
		"", "#", "?", "z", "#12", "1234", "#12345", "#1234567", "#gggggg", "Light Blue", "light_blue", "LightBlu",
		"LightBluee", "lightblue1", "red-blue-green", "red--blue", "-red", "red-", "red-nocolor",
		"?red::", "?:", "?red", "?red:blue:nocolor", "?red:blue:green:yellow",
		"WH\u0130TE", // U+0130 is not an ASCII letter for the trie
		"réd",
	} {
		if c := set.GetColorOrNull(in); c != nil {
			t.Errorf("%q: got %v, want nil", in, c)
		}
		_, err := set.GetColor(in)
		if e, ok := errors.AsType[*NoSuchColorError](err); !ok || e.Error() != "NoSuchColorException" {
			t.Errorf("%q: GetColor error = %v", in, err)
		}
		if c := set.GetColorOrWhite(in); c != White {
			t.Errorf("%q: GetColorOrWhite = %v, want White", in, c)
		}
	}
}

func TestParseTransparentAndAutomatic(t *testing.T) {
	set := Instance()
	// Java's equalsIgnoreCase also matches U+017F (long s), the Kelvin sign
	// and the dotless and dotted i with their ASCII letters.
	for _, in := range []string{"transparent", "#TRANSPARENT", "background", "#background", "tran\u017Fparent",
		"bac\u212Aground"} {
		if c := set.GetColorOrNull(in); c != Transparent() {
			t.Errorf("%q: got %v, want the transparent color", in, c)
		}
	}
	for _, in := range []string{"automatic", "AUTOMATIC", "#automatic", "automat\u0131c", "automat\u0130c"} {
		c := set.GetColorOrNull(in)
		a, ok := c.(*HColorAutomagic)
		if !ok {
			t.Fatalf("%q: got %T", in, c)
		}
		if got := a.AsString(); got != "?HColorAutomagic" {
			t.Errorf("AsString = %q", got)
		}
		if got := a.AppropriateColor(White).AsString(); got != "#000000" {
			t.Errorf("on white: %s", got)
		}
		if got := a.AppropriateColor(Black).AsString(); got != "#FFFFFF" {
			t.Errorf("on black: %s", got)
		}
	}
	a1, a2 := set.GetColorOrNull("automatic"), set.GetColorOrNull("automatic")
	if a1.Equals(a2) || !a1.Equals(a1) {
		t.Error("automatic colors are only equal to themselves")
	}
	assertPanics(t, "java.lang.UnsupportedOperationException: class net.sourceforge.plantuml.klimt.color.HColorAutomagic",
		func() { a1.ToSvg(IdentityMapper) })
}

func TestParseGradients(t *testing.T) {
	tests := []struct {
		in       string
		c1, c2   string
		policy   rune
		svgColor string
	}{
		{"red-blue", "#FF0000", "#0000FF", '-', "#FF0000"},
		{"#red-blue", "#FF0000", "#0000FF", '-', "#FF0000"},
		{"red-#blue", "#FF0000", "#0000FF", '-', "#FF0000"},
		{`#red\blue`, "#FF0000", "#0000FF", '\\', "#FF0000"},
		{"red|blue", "#FF0000", "#0000FF", '|', "#FF0000"},
		{"red/blue", "#FF0000", "#0000FF", '/', "#FF0000"},
		{"#f00-#00f", "#FF0000", "#0000FF", '-', "#FF0000"},
		{"1|2", "#111111", "#222222", '|', "#111111"},
	}
	for _, tt := range tests {
		c := Instance().GetColorOrNull(tt.in)
		g, ok := c.(*HColorGradient)
		if !ok {
			t.Errorf("%q: got %T", tt.in, c)
			continue
		}
		if g.Color1().AsString() != tt.c1 || g.Color2().AsString() != tt.c2 || g.Policy() != tt.policy {
			t.Errorf("%q: got %s %s %q", tt.in, g.Color1().AsString(), g.Color2().AsString(), g.Policy())
		}
		if got := g.ToSvg(IdentityMapper); got != tt.svgColor {
			t.Errorf("%q: ToSvg = %s", tt.in, got)
		}
		if got := g.AsString(); got != "?HColorGradient" {
			t.Errorf("%q: AsString = %s", tt.in, got)
		}
	}
}

func TestGradientColor(t *testing.T) {
	// Gradients of gradients keep the first color of the first one and the
	// second color of the second one.
	g := Gradient(Instance().GetColorOrNull("red-blue"), Instance().GetColorOrNull("green-yellow"), '|')
	if g.Color1().AsString() != "#FF0000" || g.Color2().AsString() != "#FFFF00" {
		t.Fatalf("got %s %s", g.Color1().AsString(), g.Color2().AsString())
	}
	// Green moves by (int)(0.5 * 255) = 127.
	if got := g.Color(IdentityMapper, 0.5, 128); got != NewXColorAlpha(255, 127, 0, 128) {
		t.Errorf("Color(0.5) = %v", got)
	}
	if got := g.RGB(IdentityMapper, 1, 255); got != NewXColor(255, 255, 0).RGB() {
		t.Errorf("RGB(1) = %d", got)
	}
	assertPanics(t, "java.lang.IllegalArgumentException: c=1.5", func() { g.Color(IdentityMapper, 1.5, 255) })
	assertPanics(t, "java.lang.IllegalArgumentException: c=-1.0E-4", func() { g.Color(IdentityMapper, -0.0001, 255) })
	if NoGradient(g) != g.Color1() || NoGradient(Red) != Red {
		t.Error("NoGradient")
	}
}

func TestParseSchemes(t *testing.T) {
	tests := []struct {
		in                        string
		onWhite, onBlack, onTrans string
	}{
		{"?red:blue", "#FF0000", "#0000FF", "#FF0000"},
		{"#?red:blue", "#FF0000", "#0000FF", "#FF0000"},
		{"?red:blue:", "#FF0000", "#0000FF", "#FF0000"}, // trailing empty strings are dropped by split
		{"?#red:#00f", "#FF0000", "#0000FF", "#FF0000"},
		{"?red:blue:green", "#FF0000", "#0000FF", "#008000"},
		{"?a:b:c", "#AAAAAA", "#BBBBBB", "#CCCCCC"},
	}
	for _, tt := range tests {
		c := Instance().GetColorOrNull(tt.in)
		if _, ok := c.(*HColorScheme); !ok {
			t.Errorf("%q: got %T", tt.in, c)
			continue
		}
		got := [3]string{c.AppropriateColor(White).AsString(), c.AppropriateColor(Black).AsString(),
			c.AppropriateColor(Transparent()).AsString()}
		if want := [3]string{tt.onWhite, tt.onBlack, tt.onTrans}; got != want {
			t.Errorf("%q: got %v, want %v", tt.in, got, want)
		}
	}
	// On a transparent background without a third color, the light color
	// carries the dark one for dark mode.
	c := Instance().GetColorOrNull("?red:blue").AppropriateColor(Transparent())
	if got := c.ToSvg(DarkModeMapper); got != "#0000FF" {
		t.Errorf("dark mode = %s", got)
	}
	assertPanics(t, "java.lang.UnsupportedOperationException: class net.sourceforge.plantuml.klimt.color.HColorScheme",
		func() { Instance().GetColorOrNull("?red:blue").ToRGB(IdentityMapper) })
}

func TestParseMalformedSchemes(t *testing.T) {
	// Upstream reads a third color that is not there.
	for _, in := range []string{"?red:nocolor", "?nocolor:red", "?red:blue-green"} {
		assertPanics(t, "java.lang.ArrayIndexOutOfBoundsException: Index 2 out of bounds for length 2",
			func() { Instance().GetColorOrNull(in) })
	}
	// With three colors, an invalid first or second one gives a scheme whose
	// color panics only when used.
	c := Instance().GetColorOrNull("?red:nocolor:blue")
	if got := c.AppropriateColor(White).AsString(); got != "#FF0000" {
		t.Errorf("on white = %s", got)
	}
	defer func() {
		if recover() == nil {
			t.Error("using the null color must panic")
		}
	}()
	c.AppropriateColor(Black).AsString()
}

func TestNames(t *testing.T) {
	names := Instance().Names()
	if len(names) != 154 {
		t.Fatalf("got %d names", len(names))
	}
	// Java's String order: upper case before lower case.
	if want := []string{"APPLICATION", "AliceBlue", "AntiqueWhite", "Aqua", "Aquamarine", "Azure", "BUSINESS",
		"Beige"}; !slices.Equal(names[:8], want) {
		t.Errorf("first names = %v", names[:8])
	}
	for _, name := range names {
		if Instance().GetColorOrNull(name) == nil {
			t.Errorf("%s does not parse", name)
		}
	}
}

func TestColorTrie(t *testing.T) {
	if c, ok := ColorTrie.Color("lIgHtBlUe"); !ok || c != XColorFrom(0xADD8E6) {
		t.Errorf("got %v %v", c, ok)
	}
	for _, name := range []string{"", "Light", "Light1", "Light Blue", "LightBlueé"} {
		if _, ok := ColorTrie.Color(name); ok {
			t.Errorf("%q must have no color", name)
		}
	}
	trie := &ColorTrieNode{}
	trie.PutColor("My", XColorRed)
	trie.PutColor("bad1", XColorBlue) // ignored: not only letters
	if c, ok := trie.Color("MY"); !ok || c != XColorRed {
		t.Errorf("MY = %v %v", c, ok)
	}
	if _, ok := trie.Color("bad"); ok {
		t.Error("bad must have no color")
	}
}

// assertPanics checks that f panics with a *RuntimeError of the given text.
func assertPanics(t *testing.T, want string, f func()) {
	t.Helper()
	defer func() {
		t.Helper()
		r := recover()
		e, ok := r.(*RuntimeError)
		if !ok {
			t.Errorf("panic = %v, want %q", r, want)
			return
		}
		if e.Error() != want {
			t.Errorf("panic = %q, want %q", e.Error(), want)
		}
	}()
	f()
}
