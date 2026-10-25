// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/url/UrlBuilder.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/url/UbrexUrlBuilder.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"regexp"
	"strings"
	"unicode/utf16"

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

// The patterns of upstream's UbrexUrlBuilder, tried in order.
var ubrexURLPatterns = func() []*ubrexExpression {
	const start, end = "[[ 〇*〴s", "〇*〴s  ]]"
	var result []*ubrexExpression
	for _, s := range []string{
		start + "〃  〶$URL=〇+「〤〃」 〃" + "〇?〘 〇*〴s { 〶$TOOLTIP=〇*「〤{}」 } 〙" +
			"〇?〘 〴s 〶$LABEL=〘「〤〴s{}[]」  〇*「〤[]」〙  〙" + end,
		start + "{  〶$TOOLTIP=〄>〘 }〇*〴s]]〙",
		start + "{  〶$TOOLTIP=〇+「〤{}」 }" + "〇*〴s" + "〶$LABEL=〘「〤〴s[]{}」 〇*「〤[]」〙" + end,
		start + "〶$URL=〇+「〤〴s〴g{}[]」" + "〇*〴s" + "{  〶$TOOLTIP=〄>〘 }〇*〴s]]〙",
		start + "〶$URL=〇+「〤〴s〴g{}[]」" + "〇?〘  〇*〴s  {  〶$TOOLTIP=〇*「〤{}」  } 〙" +
			"〇?〘  〴s 〶$LABEL=〘「〤〴s{}[]」 〇*「〤[]」〙  〙" + end,
		start + "〶$URL=〇+「〤〴s〴g[]」" + "〇?〘  〴s 〶$LABEL=〘「〤〴s{}[]」 〇*「〤[]」〙  〙" + end,
	} {
		result = append(result, ubrexBuild(s))
	}
	return result
}()

// UbrexURLBuilder is upstream's UbrexUrlBuilder: the link builder of the
// commands written with UBrex.
type UbrexURLBuilder struct {
	topurl string
	mode   URLMode
}

// NewUbrexURLBuilder is new UbrexUrlBuilder(topurl, mode).
func NewUbrexURLBuilder(topurl string, mode URLMode) *UbrexURLBuilder {
	return &UbrexURLBuilder{topurl: topurl, mode: mode}
}

// URL is getUrl: the link of s, nil when there is none.
func (b *UbrexURLBuilder) URL(s string) *URL {
	for _, p := range ubrexURLPatterns {
		if m, ok := b.matcher(p, s); ok && b.matchesOrFind(m) {
			value := func(key string) (string, bool) {
				values := m.findValuesByKey(key)
				if len(values) == 0 {
					return "", false
				}
				return values[0], true
			}
			url, urlOK := value("URL")
			tooltip, tooltipOK := value("TOOLTIP")
			label, labelOK := value("LABEL")
			return NewURL(b.withTopURL(url, urlOK), tooltip, tooltipOK, label, labelOK)
		}
	}
	return nil
}

func (b *UbrexURLBuilder) matcher(p *ubrexExpression, s string) (*ubrexMatcher, bool) {
	u := utf16.Encode([]rune(s))
	if b.mode == URLModeStrict {
		return p.match(u, 0), true
	}
	for i := 0; i < len(u)-2; i++ {
		if m := p.match(u, i); m.startMatch() {
			return m, true
		}
	}
	return nil, false
}

func (b *UbrexURLBuilder) matchesOrFind(m *ubrexMatcher) bool {
	switch b.mode {
	case URLModeStrict:
		return m.exactMatch()
	case URLModeAnywhere:
		return m.startMatch()
	}
	throw(illegalState, "")
	return false
}

func (b *UbrexURLBuilder) withTopURL(url string, ok bool) string {
	if !ok {
		return ""
	}
	if !strings.HasPrefix(url, "http:") && !strings.HasPrefix(url, "https:") && !strings.HasPrefix(url, "file:") {
		return b.topurl + url
	}
	return url
}
