#!/usr/bin/env bash
# Fetches the .puml inputs of upstream's Vega test suite at tag v1.2026.9, and
# the .zip archives they import, into .cache/vega/, keeping their paths
# relative to src/test/resources/vega/.
# The files are used as conformance inputs only and are never committed.
#
# Re-running the script is a no-op once the corpus of TAG is complete.
set -euo pipefail

TAG="v1.2026.9"
REPO="https://github.com/plantuml/plantuml.git"
VEGA_PATH="src/test/resources/vega"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
DEST="${VEGA_CACHE:-${ROOT}/.cache/vega}"

# .complete holds the number of files and the tag of the corpus.
if [[ -f "${DEST}/.complete" && "$(cut -d' ' -f2 "${DEST}/.complete")" == "${TAG}" ]]; then
	echo "vega corpus already present: $(cut -d' ' -f1 "${DEST}/.complete") files in ${DEST}"
	exit 0
fi

work="$(mktemp -d)"
trap 'rm -rf "${work}"' EXIT

git -c advice.detachedHead=false clone --quiet --depth 1 --branch "${TAG}" --filter=blob:none --sparse "${REPO}" "${work}/plantuml"
git -C "${work}/plantuml" sparse-checkout set "${VEGA_PATH}"

rm -rf "${DEST}"
mkdir -p "${DEST}"
(
	cd "${work}/plantuml/${VEGA_PATH}"
	find . -name '*.puml' -o -name '*.zip' | while read -r f; do
		mkdir -p "${DEST}/$(dirname "$f")"
		cp "$f" "${DEST}/$f"
	done
)

count="$(find "${DEST}" -name '*.puml' | wc -l | tr -d ' ')"
if [[ "${count}" -lt 425 ]]; then
	echo "fetch-vega.sh: expected at least 425 .puml files, got ${count}" >&2
	exit 1
fi
echo "${count} ${TAG}" > "${DEST}/.complete"
echo "vega corpus ready: ${count} files in ${DEST}"
