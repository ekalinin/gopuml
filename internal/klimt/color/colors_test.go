// SPDX-License-Identifier: MIT
// Provenance: original

package color

import (
	"errors"
	"testing"

	"github.com/ekalinin/gopuml/internal/command/regex"
)

// The expected values below were checked against upstream's classes run on
// the reference JVM.

func TestNewColors(t *testing.T) {
	tests := []struct {
		data, want string
		empty      bool
		shadow     string
	}{
		{"#red", "{BACK=[r=255,g=0,b=0,a=255] α=255} null", false, "null"},
		{"#red;line:blue", "{LINE=[r=0,g=0,b=255,a=255] α=255, BACK=[r=255,g=0,b=0,a=255] α=255} null", false, "null"},
		{"#back:red;line:blue;text:green",
			"{TEXT=[r=0,g=128,b=0,a=255] α=255, LINE=[r=0,g=0,b=255,a=255] α=255, BACK=[r=255,g=0,b=0,a=255] α=255} null",
			false, "null"},
		{"#line.dashed;header:yellow", "{HEADER=[r=255,g=255,b=0,a=255] α=255} DASHED(null)", false, "null"},
		{"#shadowing:false", "{} null", true, "false"},
		{"#shadowing:TRUE", "{} null", true, "true"},
		{"#text:nocolor", "{TEXT=null} null", false, "null"}, // present with a nil color
		{"#red;line.bold", "{BACK=[r=255,g=0,b=0,a=255] α=255} BOLD(null)", false, "null"},
		{"#line.dotted:red", "{LINE=[r=255,g=0,b=0,a=255] α=255} DOTTED(null)", false, "null"},
		{"#LINE.DASHED", "{} DASHED(null)", true, "null"},
		{"#arrow:red", "{ARROW=[r=255,g=0,b=0,a=255] α=255} null", false, "null"},
		{"#red;;blue", "{BACK=[r=0,g=0,b=255,a=255] α=255} null", false, "null"},
		{"#line.bold;line.dashed", "{} DASHED(null)", true, "null"}, // dashed is checked first
		{"#back:red-blue", "{BACK=net.sourceforge.plantuml.klimt.color.HColorGradient} null", false, "null"},
		{"#text:#red", "{TEXT=[r=255,g=0,b=0,a=255] α=255} null", false, "null"},
		{"#transparent", "{BACK=[r=0,g=0,b=0,a=0] α=0 transparent} null", false, "null"},
		{"#back:red;shadowing:true;line:blue",
			"{LINE=[r=0,g=0,b=255,a=255] α=255, BACK=[r=255,g=0,b=0,a=255] α=255} null", false, "true"},
	}
	for _, tt := range tests {
		c, err := NewColors(tt.data, Instance(), TypeBack)
		if err != nil {
			t.Errorf("%q: %v", tt.data, err)
			continue
		}
		if got := c.String(); got != tt.want {
			t.Errorf("%q: String = %q, want %q", tt.data, got, tt.want)
		}
		if c.IsEmpty() != tt.empty {
			t.Errorf("%q: IsEmpty = %v", tt.data, c.IsEmpty())
		}
		shadow := "null"
		if v, ok := c.Shadowing(); ok {
			shadow = map[bool]string{true: "true", false: "false"}[v]
		}
		if shadow != tt.shadow {
			t.Errorf("%q: shadowing = %s, want %s", tt.data, shadow, tt.shadow)
		}
	}
}

func TestNewColorsErrors(t *testing.T) {
	// A plain part must be a color; "shadowing" alone is parsed as one.
	for _, data := range []string{"#nocolor", "#shadowing", "#wh\u0130te"} {
		_, err := NewColors(data, Instance(), TypeBack)
		if _, ok := errors.AsType[*NoSuchColorError](err); !ok {
			t.Errorf("%q: err = %v", data, err)
		}
	}
	assertPanics(t, "java.lang.IllegalArgumentException: No enum constant net.sourceforge.plantuml.klimt.color.ColorType.FOO",
		func() { _, _ = NewColors("#foo:red", Instance(), TypeBack) })
	assertPanics(t, "java.lang.IllegalArgumentException: No enum constant net.sourceforge.plantuml.klimt.color.ColorType.",
		func() { _, _ = NewColors("#:red", Instance(), TypeBack) })
}

func TestColorsOperations(t *testing.T) {
	c, _ := NewColors("#red;line.dashed;shadowing:true", Instance(), TypeBack)
	if c.Color(TypeBack) == nil || c.Color(TypeLine) != nil || c.FirstColor(TypeLine, TypeBack) != c.Color(TypeBack) {
		t.Error("Color / FirstColor")
	}
	added := c.Add(TypeText, Blue)
	if added.Color(TypeText) != Blue || c.Color(TypeText) != nil {
		t.Error("Add must return a new value")
	}
	// The copy keeps the line style but not the shadowing, as upstream.
	if added.LineStyle() != LineStyleDashed {
		t.Errorf("LineStyle = %v", added.LineStyle())
	}
	if _, ok := added.Shadowing(); ok {
		t.Error("shadowing must not be copied")
	}
	if c.Add(TypeText, nil) != c || c.MergeWith(nil) != c {
		t.Error("nil arguments return the receiver")
	}
	other, _ := NewColors("#line:green;line.bold", Instance(), TypeBack)
	merged := c.MergeWith(other)
	if merged.Color(TypeLine).AsString() != "#008000" || merged.Color(TypeBack).AsString() != "#FF0000" ||
		merged.LineStyle() != LineStyleBold {
		t.Errorf("MergeWith = %v", merged)
	}
	for s, want := range map[string]LineStyle{"dashed": LineStyleDashed, "DOTTED": LineStyleDotted,
		"bold": LineStyleBold, "hidden": LineStyleInvisible, "plain": LineStyleNormal} {
		if got := EmptyColors().AddLegacyStroke(s).LineStyle(); got != want {
			t.Errorf("AddLegacyStroke(%q) = %v", s, got)
		}
	}
}

func TestColorTypeAndOrder(t *testing.T) {
	if ColorTypeOf("line.dashed") != TypeLine || ColorTypeOf("Back") != TypeBack {
		t.Error("ColorTypeOf")
	}
	if o, ok := ColorOrderFromString("gbr"); !ok || o != OrderGBR {
		t.Errorf("ColorOrderFromString(gbr) = %v %v", o, ok)
	}
	if _, ok := ColorOrderFromString("xyz"); ok {
		t.Error("xyz is not an order")
	}
	if got := OrderBRG.Color(NewXColorAlpha(1, 2, 3, 4)); got != NewXColor(3, 1, 2) {
		t.Errorf("BRG = %v", got)
	}
}

// colorCommand stands for a command pattern ending with an optional
// ":label", as upstream's timing highlight command.
func colorCommand(p *ColorParser) *regex.Concat {
	return regex.NewConcat(regex.Start(), p.Regex(), regex.SpaceZeroOrMore(),
		regex.Optional(regex.NewConcat(regex.NewLeaf(":"), regex.SpaceZeroOrMore(),
			regex.NewNamedLeaf(1, "REST", "(.*)"))), regex.End())
}

func TestColorParserRegex(t *testing.T) {
	tests := []struct {
		line, color, rest string // "-" for no match, "" color for none
	}{
		{"", "", ""},
		{"#red", "#red", ""},
		{"#Red-Blue", "#Red-Blue", ""},
		{`#red\blue`, `#red\blue`, ""},
		{"#red/blue", "#red/blue", ""},
		{"#a", "-", ""}, // a single color needs two word characters
		{"#ab", "#ab", ""},
		{"#back:red;line:blue", "#back:red;line:blue", ""},
		{"#red;line.dashed", "#red;line.dashed", ""},
		{"#line.dotted;text:red", "#line.dotted;text:red", ""},
		{"#LINE.DASHED;TEXT:RED", "#LINE.DASHED;TEXT:RED", ""},
		{"#line.dashed:red", "#line.dashed:red", ""},
		{"#shadowing:false", "#shadowing:false", ""},
		{"#text:red;", "#text:red;", ""},
		{"#text:red;:x", "#text:red;", "x"},
		{"#red;", "-", ""},
		{"#back;", "#back;", ""},
		{"#back:", "#back", ""},
		{"#back:red :caption", "#back:red", "caption"},
		{"#red:caption", "#red", "caption"},
		{"#line.dash", "-", ""},
		{"#red-", "-", ""},
		// Upstream's lookahead refuses ':' right after "#back:red" and falls
		// back to the single color "#back"; RE2 cannot express it, so the Go
		// pattern keeps "#back:red". This is the documented difference.
		{"#back:red:caption", "#back:red", "caption"},
	}
	cmd := colorCommand(SimpleColor(TypeBack))
	for _, tt := range tests {
		r := cmd.Matcher(tt.line)
		if tt.color == "-" {
			if r != nil {
				t.Errorf("%q must not match", tt.line)
			}
			continue
		}
		if r == nil {
			t.Errorf("%q must match", tt.line)
			continue
		}
		if got := r.Str("COLOR", 0); got != tt.color {
			t.Errorf("%q: COLOR = %q, want %q", tt.line, got, tt.color)
		}
		if got := r.Str("REST", 0); got != tt.rest {
			t.Errorf("%q: REST = %q, want %q", tt.line, got, tt.rest)
		}
	}

	// "line.dashed" is found whatever follows: upstream's lookahead makes
	// "line" fail before '.'.
	greedy := regex.NewConcat(regex.Start(), SimpleColor(TypeBack).Regex(), regex.NewNamedLeaf(1, "REST", "(.*)"),
		regex.End())
	if r := greedy.Matcher("#line.dashed:red"); r == nil || r.Str("COLOR", 0) != "#line.dashed:red" {
		t.Errorf("greedy context: %v", r)
	}
}

func TestColorParserColor(t *testing.T) {
	p := SimpleColor(TypeBack)
	cmd := colorCommand(p)
	colors, err := p.Color(cmd.Matcher("#back:red;line:blue"), Instance())
	if err != nil || colors.Color(TypeBack).AsString() != "#FF0000" || colors.Color(TypeLine).AsString() != "#0000FF" {
		t.Errorf("Color = %v, %v", colors, err)
	}
	colors, err = p.Color(cmd.Matcher(""), Instance())
	if err != nil || !colors.IsEmpty() {
		t.Errorf("no color: %v, %v", colors, err)
	}
	_, err = p.Color(cmd.Matcher("#nocolor"), Instance())
	if _, ok := errors.AsType[*NoSuchColorError](err); !ok {
		t.Errorf("invalid color: %v", err)
	}

	prefixed := regex.NewConcat(regex.Start(), SimpleColorPrefix("as", TypeText).Regex(), regex.End())
	if r := prefixed.Matcher("as #red;line:blue"); r == nil || r.Str("COLOR", 0) != "#red;line:blue" {
		t.Errorf("prefix: %v", r)
	}
	mandatory := regex.NewConcat(regex.Start(), MandatoryColor(TypeBack).Regex(), regex.End())
	if mandatory.Match("") || !mandatory.Match("#red") {
		t.Error("MandatoryColor")
	}
}

func TestColorParserLegacyLeaves(t *testing.T) {
	rest := func() regex.Part { return regex.NewNamedLeaf(1, "REST", "(.*)") }
	tests := []struct {
		cmd           *regex.Concat
		line          string
		name, color   string
		wantRest      string
		wantRestFound bool
	}{
		{regex.NewConcat(regex.Start(), Exp4(), rest(), regex.End()), "#red:label", "COLOR", "#red", "label", true},
		{regex.NewConcat(regex.Start(), Exp4(), rest(), regex.End()), "label", "COLOR", "", "label", true},
		{regex.NewConcat(regex.Start(), Exp6(), rest(), regex.End()), "#red|label", "COLOR", "#red", "label", true},
		{regex.NewConcat(regex.Start(), Exp7(), rest(), regex.End()), "#back:red", "COLOR", "#back", ":red", true},
		{regex.NewConcat(regex.Start(), Exp2(), Exp3(), regex.End()), "#red#blue", "BACKCOLOR", "#red", "", false},
		{regex.NewConcat(regex.Start(), Exp1(), rest(), regex.End()), "#red;line.bold x", "COLOR", "#red;line.bold", " x", true},
	}
	for _, tt := range tests {
		r := tt.cmd.Matcher(tt.line)
		if r == nil {
			t.Errorf("%q must match", tt.line)
			continue
		}
		if got := r.Str(tt.name, 0); got != tt.color {
			t.Errorf("%q: %s = %q, want %q", tt.line, tt.name, got, tt.color)
		}
		if got, ok := r.Get("REST", 0); got != tt.wantRest || ok != tt.wantRestFound {
			t.Errorf("%q: REST = %q %v", tt.line, got, ok)
		}
	}
	r := regex.NewConcat(regex.Start(), Exp2(), Exp3(), regex.End()).Matcher("#red#blue")
	if r.Str("BACKCOLOR2", 0) != "#blue" {
		t.Errorf("BACKCOLOR2 = %q", r.Str("BACKCOLOR2", 0))
	}
}

func TestUBrexColorParser(t *testing.T) {
	p := UBrexSimpleColor(TypeBack)
	// The legacy partition command: color, optional stereotype, spaces, "{".
	cmd := regex.NewConcat(regex.Start(), regex.NewLeaf("partition"), regex.SpaceOneOrMore(),
		regex.NewNamedLeaf(1, "NAME", `([^#\s]+)`), regex.SpaceZeroOrMore(), p.Regex(),
		regex.Optional(regex.NewNamedLeaf(1, "STEREOTYPE", "(<<.+>>)")), regex.SpaceZeroOrMore(),
		regex.Optional(regex.NewLeaf(`\{`)), regex.End())
	for line, want := range map[string]string{
		"partition P #red {":         "#red",
		"partition P #a":             "#a", // unlike ColorParser, one word character is enough
		"partition P #red-blue<<s>>": "#red-blue",
		"partition P #red/blue {":    "#red/blue",
	} {
		r := cmd.Matcher(line)
		if r == nil {
			t.Errorf("%q must match", line)
			continue
		}
		if got := r.Str("COLOR", 0); got != want {
			t.Errorf("%q: COLOR = %q, want %q", line, got, want)
		}
		colors, err := p.Color(r, Instance())
		if err != nil || colors.Color(TypeBack) == nil {
			t.Errorf("%q: Color = %v, %v", line, colors, err)
		}
	}
	if cmd.Match("partition P {") {
		t.Error("the UBrex color is mandatory")
	}
}
