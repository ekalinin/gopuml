// SPDX-License-Identifier: MIT
// Provenance: original

package abel_test

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/classdiagram"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/security"
	"github.com/ekalinin/gopuml/internal/tim"
	"github.com/ekalinin/gopuml/internal/uml"
)

// The golden file compared here is written by
// internal/abel/testdata/javagen/BodyGolden.java from the cases of
// testdata/body_inputs.txt.

type bodyCase struct {
	name, source string
}

func readBodyCases(t *testing.T) []bodyCase {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "body_inputs.txt"))
	if err != nil {
		t.Fatal(err)
	}
	var cases []bodyCase
	var cur *bodyCase
	var sb strings.Builder
	lines := strings.Split(string(data), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	for _, line := range lines {
		if name, ok := strings.CutPrefix(line, "=== "); ok {
			if cur != nil {
				cur.source = sb.String()
				cases = append(cases, *cur)
			}
			cur = &bodyCase{name: name}
			sb.Reset()
			continue
		}
		sb.WriteString(line + "\n")
	}
	if cur != nil {
		cur.source = sb.String()
		cases = append(cases, *cur)
	}
	return cases
}

// parseBody parses a source as a class diagram, as SourceStringReader does
// for the cases.
func parseBody(t *testing.T, source string) core.Diagram {
	t.Helper()
	source = "@startuml\n" + source + "@enduml\n"
	rc := core.NewRenderContext()
	blocks := core.BuildBlocks(source, rc.Desc, rc.Config)
	access := &security.Access{}
	access.Config.Profile, _ = security.ParseProfile(rc.SecurityProfile)
	res, err := tim.PreprocessWith(rc, access, blocks[0].Lines,
		tim.Options{Definitions: &tim.BlockDefinitions{}, Defines: tim.NewDefines()})
	if err != nil {
		t.Fatal(err)
	}
	src := core.NewUmlSource(res.Lines, blocks[0].Lines, true)
	return classdiagram.Factory.CreateSystem(rc, src)
}

func rectString(r *geom.XRectangle2D) string {
	if r == nil {
		return "null"
	}
	return d2s(r.MinX()) + "," + d2s(r.MinY()) + "," + d2s(r.Width) + "," + d2s(r.Height)
}

func TestGoldenBodies(t *testing.T) {
	var out []string
	for _, c := range readBodyCases(t) {
		out = append(out, "=== "+c.name)
		d := parseBody(t, c.source)
		cd, ok := d.(abel.Diagram)
		if !ok {
			out = append(out, "not a CucaDiagram: "+c.name)
			continue
		}
		cuca := cd.Cuca()
		sp := cuca.SkinParam()
		for _, leaf := range cuca.Leafs() {
			lt := leaf.LeafType()
			var style *klimt.Style
			var sm, sf bool
			var fc *klimt.FontConfiguration
			switch {
			case lt.IsLikeClass():
				style = klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameClassDiagram,
					klimt.SNameClass).WithTOBECHANGED(leaf.Stereotype()).With(leaf.Stereostyles()).
					MergedStyle(leaf.CurrentStyleBuilder())
				sm = cuca.ShowPortion(uml.PortionMethod, leaf)
				sf = cuca.ShowPortion(uml.PortionField, leaf)
			case lt == abel.LeafTypeObject:
				style = klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameObjectDiagram,
					klimt.SNameObject).WithTOBECHANGED(leaf.Stereotype()).MergedStyle(sp.CurrentStyleBuilder())
				sf = cuca.ShowPortion(uml.PortionField, leaf)
			case lt == abel.LeafTypeMap || lt == abel.LeafTypeJSON:
				name := klimt.SNameJson
				if lt == abel.LeafTypeMap {
					name = klimt.SNameMap
				}
				style = klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameObjectDiagram,
					name).WithTOBECHANGED(leaf.Stereotype()).MergedStyle(sp.CurrentStyleBuilder())
				fc = style.FontConfiguration(sp.IHtmlColorSet(), nil)
			default:
				style = klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameClassDiagram,
					klimt.SNameClass).WithTOBECHANGED(leaf.Stereotype()).MergedStyle(leaf.CurrentStyleBuilder())
				sm, sf = true, true
			}
			out = append(out, leaf.Name()+" "+lt.String()+" sm="+strconv.FormatBool(sm)+" sf="+
				strconv.FormatBool(sf))
			for _, b := range bounders {
				var log []string
				step(&log, func() {
					tb := leaf.Bodier().Body(sp, sm, sf, leaf.Stereotype(), style, fc)
					if tb == nil {
						log = append(log, "null")
						return
					}
					dim := tb.CalculateDimension(b.bounder)
					log = append(log, "dim "+d2s(dim.Width())+"x"+d2s(dim.Height()))
					switch body := tb.(type) {
					case *abel.TextBlockMap:
						body.SetTotalWidth(dim.Width() + 20)
					case *abel.TextBlockCucaJSon:
						body.SetTotalWidth(dim.Width() + 20)
					}
					if wp, ok := tb.(klimt.WithPorts); ok {
						log = append(log, "ports "+wp.Ports(b.bounder).String())
					}
					if lt.IsLikeClass() || lt == abel.LeafTypeObject {
						var members []string
						for _, cs := range leaf.Bodier().RawBody() {
							members = append(members, cs.String())
						}
						members = append(members, "zzz")
						for _, m := range members {
							step(&log, func() {
								log = append(log, "inner '"+m+"' "+rectString(tb.InnerPosition(m, b.bounder)))
							})
						}
					}
					tb.DrawU(newRec(&log, b.bounder).Apply(klimt.NewUTranslate(3, 4)))
				})
				for _, l := range log {
					out = append(out, c.name+"\t"+b.name+"\t"+l)
				}
			}
		}
		var done *abel.LinkConstraint
		for _, link := range cuca.Links() {
			lc := link.LinkConstraint()
			if lc == nil || lc == done {
				continue
			}
			done = lc
			for _, b := range bounders {
				var log []string
				step(&log, func() {
					var l1, l2 *abel.Link
					for _, l := range cuca.Links() {
						if l.LinkConstraint() == lc {
							if l1 == nil {
								l1 = l
							} else {
								l2 = l
							}
						}
					}
					lc.SetPosition(l1, geom.NewXPoint2D(10, 20))
					lc.SetPosition(l2, geom.NewXPoint2D(60, 90))
					lc.DrawMe(newRec(&log, b.bounder), sp)
				})
				for _, l := range log {
					out = append(out, c.name+"\tconstraint "+b.name+"\t"+l)
				}
			}
		}
	}
	compareGolden(t, "body.golden", out)
}
