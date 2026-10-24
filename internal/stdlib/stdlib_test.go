// SPDX-License-Identifier: MIT
// Provenance: original

package stdlib

import (
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRetrieveLinks(t *testing.T) {
	for name, want := range map[string]string{
		"awslib":    "awslib14",
		"bootstrap": "bootstrap1.13.1",
		"material":  "material2.1.19",
		"material2": "material2.1.19",
		"material7": "material7.4.47",
		"c4":        "c4",
	} {
		lib, err := Retrieve(name)
		if err != nil {
			t.Fatalf("Retrieve(%q): %v", name, err)
		}
		if lib.Name() != want {
			t.Errorf("Retrieve(%q) = %s, want %s", name, lib.Name(), want)
		}
	}
}

func TestRetrieveMissing(t *testing.T) {
	for name, path := range map[string]string{
		"foo":   "stdlib/foo/info.spm",
		"C4":    "stdlib/C4/info.spm", // folder names are lower case, as in the jar
		"adaml": "stdlib/adaml/info.spm",
	} {
		_, err := Retrieve(name)
		var nf *FileNotFoundError
		if !errors.As(err, &nf) || nf.Path != path || err.Error() != path+" (No such file or directory)" {
			t.Errorf("Retrieve(%q): %v, want the missing %s", name, err, path)
		}
	}
}

func TestChannels(t *testing.T) {
	c4, err := Retrieve("c4")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c4.Puml("c4_container"); err != nil {
		t.Fatal(err)
	}
	if n := len(c4.puml.v); n != 37 {
		t.Errorf("c4 has %d .puml files, want 37", n)
	}
	if data, _ := c4.Puml("c4_container"); !strings.Contains(string(data), "!include <C4/C4_Context>") {
		t.Errorf("c4_container: %.80q", data)
	}
	if data, _ := c4.Puml("missing"); data != nil {
		t.Errorf("missing file: %q", data)
	}
	sprite, err := c4.Sprite("c4_context:person")
	if err != nil || sprite == nil || sprite.Width != 48 || sprite.Height != 48 || len(sprite.Data) != 48*24 {
		t.Errorf("person sprite %+v, %v", sprite, err)
	}
	_, err = c4.Images()
	var nf *FileNotFoundError
	if !errors.As(err, &nf) || nf.Path != "stdlib/c4/image.spm" {
		t.Errorf("c4 images: %v, want the missing image channel", err)
	}

	aws, err := Retrieve("awslib14")
	if err != nil {
		t.Fatal(err)
	}
	images, err := aws.Images()
	if err != nil || len(images) != 1706 {
		t.Fatalf("awslib14: %d images, %v", len(images), err)
	}
	pixels, err := images[0].ARGB()
	if err != nil || images[0].Width != 64 || images[0].Height != 64 || len(pixels) != 64*64 {
		t.Errorf("image 0: %dx%d, %d pixels, %v", images[0].Width, images[0].Height, len(pixels), err)
	}
	if data, err := aws.JSON("all.json"); err != nil || !strings.HasPrefix(strings.TrimSpace(string(data)), "[") {
		t.Errorf("all.json: %.40q, %v", data, err)
	}

	mdi, err := Retrieve("material7")
	if err != nil {
		t.Fatal(err)
	}
	svg, ok, err := mdi.SvgSprite("accountuser/account:mdiAccount")
	if err != nil || !ok || !strings.HasPrefix(svg, `<svg width="48" height="48">`) {
		t.Errorf("mdiAccount: %.60q, %v, %v", svg, ok, err)
	}
}

func TestFolderNames(t *testing.T) {
	names := FolderNames()
	if len(names) != 25 || names[0] != "archimate" || names[len(names)-1] != "tupadr3" {
		t.Errorf("names %q", names)
	}
	for _, n := range names {
		if IsExcluded(n) {
			t.Errorf("excluded %s listed", n)
		}
	}
}

func TestReadUTF(t *testing.T) {
	for in, want := range map[string]string{
		"\x00\x03abc":                      "abc",
		"\x00\x02\xc3\xa9":                 "é",
		"\x00\x02\xc0\x80":                 "\x00",
		"\x00\x06\xed\xa0\xbd\xed\xb8\x80": "😀", // a surrogate pair, as Java writes U+1F600
	} {
		r := &dataReader{data: []byte(in)}
		if got := r.readUTF(); got != want || r.err != nil {
			t.Errorf("readUTF(%q) = %q, %v, want %q", in, got, r.err, want)
		}
	}
	for _, in := range []string{"\x00\x02\xc3", "\x00\x01\xff", "\x00"} {
		r := &dataReader{data: []byte(in)}
		if r.readUTF(); r.err == nil {
			t.Errorf("readUTF(%q): no error", in)
		}
	}
}

// TestConcurrentFirstUse reads one channel from several goroutines at
// once; run with -race.
func TestConcurrentFirstUse(t *testing.T) {
	lib, err := Retrieve("kubernetes")
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			if _, err := lib.Sprite("x"); err != nil {
				t.Error(err)
			}
			if _, err := Retrieve("k8s"); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
}

// TestDecodeTime logs the time to decode the largest channels.
func TestDecodeTime(t *testing.T) {
	if testing.Short() {
		t.Skip("decodes about 100 MB")
	}
	for _, f := range []struct {
		name string
		c    Channel
	}{{"ibm", ChannelImage}, {"tupadr3", ChannelSprite}} {
		start := time.Now()
		data, err := readChannel(f.name, f.c)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%s/%s: %d bytes in %v", f.name, f.c.fileName(), len(data), time.Since(start))
	}
}
