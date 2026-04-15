---
Title: Read-Only Self-Contained HTML Exports with Analysis Requirements
Ticket: GST-2026-04-13
Status: active
Topics:
    - gstreamer
    - go-minitrace
    - transcript-analysis
    - ui-design
    - html-export
    - readonly-exports
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ttmp/2026/04/14/GST-2026-04-13--gstreamer-pi-sessions-analysis-with-go-minitrace/query-commands/gstreamer-tool-analysis.sql
      Note: Example query that could feed into export-time statistics module
ExternalSources: []
Summary: Revised HTML export designs for read-only self-contained documents with client-side JS, including required transcript analysis and synthetic annotations
LastUpdated: 2026-04-14T16:00:00-04:00
WhatFor: Designing self-contained HTML transcript viewers that work offline without backend dependencies
WhenToUse: When exporting transcripts for archival, sharing, or offline review
---


# Read-Only Self-Contained HTML Exports with Analysis Requirements

## Executive Summary

This document revises the HTML export proposals to focus on **read-only, self-contained HTML documents** where:
- All interactivity is client-side JavaScript
- No server or backend is required after export
- The HTML file contains all data (embedded JSON) and code (embedded JS/CSS)
- Users can view, search, filter, and navigate but cannot modify the underlying transcript
- All "annotations" are either: (a) original annotations from the transcript, or (b) **synthetic annotations** computed during the export process

The key innovation is the **export-time analysis pipeline** that generates derived data structures, visualizations, and synthetic annotations from the raw minitrace data, embedding everything needed for rich interaction in a single HTML file.

---

## Core Design Principles

### 1. Self-Contained Single File
```
transcript-export-2026-04-13-bbf1bdf1.html
├── Embedded CSS (styles)
├── Embedded JS (interaction logic)
├── Embedded JSON (transcript data + derived analysis)
└── No external dependencies (works offline, file:// protocol)
```

### 2. Read-Only Guarantee
- All state changes are view-state only (filters, collapses, selections)
- No mutations to transcript data
- No persistence mechanism
- Each reload returns to initial state

### 3. Client-Side Interactivity
- Search: JS filtering of embedded data
- Navigation: JS DOM manipulation (show/hide turns)
- Visualizations: JS-generated (canvas, SVG, or DOM)
- Charts: Computed from embedded derived statistics

### 4. Export-Time Analysis
The export process (Go/minitrace) performs analysis and embeds results:
- Raw transcript data (sessions, turns, tool calls)
- Derived statistics (aggregations, timelines)
- Synthetic annotations (auto-detected patterns)
- Index structures (search, navigation)

---

## Data Architecture for Self-Contained Exports

### Embedded Data Structure

```javascript
// Embedded in HTML as <script type="application/json" id="transcript-data">
{
  "metadata": {
    "export_version": "1.0",
    "generated_at": "2026-04-14T16:00:00Z",
    "generator": "go-minitrace-html-export"
  },
  
  // Raw minitrace data
  "session": { /* ... full session object ... */ },
  "turns": [ /* ... array of turn objects ... */ ],
  "tool_calls": [ /* ... flattened or nested ... */ ],
  "annotations": [ /* ... original annotations ... */ ],
  
  // Derived analysis (computed at export time)
  "derived": {
    "statistics": {
      "total_turns": 1170,
      "total_tools": 1385,
      "operation_breakdown": { "EXECUTE": 529, "READ": 402, "MODIFY": 189 },
      "file_activity": [ { "path": "recording.go", "reads": 19, "modifies": 27 } ],
      "timeline_buckets": [ /* 30-minute activity buckets */ ],
      "duration_metrics": { "ttfa": 45.2, "avg_turn_duration": 73.8 }
    },
    
    "indices": {
      "file_to_turns": { "recording.go": [47, 234, 567, 892] },
      "tool_to_turns": { "bash": [1, 3, 5, 8] },
      "search_index": { /* inverted index for client-side search */ }
    },
    
    "synthetic_annotations": [
      {
        "id": "synth-001",
        "type": "auto-detected-pattern",
        "scope": { "type": "turn_range", "start": 234, "end": 567 },
        "category": "high-velocity-editing",
        "title": "Intensive editing of recording.go",
        "detail": "27 modifications over 45 minutes suggests design iteration",
        "derived_from": [47, 234, 289, 345, 412, 567],
        "confidence": 0.92,
        "signals": ["edit_frequency", "file_repetition", "time_density"]
      }
    ],
    
    "visualization_data": {
      "activity_heatmap": [ /* pre-computed 2D array for heatmap */ ],
      "tool_distribution": [ /* chart-ready data */ ],
      "file_timeline": [ /* chronological file operations */ ]
    }
  },
  
  // Configuration for the viewer
  "viewer_config": {
    "default_view": "chronological",
    "available_views": ["dashboard", "reader", "investigation", "timeline"],
    "initial_filters": {},
    "color_scheme": "light"
  }
}
```

---

## Required Export-Time Analysis

### Analysis Pipeline (Go/minitrace → HTML Export)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                     Export-Time Analysis Pipeline                           │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐                  │
│  │ Raw Minitrace│───→│ Analysis     │───→│ HTML         │                  │
│  │ JSON         │    │ Engine       │    │ Generator    │                  │
│  └──────────────┘    └──────────────┘    └──────────────┘                  │
│         │                   │                   │                          │
│         │            ┌──────┴──────┐         │                          │
│         │            │ Analysis      │         │                          │
│         │            │ Modules       │         │                          │
│         │            ├─────────────┤         │                          │
│         │            │ • Statistics│         │                          │
│         │            │ • Patterns  │         │                          │
│         │            │ • Clusters  │         │                          │
│         │            │ • Threads   │         │                          │
│         │            │ • Errors    │         │                          │
│         │            │ • Search Idx│         │                          │
│         │            └─────────────┘         │                          │
│         │                                    │                          │
│         └────────────────────────────────────┘                          │
│                    All results embedded in HTML                          │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Analysis Modules Required

#### 1. Statistics Module
**Computes:**
- Aggregate counts (turns, tools, by type)
- Duration metrics (total, TTFA, active/idle ratio)
- Token usage (input/output)
- Tool operation distribution
- File operation tallies

**Synthetic annotations generated:**
```json
{
  "type": "statistical-marker",
  "category": "high-activity-session",
  "condition": "tool_count > 1000",
  "message": "High-activity session (1,385 tools)"
}
```

#### 2. Pattern Detection Module
**Computes:**
- Repeated file operations ("hot files")
- Repeated errors ("stuck patterns")
- Tool sequences (e.g., read → edit → test cycle)
- Time-density clusters (rapid activity periods)
- Tool dependencies (which tools often follow others)

**Synthetic annotations generated:**
```json
{
  "type": "pattern-detected",
  "category": "stuck-on-error",
  "scope": { "turns": [145, 203, 289, 412] },
  "pattern": "repeated_gst_element_not_found",
  "suggestion": "Consider dependency check"
}
```

#### 3. Thread Reconstruction Module
**Computes:**
- Identifies semantically related turns that may be non-consecutive
- Groups by: file touched, topic keywords, time proximity, tool patterns
- Creates "narrative threads" through the session

**Synthetic annotations generated:**
```json
{
  "type": "thread-marker",
  "category": "implementation-thread",
  "title": "The GStreamer Bridge Implementation",
  "turns": [234, 345, 412, 567, 734, 892],
  "coherence_score": 0.87,
  "keywords": ["appsink", "appsrc", "bridge", "pipeline"]
}
```

#### 4. Error/Anomaly Detection Module
**Computes:**
- Failed tool calls with patterns
- Repeated attempts at same operation
- Long-running operations (outliers)
- Token usage spikes
- Error recovery patterns (success after failure)

**Synthetic annotations generated:**
```json
{
  "type": "anomaly",
  "category": "repeated-failure",
  "severity": "warning",
  "scope": { "tool_calls": [tc-452, tc-453, tc-454] },
  "pattern": "three_consecutive_bash_failures",
  "resolution_turn": 456
}
```

#### 5. Search Index Module
**Computes:**
- Inverted index for full-text search
- Tokenization of turn content, tool inputs/outputs
- File path index
- Code snippet extraction

**Data structure embedded:**
```json
{
  "search_index": {
    "terms": {
      "gstreamer": [12, 45, 67, 234, 567],
      "appsink": [234, 236, 240, 567],
      "error": [145, 203, 289, 412]
    },
    "code_snippets": [
      { "turn": 234, "language": "go", "content": "..." }
    ]
  }
}
```

---

## Revised View Proposals (Read-Only, Self-Contained)

### Proposal 1: The Static Dashboard (with JS Drill-Down)

**Paradigm:** Static summary with client-side filtering
**Target:** Team Leads, Researchers
**Analysis Required:** Statistics, Pattern Detection

#### ASCII Layout

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  🔬 Transcript Dashboard: bbf1bdf1...                     [⚙️ View Options] │
│  GStreamer Pipeline Migration (SCS-0014)                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐         │
│  │ 🕐 24h 13m   │ │ 🔄 1,170     │ │ 🛠️ 1,385    │ │ 📊 36%       │         │
│  │ Duration     │ │ Turns        │ │ Tool Calls   │ │ Read Ratio   │         │
│  └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘         │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ Filter: [All Operations ▼] [All Tools ▼] [Time Range: All ▼]        │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  ┌─────────────────────────────────────┐ ┌───────────────────────────────┐   │
│  │ Tool Operations (click to filter)   │ │ Activity Timeline (30m buckets)│   │
│  │                                     │ │ ████▓▓░░▒▒██▓▓░░▒▒██▓▓░░▒▒    │   │
│  │  ████████ EXECUTE (529)            │ │ ████▓▓▓▓░░░░▒▒▒▒████▓▓▓▓    │   │
│  │  ██████░░ READ (402)               │ │ ░░░░▒▒▒▒▓▓▓▓████░░░░▒▒▒▒    │   │
│  │  ███░░░░░ MODIFY (189)              │ │ ░░░░▒▒▒▒▓▓▓▓████░░░░▒▒▒▒    │   │
│  │  ██░░░░░░ NEW (82)                  │ │ ░░░░▒▒▒▒▓▓▓▓████░░░░▒▒▒▒    │   │
│  │                                     │ │ ▓ = bash  ░ = read  ▒ = edit  │   │
│  │ [Download CSV] [View Raw JSON]     │ │ ████▓▓▓▓░░░░▒▒▒▒████▓▓▓▓    │   │
│  └─────────────────────────────────────┘ └───────────────────────────────┘   │
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐   │
│  │ Top Modified Files (click to see turns)                               │   │
│  │ ───────────────────────────────────────────────────────────────────  │   │
│  │ ████████████████████ recording.go                    27 edits  [📄]   │   │
│  │ ███████████████░░░░░ preview.go                     18 edits  [📄]   │   │
│  │ ████████████░░░░░░░░░ shared_video.go                15 edits  [📄]   │   │
│  │                                                                       │   │
│  │ [📄] = click to open file detail view (shows all operations on file)  │   │
│  └───────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐   │
│  │ Detected Patterns (auto-generated during export)                      │   │
│  │                                                                       │   │
│  │ ⚠️ High-Velocity Editing    recording.go edited 27× in 45m      [👁️]  │   │
│  │ 🔁 Repeated Error Pattern   "gst element not found" × 12       [👁️]  │   │
│  │ 📈 Long Session              24h duration (95th percentile)   [ℹ️]  │   │
│  │ 🔗 Implementation Thread    Bridge pattern (turns 234-892)   [👁️]  │   │
│  │                                                                       │   │
│  │ [👁️] = click to jump to relevant turns in reader view               │   │
│  └───────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

#### Interactivity (Client-Side JS)

| Interaction | JS Behavior | Data Source |
|-------------|-------------|-------------|
| Click operation bar | Filter all views to show only that operation | `derived.statistics.operation_breakdown` |
| Click file row | Show modal with file's full operation timeline | `derived.indices.file_to_turns` |
| Click pattern | Jump to reader view, pre-filtered to relevant turns | `derived.synthetic_annotations` |
| Click heatmap bucket | Reader view jumps to that time range | `derived.timeline_buckets` |
| Download CSV | JS generates CSV from embedded data | `derived.statistics` |

#### Analysis Required for This View

```yaml
statistics:
  - operation_breakdown by tool and type
  - duration_metrics (total, active, idle)
  - file_activity (reads, modifies per file)
  - timeline_buckets (30-minute granularity)
  
pattern_detection:
  - hot_files: files with >10 operations
  - velocity_clusters: periods with >5 edits/hour
  - repeated_errors: same error message >3 times
  
synthetic_annotations:
  - type: "high-velocity-editing"
    detection: "file.edit_count > 10 AND time_span < 1 hour"
  
  - type: "repeated-error-pattern"
    detection: "same_error_message.count > 3"
    
  - type: "long-session"
    detection: "duration > 8 hours"
    
  - type: "implementation-thread"
    detection: "semantic_clustering(keywords, file_coherence)"
```

---

### Proposal 2: The Chronological Reader (Collapsible Turns)

**Paradigm:** Linear reading with progressive disclosure
**Target:** Developers, Future Self
**Analysis Required:** Search Index, Thread Markers

#### ASCII Layout

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  📖 Session Reader: GStreamer Migration                 [≡ Menu] [🔍 Search] │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ 🔍 Search: [_______________]  [Options ▼]  Found: 47 matches       │   │
│  │ Filters: [All turns] [With edits only ▼] [Threads: All ▼]         │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ ⬆️ Jump to: [Start ▼] [Thread: Bridge Implementation →] [Prev 🔴] │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ Turn #1 · 6:11 PM · 🎙️ User                                         │   │
│  │ ┌─────────────────────────────────────────────────────────────────┐ │   │
│  │ │ Create a new docmgr ticket to port this from ffmpeg to         │ │   │
│  │ │ gstreamer. Keep a detailed diary...                              │ │   │
│  └─┴─────────────────────────────────────────────────────────────────┴─┘   │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ Turn #2 · 6:12 PM · 🤖 Assistant                                    │   │
│  │ ┌─────────────────────────────────────────────────────────────────┐ │   │
│  │ │ I'll help you create the GStreamer migration ticket. First...  │ │   │
│  │ └─────────────────────────────────────────────────────────────────┘ │   │
│  │ ▶ 🛠️ 3 tool calls · 45s  [Click to expand ▼]                      │   │
│  │                                                                     │   │
│  │ 🔗 Part of thread: "GStreamer Migration" (234 turns in thread)     │   │
│  │                                                                     │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ Turn #47 · 8:42 PM · 🤖 Assistant        [📌 Bookmark] [🔗 Copy Link]│   │
│  │ 🏷️ Pattern: high-velocity-editing  🏷️ Thread: Bridge Implementation │   │
│  │ ┌─────────────────────────────────────────────────────────────────┐ │   │
│  │ │ I'll implement the shared video bridge using appsink/appsrc... │ │   │
│  │ └─────────────────────────────────────────────────────────────────┘ │   │
│  │ ▼ 🛠️ 3 tool calls · 45s  [Click to collapse ▲]                      │   │
│  │                                                                     │   │
│  │ ┌─────────────────────────────────────────────────────────────────┐ │   │
│  │ │ 📄 read: pkg/media/gst/shared_video.go (lines 45-120)        │ │   │
│  │ │     · 3s · ✅ Success · [📄 View file]                          │ │   │
│  │ ├─────────────────────────────────────────────────────────────────┤ │   │
│  │ │ ✏️ edit: pkg/media/gst/recording.go (lines 200-245)            │ │   │
│  │ │     · 12s · ✅ Success · [📄 View file] [🔍 View diff]          │ │   │
│  │ │                                                                   │ │   │
│  │ │     @@ -200,7 +200,12 @@                                        │ │   │
│  │ │     -    sink, err := gst.NewElement("appsink")                 │ │   │
│  │ │     +    // Use shared source bridge pattern                    │ │   │
│  │ │     +    sink, err := gst.NewElement("appsink")                 │ │   │
│  │ │     +    if err != nil {                                        │ │   │
│  │ │     +        return fmt.Errorf("failed: %w", err)               │ │   │
│  │ │     +    }                                                      │ │   │
│  │ │                                                                   │ │   │
│  │ │     [Show full diff (45 lines)]                                 │ │   │
│  │ ├─────────────────────────────────────────────────────────────────┤ │   │
│  │ │ ⚡ bash: go test ./pkg/media/gst/...                            │ │   │
│  │ │     · 45s · ✅ Success · Exit 0 · [📄 Full output]               │ │   │
│  │ │     ok  screencast/pkg/media/gst  0.234s                        │ │   │
│  │ │     PASS                                                        │ │   │
│  │ └─────────────────────────────────────────────────────────────────┘ │   │
│  │                                                                     │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ Turn #48 · 8:43 PM · 🎙️ User                                        │   │
│  │ [Filtered: doesn't match current filter criteria]                   │   │
│  │ [Click to show anyway]                                              │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ Turn #892 · 11:42 PM · 🤖 Assistant   [🏁 Thread End: Bridge Impl]  │   │
│  │ 🏷️ Pattern: test-completion                                         │   │
│  │ ┌─────────────────────────────────────────────────────────────────┐ │   │
│  │ │ ✅ All tests passing. The GStreamer bridge is now complete...  │ │   │
│  │ └─────────────────────────────────────────────────────────────────┘ │   │
│  │ ▶ 🛠️ 2 tool calls · 23s                                            │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ ⬇️ End of session. [Export filtered view] [Back to top ⬆️]         │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

#### Interactivity (Client-Side JS)

| Interaction | JS Behavior | Data Source |
|-------------|-------------|-------------|
| Search query | JS searches `derived.search_index.terms` | `derived.search_index` |
| Filter turns | JS toggles visibility based on embedded turn metadata | `turns[] + derived.synthetic_annotations` |
| Expand tool calls | JS shows pre-rendered hidden content | Embedded in HTML as `<template>` or hidden `<div>` |
| View diff | JS modal shows pre-computed diff from `tool_calls` | `tool_calls[].input.file_path + output` |
| Thread navigation | JS filters to turns in thread | `derived.synthetic_annotations[].scope` |
| Copy link | JS generates URL hash: `#turn-47` | Current scroll position |

#### Analysis Required for This View

```yaml
search_index:
  - inverted_index: term -> [turn_ids]
  - code_snippet_extraction: tool_calls with code blocks
  - file_path_index: all file references

thread_detection:
  - semantic_clustering: group turns by topic similarity
  - file_coherence: turns touching same files
  - keyword_tracking: "gstreamer", "appsink", "bridge" mentions
  
synthetic_annotations:
  - type: "thread-boundary"
    detection: "semantic_clustering.start/end"
    
  - type: "pattern-marker"
    detection: "pattern_detection match on this turn"
```

---

### Proposal 3: The Investigation View (Pattern-Focused)

**Paradigm:** Pattern clusters with evidence linking
**Target:** Auditors, Safety Reviewers
**Analysis Required:** Pattern Detection, Error Detection, Thread Reconstruction

#### ASCII Layout

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  🕵️ Investigation View: bbf1bdf1...                                    [📋] │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │ Focus: [All Patterns ▼] [Severity: All ▼] [Auto-detected Only ☑️]   │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│  ┌──────────────────┐ ┌──────────────────┐ ┌──────────────────┐              │
│  │ ⚠️ CLUSTER #1    │ │ 🔁 CLUSTER #2    │ │ ⏱️ CLUSTER #3    │              │
│  │                  │ │                  │ │                  │              │
│  │ Type: Error      │ │ Type: Repetition │ │ Type: Velocity   │              │
│  │ Severity: High   │ │ Severity: Medium │ │ Severity: Low    │              │
│  │                  │ │                  │ │                  │              │
│  │ "gst element not│ │ recording.go     │ │ 27 edits in 45m  │              │
│  │  found" × 12    │ │  edited 27×     │ │ suggests design  │              │
│  │                  │ │                  │ │ iteration        │              │
│  │ Turns: 145, 203,│ │ Signal: unclear│ │                  │              │
│  │  289, 412...    │ │ requirements or│ │ File: recording.go│             │
│  │                  │ │ tooling issue   │ │ Turns: 234-567  │              │
│  │ Pattern: Missing│ │                  │ │                  │              │
│  │  dependency     │ │                  │ │                  │              │
│  │                  │ │                  │ │                  │              │
│  │ [👁️ View 12]    │ │ [👁️ View 27]     │ │ [👁️ View 47]     │              │
│  │ [🔗 Evidence]    │ │ [🔗 Evidence]    │ │ [🔗 Evidence]    │              │
│  │                  │ │                  │ │                  │              │
│  │ Auto-detected    │ │ Auto-detected    │ │ Auto-detected    │              │
│  │ Confidence: 94%  │ │ Confidence: 87%  │ │ Confidence: 91%  │              │
│  └──────────────────┘ └──────────────────┘ └──────────────────┘              │
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │ 🧵 THREAD: "The GStreamer Bridge Implementation" (Auto-reconstructed)│ │
│  │                                                                       │ │
│  │ Coherence: 87% · Turns: 47 · Duration: 3h 27m · Files: 12 modified    │ │
│  │                                                                       │ │
│  │ [Turn 234] Initial attempt with appsink ───────────────────────────→  │ │
│  │      ↓ [Turn 345] Error: caps negotiation failed (Error Cluster #1)  │ │
│  │      ↓ [Turn 412] Pivot to tee + queue pattern                        │ │
│  │      ↓ [Turn 567] Testing with shared_video bridge                   │ │
│  │      ↓ [Turn 734] Integration with preview manager                   │ │
│  │      ↓ [Turn 892] ✅ Tests passing, thread complete                 │ │
│  │                                                                       │ │
│  │ Evidence links:                                                       │ │
│  │ • Turn 234 → 345: Same error pattern (Error Cluster #1)              │ │
│  │ • Turn 567 → 734: Same file (shared_video.go)                        │ │
│  │ • Turn 412 → 892: Keyword bridge in both                             │ │
│  │                                                                       │ │
│  │ [👁️ View Full Thread (47 turns)] [📊 Thread Metrics] [🔗 Evidence Map]│ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │ 📝 Audit Checklist (computed from session data)                        │ │
│  │                                                                       │ │
│  │ ☑️ No credential exposure in tool outputs                             │ │
│  │    (scanned: 529 bash outputs, 0 credential patterns found)         │ │
│  │ ☑️ All bash commands had descriptions                                  │ │
│  │    (96/96 bash tool calls had non-empty descriptions)                 │ │
│  │ ☑️ Error recovery patterns present                                     │ │
│  │    (12 errors found, 11 followed by successful recovery)              │ │
│  │ ☐ High token usage in turns 400-500                                    │ │
│  │    (investigate: 45k tokens in 10-turn window)                        │ │
│  │ ☑️ No PII detected in conversation                                    │ │
│  │    (scanned all turns: no PII patterns)                              │ │
│  │                                                                       │ │
│  │ [📄 Export Audit Report] [🖨️ Print]                                    │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

#### Interactivity (Client-Side JS)

| Interaction | JS Behavior | Data Source |
|-------------|-------------|-------------|
| View cluster | Modal shows all turns in cluster | `derived.synthetic_annotations[].scope` |
| View thread | Reader view filtered to thread turns | `derived.synthetic_annotations[].turns` |
| Evidence links | Shows relationship details | `derived.thread_evidence_links` |
| Export audit | JS generates PDF/text report | Computed from audit checklist rules |

#### Analysis Required for This View

```yaml
error_detection:
  - error_pattern_clustering: group same error messages
  - failure_recovery_tracking: did error resolve? how many tries?
  - severity_scoring: based on error type, frequency, recovery time
  
pattern_clustering:
  - cluster_by_similarity: turns with similar tool patterns
  - file_based_clusters: turns touching same file
  - time_based_clusters: rapid activity bursts
  
thread_reconstruction:
  - semantic_similarity: NLP/keyword based grouping
  - file_coherence_graph: files → turns mapping
  - temporal_proximity: time-based grouping with gaps
  - evidence_links: why are these turns connected?
  
audit_checklist:
  - credential_scanner: regex patterns in tool outputs
  - description_checker: verify bash commands have descriptions
  - error_recovery_analyzer: success after failure detection
  - pii_scanner: PII pattern detection in content
  - token_usage_analyzer: detect unusual token spikes
  
synthetic_annotations:
  - type: "error-cluster"
    detection: "same_error_message.count >= 3"
    scope: "turn_range + specific_tool_calls"
    
  - type: "thread"
    detection: "semantic_clustering.coherence > 0.8"
    evidence: "file_links, keyword_overlap, time_proximity"
    
  - type: "audit-item"
    detection: "rule_based_check(session_data)"
    result: "pass/fail/warning with evidence"
```

---

### Proposal 4: The Timeline Visualization

**Paradigm:** Time-based activity visualization
**Target:** All personas (overview context)
**Analysis Required:** Timeline bucketing, Activity metrics

#### ASCII Layout

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  ⏱️ Timeline View: 24h Session · Apr 13-14, 2026                    [📊📁🔍] │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │ Zoom: [Session ▼] [1h ▼] [30m ▼] [10m ▼]  Navigate: [← 6h] [6h →]   │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │ Activity Heatmap (30-minute buckets)                                  │ │
│  │                                                                       │ │
│  │        18:00  19:00  20:00  21:00  22:00  23:00  00:00  01:00...      │ │
│  │ Tool:   │      │      │      │      │      │      │                   │ │
│  │ bash    ██▓▓░░ ████▓▓ ░░▒▒▒▒ ████▓▓ ░░░░▒▒ ████▓▓ ░░▒▒▒▒ ████▓▓      │ │
│  │ read    ░░▒▒▒▒ ▓▓▓▓░░ ░░▒▒██ ██▓▓░░ ░░▒▒▒▒ ▓▓▓▓░░ ░░▒▒██ ██▓▓░░      │ │
│  │ edit    ░░░░▒▒ ▓▓░░░░ ▒▒▒▒▓▓ ░░░░▒▒ ▓▓░░░░ ▒▒▒▒▓▓ ░░░░▒▒ ▓▓░░░░      │ │
│  │ write   ░░░░░░ ▒▒░░░░ ░░░░▒▒ ▒▒░░░░ ░░░░▒▒ ▒▒░░░░ ░░░░▒▒ ▒▒░░░░      │ │
│  │                                                                       │ │
│  │ Legend: █ high (>20) ▓ medium (10-20) ░ low (5-10) · idle (<5)      │ │
│  │                                                                       │ │
│  │ [Click any cell to jump to turns in that period →]                   │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │ File Activity Timeline                                                │ │
│  │                                                                       │ │
│  │ recording.go    ████████████████████████░░░░░░░░░░░░░░░░░░░░░░░░░░    │ │
│  │   ├─ 18:00-21:00 First implementation phase                          │ │
│  │   ├─ 21:00-02:00 Debugging (Error Cluster #1 active)                 │ │
│  │   └─ 14:00-17:00 Final polish                                        │ │
│  │                                                                       │ │
│  │ shared_video.go ░░░░░░░░░░████████████████████░░░░░░░░░░░░░░░░░░░░    │ │
│  │ preview.go      ░░░░░░░░░░░░░░░░░░░░░░████████████████░░░░░░░░░░░░    │ │
│  │ bridge.go       ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░████████████░░░░    │ │
│  │                                                                       │ │
│  │ [Click file name to filter heatmap to that file only →]              │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │ Thread Timeline (reconstructed)                                       │ │
│  │                                                                       │ │
│  │ Thread: "Bridge Implementation" ────────────────────────────────→     │ │
│  │   [18:30]═══════[21:45]═══════════════[14:30 next day]              │ │
│  │   Phase 1        Debugging phase       Completion                      │ │
│  │                                                                       │ │
│  │ Thread: "Preview Manager Integration" ────────────────────────→       │ │
│  │   [20:15]═══════════════[23:30]                                    │ │
│  │                                                                       │ │
│  │ [👁️ Click thread to see all turns in thread view →]                   │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

#### Analysis Required

```yaml
timeline_bucketing:
  - granularity_levels: [10min, 30min, 1h, 6h, 24h]
  - bucket_computation: count_tools_per_type_per_bucket
  - heatmap_data: 2D array [tool_types × time_buckets]
  
file_activity_timeline:
  - file_operation_timeline: for each file, which buckets had activity
  - phase_detection: detect contiguous activity periods
  
thread_timeline:
  - thread_time_bounds: start/end times per thread
  - thread_phase_detection: sub-periods within thread
```

---

## Implementation: Export Command

### Proposed go-minitrace Command

```bash
# Export single session to self-contained HTML
go-minitrace export html \
  --session-id bbf1bdf1-364a-44cb-8cd0-ebcba86dd1ad \
  --archive-glob './output/active/*/*.minitrace.json' \
  --output ./exports/scs-gstreamer-migration.html \
  --analysis full \
  --views dashboard,reader,investigation,timeline \
  --synthetic-annotations all

# Export with specific analysis modules
go-minitrace export html \
  --session-id bbf1bdf1-... \
  --analysis-modules statistics,patterns,threads \
  --views reader \
  --compact  # minimize file size
```

### Export Process Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    HTML Export Process (go-minitrace)                      │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Input: Session ID + Archive Glob + Analysis Config                          │
│                                                                             │
│  Step 1: LOAD                                                             │
│    • Load minitrace JSON                                                  │
│    • Validate schema version                                              │
│                                                                             │
│  Step 2: ANALYZE (Parallel)                                               │
│    • Run statistics module → derived.statistics                            │
│    • Run pattern detection → derived.synthetic_annotations[]               │
│    • Run thread reconstruction → derived.threads[]                       │
│    • Run search indexing → derived.search_index                            │
│    • Run error detection → derived.anomalies[]                           │
│    • Run audit checklist → derived.audit_items[]                         │
│                                                                             │
│  Step 3: VISUALIZE (Parallel)                                             │
│    • Compute chart data → derived.charts                                   │
│    • Compute heatmap data → derived.heatmaps                               │
│    • Compute timeline data → derived.timelines                             │
│                                                                             │
│  Step 4: RENDER                                                           │
│    • Generate view templates (Go html/template)                          │
│    • Embed CSS (minified)                                                 │
│    • Embed JS (minified)                                                  │
│    • Embed JSON data (compressed)                                         │
│                                                                             │
│  Step 5: OUTPUT                                                           │
│    • Single HTML file                                                     │
│    • Optional: external data file (if >10MB)                            │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## JavaScript Architecture (Client-Side)

### Core Modules (Embedded in HTML)

```javascript
// Embedded as <script> in the HTML file

// Module: DataManager
const DataManager = {
  raw: JSON.parse(document.getElementById('transcript-data').textContent),
  derived: null,
  
  init() {
    this.derived = this.raw.derived;
    return this;
  },
  
  getTurn(id) { return this.raw.turns.find(t => t.id === id); },
  getFileActivity(path) { return this.derived.statistics.file_activity.find(f => f.path === path); },
  getAnnotationsForTurn(turnId) {
    return [
      ...this.raw.annotations.filter(a => a.scope.target_id === turnId),
      ...this.derived.synthetic_annotations.filter(a => a.scope.turns?.includes(turnId))
    ];
  }
};

// Module: SearchEngine
const SearchEngine = {
  index: null,
  
  init() {
    this.index = DataManager.derived.search_index;
  },
  
  search(query) {
    // Client-side search using embedded inverted index
    const terms = query.toLowerCase().split(/\s+/);
    const results = terms.map(term => this.index.terms[term] || []);
    return this.intersect(results);
  }
};

// Module: ViewManager
const ViewManager = {
  currentView: 'reader',
  filters: {},
  
  switchView(viewName) {
    // Hide all views, show selected
    document.querySelectorAll('.view').forEach(v => v.classList.add('hidden'));
    document.getElementById(`view-${viewName}`).classList.remove('hidden');
    this.currentView = viewName;
  },
  
  applyFilter(filterType, value) {
    this.filters[filterType] = value;
    this.refreshCurrentView();
  },
  
  refreshCurrentView() {
    // Re-render current view with applied filters
    const turns = this.getFilteredTurns();
    this.renderTurns(turns);
  }
};

// Module: ChartRenderer
const ChartRenderer = {
  renderHeatmap(containerId, data) {
    // Generate SVG or Canvas heatmap from embedded data
    const svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
    // ... render logic using derived.heatmaps ...
    document.getElementById(containerId).appendChild(svg);
  }
};

// Initialization
document.addEventListener('DOMContentLoaded', () => {
  DataManager.init();
  SearchEngine.init();
  ViewManager.switchView('reader');
});
```

---

## Size & Performance Considerations

### Estimated File Sizes

| Component | Size | Notes |
|-----------|------|-------|
| Raw transcript (1,170 turns) | ~500KB | JSON data |
| Derived analysis | ~200KB | Statistics, indices, annotations |
| Search index | ~150KB | Inverted index for client search |
| Visualization data | ~100KB | Pre-computed chart data |
| CSS (embedded) | ~30KB | Minified styles |
| JS (embedded) | ~50KB | Minified interaction logic |
| **Total (single session)** | **~1MB** | Single HTML file |

### For Large Sessions (10k+ turns)

| Strategy | Approach | Tradeoff |
|----------|----------|----------|
| **Full embed** | Everything in HTML | ~5MB file, slow initial load |
| **Chunked** | Embed first 100 turns, lazy load rest | Faster start, requires file access |
| **External data** | HTML shell + separate data.js | Multiple files, not truly self-contained |
| **Compressed** | gzip embedded JSON | ~30% size reduction |
| **Summarized** | Embed summary only, full in external | Fast start, limited offline use |

### Recommended Approach by Use Case

| Use Case | Session Size | Strategy | Output |
|----------|--------------|----------|--------|
| Email sharing | < 1k turns | Full embed | Single 1-2MB HTML |
| Archival | > 5k turns | Chunked + compression | 2-5MB HTML |
| Dashboard embed | Any | Summary only | 200KB HTML + data URL |
| Offline review | < 2k turns | Full embed | Single file, works on plane |
| Research dataset | > 10k turns | External data | HTML + data/ folder |

---

## Summary: What Analysis is Required for Each View

| View | Primary Analysis Modules | Key Synthetic Annotations | Client-Side Features |
|------|--------------------------|---------------------------|----------------------|
| **Dashboard** | Statistics, Pattern Detection | `high-velocity-editing`, `repeated-error`, `long-session` | Filter, drill-down, CSV export |
| **Reader** | Search Index, Thread Detection | `thread-boundary`, `pattern-marker` | Search, filter, expand/collapse, thread nav |
| **Investigation** | Error Detection, Pattern Clustering, Thread Reconstruction | `error-cluster`, `thread`, `audit-item` | Cluster view, thread view, evidence map |
| **Timeline** | Timeline Bucketing, Activity Metrics | `phase-marker` | Zoom, pan, heatmap interaction |

---

## Open Questions

1. **Search implementation**: Use embedded inverted index (fast, large) or client-side indexing on load (slower, smaller HTML)?

2. **Diff rendering**: Embed pre-computed diffs (larger) or compute client-side with diff library (slower, smaller)?

3. **Thread detection algorithm**: Simple (file-based only) or sophisticated (NLP/semantic)? Tradeoff between accuracy and export time.

4. **Update mechanism**: If annotations are added to original transcript, how to update HTML export? (Re-export required - that's fine for read-only)

5. **Browser support**: Target modern browsers only, or include polyfills for older browsers in embedded JS?

---

## Next Steps

1. Implement core analysis modules in go-minitrace:
   - `pkg/analysis/statistics`
   - `pkg/analysis/patterns`
   - `pkg/analysis/threads`
   - `pkg/analysis/search`

2. Create HTML template system:
   - `templates/export/reader.html`
   - `templates/export/dashboard.html`
   - `templates/export/investigation.html`
   - `templates/export/base.html` (CSS + JS)

3. Implement `go-minitrace export html` command

4. Test with GStreamer session data:
   - Verify file size targets
   - Test client-side performance
   - Validate search accuracy

5. Document usage and customization options
