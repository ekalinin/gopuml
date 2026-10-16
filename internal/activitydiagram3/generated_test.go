// SPDX-License-Identifier: MIT
// Provenance: original

package activitydiagram3

import (
	"strconv"
	"strings"
)

// The generated cases combine the variants of the if, while, repeat, fork,
// split and switch syntaxes. They are dumped compactly (see isCompact).

var (
	ifHeads = []string{
		"if (c) then (yes)",
		"if (c) then",
		"if (c)",
		"if (c) then (yes);",
		"if (c) is (v) then",
		"if (c) equals (v) then;",
		"if (c) equal (v) then <<#red>>",
		"#red:if (c) then (yes)",
		"#red:if (c) is (v) then",
		"if (c) then (yes) <<#blue>>",
		"if (c) then (yes) <<#blue>> <<##green>>",
		"[[http://x.org]] if (c) then (yes)",
		"[[http://x.org{tip}]] #red:if (c) then (yes)",
		"if <<s>> (c) then",
		"if (c) then when yes",
		"if ()",
		"IF (C) THEN (YES)",
		"if(c)then(yes)",
		"if (a\\nb) then (yes\\nsure)",
		"if (c) then (-> yes)",
	}
	ifElses = []string{
		"",
		"else",
		"else (no)",
		"else (-> no)",
		"else (-[#red]-> no)",
		"else (-[#red,dashed]->)",
		"else when no",
		"else;",
		"elseif (d) then (z)",
		"(in) elseif (d) then (z)",
		"(-[#green]-> in) elseif (d) then (-[#blue]-> z)",
		"elseif (d) is (v) then",
		"(in) elseif (d) is (v) then (->)",
		"#pink:elseif (d) is (v) then",
		"elseif (d) then (z) <<#yellow>>",
		"else if (d) then (z)",
		"elseif (d) then (z)\n:B;\nelse (no)",
		"elseif (d)\n:B;\nelseif (e)\n:C;\nelse",
	}
	ifEnds = []string{"endif", "end if", "endif;", "endif <<#pink>>", "endif <<#pink>> <<#orange>>"}
	bodies = []string{
		":A;",
		"",
		":A;\nkill",
		":A;\ndetach",
		"stop",
		"end",
		"break",
		":A;\nnote right: n",
		"note left: n",
		"(A)",
		"-> lbl;\n:A;",
		":A;\n-[#red]->",
		"label l\ngoto l",
	}
	whileHeads = []string{
		"while (c)",
		"while (c) is (yes)",
		"while (c) equals (yes)",
		"while (c) is (yes) <<#red>>",
		"#pink:while (c)",
		"while (c);",
		"while ()",
		"while (c)is(yes)",
	}
	whileEnds   = []string{"endwhile", "endwhile (no)", "end while", "while end (x)", "endwhile;", "endwhile (a\\nb)"}
	afters      = []string{"", "stop", "end", "-> lbl;", "-[#red]->", ":Z;", "kill", "note right: after"}
	repeatHeads = []string{
		"repeat",
		"repeat :start;",
		"repeat :s; <<#red>>",
		"#pink:repeat",
		"repeat <<#red>>",
		"repeat :s; <<output>>",
	}
	repeatEnds = []string{
		"repeat while (c)",
		"repeat while (c) is (yes)",
		"repeat while (c) equals (yes)",
		"repeat while (c) is (y) not (n)",
		"repeat while (c) not (n)",
		"repeat while (c) -> back;",
		"repeat while (c) -[#red]-> back;",
		"repeat while (c) is (y) -[#blue,dashed]->",
		"repeat while",
		"repeatwhile (c);",
		"repeat while (c) <<#green>>",
		"repeat while (c\nd)",
		"repeat while (c\nd) is (y)",
		"repeat while ()",
	}
	repeatBodies = []string{
		":A;",
		":A;\nbackward :B;",
		":A;\nbackward :B; <<input>>",
		":A;\nbackward :B;\nnote right: n",
		"(in) backward :B; (out)",
		":A;\nbreak",
		"",
		":A;\nkill",
	}
	forkStarts  = []string{"fork", "fork;", "fork <<#red>>"}
	forkAgains  = []string{"fork again", "fork again;", "forkagain"}
	forkEnds    = []string{"end fork", "fork end", "end merge", "end fork {and}", "endfork;", "end fork {or};", "endmerge"}
	splitStarts = []string{"split", "split;"}
	splitAgains = []string{"split again", "splitagain;"}
	splitEnds   = []string{"end split", "split end", "endsplit;"}
	switchHeads = []string{"switch (x)", "#red:switch (x)", "switch (x) <<#blue>>", "switch ()"}
	switchCases = []string{"case (a)", "case ()", "case (a) ", "CASE (A)"}
	switchEnds  = []string{"endswitch", "endswitch <<#pink>>", "endswitch;"}
	lanes       = []string{"", "|L1|\n", "|L1|\n|L2|\n"}
)

func join(lines ...string) string {
	var parts []string
	for _, l := range lines {
		if l != "" {
			parts = append(parts, l)
		}
	}
	return strings.Join(parts, "\n") + "\n"
}

// generatedCases are the generated combinations of the if, while, repeat,
// fork, split and switch syntaxes.
func generatedCases() []testCase {
	var cases []testCase
	add := func(group string, source string) {
		cases = append(cases, testCase{name: "combos/" + group + " " + strconv.Itoa(len(cases)), source: source})
	}
	for _, h := range ifHeads {
		for _, e := range ifElses {
			add("if", join(h, ":A;", e, ":E;", "endif", ":after;"))
		}
	}
	for _, end := range ifEnds[1:] {
		for _, h := range ifHeads[:4] {
			for _, e := range ifElses[:4] {
				add("if ends", join(h, ":A;", e, ":E;", end, ":after;"))
			}
		}
	}
	for _, b1 := range bodies {
		for _, b2 := range bodies {
			add("if bodies", join("if (c) then (yes)", b1, "else (no)", b2, "endif"))
		}
	}
	for _, b1 := range bodies[:8] {
		for _, b2 := range bodies[:8] {
			add("if bodies elseif", join("start", "if (c) then (yes)", b1, "elseif (d) then", b2, "endif", "stop"))
		}
	}
	for _, h := range whileHeads {
		for _, end := range whileEnds {
			for _, after := range afters {
				add("while", join(h, ":A;", end, after))
			}
		}
	}
	for _, b := range bodies {
		for _, after := range afters {
			add("while bodies", join("while (c) is (yes)", b, "endwhile (no)", after))
		}
	}
	for _, h := range repeatHeads {
		for _, end := range repeatEnds {
			add("repeat", join(h, ":A;", end, ":after;"))
		}
	}
	for _, b := range repeatBodies {
		for _, end := range repeatEnds[:5] {
			for _, after := range afters[:4] {
				add("repeat bodies", join("repeat", b, end, after))
			}
		}
	}
	for _, s := range forkStarts {
		for _, a := range forkAgains {
			for _, e := range forkEnds {
				add("fork", join(s, ":A;", a, ":B;", e, ":after;"))
			}
		}
	}
	for _, b1 := range bodies {
		for _, b2 := range bodies {
			add("fork bodies", join("fork", b1, "fork again", b2, "end fork"))
		}
	}
	for _, s := range splitStarts {
		for _, a := range splitAgains {
			for _, e := range splitEnds {
				for _, b := range bodies[:6] {
					add("split", join(s, b, a, ":B;", e))
				}
			}
		}
	}
	for _, h := range switchHeads {
		for _, c := range switchCases {
			for _, e := range switchEnds {
				for _, b := range bodies[:4] {
					add("switch", join(h, c, b, "case (b)", ":B;", e))
				}
			}
		}
	}
	nested := []string{
		join("if (c) then", "%s", "endif"),
		join("while (c)", "%s", "endwhile"),
		join("repeat", "%s", "repeat while (c)"),
		join("fork", "%s", "fork again", ":X;", "end fork"),
		join("split", "%s", "split again", ":X;", "end split"),
		join("switch (x)", "case (a)", "%s", "endswitch"),
		join("partition P {", "%s", "}"),
	}
	for _, lane := range lanes {
		for _, outer := range nested {
			for _, inner := range nested {
				body := strings.Replace(inner, "%s", ":A;\n"+strings.TrimPrefix(lane, "|L1|\n"), 1)
				add("nested", lane+strings.Replace(outer, "%s", strings.TrimSuffix(body, "\n"), 1))
			}
		}
	}
	return cases
}
