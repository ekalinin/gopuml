// SPDX-License-Identifier: MIT
// Provenance: original

package descdiagram

import (
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/core"
)

func domainCommands() []cmd { return []cmd{commandCreateDomain()} }

func TestCreateDomain(t *testing.T) {
	for _, c := range []struct {
		line, code, display, stereotype string
		typ                             abel.LeafType
	}{
		{`domain "Controller" as M <<Machine>>`, "M", "Controller", "Machine", abel.LeafTypeDomain},
		{`domain "Heater" as H <<Causal>>`, "H", "Heater", "Causal", abel.LeafTypeDomain},
		{`domain "Plan" as P <<Designed>>`, "P", "Plan", "Designed", abel.LeafTypeDomain},
		{`domain "Log" as L <<Lexical>>`, "L", "Log", "Lexical", abel.LeafTypeDomain},
		{`domain "Operator" as O <<Biddable>>`, "O", "Operator", "Biddable", abel.LeafTypeDomain},
		{`domain "Room" as R`, "R", "Room", "", abel.LeafTypeDomain},
		{`requirement "Keep the temperature" as K`, "K", "Keep the temperature", "", abel.LeafTypeRequirement},
	} {
		result := parseWith(t, domainCommands, c.line)
		d, ok := result.(*DescriptionDiagram)
		if !ok {
			t.Fatalf("%s: result %T", c.line, result)
		}
		e := entity(t, d, c.code)
		if e.LeafType() != c.typ || e.USymbol() != abel.USymbolNull {
			t.Errorf("%s: type %v, symbol %v", c.line, e.LeafType(), e.USymbol())
		}
		if got := e.Display().Get(0).String(); got != c.display {
			t.Errorf("%s: display %q", c.line, got)
		}
		if st := e.Stereotype(); (c.stereotype == "") != (st == nil) ||
			(st != nil && !strings.Contains(st.String(), c.stereotype)) {
			t.Errorf("%s: stereotype %v", c.line, st)
		}
	}
}

func TestCreateDomainGroup(t *testing.T) {
	d, ok := parseWith(t, func() []cmd { return []cmd{commandCreateDomain(), abel.CommandEndPackage[*DescriptionDiagram]()} },
		`domain "Plant" as P {`, "}").(*DescriptionDiagram)
	if !ok {
		t.Fatal("not a description diagram")
	}
	if e := entity(t, d, "P"); !e.IsGroup() || e.GroupType() != abel.GroupTypeDomain {
		t.Errorf("group %v, type %v", e.IsGroup(), e.GroupType())
	}
	if _, ok := parseWith(t, domainCommands, `domain "A" as X`, `requirement "B" as X`).(*core.ErrorDiagram); !ok {
		t.Error("a second element with the same code was accepted")
	}
}
