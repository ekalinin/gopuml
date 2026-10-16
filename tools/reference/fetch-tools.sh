#!/usr/bin/env bash
# Prepares the host-side tools of the conformance comparison:
#   .cache/tools/resvg   resvg 0.48.1 for the host platform (pinned by checksum)
#   .cache/fonts/dejavu  the DejaVu fonts of the reference image
# Comparing on the host avoids running Go binaries under amd64 emulation,
# while rasterization still uses the same resvg version and fonts everywhere.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
IMAGE="${REFERENCE_IMAGE:-plantuml-go-reference}"
VERSION="0.48.1"
TOOLS="${ROOT}/.cache/tools"
FONTS="${ROOT}/.cache/fonts"

case "$(uname -s)-$(uname -m)" in
Linux-x86_64) asset="resvg-linux-x86_64.tar.gz" sum="fa8c26495a187e592c501db15bf9e8a9fdc051d4b2b336b39703d5b59f912b9d" ;;
Darwin-arm64) asset="resvg-macos-aarch64.zip" sum="06440eb5aa14a28cbfc7e40ae39e1ffa71adc051b89fbaa913b4f1d9b905d09f" ;;
Darwin-x86_64) asset="resvg-macos-x86_64.zip" sum="0135923e443863db251a26bd78eabc6efb4b59d67b8cdc5469e3e1da26bc0ce2" ;;
*)
	echo "fetch-tools.sh: no pinned resvg build for $(uname -s)-$(uname -m)" >&2
	exit 1
	;;
esac

if [[ ! -x "${TOOLS}/resvg" ]]; then
	mkdir -p "${TOOLS}"
	tmp="$(mktemp -d)"
	trap 'rm -rf "${tmp}"' EXIT
	curl -fsSL -o "${tmp}/${asset}" "https://github.com/linebender/resvg/releases/download/v${VERSION}/${asset}"
	echo "${sum}  ${tmp}/${asset}" | shasum -a 256 -c - >/dev/null
	case "${asset}" in
	*.zip) unzip -q -o "${tmp}/${asset}" resvg -d "${TOOLS}" ;;
	*) tar -xzf "${tmp}/${asset}" -C "${TOOLS}" resvg ;;
	esac
	chmod +x "${TOOLS}/resvg"
fi

if [[ ! -d "${FONTS}/dejavu" ]]; then
	mkdir -p "${FONTS}"
	docker run --rm --platform linux/amd64 "${IMAGE}" tar -C /usr/share/fonts/truetype -cf - dejavu | tar -C "${FONTS}" -xf -
fi

echo "resvg $("${TOOLS}/resvg" --version), fonts: $(ls "${FONTS}/dejavu" | wc -l | tr -d ' ') files"
