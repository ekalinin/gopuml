// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/classdiagram/command/CommandHideShowByGender.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/classdiagram/command/CommandHideShowByVisibility.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package classdiagram

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/uml"
)

var (
	_ uml.GenderHider     = (*ClassDiagram)(nil)
	_ uml.VisibilityHider = (*ClassDiagram)(nil)
)

func emptyByGender(portion uml.EntityPortion) abel.EntityGender {
	switch portion {
	case uml.PortionMethod:
		return abel.GenderEmptyMethods()
	case uml.PortionField:
		return abel.GenderEmptyFields()
	case uml.PortionMember:
		uml.Throw(uml.IllegalArgumentException, "")
	}
	return abel.GenderAll()
}

// genderTypes are the type keywords of "hide <type> <portion>".
var genderTypes = []struct {
	name string
	typ  abel.LeafType
}{
	{"class", abel.LeafTypeClass}, {"object", abel.LeafTypeObject}, {"interface", abel.LeafTypeInterface},
	{"enum", abel.LeafTypeEnum}, {"abstract", abel.LeafTypeAbstractClass},
	{"annotation", abel.LeafTypeAnnotation}, {"protocol", abel.LeafTypeProtocol}, {"struct", abel.LeafTypeStruct},
	{"exception", abel.LeafTypeException}, {"metaclass", abel.LeafTypeMetaclass},
	{"stereotype", abel.LeafTypeStereotype}, {"dataclass", abel.LeafTypeDataclass},
	{"record", abel.LeafTypeRecord},
}

// HideShowByGender implements uml.GenderHider: upstream's
// CommandHideShowByGender.executeClassDiagram.
func (d *ClassDiagram) HideShowByGender(arg *regex.Result) command.Result {
	portion := uml.GenderPortion(arg.Str("PORTION", 0))
	var gender abel.EntityGender
	arg1, ok := arg.Get("GENDER", 0)
	found := false
	if !ok {
		gender, found = abel.GenderAll(), true
	} else {
		for _, g := range genderTypes {
			if uml.EqualsIgnoreCase(arg1, g.name) {
				gender, found = abel.GenderByEntityType(g.typ), true
				break
			}
		}
	}
	if !found {
		if strings.HasPrefix(arg1, "<<") {
			gender = abel.GenderByStereotype(arg1)
		} else {
			arg1 = uml.RemoveQuotes(arg1)
			quark := d.QuarkInContext(true, d.CleanID(arg1))
			if portion == uml.PortionMethod {
				gender = abel.GenderByClassName(arg1)
			} else {
				entity := quark.Data()
				if entity == nil {
					return command.Error("No such element " + quark.Name())
				}
				gender = abel.GenderByEntityAlone(entity)
			}
		}
	}
	_, empty := arg.Get("EMPTY", 0)
	emptyMembers := empty && portion == uml.PortionMember
	if empty && !emptyMembers {
		gender = abel.GenderAnd(gender, emptyByGender(portion))
	}
	if !d.CurrentGroup().IsRoot() {
		gender = abel.GenderAnd(gender, abel.GenderByPackage(d.CurrentGroup()))
	}
	show := uml.EqualsIgnoreCase(arg.Str("COMMAND", 0), "show")
	if emptyMembers {
		d.HideOrShow(abel.GenderAnd(gender, emptyByGender(uml.PortionField)), uml.PortionField, show)
		d.HideOrShow(abel.GenderAnd(gender, emptyByGender(uml.PortionMethod)), uml.PortionMethod, show)
	} else {
		d.HideOrShow(gender, portion, show)
	}
	return command.Ok()
}

// HideShowByVisibility implements uml.VisibilityHider: upstream's
// CommandHideShowByVisibility.executeArgClass.
func (d *ClassDiagram) HideShowByVisibility(arg *regex.Result) command.Result {
	portion := visibilityPortion(arg.Str("PORTION", 0))
	var visibilities abel.VisibilitySet
	for _, token := range strings.FieldsFunc(uml.LowerCase(arg.Str("VISIBILITY", 0)), func(r rune) bool {
		return r == ' ' || r == ','
	}) {
		visibilities = addVisibilities(token, portion, visibilities)
	}
	d.HideOrShowVisibilityModifier(visibilities, uml.EqualsIgnoreCase(arg.Str("COMMAND", 0), "show"))
	return command.Ok()
}

func addVisibilities(token string, portion uml.EntityPortion, result abel.VisibilitySet) abel.VisibilitySet {
	field := portion == uml.PortionMember || portion == uml.PortionField
	method := portion == uml.PortionMember || portion == uml.PortionMethod
	add := func(ok bool, v klimt.VisibilityModifier) {
		if ok {
			result = result.With(v)
		}
	}
	switch token {
	case "public":
		add(field, klimt.VisibilityPublicField)
		add(method, klimt.VisibilityPublicMethod)
	case "private":
		add(field, klimt.VisibilityPrivateField)
		add(method, klimt.VisibilityPrivateMethod)
	case "protected":
		add(field, klimt.VisibilityProtectedField)
		add(method, klimt.VisibilityProtectedMethod)
	case "package":
		add(field, klimt.VisibilityPackagePrivateField)
		add(method, klimt.VisibilityPackagePrivateMethod)
	}
	return result
}

// visibilityPortion is CommandHideShowByVisibility.getEntityPortion: the
// portion named by the first three letters.
func visibilityPortion(s string) uml.EntityPortion {
	if uml.JavaLength(s) < 3 {
		uml.Throw(uml.StringIndexOutOfBoundsException, "")
	}
	switch uml.LowerCase(s[:3]) {
	case "met":
		return uml.PortionMethod
	case "mem":
		return uml.PortionMember
	case "att", "fie":
		return uml.PortionField
	}
	uml.Throw(uml.IllegalArgumentException, "")
	return 0
}
