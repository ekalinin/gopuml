// SPDX-License-Identifier: MIT
// Provenance: original

package activitydiagram3

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/security"
	"github.com/ekalinin/gopuml/internal/tim"
	"github.com/ekalinin/gopuml/internal/uml"
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

// parseCase preprocesses the first block of a source and runs the activity
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

// dumpCase is the text Harness.java writes for a case.
func dumpCase(c testCase) []string {
	out := []string{"=== " + c.name}
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
			if keepCompact(line) {
				compact = append(compact, line)
			}
		}
		return append(out, compact...)
	}
	return dumpDiagram(out, d)
}

// isCompact is Harness.isCompact: the generated cases only dump the result,
// the errors, the warnings and the instruction tree.
func isCompact(name string) bool { return strings.HasPrefix(name, "combos/") }

// keepCompact is Harness.keepCompact.
func keepCompact(line string) bool {
	for _, p := range []string{"result", "error", "warning", "ins ", "branch ", "current "} {
		if strings.HasPrefix(line, p) {
			return true
		}
	}
	return false
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
	case *ActivityDiagram3:
		return dumpActivity(out, d)
	}
	return append(out, fmt.Sprintf("result %T", d))
}

func jstr(s string, ok bool) string {
	if !ok {
		return "null"
	}
	return s
}

func jbool(b bool) string { return strconv.FormatBool(b) }

func dumpActivity(out []string, ad *ActivityDiagram3) []string {
	out = append(out, "result ActivityDiagram3",
		"description "+ad.Description(),
		"pages "+strconv.Itoa(ad.Pages()),
		"title "+dumpDP(ad.Title()),
		"caption "+dumpDP(ad.Caption()),
		"legend "+dumpDP(ad.Legend()),
		"header "+dumpDP(ad.Header()),
		"footer "+dumpDP(ad.Footer()),
		"mainframe "+dumpDP(ad.MainFrame()),
		"skinparamused "+jbool(ad.IsSkinParamUsed()))
	pragma := ad.Pragma()
	for k := klimt.PragmaKey(0); !strings.HasPrefix(k.String(), "PragmaKey("); k++ {
		if pragma.IsDefine(k) {
			out = append(out, "pragma "+k.String()+"="+jstr(pragma.Value(k)))
		}
	}
	for _, w := range pragma.Warnings() {
		out = append(out, "warning "+strings.Join(w.Message(), "\\n"))
	}
	values := ad.Skin().Values()
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, k := range keys {
		out = append(out, "skinparam "+k+"="+values[k])
	}
	out = append(out, "hasurl "+jbool(ad.HasURL()), "strategy "+ad.SwimlaneStrategy().String())
	sw := ad.Swimlanes()
	for _, s := range sw.List() {
		out = append(out, "swimlane "+s.Name()+" order="+strconv.Itoa(s.Order())+" display="+dumpDisplay(s.Display())+
			" colors="+s.Colors().String())
	}
	out = append(out, "currentswimlane "+dumpLane(sw.CurrentSwimlane()), "nextlink "+dumpLR(sw.NextLinkRenderer()))
	dm := &dumper{out: out, ids: map[any]int{}}
	dm.ins(sw.Root(), 0, "root")
	current := "?"
	if id, ok := dm.ids[sw.Current()]; ok {
		current = strconv.Itoa(id)
	}
	return append(dm.out, "current "+current)
}

func dumpDisplay(d *klimt.Display) string {
	if d == nil {
		return "null"
	}
	if klimt.DisplayIsNull(d) {
		return "NULL"
	}
	var parts []string
	for _, cs := range d.AsList() {
		prefix := ""
		if _, ok := cs.(*klimt.Stereotype); ok {
			prefix = "S:"
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

func dumpColor(c color.HColor) string {
	if c == nil {
		return "null"
	}
	return c.String()
}

func dumpColors(c *color.Colors) string {
	if c == nil {
		return "null"
	}
	return c.String()
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

func dumpStereogroup(g *klimt.Stereogroup) string {
	if g == nil {
		return "nil"
	}
	if g.IsEmpty() {
		return "null"
	}
	return "[" + strings.Join(g.Labels(), ", ") + "]"
}

func dumpLane(s *Swimlane) string {
	if s == nil {
		return "null"
	}
	return s.Name()
}

// compareUTF16 orders strings as Java's String.compareTo.
func compareUTF16(a, b string) int {
	return slices.Compare(utf16.Encode([]rune(a)), utf16.Encode([]rune(b)))
}

func dumpLanes(set SwimlaneSet) string {
	var names []string
	for _, s := range set {
		name := dumpLane(s)
		if !slices.Contains(names, name) {
			names = append(names, name)
		}
	}
	slices.SortFunc(names, compareUTF16)
	return "[" + strings.Join(names, ", ") + "]"
}

func dumpRainbow(r *Rainbow) string {
	if r == nil {
		return "null"
	}
	var parts []string
	for _, c := range r.Colors() {
		parts = append(parts, dumpColor(c.ArrowColor())+"/"+dumpColor(c.ArrowHeadColor())+"/"+c.Style().String())
	}
	return "[" + strings.Join(parts, ";") + "]/" + strconv.Itoa(r.ColorArrowSeparationSpace())
}

func dumpLR(l *LinkRendering) string {
	if l == nil {
		return "null"
	}
	return "{" + dumpDisplay(l.Display()) + " " + dumpRainbow(l.Rainbow()) + "}"
}

func dumpBoxStyle(b BoxStyle, ok bool) string {
	if !ok {
		return "nil"
	}
	return jstr(b.Name())
}

func dumpNote(n *PositionedNote) string {
	if n == nil {
		return "null"
	}
	return "{" + dumpDisplay(n.Display()) + " " + n.NotePosition().String() + " " + n.Type().String() + " lane=" +
		dumpLane(n.SwimlaneNote()) + " colors=" + dumpColors(n.Colors()) + " stereo=" + dumpStereo(n.Stereotype()) + "}"
}

func dumpNotes(notes []*PositionedNote) string {
	var sb strings.Builder
	sb.WriteString("[")
	for _, n := range notes {
		sb.WriteString(dumpNote(n))
	}
	sb.WriteString("]")
	return sb.String()
}

// safe returns the value, or the simple class name of the Java exception
// it panics with.
func safe(f func() string) (s string) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(*color.RuntimeError); ok {
				class := e.Class
				if i := strings.LastIndexByte(class, '.'); i >= 0 {
					class = class[i+1:]
				}
				s = "!" + class
				return
			}
			panic(r)
		}
	}()
	return f()
}

type dumper struct {
	out []string
	ids map[any]int
}

func (dm *dumper) id(o Instruction) string {
	if o == nil {
		return "null"
	}
	if id, ok := dm.ids[o]; ok {
		return strconv.Itoa(id)
	}
	return "?"
}

func insName(ins Instruction) string {
	return strings.TrimPrefix(fmt.Sprintf("%T", ins), "*activitydiagram3.")
}

func inOut(ins Instruction) string {
	return " in=" + dumpLane(ins.SwimlaneIn()) + " out=" + dumpLane(ins.SwimlaneOut())
}

func (dm *dumper) ins(ins Instruction, depth int, role string) {
	id := len(dm.ids)
	dm.ids[ins] = id
	var sb strings.Builder
	sb.WriteString("ins " + strconv.Itoa(id) + " " + strconv.Itoa(depth) + " " + role + " " + insName(ins))
	sb.WriteString(" lanes=" + safe(func() string { return dumpLanes(ins.Swimlanes()) }))
	sb.WriteString(" break=" + safe(func() string { return jbool(ins.ContainsBreak()) }))
	sb.WriteString(" inlink=" + safe(func() string { return dumpLR(ins.InLinkRendering()) }))
	var children []func()
	switch i := ins.(type) {
	case *InstructionList:
		sb.WriteString(" default=" + dumpLane(i.defaultSwimlane) + " outlink=" + dumpLR(i.outlinkRendering) +
			" empty=" + jbool(i.IsEmpty()) + " onlystop=" + jbool(i.IsOnlySingleStopOrSpot()) +
			" notes=" + dumpNotes(i.notes))
		for _, child := range i.all {
			children = append(children, func() { dm.ins(child, depth+1, "item") })
		}
	case *InstructionSimple:
		sb.WriteString(inOut(i) + " label=" + dumpDisplay(i.label) + " colors=" + dumpColors(i.colors) +
			" box=" + dumpBoxStyle(i.boxStyle, true) + " url=" + dumpURL(i.url) + " stereo=" + dumpStereo(i.stereotype) +
			" killed=" + jbool(i.killed) + " notes=" + dumpNotes(i.notes))
	case *InstructionSpot:
		sb.WriteString(inOut(i) + " spot=" + i.spot + " color=" + dumpColor(i.color) + " killed=" + jbool(i.killed) +
			" notes=" + dumpNotes(i.notes))
	case *InstructionGoto:
		sb.WriteString(inOut(i) + " name=" + i.name + " notes=" + dumpNotes(i.notes))
	case *InstructionLabel:
		sb.WriteString(inOut(i) + " name=" + i.name + " notes=" + dumpNotes(i.notes))
	case *InstructionStart:
		sb.WriteString(inOut(i) + " colors=" + dumpColors(i.colors) + " notes=" + dumpNotes(i.notes))
	case *InstructionStop:
		sb.WriteString(inOut(i) + " colors=" + dumpColors(i.colors) + " notes=" + dumpNotes(i.notes))
	case *InstructionEnd:
		sb.WriteString(inOut(i) + " colors=" + dumpColors(i.colors) + " notes=" + dumpNotes(i.notes))
	case *InstructionBreak:
		sb.WriteString(inOut(i) + " notes=" + dumpNotes(i.notes))
	case *InstructionIf:
		current := "else"
		if i.current != i.elseBranch {
			current = strconv.Itoa(slices.Index(i.thens, i.current))
		}
		sb.WriteString(" parent=" + dm.id(i.parent) + " lane=" + dumpLane(i.swimlane) + " url=" + dumpURL(i.url) +
			" stereo=" + dumpStereo(i.stereotype) + " outcolor=" + dumpLR(i.outColor) + " endif=" + jbool(i.endifCalled) +
			" current=" + current + " notes=" + dumpNotes(i.notes))
		for k, b := range i.thens {
			children = append(children, func() { dm.branch(b, depth+1, "then"+strconv.Itoa(k)) })
		}
		if i.elseBranch != nil {
			children = append(children, func() { dm.branch(i.elseBranch, depth+1, "else") })
		}
	case *InstructionSwitch:
		sb.WriteString(" parent=" + dm.id(i.parent) + " lane=" + dumpLane(i.swimlane) + " test=" +
			dumpDisplay(i.labelTest) + " colors=" + dumpColors(i.colors) + " endcolors=" + dumpColors(i.endColors) +
			" after=" + dumpLR(i.afterEndwhile) + " current=" + strconv.Itoa(slices.Index(i.switches, i.current)) +
			" notes=" + dumpNotes(i.notes))
		for k, b := range i.switches {
			children = append(children, func() { dm.branch(b, depth+1, "case"+strconv.Itoa(k)) })
		}
	case *InstructionWhile:
		sb.WriteString(" parent=" + dm.id(i.parent) + " lane=" + dumpLane(i.swimlane) + " test=" + dumpDisplay(i.test) +
			" yes=" + dumpDisplay(i.yes) + " color=" + dumpColor(i.color) + " killed=" + jbool(i.killed) +
			" testcalled=" + jbool(i.testCalled) + " outcolor=" + dumpLR(i.outColor) +
			" box=" + dumpBoxStyle(i.boxStyle, i.hasBoxStyle) + " backward=" + dumpDisplay(i.backward) +
			" stereo=" + dumpStereo(i.stereotype) + " in1=" + dumpLR(i.incoming1) + " in2=" + dumpLR(i.incoming2) +
			" backwardcalled=" + jbool(i.backwardCalled) + " notes=" + dumpNotes(i.notes))
		if i.specialOut != nil {
			children = append(children, func() { dm.ins(i.specialOut, depth+1, "special") })
		}
		children = append(children, func() { dm.ins(i.repeatList, depth+1, "loop") })
	case *InstructionRepeat:
		sb.WriteString(" parent=" + dm.id(i.parent) + " lane=" + dumpLane(i.swimlane) +
			" laneout=" + dumpLane(i.swimlaneOut) + " lanebackward=" + dumpLane(i.swimlaneBackward) +
			" start=" + dumpDisplay(i.startLabel) + " boxin=" + dumpBoxStyle(i.boxStyleIn, true) +
			" sgloop=" + dumpStereogroup(i.stereogroupLoop) + " test=" + dumpDisplay(i.test) +
			" yes=" + dumpDisplay(i.yes) + " out=" + dumpDisplay(i.out) + " sg2=" + dumpStereogroup(i.stereotype2) +
			" testcalled=" + jbool(i.testCalled) + " killed=" + jbool(i.killed) +
			" endlink=" + dumpLR(i.endRepeatLinkRendering) + " backward=" + dumpDisplay(i.backward) +
			" hasbackward=" + jbool(i.HasBackward()) + " box=" + dumpBoxStyle(i.boxStyle, i.hasBoxStyle) +
			" stereoback=" + dumpStereo(i.stereotypeBack) + " in1=" + dumpLR(i.incoming1) + " in2=" + dumpLR(i.incoming2) +
			" backnotes=" + dumpNotes(i.backwardNotes))
		children = append(children, func() { dm.ins(i.repeatList, depth+1, "loop") })
	case *InstructionFork:
		sb.WriteString(" parent=" + dm.id(i.parent) + inOut(i) + " style=" + i.style.String() +
			" label=" + jstr(i.label, i.hasLabel) + " finished=" + jbool(i.finished) + " colors=" + dumpColors(i.colors) +
			" notes=" + dumpNotes(i.notes))
		for k, l := range i.forks {
			children = append(children, func() { dm.ins(l, depth+1, "fork"+strconv.Itoa(k)) })
		}
	case *InstructionSplit:
		sb.WriteString(" parent=" + dm.id(i.parent) + " in=" + dumpLane(i.swimlaneIn) + " out=" + dumpLane(i.swimlaneOut) +
			" colors=" + dumpColors(i.colors))
		for k, l := range i.splits {
			children = append(children, func() { dm.ins(l, depth+1, "split"+strconv.Itoa(k)) })
		}
	case *InstructionGroup:
		set := color.Instance()
		sb.WriteString(" parent=" + dm.id(i.parent) + " title=" + dumpDisplay(i.title) + " type=" + i.typ.String() +
			" back=" + dumpColor(i.backColor) +
			" styleline=" + dumpColor(i.style.Value(klimt.PNameLineColor).AsColor(set)) +
			" styleback=" + dumpColor(i.style.Value(klimt.PNameBackGroundColor).AsColor(set)) +
			" note=" + dumpNote(i.note))
		children = append(children, func() { dm.ins(i.list, depth+1, "body") })
	}
	dm.out = append(dm.out, sb.String())
	for _, c := range children {
		c()
	}
}

func (dm *dumper) branch(b *Branch, depth int, role string) {
	dm.out = append(dm.out, "branch "+strconv.Itoa(depth)+" "+role+" test="+dumpDisplay(b.labelTest)+
		" positive="+dumpLR(b.labelPositive)+" inlink="+dumpLR(b.inlinkRendering)+" inlabel="+dumpLR(b.inlabel)+
		" special="+dumpLR(b.special)+" color="+dumpColor(b.color)+" specialcolors="+dumpColors(b.specialColors)+
		" getcolor="+dumpColor(b.Color()))
	dm.ins(b.list, depth+1, "body")
}
