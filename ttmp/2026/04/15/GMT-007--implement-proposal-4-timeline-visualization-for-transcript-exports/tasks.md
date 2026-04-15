# Tasks

## Proposal 4 timeline implementation

### Scope and payload design
- [x] Finalize timeline export payload schema
- [x] Separate original annotations from derived timeline markers in the payload model
- [x] Define jump-target contracts from timeline elements back into reader turns
- [x] Decide first bucket size default (15m vs 30m)

### SQL-first analysis foundation
- [x] Create `query-commands/timeline/timeline-buckets.sql`
- [x] Create `query-commands/timeline/timeline-file-activity.sql`
- [x] Create `query-commands/timeline/timeline-idle-windows.sql`
- [x] Create `query-commands/timeline/timeline-phase-signals.sql`
- [x] Create `query-commands/timeline/timeline-thread-signals.sql` (even if initially partial/simple)
- [x] Validate each query command against real fixture sessions
- [x] Document expected columns/result contracts for each query command

### Go-side payload assembly
- [x] Add timeline export package and types
- [x] Add query-result loader/merger for timeline query commands
- [x] Build bucket arrays from SQL output
- [x] Build file series from SQL output
- [x] Build idle windows from SQL output
- [x] Merge manual/imported phase spans with SQL-derived phase signals
- [x] Merge manual/imported thread spans with SQL-derived thread signals
- [x] Emit deterministic reader jump targets for buckets/phases/threads

### Browser rendering
- [x] Add self-contained timeline HTML template/runtime
- [x] Render heatmap from bucket arrays
- [x] Render file activity bands
- [x] Render idle window shading
- [x] Render phase ribbon
- [x] Render thread bars with segmented support
- [x] Add hover labels / legends
- [x] Add click-to-reader navigation

### Validation
- [x] Add backend tests for bucket parsing and timeline payload rendering
- [ ] Add golden tests for query-command outputs on fixture sessions
- [x] Validate timeline export against session `bbf1bdf1-364a-44cb-8cd0-ebcba86dd1ad`
- [x] Verify browser view has no external asset/data requests
- [ ] Verify timeline clicks land on sensible reader targets
- [ ] Verify output remains understandable and responsive on a long session

### Documentation and workflow
- [ ] Add a workflow note explaining how timeline query commands feed the exporter
- [ ] Add a playbook for manual/imported phase and thread span preparation
- [ ] Keep changelog updated as implementation phases complete

