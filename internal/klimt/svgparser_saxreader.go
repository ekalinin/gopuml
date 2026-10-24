// SPDX-License-Identifier: MIT
// Provenance: original

package klimt

import (
	"encoding/xml"
	"errors"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// The SAX events of Java's SAXParser, as SvgSaxParser configures it
// (namespace aware, DOCTYPE disallowed), from Go's encoding/xml tokenizer.
// The handler sees the qualified names (prefix:local) of elements and
// attributes, the attributes in document order without the namespace
// declarations (xmlns, xmlns:p), and the text in one or more characters
// events. Parsing stops at the first error that Xerces reports for
// malformed XML: the events before it are delivered, as SAX delivers them
// while parsing.
//
// Known differences with Xerces: attribute values are normalized as XML 1.0
// requires (each tab, newline and carriage return becomes a space), but
// character references to those characters (&#9; &#10; &#13;) are
// normalized too, where Xerces keeps them; the error messages are not
// Xerces' messages; XML 1.1 documents are rejected.

// saxAttr is an attribute: its qualified name and value.
type saxAttr struct {
	qName string
	value string
}

// saxAttributes is upstream's org.xml.sax.Attributes.
type saxAttributes []saxAttr

// value returns the value of the attribute qName; ok is false when there is
// none (null in Java).
func (a saxAttributes) value(qName string) (string, bool) {
	for _, attr := range a {
		if attr.qName == qName {
			return attr.value, true
		}
	}
	return "", false
}

// saxHandler is the part of org.xml.sax.helpers.DefaultHandler that the SVG
// sprite parser overrides.
type saxHandler interface {
	startElement(qName string, attrs saxAttributes)
	endElement(qName string)
	characters(text string)
}

// saxParseError is a SAXParseException: the document is not well-formed.
type saxParseError struct{ message string }

func (e *saxParseError) Error() string { return "org.xml.sax.SAXParseException: " + e.message }

func saxError(message string) error { return &saxParseError{message: message} }

// The XML declaration as XML 1.0 and 1.1 define it, and its version 1.1,
// which encoding/xml rejects.
var (
	saxXMLDecl = regexp.MustCompile(`^<\?xml[ \t\r\n]+version[ \t\r\n]*=[ \t\r\n]*("1\.[01]"|'1\.[01]')` +
		`([ \t\r\n]+encoding[ \t\r\n]*=[ \t\r\n]*("[A-Za-z][A-Za-z0-9._-]*"|'[A-Za-z][A-Za-z0-9._-]*'))?` +
		`([ \t\r\n]+standalone[ \t\r\n]*=[ \t\r\n]*("(yes|no)"|'(yes|no)'))?[ \t\r\n]*\?>$`)
	saxXMLDeclVersion11 = regexp.MustCompile(`^<\?xml[ \t\r\n]+version[ \t\r\n]*=[ \t\r\n]*["']1\.(1)["']`)
	saxCharRef          = regexp.MustCompile(`&#(x[0-9a-fA-F]+|[0-9]+);`)
)

// saxParse parses content and sends its events to h. It returns an error
// when the document is not well-formed; the panics of h go through.
//
// encoding/xml is less strict than Xerces on a few points, which the raw
// text of the tokens checks: the XML declaration, whitespace between
// attributes, and character references to characters that XML does not
// allow (surrogates). It rejects version 1.1, which Xerces reads: the
// tokenizer sees version 1.0 instead (XML 1.1 documents are then read with
// the rules of XML 1.0).
func saxParse(content string, h saxHandler) error {
	input := content
	if m := saxXMLDeclVersion11.FindStringSubmatchIndex(content); m != nil {
		input = content[:m[2]] + "0" + content[m[3]:]
	}
	d := xml.NewDecoder(strings.NewReader(input))
	d.Strict = true
	// Java reads characters from a StringReader: the declared encoding does
	// not matter.
	d.CharsetReader = func(_ string, input io.Reader) (io.Reader, error) { return input, nil }

	type element struct {
		qName    string
		prefixes []string // prefixes declared by the element ("" for the default)
	}
	var stack []element
	bound := map[string]int{} // prefix to the number of declarations in scope
	seenRoot := false
	rootClosed := false

	for {
		offset := d.InputOffset()
		tok, err := d.RawToken()
		if err == io.EOF {
			if !seenRoot {
				return saxError("Premature end of file.")
			}
			if len(stack) > 0 {
				return saxError("XML document structures must start and end within the same entity.")
			}
			return nil
		}
		if err != nil {
			var syntaxErr *xml.SyntaxError
			if errors.As(err, &syntaxErr) {
				return saxError(syntaxErr.Msg)
			}
			return saxError(err.Error())
		}

		raw := content[offset:d.InputOffset()]

		switch t := tok.(type) {
		case xml.StartElement:
			if rootClosed {
				return saxError("The markup in the document following the root element must be well-formed.")
			}
			seenRoot = true
			qName := saxQName(t.Name)
			if err := saxCheckStartTag(raw, qName); err != nil {
				return err
			}
			if err := saxCheckCharRefs(raw); err != nil {
				return err
			}
			var declared []string
			var attrs saxAttributes
			seen := map[string]bool{}
			for _, a := range t.Attr {
				aName := saxQName(a.Name)
				if seen[aName] {
					return saxError("Attribute \"" + aName + "\" was already specified for element \"" + qName +
						"\".")
				}
				seen[aName] = true
				switch {
				case a.Name.Space == "" && a.Name.Local == "xmlns":
					declared = append(declared, "")
				case a.Name.Space == "xmlns":
					if a.Value == "" {
						return saxError("Prefixed namespace bindings may not be empty (" + aName + ").")
					}
					declared = append(declared, a.Name.Local)
				default:
					attrs = append(attrs, saxAttr{qName: aName, value: saxNormalizeAttribute(a.Value)})
				}
			}
			for _, p := range declared {
				bound[p]++
			}
			stack = append(stack, element{qName: qName, prefixes: declared})
			if p := t.Name.Space; p != "" && p != "xml" && bound[p] == 0 {
				return saxError("The prefix \"" + p + "\" for element \"" + qName + "\" is not bound.")
			}
			for _, a := range attrs {
				if p, _, ok := strings.Cut(a.qName, ":"); ok && p != "xml" && bound[p] == 0 {
					return saxError("The prefix \"" + p + "\" for attribute \"" + a.qName +
						"\" associated with an element type \"" + qName + "\" is not bound.")
				}
			}
			h.startElement(qName, attrs)

		case xml.EndElement:
			qName := saxQName(t.Name)
			if len(stack) == 0 {
				return saxError("The markup in the document following the root element must be well-formed.")
			}
			top := stack[len(stack)-1]
			if top.qName != qName {
				return saxError("The element type \"" + top.qName + "\" must be terminated by the matching end-tag \"</" +
					top.qName + ">\".")
			}
			stack = stack[:len(stack)-1]
			for _, p := range top.prefixes {
				bound[p]--
			}
			h.endElement(qName)
			if len(stack) == 0 {
				rootClosed = true
			}

		case xml.CharData:
			if !strings.HasPrefix(raw, "<![CDATA[") {
				if err := saxCheckCharRefs(raw); err != nil {
					return err
				}
			}
			if len(stack) == 0 {
				if !saxIsWhitespace(string(t)) {
					if rootClosed {
						return saxError("Content is not allowed in trailing section.")
					}
					return saxError("Content is not allowed in prolog.")
				}
				continue
			}
			h.characters(string(t))

		case xml.ProcInst:
			if strings.EqualFold(t.Target, "xml") {
				if t.Target != "xml" || offset != 0 {
					return saxError("The processing instruction target matching \"[xX][mM][lL]\" is not allowed.")
				}
				if !saxXMLDecl.MatchString(raw) {
					return saxError("The XML declaration is not well-formed.")
				}
			}

		case xml.Directive:
			if strings.HasPrefix(string(t), "DOCTYPE") {
				return saxError("DOCTYPE is disallowed when the feature " +
					"\"http://apache.org/xml/features/disallow-doctype-decl\" set to true.")
			}
			return saxError("The markup in the document preceding the root element must be well-formed.")

		case xml.Comment:
			// ignored, as DefaultHandler does
		}
	}
}

// saxCheckStartTag checks that whitespace separates the attributes of a
// start tag (raw is its text, from "<" to ">").
func saxCheckStartTag(raw, qName string) error {
	var quote byte
	for i := 1; i < len(raw); i++ {
		c := raw[i]
		switch {
		case quote != 0:
			if c != quote {
				continue
			}
			quote = 0
			if i+1 < len(raw) {
				switch raw[i+1] {
				case ' ', '\t', '\n', '\r', '/', '>':
				default:
					return saxError("Element type \"" + qName +
						"\" must be followed by either attribute specifications, \">\" or \"/>\".")
				}
			}
		case c == '"' || c == '\'':
			quote = c
		}
	}
	return nil
}

// saxCheckCharRefs checks that the character references of raw text are
// characters that XML allows.
func saxCheckCharRefs(raw string) error {
	if !strings.Contains(raw, "&#") {
		return nil
	}
	for _, m := range saxCharRef.FindAllStringSubmatch(raw, -1) {
		var v uint64
		var err error
		if hex, ok := strings.CutPrefix(m[1], "x"); ok {
			v, err = strconv.ParseUint(hex, 16, 32)
		} else {
			v, err = strconv.ParseUint(m[1], 10, 32)
		}
		if err != nil || !saxIsXMLChar(v) {
			return saxError("Character reference \"" + m[0][:len(m[0])-1] + "\" is an invalid XML character.")
		}
	}
	return nil
}

// saxIsXMLChar reports whether XML 1.0 allows the character (production
// Char).
func saxIsXMLChar(v uint64) bool {
	return v == 0x9 || v == 0xA || v == 0xD || v >= 0x20 && v <= 0xD7FF || v >= 0xE000 && v <= 0xFFFD ||
		v >= 0x10000 && v <= 0x10FFFF
}

// saxQName returns the qualified name of a raw name.
func saxQName(n xml.Name) string {
	if n.Space == "" {
		return n.Local
	}
	return n.Space + ":" + n.Local
}

// saxNormalizeAttribute is the attribute value normalization of XML 1.0 for
// CDATA attributes: each whitespace character becomes a space. The
// tokenizer has already turned line ends into "\n".
func saxNormalizeAttribute(v string) string {
	if !strings.ContainsAny(v, "\t\n\r") {
		return v
	}
	return strings.Map(func(r rune) rune {
		if r == '\t' || r == '\n' || r == '\r' {
			return ' '
		}
		return r
	}, v)
}

// saxIsWhitespace reports whether s is only XML whitespace (S).
func saxIsWhitespace(s string) bool {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case ' ', '\t', '\n', '\r':
		default:
			return false
		}
	}
	return true
}
