// SPDX-License-Identifier: MIT
// Provenance: original

package core

import (
	"slices"
	"testing"

	"github.com/ekalinin/plantuml-in-go/internal/text"
)

func lines(b *Block) []string {
	var out []string
	for _, l := range b.Lines {
		out = append(out, l.S)
	}
	return out
}

func positions(b *Block) []int {
	var out []int
	for _, l := range b.Lines {
		out = append(out, l.Location.Position)
	}
	return out
}

func TestBuildBlocks(t *testing.T) {
	tests := []struct {
		name   string
		src    string
		config []string
		want   [][]string
	}{
		{
			name: "two blocks in one input",
			src:  "@startuml\nAlice -> Bob\n@enduml\ntext between\n@startuml\nBob -> Carol\n@enduml\n",
			want: [][]string{{"@startuml", "Alice -> Bob", "@enduml"}, {"@startuml", "Bob -> Carol", "@enduml"}},
		},
		{
			name: "diagram embedded in a Java comment",
			src:  "/**\n * @startuml\n * Alice -> Bob\n * @enduml\n */\nclass Foo {}\n",
			// The prefix is " *": the whitespace before @startuml is kept.
			want: [][]string{{" @startuml", " Alice -> Bob", " @enduml"}},
		},
		{
			name: "no block",
			src:  "just text\nAlice -> Bob\n",
			want: nil,
		},
		{
			name: "backslash tags",
			src:  "\\startuml\nA -> B\n\\enduml\n",
			want: [][]string{{"\\startuml", "A -> B", "\\enduml"}},
		},
		{
			name: "upper-case start keyword is not a directive",
			src:  "@STARTUML\nA -> B\n@enduml\n",
			want: nil,
		},
		{
			name: "start keyword needs a following character",
			src:  "@start\nA\n@end\n",
			want: nil,
		},
		{
			name: "unterminated block is dropped",
			src:  "@startuml\nA -> B\n",
			want: nil,
		},
		{
			name: "yaml header after the start line",
			src:  "@startuml\n---\ntitle: x\n---\nA -> B\n@enduml\n",
			want: [][]string{{"@startuml", "A -> B", "@enduml"}},
		},
		{
			name: "unterminated yaml header is kept",
			src:  "@startuml\n---\nA -> B\n@enduml\n",
			want: [][]string{{"@startuml", "---", "A -> B", "@enduml"}},
		},
		{
			name: "backslash continuation skips comments",
			src:  "@startuml\nA -> \\\n' comment\nB\n@enduml\n",
			want: [][]string{{"@startuml", "A -> B", "@enduml"}},
		},
		{
			name: "double backslash is not a continuation",
			src:  "@startuml\nA -> B : x\\\\\nC -> D\n@enduml\n",
			want: [][]string{{"@startuml", "A -> B : x\\\\", "C -> D", "@enduml"}},
		},
		{
			// The MIT variant does not know ditaa, so the backslash merges.
			name: "ditaa lines merge in the MIT variant",
			src:  "@startditaa\n+--\\\n|  |\n@endditaa\n",
			want: [][]string{{"@startditaa", "+--|  |", "@endditaa"}},
		},
		{
			name:   "config lines after the start line",
			src:    "@startuml\nA -> B\n@enduml\n",
			config: []string{"skinparam monochrome true", "' a comment"},
			want:   [][]string{{"@startuml", "skinparam monochrome true", "A -> B", "@enduml"}},
		},
		{
			name: "pause and append",
			src:  "@startuml\nA -> B\n@pause\nignored\n@append C -> D\n@unpause\nE -> F\n@enduml\n",
			// The @unpause line itself is read while paused and is dropped.
			want: [][]string{{"@startuml", "A -> B", "C -> D", "E -> F", "@enduml"}},
		},
		{
			name: "exit truncates the block",
			src:  "@startuml\nA -> B\n!exit\nC -> D\n@enduml\n",
			want: [][]string{{"@startuml", "A -> B", "@enduml"}},
		},
		{
			name: "byte order mark, en dash and CRLF",
			src:  "\uFEFF@startuml\r\nA \u2013> B\r\n@enduml\r\n",
			want: [][]string{{"@startuml", "A -> B", "@enduml"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocks := BuildBlocks(tt.src, "string", tt.config)
			if len(blocks) != len(tt.want) {
				t.Fatalf("got %d blocks, want %d", len(blocks), len(tt.want))
			}
			for i, b := range blocks {
				if got := lines(b); !slices.Equal(got, tt.want[i]) {
					t.Errorf("block %d = %q, want %q", i, got, tt.want[i])
				}
			}
		})
	}
}

func TestBuildBlocksLocations(t *testing.T) {
	blocks := BuildBlocks("header\n@startuml\nA -> \\\nB\nC\n@enduml\n", "f.puml", []string{"x"})
	if len(blocks) != 1 {
		t.Fatal("want one block")
	}
	// Config lines share the start line's location; merged lines keep the
	// location of their first line.
	if got, want := positions(blocks[0]), []int{1, 1, 2, 4, 5}; !slices.Equal(got, want) {
		t.Errorf("positions = %v, want %v", got, want)
	}
	if blocks[0].Lines[0].Location.Desc != "f.puml" {
		t.Errorf("desc = %q", blocks[0].Lines[0].Location.Desc)
	}
}

func TestBlockName(t *testing.T) {
	tests := []struct {
		start string
		name  string
		ok    bool
	}{
		{"@startuml", "", false},
		{"@startuml login-flow", "login-flow", true},
		{"@startuml(id=foo)", "", false},
		{`@startuml "my diagram"`, "my diagram", true},
		{"@startuml{braced}", "braced", true},
		{"@startuml {braced}", "{braced", true},
		{"@startuml diagram.png", "diagram", true},
		{"@startuml a,b", "a", true},
		{"@startuml file://out/x", "out/x", true},
		{"@startuml <b>bold</b>", "", false},
		{"  @startmindmap   map1  ", "map1", true},
		{"@STARTUML upper", "upper", true},
	}
	for _, tt := range tests {
		b := &Block{Lines: []text.StringLocated{text.NewStringLocated(tt.start, nil)}}
		name, ok := b.Name()
		if name != tt.name || ok != tt.ok {
			t.Errorf("%q: name %q ok %v, want %q %v", tt.start, name, ok, tt.name, tt.ok)
		}
	}
}
