// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/abel/EntityGender.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/abel/EntityGenderUtils.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package abel

import (
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

// EntityGender selects entities, for "hide" and "show" on portions.
type EntityGender interface {
	// Contains reports whether the entity is selected.
	Contains(test *Entity) bool
	// Gender returns the name of the selection (a type, a stereotype, a
	// uid or a class name); ok is false for the other selections (null
	// upstream).
	Gender() (gender string, ok bool)
}

type gender struct {
	contains func(test *Entity) bool
	name     string
	hasName  bool
}

func (g gender) Contains(test *Entity) bool { return g.contains(test) }

func (g gender) Gender() (string, bool) { return g.name, g.hasName }

// GenderByEntityType selects the entities of a type.
func GenderByEntityType(typ LeafType) EntityGender {
	return gender{contains: func(test *Entity) bool { return test.LeafType() == typ }, name: typ.String(),
		hasName: true}
}

// GenderByEntityAlone selects one entity, by uid.
func GenderByEntityAlone(entity *Entity) EntityGender {
	return gender{contains: func(test *Entity) bool { return test.UID() == entity.UID() }, name: entity.UID(),
		hasName: true}
}

// GenderByStereotype selects the entities with a stereotype label, such as
// "<<service>>".
func GenderByStereotype(stereotype string) EntityGender {
	return gender{contains: func(test *Entity) bool {
		if test.Stereotype() == nil {
			return false
		}
		for _, label := range test.Stereotype().Labels(klimt.GuillemetDoubleComparator()) {
			if label == stereotype {
				return true
			}
		}
		return false
	}, name: stereotype, hasName: true}
}

// GenderByPackage selects the entities directly in a group. The root
// panics, as upstream throws IllegalArgumentException.
func GenderByPackage(group *Entity) EntityGender {
	if group.IsRoot() {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	return gender{contains: func(test *Entity) bool {
		if test.ParentContainer().IsRoot() {
			return false
		}
		return group == test.ParentContainer()
	}}
}

// GenderAnd selects the entities that both selections select.
func GenderAnd(g1, g2 EntityGender) EntityGender {
	return gender{contains: func(test *Entity) bool { return g1.Contains(test) && g2.Contains(test) }}
}

// GenderAll selects every entity.
func GenderAll() EntityGender { return gender{contains: func(*Entity) bool { return true }} }

// GenderEmptyMethods selects the entities without methods to display.
func GenderEmptyMethods() EntityGender {
	return gender{contains: func(test *Entity) bool { return test.Bodier().MethodsToDisplay().Size() == 0 }}
}

// GenderEmptyFields selects the entities without fields to display.
func GenderEmptyFields() EntityGender {
	return gender{contains: func(test *Entity) bool { return test.Bodier().FieldsToDisplay().Size() == 0 }}
}

// GenderByClassName selects the entities of a name.
func GenderByClassName(className string) EntityGender {
	return gender{contains: func(test *Entity) bool { return className == test.Name() }, name: className,
		hasName: true}
}
