#!/usr/bin/env bash
# 01-discover-sessions.sh
# Discover Pi sessions for the screencast-studio GStreamer work

set -euo pipefail

TARGET_CWD="/home/manuel/code/wesen/2026-04-09--screencast-studio"
SOURCE_ROOT="${HOME}/.pi/agent/sessions"
LIMIT="${1:-20}"

echo "=== Discovering Pi sessions for: ${TARGET_CWD} ==="
echo ""

# The Pi session directory is slugged
slug_path() {
  local path="$1"
  printf '%s' "${path}" | sed 's#^/##; s#/#-#g; s#^#--#; s#$#--#'
}

SESSION_DIR="${SOURCE_ROOT}/$(slug_path "${TARGET_CWD}")"

if [[ ! -d "${SESSION_DIR}" ]]; then
  echo "ERROR: Session directory not found: ${SESSION_DIR}" >&2
  exit 1
fi

echo "Session directory: ${SESSION_DIR}"
echo ""

# Output TSV header
printf '%s\t%s\t%s\t%s\t%s\n' "timestamp" "session_id" "bytes" "human_size" "path"

# Find and list sessions
find "${SESSION_DIR}" -maxdepth 1 -type f -name '*.jsonl' | sort -r | while read -r file; do
  base="$(basename "${file}")"
  session_id="${base##*_}"
  session_id="${session_id%.jsonl}"
  timestamp="${base%%_*}"
  bytes="$(stat -c '%s' "${file}")"
  
  # Human readable size
  if command -v numfmt >/dev/null 2>&1; then
    human_size="$(numfmt --to=iec-i --suffix=B "${bytes}" 2>/dev/null || echo "${bytes}B")"
  else
    human_size="${bytes}B"
  fi
  
  printf '%s\t%s\t%s\t%s\t%s\n' "${timestamp}" "${session_id}" "${bytes}" "${human_size}" "${file}"
done | head -n "${LIMIT}"

echo ""
echo "Total sessions found: $(find "${SESSION_DIR}" -maxdepth 1 -type f -name '*.jsonl' | wc -l)"
