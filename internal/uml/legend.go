// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/TitledDiagram.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package uml

// LegendSetter is a diagram with its own legend placement, as upstream's
// overrides of TitledDiagram.setLegend (class diagrams put the legend of a
// package in the package). Every diagram embedding Base is one; the legend
// commands call the most derived SetLegend.
type LegendSetter interface {
	SetLegend(legend DisplayPositioned)
}

// setLegend calls the SetLegend of the diagram, or of its Base.
func setLegend[D Diagram](d D, legend DisplayPositioned) {
	if s, ok := any(d).(LegendSetter); ok {
		s.SetLegend(legend)
		return
	}
	d.Titled().SetLegend(legend)
}
