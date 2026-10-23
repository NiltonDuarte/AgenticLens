# docs(plan): live observability refinements, comparison scatter plot, and agent/skill tracking (PO-11)

## Summary
Defines the product specification, advisor risk analysis, probing questions, resolved steering decisions, and 12 vertical tracer-bullet Jira tickets (AL-44 through AL-55) covering live observability inspector refinements, cross-ticket correlation visualization, debug prompt streamlining, and complete agent and skill invocation observability:
1. Global localized millisecond timestamp formatting (`HH:mm:ss.SSS`) across Live Inspector, Waterfall, and Deep Dive views, with header timezone indicator (AL-44).
2. Monospace debug ID token middle truncation (`ses_0192...304a`) and flex container overflow containment (AL-45).
3. Collapsible Failed Tool Calls section with default-collapsed state and prominent failure count badge (AL-46).
4. Real-time live trace dynamic refresh and active session polling in the right inspector pane (AL-47).
5. Interactive multi-metric ticket scatter plot in the Cross-Ticket Comparison view with default X=Cost, Y=Net Lines of Code, outlier exclusion filter, and dynamic linear/log scaling (AL-48).
6. Reverse event chronology (newest first), chronological order toggle, pause-on-scroll lock, and quick-scroll navigation in the Live Trace Inspector (AL-49).
7. Audit and streamline debug prompt generation and eliminate destructive payload sanitization in `web/src/utils/debugPrompt.ts` (AL-50).
8. Additive domain contracts, OpenCode SQLite extractor, and live stream ingestion for modular skills and subagent task linking (AL-51).
9. Rich timeline badges (`[skill: name]`, `[task -> role]`) and clickable subagent session jump links in Trace Waterfall and Live Inspector (AL-52).
10. Dedicated Skills & Subagents autopsy panel in Ticket Deep Dive (AL-53).
11. Fleet-wide skill adoption analytics and multi-select comparison matrix filter (AL-54).
12. Formal Schema v4 evolution, domain migration step (`StepV3ToV4`), and CLI backfill runner (AL-55).

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

8. **Additive Domain Contracts, OpenCode Extractor, and Live Stream Ingestion for Skills and Tasks (AL-51)**:
   - Defines `SkillFact` and `TaskLinkFact` immutable domain structs with `omitempty` JSON tags.
   - Extends `LivePartFact` in Go and TypeScript with optional `SkillName`, `ChildSessionID`, and `SubagentType`.
   - Maintains `CurrentSchemaVersion = 3` initially to permit rapid UI iteration without premature migration churn.
   - Parses `tool == "skill"` and `tool == "task"` in both offline OpenCode extractor and real-time live SQLite watcher.

9. **Rich Timeline Badges and Subagent Session Linking in Trace Views (AL-52)**:
   - Renders distinct badges for `[skill: <name>]` across Trace Waterfall and Live Inspector.
   - Renders `[task -> <role>]` with one-click navigation to the spawned child session.

10. **Dedicated Skills & Subagents Autopsy Panel in Ticket Deep Dive (AL-53)**:
    - Adds an interactive Skills Table in Ticket Deep Dive detailing skill names, caller roles, call counts, and errors.
    - Displays a Subagent Delegation Tree mapping parent-to-child session handoffs and task goals.
    - Gracefully handles legacy un-migrated ledgers with empty states.

11. **Fleet-Wide Skill Adoption Analytics and Comparison Matrix Filter (AL-54)**:
    - Introduces `SkillAdoptionFact` and pure aggregation in domain.
    - Surfaces top invoked skills and adoption metrics in Fleet Overview.
    - Adds a "Skills Used" filter and column to the Cross-Ticket Comparison matrix.

12. **Formal Schema v4 Evolution, Migration Step, and Backfill Runner (AL-55)**:
    - Formally bumps `CurrentSchemaVersion` from 3 to 4 as the final stabilization step.
    - Implements `StepV3ToV4` in Go and TypeScript ensuring non-null array normalization.
    - Enhances `agentic-lens backfill` CLI to upgrade historical workspace ledgers.

## Ticket Breakdown
- Ready: `AL-44: Global Localized Millisecond Timestamp Formatting and Timezone Header` (Phase 7 - Ready)
- Ready: `AL-45: Debug ID Token Middle Truncation and Responsive Layout Containment` (Phase 7 - Ready)
- Ready: `AL-46: Collapsible Failed Tool Calls Section with Prominent Failure Badge and Default-Collapsed State` (Phase 7 - Ready)
- Ready: `AL-47: Real-Time Live Trace Dynamic Refresh and Active Session Polling` (Phase 7 - Ready)
- Ready: `AL-48: Interactive Multi-Metric Ticket Scatter Plot with Outlier Exclusion and Dynamic Scaling in Comparison View` (Phase 7 - Ready)
- Ready: `AL-49: Reverse Event Chronology, Scroll-Lock, and Quick-Scroll Navigation in Live Trace Inspector` (Phase 7 - Ready)
- Ready: `AL-50: Audit and Streamline Debug Prompt Generation and Sanitization Pipeline` (Phase 7 - Ready)
- Ready: `AL-51: Additive Domain Contracts, OpenCode Extractor, and Live Stream Ingestion for Skills and Tasks` (Phase 8 - Ready, blocks AL-52, AL-53, AL-54)
- Pending: `AL-52: Rich Timeline Badges and Subagent Session Linking in Trace Views` (Phase 8 - Pending, blocked by AL-51)
- Pending: `AL-53: Dedicated Skills and Subagents Autopsy Panel in Ticket Deep Dive` (Phase 8 - Pending, blocked by AL-51)
- Pending: `AL-54: Fleet-Wide Skill Adoption Analytics and Comparison Matrix Filter` (Phase 8 - Pending, blocked by AL-51)
- Pending: `AL-55: Formal Schema v4 Evolution, Migration Step, and Backfill Runner` (Phase 8 - Pending, blocked by AL-51, AL-52, AL-53, AL-54)

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
  - `alspaces/plan/to-do/AL-51 additive domain contracts opencode extractor and live stream ingestion for skills and tasks.md`
  - `alspaces/plan/to-do/AL-52 rich timeline badges and subagent session linking in trace views.md`
  - `alspaces/plan/to-do/AL-53 dedicated skills and subagents autopsy panel in ticket deep dive.md`
  - `alspaces/plan/to-do/AL-54 fleet wide skill adoption analytics and comparison matrix filter.md`
  - `alspaces/plan/to-do/AL-55 formal schema v4 evolution migration step and backfill runner.md`
- Session Summary: `alspaces/PO-11/SUMMARY.md`
