// SPDX-License-Identifier: MIT
// Provenance: original

package rose

import (
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/ekalinin/plantuml-in-go/internal/command"
	"github.com/ekalinin/plantuml-in-go/internal/core"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
	"github.com/ekalinin/plantuml-in-go/internal/text"
)

// The golden file compared here is written by
// testdata/javagen/RoseGolden.java (instructions in its header).

func newSkinParam() *klimt.SkinParam {
	return klimt.SkinParamCreate(map[string]string{}, nil, core.Sequence, klimt.PragmaCreateEmpty(),
		klimt.NewConfigurationStore[klimt.OptionKey]())
}

var setups = []string{"default", "params", "style", "strict", "style2"}

func setup(t *testing.T, name string) *klimt.SkinParam {
	sp := newSkinParam()
	switch name {
	case "params":
		sp.SetParam("roundcorner", "12")
		sp.SetParam("shadowing", "true")
		sp.SetParam("defaultFontSize", "15")
		sp.SetParam("actorStyle", "awesome")
		sp.SetParam("responseMessageBelowArrow", "true")
		sp.SetParam("maxMessageSize", "60")
		sp.SetParam("noteTextAlignment", "center")
		sp.SetParam("SequenceMessagePadding", "3")
		sp.SetParam("SequenceMessageMargin", "2")
		sp.SetParam("SequenceMessageBorderColor", "red")
		sp.SetParam("SequenceMessageBackgroundColor", "#EEEEEE")
		sp.SetParam("arrowThickness", "2.5")
		sp.SetParam("participantBackgroundColor", "#CCE5FF")
		sp.SetParam("noteBackgroundColor", "#FFFFAA")
		sp.SetParam("sequenceDividerBackgroundColor", "#F0E0D0")
		sp.SetParam("sequenceLifeLineBorderThickness", "2")
		sp.SetParam("sequenceGroupBorderThickness", "3")
	case "style", "style2":
		file := "rose_style.css"
		if name == "style2" {
			file = "rose_style2.css"
		}
		data, err := os.ReadFile(filepath.Join(goldenDir, file))
		if err != nil {
			t.Fatal(err)
		}
		all := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
		var lines []text.StringLocated
		for _, l := range all[1 : len(all)-1] {
			lines = append(lines, text.NewStringLocated(l, nil))
		}
		styles, err := klimt.NewStyleParserCounter(sp.CurrentStyleBuilder()).Parse(command.NewBlocLines(lines))
		if err != nil {
			t.Fatal(err)
		}
		sp.MuteStyle(styles)
		sp.ApplyPendingStyleMigration()
	case "strict":
		sp.SetParam("style", "strictuml")
		sp.SetParam("actorStyle", "hollow")
		sp.SetParam("noteTextAlignment", "right")
		sp.SetParam("SequenceMessageAlignment", "right")
	}
	return sp
}

func seq(sp *klimt.SkinParam, names ...klimt.SName) *klimt.Style {
	all := append([]klimt.SName{klimt.SNameRoot, klimt.SNameElement, klimt.SNameSequenceDiagram}, names...)
	return klimt.StyleSignatureBasicOf(all...).MergedStyle(sp.CurrentStyleBuilder())
}

func className(c any) string {
	if c == nil {
		return "null"
	}
	return reflect.TypeOf(c).Elem().Name()
}

type area [6]float64

var (
	areaPref  = []area{{}}
	areaWide  = []area{{}, {30, 12}}
	areaNote  = []area{{}, {40, 6}, {-1, 0}}
	areaArrow = []area{{}, {40, 0}, {40, 0, 0, 0, 0, 9}, {40, 0, 0, 0, 0, -9}}
	areaSelf  = []area{{}, {20, 0, -5, 0, 5, 0}, {20, 0, 5, 1, 5, 0}, {20, 0, 5, 2, 5, 0}, {20, 0, -5, 2, 5, 0},
		{20, 0, 0, 3, 5, 0}}
	areaLine = []area{{}, {20, 30}, {9, -20}}
)

// record is the harness's record: the class, the preferred size, the arrow
// points and the drawings in the areas.
func record(out *[]string, caseName string, comp klimt.Component, bn string, b font.StringBounder, areas []area) {
	var log []string
	log = append(log, "class "+className(comp))
	step(&log, func() {
		dim := comp.PreferredDimension(b)
		log = append(log, "pref "+d2s(dim.Width())+"x"+d2s(dim.Height()))
	})
	if a, ok := comp.(klimt.ArrowComponent); ok {
		full := geom.NewXDimension2D(120, 40)
		step(&log, func() {
			p := a.StartPoint(b, full)
			log = append(log, "start "+d2s(p.X)+","+d2s(p.Y))
		})
		step(&log, func() {
			p := a.EndPoint(b, full)
			log = append(log, "end "+d2s(p.X)+","+d2s(p.Y))
		})
		step(&log, func() { log = append(log, "ypoint "+d2s(a.YPoint(b))) })
		step(&log, func() { log = append(log, "paddingy "+d2s(a.PaddingY())) })
		step(&log, func() { log = append(log, "posarrow "+d2s(a.PosArrow(b))) })
		if self, ok := comp.(*ComponentRoseSelfArrow); ok {
			step(&log, func() { log = append(log, "drawn "+d2s(self.DrawnWidth(b))) })
		}
	}
	for _, ar := range areas {
		step(&log, func() {
			pref := comp.PreferredDimension(b)
			a := klimt.AreaCreate(pref.Width()+ar[0], pref.Height()+ar[1])
			a.SetDeltaX1(ar[2])
			a.SetLevel(int(ar[3]))
			a.SetLiveDeltaSize(ar[4])
			a.SetTextDeltaX(ar[5])
			log = append(log, "area "+a.String()+" lvl="+strconv.Itoa(a.Level())+" live="+d2s(a.LiveDeltaSize())+
				" tdx="+d2s(a.TextDeltaX()))
			ug := newRec(&log, b)
			log = append(log, "background")
			step(&log, func() { comp.DrawU(ug, a, klimt.NewSimpleContext2D(true)) })
			log = append(log, "foreground")
			step(&log, func() { comp.DrawU(ug, a, klimt.NewSimpleContext2D(false)) })
		})
	}
	for _, l := range log {
		*out = append(*out, caseName+"\t"+bn+"\t"+l)
	}
}

func comp(out *[]string, caseName string, f func() klimt.Component, areas []area) {
	for _, b := range bounders {
		var c klimt.Component
		var log []string
		step(&log, func() { c = f() })
		if len(log) > 0 {
			*out = append(*out, caseName+"\t"+b.name+"\t"+log[0])
			continue
		}
		record(out, caseName, c, b.name, b.bounder, areas)
	}
}

func display(spec string) *klimt.Display {
	switch spec {
	case "simple":
		return klimt.DisplayCreate("Alice")
	case "multi":
		return klimt.DisplayWithNewlines(klimt.PragmaCreateEmpty(), `Long participant\nsecond line`)
	case "creole":
		return klimt.DisplayCreate("**Bold** and <color:red>red</color>", "----", "<i>after</i> rule")
	case "empty":
		return klimt.DisplayCreate("")
	case "hello":
		return klimt.DisplayCreate("hello")
	case "message":
		return klimt.DisplayWithNewlines(klimt.PragmaCreateEmpty(),
			`a much longer message text\nwith <b>two</b> lines and some more words to wrap`)
	case "note":
		return klimt.DisplayCreate("note text")
	case "notesep":
		return klimt.DisplayCreate("first", "----", "after line", "==title==", "end")
	case "notecreole":
		return klimt.DisplayCreate("* bullet one", "* bullet **two**", "<u>under</u> text")
	case "notelong":
		return klimt.DisplayCreate("This is a rather long note text that should wrap when a maximum width is set")
	}
	panic(spec)
}

// participantType is upstream's sequencediagram.ParticipantType, for the
// test participant.
type participantType int

const (
	ptParticipant participantType = iota
	ptActor
	ptBoundary
	ptControl
	ptEntity
	ptQueue
	ptDatabase
	ptCollections
)

var participantSNames = []klimt.SName{klimt.SNameParticipant, klimt.SNameActor, klimt.SNameBoundary,
	klimt.SNameControl, klimt.SNameEntity, klimt.SNameQueue, klimt.SNameDatabase, klimt.SNameCollections}

func headType(t participantType, head bool) klimt.ComponentType {
	heads := []klimt.ComponentType{klimt.ComponentTypeParticipantHead, klimt.ComponentTypeActorHead,
		klimt.ComponentTypeBoundaryHead, klimt.ComponentTypeControlHead, klimt.ComponentTypeEntityHead,
		klimt.ComponentTypeQueueHead, klimt.ComponentTypeDatabaseHead, klimt.ComponentTypeCollectionsHead}
	tails := []klimt.ComponentType{klimt.ComponentTypeParticipantTail, klimt.ComponentTypeActorTail,
		klimt.ComponentTypeBoundaryTail, klimt.ComponentTypeControlTail, klimt.ComponentTypeEntityTail,
		klimt.ComponentTypeQueueTail, klimt.ComponentTypeDatabaseTail, klimt.ComponentTypeCollectionsTail}
	if head {
		return heads[t]
	}
	return tails[t]
}

// testParticipant computes its styles as upstream's
// sequencediagram.Participant.getUsedStyles.
type testParticipant struct {
	typ          participantType
	stereotype   *klimt.Stereotype
	colors       *color.Colors
	styleBuilder *klimt.StyleBuilder
}

var _ Participant = (*testParticipant)(nil)

func (p *testParticipant) UsedStyles() []*klimt.Style {
	signature := klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameSequenceDiagram,
		participantSNames[p.typ]).AddClickable(nil)
	tmp := signature.WithTOBECHANGED(p.stereotype).MergedStyle(p.styleBuilder)
	tmp = tmp.EventuallyOverrideColors(p.colors)
	stereo := signature.ForStereotypeItself(p.stereotype).MergedStyle(p.styleBuilder)
	if tmp != nil {
		stereo = tmp.MergeWith(stereo, klimt.MergeOverwriteExistingValue)
	}
	return []*klimt.Style{tmp, stereo}
}

func participants(t *testing.T, out *[]string, setupName string) {
	type variant struct {
		display, stereo, back string
	}
	variants := []variant{{"simple", "", ""}, {"multi", "<<foo>>", ""}, {"creole", "", "#FFAAAA"},
		{"simple", "<<bar>> <<foo>>", "#AAFFAA"}}
	for pt := ptParticipant; pt <= ptCollections; pt++ {
		for _, head := range []bool{true, false} {
			for v, vr := range variants {
				caseName := setupName + "/participant/" + headType(pt, head).String() + "/" + strconv.Itoa(v)
				comp(out, caseName, func() klimt.Component {
					sp := setup(t, setupName)
					p := &testParticipant{typ: pt, colors: color.EmptyColors(), styleBuilder: sp.CurrentStyleBuilder()}
					d := display(vr.display)
					if vr.stereo != "" {
						p.stereotype = klimt.StereotypeBuild(vr.stereo)
						d = d.Add(p.stereotype)
					}
					if vr.back != "" {
						c, err := color.Instance().GetColor(vr.back)
						if err != nil {
							t.Fatal(err)
						}
						p.colors = p.colors.Add(color.TypeBack, c)
					}
					return NewRose().CreateComponentParticipant(p, headType(pt, head), nil, sp, d)
				}, areaWide)
			}
		}
	}
}

func arrowConfigs() []*klimt.ArrowConfiguration {
	n := klimt.ArrowConfigurationWithDirectionNormal()
	b := klimt.ArrowConfigurationWithDirectionBoth()
	return []*klimt.ArrowConfiguration{
		n,
		n.WithBody(klimt.ArrowBodyDotted),
		n.WithBody(klimt.ArrowBodyHidden),
		n.Reverse(),
		b,
		b.WithBody(klimt.ArrowBodyDotted),
		n.WithHead2(klimt.ArrowHeadAsync),
		n.WithHead1(klimt.ArrowHeadAsync).WithHead2(klimt.ArrowHeadAsync),
		b.WithHead1(klimt.ArrowHeadAsync),
		n.WithPart(klimt.ArrowPartTopPart),
		n.WithPart(klimt.ArrowPartBottomPart),
		b.WithPart(klimt.ArrowPartTopPart),
		n.WithHead2(klimt.ArrowHeadAsync).WithPart(klimt.ArrowPartBottomPart),
		n.WithHead2(klimt.ArrowHeadCrossx),
		n.WithHead1(klimt.ArrowHeadCrossx),
		n.WithDecoration2(klimt.ArrowDecorationCircle),
		n.WithDecoration1(klimt.ArrowDecorationCircle),
		n.WithDecoration1(klimt.ArrowDecorationCircle).WithDecoration2(klimt.ArrowDecorationCircle),
		b.WithDecoration1(klimt.ArrowDecorationCircle).WithDecoration2(klimt.ArrowDecorationCircle),
		n.WithHead1(klimt.ArrowHeadAsync).WithDecoration1(klimt.ArrowDecorationCircle),
		n.WithHead2(klimt.ArrowHeadCrossx).WithDecoration2(klimt.ArrowDecorationCircle),
		n.WithHead2(klimt.ArrowHeadNone).WithDecoration2(klimt.ArrowDecorationCircle),
		n.WithHead1(klimt.ArrowHeadCrossx).WithDecoration1(klimt.ArrowDecorationCircle),
		n.ReverseDefine(),
		n.ReverseDefine().WithBody(klimt.ArrowBodyDotted).Reverse(),
		n.WithInclination(3),
		n.WithInclination(3).Reverse(),
		b.WithInclination(4).WithHead1(klimt.ArrowHeadAsync).WithHead2(klimt.ArrowHeadAsync),
		n.WithColor(color.Red).WithThickness(3),
		n.Self(),
		n.Self().ReverseDefine(),
		klimt.ArrowConfigurationWithDirectionSelf(false).WithBody(klimt.ArrowBodyDotted),
		klimt.ArrowConfigurationWithDirectionSelf(true),
		n.Self().WithHead2(klimt.ArrowHeadAsync),
		n.Self().WithHead1(klimt.ArrowHeadAsync).WithHead2(klimt.ArrowHeadAsync).WithPart(klimt.ArrowPartTopPart),
		n.Self().WithHead1(klimt.ArrowHeadAsync).WithPart(klimt.ArrowPartBottomPart),
		n.Self().WithHead2(klimt.ArrowHeadCrossx),
		n.Self().WithHead1(klimt.ArrowHeadCrossx),
		n.Self().WithHead1(klimt.ArrowHeadNormal),
		n.Self().WithPart(klimt.ArrowPartTopPart),
		n.Self().WithPart(klimt.ArrowPartBottomPart).ReverseDefine(),
		n.Self().WithDecoration1(klimt.ArrowDecorationCircle).WithDecoration2(klimt.ArrowDecorationCircle),
		n.Self().WithDecoration1(klimt.ArrowDecorationCircle).WithHead1(klimt.ArrowHeadNormal).ReverseDefine(),
		n.Self().WithDecoration2(klimt.ArrowDecorationCircle).ReverseDefine(),
		n.Self().WithHead1(klimt.ArrowHeadCrossx).WithHead2(klimt.ArrowHeadCrossx).ReverseDefine(),
		n.Self().WithHead2(klimt.ArrowHeadAsync).ReverseDefine(),
		n.Self().WithBody(klimt.ArrowBodyHidden),
		n.WithPart(klimt.ArrowPartTopPart).Reverse(),
		n.WithPart(klimt.ArrowPartBottomPart).Reverse(),
		n.Self().WithDecoration1(klimt.ArrowDecorationCircle).WithHead1(klimt.ArrowHeadAsync),
		n.Self().WithDecoration1(klimt.ArrowDecorationCircle).ReverseDefine(),
	}
}

func arrows(t *testing.T, out *[]string, setupName string) {
	configs := arrowConfigs()
	for i, config := range configs {
		labels := []string{"hello"}
		if i < 2 {
			labels = []string{"hello", "empty", "message", "creole"}
		}
		areas := areaArrow
		if config.IsSelfArrow() {
			areas = areaSelf
		}
		for _, label := range labels {
			caseName := setupName + "/arrow/" + strconv.Itoa(i) + "/" + label
			comp(out, caseName, func() klimt.Component {
				sp := setup(t, setupName)
				return NewRose().CreateComponentArrow([]*klimt.Style{seq(sp, klimt.SNameArrow)}, config, sp,
					display(label))
			}, areas)
		}
		if config.IsSelfArrow() && i < 32 {
			for _, label := range []string{"empty", "message"} {
				caseName := setupName + "/arrow/" + strconv.Itoa(i) + "/" + label
				comp(out, caseName, func() klimt.Component {
					sp := setup(t, setupName)
					return NewRose().CreateComponentArrow([]*klimt.Style{seq(sp, klimt.SNameArrow)}, config, sp,
						display(label))
				}, areaSelf)
			}
		}
	}
}

func notes(t *testing.T, out *[]string, setupName string) {
	types := []klimt.ComponentType{klimt.ComponentTypeNote, klimt.ComponentTypeNoteHexagonal,
		klimt.ComponentTypeNoteBox}
	positions := []string{"null", "LEFT", "OVER_SEVERAL"}
	texts := []string{"note", "notesep", "notecreole", "notelong"}
	for _, typ := range types {
		for _, pos := range positions {
			for _, txt := range texts {
				for _, colored := range []bool{false, true} {
					if colored && txt != "note" {
						continue
					}
					caseName := setupName + "/note/" + typ.String() + "/" + pos + "/" + txt
					if colored {
						caseName += "/colored"
					}
					comp(out, caseName, func() klimt.Component {
						sp := setup(t, setupName)
						var st *klimt.Style
						switch typ {
						case klimt.ComponentTypeNoteHexagonal:
							st = seq(sp, klimt.SNameNote, klimt.SNameHnote)
						case klimt.ComponentTypeNoteBox:
							st = seq(sp, klimt.SNameNote, klimt.SNameRnote)
						default:
							st = seq(sp, klimt.SNameNote)
						}
						colors := color.EmptyColors()
						if colored {
							c, err := color.Instance().GetColor("#AAFFAA")
							if err != nil {
								t.Fatal(err)
							}
							colors = colors.Add(color.TypeBack, c).Add(color.TypeLine, color.Blue)
						}
						if pos == "null" {
							return NewRose().CreateComponentNote([]*klimt.Style{st}, typ, sp, display(txt), colors)
						}
						return NewRose().CreateComponentNoteAt([]*klimt.Style{st}, typ, sp, display(txt), colors,
							NotePositionValueOf(pos))
					}, areaNote)
				}
			}
		}
	}
}

func others(t *testing.T, out *[]string, setupName string) {
	rose := NewRose()
	headers := [][]string{{"alt", "x > 5"}, {"loop"}, {"group", `multi\nline comment`}, {"critical", ""}}
	for _, teoz := range []bool{false, true} {
		ht := klimt.ComponentTypeGroupingHeaderLegacy
		if teoz {
			ht = klimt.ComponentTypeGroupingHeaderTeoz
		}
		for i, h := range headers {
			comp(out, setupName+"/"+ht.String()+"/"+strconv.Itoa(i), func() klimt.Component {
				sp := setup(t, setupName)
				return rose.CreateComponent([]*klimt.Style{seq(sp, klimt.SNameGroup), seq(sp, klimt.SNameGroupHeader)},
					ht, nil, sp, klimt.DisplayCreate(h...))
			}, areaWide)
		}
		et := klimt.ComponentTypeGroupingElseLegacy
		if teoz {
			et = klimt.ComponentTypeGroupingElseTeoz
		}
		for _, comment := range []string{"x <= 5", "", "<b>bold</b> case"} {
			comp(out, setupName+"/"+et.String()+"/"+comment, func() klimt.Component {
				sp := setup(t, setupName)
				return rose.CreateComponent([]*klimt.Style{seq(sp, klimt.SNameGroup)}, et, nil, sp,
					klimt.DisplayCreate(comment))
			}, areaWide)
		}
	}
	comp(out, setupName+"/GROUPING_SPACE", func() klimt.Component {
		return rose.CreateComponent(nil, klimt.ComponentTypeGroupingSpace, nil, setup(t, setupName), nil)
	}, areaWide)

	for _, typ := range []klimt.ComponentType{klimt.ComponentTypeActivationBoxCloseClose,
		klimt.ComponentTypeActivationBoxCloseOpen, klimt.ComponentTypeActivationBoxOpenClose,
		klimt.ComponentTypeActivationBoxOpenOpen} {
		for _, withDisplay := range []bool{false, true} {
			comp(out, setupName+"/"+typ.String()+"/"+strconv.FormatBool(withDisplay), func() klimt.Component {
				sp := setup(t, setupName)
				var d *klimt.Display
				if withDisplay {
					d = klimt.DisplayCreate("Alice")
				}
				return rose.CreateComponent([]*klimt.Style{seq(sp, klimt.SNameActivationBox)}, typ, nil, sp, d)
			}, []area{{}, {0, 40}, {15, 33}})
		}
	}

	for _, withDisplay := range []bool{false, true} {
		comp(out, setupName+"/PARTICIPANT_LINE/"+strconv.FormatBool(withDisplay), func() klimt.Component {
			sp := setup(t, setupName)
			var d *klimt.Display
			if withDisplay {
				d = klimt.DisplayCreate("Alice")
			}
			return rose.CreateComponent([]*klimt.Style{seq(sp, klimt.SNameLifeLine)},
				klimt.ComponentTypeParticipantLine, nil, sp, d)
		}, []area{{}, {20, 30}, {9, -20}})
	}
	comp(out, setupName+"/DELAY_LINE", func() klimt.Component {
		sp := setup(t, setupName)
		return rose.CreateComponent([]*klimt.Style{seq(sp, klimt.SNameLifeLine, klimt.SNameDelay)},
			klimt.ComponentTypeDelayLine, nil, sp, nil)
	}, areaLine)
	for _, txt := range []string{"5 minutes later", `two\nlines`} {
		comp(out, setupName+"/DELAY_TEXT/"+txt, func() klimt.Component {
			sp := setup(t, setupName)
			return rose.CreateComponent([]*klimt.Style{seq(sp, klimt.SNameLifeLine, klimt.SNameDelay)},
				klimt.ComponentTypeDelayText, nil, sp, klimt.DisplayWithNewlines(klimt.PragmaCreateEmpty(), txt))
		}, areaWide)
	}
	comp(out, setupName+"/DESTROY", func() klimt.Component {
		sp := setup(t, setupName)
		return rose.CreateComponent(
			[]*klimt.Style{klimt.ComponentTypeDestroy.StyleSignature().MergedStyle(sp.CurrentStyleBuilder())},
			klimt.ComponentTypeDestroy, nil, sp, nil)
	}, areaWide)
	comp(out, setupName+"/NEWPAGE", func() klimt.Component {
		sp := setup(t, setupName)
		return rose.CreateComponentNewPage([]*klimt.Style{seq(sp, klimt.SNameNewpage)}, sp)
	}, areaWide)

	for _, txt := range []string{"", "Initialization", `two\nlines`} {
		comp(out, setupName+"/DIVIDER/"+txt, func() klimt.Component {
			sp := setup(t, setupName)
			return rose.CreateComponent([]*klimt.Style{seq(sp, klimt.SNameSeparator)}, klimt.ComponentTypeDivider,
				nil, sp, klimt.DisplayWithNewlines(klimt.PragmaCreateEmpty(), txt))
		}, areaWide)
	}
	refs := [][]string{{"ref", "Init"}, {"ref over A, B", "first line", "<b>second</b> line"}, {"ref"}}
	for i, r := range refs {
		comp(out, setupName+"/REFERENCE/"+strconv.Itoa(i), func() klimt.Component {
			sp := setup(t, setupName)
			return rose.CreateComponent([]*klimt.Style{seq(sp, klimt.SNameReference),
				seq(sp, klimt.SNameReferenceHeader)}, klimt.ComponentTypeReference, nil, sp, klimt.DisplayCreate(r...))
		}, areaWide)
	}
	for _, txt := range []string{"Box", `Multi\nline box`, ""} {
		comp(out, setupName+"/ENGLOBER/"+txt, func() klimt.Component {
			sp := setup(t, setupName)
			return rose.CreateComponent([]*klimt.Style{seq(sp, klimt.SNameBox)}, klimt.ComponentTypeEnglober, nil, sp,
				klimt.DisplayWithNewlines(klimt.PragmaCreateEmpty(), txt))
		}, areaWide)
	}
}

func alignments(out *[]string) {
	n := klimt.ArrowConfigurationWithDirectionNormal()
	configs := []*klimt.ArrowConfiguration{n, n.Reverse(), klimt.ArrowConfigurationWithDirectionBoth(),
		n.ReverseDefine(), n.Reverse().ReverseDefine()}
	for _, align := range []string{"first", "direction", "reversedirection", "center", "right", "left"} {
		for i, config := range configs {
			comp(out, "align/"+align+"/"+strconv.Itoa(i), func() klimt.Component {
				sp := newSkinParam()
				sp.SetParam("SequenceMessageAlignment", align)
				return NewRose().CreateComponentArrow([]*klimt.Style{seq(sp, klimt.SNameArrow)}, config, sp,
					display("hello"))
			}, areaArrow)
		}
	}
}

func misc(out *[]string) {
	var log []string
	rose := NewRose()
	sp := newSkinParam()
	arrow := []*klimt.Style{seq(sp, klimt.SNameArrow)}
	for _, typ := range klimt.ComponentTypeValues() {
		log = append(log, "nostyle "+typ.String())
		step(&log, func() {
			log = append(log, className(rose.CreateComponent(nil, typ, nil, sp, klimt.DisplayCreate("x"))))
		})
		log = append(log, "styles "+typ.String())
		step(&log, func() {
			log = append(log, className(rose.CreateComponent([]*klimt.Style{seq(sp, klimt.SNameGroup),
				seq(sp, klimt.SNameGroupHeader)}, typ, klimt.ArrowConfigurationWithDirectionNormal(), sp,
				klimt.DisplayCreate("x", "y"))))
		})
		log = append(log, "note "+typ.String())
		step(&log, func() {
			log = append(log, className(rose.CreateComponentNote(arrow, typ, sp, klimt.DisplayCreate("x"),
				color.EmptyColors())))
		})
	}
	for _, typ := range klimt.ComponentTypeValues() {
		log = append(log, "participant "+typ.String())
		step(&log, func() {
			p := &testParticipant{typ: ptParticipant, colors: color.EmptyColors(), styleBuilder: sp.CurrentStyleBuilder()}
			log = append(log, className(rose.CreateComponentParticipant(p, typ, nil, sp, klimt.DisplayCreate("A"))))
		})
	}
	step(&log, func() {
		log = append(log, "participant line "+className(rose.CreateComponent(arrow, klimt.ComponentTypeParticipantLine,
			nil, sp, nil)))
	})
	sp2 := newSkinParam()
	sp2.SetParam("sequenceLifeLineBorderThickness", "2.5")
	for _, p := range []*klimt.SkinParam{sp, sp2} {
		for _, lp := range []klimt.LineParam{klimt.LineParamSequenceLifeLineBorder, klimt.LineParamSequenceArrow} {
			log = append(log, "stroke "+lp.String())
			step(&log, func() { log = append(log, Stroke(p, lp, 1.5).String()) })
		}
	}
	step(&log, func() { log = append(log, "color "+hc(rose.HtmlColor(sp, klimt.ColorParamBackground))) })
	step(&log, func() { log = append(log, "color "+hc(rose.HtmlColor(sp, klimt.ColorParamSequenceLifeLineBorder))) })
	step(&log, func() {
		log = append(log, "color "+hc(rose.HtmlColorStereo(sp2, nil, klimt.ColorParamNoteBackground,
			klimt.ColorParamArrowHead)))
	})
	step(&log, func() { log = append(log, "color "+hc(rose.HtmlColorStereo(sp2, nil))) })
	for _, pos := range []string{"left", "Over", "over_several", "bottom", "", "middle"} {
		label := pos
		if pos == "" {
			label = "null"
		}
		log = append(log, "noteposition "+label)
		step(&log, func() { log = append(log, NotePositionDefaultLeft(pos, pos != "").String()) })
	}
	for _, l := range log {
		*out = append(*out, "misc\t"+l)
	}
}

func TestGoldenRose(t *testing.T) {
	var out []string
	for _, setupName := range setups {
		participants(t, &out, setupName)
		arrows(t, &out, setupName)
		notes(t, &out, setupName)
		others(t, &out, setupName)
	}
	alignments(&out)
	misc(&out)
	compareGolden(t, "rose.golden.gz", out)
}
