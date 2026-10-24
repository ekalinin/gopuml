// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/style/parser/RawStyleParser.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/style/parser/RawStyleSheet.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/style/parser/RawStyleRule.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/style/parser/RawSelector.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/style/parser/MergedStyleNode.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/style/parser/PrioritizedValue.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/style/LegacyStyleFlattener.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/style/parser/StyleParsingException.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/utils/CharInspector.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/utils/CharInspectorImpl.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/ekalinin/gopuml/internal/command"
)

// StyleParsingError is upstream's StyleParsingException: a style sheet that
// cannot be parsed.
type StyleParsingError struct {
	Message string
}

func (e *StyleParsingError) Error() string { return e.Message }

// charInspector walks the characters of lines, upstream's CharInspectorImpl:
// each line is trimmed, a newline is added after each line when
// insertNewlines is set, empty lines are skipped, and peek does not look
// past the current line (0 there).
type charInspector struct {
	lines          [][]rune
	insertNewlines bool
	line           int
	pos            int
}

func newCharInspector(lines command.BlocLines, insertNewlines bool) *charInspector {
	ci := &charInspector{insertNewlines: insertNewlines}
	for _, l := range lines.Lines() {
		s := l.Trimmed().S
		if insertNewlines {
			s += "\n"
		}
		ci.lines = append(ci.lines, []rune(s))
	}
	if len(ci.lines) == 0 {
		// Upstream reads the first line unconditionally and throws.
		throw(indexOutOfBounds, "Index 0 out of bounds for length 0")
	}
	return ci
}

func (ci *charInspector) peek(ahead int) rune {
	if ci.line == -1 {
		return 0
	}
	current := ci.lines[ci.line]
	if ci.pos+ahead >= len(current) {
		return 0
	}
	return current[ci.pos+ahead]
}

func (ci *charInspector) jump() {
	if ci.line == -1 {
		throw(illegalState, "")
	}
	ci.pos++
	if ci.pos >= len(ci.lines[ci.line]) {
		ci.line++
		ci.pos = 0
	}
	for ci.line < len(ci.lines) && len(ci.lines[ci.line]) == 0 {
		ci.line++
	}
	if ci.line >= len(ci.lines) {
		ci.line = -1
	}
}

// rawSelectorKind is upstream's RawSelector.Kind.
type rawSelectorKind int

const (
	rawSelectorName rawSelectorKind = iota
	rawSelectorStereotype
	rawSelectorDepth
	rawSelectorUnknown
)

// rawSelector is one name of a selector list, upstream's RawSelector: a
// style name, a ".stereotype", a "depth(n)" or an unknown word.
type rawSelector struct {
	kind  rawSelectorKind
	raw   string
	sname SName
	depth int
}

func classifyRawSelector(raw string) rawSelector {
	if strings.HasPrefix(raw, ".") {
		return rawSelector{kind: rawSelectorStereotype, raw: raw, depth: -1}
	}
	if strings.HasPrefix(raw, "depth(") && strings.HasSuffix(raw, ")") && len(raw) >= len("depth()") {
		inside := strings.TrimFunc(raw[len("depth("):len(raw)-1], func(r rune) bool { return r <= ' ' })
		if depth, err := strconv.ParseInt(inside, 10, 32); err == nil && inside != "+" {
			return rawSelector{kind: rawSelectorDepth, raw: raw, depth: int(depth)}
		}
		// Not a depth(...) selector after all.
	}
	sname, ok := SNameRetrieve(raw)
	if !ok {
		return rawSelector{kind: rawSelectorUnknown, raw: raw, depth: -1}
	}
	return rawSelector{kind: rawSelectorName, raw: raw, sname: sname, depth: -1}
}

// canonicalOtherKey is the key of a selector that is not a style name.
func (r rawSelector) canonicalOtherKey() string {
	switch r.kind {
	case rawSelectorStereotype:
		return javaLower(r.raw[1:])
	case rawSelectorDepth:
		return "depth(" + itoa(r.depth) + ")"
	case rawSelectorUnknown:
		return javaLower(r.raw)
	}
	throw(illegalState, "Not applicable")
	return ""
}

// rawStyleRule is a "selectors { ... }" or "@media { ... }" block as
// written, upstream's RawStyleRule: its properties in declaration order and
// its nested blocks.
type rawStyleRule struct {
	selectors     []rawSelector
	star          bool
	media         bool
	propertyOrder []PName
	properties    map[PName]string
	children      []*rawStyleRule
}

func (r *rawStyleRule) putProperty(name PName, value string) {
	if r.properties == nil {
		r.properties = map[PName]string{}
	}
	if _, ok := r.properties[name]; !ok {
		r.propertyOrder = append(r.propertyOrder, name)
	}
	r.properties[name] = value
}

// rawStyleParser turns the text of a style sheet into its blocks, upstream's
// RawStyleParser. Comments ("//", "/* */" and "/' '/") and blanks between
// tokens are skipped; "--name: value" declares a variable, used as
// "var(--name)".
type rawStyleParser struct {
	ins       *charInspector
	variables map[string]string
	stack     []*rawStyleRule
	topLevel  []*rawStyleRule
}

// parseRawStyleSheet returns the top-level blocks of lines.
func parseRawStyleSheet(lines command.BlocLines) ([]*rawStyleRule, error) {
	if lines.Size() == 0 {
		return nil, nil
	}
	p := &rawStyleParser{ins: newCharInspector(lines, true), variables: map[string]string{}}
	return p.run()
}

func (p *rawStyleParser) run() ([]*rawStyleRule, error) {
	for {
		p.skipWhitespaceAndComments()
		c := p.ins.peek(0)
		if c == 0 {
			break
		}
		var err error
		switch c {
		case '}':
			p.ins.jump()
			p.closeCurrentRule()
		case '@':
			err = p.parseAtRule()
		case ':':
			// A selector may be written with a leading colon, e.g. ":foo {".
			p.ins.jump()
			err = p.parseSelectorHead(p.readIdentifier())
		default:
			err = p.parseStatementStartingWithIdentifier()
		}
		if err != nil {
			return nil, err
		}
	}
	// A block still open at the end is closed, as if its "}" were there.
	for len(p.stack) > 0 {
		p.closeCurrentRule()
	}
	return p.topLevel, nil
}

func (p *rawStyleParser) parseStatementStartingWithIdentifier() error {
	head := p.readIdentifier()
	if head == "" {
		return &StyleParsingError{Message: "Unexpected character '" + string(p.ins.peek(0)) + "'"}
	}
	if strings.HasPrefix(head, "--") {
		p.parseVariableDeclaration(head)
		return nil
	}
	p.skipWhitespaceAndComments()
	if next := p.ins.peek(0); next == ',' || next == '*' || next == '{' {
		return p.parseSelectorHead(head)
	}
	return p.parsePropertyDeclaration(head)
}

func (p *rawStyleParser) parseSelectorHead(firstName string) error {
	names := []string{firstName}
	p.skipWhitespaceAndComments()
	for p.ins.peek(0) == ',' {
		p.ins.jump()
		p.skipWhitespaceAndComments()
		names = append(names, p.readIdentifier())
		p.skipWhitespaceAndComments()
	}
	star := false
	if p.ins.peek(0) == '*' {
		star = true
		p.ins.jump()
		p.skipWhitespaceAndComments()
	}
	if p.ins.peek(0) != '{' {
		return &StyleParsingError{Message: "Expected '{' after selector [" + strings.Join(names, ", ") + "]"}
	}
	p.ins.jump()
	rule := &rawStyleRule{star: star}
	for _, name := range names {
		rule.selectors = append(rule.selectors, classifyRawSelector(name))
	}
	p.stack = append(p.stack, rule)
	return nil
}

func (p *rawStyleParser) parseAtRule() error {
	p.ins.jump() // '@'
	var header strings.Builder
	for {
		c := p.ins.peek(0)
		if c == 0 || c == '}' || c == ';' {
			return &StyleParsingError{Message: "Unterminated @ rule: '@" + header.String() + "'"}
		}
		if c == '{' {
			break
		}
		header.WriteRune(c)
		p.ins.jump()
	}
	p.ins.jump() // '{'
	p.stack = append(p.stack, &rawStyleRule{media: true})
	return nil
}

func (p *rawStyleParser) parseVariableDeclaration(name string) {
	value := p.readValue()
	p.variables[name[2:]] = p.resolveVariables(value)
}

func (p *rawStyleParser) parsePropertyDeclaration(name string) error {
	if len(p.stack) == 0 {
		return &StyleParsingError{Message: "Property '" + name + "' declared outside of any block"}
	}
	rawValue := p.readValue()
	pname, ok := PNameFromName(name, StyleSchemeRegular)
	if !ok {
		return nil // An unknown property name is ignored.
	}
	p.stack[len(p.stack)-1].putProperty(pname, p.resolveVariables(rawValue))
	return nil
}

// closeCurrentRule closes the innermost open block; a stray "}" is ignored.
func (p *rawStyleParser) closeCurrentRule() {
	if len(p.stack) == 0 {
		return
	}
	closed := p.stack[len(p.stack)-1]
	p.stack = p.stack[:len(p.stack)-1]
	if len(p.stack) == 0 {
		p.topLevel = append(p.topLevel, closed)
	} else {
		parent := p.stack[len(p.stack)-1]
		parent.children = append(parent.children, closed)
	}
}

// readValue reads up to the next ";", end of line or "}"; a double-quoted
// part is copied without its quotes.
func (p *rawStyleParser) readValue() string {
	p.skipInlineSpaces()
	for p.ins.peek(0) == ':' {
		p.ins.jump()
	}
	p.skipInlineSpaces()
	var sb strings.Builder
	for {
		c := p.ins.peek(0)
		if c == 0 || c == '\n' || c == '\r' || c == '}' {
			break
		}
		if c == ';' {
			p.ins.jump()
			break
		}
		if c == '"' {
			p.ins.jump()
			for p.ins.peek(0) != 0 && p.ins.peek(0) != '"' {
				sb.WriteRune(p.ins.peek(0))
				p.ins.jump()
			}
			if p.ins.peek(0) == '"' {
				p.ins.jump()
			}
			continue
		}
		sb.WriteRune(c)
		p.ins.jump()
	}
	return strings.TrimFunc(sb.String(), func(r rune) bool { return r <= ' ' })
}

// readIdentifier reads a bare name or a double-quoted string. A name that
// starts with "." (a stereotype) may contain spaces.
func (p *rawStyleParser) readIdentifier() string {
	if p.ins.peek(0) == '"' {
		return p.readQuoted()
	}
	stereotypeLed := p.ins.peek(0) == '.'
	var sb strings.Builder
	for {
		c := p.ins.peek(0)
		if c == ' ' && stereotypeLed {
			sb.WriteRune(c)
			p.ins.jump()
			continue
		}
		if isRawStyleDelimiter(c) {
			break
		}
		sb.WriteRune(c)
		p.ins.jump()
	}
	if stereotypeLed {
		return strings.TrimFunc(sb.String(), func(r rune) bool { return r <= ' ' })
	}
	return sb.String()
}

func (p *rawStyleParser) readQuoted() string {
	p.ins.jump() // opening quote
	var sb strings.Builder
	for p.ins.peek(0) != 0 && p.ins.peek(0) != '"' {
		sb.WriteRune(p.ins.peek(0))
		p.ins.jump()
	}
	if p.ins.peek(0) == '"' {
		p.ins.jump()
	}
	return sb.String()
}

func isRawStyleDelimiter(c rune) bool {
	switch c {
	case 0, ' ', '\t', '\n', '\r', '{', '}', ';', ',', ':', '*':
		return true
	}
	return false
}

func (p *rawStyleParser) skipInlineSpaces() {
	for p.ins.peek(0) == ' ' || p.ins.peek(0) == '\t' {
		p.ins.jump()
	}
}

func (p *rawStyleParser) skipWhitespaceAndComments() {
	for {
		c := p.ins.peek(0)
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			p.ins.jump()
		case c == '/' && p.ins.peek(1) == '/':
			for p.ins.peek(0) != 0 && p.ins.peek(0) != '\n' {
				p.ins.jump()
			}
		case c == '/' && p.ins.peek(1) == '*':
			p.ins.jump()
			p.ins.jump()
			p.skipUntilSequence('*', '/')
		case c == '/' && p.ins.peek(1) == '\'':
			p.ins.jump()
			p.ins.jump()
			p.skipUntilSequence('\'', '/')
		default:
			return
		}
	}
}

func (p *rawStyleParser) skipUntilSequence(first, second rune) {
	for p.ins.peek(0) != 0 {
		if p.ins.peek(0) == first && p.ins.peek(1) == second {
			p.ins.jump()
			p.ins.jump()
			return
		}
		p.ins.jump()
	}
}

// styleVarRef matches var(--name), var(-name) or var(name); \s is Java's.
var styleVarRef = regexp.MustCompile(`var\([ \t\n\x0B\f\r]*-*([A-Za-z_][-\w]*)[ \t\n\x0B\f\r]*\)`)

// resolveVariables replaces each var(--name) with the value of name, when
// it is known.
func (p *rawStyleParser) resolveVariables(value string) string {
	if !strings.Contains(value, "var(") {
		return value
	}
	return styleVarRef.ReplaceAllStringFunc(value, func(m string) string {
		if v, ok := p.variables[styleVarRef.FindStringSubmatch(m)[1]]; ok {
			return v
		}
		return m
	})
}

// prioritizedValue is the regular and dark texts of a property with the
// order of its declaration, upstream's PrioritizedValue.
type prioritizedValue struct {
	light, dark *string
	priority    int
}

// mergeWith folds a declaration over existing (which may be nil): a regular
// and a dark value combine, otherwise the later declaration wins.
func (v *prioritizedValue) mergeWith(existing *prioritizedValue) *prioritizedValue {
	if existing == nil {
		return v
	}
	if (v.dark == nil && existing.dark == nil) || (v.light == nil && existing.light == nil) {
		if v.priority > existing.priority {
			return v
		}
		return existing
	}
	if v.dark == nil && existing.light == nil {
		return &prioritizedValue{light: v.light, dark: existing.dark, priority: v.priority}
	}
	if existing.dark == nil && v.light == nil {
		return &prioritizedValue{light: existing.light, dark: v.dark, priority: existing.priority}
	}
	if v.priority > existing.priority {
		return v
	}
	return existing
}

// mergedStyleNode is a selector of the merged tree of a style sheet,
// upstream's MergedStyleNode: its properties, and its children by style
// name (visited in enum order) and by other key (in insertion order), a
// plain and a starred selector of the same name being separate children.
type mergedStyleNode struct {
	named, starredNamed      map[SName]*mergedStyleNode
	otherOrder, starredOrder []string
	other, starredOther      map[string]*mergedStyleNode
	otherKinds               map[string]rawSelectorKind
	properties               [pnameCount]*prioritizedValue
	star                     bool
}

func newMergedStyleNode(star bool) *mergedStyleNode {
	return &mergedStyleNode{star: star, named: map[SName]*mergedStyleNode{},
		starredNamed: map[SName]*mergedStyleNode{}, other: map[string]*mergedStyleNode{},
		starredOther: map[string]*mergedStyleNode{}, otherKinds: map[string]rawSelectorKind{}}
}

// mergeAll merges the top-level blocks, the content of "@media" blocks as
// dark values (the condition is not read).
func (n *mergedStyleNode) mergeAll(rules []*rawStyleRule, counter AutomaticCounter) {
	for _, rule := range rules {
		if !rule.media {
			n.mergeRule(rule, counter, false, map[*rawStyleRule]*[pnameCount]*prioritizedValue{})
			continue
		}
		for _, nested := range rule.children {
			n.mergeRule(nested, counter, true, map[*rawStyleRule]*[pnameCount]*prioritizedValue{})
		}
	}
}

// mergeRule merges a block and its nested blocks under n. The values of a
// block are numbered once, even when a selector list makes it reached
// several times. An "@media" block inside a selector panics, as upstream
// throws IllegalArgumentException.
func (n *mergedStyleNode) mergeRule(rule *rawStyleRule, counter AutomaticCounter, dark bool,
	ownValuesCache map[*rawStyleRule]*[pnameCount]*prioritizedValue) {
	if rule.media {
		throw(illegalArgument, "An @media block is not a selector; merge its children instead")
	}
	ownValues := ownValuesCache[rule]
	if ownValues == nil {
		ownValues = &[pnameCount]*prioritizedValue{}
		for _, name := range rule.propertyOrder {
			v := rule.properties[name]
			pv := &prioritizedValue{priority: counter.NextInt()}
			if dark {
				pv.dark = &v
			} else {
				pv.light = &v
			}
			ownValues[name] = pv
		}
		ownValuesCache[rule] = ownValues
	}
	for _, selector := range rule.selectors {
		child := n.getOrCreateChild(selector, rule.star)
		for name, pv := range ownValues {
			if pv != nil {
				child.properties[name] = pv.mergeWith(child.properties[name])
			}
		}
		for _, nested := range rule.children {
			child.mergeRule(nested, counter, dark, ownValuesCache)
		}
	}
}

func (n *mergedStyleNode) getOrCreateChild(selector rawSelector, star bool) *mergedStyleNode {
	if selector.kind == rawSelectorName {
		children := n.named
		if star {
			children = n.starredNamed
		}
		child := children[selector.sname]
		if child == nil {
			child = newMergedStyleNode(star)
			children[selector.sname] = child
		}
		return child
	}
	key := selector.canonicalOtherKey()
	children, order := n.other, &n.otherOrder
	if star {
		children, order = n.starredOther, &n.starredOrder
	}
	child := children[key]
	if child == nil {
		child = newMergedStyleNode(star)
		children[key] = child
		*order = append(*order, key)
		n.otherKinds[key] = selector.kind
	}
	return child
}

// flattenStyles is upstream's LegacyStyleFlattener: one style per node of
// the tree that has properties, with the query of its path. The values of
// a selector with stereotypes count them in their specificity, and an
// unknown selector word is taken as a stereotype.
func flattenStyles(root *mergedStyleNode) []*Style {
	var result []*Style
	flattenStyleNode(root, StyleSignatureBasicEmpty(), &result)
	return result
}

func flattenStyleNode(node *mergedStyleNode, pathSoFar StyleSignatureBasic, result *[]*Style) {
	var values styleValues
	count := 0
	for name, pv := range node.properties {
		if pv != nil {
			specificity := SpecificityAtOrder(pv.priority).WithStereotypeCount(len(pathSoFar.Stereotypes()))
			values[name] = ValueImplOf(pv.light, pv.dark, specificity)
			count++
		}
	}
	if count > 0 {
		*result = append(*result, newStyleValues(pathSoFar, values))
	}
	flattenNamedStyles(node.named, pathSoFar, result)
	flattenNamedStyles(node.starredNamed, pathSoFar, result)
	flattenOtherStyles(node, node.otherOrder, node.other, pathSoFar, result)
	flattenOtherStyles(node, node.starredOrder, node.starredOther, pathSoFar, result)
}

func flattenNamedStyles(children map[SName]*mergedStyleNode, pathSoFar StyleSignatureBasic, result *[]*Style) {
	for name := SName(0); name < snameCount; name++ {
		child := children[name]
		if child == nil {
			continue
		}
		childSignature := pathSoFar.AddSName(name)
		if child.star {
			childSignature = childSignature.AddStar()
		}
		flattenStyleNode(child, childSignature, result)
	}
}

func flattenOtherStyles(node *mergedStyleNode, order []string, children map[string]*mergedStyleNode,
	pathSoFar StyleSignatureBasic, result *[]*Style) {
	for _, key := range order {
		child := children[key]
		var childSignature StyleSignatureBasic
		if node.otherKinds[key] == rawSelectorDepth {
			childSignature = pathSoFar.AddLevel(javaParseInt(key[len("depth(") : len(key)-1]))
		} else {
			childSignature = pathSoFar.AddStereotype(key)
		}
		if child.star {
			childSignature = childSignature.AddStar()
		}
		flattenStyleNode(child, childSignature, result)
	}
}
