# Changelog

## 2026-04-15

Step 1: Created GMT-007 for Proposal 4 timeline visualization work, added a detailed implementation/analysis guide, and reframed the plan as SQL-first so the temporal analyses can also be packaged and reused as query commands.

## 2026-04-15

Step 2: Added the first Go-side bridge package for Proposal 4 under `pkg/exporttimeline`, with typed row models and a SQL-backed loader that executes the timeline query commands, decodes their result sets, and returns them as a single `SQLTimelineData` bundle.

## 2026-04-15

Step 3: Added the first normalized timeline payload layer on top of the SQL results, including dense file-series generation, idle-window normalization, and Proposal 2-compatible `#turn-*` jump hashes for buckets, phase signals, and thread signals.

## 2026-04-15

Step 4: Added merged Proposal 4 phase/thread structures on top of the SQL signals, including annotation-backed manual/import support via `timeline-phase` / `timeline-thread` tags plus detail JSON, grouped derived phase markers/thread spans, and merge rules that let curated markers coexist with or override matching derived candidates.

## 2026-04-15

Step 5: Added the first browser-rendering slice for Proposal 4 with a self-contained HTML/SVG runtime (`pkg/exporttimeline/templates/*`), a new `go-minitrace export timeline` command, timeline export builder/renderer plumbing, and interactive visuals for heatmap/file bands/idle shading/phase ribbon/thread bars with hover labels and click-to-reader hash navigation.
