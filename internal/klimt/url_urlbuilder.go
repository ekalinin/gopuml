// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/url/UrlBuilder.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/url/UbrexUrlBuilder.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"regexp"
	"strings"

	"github.com/ekalinin/gopuml/internal/command/regex"
)

// URLBuilderKey is the regex group name of a link in command patterns.
const URLBuilderKey = "URL"

const (
	urlStartPart = `\[\[[%s]*`
	urlEndPart   = `[%s]*\]\]`
	urlSQuoted   = urlStartPart +
		`[%g]([^%g]+)[%g]` + // Quoted part
		`(?:[%s]*\{([^{}]*)\})?` + // Optional tooltip
		`(?:[%s]([^%s\{\}\[\]][^\[\]]*))?` + // Optional label
		urlEndPart
	urlSOnlyTooltip = urlStartPart +
		`\{(` + javaDot + `*)\}` + // Tooltip
		urlEndPart
	urlSOnlyTooltipAndLabel = urlStartPart +
		`\{([^{}]*)\}` + // Tooltip
		`[%s]*` +
		`([^\[%s\{\}\[\]][^\[\]]*)` + // Label
		urlEndPart
	urlSLinkTooltipNoLabel = urlStartPart +
		`([^\s%g{}\[\]]+?)` + // Link
		`[%s]*` +
		`\{(` + javaDot + `+)\}` + // Tooltip
		urlEndPart
	urlSLinkWithOptionalTooltipWithOptionalLabel = urlStartPart +
		`([^%s%g\[\]]+?)` + // Link
		`(?:[%s]*\{([^{}]*)\})?` + // Optional tooltip
		`(?:[%s]([^%s\{\}\[\]][^\[\]]*))?` + // Optional label
		urlEndPart
)

// URLBuilderRegexp returns the pattern of a link, for command patterns.
func URLBuilderRegexp() string {
	return urlSQuoted + "|" + urlSOnlyTooltip + "|" + urlSOnlyTooltipAndLabel + "|" + urlSLinkTooltipNoLabel + "|" +
		urlSLinkWithOptionalTooltipWithOptionalLabel
}

// URLBuilderMandatory returns the leaf capturing a link as "URL".
func URLBuilderMandatory() *regex.Leaf {
	return regex.NewNamedLeaf(12, URLBuilderKey, "("+URLBuilderRegexp()+")")
}

// URLBuilderOptional returns the optional link leaf.
func URLBuilderOptional() regex.Part { return regex.Optional(URLBuilderMandatory()) }

// urlPatterns holds the five link patterns compiled twice: anchored for the
// strict mode (Java's matches) and unanchored for the anywhere mode (find).
type urlPatterns struct {
	strict, anywhere *regexp.Regexp
}

func compileURLPattern(p string) urlPatterns {
	return urlPatterns{strict: regex.MustCompile("^(?:" + p + ")$"), anywhere: regex.MustCompile(p)}
}

var (
	urlQuoted                               = compileURLPattern(urlSQuoted)
	urlOnlyTooltip                          = compileURLPattern(urlSOnlyTooltip)
	urlOnlyTooltipAndLabel                  = compileURLPattern(urlSOnlyTooltipAndLabel)
	urlLinkTooltipNoLabel                   = compileURLPattern(urlSLinkTooltipNoLabel)
	urlLinkWithOptionalTooltipOptionalLabel = compileURLPattern(urlSLinkWithOptionalTooltipWithOptionalLabel)
)

// URLBuilder finds links ("[[target{tooltip} label]]") in texts.
type URLBuilder struct {
	topurl    string
	topurlSet bool
	mode      URLMode
}

// NewURLBuilder returns a builder; topurl, when set, prefixes relative
// targets.
func NewURLBuilder(topurl string, topurlSet bool, mode URLMode) *URLBuilder {
	return &URLBuilder{topurl: topurl, topurlSet: topurlSet, mode: mode}
}

// group returns a submatch and whether it took part in the match.
func group(s string, loc []int, i int) (string, bool) {
	if loc[2*i] < 0 {
		return "", false
	}
	return s[loc[2*i]:loc[2*i+1]], true
}

func (b *URLBuilder) match(p urlPatterns, s string) []int {
	switch b.mode {
	case URLModeStrict:
		return p.strict.FindStringSubmatchIndex(s)
	case URLModeAnywhere:
		return p.anywhere.FindStringSubmatchIndex(s)
	}
	throw(illegalState, "")
	return nil
}

// URL returns the first link found in s, nil when there is none.
func (b *URLBuilder) URL(s string) *URL {
	if m := b.match(urlQuoted, s); m != nil {
		g1, _ := group(s, m, 1)
		g2, ok2 := group(s, m, 2)
		g3, ok3 := group(s, m, 3)
		return NewURL(b.withTopURL(g1), g2, ok2, g3, ok3)
	}
	if m := b.match(urlOnlyTooltip, s); m != nil {
		g1, ok1 := group(s, m, 1)
		return NewURL("", g1, ok1, "", false)
	}
	if m := b.match(urlOnlyTooltipAndLabel, s); m != nil {
		g1, ok1 := group(s, m, 1)
		g2, ok2 := group(s, m, 2)
		return NewURL("", g1, ok1, g2, ok2)
	}
	if m := b.match(urlLinkTooltipNoLabel, s); m != nil {
		g1, _ := group(s, m, 1)
		g2, ok2 := group(s, m, 2)
		return NewURL(b.withTopURL(g1), g2, ok2, "", false)
	}
	if m := b.match(urlLinkWithOptionalTooltipOptionalLabel, s); m != nil {
		g1, _ := group(s, m, 1)
		g2, ok2 := group(s, m, 2)
		g3, ok3 := group(s, m, 3)
		return NewURL(b.withTopURL(g1), g2, ok2, g3, ok3)
	}
	return nil
}

func (b *URLBuilder) withTopURL(url string) string {
	if !strings.HasPrefix(url, "http:") && !strings.HasPrefix(url, "https:") && !strings.HasPrefix(url, "file:") &&
		b.topurlSet {
		return b.topurl + url
	}
	return url
}

// UbrexURLBuilder is upstream's link builder on its UBrex engine, which is
// not ported: using it panics.
type UbrexURLBuilder struct{}

// NewUbrexURLBuilder panics: the UBrex engine (com.plantuml.ubrex) is not
// ported.
func NewUbrexURLBuilder(topurl string, mode URLMode) *UbrexURLBuilder {
	notPorted("url/UbrexUrlBuilder", "the com.plantuml.ubrex engine")
	return nil
}
