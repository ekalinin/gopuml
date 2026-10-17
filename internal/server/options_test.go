// SPDX-License-Identifier: MIT
// Provenance: original

package server

import (
	"reflect"
	"testing"

	"github.com/ekalinin/gopuml"
)

func TestParseOptions(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want gopuml.Options
	}{
		{"none", nil, gopuml.Options{Format: gopuml.PNG}},
		{"charset padding", []string{"--charset", "UTF-8"}, gopuml.Options{Format: gopuml.PNG}},
		{"svg", []string{"--svg"}, gopuml.Options{Format: gopuml.SVG}},
		{"deprecated alias", []string{"-tsvg"}, gopuml.Options{Format: gopuml.SVG}},
		{"case", []string{"-TSVG"}, gopuml.Options{Format: gopuml.SVG}},
		{"txt", []string{"-txt"}, gopuml.Options{Format: gopuml.TXT}},
		{"utxt", []string{"--utxt"}, gopuml.Options{Format: gopuml.UTXT}},
		{"declaration order", []string{"--utxt", "--svg"}, gopuml.Options{Format: gopuml.SVG}},
		{"given twice", []string{"--svg", "--svg", "--txt"}, gopuml.Options{Format: gopuml.TXT}},
		{"all twice", []string{"--svg", "-svg"}, gopuml.Options{Format: gopuml.PNG}},
		{"format flag", []string{"--svg", "-f", "txt"}, gopuml.Options{Format: gopuml.TXT}},
		{"unknown format name", []string{"--svg", "--format", "nope"}, gopuml.Options{Format: gopuml.SVG}},
		{"dark mode", []string{"-darkmode"}, gopuml.Options{Format: gopuml.PNG, DarkMode: true}},
		{"no metadata", []string{"--disable-metadata", "--svg"}, gopuml.Options{Format: gopuml.SVG, NoMetadata: true}},
		{"defines", []string{"-DA=1", "-D", "B=2", "--define", "A=3", "-DC"}, gopuml.Options{
			Format:  gopuml.PNG,
			Defines: []gopuml.Define{{Name: "A", Value: "3"}, {Name: "B", Value: "2"}, {Name: "C"}},
		}},
		{"config order", []string{"-Sshadowing=false", "-Pteoz=true", "--theme", "cerulean", "--skinparam", "dpi=200"}, gopuml.Options{
			Format: gopuml.PNG,
			Config: []string{"!theme cerulean", "!pragma teoz true", "skinparamlocked shadowing false", "skinparamlocked dpi 200"},
		}},
		{"filename", []string{"-filename", "a.puml"}, gopuml.Options{Format: gopuml.PNG, Filename: "a.puml"}},
		{"local files ignored", []string{"--config", "/etc/passwd", "-I/etc/passwd", "-filedir", "/etc"}, gopuml.Options{Format: gopuml.PNG}},
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
	if got, err := parseOptions([]string{"--pdf", "--pdf", "--svg"}); err != nil || got.Format != gopuml.SVG {
		t.Errorf("got %v, %v", got.Format, err)
	}
}
