#!/usr/bin/env bash
# 04-run-query-commands.sh
# Run all custom query commands against the GStreamer sessions

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TICKET_DIR="$(dirname "${SCRIPT_DIR}")"
OUTPUT_DIR="${TICKET_DIR}/output"
QUERY_DIR="${TICKET_DIR}/query-commands"
ARCHIVE_GLOB="${OUTPUT_DIR}/active/*/*.minitrace.json"

echo "=== Running GStreamer Query Commands ==="
echo "Archive: ${ARCHIVE_GLOB}"
echo "Queries: ${QUERY_DIR}"
echo ""

# Main GStreamer session ID
MAIN_SESSION="bbf1bdf1-364a-44cb-8cd0-ebcba86dd1ad"

echo "--- 1. GStreamer Session List ---"
go-minitrace query commands gstreamer-session-list \
  --query-repository "${QUERY_DIR}" \
  --archive-glob "${ARCHIVE_GLOB}" \
  2>&1

echo ""
echo "--- 2. Tool Analysis for Main Session (${MAIN_SESSION}) ---"
go-minitrace query commands gstreamer-tool-analysis \
  --query-repository "${QUERY_DIR}" \
  --archive-glob "${ARCHIVE_GLOB}" \
  --session-id "${MAIN_SESSION}" \
  --limit 20 \
  2>&1

echo ""
echo "--- 3. Top Modified Files (all sessions) ---"
go-minitrace query commands top-modified-files \
  --query-repository "${QUERY_DIR}" \
  --archive-glob "${ARCHIVE_GLOB}" \
  --limit 30 \
  2>&1

echo ""
echo "=== Query commands complete ==="
