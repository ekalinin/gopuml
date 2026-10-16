// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/regex/IRegex.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/regex/RegexComposed.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/regex/RegexConcat.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/regex/RegexLeaf.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/regex/RegexOptional.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/regex/RegexOr.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/regex/RegexRepeatedOneOrMore.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/regex/RegexRepeatedZeroOrMore.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/regex/RegexResult.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/regex/RegexPartialMatch.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package regex

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

// Group is one captured group; OK is false when the group did not take part
// in the match (null in upstream).
type Group struct {
	S  string
	OK bool
}

// Part is a piece of a composed command pattern.
type Part interface {
	// Pattern is the upstream pattern text, before macro expansion.
	Pattern() string
	// Count is the number of capturing groups in Pattern.
	Count() int
	// collect assigns captured groups to named parts, in pattern order.
	collect(groups *[]Group, r *Result)
}

// Leaf is a pattern fragment with an optional name for its groups.
type Leaf struct {
	name    string
	count   int
	pattern string
}

var (
	start = &Leaf{pattern: "^"}
	end   = &Leaf{pattern: "$"}
)

// Start is the "^" anchor.
func Start() *Leaf { return start }

// End is the "$" anchor.
func End() *Leaf { return end }

// NewLeaf returns an unnamed fragment without capturing groups.
func NewLeaf(pattern string) *Leaf { return NewNamedLeaf(0, "", pattern) }

// NewNamedLeaf returns a fragment whose count capturing groups are reported
// under name. It panics when the pattern does not have count groups, as
// upstream asserts.
func NewNamedLeaf(count int, name, pattern string) *Leaf {
	t, err := Transform(pattern)
	if err != nil {
		panic(err)
	}
	re, err := regexp.Compile(t)
	if err != nil {
		panic(fmt.Sprintf("regex: leaf %q: %v", pattern, err))
	}
	if re.NumSubexp() != count {
		panic(fmt.Sprintf("regex: leaf %q has %d groups, declared %d", pattern, re.NumSubexp(), count))
	}
	return &Leaf{name: name, count: count, pattern: pattern}
}

// Spaces between tokens.
func SpaceZeroOrMore() *Leaf { return NewLeaf("[%s]*") }
func SpaceOne() *Leaf        { return NewLeaf("[%s]") }
func SpaceOneOrMore() *Leaf  { return NewLeaf("[%s]+") }
func SpaceZeroOrOne() Part   { return Optional(SpaceOne()) }

func (l *Leaf) Pattern() string { return l.pattern }
func (l *Leaf) Count() int      { return l.count }

func (l *Leaf) collect(groups *[]Group, r *Result) {
	taken := (*groups)[:l.count]
	*groups = (*groups)[l.count:]
	if l.name != "" {
		r.put(l.name, append([]Group(nil), taken...))
	}
}

// composed holds sub-parts and the shared match logic.
type composed struct {
	parts []Part
}

func (c *composed) countParts() int {
	n := 0
	for _, p := range c.parts {
		n += p.Count()
	}
	return n
}

func (c *composed) collectParts(groups *[]Group, r *Result) {
	for _, p := range c.parts {
		p.collect(groups, r)
	}
}

type optional struct{ composed }

// Optional makes a part optional.
func Optional(p Part) Part { return &optional{composed{[]Part{p}}} }

func (o *optional) Pattern() string                    { return "(?:" + o.parts[0].Pattern() + ")?" }
func (o *optional) Count() int                         { return o.countParts() }
func (o *optional) collect(groups *[]Group, r *Result) { o.collectParts(groups, r) }

type zeroOrMore struct{ composed }

// ZeroOrMore repeats a part zero or more times.
func ZeroOrMore(p Part) Part { return &zeroOrMore{composed{[]Part{p}}} }

func (z *zeroOrMore) Pattern() string                    { return "(?:" + z.parts[0].Pattern() + ")*" }
func (z *zeroOrMore) Count() int                         { return z.countParts() }
func (z *zeroOrMore) collect(groups *[]Group, r *Result) { z.collectParts(groups, r) }

// named wraps alternatives or repetitions in a capturing group reported
// under name, when name is not empty.
type named struct {
	composed
	name   string
	prefix string // "(" + body + suffix
	sep    string
	suffix string
}

// Or matches one of the parts. A non-empty name captures the whole choice.
func Or(name string, parts ...Part) Part {
	return &named{composed: composed{parts}, name: name, sep: "|", suffix: ")"}
}

// OneOrMore repeats a part one or more times, capturing the whole repetition
// under name.
func OneOrMore(name string, p Part) Part {
	return &named{composed: composed{[]Part{p}}, name: name, suffix: ")+"}
}

func (n *named) Pattern() string {
	var b strings.Builder
	b.WriteString("(")
	if n.name == "" && n.sep == "|" {
		b.WriteString("?:")
	}
	for i, p := range n.parts {
		if i > 0 {
			b.WriteString(n.sep)
		}
		b.WriteString(p.Pattern())
	}
	b.WriteString(n.suffix)
	return b.String()
}

// Count includes the group upstream always opens: Or and OneOrMore count one
// group even without a name, matching RegexOr and RegexRepeatedOneOrMore.
func (n *named) Count() int { return 1 + n.countParts() }

// collect consumes the whole-match group only for a named part, as upstream
// does: an unnamed OneOrMore still opens a capturing group, so the groups
// after it shift, exactly like upstream.
func (n *named) collect(groups *[]Group, r *Result) {
	var whole Group
	if n.name != "" {
		whole = (*groups)[0]
		*groups = (*groups)[1:]
	}
	n.collectParts(groups, r)
	if n.name != "" {
		r.put(n.name, []Group{whole})
	}
}

// Concat is a full command pattern: parts concatenated and compiled once.
type Concat struct {
	composed
	once sync.Once
	re   *regexp.Regexp

	namesOnce sync.Once
	names     []string
}

// javaOrder returns the names of the pattern in upstream's HashMap order
// (see javaorder.go), computed once.
func (c *Concat) javaOrder() []string {
	c.namesOnce.Do(func() { c.names = javaNames(c) })
	return c.names
}

// NewConcat concatenates parts; upstream's commands start with Start() and
// end with End().
func NewConcat(parts ...Part) *Concat { return &Concat{composed: composed{parts}} }

func (c *Concat) Pattern() string {
	var b strings.Builder
	for _, p := range c.parts {
		b.WriteString(p.Pattern())
	}
	return b.String()
}

func (c *Concat) Count() int                         { return c.countParts() }
func (c *Concat) collect(groups *[]Group, r *Result) { c.collectParts(groups, r) }

func (c *Concat) compiled() *regexp.Regexp {
	c.once.Do(func() { c.re = MustCompile(c.Pattern()) })
	return c.re
}

// Match reports whether the pattern matches s.
func (c *Concat) Match(s string) bool { return c.compiled().MatchString(s) }

// Matcher matches s and returns the named groups, or nil when s does not
// match.
func (c *Concat) Matcher(s string) *Result {
	idx := c.compiled().FindStringSubmatchIndex(s)
	if idx == nil {
		return nil
	}
	groups := make([]Group, 0, len(idx)/2-1)
	for i := 2; i < len(idx); i += 2 {
		if idx[i] < 0 {
			groups = append(groups, Group{})
		} else {
			groups = append(groups, Group{S: s[idx[i]:idx[i+1]], OK: true})
		}
	}
	r := &Result{data: map[string][]Group{}, javaOrder: c.javaOrder()}
	c.collect(&groups, r)
	return r
}

// Result holds the groups of a match by name.
type Result struct {
	data  map[string][]Group
	order []string
	// javaOrder holds the names in the iteration order of upstream's
	// HashMap of groups, which getLazzy follows.
	javaOrder []string
}

func (r *Result) put(name string, groups []Group) {
	if _, ok := r.data[name]; !ok {
		r.order = append(r.order, name)
	}
	r.data[name] = groups
}

// Get returns group num of name; ok is false when the group did not match.
func (r *Result) Get(name string, num int) (string, bool) {
	g, ok := r.data[name]
	if !ok || num >= len(g) {
		return "", false
	}
	return g[num].S, g[num].OK
}

// Str returns group num of name, or "" when it did not match.
func (r *Result) Str(name string, num int) string {
	s, _ := r.Get(name, num)
	return s
}

// Lazy returns group num of the first name starting with prefix whose group
// matched, trying the names in the iteration order of upstream's HashMap
// of groups (which matters when several such groups match at once).
func (r *Result) Lazy(prefix string, num int) (string, bool) {
	order := r.javaOrder
	if order == nil {
		order = r.order
	}
	for _, name := range order {
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		if g := r.data[name]; num < len(g) && g[num].OK {
			return g[num].S, true
		}
	}
	return "", false
}
