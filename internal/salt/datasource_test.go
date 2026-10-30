// SPDX-License-Identifier: MIT
// Provenance: original

package salt

import (
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/klimt/color"
)

// TestDataSourceImpl checks the tokens and terminators against upstream's
// DataSourceImpl on the same lines.
func TestDataSourceImpl(t *testing.T) {
	tests := []struct {
		lines []string
		want  string
	}{
		{[]string{`Login | "me"`}, `[Login NEWCOL] ["me" NEWLINE]`},
		{[]string{`{# a | b }`}, `[{# NEWCOL] [a NEWCOL] [b NEWCOL] [} NEWLINE]`},
		{[]string{`{T#`}, `[{ NEWCOL] [T# NEWLINE]`},
		{[]string{`{SI`}, `[{SI NEWLINE]`},
		{[]string{`{S-`}, `[{S- NEWLINE]`},
		{[]string{`{S`}, `[{S NEWLINE]`},
		// The pattern is a Pattern2, so case-insensitive.
		{[]string{`{s`}, `[{s NEWLINE]`},
		{[]string{`{* File | Edit`}, `[{* NEWCOL] [File NEWCOL] [Edit NEWLINE]`},
		{[]string{`[[http://x.com\{tooltip\}]]`}, `[[[http://x.com{tooltip}]] NEWLINE]`},
		{[]string{`3\|4`}, `[3|4 NEWLINE]`},
		{[]string{`a |  | b`}, `[a NEWCOL] [b NEWLINE]`},
		{[]string{`{`, `Name | "text"`, `}`}, `[{ NEWLINE] [Name NEWCOL] ["text" NEWLINE] [} NEWLINE]`},
		{[]string{`{ (X) public | () default }`}, `[{ NEWCOL] [(X) public NEWCOL] [() default NEWCOL] [} NEWLINE]`},
		{[]string{`N={ [Top] }`}, `[N= NEWCOL] [{ NEWCOL] [[Top] NEWCOL] [} NEWLINE]`},
	}
	for _, tt := range tests {
		d := newDataSourceImpl(tt.lines)
		var got []string
		for d.hasNext() {
			n := d.next()
			got = append(got, "["+n.element+" "+n.terminator.String()+"]")
		}
		if s := strings.Join(got, " "); s != tt.want {
			t.Errorf("%q: tokens %s, want %s", tt.lines, s, tt.want)
		}
	}
}

// TestDataSourcePeekPastTheEnd checks the IndexOutOfBoundsException that an
// unclosed group reaches.
func TestDataSourcePeekPastTheEnd(t *testing.T) {
	d := newDataSourceImpl([]string{"{", "[OK]"})
	defer func() {
		r := recover()
		e, ok := r.(*color.RuntimeError)
		if !ok || e.Class != "java.lang.IndexOutOfBoundsException" || e.Message != "Index 2 out of bounds for length 2" {
			t.Errorf("recovered %#v, want the IndexOutOfBoundsException of upstream", r)
		}
	}()
	d.peek(2)
}

func TestPositionner2(t *testing.T) {
	a, b, c := &elementEmpty{}, &elementEmpty{}, &elementEmpty{}
	var p positionner2
	p.add(terminated[element]{a, newCol})
	p.add(terminated[element]{b, newLine})
	p.add(terminated[element]{c, newCol})
	// A "*" cell: c spans its own column and the next one.
	p.mergeLeft(newLine)
	all := p.all()
	if len(all) != 3 {
		t.Fatalf("elements %v, want a, b and c", all)
	}
	_, _, _ = a, b, c
	if got := all[0].cell.String() + all[1].cell.String() + all[2].cell.String(); got != "(0,0)-(0,0)(0,1)-(0,1)(1,0)-(1,1)" {
		t.Errorf("cells %s", got)
	}
	// mergeLeft updates the maximum before it moves, as upstream: the row
	// after the "*" is not counted until an element is added to it.
	if p.nbRows() != 2 || p.nbCols() != 2 {
		t.Errorf("%d rows and %d columns, want 2 and 2", p.nbRows(), p.nbCols())
	}
}

func TestPositionner2MergeWithoutCell(t *testing.T) {
	var p positionner2
	defer func() {
		e, ok := recover().(*color.RuntimeError)
		if !ok || e.Class != "java.lang.NullPointerException" {
			t.Errorf("recovered %v, want a NullPointerException", e)
		}
	}()
	p.mergeLeft(newCol)
}
