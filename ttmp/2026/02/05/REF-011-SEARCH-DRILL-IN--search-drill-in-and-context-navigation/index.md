---
Title: Search Drill-In and Context Navigation
Ticket: REF-011-SEARCH-DRILL-IN
Status: active
Topics:
    - ui
    - refactorio
    - frontend
    - search
DocType: index
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-06T00:00:00-05:00
WhatFor: "Design and implement stable drill-in navigation from unified search results into contextual destination views."
WhenToUse: "Use when implementing result navigation, deep links, and target hydration across Search and destination pages."
---

# Search Drill-In and Context Navigation

## Overview

This ticket adds the missing navigation layer between unified search and contextual investigation views.
The goal is to make every search result actionable, URL-addressable, and reload/share-safe by defining per-type drill-in routes and destination page hydration behavior.
The detailed design and implementation plan is captured in `design/01-search-drill-in-detailed-analysis-and-implementation-guide.md`.

## Key Links

- **Detailed Guide**: [design/01-search-drill-in-detailed-analysis-and-implementation-guide.md](./design/01-search-drill-in-detailed-analysis-and-implementation-guide.md)
- **Analysis Overview**: [analysis/01-search-drill-in-analysis-overview.md](./analysis/01-search-drill-in-analysis-overview.md)
- **Diary**: [reference/01-diary.md](./reference/01-diary.md)
- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- ui
- refactorio
- frontend
- search

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
