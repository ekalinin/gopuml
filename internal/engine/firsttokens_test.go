// SPDX-License-Identifier: MIT
// Provenance: original

package engine

import (
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/command"
)

// firstTokensEntry is a line of internal/command/testdata/first_tokens.golden,
// written by FirstTokensGolden.java from the upstream jar.
type firstTokensEntry struct {
	factory, class, pattern string
	fast, slow              []string // nil for any line
}

func readFirstTokensGolden(t *testing.T) []firstTokensEntry {
	t.Helper()
	data, err := os.ReadFile("../command/testdata/first_tokens.golden")
	if err != nil {
		t.Fatal(err)
	}
	tokens := func(s string) []string {
		if s == "ANYTHING" {
			return nil
		}
		result := []string{}
		for _, tok := range strings.Fields(s) {
			if tok == "<empty>" {
				tok = ""
			}
			result = append(result, unescapeGolden(tok))
		}
		slices.Sort(result)
		return result
	}
	var result []firstTokensEntry
	for line := range strings.SplitSeq(strings.TrimRight(string(data), "\n"), "\n") {
		f := strings.Split(line, "\t")
		e := firstTokensEntry{factory: f[0], class: f[2], pattern: unescapeGolden(f[3]), fast: tokens(f[4]),
			slow: tokens(f[5])}
		if f[3] == "-" {
			e.pattern = ""
		}
		result = append(result, e)
	}
	return result
}

// javaDotClass is how the commands write Java's ".".
const javaDotClass = `[^\n\r\x{85}\x{2028}\x{2029}]`

// unescapeGolden decodes the \uXXXX escapes of the golden file.
func unescapeGolden(s string) string {
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+5 < len(s) && s[i+1] == 'u' {
			v, err := strconv.ParseUint(s[i+2:i+6], 16, 32)
			if err == nil {
				sb.WriteRune(rune(v))
				i += 5
				continue
			}
		}
		sb.WriteByte(s[i])
	}
	return sb.String()
}

// Every command of the factories of the supported diagram types declares the
// first tokens of its upstream command: the command of the same factory with
// the same starting pattern, or, for a command without one, one of the
// upstream commands without pattern that declare the same tokens.
func TestFirstTokensMatchUpstream(t *testing.T) {
	golden := readFirstTokensGolden(t)
	names := []string{"SequenceDiagramFactory", "ClassDiagramFactory", "ActivityDiagramFactory",
		"DescriptionDiagramFactory", "StateDiagramFactory", "ActivityDiagramFactory3", "MindMapDiagramFactory",
		"WBSDiagramFactory",
		"PacketDiagramFactory", "GanttDiagramFactory", "TimingDiagramFactory"}
	factories := Factories[1:]
	if len(factories) != len(names) {
		t.Fatalf("%d factories, want %d", len(factories), len(names))
	}
	same := func(a, b []string) bool { return (a == nil) == (b == nil) && slices.Equal(a, b) }
	for i, f := range factories {
		infos := f.(interface{ CommandInfos() []command.CommandInfo }).CommandInfos()
		var upstream []firstTokensEntry
		for _, e := range golden {
			if e.factory == names[i] {
				upstream = append(upstream, e)
			}
		}
		for j, info := range infos {
			tokens := info.Tokens
			if tokens != nil {
				tokens = slices.Clone(tokens)
				slices.Sort(tokens)
			}
			pattern := strings.ReplaceAll(info.Pattern, javaDotClass, ".")
			k := slices.IndexFunc(upstream, func(e firstTokensEntry) bool { return pattern != "" && e.pattern == pattern })
			if k < 0 {
				// A command without pattern, or whose pattern is written
				// differently: one of the upstream commands without pattern, or
				// one whose pattern matches nothing here.
				k = slices.IndexFunc(upstream, func(e firstTokensEntry) bool { return same(e.fast, tokens) })
				t.Logf("%s command %d: pattern %q matched by its tokens %q", names[i], j, pattern, tokens)
			}
			if k < 0 || !same(upstream[k].fast, tokens) {
				t.Errorf("%s command %d: pattern %q, tokens %q: no upstream command declares them", names[i], j,
					pattern, tokens)
			}
		}
	}
}
