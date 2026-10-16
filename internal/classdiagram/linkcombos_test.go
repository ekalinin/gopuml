// SPDX-License-Identifier: MIT
// Provenance: original

package classdiagram

import "strconv"

// linkComboCases are generated relations: every pair of heads, the bodies
// with directions, styles and middle circles, and the end texts.
func linkComboCases() []testCase {
	heads1 := []string{"", "<|", "^", "*", "o", "x", "<||", "<|:", "}", "}o", "|o", "||", "}|", "<", "<_", "<<",
		"0", "@", "0)", ")", "#", "+", "o<", "<*"}
	heads2 := []string{"", "|>", "^", "*", "o", "x", "||>", ":|>", "{", "o{", "o|", "||", "|{", ">", "_>", ">>",
		"0", "@", "(0", "(", "#", "+", `\\`, "//", ">o"}
	var cases []testCase
	add := func(src string) {
		cases = append(cases, testCase{name: "linkcombos/" + strconv.Itoa(len(cases)) + " " + src, source: src + "\n"})
	}
	for _, h1 := range heads1 {
		for _, h2 := range heads2 {
			add("A " + h1 + "--" + h2 + " B")
		}
	}
	bodies := []string{"-", "--", "---", ".", "..", "...", "=", "==", "-.-", ".-.", "-up-", "-u-", "-do-", "-down-",
		"-le-", "-l-", "-left-", "-ri-", "-right-", "-r-", ".up.", ".left.", "-[#red]-", "-[dashed]-",
		"-[#red,dashed]-", "-[bold]-", "-[dotted]-", "-[hidden]-", "-[thickness=3]-", "-[#red;#blue]-",
		"-[norank]-", "-[single]-", "-[plain]-", "-[node]-", "-up[#red]-", "-[#red]up-", "-0-", "-(0-", "-0)-",
		"-(0)-", "--0--", ".0.", "-0[#red]-"}
	for _, body := range bodies {
		for _, heads := range [][2]string{{"", ""}, {"", ">"}, {"<|", ""}, {"*", ""}, {"o", ">"}, {"<", ">"}} {
			add("A " + heads[0] + body + heads[1] + " B")
		}
	}
	for _, src := range []string{
		"A --0 B", "A --(0 B", "A 0)-- B", "A 0-- B", "A -0 B", "A --0) B", "A -0~B", "A -0.B", "A --o B",
		"A--B", "A-->B", "Ao--B", "A o-- B", "A --o B", "A --> B : label", "A --> B : label >", "A --> B : < label",
		"A --> B : >", "A --> B : <", `A "1" --> "many" B`, `A "1" --> "many" B : contains`,
		`A --> B : "1" label "2"`, `A --> B : "1" label`, `A --> B : label "2"`, `A "x" --> B : "1" label "2"`,
		"A /role1 --> /role2 B", `A /"role 1" --> B`, "A [qual] --> B", "A --> [q2] B", "A [q1] --> [q2] B",
		"@2 A --> B", "@1.5 A --> B", "@1.2.3 A --> B", "A --> B #red", "A --> B #red;line.dashed",
		"A --> B [[http://x]]", "A --> B <<st>>", "A --> B <<st>> : label", "A --> B #red [[http://x]] <<s>> : l",
		`"A x" --> "B y"`, "a.b.C --> d.E", ".A --> B", "A::p1 --> B::p2", "A::p1 --> B", "A -up-> A",
		"A -left-> B : l", "A -up-> B", "A <-left- B", "A <|-up- B", "A ..|> B", "A ..> B", "A --* B", "A *--* B",
		"A x--x B", "A }o--o{ B", "A ||--|| B", "A |o--o| B", "A }|--|{ B", "A +-- B", "A #--# B", "A ^--^ B",
		"A <|--|> B", "A <||--||> B", "A <|:--:|> B", "A @--@ B", "A 0--0 B", "A )--( B", `A \\-- B`,
		"A --// B", "A -[#red]-> B", "A -[#red]up-> B", "A -up[#red]-> B", "A -[#red]-[#blue]-> B",
		"A ->> B", "A <<- B", "A <_-- B", "A --_> B", "A =[#red]=> B",
	} {
		add(src)
	}
	return cases
}

// declComboCases are generated declarations: the element keywords with
// names, generics, stereotypes, tags, links, colors and parents, the
// notes, and the hide and show commands.
func declComboCases() []testCase {
	var cases []testCase
	add := func(src string) {
		cases = append(cases, testCase{name: "declcombos/" + strconv.Itoa(len(cases)) + " " + firstLine(src),
			source: src + "\n"})
	}
	keywords := []string{"class", "abstract class", "abstract", "interface", "enum", "annotation", "entity",
		"circle", "diamond", "protocol", "struct", "exception", "metaclass", "stereotype", "dataclass", "record",
		"map", "static class", "object", "json"}
	names := []string{"A", `"A b"`, `"Disp" as A`, `A as "Disp"`, "a.b.C", "A<T>", `"G<K>" as G`, "A_1",
		"Été"}
	tails := []string{"", " <<st>>", " << (S,#FF7700) S >>", " $tag", " #red", " ##[dashed]blue",
		" [[http://x]]", " extends B", " implements I, J", " {}", " <<st>> #red [[http://x]] ##blue",
		" $a <<s>> $b", " #back:red;line:blue"}
	for _, kw := range keywords {
		for _, name := range names {
			add(kw + " " + name)
		}
		for _, tail := range tails {
			add(kw + " Name" + tail)
		}
		add(kw + " Name {\n  -field : int\n  +method()\n}")
		add("-" + kw + " Name")
	}
	for _, portion := range []string{"members", "member", "attributes", "attribute", "fields", "field", "methods",
		"method", "circle", "circles", "circled", "stereotype", "stereotypes"} {
		for _, gender := range []string{"", "class ", "interface ", "enum ", "abstract ", "<<st>> ", "A ",
			`"A" `, "empty ", "class empty ", "A empty "} {
			add("class A <<st>> {\n  f\n  m()\n}\ninterface I\nhide " + gender + portion)
			add("class A\nshow " + gender + portion)
		}
	}
	for _, visibility := range []string{"private", "public", "protected", "package", "private,public",
		"private public", "private, protected"} {
		for _, portion := range []string{"members", "fields", "methods", "attributes"} {
			add("class A {\n  -a\n  +b\n  #c\n  ~d\n  -e()\n  +f()\n  #g()\n  ~h()\n}\nhide " + visibility + " " +
				portion)
		}
	}
	for _, pos := range []string{"left", "right", "top", "bottom"} {
		for _, tail := range []string{" : text", " <<st>> : text", " #red : text", " $t : text",
			" [[http://x]] : text", "\n  multi\nend note", " {\n  bracket\n}", " #pink\n  multi\nendnote"} {
			add("class A\nnote " + pos + " of A" + tail)
			add("class A\nnote " + pos + tail)
		}
	}
	return cases
}

func firstLine(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			return s[:i] + " ..."
		}
	}
	return s
}
