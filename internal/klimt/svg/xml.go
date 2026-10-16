// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/svg/IElement.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/svg/PortableSvgDocument.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/svg/XmlContent.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/svg/XmlDocument.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/svg/XmlLeaf.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/svg/XmlNode.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/svg/XmlWriter.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svg

import (
	"strings"

	"github.com/ekalinin/plantuml-in-go/internal/klimt"
)

// XMLContent is a piece of content inside an XMLNode: a nested element or a
// leaf (text, comment, processing instruction, CDATA section or raw markup).
type XMLContent interface {
	writeTo(w *XMLWriter)
}

// XMLNode is upstream's minimal DOM element: a tag name, attributes in
// insertion order and children. The tree is built first, so that
// attributes set late (once the canvas size is known) are written, then
// serialized in one pass.
type XMLNode struct {
	tagName  string
	attrs    []xmlAttr
	children []XMLContent
}

// xmlAttr is one attribute. Upstream stores attributes in a LinkedHashMap:
// setting an existing name again keeps its position.
type xmlAttr struct{ name, value string }

func newXMLNode(tagName string) *XMLNode { return &XMLNode{tagName: tagName} }

// TagName returns the element name.
func (n *XMLNode) TagName() string { return n.tagName }

// SetAttribute sets an attribute, keeping the position of an existing one.
func (n *XMLNode) SetAttribute(name, value string) {
	for i := range n.attrs {
		if n.attrs[i].name == name {
			n.attrs[i].value = value
			return
		}
	}
	n.attrs = append(n.attrs, xmlAttr{name: name, value: value})
}

// Attribute returns the value of an attribute; ok is false when it is not
// set.
func (n *XMLNode) Attribute(name string) (value string, ok bool) {
	for _, a := range n.attrs {
		if a.name == name {
			return a.value, true
		}
	}
	return "", false
}

// AppendChild appends a child element.
func (n *XMLNode) AppendChild(child *XMLNode) { n.children = append(n.children, child) }

// SetTextContent replaces all content with one text node, as
// org.w3c.dom.Node.setTextContent.
func (n *XMLNode) SetTextContent(value string) {
	n.children = nil
	n.AppendText(value)
}

// AppendText appends a text node.
func (n *XMLNode) AppendText(value string) { n.children = append(n.children, xmlText(value)) }

// AppendComment appends a comment.
func (n *XMLNode) AppendComment(value string) { n.children = append(n.children, xmlComment(value)) }

// AppendProcessingInstruction appends a processing instruction
// <?target data?>.
func (n *XMLNode) AppendProcessingInstruction(target, data string) {
	n.children = append(n.children, xmlProcessingInstruction{target: target, data: data})
}

// AppendCData appends a CDATA section.
func (n *XMLNode) AppendCData(value string) { n.children = append(n.children, xmlCData(value)) }

// AppendRaw appends already serialized markup, written verbatim.
func (n *XMLNode) AppendRaw(markup string) { n.children = append(n.children, xmlRaw(markup)) }

// HasChildren reports whether the element has content, upstream's
// getFirstChild() != null.
func (n *XMLNode) HasChildren() bool { return len(n.children) > 0 }

func (n *XMLNode) writeTo(w *XMLWriter) {
	w.StartElement(n.tagName)
	for _, a := range n.attrs {
		w.Attribute(a.name, a.value)
	}
	for _, c := range n.children {
		c.writeTo(w)
	}
	w.EndElement()
}

// ToXML serializes the element and its subtree.
func (n *XMLNode) ToXML(indentSpaces int) string {
	w := NewXMLWriter(indentSpaces)
	n.writeTo(w)
	return w.XML()
}

// The leaves of upstream's XmlLeaf, one type per kind.
type (
	xmlText                  string
	xmlComment               string
	xmlCData                 string
	xmlRaw                   string
	xmlProcessingInstruction struct{ target, data string }
)

func (t xmlText) writeTo(w *XMLWriter)                  { w.Text(string(t)) }
func (c xmlComment) writeTo(w *XMLWriter)               { w.Comment(string(c)) }
func (c xmlCData) writeTo(w *XMLWriter)                 { w.CData(string(c)) }
func (r xmlRaw) writeTo(w *XMLWriter)                   { w.Raw(string(r)) }
func (p xmlProcessingInstruction) writeTo(w *XMLWriter) { w.ProcessingInstruction(p.target, p.data) }

// XMLDocument is upstream's minimal document: an element factory and a
// root.
type XMLDocument struct {
	root *XMLNode
}

// CreateElement returns a new element.
func (d *XMLDocument) CreateElement(name string) *XMLNode { return newXMLNode(name) }

// SetRoot sets the root element.
func (d *XMLDocument) SetRoot(root *XMLNode) { d.root = root }

// Root returns the root element.
func (d *XMLDocument) Root() *XMLNode { return d.root }

// ToXML serializes the document from its root. It panics without root, as
// upstream throws IllegalStateException.
func (d *XMLDocument) ToXML(indentSpaces int) string {
	if d.root == nil {
		throw(illegalState, "No root element set.")
	}
	w := NewXMLWriter(indentSpaces)
	d.root.writeTo(w)
	return w.XML()
}

// ApplyGroupAttribute writes one group attribute on element, upstream's
// PortableSvgDocument.applyGroupAttribute: TITLE becomes a <title> child,
// DATA_UID the id, the entity and participant keys data-entity-1/2, and a
// few keys their own attribute name; the others are ignored.
func (d *XMLDocument) ApplyGroupAttribute(element *XMLNode, key klimt.UGroupType, value string) {
	switch key {
	case klimt.UGroupTitle:
		title := d.CreateElement(klimt.UGroupTitle.SvgKeyAttributeName())
		title.SetTextContent(value)
		element.AppendChild(title)
	case klimt.UGroupID:
		// ignored
	case klimt.UGroupDataUID:
		element.SetAttribute("id", value)
	case klimt.UGroupDataParticipant1, klimt.UGroupDataEntity1UID:
		element.SetAttribute("data-entity-1", value)
	case klimt.UGroupDataParticipant2, klimt.UGroupDataEntity2UID:
		element.SetAttribute("data-entity-2", value)
	case klimt.UGroupClass, klimt.UGroupDataSourceLine, klimt.UGroupDataQualifiedName, klimt.UGroupDataEntityUID,
		klimt.UGroupDataVisibilityModifier, klimt.UGroupDataLinkType:
		element.SetAttribute(key.SvgKeyAttributeName(), value)
	}
}

// XMLWriter is upstream's streaming XML writer. It collapses empty elements
// to self-closing tags, escapes "&" and "<" in text and also '"' in
// attribute values, and indents only when indentSpaces is positive (the SVG
// output uses 0: no newline at all).
type XMLWriter struct {
	out          strings.Builder
	indentSpaces int
	openTags     []string
	// pendingClose is true once "<name" is written but not the ">" of the
	// start tag.
	pendingClose bool
	// inlineContent is true when the current element holds text, so that
	// its end tag is not indented.
	inlineContent bool
}

// NewXMLWriter returns a writer.
func NewXMLWriter(indentSpaces int) *XMLWriter { return &XMLWriter{indentSpaces: indentSpaces} }

// StartElement opens an element.
func (w *XMLWriter) StartElement(name string) *XMLWriter {
	w.closePendingStartTag(true)
	w.indent(len(w.openTags))
	w.out.WriteByte('<')
	w.out.WriteString(name)
	w.openTags = append(w.openTags, name)
	w.pendingClose = true
	w.inlineContent = false
	return w
}

// Attribute adds an attribute to the open start tag; it panics outside of
// one, as upstream throws IllegalStateException.
func (w *XMLWriter) Attribute(name, value string) *XMLWriter {
	if !w.pendingClose {
		throw(illegalState, "Cannot add an attribute outside of an open start tag.")
	}
	w.out.WriteByte(' ')
	w.out.WriteString(name)
	w.out.WriteString(`="`)
	w.escape(value, true)
	w.out.WriteByte('"')
	return w
}

// Text writes escaped text.
func (w *XMLWriter) Text(value string) *XMLWriter {
	w.closePendingStartTag(false)
	w.inlineContent = true
	w.escape(value, false)
	return w
}

// Comment writes <!--value--> without spacing; "--" becomes "- -" and a
// trailing "-" gets a space.
func (w *XMLWriter) Comment(value string) *XMLWriter {
	w.closePendingStartTag(true)
	w.indent(len(w.openTags))
	w.out.WriteString("<!--")
	safe := strings.ReplaceAll(value, "--", "- -")
	if strings.HasSuffix(safe, "-") {
		safe += " "
	}
	w.out.WriteString(safe)
	w.out.WriteString("-->")
	w.newline()
	w.inlineContent = false
	return w
}

// ProcessingInstruction writes <?target data?>, without the space when data
// is empty.
func (w *XMLWriter) ProcessingInstruction(target, data string) *XMLWriter {
	w.closePendingStartTag(true)
	w.indent(len(w.openTags))
	w.out.WriteString("<?")
	w.out.WriteString(target)
	if data != "" {
		w.out.WriteByte(' ')
		w.out.WriteString(data)
	}
	w.out.WriteString("?>")
	w.newline()
	w.inlineContent = false
	return w
}

// CData writes a CDATA section; "]]>" is split across two sections.
func (w *XMLWriter) CData(value string) *XMLWriter {
	w.closePendingStartTag(true)
	w.indent(len(w.openTags))
	w.out.WriteString("<![CDATA[")
	w.out.WriteString(strings.ReplaceAll(value, "]]>", "]]]]><![CDATA[>"))
	w.out.WriteString("]]>")
	w.newline()
	w.inlineContent = false
	return w
}

// Raw writes markup verbatim, without indentation.
func (w *XMLWriter) Raw(markup string) *XMLWriter {
	w.closePendingStartTag(true)
	w.out.WriteString(markup)
	w.inlineContent = false
	return w
}

// EndElement closes the current element, as a self-closing tag when it is
// empty. It panics when no element is open.
func (w *XMLWriter) EndElement() *XMLWriter {
	if len(w.openTags) == 0 {
		throw(illegalState, "No element to close.")
	}
	name := w.openTags[len(w.openTags)-1]
	w.openTags = w.openTags[:len(w.openTags)-1]
	if w.pendingClose {
		w.out.WriteString("/>")
		w.newline()
		w.pendingClose = false
	} else {
		if !w.inlineContent {
			w.indent(len(w.openTags))
		}
		w.out.WriteString("</")
		w.out.WriteString(name)
		w.out.WriteByte('>')
		w.newline()
	}
	w.inlineContent = false
	return w
}

// XML returns the document; it panics when an element is still open.
func (w *XMLWriter) XML() string {
	if len(w.openTags) != 0 {
		throw(illegalState, "Some elements were left open: ["+strings.Join(w.openTags, ", ")+"]")
	}
	return w.out.String()
}

// RawXML returns the markup written so far, open elements allowed.
func (w *XMLWriter) RawXML() string { return w.out.String() }

func (w *XMLWriter) closePendingStartTag(openingChild bool) {
	if !w.pendingClose {
		return
	}
	w.out.WriteByte('>')
	if openingChild {
		w.newline()
	}
	w.pendingClose = false
}

func (w *XMLWriter) newline() {
	if w.indentSpaces > 0 {
		w.out.WriteByte('\n')
	}
}

func (w *XMLWriter) indent(level int) {
	if w.indentSpaces <= 0 {
		return
	}
	w.out.WriteString(strings.Repeat(" ", level*w.indentSpaces))
}

// escape writes s with "&" and "<" escaped, and '"' too in attributes.
func (w *XMLWriter) escape(s string, attribute bool) {
	start := 0
	for i := 0; i < len(s); i++ {
		var rep string
		switch s[i] {
		case '&':
			rep = "&amp;"
		case '<':
			rep = "&lt;"
		case '"':
			if !attribute {
				continue
			}
			rep = "&quot;"
		default:
			continue
		}
		w.out.WriteString(s[start:i])
		w.out.WriteString(rep)
		start = i + 1
	}
	w.out.WriteString(s[start:])
}
