---
Title: Diary
Ticket: GMT-007
Status: active
Topics:
    - go-minitrace
    - transcript-analysis
    - html-export
    - readonly-exports
    - ui-design
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: pkg/exporttimeline/loader.go
      Note: Programmatic loader that executes timeline query commands and decodes their results
    - Path: pkg/exporttimeline/types.go
      Note: Typed row/result models for the SQL-backed Proposal 4 bridge layer
    - Path: ttmp/2026/04/15/GMT-007--implement-proposal-4-timeline-visualization-for-transcript-exports/query-commands/timeline/timeline-buckets.sql
      Note: Initial SQL-first bucket contract for Proposal 4
    - Path: ttmp/2026/04/15/GMT-007--implement-proposal-4-timeline-visualization-for-transcript-exports/query-commands/timeline/timeline-file-activity.sql
      Note: Per-file per-bucket activity query used to seed file bands
    - Path: ttmp/2026/04/15/GMT-007--implement-proposal-4-timeline-visualization-for-transcript-exports/query-commands/timeline/timeline-idle-windows.sql
      Note: Contiguous idle-window detection query
    - Path: ttmp/2026/04/15/GMT-007--implement-proposal-4-timeline-visualization-for-transcript-exports/query-commands/timeline/timeline-phase-signals.sql
      Note: Bucket-level phase-signal query
    - Path: ttmp/2026/04/15/GMT-007--implement-proposal-4-timeline-visualization-for-transcript-exports/query-commands/timeline/timeline-thread-signals.sql
      Note: Simple file-based thread candidate query
ExternalSources: []
Summary: Chronological implementation diary for Proposal 4 timeline visualization work.
LastUpdated: 2026-04-15T02:05:00-04:00
WhatFor: Capture the implementation sequence, rationale, validation results, and review guidance for the timeline export work.
WhenToUse: Read this when continuing Proposal 4 work or reviewing how the SQL-first timeline foundation was established.
---



# Diary

## Goal

Track the implementation of Proposal 4: a SQL-first, read-only, self-contained timeline visualization export for transcript sessions.

## Step 1: SQL-first foundation for Proposal 4 timeline analysis

I started Proposal 4 by treating the ticket itself as an implementation boundary, not just a design placeholder. The first slice needed to answer a practical question before any browser visualization work could begin: what temporal analysis should live in SQL, what should the payload contract look like, and what reusable query commands do we need before writing exporter code?

This was the right place to begin because Proposal 4 is supposed to be more than a one-off visualization. The user explicitly wanted as much of the analysis as possible to be reusable as SQL commands. That changes the architecture materially: Go becomes an orchestration and payload-merging layer, while SQL becomes the first-class home for buckets, file activity, idle detection, and signal extraction.

### Prompt Context

**User prompt (verbatim):** "Ok, go ahead and work task by task, committing at appropriate intervals and keeping a diary"

**Assistant interpretation:** Start executing Proposal 4 in focused slices, record the work in a formal ticket diary, and keep commits tight enough that each slice is reviewable.

**Inferred user intent:** Turn Proposal 4 from planning into real implementation work while keeping strong ticket hygiene and a clear narrative for future continuation.

**Commit (code):** pending — "docs: add SQL-first proposal 4 timeline foundation"

### What I did

- Created a real diary document for GMT-007.
- Updated the Proposal 4 implementation guide to make the architecture explicitly SQL-first.
- Finalized the initial payload and integration decisions for v1:
  - default bucket size is `30` minutes,
  - original annotations remain separate from derived timeline markers,
  - all timeline clicks should resolve to Proposal 2-compatible `#turn-<idx>` reader jumps.
- Added a result-contract section documenting the expected output columns for the first wave of timeline query commands.
- Created these ticket-local query commands under:
  - `ttmp/2026/04/15/GMT-007--implement-proposal-4-timeline-visualization-for-transcript-exports/query-commands/timeline/`
  - `timeline-buckets.sql`
  - `timeline-file-activity.sql`
  - `timeline-idle-windows.sql`
  - `timeline-phase-signals.sql`
  - `timeline-thread-signals.sql`
- Validated the commands with `go-minitrace query commands ... --query-repository ...` against real local archive sessions in `./output/active/*/*.minitrace.json`.
- Fixed an early issue where file-path extraction treated JSON `null` as the literal string `"null"`, which polluted file-activity and thread-signal outputs.
- Marked the first task slice complete in `tasks.md`.

### Why

Proposal 4 will be much easier to maintain if the temporal analysis exists as reusable query commands that can be:

- inspected directly,
- run outside the timeline UI,
- compared with exported payloads,
- and reused in other research or CLI workflows.

Starting with query commands also makes it easier to validate the timeline logic without simultaneously debugging HTML rendering and browser behavior.

### What worked

- `timeline-buckets.sql` executed successfully and produced bucket rows with useful counts and jump-turn indices.
- `timeline-idle-windows.sql` successfully detected contiguous low-activity windows on a real session.
- `timeline-phase-signals.sql` produced bucket-level dominant signal rows suitable for later phase labeling.
- `timeline-file-activity.sql` and `timeline-thread-signals.sql` both executed after the null-path fix and now emit file-based activity/thread candidates rather than `"null"` pseudo-files.
- The SQL-first direction now has a documented contract in both the design doc and tasks.

### What didn't work

The first version of file-path extraction used a simple `NULLIF(..., '')`, which did not treat JSON `null` as absent. That caused bad rows such as:

```text
file_path = null
```

to be treated like a real file key in file-activity and thread-signal outputs.

The fix was to normalize both empty string and literal `null` string values to SQL `NULL` before aggregation.

### What I learned

A few important patterns became clear immediately:

1. Proposal 4 really does want a different architecture than Proposal 2.
   - Proposal 2 is payload-first and literal.
   - Proposal 4 is analysis-first and derived.

2. SQL is a good fit for:
   - bucketization,
   - per-file activity series,
   - idle candidate detection,
   - bucket-level phase signals.

3. SQL is only a partial fit for thread reconstruction.
   - file-based thread candidates are easy enough,
   - semantic threads will probably still need imported/manual spans or a later higher-level analysis pass.

### What was tricky to build

The sharp edge here was not the overall SQL syntax. It was the mismatch between JSON null handling and file-oriented aggregation. In transcript analytics, a lot of fields are optional. If you do not normalize those optional JSON values early, you can accidentally create fake categories, fake files, and misleading aggregate rows.

The other tricky design choice was deciding where to stop with SQL. It is tempting to force every future concept into SQL, but Proposal 4 should be SQL-first, not SQL-only. Manual/imported phases and thread spans still make sense as higher-level timeline inputs, especially for the first trustworthy implementation.

### What warrants a second pair of eyes

- Whether the current query-command result contracts are the right long-term handoff boundary into the future Go payload merger.
- Whether `timeline-phase-signals.sql` is too biased toward execution-heavy sessions and needs stronger signals for read/modify/test-heavy projects.
- Whether the file-based thread candidate query is useful enough as a first pass or whether it should already incorporate more command-pattern logic.

### What should be done in the future

- Build the Go-side merger that consumes these query-command outputs into a single timeline payload.
- Add a workflow note documenting how these query commands feed the exporter.
- Decide whether phase/thread spans should be imported through a dedicated file format or derived from existing annotation/import machinery.

### Code review instructions

Start here:
- `ttmp/2026/04/15/GMT-007--implement-proposal-4-timeline-visualization-for-transcript-exports/design-doc/01-proposal-4-timeline-visualization-implementation-and-analysis-guide.md`
- `ttmp/2026/04/15/GMT-007--implement-proposal-4-timeline-visualization-for-transcript-exports/query-commands/timeline/timeline-buckets.sql`
- `ttmp/2026/04/15/GMT-007--implement-proposal-4-timeline-visualization-for-transcript-exports/query-commands/timeline/timeline-file-activity.sql`
- `ttmp/2026/04/15/GMT-007--implement-proposal-4-timeline-visualization-for-transcript-exports/query-commands/timeline/timeline-idle-windows.sql`
- `ttmp/2026/04/15/GMT-007--implement-proposal-4-timeline-visualization-for-transcript-exports/query-commands/timeline/timeline-phase-signals.sql`
- `ttmp/2026/04/15/GMT-007--implement-proposal-4-timeline-visualization-for-transcript-exports/query-commands/timeline/timeline-thread-signals.sql`

Validate with commands like:

```bash
cd /home/manuel/code/wesen/corporate-headquarters/go-minitrace
repo='./ttmp/2026/04/15/GMT-007--implement-proposal-4-timeline-visualization-for-transcript-exports/query-commands'

# bucket view
go run ./cmd/go-minitrace query commands timeline-buckets \
  --query-repository "$repo" \
  --archive-glob './output/active/*/*.minitrace.json' \
  --session-id 019d0112-69ba-7232-9b14-875797183903

# file activity view
go run ./cmd/go-minitrace query commands timeline-file-activity \
  --query-repository "$repo" \
  --archive-glob './output/active/*/*.minitrace.json' \
  --session-id 019d03aa-ddee-7403-83d1-2ff075e82d50
```

### Technical details

**Task slice completed:**
- finalize payload/jump-target defaults
- create initial SQL-first timeline query commands
- validate them on real archive sessions
- document result contracts

**Key local validation sessions:**
- `019d0112-69ba-7232-9b14-875797183903`
- `019d03aa-ddee-7403-83d1-2ff075e82d50`

## Step 2: Add a Go-side SQL timeline loader and typed result model

Once the SQL-first query-command layer existed, the next missing piece was a Go-side bridge. Proposal 4 cannot stop at “we have useful queries”; it needs a programmatic way to execute those queries, capture their results in typed structures, and hand them to the future timeline exporter.

This slice focused on that bridge layer and intentionally stopped short of rendering. The goal was to add a small `pkg/exporttimeline` package that knows how to load the query-command results for one session and present them as a typed `SQLTimelineData` bundle.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Continue Proposal 4 in focused slices and start converting the SQL-first analysis work into a reusable programmatic interface.

**Inferred user intent:** Make the timeline plan executable in code, not just as ticket-local SQL files.

**Commit (code):** pending — "feat: add sql-backed timeline data loader"

### What I did

- Added a new package:
  - `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/exporttimeline`
- Added typed row/result models in:
  - `pkg/exporttimeline/types.go`
- Added a SQL-backed loader in:
  - `pkg/exporttimeline/loader.go`
- Implemented `LoadSQLTimelineData(...)` to:
  - load the configured query repositories,
  - open DuckDB,
  - load the archive once,
  - render and execute the timeline query commands,
  - decode each result set into typed Go structs.
- Reused existing internals rather than inventing a new execution path:
  - `minitracecmd.LoadConfiguredCatalog(...)`
  - `minitracecmd.RenderCommand(...)`
  - `query.OpenConnection(...)`
  - `query.LoadArchive(...)`
  - `query.NormalizeValue(...)`
- Added a quick validation run with an ad hoc Go program to prove the loader can fetch all five timeline result sets for a real session.
- Fixed a first-pass bug where some command defaults were not automatically materialized through direct command rendering, causing rendered SQL like:
  ```text
  WHERE file_rank <= <no value>
  ```
  The loader now passes explicit defaults for the current timeline commands.

### Why

Proposal 4 needs a seam between reusable SQL analysis and the future exporter/browser payload. Without that seam, we would either:

- re-implement the SQL logic again in Go, or
- leave the timeline work stranded as only CLI query commands.

The new loader is the first step toward treating those SQL commands as a proper backend analysis API for the exporter.

### What worked

- `pkg/exporttimeline` compiles cleanly.
- `go test ./pkg/exporttimeline -count=1` passes.
- The ad hoc validation run confirmed the loader can execute all timeline query commands for a real archive session and return typed slices.
- The package defaults are now strong enough to call the current timeline queries without relying on CLI-layer flag-default injection.

### What didn't work

The first loader implementation assumed the command-rendering path would automatically materialize all sqleton default values. That assumption was wrong for direct programmatic rendering in this path.

Exact failure:

```text
Parser Error: syntax error at or near "<"
LINE 46:   WHERE file_rank <= <no value>
```

The fix was to pass explicit default values from the loader for the current timeline commands (`top_files`, `min_operations`, `min_idle_buckets`, etc.) instead of assuming the renderer would supply them.

### What I learned

The CLI and the direct library path do not automatically share all of the same conveniences. The command definitions contain defaults, but if we call the render path directly with sparse value maps, we cannot assume every layer that the CLI exercises is present.

That means Proposal 4 likely needs one of these long-term follow-ups:

- a generic helper that materializes command defaults before rendering, or
- explicit timeline-specific orchestration that owns the runtime defaults clearly.

For now, explicit defaults in the loader are acceptable and make the behavior obvious.

### What was tricky to build

The tricky part was deciding how much abstraction to add. It would have been easy to overbuild a generalized query-command execution framework. Instead, I kept this slice narrow:

- typed row structs,
- one loader,
- one generic row decoder,
- and one direct bridge from query commands into `SQLTimelineData`.

That is enough to unblock the next task slice without prematurely locking in a bigger abstraction.

### What warrants a second pair of eyes

- Whether `LoadSQLTimelineData(...)` should eventually consume command defaults automatically from the command definitions rather than hard-coding them for the first wave of timeline commands.
- Whether the current typed row models are the right boundary, or whether a second normalized payload layer should sit on top before the actual exporter.
- Whether the package should stay named `exporttimeline` or eventually merge into a broader export subsystem alongside `exporthtml`.

### What should be done in the future

- Build the next layer that converts these typed SQL result sets into normalized bucket arrays and final timeline payload structs.
- Decide whether to add tests around generic row decoding and default handling.
- Start the first renderer-oriented slice only after the payload merger exists.

### Code review instructions

Start here:
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/exporttimeline/types.go`
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/exporttimeline/loader.go`

Then cross-check the SQL contracts they depend on:
- `ttmp/2026/04/15/GMT-007--implement-proposal-4-timeline-visualization-for-transcript-exports/query-commands/timeline/timeline-buckets.sql`
- `ttmp/2026/04/15/GMT-007--implement-proposal-4-timeline-visualization-for-transcript-exports/query-commands/timeline/timeline-file-activity.sql`
- `ttmp/2026/04/15/GMT-007--implement-proposal-4-timeline-visualization-for-transcript-exports/query-commands/timeline/timeline-idle-windows.sql`
- `ttmp/2026/04/15/GMT-007--implement-proposal-4-timeline-visualization-for-transcript-exports/query-commands/timeline/timeline-phase-signals.sql`
- `ttmp/2026/04/15/GMT-007--implement-proposal-4-timeline-visualization-for-transcript-exports/query-commands/timeline/timeline-thread-signals.sql`

Validate with:

```bash
cd /home/manuel/code/wesen/corporate-headquarters/go-minitrace
go test ./pkg/exporttimeline -count=1
```

Optional ad hoc validation:

```bash
cd /home/manuel/code/wesen/corporate-headquarters/go-minitrace
cat <<'EOF' >/tmp/check_timeline_loader.go
package main
import (
  "context"
  "fmt"
  "github.com/go-go-golems/go-minitrace/pkg/exporttimeline"
)
func main(){
  data, err := exporttimeline.LoadSQLTimelineData(context.Background(), exporttimeline.LoadOptions{
    QueryRepositories: []string{"./ttmp/2026/04/15/GMT-007--implement-proposal-4-timeline-visualization-for-transcript-exports/query-commands"},
    ArchiveGlobs: []string{"./output/active/*/*.minitrace.json"},
    SessionID: "019d03aa-ddee-7403-83d1-2ff075e82d50",
  })
  if err != nil { panic(err) }
  fmt.Printf("buckets=%d fileActivity=%d idle=%d phaseSignals=%d threadSignals=%d\n", len(data.Buckets), len(data.FileActivity), len(data.IdleWindows), len(data.PhaseSignals), len(data.ThreadSignals))
}
EOF
go run /tmp/check_timeline_loader.go
```

### Technical details

**New package:**
- `pkg/exporttimeline`

**Current exported entrypoint:**
- `LoadSQLTimelineData(ctx, opts)`

**Current task slice completed:**
- add timeline export package and typed row models
- add SQL query-result loader for Proposal 4 timeline commands
