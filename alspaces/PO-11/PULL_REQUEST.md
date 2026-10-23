# docs(plan): live observability refinements, comparison scatter plot, and agent/skill tracking (PO-11)

## Summary
Defines the product specification, advisor risk analysis, probing questions, resolved steering decisions, and 13 vertical tracer-bullet Jira tickets (AL-44 through AL-56) covering live observability inspector refinements, cross-ticket correlation visualization, debug prompt streamlining, and complete agent and skill invocation observability:
1. Global localized millisecond timestamp formatting (`HH:mm:ss.SSS`) across Live Inspector, Waterfall, and Deep Dive views, with header timezone indicator (AL-44).
2. Monospace debug ID token middle truncation (`ses_0192...304a`) and flex container overflow containment (AL-45).
3. Collapsible Failed Tool Calls section with default-collapsed state and prominent failure count badge (AL-46).
4. Real-time live trace dynamic refresh and active session polling in the right inspector pane (AL-47).
5. Interactive multi-metric ticket scatter plot in the Cross-Ticket Comparison view with default X=Cost, Y=Net Lines of Code, outlier exclusion filter, and dynamic linear/log scaling (AL-48).
6. Reverse event chronology (newest first), chronological order toggle, pause-on-scroll lock, and quick-scroll navigation in the Live Trace Inspector (AL-49).
7. Audit and streamline debug prompt generation and eliminate destructive payload sanitization in `web/src/utils/debugPrompt.ts` (AL-50).
8. Schema v4 evolution, domain invariants, and backward-compatible V3->V4 migration engine (AL-51).
9. OpenCode SQLite database extractor and backfill engine for skills and subagent task linking (AL-52).
10. Live observability stream extraction and dual-plane skill/task ingestion (AL-53).
11. Rich timeline badges (`[skill: name]`, `[task -> role]`) and clickable subagent session jump links in Trace Waterfall and Live Inspector (AL-54).
12. Dedicated Skills & Subagents autopsy panel in Ticket Deep Dive (AL-55).
13. Fleet-wide skill adoption analytics and multi-select comparison matrix filter (AL-56).

## Initiatives Covered

1. **Global Localized Millisecond Timestamp Formatting and Timezone Header (AL-44)**:
   - Formats event and turn timestamps globally across Live Trace Inspector, Trace Waterfall, and Ticket Deep Dive as `HH:mm:ss.SSS` in client local time.
   - Retains full localized ISO timestamp in hover tooltip for complete date context.
   - Surfaces detected client timezone badge in the inspector header.

2. **Debug ID Token Middle Truncation and Responsive Layout Containment (AL-45)**:
   - Implements middle truncation (`ses_0192...304a`) for visible debug tokens in `.debug-id-token-group` to preserve unique trailing entropy.
   - Constrains layout with `min-width: 0`, `max-width`, and flex shrinking.

3. **Collapsible Failed Tool Calls Section with Prominent Failure Badge (AL-46)**:
   - Sets the Failed Tool Calls section to collapsed by default so the primary execution stream is immediately visible.
   - Displays a prominent summary header with failure count badge (e.g. `Failed Tool Calls: N`) and an accordion toggle button.

4. **Real-Time Live Trace Dynamic Refresh & Active Session Polling (AL-47)**:
   - Ensures active session execution traces update automatically without requiring manual page reload.
   - Employs active session polling and SSE trace update ingestion while preserving user scroll state and deliberation toggles.

5. **Interactive Multi-Metric Ticket Scatter Plot with Outlier Exclusion (AL-48)**:
   - Adds a native SVG scatter plot to the Cross-Ticket Comparison view.
   - Defaults to X = Total Spend (`cost_usd`) and Y = Net Lines of Code (`net_lines_added`).
   - Includes an outlier ticket exclusion filter allowing operators to check/uncheck tickets to dynamically rescale the chart.

6. **Reverse Event Chronology, Scroll-Lock, and Quick-Scroll Navigation (AL-49)**:
   - Reverses message list ordering so latest model turns appear at the top of the inspector stream by default.
   - Adds an order toggle allowing operators to switch back to chronological order (Oldest First).
   - Enforces pause-on-scroll lock when scrolled down.

7. **Audit and Streamline Debug Prompt Generation and Sanitization Pipeline (AL-50)**:
   - Audits `web/src/utils/debugPrompt.ts` introduced in `FIX-3` (420+ LOC).
   - Eliminates destructive regex sanitization on raw technical payloads, ensuring code fences preserve exact runtime artifacts.
   - Streamlines prompt messages and checklists across all 4 operational layers to be direct and concise.

8. **Schema v4 Evolution, Domain Invariants, and V3->V4 Migration Engine (AL-51)**:
   - Defines `SkillFact` and `TaskLinkFact` immutable domain structs.
   - Bumps `SchemaVersion` to 4 and extends `AgenticFactLedger` and `EventFact`.
   - Implements pure domain migration step `StepV3ToV4` in `internal/domain/migration/step_v3_v4.go`.

9. **OpenCode Extractor and Backfill Engine for Skills and Tasks (AL-52)**:
   - Parses `p.data` for `tool == "skill"` to extract skill name, caller role, and first invocation turn.
   - Parses `p.data` for `tool == "task"` to extract subagent type and correlate with spawned child `session_id`.

10. **Live Stream Extraction and Dual-Plane Ingestion (AL-53)**:
    - Extends `LiveService` and SQLite live tailing to parse streaming skill and task tool parts.
    - Enriches `LivePartFact` with `SkillName` and `ChildSessionID` over SSE `LiveEventTraceUpdate`.

11. **Rich Timeline Badges and Subagent Session Linking in Trace Views (AL-54)**:
    - Renders distinct violet badges for `[skill: <name>]` across Trace Waterfall and Live Inspector.
    - Renders `[task -> <role>]` with one-click navigation to the spawned child session.

12. **Dedicated Skills & Subagents Autopsy Panel in Ticket Deep Dive (AL-55)**:
    - Adds an interactive Skills Table in Ticket Deep Dive detailing skill names, caller roles, call counts, and errors.
    - Displays a Subagent Delegation Tree mapping parent-to-child session handoffs and task goals.

13. **Fleet-Wide Skill Adoption Analytics and Comparison Matrix Filter (AL-56)**:
    - Introduces `SkillUsageSummary` in domain aggregations.
    - Surfaces top invoked skills and adoption metrics in Fleet Overview.
    - Adds a "Skills Used" filter and column to the Cross-Ticket Comparison matrix.

## Ticket Breakdown
- Ready: `AL-44: Global Localized Millisecond Timestamp Formatting and Timezone Header` (Phase 7 - Ready)
- Ready: `AL-45: Debug ID Token Middle Truncation and Responsive Layout Containment` (Phase 7 - Ready)
- Ready: `AL-46: Collapsible Failed Tool Calls Section with Prominent Failure Badge and Default-Collapsed State` (Phase 7 - Ready)
- Ready: `AL-47: Real-Time Live Trace Dynamic Refresh and Active Session Polling` (Phase 7 - Ready)
- Ready: `AL-48: Interactive Multi-Metric Ticket Scatter Plot with Outlier Exclusion and Dynamic Scaling in Comparison View` (Phase 7 - Ready)
- Ready: `AL-49: Reverse Event Chronology, Scroll-Lock, and Quick-Scroll Navigation in Live Trace Inspector` (Phase 7 - Ready)
- Ready: `AL-50: Audit and Streamline Debug Prompt Generation and Sanitization Pipeline` (Phase 7 - Ready)
- Ready: `AL-51: Schema v4 Evolution, Domain Invariants, and V3->V4 Migration Engine` (Phase 8 - Ready, blocks AL-52, AL-53)
- Pending: `AL-52: OpenCode Extractor and Backfill Engine for Skills and Tasks` (Phase 8 - Pending, blocked by AL-51)
- Pending: `AL-53: Live Stream Extraction and Dual-Plane Skill and Task Ingestion` (Phase 8 - Pending, blocked by AL-51)
- Pending: `AL-54: Rich Timeline Badges and Subagent Session Linking in Trace Views` (Phase 8 - Pending, blocked by AL-52, AL-53)
- Pending: `AL-55: Dedicated Skills and Subagents Autopsy Panel in Ticket Deep Dive` (Phase 8 - Pending, blocked by AL-52)
- Pending: `AL-56: Fleet-Wide Skill Adoption Analytics and Comparison Matrix Filter` (Phase 8 - Pending, blocked by AL-52)

## Pointers
- Specification: `alspaces/PO-11/SPEC.md`
- Shared Roadmap: `alspaces/plan/TICKETS.md`
- Development Tickets:
  - `alspaces/plan/to-do/AL-44 localized timestamp and timezone header in live trace inspector.md`
  - `alspaces/plan/to-do/AL-45 debug id token overflow truncation and containment.md`
  - `alspaces/plan/to-do/AL-46 collapsible failed tool calls section in live inspector.md`
  - `alspaces/plan/to-do/AL-47 live trace dynamic refresh and active session polling.md`
  - `alspaces/plan/to-do/AL-48 multi-metric ticket scatter plot in comparison view.md`
  - `alspaces/plan/to-do/AL-49 reverse event ordering and quick scroll navigation in live inspector.md`
  - `alspaces/plan/to-do/AL-50 audit and streamline debug prompt generation and sanitization.md`
  - `alspaces/plan/to-do/AL-51 schema v4 evolution domain invariants and migration engine.md`
  - `alspaces/plan/to-do/AL-52 opencode extractor and backfill engine for skills and tasks.md`
  - `alspaces/plan/to-do/AL-53 live stream extraction and dual plane skill task ingestion.md`
  - `alspaces/plan/to-do/AL-54 rich timeline badges and subagent session linking in trace views.md`
  - `alspaces/plan/to-do/AL-55 dedicated skills and subagents autopsy panel in ticket deep dive.md`
  - `alspaces/plan/to-do/AL-56 fleet wide skill adoption analytics and comparison matrix filter.md`
- Session Summary: `alspaces/PO-11/SUMMARY.md`
  - `alspaces/plan/to-do/AL-47 live trace dynamic refresh and active session polling.md`
  - `alspaces/plan/to-do/AL-48 multi-metric ticket scatter plot in comparison view.md`
  - `alspaces/plan/to-do/AL-49 reverse event ordering and quick scroll navigation in live inspector.md`
- Session Summary: `alspaces/PO-11/SUMMARY.md`
