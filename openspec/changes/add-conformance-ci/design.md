# Design

## Context

See proposal.md for the motivation. The current state that shapes the approach:

- `.github/workflows/ci.yml` has one job, `test`, on `ubuntu-latest` (linux/amd64): lint, unit tests, provenance and the WASM build. It caches `.cache/upstream`.
- `make conformance` runs three steps:
  1. `tools/reference/fetch-tools.sh` downloads resvg for the host and copies the DejaVu fonts out of the reference image.
  2. `go run ./tools/refgen -impl go -png -out .cache/conformance/actual` renders every case with the Go implementation, natively on the host. Class diagrams use whatever `dot` the host has.
  3. `go test -tags conformance ./internal/conformance` compares the outputs with the references and writes `conformance-report.md`.
- References come from Java PlantUML in the reference image (`tools/reference/Dockerfile`, all inputs pinned by digest or checksum, except the two apt packages `fontconfig` and `fonts-dejavu`). `go run ./tools/refgen -image plantuml-go-reference -png` renders them:
  - curated references go to `testdata/conformance/reference/` and are committed
  - Vega references go to `.cache/reference/vega/` and PNG references to `.cache/reference-png/`, and are not committed
  - CLI cases start one container per case, and HTTP cases also use the pinned `plantuml/plantuml-server` image
- The reference image has `dot` 14.0.1, `resvg` and `curl`, but no `git`, `make` or C compiler.
- A local run on 2026-10-17 (macOS arm64, warm caches, host `dot` 15.0.0) with `REQUIRE=all` passed: 455 cases, 0 failures, 11 known differences, 9.7 seconds in total.

## Goals / Non-Goals

**Goals:**

- One CI job that runs the conformance suite as the specs describe, with the same entry point (`make conformance`) as a local run.
- Runs with warm caches spend their time on the Go render and the comparison, not on building the image or rendering references with Java.
- No change to the reference image, the corpus, the comparison thresholds or the known differences.

**Non-Goals:**

- Making the check a required status check in branch protection. That is a repository setting, outside this change.
- arm64 or macOS runners.
- Publishing the reference image to a container registry.
- Running `make reference-check` (two renders of the curated cases) in CI.
- Pinning the apt package versions of the reference image.

## Decisions

### D1. A second job in the existing workflow

The `conformance` job is added to `.github/workflows/ci.yml` next to `test`, with the same triggers (`pull_request`, `push` to `main`) and the same `contents: read` permission. The two jobs run in parallel.

Alternative: a separate workflow file. It would repeat the triggers and the Go setup without any benefit, because the triggers are the same.

### D2. Build the reference image in CI with a layer cache

The job builds `tools/reference/Dockerfile` for `linux/amd64` with Docker Buildx, using the GitHub Actions cache as the layer cache, and loads the result into the local Docker daemon as `plantuml-go-reference`, the name refgen and `fetch-tools.sh` expect. With a warm cache every layer comes from the cache, so the build does not download anything.

Alternatives:

- Push the image to GHCR and pull it. That needs `packages: write`, image retention and a naming scheme, for the same result: every input is already pinned, so a rebuild from the layer cache gives the same image.
- Build without a cache. Every run would download the base images, the jar and resvg again.

### D3. Cache the Vega corpus and the uncommitted references

- `.cache/vega` is cached with a key on `tools/corpus/fetch-vega.sh`. The job runs the script on every run, because the Go render and the comparison read the corpus too. The script is a no-op when the corpus is complete.
- `.cache/reference/vega` and `.cache/reference-png` are cached together. The key combines the ID of the built image with a hash of `tools/reference/**`, `tools/refgen/**`, `internal/conformance/**`, `tools/corpus/fetch-vega.sh`, `testdata/conformance/cases/**` and `testdata/conformance/reference/**`. Every input that can change a reference is in the key, so a cache hit is safe to use.
- On a miss, the job runs `go run ./tools/refgen -image plantuml-go-reference -png`, which renders all suites with Java. It does not use `make reference`, because that target runs `docker build` again outside Buildx and its cache.

Alternatives:

- Commit the Vega references. The foundation design rejected that, because they embed upstream's Vega sources.
- Render the references on every run. Simpler, but every run would start a JVM container per CLI case.

### D4. Check the committed references when they are rendered

After a Java render (a cache miss in D3), the job fails if `git status --porcelain -- testdata/conformance/reference` is not empty, and prints the differing files. The committed curated references are then out of date for the pinned image, or the image itself has drifted (see the apt risk below). Either way, the comparison would otherwise run against references that the pinned image no longer produces.

The check costs nothing extra: the render already rewrote those files. Because `testdata/conformance/reference/**` is in the cache key, a pull request that edits a committed reference always gets this check.

Alternative: `make reference-check`. It compares two renders with each other, not with the committed files, and it needs a second render.

### D5. Render the Go outputs in the reference image

`make conformance` gets an option, `GO_IN_IMAGE=1`, that changes only step 2 of the target: `go run ./tools/refgen -impl go ...` runs in a container from the reference image instead of on the host. The container:

- runs for `linux/amd64` as the host user, so the files it writes belong to the runner
- mounts the repository, the host's Go installation (`go env GOROOT`), module cache and build cache, so it needs neither a Go install in the image nor network access
- sets `CGO_ENABLED=0`, because the image has no C compiler, and `GOFLAGS=-buildvcs=false`, because it has no `git`. Nothing in the comparison depends on the VCS stamp of the version string.

Steps 1 and 3 stay on the host, as today. Without `GO_IN_IMAGE` the target works exactly as before, so local runs on arm64 hosts do not change. CI runs `make conformance REQUIRE=all GO_IN_IMAGE=1`, and a linux/amd64 developer machine can run the same command.

Alternatives:

- Install Graphviz on the runner from apt. Ubuntu's version is not 14.0.1, so class layouts would drift from the references, which defeats the pinned reference.
- Copy `dot` out of the image to the host. Its plugin directory is compiled in as `/usr/local/lib/graphviz`, and it links against the image's Ubuntu 22.04 libraries, so it is fragile on the runner's newer Ubuntu.
- Point `GRAPHVIZ_DOT` at a wrapper that runs `dot` through `docker exec` in a long-running container. It works, but it adds a container lifecycle to manage and a process round trip per layout, while only `dot` would run in the image.
- Add Go to the reference image. That ties the Go toolchain version to the pinned reference image, and a toolchain bump would invalidate the image and reference caches.

### D6. Every group is required

CI passes `REQUIRE=all`. A case that reports "not supported yet" for an unsupported type has the status "unsupported", not "fail", so `all` only fails on real mismatches. This keeps the rule that an unsupported type never produces a wrong drawing. `testdata/conformance/supported.yaml` stays absent, and the local default stays report-only.

Alternative: a `supported.yaml` that lists the supported groups. Wrong output in an unsupported group would then only show up in the report.

### D7. Publish the report

The job uploads `conformance-report.md` as a run artifact and appends it to the run summary (`$GITHUB_STEP_SUMMARY`), in steps that run whether the comparison passed or failed. The failing cases are also in the test log, one line per case with its ID, from the existing `t.Errorf`.

## Risks / Trade-offs

- **[Unknown duration of a cold run]** The Java render with CLI and HTTP cases has never run on linux/amd64 without emulation. → Measure it in the first run. The job gets a timeout above the 30-minute `go test` timeout. Cold runs happen only when a key in D3 changes.
- **[Cache eviction]** GitHub evicts caches unused for 7 days and caps the repository at 10 GB. → An evicted cache only makes the next run a cold run. The image layers and references take a few hundred MB.
- **[Unstable image ID]** If a build from a warm layer cache gives a different image ID each time, the reference cache never hits. → Runs stay correct but slow. Check the ID in two consecutive runs. If it is not stable, key on the Dockerfile hash instead.
- **[apt drift]** `fontconfig` and `fonts-dejavu` are not pinned to a version. After a layer cache miss, a newer package could change the rendering. → The image ID changes, the references are rendered again, and D4 fails loudly instead of comparing against stale references. Pinning the packages is a separate change.
- **[Docker Hub limits]** A cold run pulls the base images and the plantuml-server image. → The layer cache avoids most pulls. GitHub-hosted runners have higher pull limits than anonymous clients.
- **[CI minutes]** The repository is private, and the job runs on every pull request. → Warm runs are short. The triggers can be narrowed later without changing the job.
- **[Stricter merges]** A follow-up change that adds a diagram type must make its cases pass or list them as known differences before CI passes. → That is the support rule of the upstream-compatibility spec, now enforced.

## Migration Plan

1. Merge the change. The first run on the pull request is a cold run that fills the caches.
2. Read the duration of the cold and warm runs from the job log and adjust the timeout if needed.
3. Rollback: remove the `conformance` job. `GO_IN_IMAGE` is opt-in and does not affect other targets.

## Open Questions

- How long does a cold run take on GitHub's runners? The answer only tunes the job timeout.
