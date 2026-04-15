---
Title: HTML Transcript Export Proposals
Ticket: GST-2026-04-13
Status: active
Topics:
    - gstreamer
    - go-minitrace
    - transcript-analysis
    - ui-design
    - html-export
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ttmp/2026/04/14/GST-2026-04-13--gstreamer-pi-sessions-analysis-with-go-minitrace/output/active/2026-04/bbf1bdf1-364a-44cb-8cd0-ebcba86dd1ad.minitrace.json
      Note: Example session data for HTML export prototype
    - Path: ttmp/2026/04/14/GST-2026-04-13--gstreamer-pi-sessions-analysis-with-go-minitrace/query-commands/gstreamer-session-list.sql
      Note: Example query command that could feed HTML dashboard
ExternalSources: []
Summary: Design proposals for rich HTML transcript exports with multiple user personas, affordances, and interactive features
LastUpdated: 2026-04-14T15:00:00-04:00
WhatFor: Designing next-generation transcript browser and export capabilities
WhenToUse: When implementing HTML export features in go-minitrace or related tools
---



# HTML Transcript Export Proposals

## Executive Summary

This document proposes multiple design approaches for rich HTML exports of AI coding session transcripts. The goal is to create navigable, readable, and information-dense HTML documents that serve different user needs—from researchers studying AI behavior to developers reviewing their own coding sessions.

The proposals leverage the existing minitrace data model (sessions, turns, tool calls, annotations, metrics) and the sqleton query command system to generate dynamic, interactive HTML views.

---

## User Personas & Their Needs

### 1. The Researcher (Dr. Alice)
**Goal:** Study AI coding assistant behavior patterns across multiple sessions
**Needs:**
- Statistical aggregations and visualizations
- Annotation taxonomy filtering (MAST, ToolEmu, minitrace classifications)
- Pattern detection across sessions
- Export to research formats (CSV, JSON)
- Query interface for custom analysis

### 2. The Developer (Bob)
**Goal:** Review his own coding session to recall decisions or find lost code
**Needs:**
- Timeline navigation with jump-to-specific-turn
- Code diff views showing what changed
- Search across all turns and tool outputs
- File tree view of touched files
- Bookmarking important moments

### 3. The Auditor (Carol)
**Goal:** Review AI safety and quality of assistant interactions
**Needs:**
- Annotation visibility at turn/tool level
- Success/failure rate metrics
- Tool call categorization (READ/MODIFY/EXECUTE)
- Red-flag indicators (errors, repeated failures)
- Compliance checklists

### 4. The Team Lead (David)
**Goal:** Understand team productivity and AI tool usage
**Needs:**
- High-level session summaries
- Tool usage breakdowns
- Time-to-first-action metrics
- Comparison across team members
- Integration with project management

### 5. The Future Self (Eve)
**Goal:** Understand what she was thinking 6 months ago
**Needs:**
- Context preservation (git state, environment)
- Rich linking (file versions, external docs)
- Thought process reconstruction
- Decision rationale documentation

---

## Data Mapping: Minitrace → UI Affordances

### Core Data Model

```
Session
├── Metadata (id, title, timestamp, model, framework)
├── Environment (cwd, system info)
├── Timing (duration, TTFA, idle time)
├── Metrics (turn count, tool count, token usage)
├── Turns[]
│   ├── Index, Role, Timestamp
│   ├── Content (message text)
│   ├── Tool Calls[]
│   │   ├── Tool Name, Operation Type
│   │   ├── Input (file path, command, query)
│   │   ├── Output (result, success/failure)
│   │   └── Duration, Timestamp
│   └── Annotations[]
├── Annotations[]
│   ├── Scope (session/turn/tool)
│   ├── Category (ai-failure, question, etc.)
│   ├── Content (title, detail, tags)
│   └── Taxonomy Mappings
└── Derived Queries
    ├── File operation timeline
    ├── Tool frequency stats
    ├── Bash command patterns
    └── Web search topics
```

### UI Affordance Mapping

| Data Element | UI Affordance | Purpose |
|--------------|---------------|---------|
| Turn sequence | Vertical timeline | Chronological navigation |
| Tool calls per turn | Collapsible panels | Detail-on-demand |
| Operation types | Color coding | Quick pattern recognition |
| File paths | Clickable links | Navigation to related work |
| Code diffs | Syntax-highlighted blocks | Readability |
| Annotations | Badge chips + tooltips | Metadata without clutter |
| Metrics | Sparklines/stat boxes | At-a-glance understanding |
| Duration data | Progress bars | Time perception |
| Taxonomy codes | Filter sidebar | Categorical exploration |

---

## Proposal 1: The Dashboard View

**Target:** Team Leads, Researchers  
**Metaphor:** Analytics dashboard with drill-down capability

### ASCII Layout

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  🔬 Transcript Explorer: GST-0014 - GStreamer Pipeline Migration     [🔍]  │
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐         │
│  │ 🕐 24h 13m   │ │ 🔄 1,170     │ │ 🛠️ 1,385    │ │ 📊 36%       │         │
│  │ Duration     │ │ Turns        │ │ Tool Calls   │ │ Read Ratio   │         │
│  └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘         │
│                                                                             │
│  ┌─────────────────────────────────────┐ ┌───────────────────────────────┐  │
│  │ Tool Operations Distribution        │ │ Activity Timeline             │  │
│  │                                     │ │ ██▓▓░░▒▒██▓▓░░▒▒██▓▓░░▒▒    │  │
│  │  ████████ EXECUTE (529)            │ │ ████▓▓▓▓░░░░▒▒▒▒████▓▓▓▓    │  │
│  │  ██████░░ READ (402)               │ │ ░░░░▒▒▒▒▓▓▓▓████░░░░▒▒▒▒    │  │
│  │  ███░░░░░ MODIFY (189)              │ │ ████▓▓▓▓░░░░▒▒▒▒████▓▓▓▓    │  │
│  │  ██░░░░░░ NEW (82)                  │ │                               │  │
│  │                                     │ │ ▓ = bash  ░ = read  ▒ = edit  │  │
│  └─────────────────────────────────────┘ └───────────────────────────────┘  │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐  │
│  │ Top Modified Files                              [Filter by session ▼]│  │
│  │ ───────────────────────────────────────────────────────────────────  │  │
│  │ 📄 pkg/media/gst/recording.go                    27 edits  🔍 👁️ 📊 │  │
│  │ 📄 pkg/media/gst/shared_video_recording_bridge.go 11 edits  🔍 👁️ 📊│  │
│  │ 📄 internal/web/preview_manager.go              8 edits  🔍 👁️ 📊 │  │
│  │ 📄 pkg/media/gst/shared_video.go                  5 edits  🔍 👁️ 📊 │  │
│  │                                                                    │  │
│  │ [View full file tree]    [Export CSV]    [Generate Report]          │  │
│  └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐  │
│  │ Annotations Summary                                                 │  │
│  │ 🏷️ ai-behavior: 12  🐛 error-pattern: 8  ❓ question: 5  ✅ good: 23  │  │
│  │ [Filter by category]  [View annotation timeline]                      │  │
│  └─────────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Interactive Features
- **Click any metric card** → Drill down to filtered transcript view
- **Click file row** → Side-by-side diff view of all edits
- **Hover on timeline** → See what happened at that moment
- **Filter dropdowns** → Narrow by date range, operation type, tool

---

## Proposal 2: The Chronological Reader

**Target:** Developers, Future Self  
**Metaphor:** Rich chat/document reader with deep linking

### ASCII Layout - Overview Mode

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  📖 Session: GStreamer Pipeline Migration (SCS-0014)           [≡] [🔍] [⚙️]│
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌────────────────┐                                                        │
│  │ 📑 Navigation  │    ┌────────────────────────────────────────────────┐   │
│  ├────────────────┤    │ 👤 User · Apr 13, 6:11 PM                       │   │
│  │ 🔍 Search...   │    │─────────────────────────────────────────────────│   │
│  ├────────────────┤    │ Create a new docmgr ticket to port this from    │   │
│  │ Jump to turn:  │    │ ffmpeg streams to gstreamer. Keep a detailed    │   │
│  │ [#________]    │    │ diary and create the analysis scripts...        │   │
│  ├────────────────┤    │                                                 │   │
│  │ 📁 File Tree   │    │ 🤖 Assistant · 6:12 PM                           │   │
│  │ ▼ screencast/  │    │─────────────────────────────────────────────────│   │
│  │   ▶ 📄 main.go │    │ I'll help you create the GStreamer migration    │   │
│  │   ▶ 📄 gst/    │    │ ticket. First, let me explore the existing       │   │
│  │     📄 ◉ rec.  │    │ FFmpeg implementation...                        │   │
│  │     📄 ◉ prev. │    │                                                 │   │
│  │     📄 ◉ shar. │    │ [📝 3 tool calls]  [⏱️ 45s]  [📊 Context: 12k]   │   │
│  │   ▶ 📄 web/    │    │                                                 │   │
│  ├────────────────┤    │ ─────────────────────────────────────────────── │   │
│  │ 🏷️ Annotations │    │ 👤 User · 6:13 PM                               │   │
│  │ ☑️ ai-failure  │    │─────────────────────────────────────────────────│   │
│  │ ☑️ error       │    │ Check the current pipeline implementation...    │   │
│  │ ☑️ question    │    │                                                 │   │
│  │ ☐ good-pattern │    │ 🤖 Assistant · 6:14 PM [🔴 ERROR RECOVERY]      │   │
│  │ ☐ refactor     │    │─────────────────────────────────────────────────│   │
│  │                │    │ I see the issue. The FFmpeg pipeline is failing │   │
│  │ [+ Add Tag]    │    │ with the new multi-source setup. Let me fix...  │   │
│  └────────────────┘    │                                                 │   │
│                        │ [🔧 edit: recording.go L45-67] [⏱️ 12s] [✅]      │   │
│                        │ [📖 read: shared_video.go L120-145] [⏱️ 3s] [✅]│   │
│                        └────────────────────────────────────────────────┘   │
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │ ⚡ Quick Stats:  Turn 47 of 1170  ·  2h 14m elapsed  ·  234 tools used │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────┘
```

### ASCII Layout - Turn Detail Expanded

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  🤖 Assistant · Turn #47 · Apr 13, 8:42 PM · gpt-5.4              [📌] [🔗] │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  I'll implement the shared video bridge using the appsink/appsrc pattern.    │
│  First, let me check the current GStreamer pipeline structure:              │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ TOOL CALLS (3)                                                      │   │
│  ├─────────────────────────────────────────────────────────────────────┤   │
│  │ 1. read: pkg/media/gst/shared_video.go                             │   │
│  │    Lines 45-120 · 3s · ✅ Success                                   │   │
│  │    ─────────────────────────────────────────────────────────────   │   │
│  │    [View file content...]                                           │   │
│  │                                                                     │   │
│  │ 2. edit: pkg/media/gst/recording.go                                │   │
│  │    Lines 200-245 · 12s · ✅ Success · [Diff View]                  │   │
│  │    ─────────────────────────────────────────────────────────────   │
│  │    @@ -200,7 +200,12 @@ func (r *Recorder) Start() error {        │   │
│  │    -    sink, err := gst.NewElement("appsink")                     │   │
│  │    +    // Use shared source bridge pattern                         │   │
│  │    +    sink, err := gst.NewElement("appsink")                     │   │
│  │    +    if err != nil {                                             │   │
│  │    +        return fmt.Errorf("failed to create appsink: %w", err)  │   │
│  │    +    }                                                           │   │
│  │                                                                     │   │
│  │    [Show full diff (45 lines)]                                      │   │
│  │                                                                     │   │
│  │ 3. bash: go test ./pkg/media/gst/...                               │   │
│  │    Duration: 45s · ✅ Success · Exit 0                             │   │
│  │    ─────────────────────────────────────────────────────────────   │   │
│  │    ok      screencast/pkg/media/gst    0.234s                       │   │
│  │    PASS                                                               │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  🏷️ Annotations:   [ai-pattern: bridge-implementation] [good: test-coverage]│
│                                                                             │
│  [💬 Add Comment]  [🔖 Bookmark]  [📤 Share]  [📊 Analyze Turn]           │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Interactive Features
- **Code blocks are syntax highlighted** with line numbers
- **Click file reference** → Jump to file explorer with blame view
- **Click [Diff View]** → Side-by-side before/after with line mapping
- **Click annotation tag** → Filter transcript to similar patterns
- **Keyboard navigation** (j/k for next/prev turn, / for search)

---

## Proposal 3: The Investigation Board

**Target:** Auditors, Safety Reviewers  
**Metaphor:** Detective investigation board with evidence clustering

### ASCII Layout

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  🕵️ Investigation Board: Session bbf1bdf1...                     [Export 🖨️]│
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │ 🔍 Focus Areas:  [All] [Errors ❌ 12] [Repeated Patterns 🔄 8]         │ │
│  │                  [Slow Operations ⏱️ 23] [Safety Flags ⚠️ 3]           │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│                                                                              │
│  ┌──────────────────┐      ┌──────────────────┐      ┌──────────────────┐  │
│  │ ⚠️ CLUSTER #1    │      │ 🔁 CLUSTER #2    │      │ ⏱️ CLUSTER #3    │  │
│  │                  │      │                  │      │                  │  │
│  │ Error Pattern:   │      │ Repeated Edit:   │      │ Slow Operation:  │  │
│  │ "gst element     │      │ recording.go     │      │ web_search for   │  │
│  │  not found"      │      │ modified 27x     │      │ "gstreamer       │  │
│  │                  │      │                  │      │  appsink caps"   │  │
│  │ Turns: 145, 203, │      │ Suggests:        │      │                  │  │
│  │  289, 412...     │      │ unclear design   │      │ Avg: 8.5s each   │  │
│  │                  │      │ or requirements  │      │ Total: 142s      │  │
│  │ 🔗 [View All 12] │      │ 🔗 [View Timeline│      │ 🔗 [View Queries]│  │
│  │ 📊 [Pattern Graph]│     │ 📊 [File History]│      │ 📊 [Duration Dist│  │
│  │ 🏷️ Auto-tagged:  │      │ 🏷️ ai-pattern:  │      │ 🏷️ performance:  │  │
│  │  element-error   │      │  high-velocity   │      │  slow-io         │  │
│  └──────────────────┘      └──────────────────┘      └──────────────────┘  │
│                                                                              │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │ 🧵 Thread: The GStreamer Bridge Implementation                        │ │
│  │                                                                       │ │
│  │ Start: Turn #234 (8:15 PM) → End: Turn #892 (11:42 PM)               │ │
│  │ Duration: 3h 27m · 47 tool calls · 12 files modified                 │ │
│  │                                                                       │ │
│  │ [Turn 234] Initial attempt with appsink ──────────────────────────→  │ │
│  │      ↓ [Turn 345] Error: caps negotiation failed                     │ │
│  │      ↓ [Turn 412] Pivot to tee + queue pattern                        │ │
│  │      ↓ [Turn 567] Testing with shared_video bridge                    │ │
│  │      ↓ [Turn 734] Integration with preview manager                    │ │
│  │      ↓ [Turn 892] ✅ Tests passing, PR ready                         │ │
│  │                                                                       │ │
│  │ 🔗 [View Full Thread]  [📊 Thread Metrics]  [🎥 Replay Thread]        │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│                                                                              │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │ 📝 Audit Checklist                                                    │ │
│  │ ☑️ No credential exposure in tool outputs                              │ │
│  │ ☑️ All bash commands had descriptions                                  │ │
│  │ ☑️ Error recovery patterns present                                     │ │
│  │ ☐ High token usage in turns 400-500 (investigate)                    │ │
│  │ ☑️ No PII detected in conversation                                   │ │
│  │                                                                       │ │
│  │ [Generate Audit Report]  [Export to PDF]                              │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Interactive Features
- **Drag clusters** to reorganize board
- **Draw connections** between related turns with annotation
- **Zoom into thread** → See full sub-transcript
- **Auto-clustering** suggestions based on patterns

---

## Proposal 4: The Query Interface View

**Target:** Power Users, Researchers  
**Metaphor:** SQL workbench + results browser

### ASCII Layout

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  🔬 Query Interface                                          [Save] [Export] │
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │ Query Repository: ▼ Custom / Saved / Built-in                          │ │
│  │                                                                       │ │
│  │ ┌─────────────────────────────────────────────────────────────────┐   │ │
│  │ │ /* sqleton                                                      │   │ │
│  │ │ name: gstreamer-file-heatmap                                     │   │ │
│  │ │ short: Visualize file activity as heatmap                        │   │ │
│  │ │ flags:                                                         │   │ │
│  │ │   - name: min_edits                                            │   │ │
│  │ │     type: int                                                  │   │ │
│  │ │     default: 5                                                 │   │ │
│  │ │ */                                                             │   │ │
│  │ │                                                                │   │ │
│  │ │ SELECT file_path, COUNT(*) as edits,                           │   │ │
│  │ │        COUNT(DISTINCT session_id) as sessions,                 │   │ │
│  │ │        STRING_AGG(DISTINCT tool_name, ', ') as tools_used      │   │ │
│  │ │ FROM {{TABLE_NAME}}, UNNEST(tool_calls) AS tc                  │   │ │
│  │ │ WHERE tc.operation_type IN ('MODIFY', 'NEW')                   │   │ │
│  │ │ GROUP BY file_path                                             │   │ │
│  │ │ HAVING COUNT(*) >= {{ .min_edits }}                            │   │ │
│  │ │ ORDER BY edits DESC;                                           │   │ │
│  │ └─────────────────────────────────────────────────────────────────┘   │ │
│  │                                                                       │ │
│  │ Parameters:  min_edits: [5] [▲▼]                                     │ │
│  │                                                                       │ │
│  │ [▶️ Run Query]  [📊 Visualize]  [💾 Save Query]  [📤 Export Results]   │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│                                                                              │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │ Results (25 rows) · Query time: 0.234s                                │ │
│  │                                                                       │ │
│  │ File Path                              │ Edits │ Sessions │ Tools      │ │
│  │ ────────────────────────────────────────┼───────┼──────────┼────────────│ │
│  │ ████████████████████ recording.go      │   27  │     1    │ read, edit │ │
│  │ ███████████████░░░░░ preview.go       │   18  │     1    │ read, edit │ │
│  │ ████████████░░░░░░░░░ shared_video.go   │   15  │     2    │ read, edit │ │
│  │ ████████░░░░░░░░░░░░░ bridge.go        │   12  │     1    │ read, edit │ │
│  │ ██████░░░░░░░░░░░░░░░ server.go        │    9  │     1    │ read, edit │ │
│  │                                                                       │ │
│  │ [📊 Heatmap View]  [📈 Trend Chart]  [📑 Raw JSON]                    │ │
│  │                                                                       │ │
│  │ Heatmap Visualization:                                                  │ │
│  │ ┌─────────────────────────────────────────────────────────────────┐   │ │
│  │ │ recording.go    [█][█][░][█][█][░][█][░][█][█][█][█]  12h span │   │ │
│  │ │ preview.go      [░][░][█][█][░][█][░][░][█][░][░][░]   8h span   │   │ │
│  │ │ shared_video.go [█][░][░][░][█][█][█][░][░][░][░][░]   6h span   │   │ │
│  │ │ bridge.go       [░][░][░][░][░][░][█][█][█][█][░][░]   4h span   │   │ │
│  │ └─────────────────────────────────────────────────────────────────┘   │ │
│  │ Time buckets → 6:00   8:00   10:00   12:00   14:00   16:00             │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Interactive Features
- **Live SQL editing** with syntax highlighting
- **Auto-complete** for table names and JSON paths
- **Multiple visualization types** (table, heatmap, chart, graph)
- **Export to** CSV, JSON, HTML, or embed in report

---

## Proposal 5: The Mobile-Optimized View

**Target:** All users on mobile devices  
**Metaphor:** Chat app + collapsible cards

### ASCII Layout (Mobile Portrait)

```
┌─────────────────┐
│ GStreamer...   │
│ ≡        🔍 ⚙️ │
├─────────────────┤
│                 │
│ Stats ▼         │
│ ├─ 🕐 24h      │
│ ├─ 🔄 1,170    │
│ ├─ 🛠️ 1,385    │
│ └─ 📊 36% read │
│                 │
│ ────────────────│
│                 │
│ Turn #47 🎙️    │
│ ├─ 3 tools ▶️  │
│ │  • read: ... │
│ │  • edit: ... │
│ │  • bash: ... │
│ ├─ ⏱️ 45s      │
│ └─ 🏷️ 2 tags   │
│                 │
│ ────────────────│
│                 │
│ Turn #48 🎙️    │
│ ├─ 2 tools ▶️  │
│ └─ ⏱️ 23s      │
│                 │
│ ────────────────│
│                 │
│ [⬆️]  Load     │
│ more (50/1170) │
│                 │
├─────────────────┤
│ 🏠  🔍  📊  👤 │
└─────────────────┘
```

### Swipe Gestures
- **Swipe left/right** → Next/previous turn
- **Pull down** → Refresh/load more
- **Pinch** → Expand/collapse tool details
- **Long press** → Annotate or bookmark

---

## Technical Implementation Notes

### Data Flow Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         HTML Export Pipeline                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌───────────────┐    ┌───────────────┐    ┌───────────────┐             │
│  │ Minitrace     │───→│ Query Engine  │───→│ View Models   │             │
│  │ JSON Archive  │    │ (DuckDB)      │    │ (Aggregated)   │             │
│  └─────────────────┘    └───────────────┘    └───────────────┘             │
│          │                    │                    │                          │
│          │                    │                    ↓                          │
│          │                    │            ┌───────────────┐               │
│          │                    │            │ Template      │               │
│          │                    │            │ Renderer      │               │
│          │                    │            │ (Go/JS)       │               │
│          │                    │            └───────┬───────┘               │
│          │                    │                    │                          │
│          │                    │                    ↓                          │
│          │                    │            ┌───────────────┐               │
│          │                    │            │ HTML Output   │               │
│          │                    │            │ - Static site │               │
│          │                    │            │ - SPA bundle  │               │
│          │                    │            │ - Embed widget│               │
│          │                    │            └───────────────┘               │
│          │                    │                                            │
│          │                    └────────┐                                 │
│          │                               │                                 │
│          └───────────────────────────────┘                                 │
│                    Query Commands Repository                               │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Rendering Strategy Options

1. **Server-Side Rendering (Go templates)**
   - Pros: Fast initial load, SEO-friendly, no JS dependency
   - Cons: Less interactive, requires regeneration for updates
   - Best for: Static reports, archival exports

2. **Single Page Application (React/Vue + Go API)**
   - Pros: Highly interactive, real-time updates, rich visualizations
   - Cons: Requires JS, more complex hosting
   - Best for: Live exploration, dashboards

3. **Hybrid (Go template + Progressive enhancement)**
   - Pros: Works without JS, enhanced with JS
   - Cons: More complex implementation
   - Best for: Universal accessibility

4. **Static Site Generator (Go → Markdown → HTML)**
   - Pros: Simple hosting (GitHub Pages), version controlled
   - Cons: Limited interactivity
   - Best for: Documentation, research papers

---

## Annotation Visualization Strategies

### Strategy 1: Inline Chips
```
Turn #47: Implementing the GStreamer bridge... [ai-pattern] [good]
```

### Strategy 2: Margin Markers
```
│ [📌]  Turn #47: Implementing the GStreamer bridge...
│ [🔴]    
│ [💡]    
```

### Strategy 3: Heatmap Overlay
```
Turns: ████░░░░░░████▓▓▓▓░░░░████░░░░  (density = annotation count)
```

### Strategy 4: Filtered Views
```
[Show All] [AI Patterns Only] [Errors Only] [Questions Only] [My Bookmarks]
```

---

## File Change Visualization

### Diff View Options

1. **Unified Diff** (compact)
2. **Side-by-Side** (clear comparison)
3. **Inline Annotation** (edit locations marked in context)
4. **File Timeline** (all edits to file over session)

### Code Block Enhancements
- Syntax highlighting
- Line numbers
- Click to expand/collapse
- Link to Git blame (if git integration)
- Show surrounding context (±3 lines default)

---

## Export Formats & Use Cases

| Format | Use Case | Interactivity | Portability |
|--------|----------|---------------|-------------|
| **Static HTML** | Archival, email sharing | Low | High |
| **Interactive HTML** | Deep review, presentations | High | Medium |
| **PDF** | Print, compliance reports | None | High |
| **Markdown** | Git repos, documentation | Low | Very High |
| **JSON** | Further analysis, ML training | N/A | Very High |
| **Embed Widget** | Dashboards, wikis | Medium | Medium |

---

## Open Questions

1. How should we handle very large sessions (10k+ turns)?
   - Virtual scrolling?
   - Pagination with deep links?
   - Summary view + drill down?

2. What's the right balance between client-side and server-side rendering?
   - Pre-render first N turns, lazy load rest?
   - Full SPA for exploration, static export for sharing?

3. How do we preserve interactivity in shared exports?
   - Self-contained bundle with embedded data?
   - External data reference (requires hosting)?

4. What's the annotation editing story?
   - Read-only exports vs. collaborative annotation?
   - How to sync back to source?

---

## Next Steps

1. Create HTML export prototype using Go templates + HTMX
2. Implement query command for "export-ready" view models
3. Design annotation rendering component library
4. Test with actual GStreamer session data
5. Gather feedback from each user persona
6. Iterate on interaction patterns

---

## Related References

- go-minitrace query commands (structured SQL templates)
- Minitrace schema documentation
- Annotation taxonomy (MAST, ToolEmu, minitrace)
- Existing transcript analysis queries in this ticket
