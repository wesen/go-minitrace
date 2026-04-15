# Changelog

## 2026-04-14

- Initial workspace created


## 2026-04-14

Step 1: Set up docmgr ticket, ran go-minitrace help, discovered 5 Pi sessions, created analysis scripts


## 2026-04-14

Step 2: Converted 5 sessions to minitrace, created 7 query commands, analyzed main GStreamer session (bbf1bdf1: 1170 turns, 1385 tools, 24h duration)


## 2026-04-14

Step 3: Created HTML transcript export proposals with 5 user personas, 5 view proposals (Dashboard, Reader, Investigation Board, Query Interface, Mobile), ASCII layouts, and implementation notes


## 2026-04-14

Step 4: Created revised read-only self-contained HTML export proposals with export-time analysis pipeline, 5 analysis modules, 4 view designs, and implementation specification (44KB document)


## 2026-04-14

Step 6: Added two intern-facing implementation guides for Proposal 2 (Chronological Reader) and Proposal 4 (Timeline Visualization), covering annotation workflows, export shaping, browser architecture, pseudocode, and file/API references


## 2026-04-14

Step 7: Simplified Proposal 2 by removing synthetic annotations and heuristics; reframed it as a generic JSON-driven read-only reader and documented concrete reuse of the existing React TranscriptViewer components


## 2026-04-14

Step 8: Began real implementation of simplified Proposal 2 in go-minitrace: export command scaffold, deterministic reader payload, minimal self-contained HTML renderer, static export React entrypoint, dedicated build target, and built-bundle inlining path. Also fixed unrelated glazed config API blocker in pkg/minitracecmd/repositories.go.


## 2026-04-14

Step 9: Validated end-to-end built-bundle export against session bbf1bdf1-364a-44cb-8cd0-ebcba86dd1ad, added export-reader hash navigation (#turn-*/#tool-call-*), and cleaned the dedicated Vite build config to use codeSplitting=false instead of the deprecated inlineDynamicImports option.


## 2026-04-14

Step 10: Hardened built-bundle HTML inlining by stripping leftover <link href> and <script src> asset tags after module inlining, added regression coverage, and revalidated the exported HTML had no remaining root/external asset refs.


## 2026-04-14

Step 11: Fixed large-export payload embedding hazards and the browser crash in the React reader by switching built/template HTML replacement to literal-safe payload insertion, escaping script-tag-dangerous JSON, and guaranteeing `annotations` exports as `[]` instead of `null`. Revalidated the regenerated export in a clean Playwright tab: no console errors, no external asset/data requests, search works, and `#tool-call-*` hash navigation scrolls and highlights correctly. The remaining open validation gap is direct `file://` browser loading, which the Playwright harness itself blocks.


## 2026-04-15

Step 12: Moved the GST-2026-04-13 ticket into `go-minitrace/ttmp`, created two long-form Obsidian notes about the export-reader architecture and the under-the-hood HTML export pipeline, and copied those notes into the ticket as a project report and deep-dive design document.

