# Changelog

## 2026-04-15

Step 1: Created GMT-007 for Proposal 4 timeline visualization work, added a detailed implementation/analysis guide, and reframed the plan as SQL-first so the temporal analyses can also be packaged and reused as query commands.

## 2026-04-15

Step 2: Added the first Go-side bridge package for Proposal 4 under `pkg/exporttimeline`, with typed row models and a SQL-backed loader that executes the timeline query commands, decodes their result sets, and returns them as a single `SQLTimelineData` bundle.

## 2026-04-15

Step 3: Added the first normalized timeline payload layer on top of the SQL results, including dense file-series generation, idle-window normalization, and Proposal 2-compatible `#turn-*` jump hashes for buckets, phase signals, and thread signals.
