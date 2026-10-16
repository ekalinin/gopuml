// SPDX-License-Identifier: MIT
// Provenance: original

package klimt

import (
	"strconv"
	"strings"
	"testing"
)

// The golden file compared here is written by
// testdata/javagen/CreoleGolden.java (UBrex section).

func javaStringList(values []string) string {
	if values == nil {
		return "null"
	}
	return javaList(values)
}

func TestGoldenUbrex(t *testing.T) {
	var out []string
	for _, line := range readLines(t, "ubrex_inputs.txt") {
		parts := strings.Split(line, "\t")
		ubrex := ubrexBuild(parts[0])
		var keys []string
		if last := parts[len(parts)-1]; last != "" {
			keys = strings.Split(last, ",")
		}
		for _, input := range parts[1 : len(parts)-1] {
			m := ubrex.matchString(input)
			s := parts[0] + "\t" + input + "\t" + strconv.FormatBool(m.startMatch()) + " " +
				strconv.FormatBool(m.exactMatch()) + " '" + m.acceptedMatch() + "'"
			for _, key := range keys {
				s += " " + key + "=" + javaList(m.findValuesByKey(key))
			}
			if len(keys) > 0 {
				var extract []string
				for _, e := range m.result.capture.extractByPrefix(keys[0]) {
					extract = append(extract, e.key+"="+e.value)
				}
				s += " prefix=" + javaStringList(m.result.capture.findFirstValuesByKeyPrefix(keys[0]+"/")) +
					" extract=" + javaList(extract)
			}
			out = append(out, s)
		}
	}
	compareGolden(t, "ubrex.golden", out)
}

func TestUbrexSyntaxErrors(t *testing.T) {
	for _, expression := range []string{"┇", "〒x", "〇~a", "〄~", "〘a", "「a", "〶x=a", "〶$=a", "〶$a-b=c",
		"【a", "〴z", "a b〄+〴.〘c〙"} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%q: no panic", expression)
				}
			}()
			ubrexBuild(expression)
		}()
	}
	if f, ok := ubrexBuild("〶$V=〇+〴d").matchString("12").findFirstValueByKey("V"); !ok || f != "12" {
		t.Error(f, ok)
	}
	if _, ok := ubrexBuild("〶$V=〇+〴d").matchString("x").findFirstValueByKey("V"); ok {
		t.Error("no value expected")
	}
	if v, ok := (ubrexCapture{{"a", "1"}, {"a", "2"}}).findFirstValueByKey("a"); !ok || v != "1" {
		t.Error(v)
	}
}
