// SPDX-License-Identifier: MIT
// Provenance: original

//go:build js && wasm

// Command plantuml-wasm is the WebAssembly entry point of the renderer. It
// defines the global plantuml.render(source, options), which returns a
// Promise of the result; plantuml.mjs loads it in browsers and Node.js.
package main

import (
	"context"
	"errors"
	"syscall/js"
	"testing/fstest"

	"github.com/ekalinin/gopuml"
)

func main() {
	render := js.FuncOf(render)
	js.Global().Set("plantuml", js.ValueOf(map[string]any{"render": render}))
	select {}
}

var formats = map[string]gopuml.Format{"svg": gopuml.SVG, "png": gopuml.PNG, "txt": gopuml.TXT, "utxt": gopuml.UTXT}

// options reads the render options of the JavaScript API: format, diagram,
// page, darkMode, metadata (true by default) and files, an object mapping
// include file names to their text, or to their bytes as a Uint8Array (a zip
// archive for !import). The profile is SANDBOX: includes are read only from
// those files.
func options(o js.Value) (*gopuml.Options, error) {
	opts := &gopuml.Options{SecurityProfile: "SANDBOX", FS: fstest.MapFS{}, Dir: "/"}
	if o.Type() != js.TypeObject {
		return opts, nil
	}
	if f := o.Get("format"); f.Type() == js.TypeString {
		format, ok := formats[f.String()]
		if !ok {
			return nil, errors.New("unknown format " + f.String())
		}
		opts.Format = format
	}
	if v := o.Get("diagram"); v.Type() == js.TypeNumber {
		opts.Diagram = v.Int()
	}
	if v := o.Get("page"); v.Type() == js.TypeNumber {
		opts.Page = v.Int()
	}
	opts.DarkMode = o.Get("darkMode").Truthy()
	if v := o.Get("metadata"); v.Type() == js.TypeBoolean {
		opts.NoMetadata = !v.Bool()
	}
	if files := o.Get("files"); files.Type() == js.TypeObject {
		fsys := fstest.MapFS{}
		keys := js.Global().Get("Object").Call("keys", files)
		uint8Array := js.Global().Get("Uint8Array")
		for i := range keys.Length() {
			name := keys.Index(i).String()
			v := files.Get(name)
			var data []byte
			if v.InstanceOf(uint8Array) {
				data = make([]byte, v.Length())
				js.CopyBytesToGo(data, v)
			} else {
				data = []byte(v.String())
			}
			fsys[name] = &fstest.MapFile{Data: data}
		}
		opts.FS = fsys
	}
	return opts, nil
}

// errorKinds name the typed errors of the library for JavaScript.
var errorKinds = []struct {
	err  error
	kind string
}{
	{gopuml.ErrSyntax, "syntax"},
	{gopuml.ErrNoDiagram, "no-diagram"},
	{gopuml.ErrUnsupportedDiagram, "unsupported-diagram"},
	{gopuml.ErrUnsupportedFormat, "unsupported-format"},
	{gopuml.ErrGraphvizNotFound, "graph-layout-unavailable"},
	{gopuml.ErrGraphvizTimeout, "graphviz-timeout"},
	{gopuml.ErrGraphvizFailed, "graphviz-failed"},
	{gopuml.ErrAccessDenied, "access-denied"},
	{gopuml.ErrDecode, "decode"},
}

// result converts a render result to a JavaScript object: data (a
// Uint8Array, or null), width, height, description, pages, diagrams and
// error (null, or an object with kind, message and, for syntax errors,
// line).
func result(res *gopuml.Result, err error) js.Value {
	out := map[string]any{"data": nil, "width": 0, "height": 0, "description": "", "pages": 0, "diagrams": 0, "error": nil}
	if res != nil {
		data := js.Global().Get("Uint8Array").New(len(res.Data))
		js.CopyBytesToJS(data, res.Data)
		out["data"], out["width"], out["height"] = data, res.Width, res.Height
		out["description"], out["pages"], out["diagrams"] = res.Description, res.Pages, res.Diagrams
	}
	if err != nil {
		e := map[string]any{"kind": "other", "message": err.Error()}
		for _, k := range errorKinds {
			if errors.Is(err, k.err) {
				e["kind"] = k.kind
				break
			}
		}
		if se, ok := errors.AsType[*gopuml.SyntaxError](err); ok {
			e["line"] = se.Line
		}
		if errors.Is(err, gopuml.ErrGraphvizNotFound) {
			e["message"] = "graph layout is not available in this build"
		}
		out["error"] = e
	}
	return js.ValueOf(out)
}

func render(_ js.Value, args []js.Value) any {
	source, opts := "", js.Undefined()
	if len(args) > 0 {
		source = args[0].String()
	}
	if len(args) > 1 {
		opts = args[1]
	}
	var executor js.Func
	executor = js.FuncOf(func(_ js.Value, pa []js.Value) any {
		defer executor.Release()
		resolve := pa[0]
		o, err := options(opts)
		if err != nil {
			resolve.Invoke(result(nil, err))
			return nil
		}
		go func() {
			res, err := gopuml.Render(context.Background(), source, o)
			resolve.Invoke(result(res, err))
		}()
		return nil
	})
	return js.Global().Get("Promise").New(executor)
}
