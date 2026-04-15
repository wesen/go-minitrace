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
    - Path: cmd/go-minitrace/cmds/export/timeline.go
      Note: CLI entrypoint for self-contained Proposal 4 timeline export
    - Path: pkg/exporttimeline/builder.go
      Note: Builds timeline export metadata + payload from session and SQL timeline rows
    - Path: pkg/exporttimeline/builder_test.go
      Note: Timeline export builder coverage
    - Path: pkg/exporttimeline/export.go
      Note: Top-level timeline export envelope types and build options
    - Path: pkg/exporttimeline/loader.go
      Note: Programmatic loader that executes timeline query commands and decodes their results
    - Path: pkg/exporttimeline/manual.go
      Note: Annotation-backed manual/import marker parser and extractor for Proposal 4 phases/threads
    - Path: pkg/exporttimeline/manual_test.go
      Note: Focused coverage for Proposal 4 manual/import annotation parsing
    - Path: pkg/exporttimeline/payload.go
      Note: Normalized timeline payload builder on top of SQL result sets
    - Path: pkg/exporttimeline/payload_test.go
      Note: Focused coverage for file-series densification and jump-hash shaping
    - Path: pkg/exporttimeline/render.go
      Note: Self-contained timeline HTML renderer with embedded template assets
    - Path: pkg/exporttimeline/render_payload.go
      Note: Script-tag-safe JSON payload marshaling for timeline exports
    - Path: pkg/exporttimeline/render_test.go
      Note: Timeline renderer regression coverage for payload escaping
    - Path: pkg/exporttimeline/templates/timeline.css
      Note: Timeline export styling and layout
    - Path: pkg/exporttimeline/templates/timeline.html.tmpl
      Note: Standalone timeline export HTML shell
    - Path: pkg/exporttimeline/templates/timeline.js
      Note: SVG timeline runtime (heatmap
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
LastUpdated: 2026-04-15T04:30:00-04:00
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

**Commit (code):** `bab166a0cbc639fd0f2fab488099791c8740b4bf` — "docs: add sql-first proposal 4 timeline foundation"

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

**Commit (code):** `7c3db9a57c3be1775a7eb2f264dba7e972295364` — "feat: add sql-backed timeline data loader"

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

## Step 3: Build the first normalized timeline payload from SQL results

With the SQL loader in place, the next missing layer was the actual payload shape the future renderer will want. Raw SQL row slices are useful for debugging, but they are not yet a timeline payload. The browser will need dense bucket arrays, file series aligned to bucket indices, idle windows in a stable form, and deterministic reader jump hashes.

This slice built that first normalized payload layer in `pkg/exporttimeline` without yet committing to the final HTML renderer. That keeps the work staged: SQL first, Go merger second, browser renderer third.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Continue the Proposal 4 implementation by turning SQL outputs into a concrete timeline payload the browser can eventually consume.

**Inferred user intent:** Progress steadily from analysis primitives toward a real export pipeline while keeping each intermediate layer testable.

**Commit (code):** `db498ac363db8ed552b23a6140c9abab48f7261a` — "feat: build normalized timeline payload from sql results"

### What I did

- Added normalized timeline payload types in:
  - `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/exporttimeline/payload.go`
- Added:
  - `TimelinePayload`
  - `TimelineBucket`
  - `TimelineFileSeries`
  - `TimelineIdleWindow`
  - `TimelinePhaseSignal`
  - `TimelineThreadSignal`
- Implemented `BuildTimelinePayload(...)` to:
  - convert bucket rows into stable payload buckets,
  - attach `#turn-<idx>` jump hashes,
  - build dense file series arrays from sparse SQL rows,
  - carry idle-window rows into normalized payload structures,
  - attach bucket-derived jump targets to phase signals,
  - attach jump hashes to thread-signal spans.
- Added `payload_test.go` to verify:
  - bucket jump hashes,
  - dense series materialization,
  - dominant operation placement,
  - phase-signal jump hashing,
  - thread-signal jump hashing.
- Ran package tests:
  ```bash
  go test ./pkg/exporttimeline -count=1
  ```
- Performed an ad hoc end-to-end validation by loading SQL timeline data for a real session and then building the normalized payload.

### Why

This layer is the bridge between analysis and rendering. It prevents the future browser runtime from needing to understand sparse SQL result tables directly and gives the exporter a place to normalize jump-target behavior consistently.

### What worked

- The package now has a proper normalized payload model instead of only raw SQL row slices.
- Dense file series generation works from sparse per-file per-bucket rows.
- Bucket, phase-signal, and thread-signal objects now all carry Proposal 2-compatible reader hashes.
- `payload_test.go` gives us at least one focused unit test around payload shaping logic.
- The ad hoc validation confirmed a real session can go through:
  - SQL query commands,
  - typed SQL loader,
  - normalized timeline payload builder.

### What didn't work

The first version of `BuildTimelinePayload(...)` tried to coerce `IdleWindowRow` directly into `TimelineIdleWindow` with a type conversion. That failed because they are distinct named struct types.

Fix applied:
- switched to explicit field-by-field mapping for idle windows.

### What I learned

The timeline implementation is naturally splitting into three useful boundaries:

1. **query commands**
   - reusable temporal analysis logic
2. **Go payload normalization**
   - stable browser-oriented data model
3. **browser rendering**
   - visual layout and interaction only

That split feels right and is already making the work easier to reason about.

### What was tricky to build

The subtle part here was deciding which structures should already become dense arrays and which should remain sparse row-style records.

- buckets should remain a row-per-bucket array,
- file activity really benefits from dense aligned arrays,
- phase/thread structures should keep their own semantics and just gain jump-target normalization.

This is important because premature densification of every structure would make the payload larger and less clear.

### What warrants a second pair of eyes

- Whether `TimelinePhaseSignal` should stay as bucket-local signals or evolve into a higher-level phase-candidate structure before browser rendering begins.
- Whether file-series payloads should eventually include per-bucket jump targets for more precise file-band click behavior.
- Whether the current normalized payload belongs in `pkg/exporttimeline` permanently or should later merge with a broader export payload subsystem.

### What should be done in the future

- Add manual/imported phase-span merging.
- Add manual/imported thread-span merging.
- Build the first self-contained timeline HTML renderer on top of the normalized payload.

### Code review instructions

Start here:
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/exporttimeline/payload.go`
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/exporttimeline/payload_test.go`
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/exporttimeline/loader.go`

Validate with:

```bash
cd /home/manuel/code/wesen/corporate-headquarters/go-minitrace
go test ./pkg/exporttimeline -count=1
```

Optional end-to-end ad hoc validation:

```bash
cd /home/manuel/code/wesen/corporate-headquarters/go-minitrace
cat <<'EOF' >/tmp/check_timeline_payload.go
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
  payload, err := exporttimeline.BuildTimelinePayload(data, exporttimeline.BuildPayloadOptions{BucketMinutes: 30})
  if err != nil { panic(err) }
  fmt.Printf("bucketCount=%d fileSeries=%d idle=%d phaseSignals=%d threadSignals=%d firstBucketHash=%q\n", payload.BucketCount, len(payload.FileSeries), len(payload.IdleWindows), len(payload.PhaseSignals), len(payload.ThreadSignals), payload.Buckets[0].JumpHash)
}
EOF
go run /tmp/check_timeline_payload.go
```

### Technical details

**Current task slice completed:**
- build bucket arrays from SQL output
- build file series from SQL output
- build idle windows from SQL output
- emit deterministic reader jump targets for buckets/phases/threads

## Step 4: Merge manual/imported markers with SQL-derived phases and threads

The next slice completed the remaining Go-side payload-assembly tasks that were still open in GMT-007: bringing manual/imported phase markers and thread spans into the normalized timeline payload, and merging them with the SQL-derived signal data.

I deliberately kept this slice on the backend side. The browser renderer is still not started. The immediate goal was to make the payload itself rich enough that a future renderer can show both:

- SQL-derived temporal structure, and
- human/imported timeline markers that were added intentionally.

### Prompt Context

**User prompt (verbatim):** "ok go ahead."

**Assistant interpretation:** Continue from the normalized payload slice into the remaining Go-side merge work before starting browser rendering.

**Inferred user intent:** Finish the payload layer so Proposal 4 can move into renderer work with a stable, richer data model.

**Commit (code):** `8ee06764413c326a47448c7231ae80577fb2ff15` — "feat: merge manual and derived timeline markers"

### What I did

- Extended the normalized payload in:
  - `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/exporttimeline/payload.go`
- Added explicit higher-level timeline structures:
  - `TimelinePhaseMarker`
  - `TimelineThreadSegment`
  - `TimelineThreadSpan`
- Kept the lower-level signal arrays too:
  - `PhaseSignals`
  - `ThreadSignals`
- Added manual/import parsing support in:
  - `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/exporttimeline/manual.go`
- Added a concrete first convention for manual/imported Proposal 4 markers using existing synced annotations:
  - session-scoped annotation
  - tag `timeline-phase` or `timeline-thread`
  - JSON payload stored in `annotation.content.detail`
- Implemented `ExtractManualTimelineMarkers(...)`.
- Updated `BuildTimelinePayload(...)` so it can now:
  - parse manual phase/thread markers from annotations,
  - build derived phase markers from contiguous SQL phase-signal buckets,
  - build grouped derived thread spans from SQL thread-signal rows,
  - merge manual and derived phase/thread structures,
  - keep Proposal 2-compatible `#turn-*` jump hashes on the merged structures.
- Added tests in:
  - `pkg/exporttimeline/manual_test.go`
  - `pkg/exporttimeline/payload_test.go`
- Updated the Proposal 4 design doc to record the current annotation-based manual/import convention.

### Why

Proposal 4 is not just a heatmap. The design explicitly calls for trustworthy manual/imported phase and thread structures before heavy automatic heuristics. Without this slice, the payload still only knew about SQL-derived signals, which are good hints but not yet the full timeline story.

This merge layer also creates a practical bridge to the existing annotation workflow instead of forcing a brand-new archive schema immediately.

### What worked

- The payload now contains both low-level signals and higher-level merged markers/spans.
- Manual/imported markers can be carried through ordinary synced annotations rather than needing a new file format first.
- Derived phase signals are now grouped into phase-marker spans instead of staying only bucket-local.
- Derived thread signals are now grouped into segmented thread spans per file path.
- Manual phase markers take precedence over overlapping SQL-derived phase spans, which keeps the eventual phase ribbon cleaner.
- Exact manual thread spans can suppress exact duplicate SQL-derived thread spans for the same file path while still preserving non-duplicate derived spans.
- `go test ./pkg/exporttimeline -count=1` passes.
- A real-session ad hoc run still works after the merge changes and now emits `phaseMarkers` in addition to raw `phaseSignals`.

### What didn't work

There was no major blocker in this slice, but there was one modeling decision that needed care:

- if derived thread signals are grouped into segmented spans by file path, we need a merge rule that does not accidentally throw away non-overlapping derived segments just because one manual thread exists.

I kept the first implementation conservative:

- exact duplicate manual thread spans suppress matching derived spans,
- otherwise both can coexist,
- and manual phase markers suppress overlapping derived phase markers because the phase ribbon benefits more from precedence than the thread view does.

### What I learned

The payload is now settling into a more stable shape with three distinct layers:

1. **raw SQL-derived signals**
   - good for debugging and future analysis reuse
2. **merged timeline markers/spans**
   - the structures the browser will actually want to render
3. **reader jump links**
   - the connective tissue back into Proposal 2

That split feels strong. It lets the renderer stay simple while still preserving the lower-level evidence that produced the rendered view.

### What was tricky to build

The tricky part was not JSON decoding. It was picking a manual/import convention that is immediately usable without destabilizing the archive schema.

Using existing `minitrace.Annotation` records with tags plus detail-JSON turned out to be a good pragmatic compromise:

- it respects the current schema,
- it works with existing SQLite/import/sync workflows,
- and it gives Proposal 4 a real operator path for curated markers right away.

### What warrants a second pair of eyes

- Whether the current annotation tag/detail convention is the right long-term import path or only a short-term bridge.
- Whether phase-marker precedence should stay “manual beats overlapping derived” or become more configurable later.
- Whether derived thread grouping should eventually preserve more per-segment evidence than just grouped file-path segments.

### What should be done in the future

- Add a dedicated workflow note/playbook for preparing Proposal 4 phase/thread marker annotations.
- Start the browser rendering slice now that the payload includes merged phase markers and thread spans.
- Decide later whether a dedicated sidecar import format is still needed once the annotation-based workflow is exercised on real sessions.

### Code review instructions

Start here:
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/exporttimeline/manual.go`
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/exporttimeline/payload.go`
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/exporttimeline/manual_test.go`
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/exporttimeline/payload_test.go`

Then cross-check the updated design note:
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/ttmp/2026/04/15/GMT-007--implement-proposal-4-timeline-visualization-for-transcript-exports/design-doc/01-proposal-4-timeline-visualization-implementation-and-analysis-guide.md`

Validate with:

```bash
cd /home/manuel/code/wesen/corporate-headquarters/go-minitrace
go test ./pkg/exporttimeline -count=1
```

Optional ad hoc validation:

```bash
cd /home/manuel/code/wesen/corporate-headquarters/go-minitrace
cat <<'EOF' >/tmp/check_timeline_markers.go
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
  payload, err := exporttimeline.BuildTimelinePayload(data, exporttimeline.BuildPayloadOptions{BucketMinutes: 30})
  if err != nil { panic(err) }
  fmt.Printf("phaseMarkers=%d threadSpans=%d firstPhase=%+v\n", len(payload.PhaseMarkers), len(payload.ThreadSpans), func() any { if len(payload.PhaseMarkers)>0 { return payload.PhaseMarkers[0] }; return nil }())
}
EOF
go run /tmp/check_timeline_markers.go
```

### Technical details

**Current task slice completed:**
- merge manual/imported phase spans with SQL-derived phase signals
- merge manual/imported thread spans with SQL-derived thread signals
- add annotation-backed manual/import convention for Proposal 4 markers

## Step 5: Implement the first self-contained browser timeline renderer

With the SQL and payload layers in place, this slice moved Proposal 4 into browser rendering while staying aligned with the read-only, self-contained requirement.

The goal was to ship a practical first timeline HTML artifact end-to-end (CLI -> payload -> HTML/SVG runtime), not just isolated rendering helpers.

### Prompt Context

**User prompt (verbatim):** "Ok, continue with proposal 4 and the browser rendering"

**Assistant interpretation:** Start and complete the first browser rendering slice now that manual/import + SQL merge logic is done.

**Inferred user intent:** Deliver a usable timeline export artifact with real visuals and click behavior rather than more backend-only preparation.

**Commit (code):** `22b60ab3e749515f2f0e0c096d0013a2e11f20da` — "feat: add self-contained proposal 4 timeline renderer"

### What I did

- Added a new timeline export envelope and builder layer:
  - `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/exporttimeline/export.go`
  - `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/exporttimeline/builder.go`
- Added self-contained renderer plumbing:
  - `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/exporttimeline/assets.go`
  - `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/exporttimeline/render_payload.go`
  - `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/exporttimeline/render.go`
  - `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/exporttimeline/templates/timeline.html.tmpl`
  - `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/exporttimeline/templates/timeline.css`
  - `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/exporttimeline/templates/timeline.js`
- Added renderer-focused tests:
  - `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/exporttimeline/builder_test.go`
  - `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/exporttimeline/render_test.go`
- Added new CLI export command:
  - `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/cmd/go-minitrace/cmds/export/timeline.go`
  - wired into `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/cmd/go-minitrace/cmds/export/root.go`
- Implemented first runtime visual sections in the self-contained timeline page:
  - operation heatmap from bucket operation counts,
  - file activity bands,
  - idle window shading,
  - phase ribbon,
  - segmented thread bars,
  - legends and hover tooltips,
  - click-to-reader navigation via `#turn-*` hashes, optionally anchored to `--reader-base-url`.

### Why

Proposal 4’s value only materializes when the normalized payload can be viewed and navigated directly in a browser. Without this slice, the ticket still had analysis and payload layers but no timeline artifact for actual users.

### What worked

- `go-minitrace export timeline` now generates a standalone HTML timeline file.
- The rendered timeline contains all required first-pass visual primitives listed in the task list.
- The output is self-contained (inline CSS/JS/payload; no external `<script src>`/`<link href>` tags).
- Script-tag-safe payload embedding is now in place for the timeline renderer too.
- Unit tests for export builder and HTML renderer pass.
- Full repository tests pass after the slice.
- Real export generated successfully for:
  - `019d03aa-ddee-7403-83d1-2ff075e82d50`
  - `bbf1bdf1-364a-44cb-8cd0-ebcba86dd1ad`

### What didn't work

The Playwright MCP browser session was unavailable in this slice because the profile lock was already held by another active browser session, so I could not run in-harness Playwright click/network assertions immediately.

As a temporary structural substitute, I validated the generated HTML files directly for:

- embedded payload presence,
- no external script/link tags,
- expected timeline payload sections.

### What I learned

A plain SVG runtime is enough for a readable v1 timeline as long as the payload is already normalized and deterministic. The browser code can stay simple if the SQL + Go layers do the heavy lifting first.

### What was tricky to build

The main tricky part was balancing per-bucket detail with long-session readability. The renderer uses adaptive bucket widths and horizontal scrolling to keep the artifact usable on both shorter and longer sessions without introducing non-deterministic layout dependencies.

### What warrants a second pair of eyes

- visual scaling choices (bucket width, row caps) for very long sessions,
- click semantics when `--reader-base-url` is not provided,
- whether phase/thread labels should be made denser or remain conservative for legibility.

### What should be done in the future

- Add explicit browser-level interaction tests once Playwright is available again.
- Add golden tests for SQL command outputs on fixture sessions.
- Add a dedicated workflow note/playbook for preparing manual/import phase/thread annotations.
- Consider moving timeline SQL commands into embedded/core command catalogs for easier out-of-the-box CLI usage.

### Code review instructions

Start here:
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/cmd/go-minitrace/cmds/export/timeline.go`
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/exporttimeline/export.go`
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/exporttimeline/builder.go`
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/exporttimeline/render.go`
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/exporttimeline/templates/timeline.js`
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/exporttimeline/templates/timeline.css`

Then run:

```bash
cd /home/manuel/code/wesen/corporate-headquarters/go-minitrace
go test ./pkg/exporttimeline ./cmd/go-minitrace/cmds/export ./cmd/go-minitrace -count=1
go test ./... -count=1
```

Ad hoc timeline export checks used in this slice:

```bash
# session from local output/active
cd /home/manuel/code/wesen/corporate-headquarters/go-minitrace
go run ./cmd/go-minitrace export timeline \
  --query-repository ./ttmp/2026/04/15/GMT-007--implement-proposal-4-timeline-visualization-for-transcript-exports/query-commands \
  --archive-glob './output/active/*/*.minitrace.json' \
  --session-id 019d03aa-ddee-7403-83d1-2ff075e82d50 \
  --reader-base-url /tmp/gstreamer-reader-react.html \
  --output /tmp/gstreamer-timeline.html

# target GST session
cd /home/manuel/code/wesen/corporate-headquarters/go-minitrace
go run ./cmd/go-minitrace export timeline \
  --query-repository ./ttmp/2026/04/15/GMT-007--implement-proposal-4-timeline-visualization-for-transcript-exports/query-commands \
  --archive-glob './ttmp/2026/04/14/GST-2026-04-13--gstreamer-pi-sessions-analysis-with-go-minitrace/output/active/*/*.minitrace.json' \
  --session-id bbf1bdf1-364a-44cb-8cd0-ebcba86dd1ad \
  --reader-base-url /tmp/gstreamer-reader-react.html \
  --output /tmp/gstreamer-timeline-bbf1.html
```

### Technical details

**Current task slice completed:**
- add self-contained timeline HTML template/runtime
- render heatmap from bucket arrays
- render file activity bands
- render idle window shading
- render phase ribbon
- render thread bars with segmented support
- add hover labels / legends
- add click-to-reader navigation
- add backend rendering tests and real-session timeline export validation

## Step 6: Run Playwright browser validation for the timeline renderer

With Playwright available again, I validated the actual timeline artifact in-browser and verified navigation back into the Proposal 2 reader.

### Prompt Context

**User prompt (verbatim):** "ok, playwright is free"

**Assistant interpretation:** Execute the pending browser validation checks immediately against the Proposal 4 export.

**Inferred user intent:** Move from structural/offline checks to real browser evidence for rendering, requests, console health, and click navigation behavior.

**Commit (code):** pending — docs-only follow-up after validation capture

### What I did

- Regenerated timeline export with a browser-servable reader link:
  - `--reader-base-url /gstreamer-reader-react.html`
  - output: `/tmp/gstreamer-timeline-bbf1-playwright.html`
- Served `/tmp` over local HTTP (`python3 -m http.server 8778 --directory /tmp`).
- Opened timeline export in Playwright and validated:
  - section rendering for heatmap, file bands, idle windows, phase ribbon, and thread bars,
  - zero console errors on clean load,
  - host-scoped network requests only to local served pages,
  - click-to-reader navigation from timeline cells into reader `#turn-*` anchors.
- Captured concrete click outcomes:
  - first timeline cell -> `.../gstreamer-reader-react.html#turn-0`
  - phase-ribbon cell -> `.../gstreamer-reader-react.html#turn-163`
- Verified the destination anchor exists in reader DOM (`document.getElementById('turn-...')` true).
- Captured screenshots:
  - `timeline-export-view.png`
  - `timeline-reader-after-click.png`

### Why

The ticket still had two validation tasks that are best proven with real browser automation:

- sensible click target landing,
- practical readability/responsiveness on a long real session.

This step closes those with direct evidence.

### What worked

- Timeline page loads and renders correctly in Playwright.
- No console errors in clean runs.
- No external network/data pulls from the timeline artifact itself when filtering to host traffic.
- Click navigation into reader works and lands on valid turn anchors.
- The long `bbf1...` session timeline is readable in the current layout with the existing adaptive-width + horizontal scroll strategy.

### What didn't work

No blocker in this run. The only early noise was a local `favicon.ico` 404 from the test server, which was resolved by adding `/tmp/favicon.ico` during validation.

### What I learned

The current Proposal 4 renderer architecture is workable as a v1:

- SQL and payload layers remain the source of truth,
- browser runtime stays lightweight and deterministic,
- click-back to Proposal 2 reader is reliable.

### What warrants a second pair of eyes

- optional refinement of reader-base-url ergonomics (for less path mismatch risk),
- richer browser automation assertions per visual section,
- whether to add a tiny built-in favicon data URL to avoid incidental server noise.

### What should be done in the future

- add dedicated Playwright regression scripts for timeline interactions,
- add golden SQL fixture outputs,
- keep tuning visual density and labeling for very long sessions.

### Code review instructions

Re-run validation with:

```bash
cd /home/manuel/code/wesen/corporate-headquarters/go-minitrace
go run ./cmd/go-minitrace export timeline \
  --query-repository ./ttmp/2026/04/15/GMT-007--implement-proposal-4-timeline-visualization-for-transcript-exports/query-commands \
  --archive-glob './ttmp/2026/04/14/GST-2026-04-13--gstreamer-pi-sessions-analysis-with-go-minitrace/output/active/*/*.minitrace.json' \
  --session-id bbf1bdf1-364a-44cb-8cd0-ebcba86dd1ad \
  --reader-base-url /gstreamer-reader-react.html \
  --output /tmp/gstreamer-timeline-bbf1-playwright.html
```

Then load `http://127.0.0.1:8778/gstreamer-timeline-bbf1-playwright.html` and validate:

- console errors = 0,
- host requests only for timeline + reader pages,
- timeline cell click -> reader `#turn-*`,
- target anchor exists.

### Technical details

**Current task slice completed:**
- browser validation of Proposal 4 timeline renderer in Playwright
- verify timeline clicks land on sensible reader targets
- verify output remains understandable/responsive for long session
