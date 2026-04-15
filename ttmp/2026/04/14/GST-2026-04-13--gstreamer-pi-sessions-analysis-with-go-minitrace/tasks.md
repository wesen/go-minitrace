# Tasks

## Implementation plan for Proposal 2 (simplified chronological reader)

### Ticket / design alignment
- [x] Finalize the simplified Proposal 2 scope in ticket docs
- [x] Keep Proposal 2 limited to raw transcript data + original annotations + deterministic indices
- [x] Keep Proposal 2 explicitly free of synthetic annotations and heuristic thread markers

### Backend export foundations
- [x] Add a new `go-minitrace export` command group
- [x] Add `go-minitrace export html` subcommand scaffolding
- [x] Add archive/session loading helpers reusable outside `serve`
- [x] Extract or duplicate the deterministic session-block builder needed for Proposal 2
- [x] Define export payload Go structs for:
  - [x] session metadata
  - [x] blocks
  - [x] turns
  - [x] tool calls
  - [x] annotations
  - [x] deterministic indices
- [x] Build a deterministic search index for the reader payload
- [x] Build annotation indices by scope (`session`, `turn`, `tool_call`)
- [x] Add payload serialization tests

### HTML export shell
- [x] Add a generic HTML shell template with:
  - [x] `#root` mount node
  - [x] embedded JSON payload script tag
  - [x] embedded CSS slot
  - [x] embedded JS slot
- [x] Add Go helpers to inject JSON safely into the template
- [x] Add a minimal export writer that emits a self-contained HTML file
- [x] Add CLI flags for:
  - [x] `--session-id`
  - [x] `--archive-glob`
  - [x] `--output`
  - [x] `--title`

### Frontend reuse / export runtime
- [x] Create a dedicated export frontend entrypoint in `web/src/export`
- [x] Add `loadEmbeddedExport()` helper
- [x] Create an export-specific transcript viewer wrapper that takes embedded JSON instead of live API data
- [x] Reuse or shallow-fork the current components:
  - [x] `BlockCard`
  - [x] `BlockBody`
  - [x] `ToolCallRow`
- [x] Remove annotation mutation UI from the export viewer
- [x] Remove router and RTK Query assumptions from the export viewer
- [x] Keep anchor navigation for turns/tool calls
- [x] Keep offline-only behavior (no network calls)

### Build / packaging
- [x] Add a dedicated Vite build target for the export viewer bundle
- [x] Decide how Go reads the built JS/CSS assets for embedding
- [x] Add a development workflow note for rebuilding export assets
- [x] Ensure final export is a single HTML file

### Validation / tests
- [x] Add tests for session indexing and session selection
- [x] Add tests for block formation and tool-call grouping
- [x] Add tests for annotation indexing
- [x] Add tests for HTML template generation
- [x] Run export against session `bbf1bdf1-364a-44cb-8cd0-ebcba86dd1ad`
- [ ] Verify output opens locally via browser without a server
  - Playwright `file://` navigation is blocked by the harness, so this still needs a manual local-file browser check outside the harness
- [x] Verify no external asset/network requests occur in exported file during browser validation
  - validated via local HTTP serving fallback: only the HTML document itself was requested; no external JS/CSS/data requests were made by the export

### Documentation / ticket hygiene
- [x] Update diary after each implementation slice
- [ ] Keep commits focused and small enough to review easily
- [ ] Relate new files to ticket docs
- [ ] Upload revised bundle to reMarkable after meaningful milestones
