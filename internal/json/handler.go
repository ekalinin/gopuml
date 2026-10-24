// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/json/JsonHandler.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package json

// Handler receives the events of a [Parser] while it reads the input.
//
// Implementations embed [BaseHandler], whose methods do nothing, and
// override the events they are interested in. They can call
// BaseHandler.Location to get the current position of the parser. The Start
// methods are called while the location points to the first character of
// the parsed element; the End methods are called while it points to the
// character that directly follows the last character of the element:
//
//	["lorem ipsum"]
//	 ^            ^
//	 startString  endString
//
// StartArray and StartObject may return a handler value of type A or O for
// the new array or object. It is then passed to every later event of this
// array or object, so that implementations can keep track of its elements.
type Handler[A, O any] interface {
	// StartNull is called at the first character of a null literal.
	StartNull()
	// EndNull is called after the last character of a null literal.
	EndNull()
	// StartBoolean is called at the first character of true or false.
	StartBoolean()
	// EndBoolean is called after the last character of true or false.
	EndBoolean(value bool)
	// StartString is called at the opening double quote of a string.
	StartString()
	// EndString is called after the closing double quote of a string, with
	// the unescaped string.
	EndString(s string)
	// StartNumber is called at the first character of a number.
	StartNumber()
	// EndNumber is called after the last character of a number, with its
	// text.
	EndNumber(s string)
	// StartArray is called at the opening square bracket of an array and
	// returns the handler value for this array.
	StartArray() A
	// EndArray is called after the closing square bracket of an array.
	EndArray(array A)
	// StartArrayValue is called at the first character of an array element,
	// before the Start method of the element itself.
	StartArrayValue(array A)
	// EndArrayValue is called after the last character of an array element,
	// after the End method of the element itself.
	EndArrayValue(array A)
	// StartObject is called at the opening curly brace of an object and
	// returns the handler value for this object.
	StartObject() O
	// EndObject is called after the closing curly brace of an object.
	EndObject(object O)
	// StartObjectName is called at the opening double quote of a member
	// name.
	StartObjectName(object O)
	// EndObjectName is called after the closing double quote of a member
	// name.
	EndObjectName(object O, name string)
	// StartObjectValue is called at the first character of a member value,
	// before the Start method of the value itself.
	StartObjectValue(object O, name string)
	// EndObjectValue is called after the last character of a member value,
	// after the End method of the value itself.
	EndObjectValue(object O, name string)

	setParser(p locator)
}

// locator gives the current position of a parser.
type locator interface {
	location() Location
}

// BaseHandler implements every [Handler] event as a no-op. Embed it in a
// handler and override the events of interest.
type BaseHandler[A, O any] struct {
	parser locator
}

func (h *BaseHandler[A, O]) setParser(p locator) { h.parser = p }

// Location returns the current position of the parser this handler is
// attached to, or the zero Location if there is none.
func (h *BaseHandler[A, O]) Location() Location {
	if h.parser == nil {
		return Location{}
	}
	return h.parser.location()
}

func (h *BaseHandler[A, O]) StartNull()                             {}
func (h *BaseHandler[A, O]) EndNull()                               {}
func (h *BaseHandler[A, O]) StartBoolean()                          {}
func (h *BaseHandler[A, O]) EndBoolean(value bool)                  {}
func (h *BaseHandler[A, O]) StartString()                           {}
func (h *BaseHandler[A, O]) EndString(s string)                     {}
func (h *BaseHandler[A, O]) StartNumber()                           {}
func (h *BaseHandler[A, O]) EndNumber(s string)                     {}
func (h *BaseHandler[A, O]) StartArray() (array A)                  { return array }
func (h *BaseHandler[A, O]) EndArray(array A)                       {}
func (h *BaseHandler[A, O]) StartArrayValue(array A)                {}
func (h *BaseHandler[A, O]) EndArrayValue(array A)                  {}
func (h *BaseHandler[A, O]) StartObject() (object O)                { return object }
func (h *BaseHandler[A, O]) EndObject(object O)                     {}
func (h *BaseHandler[A, O]) StartObjectName(object O)               {}
func (h *BaseHandler[A, O]) EndObjectName(object O, name string)    {}
func (h *BaseHandler[A, O]) StartObjectValue(object O, name string) {}
func (h *BaseHandler[A, O]) EndObjectValue(object O, name string)   {}
