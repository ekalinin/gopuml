// SPDX-License-Identifier: MIT
// Provenance: original

package sequencediagram

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/ekalinin/plantuml-in-go/internal/core"
	"github.com/ekalinin/plantuml-in-go/internal/jfmt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/security"
	"github.com/ekalinin/plantuml-in-go/internal/text"
	"github.com/ekalinin/plantuml-in-go/internal/tim"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

// testCase is one source of the differential tests.
type testCase struct {
	name, source string
}

// readCases reads a case file: "=== name" lines start cases, as
// testdata/upstream/Harness.java reads them.
func readCases(path string) ([]testCase, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cases []testCase
	var cur *testCase
	var sb strings.Builder
	flush := func() {
		if cur != nil {
			cur.source = sb.String()
			cases = append(cases, *cur)
		}
	}
	for _, line := range strings.Split(strings.TrimSuffix(string(data), "\n"), "\n") {
		if name, ok := strings.CutPrefix(line, "=== "); ok {
			flush()
			cur = &testCase{name: name}
			sb.Reset()
			continue
		}
		if cur == nil {
			continue
		}
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	flush()
	return cases, nil
}

func unescape(s string) string {
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+5 < len(s) && s[i+1] == 'u' {
			if v, err := strconv.ParseUint(s[i+2:i+6], 16, 32); err == nil {
				sb.WriteRune(rune(v))
				i += 5
				continue
			}
		}
		sb.WriteByte(s[i])
	}
	return sb.String()
}

// writeCases writes cases in the format of readCases.
func writeCases(path string, cases []testCase) error {
	var sb strings.Builder
	for _, c := range cases {
		sb.WriteString("=== " + c.name + "\n")
		sb.WriteString(c.source)
		if !strings.HasSuffix(c.source, "\n") {
			sb.WriteString("\n")
		}
	}
	return os.WriteFile(path, []byte(sb.String()), 0o644)
}

// frontMatterBody removes the YAML front matter of a conformance or Vega
// case file.
func frontMatterBody(data string) string {
	data = strings.ReplaceAll(data, "\r\n", "\n")
	first, rest, ok := strings.Cut(data, "\n")
	if !ok || strings.TrimSpace(first) != "---" {
		return data
	}
	for {
		line, after, ok := strings.Cut(rest, "\n")
		if strings.TrimSpace(line) == "---" {
			return after
		}
		if !ok {
			return data
		}
		rest = after
	}
}

// corpusCases returns the @startuml sources of the .puml files below root,
// named by prefix and their relative path.
func corpusCases(root, prefix string) ([]testCase, error) {
	var cases []testCase
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(p) != ".puml" {
			return nil
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		body := frontMatterBody(string(data))
		if !strings.HasPrefix(strings.TrimSpace(body), "@startuml") {
			return nil
		}
		rel, _ := filepath.Rel(root, p)
		cases = append(cases, testCase{name: prefix + filepath.ToSlash(rel), source: body})
		return nil
	})
	slices.SortFunc(cases, func(a, b testCase) int { return strings.Compare(a.name, b.name) })
	return cases, err
}

// parseCase preprocesses the first block of a source and runs the sequence
// factory on it, as Harness.java does with upstream's.
func parseCase(source string) (result core.Diagram, preprocessorError bool, err error) {
	if !strings.HasPrefix(strings.TrimSpace(source), "@start") {
		source = "@startuml\n" + source + "@enduml\n"
	}
	rc := core.NewRenderContext()
	blocks := core.BuildBlocks(source, rc.Desc, rc.Config)
	if len(blocks) == 0 {
		return nil, false, fmt.Errorf("no block")
	}
	access := &security.Access{}
	access.Config.Profile, _ = security.ParseProfile(rc.SecurityProfile)
	res, perr := tim.PreprocessWith(rc, access, blocks[0].Lines,
		tim.Options{Definitions: &tim.BlockDefinitions{}, Defines: tim.NewDefines()})
	if res == nil {
		return nil, false, perr
	}
	if perr != nil {
		return nil, true, nil
	}
	src := core.NewUmlSource(res.Lines, blocks[0].Lines, true)
	return Factory.CreateSystem(rc, src), false, nil
}

var formatNumbers = []int64{0, 1, 7, 12, 15, 25, 99, 123, 1234, 99999, 123456789, -5, 2147483647}

// dumpFormats is Harness.dumpFormats.
func dumpFormats(out []string, source string) []string {
	for _, pattern := range strings.Split(source, "\n") {
		if pattern == "" {
			continue
		}
		line := "format [" + pattern + "]"
		df, err := newDecimalFormat(pattern)
		if err != nil {
			line += " error " + err.Error()
		} else {
			for _, n := range formatNumbers {
				line += " | " + df.format(n)
			}
		}
		out = append(out, line)
	}
	return out
}

// dumpCase is the text Harness.java writes for a case.
func dumpCase(c testCase) []string {
	out := []string{"=== " + c.name}
	if strings.Contains(c.name, "decimalformat/") {
		return dumpFormats(out, c.source)
	}
	source := c.source
	if strings.HasSuffix(c.name, " (escaped)") {
		source = unescape(source)
	}
	d, preproc, err := func() (d core.Diagram, preproc bool, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("crash %v", r)
			}
		}()
		return parseCase(source)
	}()
	switch {
	case err != nil && strings.HasPrefix(err.Error(), "crash"):
		return append(out, err.Error())
	case err != nil:
		return append(out, "no block")
	case preproc:
		return append(out, "preprocessor error")
	}
	if isCompact(c.name) {
		var compact []string
		for _, line := range dumpDiagram(out[:1:1], d)[1:] {
			if strings.HasPrefix(line, "result") || strings.HasPrefix(line, "error") || strings.HasPrefix(line, "event") {
				compact = append(compact, line)
			}
		}
		return append(out, compact...)
	}
	return dumpDiagram(out, d)
}

// isCompact is Harness.isCompact: only the result, the error and the events
// of the generated arrow cases are dumped.
func isCompact(name string) bool {
	return strings.HasPrefix(name, "arrowcombos/") || strings.HasPrefix(name, "exocombos/")
}

func dumpDiagram(out []string, d core.Diagram) []string {
	switch d := d.(type) {
	case nil:
		return append(out, "result null")
	case *core.ErrorDiagram:
		typ := "SYNTAX_ERROR"
		if d.Error.Type == core.ExecutionError {
			typ = "EXECUTION_ERROR"
		}
		last := 0
		if n := len(d.Trace); n > 0 && d.Trace[n-1].Location != nil {
			last = d.Trace[n-1].Location.Position
		}
		return append(out, "result error",
			"error "+typ+" "+strconv.Itoa(d.Error.Position())+" "+d.Error.Error,
			"error.message "+d.Error.Message(),
			"error.score "+strconv.Itoa(d.Score()),
			"error.last "+strconv.Itoa(last))
	case *SequenceDiagram:
		return dumpSequence(out, d)
	}
	return append(out, fmt.Sprintf("result %T", d))
}

func jstr(s string, ok bool) string {
	if !ok {
		return "null"
	}
	return s
}

func jcolor(c color.HColor) string {
	if c == nil {
		return "null"
	}
	return c.String()
}

func jbool(b bool) string { return strconv.FormatBool(b) }

func dumpSequence(out []string, sd *SequenceDiagram) []string {
	out = append(out, "result SequenceDiagram",
		"description "+sd.Description(),
		"pages "+strconv.Itoa(sd.Pages()),
		"title "+dumpDP(sd.Title()),
		"caption "+dumpDP(sd.Caption()),
		"legend "+dumpDP(sd.Legend()),
		"header "+dumpDP(sd.Header()),
		"footer "+dumpDP(sd.Footer()),
		"mainframe "+dumpDP(sd.MainFrame()))
	for i := 1; i < sd.Pages(); i++ {
		out = append(out, "pagetitle "+strconv.Itoa(i)+" "+dumpDP(sd.TitleOfPage(i)))
	}
	h, v := sd.SplitPages()
	out = append(out, "scale "+dumpScale(sd.Scale()),
		"splitpages "+strconv.Itoa(h)+" "+strconv.Itoa(v),
		"namespaceseparator "+jstr(sd.NamespaceSeparator()),
		"skinparamused "+jbool(sd.IsSkinParamUsed()),
		"smetana "+jbool(sd.IsUseSmetana())+" elk "+jbool(sd.IsUseElk()),
		"footbox "+jbool(sd.IsShowFootbox()),
		"hideunlinked "+jbool(sd.IsHideUnlinkedData()),
		"autoactivate "+jbool(sd.IsAutoactivate()),
		"hidden "+dumpPortions(sd.HiddenPortions()))
	an := sd.AutoNumber()
	current := "null"
	if an.current != nil {
		current = an.current.String()
	}
	out = append(out, "autonumber running="+jbool(an.running)+" current="+current+" increment="+
		strconv.Itoa(an.increment)+" last="+an.last)
	pragma := sd.Pragma()
	for k := klimt.PragmaKey(0); !strings.HasPrefix(k.String(), "PragmaKey("); k++ {
		if pragma.IsDefine(k) {
			out = append(out, "pragma "+k.String()+"="+jstr(pragma.Value(k)))
		}
	}
	for _, w := range pragma.Warnings() {
		out = append(out, "warning "+strings.Join(w.Message(), "\\n"))
	}
	values := sd.Skin().Values()
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, k := range keys {
		out = append(out, "skinparam "+k+"="+values[k])
	}
	for _, name := range sd.Skin().AllSpriteNames() {
		out = append(out, "sprite "+name+" "+dumpSprite(sd.Skin().Sprite(name)))
	}
	out = append(out, "defaultskin "+sd.Skin().DefaultSkin())
	for _, p := range sd.Participants() {
		out = append(out, "participant "+dumpParticipant(sd, p))
	}
	for i, e := range sd.Events() {
		out = append(out, "event "+strconv.Itoa(i)+" "+dumpEvent(sd.Events(), e))
	}
	for _, a := range sd.LinkAnchors() {
		out = append(out, "linkanchor "+a.Anchor1+" "+a.Anchor2+" "+jstr(a.Message, a.HasMessage))
	}
	return out
}

func dumpPortions(s uml.EntityPortions) string {
	var names []string
	for _, p := range s.List() {
		names = append(names, p.String())
	}
	return "[" + strings.Join(names, ", ") + "]"
}

func dumpSprite(s klimt.Sprite) string {
	def, ok := s.(*uml.SpriteDef)
	if !ok || def == nil {
		if s == nil {
			return "null"
		}
		return "svg"
	}
	switch def.Kind {
	case uml.SpriteKindMonochrome:
		var sb strings.Builder
		for _, v := range def.Gray {
			sb.WriteString(strconv.FormatUint(uint64(uint32(v)), 16))
		}
		return "SpriteMonochrome " + strconv.Itoa(def.Width) + "x" + strconv.Itoa(def.Height) + "/" +
			strconv.Itoa(def.GrayLevel) + " " + sb.String()
	case uml.SpriteKindColor:
		var sb strings.Builder
		for _, v := range def.Color {
			sb.WriteString(strconv.FormatUint(uint64(uint32(v)), 16))
			sb.WriteString(",")
		}
		return "SpriteColor " + strconv.Itoa(def.Width) + "x" + strconv.Itoa(def.Height) + " " + sb.String()
	case uml.SpriteKindImage:
		return "SpriteImage"
	}
	return "svg"
}

func dumpScale(s *uml.Scale) string {
	if s == nil {
		return "null"
	}
	f := jfmt.DoubleToString
	switch s.Kind {
	case uml.ScaleSimple:
		return "ScaleSimple scale=" + f(s.A)
	case uml.ScaleWidth, uml.ScaleMaxWidth:
		return s.Kind.String() + " maxWidth=" + f(s.A)
	case uml.ScaleHeight, uml.ScaleMaxHeight:
		return s.Kind.String() + " maxHeight=" + f(s.A)
	}
	return s.Kind.String() + " maxWidth=" + f(s.A) + " maxHeight=" + f(s.B)
}

func dumpDisplay(d *klimt.Display) string {
	if klimt.DisplayIsNull(d) {
		return "NULL"
	}
	var parts []string
	for _, cs := range d.AsList() {
		prefix := ""
		switch cs.(type) {
		case *klimt.Stereotype:
			prefix = "S:"
		case klimt.MessageNumber:
			prefix = "N:"
		}
		parts = append(parts, prefix+strings.ReplaceAll(cs.String(), "\n", "\\n"))
	}
	s := "[" + strings.Join(parts, "|") + "]"
	if a := d.NaturalHorizontalAlignment(); a != klimt.HorizontalAlignmentNull {
		s += "@" + a.String()
	}
	return s
}

func dumpDP(dp uml.DisplayPositioned) string {
	h, v := "null", "null"
	if dp.HorizontalAlignment() != klimt.HorizontalAlignmentNull {
		h = dp.HorizontalAlignment().String()
	}
	if dp.VerticalAlignment() != uml.VerticalAlignmentNull {
		v = dp.VerticalAlignment().String()
	}
	at := "null"
	if dp.LineLocation() != nil {
		at = strconv.Itoa(dp.LineLocation().Position)
	}
	return dumpDisplay(dp.Display()) + " h=" + h + " v=" + v + " at=" + at
}

func dumpURL(u *klimt.URL) string {
	if u == nil {
		return "null"
	}
	return "{" + u.URL() + "|" + u.Tooltip() + "|" + u.Label() + "}"
}

func dumpStereo(s *klimt.Stereotype) string {
	if s == nil {
		return "null"
	}
	return s.String()
}

func dumpFashion(f *klimt.Fashion) string {
	if f == nil {
		return "null"
	}
	return jcolor(f.BackColor()) + "/" + jcolor(f.ForeColor())
}

func dumpLocation(l *text.LineLocation) string {
	if l == nil {
		return "null"
	}
	return strconv.Itoa(l.Position)
}

func dumpParticipant(sd *SequenceDiagram, p *Participant) string {
	var sb strings.Builder
	sb.WriteString(p.Code() + " uid=" + p.UID() + " type=" + p.Type().String() + " order=" + strconv.Itoa(p.Order()))
	sb.WriteString(" display=" + dumpDisplay(p.Display(false)))
	sb.WriteString(" underlined=" + dumpDisplay(p.Display(true)))
	sb.WriteString(" stereo=" + dumpStereo(p.Stereotype()))
	sb.WriteString(" stereotop=" + jbool(p.StereotypePositionTop()))
	sb.WriteString(" colors=" + p.Colors().String())
	sb.WriteString(" url=" + dumpURL(p.URL()))
	sb.WriteString(" initiallife=" + strconv.Itoa(p.InitialLife()))
	sb.WriteString(" lives=[")
	for _, f := range p.liveBackcolors {
		sb.WriteString(dumpFashion(f) + ";")
	}
	sb.WriteString("]")
	sb.WriteString(" box=")
	if e := sd.Englober(p); e == nil {
		sb.WriteString("null")
	} else {
		for _, g := range e.Genealogy() {
			sb.WriteString("<" + dumpDisplay(g.Title()) + " " + jcolor(g.BoxColor()) + " " + dumpStereo(g.Stereotype()) + ">")
		}
	}
	sb.WriteString(" at=" + dumpLocation(p.Location()))
	return sb.String()
}

func dumpArrow(a ArrowConfiguration) string {
	return a.Name() + " parts=" + a.Dressing1().Part.String() + "/" + a.Dressing2().Part.String() +
		" thickness=" + jfmt.DoubleToString(a.Thickness()) + " reverse=" + jbool(a.IsReverseDefine()) +
		" inclination=" + strconv.Itoa(a.Inclination())
}

func dumpNote(n *Note) string {
	code := func(p *Participant) string {
		if p == nil {
			return "null"
		}
		return p.Code()
	}
	shadowing := "null"
	if s, ok := n.Shadowing(); ok {
		shadowing = jbool(s)
	}
	return "Note p=" + code(n.Participant()) + " p2=" + code(n.Participant2()) + " pos=" + n.Position().String() +
		" style=" + n.NoteStyle().String() + " display=" + dumpDisplay(n.Display()) + " colors=" +
		n.Colors().String() + " shadowing=" + shadowing + " url=" + dumpURL(n.URL()) + " stereo=" +
		dumpStereo(n.Stereotype()) + " parallel=" + jbool(n.IsParallel())
}

func dumpNotes(notes []*Note) string {
	var sb strings.Builder
	sb.WriteString("[")
	for _, n := range notes {
		sb.WriteString("{" + dumpNote(n) + "}")
	}
	sb.WriteString("]")
	return sb.String()
}

func indexOf(events []Event, e any) int {
	for i, ev := range events {
		if any(ev) == e {
			return i
		}
	}
	return -1
}

func lifeTypes(types []LifeEventType) string {
	var names []string
	for _, t := range types {
		names = append(names, t.String())
	}
	return "[" + strings.Join(names, ", ") + "]"
}

func eventName(e Event) string {
	switch e.(type) {
	case *Message:
		return "Message"
	case *MessageExo:
		return "MessageExo"
	case *LifeEvent:
		return "LifeEvent"
	case *Note:
		return "Note"
	case *Notes:
		return "Notes"
	case *GroupingStart:
		return "GroupingStart"
	case *GroupingLeaf:
		return "GroupingLeaf"
	case *Divider:
		return "Divider"
	case *Delay:
		return "Delay"
	case *HSpace:
		return "HSpace"
	case *Reference:
		return "Reference"
	case *Newpage:
		return "Newpage"
	}
	return fmt.Sprintf("%T", e)
}

func dumpEvent(events []Event, e Event) string {
	var sb strings.Builder
	sb.WriteString(eventName(e))
	switch e := e.(type) {
	case AbstractMessage:
		m := e.Base()
		switch e := e.(type) {
		case *Message:
			sb.WriteString(" p1=" + e.p1.Code() + " p2=" + e.p2.Code() + " multicast=[")
			for _, p := range e.multicast {
				sb.WriteString(p.Code() + ";")
			}
			sb.WriteString("]")
		case *MessageExo:
			sb.WriteString(" p=" + e.participant.Code() + " type=" + e.typ.String() + " short=" + jbool(e.shortArrow))
		}
		sb.WriteString(" label=" + dumpDisplay(m.Label()))
		sb.WriteString(" numbered=" + dumpDisplay(m.LabelNumbered()))
		sb.WriteString(" number=" + jstr(m.MessageNumber()))
		sb.WriteString(" arrow=" + dumpArrow(m.ArrowConfiguration()))
		sb.WriteString(" parallel=" + jbool(m.IsParallel()))
		brother := -1
		if m.parallelBrother != nil {
			brother = indexOf(events, m.parallelBrother)
		}
		sb.WriteString(" brother=" + strconv.Itoa(brother))
		sb.WriteString(" life=" + lifeTypes(m.LifeEventTypes()))
		sb.WriteString(" first=" + jbool(m.firstIsActivate))
		sb.WriteString(" notes=" + dumpNotes(m.NoteOnMessages()))
		sb.WriteString(" url=" + dumpURL(m.url))
		sb.WriteString(" hasurl=" + jbool(m.HasURL()))
		sb.WriteString(" stereo=" + dumpStereo(m.Stereotype()))
		sb.WriteString(" anchors=" + jstr(m.Anchor()) + "/" + jstr(m.Part1Anchor()) + "/" + jstr(m.Part2Anchor()))
		sb.WriteString(" at=" + dumpLocation(m.LineLocation()))
	case *LifeEvent:
		msg := -1
		if e.Message() != nil {
			msg = indexOf(events, e.Message())
		}
		sb.WriteString(" p=" + e.Participant().Code() + " type=" + e.Type().String() + " colors=" +
			dumpFashion(e.SpecificColors()) + " message=" + strconv.Itoa(msg))
	case *Note:
		sb.WriteString(" " + dumpNote(e))
	case *Notes:
		sb.WriteString(" " + dumpNotes(e.List()))
	case Grouping:
		g := e.base()
		sb.WriteString(" type=" + g.Type().String() + " title=" + g.Title() + " comment=" + jstr(g.Comment()))
		sb.WriteString(" general=" + jcolor(e.BackColorGeneral()) + " element=" + jcolor(g.BackColorElement()))
		sb.WriteString(" level=" + strconv.Itoa(e.Level()) + " parallel=" + jbool(e.IsParallel()))
		switch e := e.(type) {
		case *GroupingLeaf:
			sb.WriteString(" start=" + strconv.Itoa(indexOf(events, e.start)) + " notes=" + dumpNotes(e.noteOnMessages))
		case *GroupingStart:
			parent := -1
			if e.parent != nil {
				parent = indexOf(events, e.parent)
			}
			sb.WriteString(" parent=" + strconv.Itoa(parent))
		}
	case *Divider:
		sb.WriteString(" text=" + dumpDisplay(e.Text()))
	case *Delay:
		sb.WriteString(" text=" + dumpDisplay(e.Text()))
	case *HSpace:
		sb.WriteString(" pixel=" + strconv.Itoa(e.Pixel()))
	case *Reference:
		sb.WriteString(" participants=[")
		for _, p := range e.Participants() {
			sb.WriteString(p.Code() + ";")
		}
		sb.WriteString("] strings=" + dumpDisplay(e.Strings()) + " url=" + dumpURL(e.URL()) + " general=" +
			jcolor(e.BackColorGeneral()) + " element=" + jcolor(e.BackColorElement()) + " notes=" +
			dumpNotes(e.NoteOnMessages()))
	}
	return sb.String()
}
