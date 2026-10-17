# Tasks

## 1. Go render in the reference image

- [x] 1.1 Add the `GO_IN_IMAGE` option to the `conformance` target of the `Makefile` (design D5). With `GO_IN_IMAGE=1`, the refgen step (`go run ./tools/refgen -impl go -png -out ...`) runs in a `linux/amd64` container of `$(REFERENCE_IMAGE)`: as the host user, with the repository, `go env GOROOT`, `go env GOMODCACHE` and `go env GOCACHE` mounted, and with `CGO_ENABLED=0` and `GOFLAGS=-buildvcs=false`. The other steps do not change. Verify that `make -n conformance` prints the same commands as before the change, and that `make -n conformance GO_IN_IMAGE=1` prints the `docker run` form of the refgen step only

## 2. CI job

- [x] 2.1 Add the `conformance` job to `.github/workflows/ci.yml` with the triggers and permissions of the workflow (D1): checkout, Go from `go.mod`, Docker Buildx, and the build of `tools/reference/Dockerfile` for `linux/amd64` with the GitHub Actions layer cache, loaded as `plantuml-go-reference` (D2). Record the image ID as a step output. Verify that `go run github.com/rhysd/actionlint/cmd/actionlint@latest` reports no errors for the workflow
- [x] 2.2 Add the caches of D3: `.cache/vega` keyed on `tools/corpus/fetch-vega.sh`, and `.cache/reference/vega` with `.cache/reference-png` keyed on the image ID and the hash of the inputs listed in D3. Run `tools/corpus/fetch-vega.sh` on every run, because the Go render and the comparison read the corpus too, and on a reference cache miss run `go run ./tools/refgen -image plantuml-go-reference -png`. Verify with actionlint, and verify that the key expression names every input path from D3
- [x] 2.3 After the Java render of 2.2, fail the job when `git status --porcelain -- testdata/conformance/reference` is not empty, printing the differing files (D4). Verify locally by running the step's script once on a clean tree (it passes) and once after editing a committed reference file (it fails and names the file), then restore the file
- [x] 2.4 Run `make conformance REQUIRE=all GO_IN_IMAGE=1` (D5, D6), and set a job timeout above the 30-minute `go test` timeout. Verify with actionlint
- [x] 2.5 Upload `conformance-report.md` as a run artifact and append it to `$GITHUB_STEP_SUMMARY`, in steps that run whether the comparison passed or failed (D7). Verify with actionlint that both steps use `if: always()` or an equivalent condition
- [x] 2.6 Add the five `.go.yaml` cases of D8 under `testdata/conformance/cases/upstream-compatibility/`, and `TestCIWorkflow` in `internal/conformance`, which checks the contract of the `conformance` job in `.github/workflows/ci.yml`. Verify that `make test` passes, and that `TestCIWorkflow` fails for copies of the workflow without `REQUIRE=all`, without the committed references in the reference cache key, without the reference check, or without the `!cancelled()` condition of the report steps
- [x] 2.7 Disable keep-alive in the HTTP client of the refgen HTTP cases (design risk "Flaky HTTP render"). Verify that `go vet ./tools/refgen` passes, and that the next Java render in CI passes the committed-reference check of 2.3, so the recorded references do not change

## 3. Verification in CI

- [x] 3.1 Push the branch and open a pull request. Verify that the cold run of the `conformance` job passes, that both caches of D3 are saved, that the report artifact exists and the run summary shows the report, and note the duration of the cold run
- [x] 3.2 Re-run the job. Verify that the image is built from the layer cache, that the image ID equals the one of 3.1, that the reference cache hits and the Java render is skipped, and note the duration of the warm run. If the image ID differs, key the reference cache on the hash of `tools/reference/Dockerfile` instead (design risk "Unstable image ID") and verify again
- [x] 3.3 Push a temporary commit that changes the Go implementation so that a sequence case no longer matches (for example, a changed message label in the SVG). Verify that the job fails, that the log names the case, and that the report artifact and the run summary are still published. Then drop the commit
- [x] 3.4 Push a temporary commit that edits one file under `testdata/conformance/reference/`. Verify that the reference cache misses, that the check of 2.3 fails and names the file, and that the comparison does not run. Then drop the commit
- [x] 3.5 Set the job timeout from the durations of 3.1 and 3.2, keeping it above the 30-minute `go test` timeout. Verify with actionlint

## 4. Documentation and validation

- [x] 4.1 Update the "Conformance suite" section of `README.md`: CI runs the suite on every pull request and every push to `main` with every group required; on a linux/amd64 host, `make reference` followed by `make conformance REQUIRE=all GO_IN_IMAGE=1` reproduces the CI run; on arm64 hosts, `make conformance` runs the Go side natively and its results are indicative. Verify that every command in the section exists in the `Makefile` with `make -n`
- [x] 4.2 Run `openspec validate add-conformance-ci --strict`. Verify that the change is reported as valid
