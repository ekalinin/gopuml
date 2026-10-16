#!/usr/bin/env bash
# Runs FontMetricsDump in the reference image and writes the metric tables to
# internal/klimt/font/metrics/ and the measured samples to its testdata/.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
IMAGE="${REFERENCE_IMAGE:-plantuml-go-reference}"

docker run --rm --platform linux/amd64 \
	--user "$(id -u):$(id -g)" \
	-v "${ROOT}:/work" \
	"${IMAGE}" \
	java -cp /opt/tools FontMetricsDump \
	/work/internal/klimt/font/metrics \
	/work/internal/klimt/font/metrics/testdata/samples.json
