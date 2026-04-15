#!/usr/bin/env bash
# 06-analyze-web-searches.sh
# Analyze web searches from GStreamer sessions

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TICKET_DIR="$(dirname "${SCRIPT_DIR}")"
OUTPUT_DIR="${TICKET_DIR}/output"
QUERY_DIR="${TICKET_DIR}/query-commands"
ARCHIVE_GLOB="${OUTPUT_DIR}/active/*/*.minitrace.json"

MAIN_SESSION="bbf1bdf1-364a-44cb-8cd0-ebcba86dd1ad"

echo "=== Analyzing Web Searches in GStreamer Sessions ==="
echo ""

echo "--- All Web Searches ---"
go-minitrace query commands web-searches \
  --query-repository "${QUERY_DIR}" \
  --archive-glob "${ARCHIVE_GLOB}" \
  --limit 30 \
  2>&1 | head -50

echo ""
echo "--- GStreamer-related Searches ---"
go-minitrace query commands web-searches \
  --query-repository "${QUERY_DIR}" \
  --archive-glob "${ARCHIVE_GLOB}" \
  --query-pattern "%gstreamer%" \
  --limit 20 \
  2>&1 | head -30

echo ""
echo "=== Web search analysis complete ==="
