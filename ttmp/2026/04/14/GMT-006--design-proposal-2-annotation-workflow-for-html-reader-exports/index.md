---
Title: Design Proposal 2 annotation workflow for HTML reader exports
Ticket: GMT-006
Status: active
Topics:
    - go-minitrace
    - transcript-analysis
    - html-export
    - readonly-exports
DocType: index
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Future-work ticket for defining a disciplined annotation workflow for Proposal 2 chronological reader exports using the existing minitrace annotation schema and tooling."
LastUpdated: 2026-04-14T23:05:00-04:00
WhatFor: "Capture the design and later implementation plan for creating high-value human/LLM-authored annotations before generating Proposal 2 HTML exports."
WhenToUse: "Use this ticket when we are ready to operationalize annotation passes for the chronological reader."
---

# Design Proposal 2 annotation workflow for HTML reader exports

## Overview

This ticket captures a future-work stream for Proposal 2: how to create, review, sync, and export high-value annotations for the read-only chronological reader without introducing synthetic annotations into the export itself.

The core goal is to define a practical workflow that uses the existing `go-minitrace annotate` commands and schema, while allowing careful LLM assistance for candidate selection and draft generation.

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
