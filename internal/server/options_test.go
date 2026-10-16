// SPDX-License-Identifier: MIT
// Provenance: original

package server

import (
	"reflect"
	"testing"

	plantuml "github.com/ekalinin/plantuml-in-go"
)

func TestParseOptions(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want plantuml.Options
	}{
		{"none", nil, plantuml.Options{Format: plantuml.PNG}},
		{"charset padding", []string{"--charset", "UTF-8"}, plantuml.Options{Format: plantuml.PNG}},
		{"svg", []string{"--svg"}, plantuml.Options{Format: plantuml.SVG}},
		{"deprecated alias", []string{"-tsvg"}, plantuml.Options{Format: plantuml.SVG}},
		{"case", []string{"-TSVG"}, plantuml.Options{Format: plantuml.SVG}},
		{"txt", []string{"-txt"}, plantuml.Options{Format: plantuml.TXT}},
		{"utxt", []string{"--utxt"}, plantuml.Options{Format: plantuml.UTXT}},
		{"declaration order", []string{"--utxt", "--svg"}, plantuml.Options{Format: plantuml.SVG}},
		{"given twice", []string{"--svg", "--svg", "--txt"}, plantuml.Options{Format: plantuml.TXT}},
		{"all twice", []string{"--svg", "-svg"}, plantuml.Options{Format: plantuml.PNG}},
		{"format flag", []string{"--svg", "-f", "txt"}, plantuml.Options{Format: plantuml.TXT}},
		{"unknown format name", []string{"--svg", "--format", "nope"}, plantuml.Options{Format: plantuml.SVG}},
		{"dark mode", []string{"-darkmode"}, plantuml.Options{Format: plantuml.PNG, DarkMode: true}},
		{"no metadata", []string{"--disable-metadata", "--svg"}, plantuml.Options{Format: plantuml.SVG, NoMetadata: true}},
		{"defines", []string{"-DA=1", "-D", "B=2", "--define", "A=3", "-DC"}, plantuml.Options{
			Format:  plantuml.PNG,
			Defines: []plantuml.Define{{Name: "A", Value: "3"}, {Name: "B", Value: "2"}, {Name: "C"}},
		}},
		{"config order", []string{"-Sshadowing=false", "-Pteoz=true", "--theme", "cerulean", "--skinparam", "dpi=200"}, plantuml.Options{
			Format: plantuml.PNG,
			Config: []string{"!theme cerulean", "!pragma teoz true", "skinparamlocked shadowing false", "skinparamlocked dpi 200"},
		}},
		{"filename", []string{"-filename", "a.puml"}, plantuml.Options{Format: plantuml.PNG, Filename: "a.puml"}},
		{"local files ignored", []string{"--config", "/etc/passwd", "-I/etc/passwd", "-filedir", "/etc"}, plantuml.Options{Format: plantuml.PNG}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseOptions(tt.args)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseOptionsErrors(t *testing.T) {
	for _, args := range [][]string{
		{"--pdf"},
		{"-teps"},
		{"--svg", "-f", "pdf"},
		{"-D"},
		{"--svg", "-S"},
	} {
		if _, err := parseOptions(args); err == nil {
			t.Errorf("%q: no error", args)
		}
	}
	// A second, unsupported format flag does not matter when given twice.
	if got, err := parseOptions([]string{"--pdf", "--pdf", "--svg"}); err != nil || got.Format != plantuml.SVG {
		t.Errorf("got %v, %v", got.Format, err)
	}
}
