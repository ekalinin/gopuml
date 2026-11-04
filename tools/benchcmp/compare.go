// SPDX-License-Identifier: MIT
// Provenance: original

package main

import (
	"bufio"
	"fmt"
	"io"
	"maps"
	"math"
	"slices"
	"strconv"
	"strings"
)

// benchPrefix starts the lines of the sub-benchmarks of BenchmarkRender.
const benchPrefix = "BenchmarkRender/"

// defaultThreshold is the increase of the median B/op or allocs/op of a
// diagram type, in percent, above which the type is a regression.
const defaultThreshold = 5.0

// metrics are the values of one benchmark result.
type metrics struct {
	ns, bytes, allocs float64
}

// results holds the results of each diagram type, one per round.
type results map[string][]metrics

// parseResults reads the BenchmarkRender lines of the output of
// go test -bench: "BenchmarkRender/<type>[-<procs>] <n> <v> ns/op <v> B/op
// <v> allocs/op". Other lines, and lines without the three values, are
// ignored.
func parseResults(r io.Reader) (results, error) {
	out := results{}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 2 {
			continue
		}
		name, ok := strings.CutPrefix(fields[0], benchPrefix)
		if !ok {
			continue
		}
		var m metrics
		var seen int
		for i := 2; i+1 < len(fields); i += 2 {
			v, err := strconv.ParseFloat(fields[i], 64)
			if err != nil {
				continue
			}
			switch fields[i+1] {
			case "ns/op":
				m.ns, seen = v, seen|1
			case "B/op":
				m.bytes, seen = v, seen|2
			case "allocs/op":
				m.allocs, seen = v, seen|4
			}
		}
		if seen == 7 {
			typ := trimProcs(name)
			out[typ] = append(out[typ], m)
		}
	}
	return out, sc.Err()
}

// trimProcs removes the -<GOMAXPROCS> suffix that go test adds to the name
// of a benchmark.
func trimProcs(name string) string {
	i := strings.LastIndexByte(name, '-')
	if i < 0 || i == len(name)-1 {
		return name
	}
	if _, err := strconv.Atoi(name[i+1:]); err != nil {
		return name
	}
	return name[:i]
}

// summary is the median of each metric over the rounds of one side, with
// the spread of ns/op around its median, in percent.
type summary struct {
	n                 int
	ns, bytes, allocs float64
	nsLow, nsHigh     float64
}

func summarize(rounds []metrics) summary {
	if len(rounds) == 0 {
		return summary{}
	}
	pick := func(f func(metrics) float64) []float64 {
		vs := make([]float64, len(rounds))
		for i, m := range rounds {
			vs[i] = f(m)
		}
		slices.Sort(vs)
		return vs
	}
	ns := pick(func(m metrics) float64 { return m.ns })
	s := summary{
		n:      len(rounds),
		ns:     median(ns),
		bytes:  median(pick(func(m metrics) float64 { return m.bytes })),
		allocs: median(pick(func(m metrics) float64 { return m.allocs })),
	}
	s.nsLow, s.nsHigh = change(s.ns, ns[0]), change(s.ns, ns[len(ns)-1])
	return s
}

// median returns the median of sorted values: the mean of the two middle
// values for an even count.
func median(sorted []float64) float64 {
	n := len(sorted)
	if n%2 == 1 {
		return sorted[n/2]
	}
	return (sorted[n/2-1] + sorted[n/2]) / 2
}

// change returns (head - base) / base in percent, or 0 when base is 0.
func change(base, head float64) float64 {
	if base == 0 {
		return 0
	}
	return (head - base) / base * 100
}

// exceeds reports whether head is more than threshold percent above base,
// without a division, so that exactly +5% is not a regression.
func exceeds(base, head, threshold float64) bool {
	return base > 0 && (head-base)*100 > threshold*base
}

// row is the comparison of one diagram type.
type row struct {
	typ        string
	base, head summary
	regression bool
}

// compare compares the sides type by type, sorted by name. A type is a
// regression when its median B/op or allocs/op is more than threshold
// percent above that of the base; a type without a result on one side is
// never one.
func compare(base, head results, threshold float64) []row {
	types := slices.Collect(maps.Keys(head))
	for typ := range base {
		if _, ok := head[typ]; !ok {
			types = append(types, typ)
		}
	}
	slices.Sort(types)
	rows := make([]row, len(types))
	for i, typ := range types {
		r := row{typ: typ, base: summarize(base[typ]), head: summarize(head[typ])}
		if r.base.n > 0 && r.head.n > 0 {
			r.regression = exceeds(r.base.bytes, r.head.bytes, threshold) ||
				exceeds(r.base.allocs, r.head.allocs, threshold)
		}
		rows[i] = r
	}
	return rows
}

// regressions returns the types of the regressed rows.
func regressions(rows []row) []string {
	var out []string
	for _, r := range rows {
		if r.regression {
			out = append(out, r.typ)
		}
	}
	return out
}

// header describes the run above the table.
type header struct {
	base, head string
	rounds     int
	threshold  float64
}

// writeMarkdown writes the report: the header, the table, the line on the
// regressions and the notes.
func writeMarkdown(w io.Writer, h header, rows []row, notes []string) {
	fmt.Fprintf(w, "# Render benchmark\n\n")
	fmt.Fprintf(w, "Base: %s. Commit: %s. Rounds: %d. The comparison fails when the median `B/op` or `allocs/op` of a diagram type is more than %s above the base; `ns/op` does not fail it.\n\n",
		h.base, h.head, h.rounds, percent(h.threshold))
	fmt.Fprintln(w, "| Type | ns/op base | ns/op commit | ns/op change | B/op base | B/op commit | B/op change | allocs/op base | allocs/op commit | allocs/op change | |")
	fmt.Fprintln(w, "|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|")
	for _, r := range rows {
		cells := []string{r.typ,
			nsCell(r.base), nsCell(r.head), changeCell(r, r.base.ns, r.head.ns),
			valueCell(r.base, r.base.bytes), valueCell(r.head, r.head.bytes), changeCell(r, r.base.bytes, r.head.bytes),
			valueCell(r.base, r.base.allocs), valueCell(r.head, r.head.allocs), changeCell(r, r.base.allocs, r.head.allocs),
			status(r)}
		fmt.Fprintf(w, "| %s |\n", strings.Join(cells, " | "))
	}
	fmt.Fprintln(w)
	if regressed := regressions(rows); len(regressed) > 0 {
		fmt.Fprintf(w, "Allocation regressions above %s: %s.\n", percent(h.threshold), strings.Join(regressed, ", "))
	} else {
		fmt.Fprintf(w, "No allocation regression above %s.\n", percent(h.threshold))
	}
	for _, n := range notes {
		fmt.Fprintf(w, "\n%s\n", n)
	}
}

func status(r row) string {
	switch {
	case r.regression:
		return "regression"
	case r.base.n == 0:
		return "no base result"
	case r.head.n == 0:
		return "no commit result"
	}
	return ""
}

// number writes a value rounded to an integer: the median of an even count
// of rounds can end in .5, which says nothing about a render.
func number(v float64) string { return strconv.FormatFloat(math.Round(v), 'f', 0, 64) }

func percent(v float64) string { return fmt.Sprintf("%+.1f%%", v) }

func valueCell(s summary, v float64) string {
	if s.n == 0 {
		return ""
	}
	return number(v)
}

func nsCell(s summary) string {
	if s.n == 0 {
		return ""
	}
	return fmt.Sprintf("%s (%s, %s)", number(s.ns), percent(s.nsLow), percent(s.nsHigh))
}

func changeCell(r row, base, head float64) string {
	if r.base.n == 0 || r.head.n == 0 {
		return ""
	}
	return percent(change(base, head))
}
