// SPDX-License-Identifier: MIT
// Provenance: original

package classdiagram

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/ekalinin/plantuml-in-go/internal/abel"
	"github.com/ekalinin/plantuml-in-go/internal/command"
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

// parseCase preprocesses the first block of a source and runs the class
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

// panicText is the text of a recovered panic, as Throwable.toString.
func panicText(r any) string {
	if e, ok := r.(*color.RuntimeError); ok {
		return e.Error()
	}
	return fmt.Sprint(r)
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
				err = fmt.Errorf("crash %s", panicText(r))
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
			if strings.HasPrefix(line, "result") || strings.HasPrefix(line, "error") ||
				strings.HasPrefix(line, "link") || (strings.HasPrefix(c.name, "declcombos/") &&
				(strings.HasPrefix(line, "entity ent") || strings.HasPrefix(line, "body "))) {
				compact = append(compact, line)
			}
		}
		return append(out, compact...)
	}
	return dumpDiagram(out, d)
}

// isCompact is Harness.isCompact.
func isCompact(name string) bool {
	return strings.HasPrefix(name, "linkcombos/") || strings.HasPrefix(name, "declcombos/")
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
	case *abel.NewpagedDiagram:
		out = append(out, "result NewpagedDiagram",
			"description "+d.Description(),
			"pages "+strconv.Itoa(d.Pages()))
		for i, p := range d.Diagrams() {
			out = append(out, "page "+strconv.Itoa(i))
			out = dumpClass(out, p.(*ClassDiagram))
		}
		return out
	case *ClassDiagram:
		out = append(out, "result ClassDiagram")
		return dumpClass(out, d)
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

// safe is Harness.safe: the text, or "EXC" and the exception (without the
// message of a PatternSyntaxException).
func safe(f func() string) (s string) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(*color.RuntimeError); ok && e.Class == "java.util.regex.PatternSyntaxException" {
				s = "EXC " + e.Class
				return
			}
			s = "EXC " + panicText(r)
		}
	}()
	return f()
}

type togetherIDs struct {
	ids   map[*abel.Together]int
	order []*abel.Together
}

func (t *togetherIDs) id(together *abel.Together) int {
	if id, ok := t.ids[together]; ok {
		return id
	}
	id := len(t.order)
	t.ids[together] = id
	t.order = append(t.order, together)
	return id
}

func dumpClass(out []string, cd *ClassDiagram) []string {
	togethers := &togetherIDs{ids: map[*abel.Together]int{}}
	out = append(out,
		"description "+cd.Description(),
		"pages "+strconv.Itoa(cd.Pages()),
		"type "+cd.DiagramType().String(),
		"title "+dumpDP(cd.Title()),
		"caption "+dumpDP(cd.Caption()),
		"legend "+dumpDP(cd.Legend()),
		"header "+dumpDP(cd.Header()),
		"footer "+dumpDP(cd.Footer()),
		"mainframe "+dumpDP(cd.MainFrame()))
	h, v := cd.SplitPages()
	out = append(out, "scale "+dumpScale(cd.Scale()),
		"splitpages "+strconv.Itoa(h)+" "+strconv.Itoa(v),
		"namespaceseparator "+jstr(cd.NamespaceSeparator()),
		"skinparamused "+jbool(cd.IsSkinParamUsed()),
		"smetana "+jbool(cd.IsUseSmetana())+" elk "+jbool(cd.IsUseElk()))
	pragma := cd.Pragma()
	for k := klimt.PragmaKey(0); !strings.HasPrefix(k.String(), "PragmaKey("); k++ {
		if pragma.IsDefine(k) {
			out = append(out, "pragma "+k.String()+"="+jstr(pragma.Value(k)))
		}
	}
	for _, w := range pragma.Warnings() {
		out = append(out, "warning "+strings.Join(w.Message(), "\\n"))
	}
	values := cd.Skin().Values()
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, k := range keys {
		out = append(out, "skinparam "+k+"="+values[k])
	}
	sprites := slices.Clone(cd.Skin().AllSpriteNames())
	slices.Sort(sprites)
	for _, name := range sprites {
		out = append(out, "sprite "+name)
	}
	out = append(out, "defaultskin "+cd.Skin().DefaultSkin(),
		"rankdir "+cd.Skin().Rankdir().String(),
		"allowmixing "+jbool(cd.IsAllowMixing())+" layoutexplicit "+strconv.Itoa(cd.UseLayoutExplicit())+
			" rawlayout "+strconv.Itoa(cd.RawLayout()),
		"visibilitymodifierpresent "+jbool(cd.IsVisibilityModifierPresent()),
		"hidevisibility "+dumpVisibilities(cd.HidesVisibilityModifier()),
		"lastentity "+uid(cd.LastEntity()))
	stack := "stack"
	for _, bag := range cd.Stacks() {
		switch b := bag.(type) {
		case *abel.Entity:
			stack += " " + b.UID()
		case *abel.Together:
			stack += " T" + strconv.Itoa(togethers.id(b))
		}
	}
	out = append(out, stack)
	for _, h := range cd.HideOrShows() {
		out = append(out, "hideorshow gender="+jstr(h.Gender.Gender())+" portion="+h.Portion.String()+" show="+
			jbool(h.Show))
	}
	for _, h := range cd.Hides2() {
		out = append(out, "hides2 "+h.String())
	}
	for _, h := range cd.Removed() {
		out = append(out, "removed "+h.String())
	}
	for _, q := range cd.Quarks() {
		parent := "null"
		if q.Parent() != nil {
			parent = "[" + q.Parent().QualifiedName() + "]"
		}
		out = append(out, "quark ["+q.QualifiedName()+"] name=["+q.Name()+"] parent="+parent+" data="+
			uid(q.Data())+" children="+strconv.Itoa(q.CountChildren()))
	}
	for _, q := range cd.Quarks() {
		e := q.Data()
		if e == nil {
			continue
		}
		out = append(out, "entity "+dumpEntity(e, togethers))
		if e.Bodier() != nil {
			out = append(out, "body "+e.UID()+" "+dumpBody(e))
		}
		if !e.IsRoot() && !e.IsGroup() && (e.LeafType().IsLikeClass() || e.LeafType() == abel.LeafTypeObject) {
			line := "portions " + e.UID()
			for p := uml.PortionField; p <= uml.PortionStereotype; p++ {
				line += " " + p.String() + "=" + safe(func() string { return jbool(cd.ShowPortion(p, e)) })
			}
			line += " stereolabels=" + safe(func() string { return dumpStringList(cd.VisibleStereotypeLabels(e)) })
			out = append(out, line)
		}
	}
	for _, l := range cd.Links() {
		out = append(out, "link "+dumpLink(l))
	}
	for i, t := range togethers.order {
		parent := "null"
		if t.Parent() != nil {
			parent = "T" + strconv.Itoa(togethers.id(t.Parent()))
		}
		out = append(out, "together T"+strconv.Itoa(i)+" parent="+parent)
	}
	return out
}

func dumpStringList(list []string) string {
	if list == nil {
		return "null"
	}
	return "[" + strings.Join(list, ", ") + "]"
}

func uid(e *abel.Entity) string {
	if e == nil {
		return "null"
	}
	return e.UID()
}

func dumpVisibilities(set abel.VisibilitySet) string {
	var sb strings.Builder
	sb.WriteString("[")
	for _, v := range set.List() {
		sb.WriteString(v.String() + ";")
	}
	sb.WriteString("]")
	return sb.String()
}

func dumpVisibility(v klimt.VisibilityModifier, ok bool) string {
	if !ok {
		return "null"
	}
	return v.String()
}

func dumpEntity(e *abel.Entity, togethers *togetherIDs) string {
	var sb strings.Builder
	sb.WriteString(e.UID() + " quark=[" + e.Quark().QualifiedName() + "]")
	if e.IsGroup() {
		sb.WriteString(" group=" + e.GroupType().String())
	} else {
		sb.WriteString(" leaf=" + e.LeafType().String())
	}
	sb.WriteString(" display=" + dumpDisplay(e.Display()))
	mode := "null"
	if !klimt.DisplayIsNull(e.Display()) {
		mode = e.Display().DefaultCreoleMode().String()
	}
	sb.WriteString(" mode=" + mode)
	sb.WriteString(" stereo=" + dumpStereo(e.Stereotype()))
	sb.WriteString(" styles=[" + strings.Join(e.Stereostyles().StyleNames(), ", ") + "]")
	sb.WriteString(" tags=[")
	for _, t := range e.Stereotags() {
		sb.WriteString(t.String() + ";")
	}
	sb.WriteString("]")
	sb.WriteString(" generic=" + e.GenericValue().String())
	sb.WriteString(" visibility=" + dumpVisibility(e.VisibilityModifier()))
	sb.WriteString(" colors=" + dumpColors(e.Colors()))
	sb.WriteString(" url=" + dumpURL(e.URL99()))
	sb.WriteString(" symbol=" + e.USymbol().String())
	sb.WriteString(" static=" + jbool(e.IsStatic()) + " packed=" + jbool(e.IsPacked()))
	sb.WriteString(" rawlayout=" + strconv.Itoa(e.RawLayout()))
	together := "null"
	if t := e.Together(); t != nil {
		together = "T" + strconv.Itoa(togethers.id(t))
	}
	sb.WriteString(" together=" + together)
	sb.WriteString(" at=" + dumpLocation(e.Location()))
	legend := "null"
	if l := e.Legend(); l != nil {
		legend = dumpDP(*l)
	}
	sb.WriteString(" legend=" + legend)
	sb.WriteString(" notestop=" + dumpNotes(e.Notes(abel.PositionTop)))
	sb.WriteString(" notesbottom=" + dumpNotes(e.Notes(abel.PositionBottom)))
	sb.WriteString(" tips=[")
	members, tips := e.Tips()
	for _, m := range members {
		t := tips[m]
		sb.WriteString("{" + m + " " + dumpDisplay(t.Display()) + " " + dumpColors(t.Colors()) + " " +
			dumpStereo(t.Stereotype()) + "}")
	}
	sb.WriteString("]")
	ports := slices.Clone(e.PortShortNamesValue())
	slices.Sort(ports)
	sb.WriteString(" ports=" + dumpStringList(append([]string{}, ports...)))
	if !e.IsRoot() {
		sb.WriteString(" hidden=" + safe(func() string { return jbool(e.IsHidden()) }))
		sb.WriteString(" removed=" + safe(func() string { return jbool(e.IsRemoved()) }))
	}
	sb.WriteString(" position=" + safe(func() string { return e.EntityPosition().String() }))
	sb.WriteString(" hasurl=" + safe(func() string { return jbool(e.HasURL()) }))
	return sb.String()
}

func dumpBody(e *abel.Entity) string {
	b := e.Bodier()
	if j, ok := b.(*abel.BodierJSon); ok {
		if j.JSON() == nil {
			return "json null"
		}
		return "json " + j.JSON().String()
	}
	var sb strings.Builder
	switch b.(type) {
	case *abel.BodierLikeClassOrObject:
		sb.WriteString("BodierLikeClassOrObject")
	case *abel.BodierMap:
		sb.WriteString("BodierMap")
	case *abel.BodierSimple:
		sb.WriteString("BodierSimple")
	}
	sb.WriteString(" raw=[")
	for _, cs := range b.RawBody() {
		sb.WriteString(cs.String() + ";")
	}
	sb.WriteString("]")
	switch b := b.(type) {
	case *abel.BodierMap:
		keys, values := b.Entries()
		parts := make([]string, len(keys))
		for i, k := range keys {
			parts[i] = k + "=" + values[k]
		}
		sb.WriteString(" map={" + strings.Join(parts, ", ") + "}")
	case *abel.BodierLikeClassOrObject:
		sb.WriteString(" type=" + b.Type().String())
		sb.WriteString(" fields=" + safe(func() string { return dumpMembers(b.FieldsToDisplay()) }))
		sb.WriteString(" methods=" + safe(func() string { return dumpMembers(b.MethodsToDisplay()) }))
	}
	return sb.String()
}

func dumpMembers(d *klimt.Display) string {
	var sb strings.Builder
	sb.WriteString("[")
	for _, cs := range d.AsList() {
		m := cs.(*abel.Member)
		mods := ""
		if m.IsStatic() {
			mods += "S"
		}
		if m.IsAbstract() {
			mods += "A"
		}
		sb.WriteString("{" + dumpVisibility(m.VisibilityModifier()) + "|" + mods + "|" + m.Display(false) + "|" +
			dumpURL(m.URL()) + "|" + m.String() + "}")
	}
	sb.WriteString("]")
	return sb.String()
}

func dumpNotes(notes []*abel.CucaNote) string {
	var sb strings.Builder
	sb.WriteString("[")
	for _, n := range notes {
		sb.WriteString("{" + dumpNote(n) + "}")
	}
	sb.WriteString("]")
	return sb.String()
}

func dumpNote(n *abel.CucaNote) string {
	if n == nil {
		return "null"
	}
	return dumpDisplay(n.Display()) + " " + n.Position().String() + " " + dumpColors(n.Colors()) + " " +
		n.Strategy().String()
}

func dumpLink(l *abel.Link) string {
	var sb strings.Builder
	sb.WriteString(l.UID() + " e1=" + l.Entity1().UID() + " e2=" + l.Entity2().UID())
	sb.WriteString(" type=" + l.Type().String() + " middle=" + l.Type().MiddleDecor().String())
	a := l.LinkArg()
	sb.WriteString(" label=" + dumpDisplay(a.Label()))
	sb.WriteString(" q1=" + a.Quantifier1().String() + " q2=" + a.Quantifier2().String())
	sb.WriteString(" r1=" + a.Role1().String() + " r2=" + a.Role2().String())
	sb.WriteString(" kal1=" + a.Kal1().String() + " kal2=" + a.Kal2().String())
	sb.WriteString(" ld=" + a.Labeldistance().String() + " la=" + a.Labelangle().String())
	sb.WriteString(" length=" + strconv.Itoa(a.Length()) + " vis=" + dumpVisibility(a.VisibilityModifier()))
	sb.WriteString(" arrow=" + l.LinkArrow().String() + " inverted=" + jbool(l.IsInverted()))
	sb.WriteString(" invis=" + jbool(l.IsInvis()) + " hiddenflag=" + jbool(l.Hidden))
	sb.WriteString(" single=" + jbool(l.IsSingle()) + " nodestyle=" + jbool(l.UseNodeStyle()))
	sb.WriteString(" constraint=" + jbool(l.IsConstraint()) + " weight=" + jfmt.DoubleToString(l.Weight()))
	sb.WriteString(" solitary=" + jbool(l.IsHorizontalSolitary()) + " opale=" + jbool(l.IsOpale()))
	sb.WriteString(" sametail=" + l.Sametail().String())
	sb.WriteString(" ports=" + l.PortName1().String() + "/" + l.PortName2().String())
	sb.WriteString(" url=" + dumpURL(l.URL()))
	sb.WriteString(" stereo=" + dumpStereo(l.Stereotype()))
	sb.WriteString(" colors=" + dumpColors(l.Colors()))
	sb.WriteString(" supplementary=[")
	for _, c := range l.SupplementaryColors() {
		sb.WriteString(dumpColors(c) + ";")
	}
	sb.WriteString("]")
	sb.WriteString(" note=" + dumpNote(l.Note()))
	if lc := l.LinkConstraint(); lc == nil {
		sb.WriteString(" linkconstraint=null")
	} else {
		l1, l2 := lc.Links()
		sb.WriteString(" linkconstraint=" + l1.UID() + "/" + l2.UID() + " " + dumpDisplay(lc.Display()))
	}
	sb.WriteString(" codeline=" + jstr(l.CodeLine()))
	sb.WriteString(" at=" + dumpLocation(l.Location()))
	sb.WriteString(" hidden=" + safe(func() string { return jbool(l.IsHidden()) }))
	sb.WriteString(" removed=" + safe(func() string { return jbool(l.IsRemoved()) }))
	return sb.String()
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

func dumpColors(c *color.Colors) string {
	if c == nil {
		return "null"
	}
	shadowing := "null"
	if s, ok := c.Shadowing(); ok {
		shadowing = jbool(s)
	}
	return "{" + c.String() + " shadowing=" + shadowing + "}"
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
		case *abel.Member:
			prefix = "M:"
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
	return dumpDisplay(dp.Display()) + " h=" + h + " v=" + v + " at=" + dumpLocation(dp.LineLocation())
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

func dumpLocation(l *text.LineLocation) string {
	if l == nil {
		return "null"
	}
	return strconv.Itoa(l.Position)
}

var _ command.Diagram = (*ClassDiagram)(nil)
