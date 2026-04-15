---
Title: 'Proposal 2 - Chronological Reader: Annotation, Export, and Visualization Guide'
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
RelatedFiles: []
ExternalSources: []
Summary: "Simplified intern-facing guide for building the read-only chronological reader using only raw transcript data plus human/LLM-authored annotations, with a concrete React reuse plan"
LastUpdated: 2026-04-14T16:45:00-04:00
WhatFor: "Detailed implementation guide for a simple Proposal 2 reader without synthetic annotations or heuristic threading"
WhenToUse: "When implementing a generic self-contained HTML transcript reader powered by embedded JSON"
---

# Proposal 2 - Chronological Reader: Annotation, Export, and Visualization Guide

## Executive Summary

This document defines a **simplified** version of Proposal 2: the **Chronological Reader**.

The simplification is intentional:

- **no synthetic annotations**,
- **no heuristic thread detection**,
- **no inferred high-signal markers**,
- **no auto-generated narrative overlays**.

Instead, the reader is built from only three kinds of information:

1. the raw transcript,
2. deterministic structural reshaping of that transcript for presentation,
3. and **original annotations** authored by humans or an LLM following an annotation guide.

This version is a better starting point because it reduces product and implementation risk. It keeps the reader focused on what the transcript actually says and what annotators explicitly added.

A new intern should think about this proposal like this:

> We are not building an intelligent transcript critic.
> We are building a very good read-only transcript browser.

The browser should be able to:

- render a session in chronological order,
- show turns and tool calls clearly,
- show original annotations attached to session / turn / tool call,
- search the transcript,
- filter by roles and tools,
- link directly to turns and tool calls,
- and run entirely from a single HTML file with embedded JSON.

---

## What Changed From the Previous Version

This guide intentionally removes the earlier complexity.

### Removed

- synthetic annotations
- heuristic thread markers
- auto-detected pivots
- inferred high-signal turns
- export-time “narrative intelligence”

### Kept

- chronological reading experience
- original annotations
- deterministic grouping and indexing
- search
- filters
- expandable tool call details
- self-contained HTML export

### Why this is better for V1

Because it makes the system:

- easier to reason about,
- easier to test,
- easier to explain to users,
- easier to implement with current code,
- and easier to trust.

---

## Product Goal

The Chronological Reader should answer these questions well:

- “What happened in this session?”
- “What did the assistant do after this prompt?”
- “Which files were read or changed in this region?”
- “What did this bash command output?”
- “What annotations did a reviewer add here?”
- “Can I jump to the exact turn or tool call someone referenced?”

It should **not** try to answer:

- “What hidden thread did the assistant probably follow?”
- “What phase do we think this turn belongs to?”
- “Which turns are high-value according to heuristics?”

Those are future problems. V1 should stay literal.

---

## What Data the Reader Uses

The reader should use only data that is either:

1. already present in minitrace,
2. or deterministically derived from it for display.

### Raw minitrace fields used directly

From the session:

- `id`
- `title`
- `summary`
- `classification`
- `provenance`
- `environment`
- `operational_context`
- `timing`
- `metrics`
- `turns[]`
- `tool_calls[]`
- `annotations[]`

From each turn:

- `index`
- `timestamp`
- `role`
- `source`
- `content`
- `thinking`
- `model`
- `usage`
- `tool_calls_in_turn`

From each tool call:

- `id`
- `emitting_turn_index`
- `timestamp`
- `tool_name`
- `operation_type`
- `input.file_path`
- `input.command`
- `input.arguments`
- `output.success`
- `output.result`
- `output.error`
- `output.duration_ms`
- `output.truncated`

From each annotation:

- `annotator`
- `scope.type`
- `scope.target_id`
- `content.category`
- `content.tags`
- `content.title`
- `content.detail`
- `taxonomy_mappings`
- `classification`

### Deterministic derived structures allowed

These are okay because they do not add interpretive meaning. They only reshape data for rendering.

Allowed examples:

- tool calls grouped by turn
- turns grouped into blocks by user prompt
- annotations indexed by target scope
- inverted search index
- file path lookup index
- URL/hash lookup maps
- counts for filters (e.g. how many bash tool calls)

These are **presentation indices**, not synthetic semantics.

---

## Annotation Policy for the Simplified Reader

This section is central to the simplified design.

### Rule 1: Only original annotations are shown as annotations

The reader may display:

- annotations already present in the minitrace session,
- or annotations deliberately added by a human or LLM annotation pass before export.

The reader may **not** invent annotation objects during export.

### Rule 2: Annotation work can still be guided

Even though the annotations are not synthetic, the process for creating them can still be guided by an LLM or a human rubric.

That means we still need an annotation guide. We just do not serialize guesses as if they were system-generated truth.

### Rule 3: Annotation scope stays exactly within the existing schema

Use only:

- `session`
- `turn`
- `tool_call`

Do not invent range annotations or timeline-only annotation object types for Proposal 2.

If someone wants to mark a range, they should annotate the start turn and mention the span in `detail`, or add multiple turn annotations.

---

## Human / LLM Annotation Guide

The purpose of annotation in the simplified reader is to help future humans understand the transcript without changing the transcript itself.

### What should be annotated

A human or LLM annotator should prefer **high-value explicit notes** over dense blanket coverage.

Good candidates:

- important user requirement clarifications,
- assistant decisions that affect later work,
- tool calls that reveal a decisive error,
- tool calls that show a successful test or milestone,
- turns where a reviewer would want to leave context for a future engineer.

### What should not be annotated

Avoid annotating:

- every ordinary file read,
- every edit,
- every successful bash command,
- routine continuation chatter,
- information that is already obvious from the immediate transcript.

### Recommended categories

Stay within the current known categories:

- `observation`
- `ai-failure`
- `user-error`
- `environment-issue`
- `success`
- `question`
- `to-discuss`
- `to-improve`

### Annotation examples

#### Turn-level observation

```json
{
  "scope": { "type": "turn", "target_id": "47" },
  "content": {
    "category": "observation",
    "title": "Bridge implementation starts here",
    "detail": "This is the first turn where the assistant explicitly commits to the appsink/appsrc bridge approach.",
    "tags": ["bridge", "gstreamer", "implementation"]
  }
}
```

#### Tool-call-level success

```json
{
  "scope": { "type": "tool_call", "target_id": "call-xyz" },
  "content": {
    "category": "success",
    "title": "First passing test",
    "detail": "This tool call is the first visible successful test run after repeated bridge-related failures.",
    "tags": ["test", "milestone"]
  }
}
```

#### Tool-call-level error

```json
{
  "scope": { "type": "tool_call", "target_id": "call-abc" },
  "content": {
    "category": "environment-issue",
    "title": "Missing GStreamer element",
    "detail": "The bash output indicates the environment is missing a required runtime/plugin element, which may not be a code bug.",
    "tags": ["gstreamer", "runtime", "dependency"]
  }
}
```

### LLM annotation prompt skeleton

```text
You are annotating a transcript for a chronological reader UI.
Use only the existing minitrace annotation schema.
Do not invent ranges, threads, or synthetic categories.

Add annotations only where a future engineer would genuinely benefit.
Prefer explicit, grounded observations over speculation.

Choose from these categories:
- observation
- ai-failure
- user-error
- environment-issue
- success
- question
- to-discuss
- to-improve

Good annotation targets:
- decisive turn-level decisions
- important failing tool calls
- important successful tests
- ambiguous requirements
- notes that explain later choices

Do not annotate routine reads or obvious actions.
```

### Existing CLI reference for manual annotation

File reference:
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/cmd/go-minitrace/cmds/annotate/add.go`

Example manual CLI usage:

```bash
go-minitrace annotate add \
  --output-dir ./output \
  --session bbf1bdf1-364a-44cb-8cd0-ebcba86dd1ad \
  --scope turn \
  --target-id 47 \
  --annotator user \
  --category observation \
  --title "Bridge implementation starts here" \
  --detail "First explicit commitment to the shared appsink/appsrc bridge approach" \
  --tags bridge,gstreamer,implementation
```

Important workflow reminder:
- if annotations are stored externally through the annotate store, they must be synced back into the archive before export if the export reads `.minitrace.json` files directly.

---

## The Simplified Export Model

The easiest path is to make the export payload match the current web transcript viewer’s mental model as closely as possible.

That means exporting a generic JSON blob that looks like:

```json
{
  "session": { ...session summary/detail... },
  "blocks": [ ...block list... ],
  "annotations": [ ...session/turn/tool_call annotations... ]
}
```

This is important because the existing React transcript viewer already works with:

- a session summary/detail object,
- a list of `SessionBlock`s,
- and a list of `Annotation`s.

### Deterministic block formation

The current web UI is block-oriented. A block is:

- one user prompt,
- plus all assistant/system turns until the next user prompt.

This is deterministic and safe. It is not heuristic interpretation.

So for Proposal 2, we should keep that structure.

### Export payload shape

Recommended generic export schema:

```json
{
  "version": "reader-export-v1",
  "session": {
    "id": "bbf1bdf1-364a-44cb-8cd0-ebcba86dd1ad",
    "title": "Create a new docmgr ticket to port this from ffmpeg streams to gstreamer...",
    "summary": null,
    "classification": "internal",
    "timing": { ... },
    "metrics": { ... },
    "environment": { ... },
    "operational_context": { ... },
    "provenance": { ... }
  },
  "blocks": [
    {
      "block_num": 1,
      "user_turn_idx": 0,
      "user_ts": "2026-04-13T18:11:37Z",
      "user_content": "Create a new docmgr ticket...",
      "agent_turns": 12,
      "tool_calls": 27,
      "gap_minutes": null,
      "turns": [ ... ],
      "artifacts": {
        "commits": [],
        "tickets_created": ["SCS-0012"],
        "docs_added": [],
        "diary_writes": 1
      }
    }
  ],
  "annotations": [ ... ],
  "indices": {
    "annotation_counts": {
      "session": 2,
      "turn": 8,
      "tool_call": 5
    },
    "turn_to_annotations": {
      "47": ["ann-1", "ann-2"]
    },
    "tool_call_to_annotations": {
      "call-xyz": ["ann-9"]
    },
    "search": {
      "terms": {
        "gstreamer": [2, 14, 47],
        "appsink": [47, 51, 58]
      }
    }
  }
}
```

### Why this payload is good

Because it is:

- generic,
- JSON-only,
- decoupled from the live API,
- easy to embed in a `<script type="application/json">` tag,
- and very close to the current web UI’s existing TypeScript types.

---

## What We Can Reuse From the Current React / JS / HTML

This is the practical part.

The current `go-minitrace/web` application already has a transcript viewer. We should reuse as much as possible, but not blindly.

### Best reusable pieces

#### 1. Type definitions

File reference:
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/web/src/types/session.ts`

This file is very valuable because it already defines:

- `SessionDetail`
- `SessionBlock`
- `Turn`
- `ToolCall`
- `Annotation`

These types are already close to what the export JSON should look like.

**Recommendation:**
- keep these types,
- or create export-specific equivalents that intentionally mirror them.

#### 2. TranscriptViewer component structure

File references:
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/web/src/components/TranscriptViewer/TranscriptViewer.tsx`
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/web/src/components/TranscriptViewer/BlockCard.tsx`
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/web/src/components/TranscriptViewer/BlockBody.tsx`
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/web/src/components/TranscriptViewer/ToolCallRow.tsx`

These are the strongest reuse candidates.

Why:
- they already render block-based transcripts,
- they already render turns and tool calls,
- they already support expansion and annotation display,
- they already know how to display diffs and bash output reasonably well.

#### 3. Virtual list support

File reference:
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/web/src/components/shared/useVirtualList.ts`

If exports become large, this is potentially reusable.

#### 4. Small shared display helpers

Potential reuse candidates:
- `FormatDuration.tsx`
- `ToolCallBadge.tsx`
- simple status chips / badges

### Parts we should not reuse directly

#### 1. RTK Query / live API layer

File reference:
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/web/src/api/minitrace.ts`

This is designed for a live backend. A self-contained export should not depend on it.

#### 2. Router/page structure

File references:
- `pages/TranscriptViewerPage.tsx`
- `App.tsx`

These are app-shell concerns, not export concerns.

#### 3. Annotation mutation UI

File references:
- `AnnotationComposer.tsx`
- `AnnotationModal.tsx`
- mutation hooks in `api/minitrace.ts`

Because the export is read-only, the HTML should not include annotation authoring or saving.

#### 4. Query editor and session browser

These are unrelated to the single-session reader export.

### Minimal reuse plan

The simplest approach is:

- reuse the transcript viewer rendering components,
- remove API dependencies,
- remove mutation UI,
- feed them with embedded JSON instead of fetched data.

---

## Recommended Architecture: Generic JSON-Driven Reader

The best V1 design is to make the reader generic.

That means the browser app should not know anything about one specific export format beyond a stable JSON schema.

### Generic export template

```html
<!doctype html>
<html>
  <head>
    <meta charset="utf-8" />
    <title>{{ .Title }}</title>
    <style>{{ .EmbeddedCSS }}</style>
  </head>
  <body>
    <div id="root"></div>
    <script id="minitrace-export-data" type="application/json">{{ .ExportJSON }}</script>
    <script>{{ .EmbeddedJS }}</script>
  </body>
</html>
```

This is already “entirely generic and just load a json.”

### Browser boot logic

The bundle startup code should do roughly this:

```ts
const raw = document.getElementById("minitrace-export-data")?.textContent ?? "{}";
const data = JSON.parse(raw) as TranscriptExportPayload;
mountReaderApp(document.getElementById("root"), data);
```

This is preferable to trying to pre-render the entire transcript to static HTML in Go for V1.

Why:
- less template complexity,
- easier reuse of the current React viewer,
- easier iteration,
- easier to keep the viewer generic.

---

## How To “Compile It Down” to an HTML Template

There are three realistic options.

### Option A: Keep React, compile to JS/CSS bundle, inject into HTML template

This is the most practical short-term approach.

#### How it works

1. Build a dedicated export entrypoint in the web app.
2. Bundle it with Vite into one JS file and one CSS file.
3. During export, Go reads those built assets.
4. Go injects JS, CSS, and export JSON into a generic HTML template.
5. Browser loads the generic shell and mounts the app from embedded data.

#### Why it is good

- maximum reuse of current viewer code,
- simple mental model,
- generic JSON-driven runtime,
- still results in a single self-contained HTML file.

#### What to add

Proposed new files:

```text
web/src/export/
  readerExportMain.tsx
  ExportTranscriptViewerPage.tsx
  loadEmbeddedExport.ts
```

#### Pseudocode

```tsx
// readerExportMain.tsx
import { createRoot } from "react-dom/client";
import { ExportTranscriptViewerPage } from "./ExportTranscriptViewerPage";
import { loadEmbeddedExport } from "./loadEmbeddedExport";

const data = loadEmbeddedExport();
const root = createRoot(document.getElementById("root")!);
root.render(<ExportTranscriptViewerPage data={data} />);
```

```ts
// loadEmbeddedExport.ts
export function loadEmbeddedExport(): TranscriptExportPayload {
  const el = document.getElementById("minitrace-export-data");
  if (!el?.textContent) throw new Error("missing embedded export data");
  return JSON.parse(el.textContent);
}
```

### Option B: Server-render React into static HTML during export

This is possible but I do **not** recommend it for V1.

Why not:
- MUI/Emotion SSR is more complex,
- you still need JS for interactions,
- export pipeline becomes much heavier.

### Option C: Rewrite the reader as Go templates + small vanilla JS

This is probably the best long-term maintenance architecture, but not the best first implementation if we want reuse.

Why:
- smaller output bundle,
- easier offline static behavior,
- less frontend dependency surface.

But it requires reimplementing UI that already exists.

### Recommendation

Use **Option A now**, and consider **Option C later** if the export product becomes important enough to justify a lighter specialized renderer.

---

## Concrete Reuse Plan

### Step 1: Create export-specific wrapper around existing TranscriptViewer

The current `TranscriptViewer.tsx` expects fetched annotations via `useGetSessionAnnotationsQuery`. That is not suitable for static export.

So create a new export-specific wrapper that passes embedded annotations directly.

You have two choices:

#### Choice A: Fork a simplified static variant

Example:

```text
web/src/components/TranscriptExportViewer/
  TranscriptExportViewer.tsx
  BlockCard.tsx        (reuse or shallow fork)
  BlockBody.tsx        (reuse or shallow fork)
  ToolCallRow.tsx      (reuse)
```

This is the cleanest approach.

#### Choice B: Refactor TranscriptViewer into data-provider-free core

Split current component into:
- a container that fetches,
- a pure presentational core that accepts `session` + `annotations` props.

This is better architecturally, but slightly more refactoring.

### Step 2: Reuse existing block and tool rendering

Best candidates for near-direct reuse:

- `BlockCard.tsx`
- `BlockBody.tsx`
- `ToolCallRow.tsx`

Necessary changes:

- remove annotation creation buttons,
- remove annotation modal usage,
- replace live fetched annotation index with precomputed embedded annotation index,
- simplify tabs if the export does not need separate annotation panel view.

### Step 3: Use export JSON that matches current UI shapes

If the export JSON matches `SessionDetail`, `SessionBlock`, and `Annotation`, then reuse becomes straightforward.

This is one of the strongest reasons to keep the export generic JSON close to the current web types.

---

## Deterministic Export-Time Processing

Even without heuristics, we still need deterministic preparation.

### Processing steps

1. load session,
2. group tool calls by emitting turn,
3. group turns into blocks by user prompts,
4. attach original annotations by scope,
5. build search index,
6. shape JSON payload,
7. embed JSON + compiled JS/CSS into HTML.

### Pseudocode

```go
func BuildSimpleReaderExport(session *minitrace.Session) (*TranscriptExportPayload, error) {
    blocks := BuildBlocksFromTurns(session.Turns)
    toolsByTurn := GroupToolCallsByTurn(session.ToolCalls)

    for i := range blocks {
        for j := range blocks[i].Turns {
            turn := &blocks[i].Turns[j]
            turn.ToolCallsInTurn = toolsByTurn[turn.Index]
        }
    }

    annotations := session.Annotations
    indices := BuildDeterministicReaderIndices(blocks, annotations)

    return &TranscriptExportPayload{
        Version:     "reader-export-v1",
        Session:     BuildSessionDetail(session),
        Blocks:      AdaptBlocks(blocks),
        Annotations: annotations,
        Indices:     indices,
    }, nil
}
```

### Search index is still okay

A search index is not a heuristic interpretation. It is a deterministic lookup structure.

So Proposal 2 can still include:
- full-text token index,
- file-path term index,
- annotation term index.

That keeps search fast without changing meaning.

---

## Visualization Implementation Guide

### Core UI sections

1. **Header**
   - title
   - session id
   - model/framework
   - search box
   - role/tool filters

2. **Block list**
   - one block per user prompt region
   - collapse/expand block body

3. **Turn body**
   - role
   - turn number
   - content
   - optional thinking
   - turn annotations

4. **Tool call rows**
   - one row per tool call
   - expandable diff/output
   - tool-call annotations

5. **Optional annotation side summary**
   - count by scope/category
   - jump links to annotated elements

### Suggested interactions

Allowed interactions should only affect local view state:

- search query
- role filter
- tool filter
- collapse/expand block
- collapse/expand tool call detail
- scroll to `#turn-47`
- scroll to `#tool-call-abc`

### Pseudocode: rendering with embedded annotations

```tsx
function ExportTranscriptViewer({ data }: { data: TranscriptExportPayload }) {
  const annotationIndex = useMemo(() => buildAnnotationIndex(data.annotations), [data.annotations]);
  const [search, setSearch] = useState("");
  const [roleFilter, setRoleFilter] = useState<string | null>(null);

  const visibleBlocks = useMemo(() => {
    return filterBlocks(data.blocks, { search, roleFilter, index: data.indices.search });
  }, [data.blocks, data.indices.search, roleFilter, search]);

  return (
    <div>
      <Header session={data.session} search={search} onSearch={setSearch} />
      {visibleBlocks.map((block) => (
        <BlockCard
          key={block.block_num}
          block={block}
          turnAnnotations={annotationIndex.byTurn}
          toolCallAnnotations={annotationIndex.byToolCall}
        />
      ))}
    </div>
  );
}
```

### Accessibility and offline behavior

The exported reader should:

- use semantic headings where possible,
- keep anchorable ids for turns and tool calls,
- degrade reasonably if search JS fails,
- never require network access,
- never reference external font/CDN/script URLs.

---

## Suggested File Layout

### In the web app

```text
web/src/export/
  readerExportMain.tsx
  ExportTranscriptViewerPage.tsx
  loadEmbeddedExport.ts
  types.ts

web/src/components/TranscriptExportViewer/
  TranscriptExportViewer.tsx
  BlockCard.tsx
  BlockBody.tsx
  ToolCallRow.tsx
```

### In go-minitrace exporter code

```text
cmd/go-minitrace/cmds/export/
  root.go
  html.go

pkg/export/html/
  builder.go
  assets.go
  reader_export.go
  template.go
```

### Generic HTML template asset

```text
templates/export/
  reader-shell.html.tmpl
```

---

## File and API References

### Current web UI files worth reusing

- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/web/src/types/session.ts`
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/web/src/components/TranscriptViewer/TranscriptViewer.tsx`
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/web/src/components/TranscriptViewer/BlockCard.tsx`
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/web/src/components/TranscriptViewer/BlockBody.tsx`
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/web/src/components/TranscriptViewer/ToolCallRow.tsx`
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/web/src/components/shared/useVirtualList.ts`
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/web/src/api/sessionProtoAdapters.ts`

### Existing backend/schema files to understand

- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/cmd/go-minitrace/main.go`
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/query/engine.go`
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/minitrace/schema.go`
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/doc/minitrace-schema.md`
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/cmd/go-minitrace/cmds/annotate/add.go`

### Ticket-local references

- `ttmp/2026/04/14/GST-2026-04-13--gstreamer-pi-sessions-analysis-with-go-minitrace/query-commands/gstreamer-tool-analysis.sql`
- `ttmp/2026/04/14/GST-2026-04-13--gstreamer-pi-sessions-analysis-with-go-minitrace/query-commands/top-modified-files.sql`
- `ttmp/2026/04/14/GST-2026-04-13--gstreamer-pi-sessions-analysis-with-go-minitrace/design-doc/03-html-readonly-exports-with-analysis.md`

---

## Implementation Recommendation

If you want the shortest path to a working generic Proposal 2 export, do this:

### Recommended path

1. **Do not invent synthetic annotations.**
2. **Export a generic JSON payload** matching current transcript viewer shapes.
3. **Fork or refactor the existing React TranscriptViewer** into a static export variant.
4. **Bundle it with Vite** as a self-contained JS/CSS export app.
5. **Inject JSON + JS + CSS into a generic Go HTML template**.

### Why this is the right tradeoff

Because it gives you:

- reuse of real UI code,
- a generic JSON-driven format,
- a read-only self-contained output,
- a much smaller product scope,
- and a clean upgrade path later.

If later you decide to replace React with a lighter template-based renderer, you can do that without changing the export payload format.

That is the key architectural win:

> make the export payload stable first,
> and the rendering implementation replaceable later.
