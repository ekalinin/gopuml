// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/style/FromSkinparamToStyle.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"strings"
)

// skinparamData is one conversion of a skinparam: the property and the
// style names it sets.
type skinparamData struct {
	styleNames   []SName
	propertyName PName
}

// String is upstream's toString, for example
// "[participant.clickable]BackGroundColor". As upstream, without style names
// the "[" is removed instead of a final ".".
func (d skinparamData) String() string {
	var sb strings.Builder
	sb.WriteString("[")
	for _, s := range d.styleNames {
		sb.WriteString(s.String())
		sb.WriteString(".")
	}
	str := sb.String()
	str = str[:len(str)-1]
	return str + "]" + d.propertyName.String()
}

// skinparamKnowledge maps the lower-case skinparam names to their
// conversions: read-only data built at initialization.
var skinparamKnowledge = buildSkinparamKnowledge()

func buildSkinparamKnowledge() map[string][]skinparamData {
	k := map[string][]skinparamData{}
	addConvert := func(skinparam string, propertyName PName, styleNames ...SName) {
		skinparam = javaLower(skinparam)
		k[skinparam] = append(k[skinparam], skinparamData{styleNames: styleNames, propertyName: propertyName})
	}
	addConFont := func(skinparam string, styleNames ...SName) {
		addConvert(skinparam+"FontSize", PNameFontSize, styleNames...)
		addConvert(skinparam+"FontStyle", PNameFontStyle, styleNames...)
		addConvert(skinparam+"FontColor", PNameFontColor, styleNames...)
		addConvert(skinparam+"FontName", PNameFontName, styleNames...)
	}
	addMagic := func(sname SName) {
		cleanName := strings.ReplaceAll(sname.String(), "_", "")
		addConvert(cleanName+"BackgroundColor", PNameBackGroundColor, sname)
		addConvert(cleanName+"BorderColor", PNameLineColor, sname)
		addConvert(cleanName+"BorderThickness", PNameLineThickness, sname)
		addConvert(cleanName+"RoundCorner", PNameRoundCorner, sname)
		addConvert(cleanName+"DiagonalCorner", PNameDiagonalCorner, sname)
		addConvert(cleanName+"BorderStyle", PNameLineStyle, sname)
		addConFont(cleanName, sname)
		addConvert(cleanName+"Shadowing", PNameShadowing, sname)
		addConvert(cleanName+"StereotypeFontSize", PNameFontSize, SNameStereotype, sname)
		addConvert(cleanName+"StereotypeFontStyle", PNameFontStyle, SNameStereotype, sname)
		addConvert(cleanName+"StereotypeFontColor", PNameFontColor, SNameStereotype, sname)
		addConvert(cleanName+"StereotypeFontName", PNameFontName, SNameStereotype, sname)
	}

	addConvert("participantClickableBackgroundColor", PNameBackGroundColor, SNameParticipant, SNameClickable)
	addConvert("participantClickableBorderColor", PNameLineColor, SNameParticipant, SNameClickable)
	addMagic(SNameParticipant)
	addMagic(SNameBoundary)
	addMagic(SNameControl)
	addMagic(SNameCollections)
	addMagic(SNameActor)
	addMagic(SNameDatabase)
	addMagic(SNameEntity)

	addConFont("header", SNameDocument, SNameHeader)
	addConFont("footer", SNameDocument, SNameFooter)
	addConFont("caption", SNameDocument, SNameCaption)

	addConvert("defaultFontSize", PNameFontSize, SNameElement)

	addConvert("sequenceStereotypeFontSize", PNameFontSize, SNameStereotype)
	addConvert("sequenceStereotypeFontStyle", PNameFontStyle, SNameStereotype)
	addConvert("sequenceStereotypeFontColor", PNameFontColor, SNameStereotype)
	addConvert("sequenceStereotypeFontName", PNameFontName, SNameStereotype)

	addConvert("SequenceReferenceBorderColor", PNameLineColor, SNameReference)
	addConvert("SequenceReferenceBorderColor", PNameLineColor, SNameReferenceHeader)
	addConvert("SequenceReferenceBackgroundColor", PNameBackGroundColor, SNameReference)
	addConvert("sequenceReferenceHeaderBackgroundColor", PNameBackGroundColor, SNameReferenceHeader)
	addConFont("sequenceReference", SNameReference)
	addConFont("sequenceReference", SNameReferenceHeader)

	addConvert("sequenceGroupBorderThickness", PNameLineThickness, SNameGroup)
	addConvert("SequenceGroupBorderColor", PNameLineColor, SNameGroup)
	addConvert("SequenceGroupBorderColor", PNameLineColor, SNameGroupHeader)
	addConvert("SequenceGroupBackgroundColor", PNameBackGroundColor, SNameGroupHeader)
	addConFont("SequenceGroup", SNameGroup)
	addConFont("SequenceGroupHeader", SNameGroupHeader)

	addConvert("SequenceBoxBorderColor", PNameLineColor, SNameBox)
	addConvert("SequenceBoxBackgroundColor", PNameBackGroundColor, SNameBox)
	addConvert("SequenceBoxFontColor", PNameFontColor, SNameBox)

	addConvert("SequenceLifeLineBorderColor", PNameLineColor, SNameLifeLine)
	addConvert("SequenceLifeLineBackgroundColor", PNameBackGroundColor, SNameActivationBox)

	addConFont("sequenceDelay", SNameDelay)
	addConvert("sequenceDelayBorderColor", PNameLineColor, SNameDelay)

	addConvert("sequenceDividerBackgroundColor", PNameBackGroundColor, SNameSeparator)
	addConvert("sequenceDividerBorderColor", PNameLineColor, SNameSeparator)
	addConFont("sequenceDivider", SNameSeparator)
	addConvert("sequenceDividerBorderThickness", PNameLineThickness, SNameSeparator)

	addConvert("SequenceMessageAlignment", PNameHorizontalAlignment, SNameArrow)

	addConFont("note", SNameNote)
	addConvert("noteBorderThickness", PNameLineThickness, SNameNote)
	addConvert("noteBorderColor", PNameLineColor, SNameNote)
	addConvert("noteBackgroundColor", PNameBackGroundColor, SNameNote)

	addConvert("packageBackgroundColor", PNameBackGroundColor, SNameGroup)
	addConvert("packageBorderColor", PNameLineColor, SNameGroup)
	addMagic(SNamePackage)

	addConvert("PartitionBorderColor", PNameLineColor, SNameComposite)
	addConvert("PartitionBackgroundColor", PNameBackGroundColor, SNameComposite)
	addConFont("Partition", SNameComposite)

	addConvert("hyperlinkColor", PNameHyperLinkColor, SNameRoot)

	addConvert("activityStartColor", PNameBackGroundColor, SNameCircle, SNameStart)
	addConvert("activityEndColor", PNameLineColor, SNameCircle, SNameEnd)
	addConvert("activityStopColor", PNameLineColor, SNameCircle, SNameStop)
	addConvert("activityBarColor", PNameBackGroundColor, SNameActivityBar)
	addConvert("activityBorderColor", PNameLineColor, SNameActivity)
	addConvert("activityBorderThickness", PNameLineThickness, SNameActivity)
	addConvert("activityBackgroundColor", PNameBackGroundColor, SNameActivity)
	addConFont("activity", SNameActivity)

	addConvert("activityDiamondBackgroundColor", PNameBackGroundColor, SNameDiamond)
	addConvert("activityDiamondBorderColor", PNameLineColor, SNameDiamond)
	addConFont("activityDiamond", SNameDiamond)

	addConFont("arrow", SNameArrow)
	addConvert("arrowThickness", PNameLineThickness, SNameArrow)
	addConvert("arrowColor", PNameLineColor, SNameArrow)
	addConvert("arrowStyle", PNameLineStyle, SNameArrow)
	addConvert("arrowHeadColor", PNameHeadColor, SNameArrow)

	addConvert("defaulttextalignment", PNameHorizontalAlignment, SNameRoot)
	addConvert("defaultFontName", PNameFontName, SNameRoot)
	addConvert("defaultFontColor", PNameFontColor, SNameRoot)

	addConFont("SwimlaneTitle", SNameSwimlane)
	addConvert("SwimlaneTitleBackgroundColor", PNameBackGroundColor, SNameSwimlane)
	addConvert("SwimlaneBorderColor", PNameLineColor, SNameSwimlane)
	addConvert("SwimlaneBorderThickness", PNameLineThickness, SNameSwimlane)

	addConvert("roundCorner", PNameRoundCorner, SNameRoot)

	addConvert("titleBorderThickness", PNameLineThickness, SNameTitle)
	addConvert("titleBorderColor", PNameLineColor, SNameTitle)
	addConvert("titleBackgroundColor", PNameBackGroundColor, SNameTitle)
	addConvert("titleBorderRoundCorner", PNameRoundCorner, SNameTitle)
	addConFont("title", SNameDocument, SNameTitle)

	addConvert("legendBorderThickness", PNameLineThickness, SNameLegend)
	addConvert("legendBorderColor", PNameLineColor, SNameLegend)
	addConvert("legendBackgroundColor", PNameBackGroundColor, SNameLegend)
	addConvert("legendBorderRoundCorner", PNameRoundCorner, SNameLegend)
	addConFont("legend", SNameLegend)

	addConvert("noteTextAlignment", PNameHorizontalAlignment, SNameNote)

	addConvert("BackgroundColor", PNameBackGroundColor, SNameDocument)

	addConvert("classBackgroundColor", PNameBackGroundColor, SNameElement, SNameClass)
	addConvert("classBorderColor", PNameLineColor, SNameElement, SNameClass)
	addConvert("classFontSize", PNameFontSize, SNameElement, SNameClass, SNameHeader)
	addConvert("classFontStyle", PNameFontStyle, SNameElement, SNameClass, SNameHeader)
	addConvert("classFontColor", PNameFontColor, SNameElement, SNameClass, SNameHeader)
	addConvert("classFontName", PNameFontName, SNameElement, SNameClass, SNameHeader)
	addConvert("classAttributeFontSize", PNameFontSize, SNameElement, SNameClass)
	addConvert("classAttributeFontStyle", PNameFontStyle, SNameElement, SNameClass)
	addConvert("classAttributeFontColor", PNameFontColor, SNameElement, SNameClass)
	addConvert("classAttributeFontName", PNameFontName, SNameElement, SNameClass)
	addConvert("classBorderThickness", PNameLineThickness, SNameElement, SNameClass)
	addConvert("classHeaderBackgroundColor", PNameBackGroundColor, SNameElement, SNameClass, SNameHeader)

	addConvert("objectBackgroundColor", PNameBackGroundColor, SNameObject)
	addConvert("objectBorderColor", PNameLineColor, SNameObject)
	addConFont("object", SNameObject)
	addConFont("objectAttribute", SNameObject)
	addConvert("objectBorderThickness", PNameLineThickness, SNameObject)

	addConvert("stateBackgroundColor", PNameBackGroundColor, SNameState)
	addConvert("stateBorderColor", PNameLineColor, SNameState)
	addConFont("state", SNameState)
	addConFont("stateAttribute", SNameState)
	addConvert("stateBorderThickness", PNameLineThickness, SNameState)

	addMagic(SNameAgent)
	addMagic(SNameArtifact)
	addMagic(SNameCard)
	addMagic(SNameInterface)
	addMagic(SNameCloud)
	addMagic(SNameComponent)
	addMagic(SNameFile)
	addMagic(SNameFolder)
	addMagic(SNameFrame)
	addMagic(SNameHexagon)
	addMagic(SNameNode)
	addMagic(SNamePerson)
	addMagic(SNameQueue)
	addMagic(SNameRectangle)
	addMagic(SNameStack)
	addMagic(SNameStorage)
	addMagic(SNameUsecase)
	addMagic(SNameMap)
	addMagic(SNameArchimate)

	addMagic(SNameHnote)
	addMagic(SNameRnote)

	addConvert("IconPrivateColor", PNameLineColor, SNameVisibilityIcon, SNamePrivate)
	addConvert("IconPrivateBackgroundColor", PNameBackGroundColor, SNameVisibilityIcon, SNamePrivate)
	addConvert("IconPackageColor", PNameLineColor, SNameVisibilityIcon, SNamePackage)
	addConvert("IconPackageBackgroundColor", PNameBackGroundColor, SNameVisibilityIcon, SNamePackage)
	addConvert("IconProtectedColor", PNameLineColor, SNameVisibilityIcon, SNameProtected)
	addConvert("IconProtectedBackgroundColor", PNameBackGroundColor, SNameVisibilityIcon, SNameProtected)
	addConvert("IconPublicColor", PNameLineColor, SNameVisibilityIcon, SNamePublic)
	addConvert("IconPublicBackgroundColor", PNameBackGroundColor, SNameVisibilityIcon, SNamePublic)

	addConvert("MinClassWidth", PNameMinimumWidth)

	addConvert("wrapWidth", PNameMaximumWidth, SNameElement)
	addConvert("HyperlinkUnderline", PNameHyperlinkUnderlineThickness, SNameElement)

	addConvert("StereotypeAlignment", PNameHorizontalAlignment, SNameStereotype)

	addConvert("stereotypeABackgroundColor", PNameBackGroundColor, SNameSpotAbstractClass)
	addConvert("stereotypeABorderColor", PNameLineColor, SNameSpotAbstractClass)
	addConvert("stereotypeCBackgroundColor", PNameBackGroundColor, SNameSpotClass)
	addConvert("stereotypeCBorderColor", PNameLineColor, SNameSpotClass)
	addConvert("stereotypeEBackgroundColor", PNameBackGroundColor, SNameSpotEnum)
	addConvert("stereotypeEBorderColor", PNameLineColor, SNameSpotEnum)
	addConvert("stereotypeIBackgroundColor", PNameBackGroundColor, SNameSpotInterface)
	addConvert("stereotypeIBorderColor", PNameLineColor, SNameSpotInterface)
	addConvert("stereotypeNBackgroundColor", PNameBackGroundColor, SNameSpotAnnotation)
	addConvert("stereotypeNBorderColor", PNameLineColor, SNameSpotAnnotation)
	addConvert("stereotypeRBackgroundColor", PNameBackGroundColor, SNameSpotRecord)
	addConvert("stereotypeRBorderColor", PNameLineColor, SNameSpotRecord)
	addConvert("stereotypeDBackgroundColor", PNameBackGroundColor, SNameSpotDataClass)
	addConvert("stereotypeDBorderColor", PNameLineColor, SNameSpotDataClass)
	return k
}

// FromSkinparamToStyle converts one skinparam ("key<<stereo>> value") to
// the styles it stands for.
type FromSkinparamToStyle struct {
	styles    []*Style
	stereo    string
	hasStereo bool
	key       string
}

// NewFromSkinparamToStyle returns the converter of a skinparam key; a key
// with "<<" carries a stereotype.
func NewFromSkinparamToStyle(key string) *FromSkinparamToStyle {
	c := &FromSkinparamToStyle{}
	if strings.Contains(key, "<<") {
		tokens := javaTokenize(key, "<>")
		if len(tokens) == 0 {
			throw(noSuchElement, "")
		}
		c.key = tokens[0]
		if len(tokens) > 1 {
			c.stereo = strings.TrimFunc(tokens[1], func(r rune) bool { return r <= ' ' })
			c.hasStereo = true
		}
	} else {
		c.key = key
	}
	return c
}

// ConvertNow converts the value; counter numbers the new values.
func (c *FromSkinparamToStyle) ConvertNow(value string, counter AutomaticCounter) {
	if strings.HasSuffix(c.key, "shadowing") {
		if javaEqualsIgnoreCase(value, "false") {
			value = "0"
		} else if javaEqualsIgnoreCase(value, "true") {
			value = "3"
		}
	} else if c.key == "hyperlinkunderline" {
		if javaEqualsIgnoreCase(value, "false") {
			value = "0"
		}
		if javaEqualsIgnoreCase(value, "true") {
			value = "1"
		}
	}
	if javaEqualsIgnoreCase(value, "right:right") {
		value = "right"
	}
	if javaEqualsIgnoreCase(value, "dotted") {
		value = "1;3"
	}
	if javaEqualsIgnoreCase(value, "dashed") {
		value = "7;7"
	}
	datas, ok := skinparamKnowledge[javaLower(c.key)]
	if !ok {
		if javaEqualsIgnoreCase(c.key, "shadowing") {
			c.addStyle(PNameShadowing, shadowingValue(value, counter), SNameRoot)
		} else if javaEqualsIgnoreCase(c.key, "noteshadowing") {
			c.addStyle(PNameShadowing, shadowingValue(value, counter), SNameRoot, SNameNote)
		}
		return
	}
	if isComplexSkinparamValue(value) {
		if strings.Contains(value, ";") {
			if strings.HasPrefix(value, ";") {
				value = " " + value
			}
			tokens := javaTokenize(value, ";")
			value = tokens[0]
			for _, read := range tokens[1:] {
				c.readValue(read, datas, counter)
			}
		} else {
			c.readValue(value, datas, counter)
			return
		}
	}
	if value != " " {
		for _, data := range datas {
			c.addStyle(data.propertyName, ValueImplRegular(value, counter), data.styleNames...)
		}
	}
}

func (c *FromSkinparamToStyle) readValue(read string, datas []skinparamData, counter AutomaticCounter) {
	switch {
	case strings.HasPrefix(read, "text:"):
		parts := javaSplit(read, ":")
		if len(parts) < 2 {
			throw("java.lang.ArrayIndexOutOfBoundsException", "Index 1 out of bounds for length "+itoa(len(parts)))
		}
		value2 := parts[1]
		for _, data := range datas {
			c.addStyle(PNameFontColor, ValueImplRegular(value2, counter), data.styleNames...)
		}
	case strings.HasPrefix(read, "line.dotted"):
		for _, data := range datas {
			c.addStyle(PNameLineStyle, ValueImplRegular("1;3", counter), data.styleNames...)
		}
	case strings.HasPrefix(read, "line.dashed"):
		for _, data := range datas {
			c.addStyle(PNameLineStyle, ValueImplRegular("7;7", counter), data.styleNames...)
		}
	case strings.Contains(javaLower(read), "bold"):
		for _, data := range datas {
			c.addStyle(PNameLineThickness, ValueImplRegular("2", counter), data.styleNames...)
		}
	}
}

func isComplexSkinparamValue(value string) bool {
	return strings.Contains(value, ";") || strings.HasPrefix(value, "text:")
}

func shadowingValue(value string, counter AutomaticCounter) *ValueImpl {
	if javaEqualsIgnoreCase(value, "false") || javaEqualsIgnoreCase(value, "no") {
		return ValueImplRegular("0", counter)
	}
	if javaEqualsIgnoreCase(value, "true") || javaEqualsIgnoreCase(value, "yes") {
		return ValueImplRegular("3", counter)
	}
	return ValueImplRegular(value, counter)
}

func (c *FromSkinparamToStyle) addStyle(propertyName PName, value Value, styleNames ...SName) {
	var values styleValues
	values[propertyName] = value
	sig := StyleSignatureBasicOf(styleNames...)
	if c.hasStereo {
		names := javaSplit(c.stereo, "&")
		values = addStereotypeCount(values, len(names))
		for _, s := range names {
			sig = sig.AddStereotype(s)
		}
	}
	c.styles = append(c.styles, newStyleValues(sig, values))
}

// Styles returns the converted styles.
func (c *FromSkinparamToStyle) Styles() []*Style { return c.styles }
