// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/Darken.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/HslColor.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/IsDark.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/IsLight.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/Lighten.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/ReverseColor.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/ReverseHsluvColor.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/text"
)

// getColor is HColorSet.instance().getColor, with upstream's "No such
// color" error.
func getColor(s string, location text.StringLocated) (color.HColor, error) {
	c, err := color.Instance().GetColor(s)
	if err != nil {
		return nil, newEaterException("No such color", location)
	}
	return c, nil
}

// darken is %darken(color, ratio): the HSL luminance decreased by ratio
// percent of itself.
type darken struct{ simpleReturnFunction }

func (darken) Signature() TFunctionSignature { return NewTFunctionSignature("%darken", 2) }

func (darken) CanCover(nbArg int, _ []string) bool { return nbArg == 2 }

func (darken) ExecuteReturnFunction(_ *TContext, _ TMemory, location text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	ratio := int(values[1].ToInt())
	c, err := getColor(values[0].String(), location)
	if err != nil {
		return nil, err
	}
	return TValueFromString(c.Darken(ratio).AsString()), nil
}

// lighten is %lighten(color, ratio): the HSL luminance increased by ratio
// percent of itself.
type lighten struct{ simpleReturnFunction }

func (lighten) Signature() TFunctionSignature { return NewTFunctionSignature("%lighten", 2) }

func (lighten) CanCover(nbArg int, _ []string) bool { return nbArg == 2 }

func (lighten) ExecuteReturnFunction(_ *TContext, _ TMemory, location text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	ratio := int(values[1].ToInt())
	c, err := getColor(values[0].String(), location)
	if err != nil {
		return nil, err
	}
	return TValueFromString(c.Lighten(ratio).AsString()), nil
}

// isDark is %is_dark(color).
type isDark struct{ simpleReturnFunction }

func (isDark) Signature() TFunctionSignature { return NewTFunctionSignature("%is_dark", 1) }

func (isDark) CanCover(nbArg int, _ []string) bool { return nbArg == 1 }

func (isDark) ExecuteReturnFunction(_ *TContext, _ TMemory, location text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	c, err := getColor(values[0].String(), location)
	if err != nil {
		return nil, err
	}
	return TValueFromBoolean(c.IsDark()), nil
}

// isLight is %is_light(color).
type isLight struct{ simpleReturnFunction }

func (isLight) Signature() TFunctionSignature { return NewTFunctionSignature("%is_light", 1) }

func (isLight) CanCover(nbArg int, _ []string) bool { return nbArg == 1 }

func (isLight) ExecuteReturnFunction(_ *TContext, _ TMemory, location text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	c, err := getColor(values[0].String(), location)
	if err != nil {
		return nil, err
	}
	return TValueFromBoolean(!c.IsDark()), nil
}

// reverseColor is %reverse_color(color): the RGB negative.
type reverseColor struct{ simpleReturnFunction }

func (reverseColor) Signature() TFunctionSignature { return NewTFunctionSignature("%reverse_color", 1) }

func (reverseColor) CanCover(nbArg int, _ []string) bool { return nbArg == 1 }

func (reverseColor) ExecuteReturnFunction(_ *TContext, _ TMemory, location text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	c, err := getColor(values[0].String(), location)
	if err != nil {
		return nil, err
	}
	return TValueFromString(c.Reverse().AsString()), nil
}

// reverseHsluvColor is %reverse_hsluv_color(color): the HSLuv lightness
// reversed.
type reverseHsluvColor struct{ simpleReturnFunction }

func (reverseHsluvColor) Signature() TFunctionSignature {
	return NewTFunctionSignature("%reverse_hsluv_color", 1)
}

func (reverseHsluvColor) CanCover(nbArg int, _ []string) bool { return nbArg == 1 }

func (reverseHsluvColor) ExecuteReturnFunction(_ *TContext, _ TMemory, location text.StringLocated,
	values []*TValue, _ map[string]*TValue) (*TValue, error) {
	c, err := getColor(values[0].String(), location)
	if err != nil {
		return nil, err
	}
	return TValueFromString(c.ReverseHsluv().AsString()), nil
}

// hslColor is %hsl_color(h, s, l[, a]): the color of a hue in degrees,
// saturation and luminance in percent, and alpha in percent.
type hslColor struct{ simpleReturnFunction }

func (hslColor) Signature() TFunctionSignature { return NewTFunctionSignature("%hsl_color", 3) }

func (hslColor) CanCover(nbArg int, _ []string) bool { return nbArg == 3 || nbArg == 4 }

func (hslColor) ExecuteReturnFunction(_ *TContext, _ TMemory, _ text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	h := float32(values[0].ToInt())
	s := float32(values[1].ToInt())
	l := float32(values[2].ToInt())
	if len(values) == 3 {
		rgb := color.NewHSLColorHSL(h, s, l).RGB()
		return TValueFromString(color.Simple(rgb).AsString()), nil
	}
	a := float32(float64(values[3].ToInt()) / 100.0)
	rgb := color.NewHSLColorHSLA(h, s, l, a).RGB()
	return TValueFromString(color.Simple(rgb).AsString()), nil
}
