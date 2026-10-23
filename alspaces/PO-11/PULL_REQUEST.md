# docs(plan): live observability refinements and cross-ticket scatter comparison (PO-11)

## Summary
Defines the product specification, advisor risk analysis, probing questions, resolved steering decisions, and 7 vertical tracer-bullet Jira tickets (AL-44 through AL-50) covering live observability inspector refinements, cross-ticket correlation visualization, and debug prompt generation streamlining:
1. Global localized millisecond timestamp formatting (`HH:mm:ss.SSS`) across Live Inspector, Waterfall, and Deep Dive views, with header timezone indicator (AL-44).
2. Monospace debug ID token middle truncation (`ses_0192...304a`) and flex container overflow containment (AL-45).
3. Collapsible Failed Tool Calls section with default-collapsed state and prominent failure count badge (AL-46).
4. Real-time live trace dynamic refresh and active session polling in the right inspector pane (AL-47).
5. Interactive multi-metric ticket scatter plot in the Cross-Ticket Comparison view with default X=Cost, Y=Net Lines of Code, outlier exclusion filter, and dynamic linear/log scaling (AL-48).
6. Reverse event chronology (newest first), chronological order toggle, pause-on-scroll lock, and quick-scroll navigation in the Live Trace Inspector (AL-49).
7. Audit and streamline debug prompt generation and eliminate destructive payload sanitization in `web/src/utils/debugPrompt.ts` (AL-50).

## Initiatives Covered

1. **Global Localized Millisecond Timestamp Formatting and Timezone Header (AL-44)**:
   - Formats event and turn timestamps globally across Live Trace Inspector, Trace Waterfall, and Ticket Deep Dive as `HH:mm:ss.SSS` in client local time to save space and eliminate manual UTC offset calculation.
   - Retains full localized ISO timestamp in hover tooltip for complete date context.
   - Surfaces detected client timezone badge in the inspector header.

2. **Debug ID Token Middle Truncation and Responsive Layout Containment (AL-45)**:
   - Implements middle truncation (`ses_0192...304a`) for visible debug tokens (`span.debug-id-token.mono`) in `.debug-id-token-group` to preserve unique trailing entropy.
   - Constrains layout with `min-width: 0`, `max-width`, and flex shrinking.
   - Preserves complete identifier strings in tooltip titles and clipboard copy actions.
   - Compatible with forthcoming merge of `fix/FIX-3-centralize-live-anomalies-debug-identifiers`.

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
   - Supports linear and logarithmic scale modes depending on metric distribution span.
   - Tooltips show autopsy facts; clicking points routes to Ticket Deep Dive.

6. **Reverse Event Chronology, Scroll-Lock, and Quick-Scroll Navigation (AL-49)**:
   - Reverses message list ordering so latest model turns appear at the top of the inspector stream by default.
   - Adds an order toggle allowing operators to switch back to chronological order (Oldest First).
   - Enforces pause-on-scroll lock: when scrolled down, incoming streaming events do not displace the user's reading position.
   - Provides dedicated "Scroll to Top" and "Scroll to Bottom" navigation buttons.

7. **Audit and Streamline Debug Prompt Generation and Sanitization Pipeline (AL-50)**:
   - Audits `web/src/utils/debugPrompt.ts` introduced in `FIX-3` (420+ LOC).
   - Eliminates destructive regex sanitization on raw technical payloads (tool inputs, outputs, errors, messages, deliberation), ensuring code fences preserve exact runtime artifacts.
   - Removes fragile regex grammar post-processing (cleaning orphaned commas and conjunctions).
   - Streamlines prompt messages and checklists across all 4 operational layers to be direct, sharp, and concise.

## Ticket Breakdown
- Ready: `AL-44: Global Localized Millisecond Timestamp Formatting and Timezone Header` (Phase 7 - Ready)
- Ready: `AL-45: Debug ID Token Middle Truncation and Responsive Layout Containment` (Phase 7 - Ready)
- Ready: `AL-46: Collapsible Failed Tool Calls Section with Prominent Failure Badge and Default-Collapsed State` (Phase 7 - Ready)
- Ready: `AL-47: Real-Time Live Trace Dynamic Refresh and Active Session Polling` (Phase 7 - Ready)
- Ready: `AL-48: Interactive Multi-Metric Ticket Scatter Plot with Outlier Exclusion and Dynamic Scaling in Comparison View` (Phase 7 - Ready)
- Ready: `AL-49: Reverse Event Chronology, Scroll-Lock, and Quick-Scroll Navigation in Live Trace Inspector` (Phase 7 - Ready)
- Ready: `AL-50: Audit and Streamline Debug Prompt Generation and Sanitization Pipeline` (Phase 7 - Ready)

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
  - `alspaces/plan/to-do/AL-47 live trace dynamic refresh and active session polling.md`
  - `alspaces/plan/to-do/AL-48 multi-metric ticket scatter plot in comparison view.md`
  - `alspaces/plan/to-do/AL-49 reverse event ordering and quick scroll navigation in live inspector.md`
- Session Summary: `alspaces/PO-11/SUMMARY.md`
