#!/usr/bin/env bash
# 02-convert-sessions.sh
# Convert Pi sessions to minitrace format

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TICKET_DIR="$(dirname "${SCRIPT_DIR}")"
OUTPUT_DIR="${TICKET_DIR}/output"

SOURCE_DIR="${HOME}/.pi/agent/sessions/--home-manuel-code-wesen-2026-04-09--screencast-studio--"

echo "=== Converting Pi sessions to minitrace format ==="
echo "Source: ${SOURCE_DIR}"
echo "Output: ${OUTPUT_DIR}"
echo ""

# Create output directory
mkdir -p "${OUTPUT_DIR}"

# Convert all sessions
go-minitrace convert pi \
  --source-dir "${SOURCE_DIR}" \
  --output-dir "${OUTPUT_DIR}"

echo ""
echo "=== Conversion complete ==="
echo ""
echo "Converted sessions:"
find "${OUTPUT_DIR}" -name '*.minitrace.json' | while read -r f; do
  echo "  - $(basename "${f}")"
done
