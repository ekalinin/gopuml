#!/usr/bin/env bash
# Downloads the MIT variant of PlantUML v1.2026.9 (binary and sources jars)
# from Maven Central into .cache/upstream/, verifies them against the pinned
# SHA-256 checksums and unpacks the sources into .cache/upstream/src/.
#
# Re-running the script is safe: existing jars are re-verified, not re-downloaded.
set -euo pipefail

VERSION="1.2026.9"
BASE_URL="https://repo1.maven.org/maven2/net/sourceforge/plantuml/plantuml-mit/${VERSION}"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
CHECKSUMS="${ROOT}/tools/upstream/checksums.sha256"
CACHE="${UPSTREAM_CACHE:-${ROOT}/.cache/upstream}"

mkdir -p "${CACHE}"
cd "${CACHE}"

for jar in "plantuml-mit-${VERSION}-sources.jar" "plantuml-mit-${VERSION}.jar"; do
	if [[ ! -f "${jar}" ]]; then
		echo "downloading ${jar}"
		curl -fsSL -o "${jar}.part" "${BASE_URL}/${jar}"
		mv "${jar}.part" "${jar}"
	fi
done

if ! shasum -a 256 -c "${CHECKSUMS}"; then
	echo "fetch.sh: checksum mismatch, refusing to use the downloaded jars" >&2
	exit 1
fi

# src/.complete holds the version of the unpacked sources.
if [[ "$(cat src/.complete 2>/dev/null)" != "${VERSION}" ]]; then
	rm -rf src
	mkdir -p src
	(cd src && unzip -q "../plantuml-mit-${VERSION}-sources.jar")
	echo "${VERSION}" > src/.complete
fi

echo "upstream sources ready in ${CACHE}/src"
