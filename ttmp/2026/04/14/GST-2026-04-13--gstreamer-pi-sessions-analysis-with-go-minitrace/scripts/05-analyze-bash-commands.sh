#!/usr/bin/env bash
# 05-analyze-bash-commands.sh
# Analyze bash commands from GStreamer sessions

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TICKET_DIR="$(dirname "${SCRIPT_DIR}")"
OUTPUT_DIR="${TICKET_DIR}/output"
QUERY_DIR="${TICKET_DIR}/query-commands"
ARCHIVE_GLOB="${OUTPUT_DIR}/active/*/*.minitrace.json"

MAIN_SESSION="bbf1bdf1-364a-44cb-8cd0-ebcba86dd1ad"

echo "=== Analyzing Bash Commands in GStreamer Sessions ==="
echo ""

echo "--- Top Bash Commands (all sessions) ---"
go-minitrace query commands bash-commands \
  --query-repository "${QUERY_DIR}" \
  --archive-glob "${ARCHIVE_GLOB}" \
  --limit 30 \
  2>&1

echo ""
echo "--- Go-related Commands ---"
go-minitrace query commands bash-commands \
  --query-repository "${QUERY_DIR}" \
  --archive-glob "${ARCHIVE_GLOB}" \
  --command-pattern "go %" \
  --limit 20 \
  2>&1

echo ""
echo "--- Git Commands ---"
go-minitrace query commands bash-commands \
  --query-repository "${QUERY_DIR}" \
  --archive-glob "${ARCHIVE_GLOB}" \
  --command-pattern "git%" \
  --limit 20 \
  2>&1

echo ""
echo "=== Bash analysis complete ==="
