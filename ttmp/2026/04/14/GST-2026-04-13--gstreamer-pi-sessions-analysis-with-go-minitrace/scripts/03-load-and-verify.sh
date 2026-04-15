#!/usr/bin/env bash
# 03-load-and-verify.sh
# Load archive and verify contents without running a query

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TICKET_DIR="$(dirname "${SCRIPT_DIR}")"
OUTPUT_DIR="${TICKET_DIR}/output"
ARCHIVE_GLOB="${OUTPUT_DIR}/active/*/*.minitrace.json"

echo "=== Loading archive to verify contents ==="
echo "Archive glob: ${ARCHIVE_GLOB}"
echo ""

# Check if files exist
file_count=$(find "${OUTPUT_DIR}" -name '*.minitrace.json' 2>/dev/null | wc -l)
if [[ "${file_count}" -eq 0 ]]; then
  echo "ERROR: No .minitrace.json files found in ${OUTPUT_DIR}"
  echo "Run 02-convert-sessions.sh first"
  exit 1
fi

echo "Found ${file_count} minitrace files"
echo ""

# Load and verify
go-minitrace query duckdb \
  --archive-glob "${ARCHIVE_GLOB}" \
  --load-only

echo ""
echo "=== Verification complete ==="
