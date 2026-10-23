GO ?= go
WASM_OUT ?= dist/wasm

.PHONY: build test lint provenance wasm site bench reference reference-check conformance

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

# Assembles the landing page (site/) with the local WebAssembly build, for a
# preview: python3 -m http.server -d dist/site
SITE_OUT ?= dist/site

site: wasm
	$(GO) run ./tools/site -out $(SITE_OUT) -wasm $(WASM_OUT)

# Compares this implementation with the upstream jar (tools/bench), e.g.
# make bench BENCH_FLAGS="-start-runs 3 -render-runs 1".
BENCH_FLAGS ?=

bench:
	mkdir -p .cache/bench
	$(GO) build -trimpath -o .cache/bench/plantuml ./cmd/plantuml
	$(GO) run ./tools/bench -cli .cache/bench/plantuml $(BENCH_FLAGS)

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

# GO_IN_IMAGE=1 renders the IMPL=go outputs in the reference image, so that
# they use its Graphviz, as the references do. It needs a linux/amd64 host:
# Go binaries crash under amd64 emulation. The image has no C compiler and
# no git, hence CGO_ENABLED=0 and -buildvcs=false.
ifeq ($(GO_IN_IMAGE),1)
GO_IN_IMAGE_DIRS := $(CURDIR) $(shell $(GO) env GOROOT GOMODCACHE GOCACHE)
REFGEN_GO = docker run --rm --platform linux/amd64 --user "$$(id -u):$$(id -g)" \
	$(foreach d,$(GO_IN_IMAGE_DIRS),-v "$(d):$(d)") -w "$(CURDIR)" \
	-e HOME=/tmp -e CGO_ENABLED=0 -e GOFLAGS=-buildvcs=false \
	-e GOMODCACHE="$(word 3,$(GO_IN_IMAGE_DIRS))" -e GOCACHE="$(word 4,$(GO_IN_IMAGE_DIRS))" \
	-e PATH="$(word 2,$(GO_IN_IMAGE_DIRS))/bin:/usr/local/bin:/usr/bin:/bin" \
	$(REFERENCE_IMAGE) go
else
REFGEN_GO = $(GO)
endif

conformance:
	tools/reference/fetch-tools.sh
	$(REFGEN_GO) run ./tools/refgen -image $(REFERENCE_IMAGE) -png -impl $(IMPL) -out $(CONFORMANCE_DIR)/actual
	$(GO) test -tags conformance -count=1 -timeout 30m -run TestConformance -v ./internal/conformance -args \
		-root $(CURDIR) -actual $(CURDIR)/$(CONFORMANCE_DIR)/actual -impl $(IMPL) \
		-resvg $(CURDIR)/.cache/tools/resvg -fonts $(CURDIR)/.cache/fonts/dejavu \
		-report $(CURDIR)/conformance-report.md $(if $(REQUIRE),-require $(REQUIRE))
