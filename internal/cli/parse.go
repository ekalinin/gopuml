// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cli/CliParsed.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cli/CliParsingException.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package cli

import (
	"fmt"
	"strings"
)

// Value is one value of a flag; Null marks a value that upstream reads as
// null, such as a missing argument after --output-dir.
type Value struct {
	S    string
	Null bool
}

// KeyValue is one -DKEY=VALUE style entry; HasValue is false without "=".
type KeyValue struct {
	Key, Value string
	HasValue   bool
}

// Parsed holds the flags and the remaining arguments of a command line.
type Parsed struct {
	lists map[Flag][]Value
	maps  map[Flag][]KeyValue
	args  []string
}

// ParseError is upstream's CliParsingException.
type ParseError struct {
	Flag    Flag
	Message string
}

// Error is upstream's message, such as "-D [DEFINE]: missing value".
func (e *ParseError) Error() string {
	return e.Flag.String() + " [" + e.Flag.Info().Name + "]: " + e.Message
}

// javaSplit splits like Java's String.split with a literal separator:
// trailing empty strings are removed, but an empty input gives [""].
func javaSplit(s, sep string) []string {
	parts := strings.Split(s, sep)
	if s == "" {
		return parts
	}
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}

// Parse reads a command line. Each argument selects the first flag, in
// declaration order, that matches it; other arguments are input paths.
func Parse(args []string) (*Parsed, error) {
	p := &Parsed{lists: map[Flag][]Value{}, maps: map[Flag][]KeyValue{}}
	put := func(f Flag, v Value) { p.lists[f] = append(p.lists[f], v) }
	for i := 0; i < len(args); i++ {
		arg := args[i]
		found := false
		for f := range Flag(flagCount) {
			if !f.match(arg) {
				continue
			}
			found = true
			switch f.Info().Arity {
			case UnaryBoolean, UnaryImmediateAction:
				put(f, Value{S: "true"})
			case BinaryNextArgumentValue:
				i++
				if i < len(args) {
					put(f, Value{S: args[i]})
				} else {
					put(f, Value{Null: true})
				}
			case UnaryInlineKeyOrKeyValue:
				s := arg[2:]
				if s == "" {
					i++
					if i >= len(args) {
						return nil, &ParseError{Flag: f, Message: "missing value"}
					}
					s = args[i]
				}
				kv := KeyValue{Key: s}
				if k, v, ok := strings.Cut(s, "="); ok {
					kv = KeyValue{Key: k, Value: v, HasValue: true}
				}
				p.maps[f] = setKeyValue(p.maps[f], kv)
			case UnaryOptionalColon:
				colon := strings.IndexByte(arg, ':')
				if colon == -1 {
					if p.lists[f] == nil {
						p.lists[f] = []Value{}
					}
					break
				}
				for _, part := range javaSplit(arg[colon+1:], ":") {
					put(f, Value{S: part})
				}
			}
			break
		}
		if !found {
			p.args = append(p.args, arg)
		}
	}
	return p, nil
}

// setKeyValue updates a key in place or appends it, like a LinkedHashMap.
func setKeyValue(list []KeyValue, kv KeyValue) []KeyValue {
	for i := range list {
		if list[i].Key == kv.Key {
			list[i] = kv
			return list
		}
	}
	return append(list, kv)
}

// IsTrue reports whether the flag was given.
func (p *Parsed) IsTrue(f Flag) bool {
	if _, ok := p.maps[f]; ok {
		return true
	}
	_, ok := p.lists[f]
	return ok
}

// String returns the first value of a flag, or its default; ok is false
// when there is neither, or the value is null.
func (p *Parsed) String(f Flag) (string, bool) {
	if l := p.lists[f]; len(l) > 0 {
		return l[0].S, !l[0].Null
	}
	if d := f.Info().Default; d != "" {
		return d, true
	}
	return "", false
}

// List returns every value of a flag.
func (p *Parsed) List(f Flag) []Value { return p.lists[f] }

// Map returns the entries of a -DKEY=VALUE style flag, in order, merged with
// the KEY=VALUE values of its long form (for example --define).
func (p *Parsed) Map(mapFlag, listFlag Flag) []KeyValue {
	out := append([]KeyValue(nil), p.maps[mapFlag]...)
	for _, v := range p.lists[listFlag] {
		switch parts := javaSplit(v.S, "="); len(parts) {
		case 1:
			out = setKeyValue(out, KeyValue{Key: parts[0]})
		case 2:
			out = setKeyValue(out, KeyValue{Key: parts[0], Value: parts[1], HasValue: true})
		}
	}
	return out
}

// Format returns the output format selected by a format flag, upstream's
// name such as "SVG", or "" when none. Like upstream, the first format flag
// in declaration order given exactly once wins.
func (p *Parsed) Format() string {
	for f := range Flag(flagCount) {
		if l, ok := p.lists[f]; ok && len(l) == 1 && f.Info().Format != "" {
			return f.Info().Format
		}
	}
	return ""
}

// ImmediateAction returns the first flag, in declaration order, that runs
// an action and exits (such as --version), given exactly once.
func (p *Parsed) ImmediateAction() (Flag, bool) {
	for f := range Flag(flagCount) {
		if l, ok := p.lists[f]; ok && len(l) == 1 && f.Info().Arity == UnaryImmediateAction {
			return f, true
		}
	}
	return 0, false
}

// Args returns the arguments that are not flags.
func (p *Parsed) Args() []string { return p.args }

// Given returns the flags present on the command line, in declaration order.
func (p *Parsed) Given() []Flag {
	var out []Flag
	for f := range Flag(flagCount) {
		if p.IsTrue(f) {
			out = append(out, f)
		}
	}
	return out
}

// UnsupportedError reports an upstream flag this implementation does not
// provide; the command line fails with exit status 1.
type UnsupportedError struct{ Flag Flag }

func (e *UnsupportedError) Error() string {
	return fmt.Sprintf("the %s option is not supported by this implementation", e.Flag)
}
