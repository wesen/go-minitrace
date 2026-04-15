---
Title: Export Reader Development Workflow
Ticket: GST-2026-04-13
Status: active
Topics:
    - gstreamer
    - go-minitrace
    - html-export
    - react
    - vite
DocType: reference
Intent: how-to
Owners: []
RelatedFiles:
    - Path: /home/manuel/code/wesen/corporate-headquarters/go-minitrace/cmd/go-minitrace/cmds/export/html.go
      Note: CLI entrypoint for HTML export
    - Path: /home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/exporthtml/render_built.go
      Note: Inlines built export-reader bundle into final HTML
    - Path: /home/manuel/code/wesen/corporate-headquarters/go-minitrace/web/src/export/TranscriptExportViewer.tsx
      Note: Read-only React export viewer
    - Path: /home/manuel/code/wesen/corporate-headquarters/go-minitrace/web/vite.export-reader.config.ts
      Note: Dedicated Vite config for single-bundle export-reader build
ExternalSources: []
Summary: Practical commands and validation steps for rebuilding and inlining the self-contained export-reader bundle
LastUpdated: 2026-04-14T15:15:00-04:00
WhatFor: Rebuilding the dedicated export-reader bundle and generating self-contained HTML exports reliably
WhenToUse: When iterating on Proposal 2 export UI or validating the end-to-end built-bundle export flow
---

# Export reader development workflow

This note documents the practical build/inlining loop for the simplified Proposal 2 self-contained HTML export in `go-minitrace`.

## Goal

Produce a **single self-contained HTML file** for one session that:
- renders a read-only transcript viewer,
- embeds the JSON payload directly,
- embeds the built React export-reader bundle directly,
- does not require a local web server.

## Relevant code paths

### Go side
- `cmd/go-minitrace/cmds/export/html.go`
- `pkg/exporthtml/builder.go`
- `pkg/exporthtml/render.go`
- `pkg/exporthtml/render_built.go`

### Web side
- `web/export-reader.html`
- `web/src/export/loadEmbeddedExport.ts`
- `web/src/export/TranscriptExportViewer.tsx`
- `web/src/export/readerExportMain.tsx`
- `web/vite.export-reader.config.ts`

## Build the dedicated export-reader bundle

From the repo root:

```bash
cd /home/manuel/code/wesen/corporate-headquarters/go-minitrace/web
pnpm build:export-reader
```

Expected output:

```text
web/dist-export-reader/export-reader.html
web/dist-export-reader/static/export-reader.js
```

This dedicated build is preferred for Go-side inlining because it emits one HTML entry and one JS bundle, which is much easier to embed into a final self-contained document than the normal multi-chunk application build.

## Generate a self-contained export HTML file

From the repo root:

```bash
cd /home/manuel/code/wesen/corporate-headquarters/go-minitrace

go run ./cmd/go-minitrace export html \
  --archive-glob '/home/manuel/code/wesen/corporate-headquarters/go-minitrace/ttmp/2026/04/14/GST-2026-04-13--gstreamer-pi-sessions-analysis-with-go-minitrace/output/active/*/*.minitrace.json' \
  --session-id bbf1bdf1-364a-44cb-8cd0-ebcba86dd1ad \
  --web-dist-dir ./web/dist-export-reader \
  --output /tmp/gstreamer-reader-react.html
```

Expected output:

```text
Wrote /tmp/gstreamer-reader-react.html
```

## Fallback path

If the dedicated built bundle is unavailable, the exporter still supports the minimal built-in renderer:

```bash
cd /home/manuel/code/wesen/corporate-headquarters/go-minitrace

go run ./cmd/go-minitrace export html \
  --archive-glob '/home/manuel/code/wesen/corporate-headquarters/go-minitrace/ttmp/2026/04/14/GST-2026-04-13--gstreamer-pi-sessions-analysis-with-go-minitrace/output/active/*/*.minitrace.json' \
  --session-id bbf1bdf1-364a-44cb-8cd0-ebcba86dd1ad \
  --output /tmp/gstreamer-reader-minimal.html
```

This fallback remains useful as a resilience path and as a debugging baseline for payload/render issues.

## Recommended validation loop

### Go validation
```bash
cd /home/manuel/code/wesen/corporate-headquarters/go-minitrace
go test ./pkg/exporthtml ./pkg/minitracecmd ./cmd/go-minitrace/cmds/export ./cmd/go-minitrace/cmds/query ./cmd/go-minitrace/cmds/serve -count=1
```

### Web validation
```bash
cd /home/manuel/code/wesen/corporate-headquarters/go-minitrace/web
pnpm build
pnpm build:export-reader
```

### Lightweight export artifact checks
Use shell/script inspection to confirm the export is plausibly self-contained:

```bash
python - <<'PY'
from pathlib import Path
import re
p = Path('/tmp/gstreamer-reader-react.html')
s = p.read_text()
print('has inline module:', '<script type="module">' in s)
print('has payload:', 'id="minitrace-export-data"' in s)
print('has root-relative refs:', bool(re.findall(r'(?:src|href)="(/[^"]+)"', s)))
print('has external refs:', bool(re.findall(r'(?:src|href)="(https?://[^"]+)"', s)))
PY
```

## Notes

- The dedicated export reader now supports direct hash navigation for:
  - `#turn-<idx>`
  - `#tool-call-<id>`
- The generated HTML is intended to be portable and opened directly from disk.
- Avoid relying on the normal `web/dist` application output for embedding because it is chunk-split and less suitable for single-file export inlining.
