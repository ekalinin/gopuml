// SPDX-License-Identifier: MIT
// Provenance: original

//go:build conformance

package code

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/conformance"
)

// TestEncodeDecodedByUpstream encodes the source of every curated diagram
// case and checks that upstream's --decode-url, run in the reference image,
// gives back the same source. Encodings need not be byte-identical.
//
//	go test -tags conformance -run TestEncodeDecodedByUpstream ./internal/code
func TestEncodeDecodedByUpstream(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	if exec.Command("docker", "image", "inspect", "plantuml-go-reference").Run() != nil {
		t.Skip("reference image plantuml-go-reference not built")
	}
	cases, err := conformance.LoadCases(os.DirFS(filepath.Join(root, "testdata/conformance")), "cases")
	if err != nil {
		t.Fatal(err)
	}
	// One container decodes every code; each result ends with a marker line.
	var script strings.Builder
	var codes, want []string
	for _, c := range cases {
		code := Encode(c.Source)
		if code == "" {
			continue
		}
		codes = append(codes, c.ID)
		// --decode-url wraps the decoded source in one more @startuml/@enduml.
		want = append(want, "@startuml\n"+decompressSource(compressSource(c.Source))+"\n@enduml\n")
		script.WriteString("java -cp /opt/plantuml.jar:/opt/tools RunCli --decode-url " + strconv.Quote(code) + "\necho '#END#'\n")
	}
	cmd := exec.Command("docker", "run", "--rm", "-i", "--platform", "linux/amd64", "plantuml-go-reference", "sh")
	cmd.Stdin = strings.NewReader(script.String())
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("docker: %v\n%s", err, stderr.String())
	}
	results := strings.Split(string(out), "#END#\n")
	if len(results) != len(want)+1 {
		t.Fatalf("got %d results for %d codes", len(results)-1, len(want))
	}
	for i := range want {
		if results[i] != want[i] {
			t.Errorf("%s: upstream decoded %q, want %q", codes[i], results[i], want[i])
		}
	}
	t.Logf("%d encodings decoded by upstream", len(want))
}

// TestEncodeCaseVariantDecodedByUpstream checks that a source whose start
// and end lines are not in lower case keeps them through an encoding by this
// system and upstream's --decode-url.
//
//	go test -tags conformance -run TestEncodeCaseVariantDecodedByUpstream ./internal/code
func TestEncodeCaseVariantDecodedByUpstream(t *testing.T) {
	if exec.Command("docker", "image", "inspect", "plantuml-go-reference").Run() != nil {
		t.Skip("reference image plantuml-go-reference not built")
	}
	code := Encode("@startUML\nAlice -> Bob : hi\n@endUML\n")
	cmd := exec.Command("docker", "run", "--rm", "--platform", "linux/amd64", "plantuml-go-reference",
		"java", "-cp", "/opt/plantuml.jar:/opt/tools", "RunCli", "--decode-url", code)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("docker: %v\n%s", err, stderr.String())
	}
	// --decode-url wraps the decoded source in one more @startuml/@enduml.
	if want := "@startuml\n@startUML\nAlice -> Bob : hi\n@endUML\n@enduml\n"; string(out) != want {
		t.Errorf("upstream decoded %q, want %q", out, want)
	}
}
