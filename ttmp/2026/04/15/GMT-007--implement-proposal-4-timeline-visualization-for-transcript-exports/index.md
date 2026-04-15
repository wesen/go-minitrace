---
Title: Implement Proposal 4 timeline visualization for transcript exports
Ticket: GMT-007
Status: active
Topics:
    - go-minitrace
    - transcript-analysis
    - html-export
    - readonly-exports
    - ui-design
DocType: index
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Implementation ticket for Proposal 4: a SQL-first, read-only, self-contained timeline visualization export for transcript sessions."
LastUpdated: 2026-04-15T01:45:00-04:00
WhatFor: "Track the implementation plan for a timeline export that derives reusable temporal analysis through SQL query commands and renders it as a self-contained HTML view."
WhenToUse: "Use this when planning or implementing the timeline export path and its supporting query commands."
---

# Implement Proposal 4 timeline visualization for transcript exports

## Overview

This ticket tracks Proposal 4 as a distinct export mode from the chronological reader. The core idea is to derive as much temporal structure as possible through SQL queries that can also be packaged as reusable query commands, then merge those results into a self-contained timeline HTML export.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- go-minitrace
- transcript-analysis
- html-export
- readonly-exports
- ui-design

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
