// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cucadiagram/Member.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package abel

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// Member is a field or a method of a class-like entity: its text without
// the modifiers, its visibility, the {static}, {classifier} and {abstract}
// modifiers and an optional link. It is a line of a Display.
type Member struct {
	display            string
	raw                string
	staticModifier     bool
	abstractModifier   bool
	url                *klimt.URL
	visibilityModifier klimt.VisibilityModifier
	hasVisibility      bool
}

// String is upstream's toString: the raw text.
func (m *Member) String() string { return m.raw }

// MemberMethod returns the method of a body line, with its modifiers.
func MemberMethod(tmpDisplay string) *Member { return newMember(true, tmpDisplay, true) }

// MemberField returns the field of a body line, with its modifiers.
func MemberField(tmpDisplay string) *Member { return newMember(true, tmpDisplay, false) }

// MemberMethodManage returns a method; manageModifier false keeps the
// modifiers in the text.
func MemberMethodManage(tmpDisplay string, manageModifier bool) *Member {
	return newMember(manageModifier, tmpDisplay, true)
}

// MemberFieldManage returns a field; manageModifier false keeps the
// modifiers in the text.
func MemberFieldManage(tmpDisplay string, manageModifier bool) *Member {
	return newMember(manageModifier, tmpDisplay, false)
}

var (
	// memberURL is upstream's Member.URL, a Pattern2 (case-insensitive,
	// with the macros).
	memberURL = uml.Compile(`^(.*?)(?:\[(` + klimt.URLBuilderRegexp() + `)\])?$`)
	// The plain java.util.regex patterns of Member: CASE_INSENSITIVE folds
	// ASCII letters, \s is Java's ASCII white space.
	memberRemoveTag                      = regexp.MustCompile(`(?i)\{(method|field)\}[\t\n\x0B\f\r ]*`)
	memberRemoveStaticClassifierAbstract = regexp.MustCompile(`(?i)\{(static|classifier|abstract)\}[\t\n\x0B\f\r ]*`)
)

func newMember(manageModifier bool, tmpDisplay string, isMethod bool) *Member {
	m := &Member{raw: tmpDisplay}
	tmpDisplay = memberRemoveTag.ReplaceAllString(tmpDisplay, "")
	if manageModifier {
		idx := memberURL.FindStringSubmatchIndex(tmpDisplay)
		if idx == nil {
			uml.Throw(uml.IllegalStateException, "")
		}
		urlStart, urlEnd := idx[4], idx[5]
		if urlStart >= 0 {
			m.url = klimt.NewURLBuilder("", false, klimt.URLModeStrict).URL(tmpDisplay[urlStart:urlEnd])
		}
		tmpDisplay = tmpDisplay[idx[2]:idx[3]]
	}
	lower := uml.LowerCase(tmpDisplay)
	if !manageModifier {
		tmpDisplay = text.Trin(tmpDisplay)
		if tmpDisplay == "" {
			m.display = " "
		} else {
			m.display = klimt.GuillemetGuillemet().ManageGuillemet(text.Trin(tmpDisplay))
		}
		return m
	}
	m.staticModifier = strings.Contains(lower, "{static}") || strings.Contains(lower, "{classifier}")
	m.abstractModifier = strings.Contains(lower, "{abstract}")
	displayClean := text.Trin(memberRemoveStaticClassifierAbstract.ReplaceAllString(tmpDisplay, ""))
	if displayClean == "" {
		displayClean = " "
	}
	if klimt.VisibilityIsVisibilityCharacter(displayClean) {
		m.visibilityModifier, m.hasVisibility = klimt.VisibilityOf(displayClean, !isMethod)
		_, size := utf8.DecodeRuneInString(displayClean)
		m.display = text.Trin(klimt.GuillemetGuillemet().ManageGuillemet(displayClean[size:]))
	} else {
		m.display = klimt.GuillemetGuillemet().ManageGuillemet(displayClean)
	}
	return m
}

// Display returns the text, with the visibility character when asked.
func (m *Member) Display(withVisibilityChar bool) string {
	if !withVisibilityChar {
		return m.display
	}
	switch {
	case m.isVisibility(klimt.VisibilityPrivateField, klimt.VisibilityPrivateMethod):
		return "-" + m.display
	case m.isVisibility(klimt.VisibilityPublicField, klimt.VisibilityPublicMethod):
		return "+" + m.display
	case m.isVisibility(klimt.VisibilityPackagePrivateField, klimt.VisibilityPackagePrivateMethod):
		return "~" + m.display
	case m.isVisibility(klimt.VisibilityProtectedField, klimt.VisibilityProtectedMethod):
		return "#" + m.display
	case m.isVisibility(klimt.VisibilityIEMandatory, klimt.VisibilityIEMandatory):
		return "*" + m.display
	}
	return m.display
}

func (m *Member) isVisibility(a, b klimt.VisibilityModifier) bool {
	return m.hasVisibility && (m.visibilityModifier == a || m.visibilityModifier == b)
}

// Equals is upstream's equals: members are equal when their texts are.
func (m *Member) Equals(other *Member) bool { return m.display == other.display }

// IsStatic reports whether the member is {static} or {classifier}.
func (m *Member) IsStatic() bool { return m.staticModifier }

// IsAbstract reports whether the member is {abstract}.
func (m *Member) IsAbstract() bool { return m.abstractModifier }

// VisibilityModifier returns the visibility; ok is false when there is
// none (null upstream).
func (m *Member) VisibilityModifier() (klimt.VisibilityModifier, bool) {
	return m.visibilityModifier, m.hasVisibility
}

// URL returns the link, nil for none.
func (m *Member) URL() *klimt.URL { return m.url }

// HasURL reports whether the member has a link.
func (m *Member) HasURL() bool { return m.url != nil }
