---
Title: Search Drill-In Analysis Overview
Ticket: REF-011-SEARCH-DRILL-IN
Status: active
Topics:
    - ui
    - refactorio
    - frontend
    - search
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Analysis overview and investigation plan to design search result drill-in flows and contextual navigation."
LastUpdated: 2026-02-05T17:05:00-05:00
WhatFor: "Guide the analysis required to design a complete search-to-context workflow." 
WhenToUse: "Use when preparing the search drill-in UX and routing design." 
---

# Search Drill-In Analysis Overview

## Goal
Design the end-to-end search drill-in flow so users can navigate from a search result into the appropriate detail view (file, symbol, code unit, diff, commit, doc hit), and preserve context along the way.

## Context
- The current Search page shows results but does not provide drill-in actions.
- The UI design spec emphasizes “read in context” and “jump between definition ↔ references ↔ history”.
- Backend search returns type-specific payloads which can support deep linking.

## Primary Questions
- What is the canonical route per result type?
- How should selection and context be preserved when navigating from search results?
- What metadata must the search result contain to enable deep links?

## In-Depth Analysis Plan
1. **Map result types to target views**
   - Define a target page for each result type (symbol, code unit, commit, diff, doc, file).
   - Identify required query params (hash, path, run_id, line number).

2. **Define navigation rules**
   - Determine whether drill-in uses full route navigation or opens side panels.
   - Decide if results should open in existing inspector panels or new pages.

3. **Search result payload requirements**
   - For each type, list required fields to build a deep link.
   - Verify backend search payloads include the necessary fields or update backend.

4. **Session scoping interaction**
   - Define how session context affects search results and deep links.
   - Decide whether drill-in should re-scope if the result is outside current session.

5. **UI/UX flows**
   - Design result item interactions (click, open in new tab, etc.).
   - Define breadcrumb or back-navigation behavior.

## Deliverables
- A routing map for each search result type.
- A list of required search payload fields for deep linking.
- Recommended UI interaction model (page navigation vs panel open).

## Suggested Review Start
- `refactorio/ui/src/pages/SearchPage.tsx`
- `refactorio/ui/src/components/data-display/SearchResults.tsx`
- `refactorio/pkg/workbenchapi/search.go`
