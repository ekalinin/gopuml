// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/atmp/SvgOption.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svg

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/security"
)

// SvgOption holds the settings of an SVG document. As upstream, the With
// methods change the option and return it.
//
// Upstream's null strings are empty strings here: an empty hover color, link
// target, font or interactive base name means none. The title is the
// exception, since an empty title still writes a <title> element.
//
// Two settings are not in upstream's class, where they come from process
// state: the security profile (upstream's SecurityUtils, INSECURE inlines
// SVG images instead of embedding them as data URLs) and whether
// "javascript:" links are kept (upstream's PLANTUML_ALLOW_JAVASCRIPT_IN_LINK
// environment variable).
type SvgOption struct {
	lengthAdjust            klimt.LengthAdjust
	preserveAspectRatio     string
	hover                   string
	svgDimensionStyle       bool
	minDim                  geom.XDimension2D
	backcolor               color.HColor
	scale                   float64
	decimal                 int
	colorMapper             *color.ColorMapper
	linkTarget              string
	font                    string
	title                   string
	hasTitle                bool
	interactiveBaseFilename string
	rootAttributes          []xmlAttr
	options                 *klimt.ConfigurationStore[klimt.OptionKey]
	securityProfile         security.Profile
	allowJavascriptInLink   bool
}

// SvgOptionBasic returns the default option: spacing length adjustment,
// preserveAspectRatio "none", dimension style, scale 1, 3 decimals,
// identity color mapper, LEGACY security profile.
func SvgOptionBasic() *SvgOption {
	return &SvgOption{lengthAdjust: klimt.LengthAdjustDefault(), preserveAspectRatio: "none",
		svgDimensionStyle: true, minDim: geom.NewXDimension2D(0, 0), scale: 1.0, decimal: 3,
		colorMapper: color.IdentityMapper, securityProfile: security.Legacy}
}

// WithConfigurationStore sets the "!option" values (SVG title and
// description); nil means none.
func (o *SvgOption) WithConfigurationStore(options *klimt.ConfigurationStore[klimt.OptionKey]) *SvgOption {
	o.options = options
	return o
}

func (o *SvgOption) valueFromOptions(key klimt.OptionKey) (string, bool) {
	if o.options == nil {
		return "", false
	}
	return o.options.Value(key)
}

// Title returns the document title: the SVG_TITLE option, else the title
// set by WithTitle; ok is false when there is none.
func (o *SvgOption) Title() (title string, ok bool) {
	if v, ok := o.valueFromOptions(klimt.OptionSvgTitle); ok {
		return v, true
	}
	return o.title, o.hasTitle
}

// Desc returns the SVG_DESC option; ok is false when it is not set.
func (o *SvgOption) Desc() (desc string, ok bool) { return o.valueFromOptions(klimt.OptionSvgDesc) }

// WithTitle sets the title to the lines of a display joined by newlines,
// when the display is not empty.
func (o *SvgOption) WithTitle(titleDisplay *klimt.Display) *SvgOption {
	if titleDisplay.Size() > 0 {
		lines := make([]string, 0, titleDisplay.Size())
		for _, cs := range titleDisplay.AsList() {
			lines = append(lines, cs.String())
		}
		o.title = strings.Join(lines, "\n")
		o.hasTitle = true
	}
	return o
}

// WithTitleText sets the title, as WithTitle does with a display of the
// lines of title.
func (o *SvgOption) WithTitleText(title string) *SvgOption {
	o.title = title
	o.hasTitle = true
	return o
}

// InteractiveBaseFilename returns the base name of the interactive assets,
// "" when the document is not interactive.
func (o *SvgOption) InteractiveBaseFilename() string { return o.interactiveBaseFilename }

// WithInteractive makes the document interactive with the assets of a base
// name ("default" or "sequencediagram").
func (o *SvgOption) WithInteractive(interactiveBaseFilename string) *SvgOption {
	o.interactiveBaseFilename = interactiveBaseFilename
	return o
}

// WithLengthAdjust sets the text length adjustment.
func (o *SvgOption) WithLengthAdjust(lengthAdjust klimt.LengthAdjust) *SvgOption {
	o.lengthAdjust = lengthAdjust
	return o
}

// WithPreserveAspectRatio sets the root preserveAspectRatio.
func (o *SvgOption) WithPreserveAspectRatio(preserveAspectRatio string) *SvgOption {
	o.preserveAspectRatio = preserveAspectRatio
	return o
}

// WithHoverPathColorRGB sets the stroke color of hovered paths; "" means
// none.
func (o *SvgOption) WithHoverPathColorRGB(hover string) *SvgOption {
	o.hover = hover
	return o
}

// WithSvgDimensionStyle sets whether the root gets style, width and height.
func (o *SvgOption) WithSvgDimensionStyle(svgDimensionStyle bool) *SvgOption {
	o.svgDimensionStyle = svgDimensionStyle
	return o
}

// WithMinDim sets the minimum canvas size.
func (o *SvgOption) WithMinDim(minDim geom.XDimension2D) *SvgOption {
	o.minDim = minDim
	return o
}

// WithBackcolor sets the background; nil means none.
func (o *SvgOption) WithBackcolor(backcolor color.HColor) *SvgOption {
	o.backcolor = backcolor
	return o
}

// WithScale sets the scale applied to every coordinate.
func (o *SvgOption) WithScale(scale float64) *SvgOption {
	o.scale = scale
	return o
}

// WithDecimal sets the number of decimals of coordinates.
func (o *SvgOption) WithDecimal(decimal int) *SvgOption {
	o.decimal = decimal
	return o
}

// WithColorMapper sets the color mapper (dark mode, monochrome...).
func (o *SvgOption) WithColorMapper(colorMapper *color.ColorMapper) *SvgOption {
	o.colorMapper = colorMapper
	return o
}

// WithLinkTarget sets the target attribute of links; "" writes an empty
// target, as upstream's null.
func (o *SvgOption) WithLinkTarget(linkTarget string) *SvgOption {
	o.linkTarget = linkTarget
	return o
}

// WithFont sets the SVG_FONT pragma value; "" means none.
func (o *SvgOption) WithFont(font string) *SvgOption {
	o.font = font
	return o
}

// RootAttributes returns the extra attributes of the root, in insertion
// order, as name and value pairs.
func (o *SvgOption) RootAttributes() [][2]string {
	result := make([][2]string, len(o.rootAttributes))
	for i, a := range o.rootAttributes {
		result[i] = [2]string{a.name, a.value}
	}
	return result
}

// WithRootAttribute adds an attribute to the root; setting a name again
// keeps its position.
func (o *SvgOption) WithRootAttribute(key, value string) *SvgOption {
	for i := range o.rootAttributes {
		if o.rootAttributes[i].name == key {
			o.rootAttributes[i].value = value
			return o
		}
	}
	o.rootAttributes = append(o.rootAttributes, xmlAttr{name: key, value: value})
	return o
}

// WithSecurityProfile sets the security profile; INSECURE inlines SVG
// images.
func (o *SvgOption) WithSecurityProfile(profile security.Profile) *SvgOption {
	o.securityProfile = profile
	return o
}

// WithAllowJavascriptInLink keeps "javascript:" links, which are emptied by
// default.
func (o *SvgOption) WithAllowJavascriptInLink(allow bool) *SvgOption {
	o.allowJavascriptInLink = allow
	return o
}

// PreserveAspectRatio returns the root preserveAspectRatio.
func (o *SvgOption) PreserveAspectRatio() string { return o.preserveAspectRatio }

// IsInteractive reports whether the interactive assets are embedded.
func (o *SvgOption) IsInteractive() bool { return o.interactiveBaseFilename != "" }

// LengthAdjust returns the text length adjustment.
func (o *SvgOption) LengthAdjust() klimt.LengthAdjust { return o.lengthAdjust }

// Hover returns the hovered path color, "" for none.
func (o *SvgOption) Hover() string { return o.hover }

// SvgDimensionStyle reports whether the root gets style, width and height.
func (o *SvgOption) SvgDimensionStyle() bool { return o.svgDimensionStyle }

// MinDim returns the minimum canvas size.
func (o *SvgOption) MinDim() geom.XDimension2D { return o.minDim }

// Backcolor returns the background, nil for none.
func (o *SvgOption) Backcolor() color.HColor { return o.backcolor }

// Scale returns the scale.
func (o *SvgOption) Scale() float64 { return o.scale }

// Decimal returns the number of decimals.
func (o *SvgOption) Decimal() int { return o.decimal }

// ColorMapper returns the color mapper.
func (o *SvgOption) ColorMapper() *color.ColorMapper { return o.colorMapper }

// LinkTarget returns the link target, "" for none.
func (o *SvgOption) LinkTarget() string { return o.linkTarget }

// Font returns the SVG_FONT value, "" for none.
func (o *SvgOption) Font() string { return o.font }

// SecurityProfile returns the security profile.
func (o *SvgOption) SecurityProfile() security.Profile { return o.securityProfile }

// AllowJavascriptInLink reports whether "javascript:" links are kept.
func (o *SvgOption) AllowJavascriptInLink() bool { return o.allowJavascriptInLink }
