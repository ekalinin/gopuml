GO ?= go
WASM_OUT ?= dist/wasm

.PHONY: build test lint provenance wasm reference reference-check conformance

build:
	$(GO) build ./...

test:
	$(GO) test ./...

lint:
	$(GO) vet ./...
	@test -z "$$(gofmt -l .)" || (echo "gofmt needed:"; gofmt -l .; exit 1)

provenance:
	$(GO) run ./tools/provenance $(if $(wildcard .cache/upstream/src/.complete),-src .cache/upstream/src)

wasm:
	mkdir -p $(WASM_OUT)
	GOOS=js GOARCH=wasm $(GO) build -trimpath -o $(WASM_OUT)/plantuml.wasm ./cmd/plantuml-wasm
	cp "$$($(GO) env GOROOT)/lib/wasm/wasm_exec.js" cmd/plantuml-wasm/plantuml.mjs $(WASM_OUT)/
	@echo "plantuml.wasm size: $$(wc -c < $(WASM_OUT)/plantuml.wasm | tr -d ' ') bytes"

REFERENCE_IMAGE ?= plantuml-go-reference

reference:
	docker build --platform linux/amd64 -t $(REFERENCE_IMAGE) tools/reference
	tools/corpus/fetch-vega.sh
	$(GO) run ./tools/refgen -image $(REFERENCE_IMAGE) -png

reference-check:
	$(GO) test -tags conformance -count=1 -timeout 30m -run TestReferencesReproducible -v ./internal/conformance -args -root $(CURDIR)

# Implementation under test: go (default) or java (checks the harness itself).
IMPL ?= go
CONFORMANCE_DIR := .cache/conformance

conformance:
	tools/reference/fetch-tools.sh
	$(GO) run ./tools/refgen -image $(REFERENCE_IMAGE) -png -impl $(IMPL) -out $(CONFORMANCE_DIR)/actual
	$(GO) test -tags conformance -count=1 -timeout 30m -run TestConformance -v ./internal/conformance -args \
		-root $(CURDIR) -actual $(CURDIR)/$(CONFORMANCE_DIR)/actual -impl $(IMPL) \
		-resvg $(CURDIR)/.cache/tools/resvg -fonts $(CURDIR)/.cache/fonts/dejavu \
		-report $(CURDIR)/conformance-report.md $(if $(REQUIRE),-require $(REQUIRE))
