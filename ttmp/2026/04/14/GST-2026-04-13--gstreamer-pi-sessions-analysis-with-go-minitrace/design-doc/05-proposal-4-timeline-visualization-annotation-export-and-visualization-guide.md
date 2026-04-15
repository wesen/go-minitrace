---
Title: 'Proposal 4 - Timeline Visualization: Annotation, Export, and Visualization Guide'
Ticket: GST-2026-04-13
Status: active
Topics:
    - gstreamer
    - go-minitrace
    - transcript-analysis
    - pipeline-debugging
    - ui-design
    - html-export
    - readonly-exports
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../../corporate-headquarters/go-minitrace/pkg/query/engine.go
      Note: DuckDB archive-loading path referenced by the timeline export guide
    - Path: ../../../../../../../../corporate-headquarters/go-minitrace/pkg/query/presets/timing-analysis.sql
      Note: Existing timing-oriented preset used as a conceptual reference
    - Path: ttmp/2026/04/14/GST-2026-04-13--gstreamer-pi-sessions-analysis-with-go-minitrace/query-commands/session-timeline.sql
      Note: Ticket-local timeline query used as a starting point for export-time bucketing
ExternalSources: []
Summary: 'Intern-facing guide for building the read-only timeline visualization export: phase annotation, export shaping, and client-side rendering'
LastUpdated: 2026-04-14T16:20:00-04:00
WhatFor: Detailed implementation guide for Proposal 4 from the read-only HTML export design
WhenToUse: When an intern or new engineer needs to implement the timeline view export from raw minitrace sessions
---


# Proposal 4 - Timeline Visualization: Annotation, Export, and Visualization Guide

## Executive Summary

This document is a detailed guide for implementing **Proposal 4: The Timeline Visualization** from the read-only self-contained HTML export design.

Where Proposal 2 (the Chronological Reader) helps a human read the transcript like a story, Proposal 4 helps a human answer a different class of questions:

- When were we most active?
- Which file dominated work at which period?
- Where were the long idle gaps?
- When did a thread start, intensify, and resolve?
- When did the work pivot from discovery to implementation to debugging to validation?

The timeline view is not a generic chart page. It is a **time-oriented explanatory layer** on top of the same transcript. It compresses a long session into a readable temporal structure.

This guide explains how to:

1. annotate and phase-label the transcript for timeline use,
2. compute timeline-oriented derived data during export,
3. and render the result as a self-contained read-only HTML visualization.

---

## Problem Statement

A raw transcript tells you what happened in order, but it does not immediately reveal the temporal structure of the session.

For example, a 24-hour GStreamer migration session may actually contain several distinct phases:

- bootstrapping and exploration,
- reading existing code,
- implementation burst,
- repeated failure loop,
- recovery and redesign,
- final validation.

Without a timeline abstraction, a new engineer has to infer those phases manually by reading hundreds of turns.

The timeline visualization solves this by turning timestamps, tool activity, file touches, and annotations into a view that shows:

- intensity over time,
- tool mix over time,
- file activity over time,
- thread spans over time,
- and synthetic phase markers.

The result should still be **read-only** and **self-contained**:

- all derived data is generated at export time,
- all rendering is client-side,
- all interaction is view-state only.

---

## What the Timeline View Is For

A user opens the timeline view when they want macro-structure, not micro-detail.

Typical questions:

- “Show me the big bursts of implementation work.”
- “When was `recording.go` hot?”
- “Did errors cluster before or after the bridge pivot?”
- “How long did the bridge thread last?”
- “Was the preview-manager work parallel with the bridge work or after it?”
- “Where should I jump into the transcript if I only want the debugging phase?”

The timeline view should be able to send the user back into the reader view at any point.

That means Proposal 4 is not standalone. It complements Proposal 2.

---

## Relevant go-minitrace Building Blocks

### Session and timing schema

File references:
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/minitrace/schema.go`
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/doc/minitrace-schema.md`

Relevant timing-related fields:

- `timing.started_at`
- `timing.ended_at`
- `timing.duration_seconds`
- `timing.active_duration_seconds`
- `tool_calls[].timestamp`
- `turns[].timestamp`
- `metrics.time_to_first_action`
- `metrics.idle_ratio`

Timeline visualization should rely primarily on tool-call timestamps, with turn timestamps as secondary data.

### Archive and query engine

File references:
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/query/engine.go`
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/query/presets/timing-analysis.sql`
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/query/presets/tool-operation-breakdown.sql`

These already establish the core idea that timing and activity analysis is queryable from the archive.

### Ticket-local query examples

Ticket file references:
- `ttmp/2026/04/14/GST-2026-04-13--gstreamer-pi-sessions-analysis-with-go-minitrace/query-commands/session-timeline.sql`
- `ttmp/2026/04/14/GST-2026-04-13--gstreamer-pi-sessions-analysis-with-go-minitrace/query-commands/top-modified-files.sql`
- `ttmp/2026/04/14/GST-2026-04-13--gstreamer-pi-sessions-analysis-with-go-minitrace/query-commands/gstreamer-tool-analysis.sql`

These are especially useful for the timeline export because they already capture:

- time-based grouping,
- file-level activity,
- operation-type breakdowns.

---

## System Overview

The timeline system has four conceptual layers.

```text
┌──────────────────────────────────────────────────────────────────────┐
│ Layer 1: Raw timestamped events                                      │
│  - session timing                                                    │
│  - turn timestamps                                                   │
│  - tool call timestamps                                              │
│  - original annotations                                              │
└──────────────────────────────────────────────────────────────────────┘
                              ↓
┌──────────────────────────────────────────────────────────────────────┐
│ Layer 2: Derived temporal structures                                 │
│  - time buckets                                                      │
│  - file activity spans                                               │
│  - phase markers                                                     │
│  - thread spans                                                      │
│  - idle windows                                                      │
│  - error clusters over time                                          │
└──────────────────────────────────────────────────────────────────────┘
                              ↓
┌──────────────────────────────────────────────────────────────────────┐
│ Layer 3: Timeline export payload                                     │
│  - chart-ready arrays                                                │
│  - labels                                                            │
│  - thread metadata                                                   │
│  - drill-down link targets                                           │
└──────────────────────────────────────────────────────────────────────┘
                              ↓
┌──────────────────────────────────────────────────────────────────────┐
│ Layer 4: Read-only interactive HTML timeline                         │
│  - heatmap                                                           │
│  - file timeline                                                     │
│  - phase ribbons                                                     │
│  - thread bars                                                       │
│  - click-to-jump                                                     │
└──────────────────────────────────────────────────────────────────────┘
```

---

## Annotation Strategy for the Timeline View

Timeline annotations are different from reader annotations.

The reader needs local, narrative, turn-level meaning. The timeline needs **temporal landmarks**.

### What timeline annotations should communicate

The timeline view needs annotations that answer:

- what phase is this time window in?
- what started here?
- what ended here?
- what error cluster dominated this period?
- what file became hot here?
- when did the session go idle?
- when did we shift from exploration to implementation?

### Recommended annotation classes

#### 1. Phase markers

These are the most important synthetic annotations for the timeline.

Examples:
- `phase:start-discovery`
- `phase:implementation-burst`
- `phase:debugging-loop`
- `phase:validation`
- `phase:wrap-up`

A phase marker is usually a **range annotation**, conceptually attached to a start and end span even if the base annotation schema is point-oriented. In practice, represent it in derived export data as a time span or turn range.

#### 2. Thread span markers

These identify the start and end of a coherent work thread.

Examples:
- `thread:bridge-implementation`
- `thread:preview-integration`
- `thread:test-recovery`

#### 3. Error cluster markers

These mark time regions dominated by related failures.

Examples:
- repeated `gst element not found`
- caps negotiation failures
- repeated failing test runs

#### 4. Idle and gap markers

These identify periods with little or no meaningful activity.

Examples:
- long pause between work bursts,
- “resume after break,”
- inactive periods due to waiting or context switch.

### Human / LLM annotation rubric for timeline

A human or LLM annotator should look for **windows**, not just individual turns.

#### A. Label phase boundaries

The annotator should inspect the session and identify where a qualitatively different mode of work begins.

Signals for a phase boundary:
- abrupt change in tool mix (e.g. reads become edits),
- abrupt change in file focus,
- repeated failures followed by new design direction,
- first sustained testing period,
- first sustained documentation period.

Example annotation intent:

```json
{
  "type": "phase-marker",
  "title": "Debugging loop begins",
  "detail": "This period begins when the assistant repeatedly edits recording.go and reruns tests after caps negotiation failures.",
  "start_turn": 345,
  "end_turn": 567,
  "tags": ["debugging", "errors", "recording.go"]
}
```

#### B. Label high-intensity windows

Mark windows where activity is unusually dense.

Signals:
- many tool calls in a short time,
- many edits in the same file within an hour,
- high alternation between `edit` and `bash`.

#### C. Label thread spans

If the same topic is revisited across non-consecutive turns, mark the overall span and notable sub-bursts.

Example:
- bridge implementation starts at 18:30,
- pauses during unrelated work,
- resumes at 21:15,
- concludes at 14:30 next day.

Timeline annotations should make these discontinuous spans visible.

#### D. Label idle or waiting periods

If there is a long gap with few events, annotate it only when it matters.

Examples:
- overnight pause,
- waiting on environment setup,
- long inactive interval after a milestone.

### Synthetic annotations the exporter should generate

#### Synthetic type 1: phase-marker

Generated from heuristics over tool mix, file focus, and error clusters.

Example heuristics:
- **discovery** if reads dominate and file diversity is high,
- **implementation** if modify/new operations dominate and one file cluster becomes hot,
- **debugging** if repeated edit→execute cycles occur with failures,
- **validation** if execute-heavy with tests and lower edit density.

#### Synthetic type 2: intensity-window

Generated when activity density exceeds a threshold.

Heuristic example:
- > 20 tool calls in 30 minutes,
- or > 5 edits in a single file in 20 minutes.

#### Synthetic type 3: idle-window

Generated when gap between adjacent timestamped events exceeds threshold.

Heuristic example:
- gap > 20 minutes for “idle,”
- gap > 2 hours for “major pause.”

#### Synthetic type 4: file-hotspan

Generated when a file dominates a time window.

Heuristic example:
- same file appears in > 40% of file-related tool calls within a 1-hour window.

### LLM annotation prompt skeleton for timeline

```text
You are annotating a minitrace session for a timeline visualization.
Focus on time windows and phase changes, not on every individual turn.

Identify:
- major phases of work,
- high-intensity windows,
- repeated-error windows,
- thread spans,
- idle or pause windows,
- validation and completion periods.

For each annotation, provide:
- title
- short detail
- start turn or start timestamp
- end turn or end timestamp
- tags
- confidence (low/medium/high)

Prefer fewer high-signal timeline annotations over many small ones.
```

---

## Export Data Model for Timeline Visualization

The timeline HTML should not compute buckets from scratch in the browser. That work belongs to export time.

### Recommended export structure

```json
{
  "session": {
    "id": "bbf1bdf1-364a-44cb-8cd0-ebcba86dd1ad",
    "started_at": "2026-04-13T18:11:37Z",
    "ended_at": "2026-04-14T18:10:33Z",
    "duration_seconds": 86358
  },
  "timeline": {
    "bucket_size_minutes": 30,
    "buckets": [],
    "tool_heatmap": [],
    "file_series": [],
    "thread_spans": [],
    "phase_markers": [],
    "error_clusters": [],
    "idle_windows": []
  }
}
```

### Time bucket format

```json
{
  "bucket_index": 0,
  "start": "2026-04-13T18:00:00Z",
  "end": "2026-04-13T18:30:00Z",
  "tool_counts": {
    "READ": 12,
    "MODIFY": 4,
    "EXECUTE": 9,
    "NEW": 1,
    "OTHER": 0
  },
  "file_counts": {
    "pkg/media/gst/recording.go": 5,
    "pkg/media/gst/shared_video.go": 3
  },
  "turn_indices": [0,1,2,3,4,5],
  "thread_ids": ["thread-bridge"],
  "synthetic_tags": ["implementation-burst"]
}
```

### File activity series format

```json
{
  "file_path": "pkg/media/gst/recording.go",
  "display_name": "recording.go",
  "activity": [0, 1, 2, 8, 12, 7, 3, 0],
  "peak_bucket": 4,
  "total_operations": 46,
  "dominant_operation": "MODIFY"
}
```

### Thread span format

```json
{
  "thread_id": "thread-bridge-implementation",
  "title": "Bridge Implementation",
  "segments": [
    {"start_bucket": 1, "end_bucket": 5},
    {"start_bucket": 9, "end_bucket": 14}
  ],
  "related_turns": [234,345,412,567,734,892],
  "coherence_score": 0.87
}
```

### Phase marker format

```json
{
  "phase_id": "phase-debugging-loop",
  "title": "Debugging loop",
  "start_bucket": 4,
  "end_bucket": 8,
  "confidence": 0.91,
  "signals": ["repeated_failures", "edit_execute_cycle", "recording.go_hot"],
  "jump_turn": 345
}
```

---

## Export-Time Analysis Required

### Analysis module 1: bucketization

This is the foundation of the entire timeline view.

Responsibilities:
- choose time granularity,
- compute contiguous buckets,
- assign turns and tool calls into buckets,
- compute per-bucket counts.

Recommended default:
- 30-minute buckets for large sessions,
- optionally derive 10-minute and 1-hour variants for zoom.

Pseudocode:

```go
func BuildBuckets(start time.Time, end time.Time, size time.Duration) []Bucket {
    var buckets []Bucket
    cur := alignDown(start, size)
    for cur.Before(end) || cur.Equal(end) {
        buckets = append(buckets, Bucket{
            Start: cur,
            End:   cur.Add(size),
        })
        cur = cur.Add(size)
    }
    return buckets
}
```

### Analysis module 2: tool heatmap computation

Responsibilities:
- count operations by type per bucket,
- optionally count by tool name per bucket,
- precompute normalized intensities for rendering.

Pseudocode:

```go
func ComputeToolHeatmap(toolCalls []ToolCall, buckets []Bucket) Heatmap {
    matrix := newMatrix(operationKinds, len(buckets))
    for _, tc := range toolCalls {
        b := bucketIndexFor(tc.Timestamp, buckets)
        matrix[tc.OperationType][b]++
    }
    return NormalizeHeatmap(matrix)
}
```

### Analysis module 3: file activity series

Responsibilities:
- decide which files are important enough to include,
- compute activity count per bucket per file,
- detect hot spans.

Heuristic for inclusion:
- include top N files by total operations,
- or include all files with >= 5 operations.

### Analysis module 4: phase detection

This is the most interpretive part.

Responsibilities:
- classify windows as discovery, implementation, debugging, validation, wrap-up,
- detect boundary conditions,
- emit phase markers.

Suggested heuristics:

- **discovery**
  - READ heavy,
  - many unique files,
  - few modifications.

- **implementation**
  - MODIFY + NEW increase,
  - a small cluster of files dominates,
  - execute operations present but not dominant.

- **debugging**
  - repeated EXECUTE and MODIFY cycle,
  - failures cluster,
  - same files repeat.

- **validation**
  - EXECUTE heavy,
  - fewer edits,
  - successful tests and summaries.

### Analysis module 5: thread timeline reconstruction

Responsibilities:
- convert thread detections into time spans,
- support non-contiguous segments,
- create jump targets back to relevant turns.

This depends on either:
- manual/LLM thread annotations,
- or automatic thread detection described in Proposal 2.

### Analysis module 6: idle window detection

Responsibilities:
- compute time gaps between adjacent events,
- label meaningful pauses,
- avoid clutter from tiny gaps.

Pseudocode:

```go
func DetectIdleWindows(events []time.Time, minGap time.Duration) []IdleWindow {
    var ret []IdleWindow
    for i := 1; i < len(events); i++ {
        gap := events[i].Sub(events[i-1])
        if gap >= minGap {
            ret = append(ret, IdleWindow{
                Start: events[i-1],
                End:   events[i],
                Gap:   gap,
            })
        }
    }
    return ret
}
```

---

## Visualization Design

The timeline visualization should prioritize readability over chart cleverness.

### Recommended visual components

#### 1. Tool activity heatmap

Purpose:
- show what kinds of work happened when.

Rows:
- operation types or tool families.

Columns:
- time buckets.

Cell intensity:
- normalized count.

#### 2. File activity bands

Purpose:
- show when important files were hot.

Each file gets a horizontal band where bucket intensity reflects operation volume.

#### 3. Phase ribbon

Purpose:
- show high-level phases across time.

This can sit above the heatmap as colored spans.

#### 4. Thread bars

Purpose:
- show reconstructed work threads.

Threads may be discontinuous. That means the same thread can have multiple visible segments.

#### 5. Jump markers

Purpose:
- let the user click a bucket, phase, or thread segment and jump into the reader view.

### ASCII layout

```text
┌─────────────────────────────────────────────────────────────────────┐
│ Header: session | zoom | bucket size | jump to reader              │
├─────────────────────────────────────────────────────────────────────┤
│ Phase ribbon: [discovery][implementation][debugging][validation]   │
├─────────────────────────────────────────────────────────────────────┤
│ Heatmap                                                           │
│ bash     ███░░░▓▓▓████░░░░▒▒▓▓                                     │
│ read     ░░███▓▓▓░░░▒▒▒▒██░░░                                     │
│ edit     ░░░░▒▒██████▓▓▓░░░░                                     │
│ write    ░░░░░░▒▒▓▓░░░░░░░░                                     │
├─────────────────────────────────────────────────────────────────────┤
│ Files                                                             │
│ recording.go    ░░░░██████████▓▓▓░░░░                            │
│ shared_video.go ░░▓▓▓▓████░░░░░░░░░░                            │
│ preview.go      ░░░░░░░░▓▓██████░░░░                            │
├─────────────────────────────────────────────────────────────────────┤
│ Threads                                                           │
│ bridge impl     [====]      [==========]                           │
│ preview thread      [=====]                                        │
└─────────────────────────────────────────────────────────────────────┘
```

### Key design decisions

#### Decision 1: precompute all chart data at export time

Why:
- keeps browser logic simple,
- avoids expensive client-side recomputation,
- makes the HTML deterministic and easy to test.

#### Decision 2: use SVG first, Canvas only if needed

Why:
- SVG is easier to debug,
- easier to wire click handlers,
- easier for interns to reason about,
- good enough for session sizes in the low thousands.

#### Decision 3: timeline clicks should always map back to turns

Why:
- the timeline is useful only if it can bring the user back to the transcript details.

That means every bucket should know which turn indices it covers.

---

## Browser Architecture

Recommended client-side modules:

- `TimelineDataManager`
- `HeatmapRenderer`
- `FileBandRenderer`
- `PhaseRenderer`
- `ThreadRenderer`
- `TimelineInteractionController`

### Pseudocode: rendering heatmap

```javascript
function renderHeatmap(container, heatmapData) {
  const svg = createSvg(container.clientWidth, 220);
  const rowHeight = 28;
  const colWidth = Math.max(6, Math.floor(container.clientWidth / heatmapData.bucketCount));

  heatmapData.rows.forEach((row, rowIndex) => {
    row.values.forEach((value, colIndex) => {
      const cell = makeRect({
        x: colIndex * colWidth,
        y: rowIndex * rowHeight,
        width: colWidth - 1,
        height: rowHeight - 2,
        fill: intensityToColor(value)
      });
      cell.dataset.bucketIndex = String(colIndex);
      cell.dataset.rowName = row.name;
      cell.addEventListener('click', onHeatmapCellClick);
      svg.appendChild(cell);
    });
  });

  container.replaceChildren(svg);
}
```

### Pseudocode: click-to-jump

```javascript
function onHeatmapCellClick(event) {
  const bucketIndex = Number(event.target.dataset.bucketIndex);
  const bucket = TimelineDataManager.getBucket(bucketIndex);
  const firstTurn = bucket.turn_indices[0];
  if (firstTurn !== undefined) {
    window.location.hash = `turn-${firstTurn}`;
    showJumpMessage(`Jump target: turn ${firstTurn}`);
  }
}
```

### Pseudocode: zoom switching

```javascript
function setZoom(level) {
  const payload = TimelineDataManager.getPayloadForZoom(level);
  renderHeatmap(document.getElementById('heatmap'), payload.heatmap);
  renderFileBands(document.getElementById('files'), payload.files);
  renderThreads(document.getElementById('threads'), payload.threads);
}
```

---

## Implementation Plan

### Phase 1: data model and bucketing

Deliverables:
- Go structs for timeline export,
- bucket builder,
- tool heatmap computation,
- file series computation.

### Phase 2: phase and idle analysis

Deliverables:
- phase detection heuristics,
- idle window detection,
- error cluster time mapping.

### Phase 3: thread time mapping

Deliverables:
- thread spans,
- non-contiguous segment support,
- jump target selection for each thread.

### Phase 4: HTML template and SVG rendering

Deliverables:
- base timeline page,
- phase ribbon,
- heatmap,
- file bands,
- thread bars.

### Phase 5: polish and integration with reader

Deliverables:
- click-to-reader behavior,
- hover tooltips,
- legends,
- accessibility labels,
- printing behavior.

---

## Suggested File Layout

```text
pkg/analysis/timeline/
  buckets.go
  heatmap.go
  files.go
  phases.go
  idle.go
  threads.go

pkg/export/html/
  timeline_export.go
  timeline_view_model.go

templates/export/
  timeline.html.tmpl
  partials/
    heatmap.svg.tmpl
    file_bands.svg.tmpl
    thread_bars.svg.tmpl

web-static/export/
  timeline.css
  timeline.js
```

---

## API and File References

### Existing go-minitrace files to understand first

- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/query/engine.go`
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/minitrace/archive.go`
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/minitrace/schema.go`
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/doc/minitrace-schema.md`
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/query/presets/timing-analysis.sql`
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/query/presets/tool-operation-breakdown.sql`
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/cmd/go-minitrace/main.go`

### Ticket-local references

- `ttmp/2026/04/14/GST-2026-04-13--gstreamer-pi-sessions-analysis-with-go-minitrace/query-commands/session-timeline.sql`
- `ttmp/2026/04/14/GST-2026-04-13--gstreamer-pi-sessions-analysis-with-go-minitrace/query-commands/top-modified-files.sql`
- `ttmp/2026/04/14/GST-2026-04-13--gstreamer-pi-sessions-analysis-with-go-minitrace/query-commands/gstreamer-tool-analysis.sql`
- `ttmp/2026/04/14/GST-2026-04-13--gstreamer-pi-sessions-analysis-with-go-minitrace/design-doc/03-html-readonly-exports-with-analysis.md`

---

## Alternatives Considered

### Alternative A: only show a single activity sparkline

Rejected because it loses too much structure.

### Alternative B: compute all buckets in the browser

Rejected for first implementation because it complicates the client unnecessarily.

### Alternative C: only show phases, no low-level activity

Rejected because users still need drill-down to concrete buckets and files.

---

## Review Checklist for the Intern

A reviewer should be able to confirm:

- the bucket counts match raw tool calls,
- file activity bands reflect real file-touch data,
- phase markers are explainable,
- thread spans link back to meaningful transcript ranges,
- clicking any timeline element results in a sensible reader jump,
- and the visualization still works offline.

---

## Implementation Checklist

- [ ] Define timeline export structs
- [ ] Implement bucketization
- [ ] Compute tool heatmap arrays
- [ ] Compute file activity series
- [ ] Detect idle windows
- [ ] Detect or import thread spans
- [ ] Detect or annotate phase ranges
- [ ] Export chart-ready payload
- [ ] Build SVG-based timeline renderer
- [ ] Add click-to-reader jump behavior
- [ ] Test on session `bbf1bdf1-364a-44cb-8cd0-ebcba86dd1ad`
- [ ] Validate browser performance and offline use

---

## Final Recommendation

Build the timeline in this order:

1. buckets,
2. heatmap,
3. file bands,
4. phase ribbon,
5. thread spans,
6. jump behavior.

Do not start with beautiful styling. Start with truthful, testable temporal data. If the counts and spans are correct, styling can always improve later. If the underlying timeline math is wrong, no amount of polish will save the feature.
