# Feature Specification: Live Observability Refinements, Comparison Scatter Plot, and Agent & Skill Invocation Tracking (AL-44 through AL-56)

## Problem Statement

During active multi-agent development and retrospective review, operators face eight distinct friction points across the Live Observability inspector, Cross-Ticket Comparison matrix, and agent runtime telemetry:

1. **Unformatted Timestamp Spatial Congestion**:
   In `LiveTraceInspector.tsx`, message turns render raw ISO-8601 UTC strings (`2026-10-22T14:32:10.123Z`). At 24+ characters, this string clutters `.message-turn-meta`, pushing token chips onto secondary lines. Operators debugging in local time must manually compute timezone offsets, and the inspector header lacks timezone context.

2. **Debug ID Token Flex Overflow**:
   Debug tokens (`span.debug-id-token.mono`) containing session UUIDs (`ses_...`) or message IDs overflow their container bounds in `.debug-id-token-group`. On medium displays or narrowed split panes, overflowing monospace text breaks layout boundaries and overlaps adjacent action buttons.

3. **Failed Tool Cards Pushing Execution Stream Below Fold**:
   In sessions encountering tool failures, `section.inspector-failed-tools-section` renders all `FailedToolAlertCard` components expanded by default. A streak of 4 or 5 failed calls forces the primary message stream far down the scroll view, requiring excessive scrolling to see what the agent is actually saying.

4. **Stale Live Trace Inspector (Missing Real-Time Trace Streaming)**:
   The right inspector pane does not react to live execution events. While `LiveService` maintains an SSE stream for `session_tree` and `anomaly_alert`, `LiveEventTraceUpdate` is defined but never broadcast. Furthermore, `LiveObservability.tsx` only fetches traces upon initial session selection or manual page reload. Operators watching an active run see a static message list until they manually refresh.

5. **Lack of Multi-Dimensional Correlation in Cross-Ticket Benchmarking**:
   `TicketComparisons.tsx` provides tabular comparisons and high-level KPI cards, but lacks an interactive visual correlation tool. Operators cannot visually plot relationships between metrics (e.g. Iterations vs Spend, Cache Hit Rate vs Cost, or Active Duration vs Churn) to identify outlier tickets or systemic regressions across workspaces.

6. **Chronological Inversion Friction in Real-Time Monitoring**:
   Message turns in the inspector list are ordered chronologically (oldest at top, newest at bottom). When monitoring an active run with dozens of turns, operators must continually scroll to the bottom to see current deliberation and tool output. When reviewing lengthy traces, there are no quick-scroll controls to jump between initial inception and current state.

7. **Over-Engineered Debug Prompt Sanitization and Verbose Messaging (FIX-3)**:
   In `web/src/utils/debugPrompt.ts` (added in `fix/FIX-3-centralize-live-anomalies-debug-identifiers`), 420+ LOC performs aggressive multi-pass regex transformations across raw tool payloads, stripping tickets and branch names from actual shell commands, test runs, and compiler error logs. Furthermore, the generated debug messages include repetitive provenance sentences ("This data was extracted from the OpenCode database.") and overly verbose, rhetorical checklists.

8. **Blindness Toward Modular Skills and Autonomous Subagent Invocations**:
   OpenCode runtime primitives for modular skills (`tool == "skill"`) and subagent delegations (`tool == "task"`) are currently treated as generic, opaque tools (`ToolFact{Name: "skill", Calls: 3}`). Telemetry discards skill names, caller roles, invocation turns, and the explicit links between parent `task` calls and spawned child `session_id`s. In the UI, trace views lack contextual skill badges and one-click subagent jump links, Ticket Deep Dive has zero skill autopsy visibility, and Fleet Overview lacks skill adoption metrics.

---

## Target Personas & Use Cases

- **AI System Engineers & Agent Debuggers**:
  Need real-time visual feedback in the live trace inspector without manual reloads. Require concise, localized timestamps to correlate agent actions with local development logs. Need to trace subagent task delegations and identify context impact from loaded skills.
- **Platform Operators & Autonomous Overseers**:
  Need high-level oversight of active sessions without being blinded by massive expanded error blocks. Want newest model turns surfaced immediately at the top of the stream. Need to track skill adoption and prevent redundant skill thrashing.
- **Engineering Leads & Performance Analysts**:
  Need flexible multi-metric scatter plots to benchmark agent efficiency across diverse tickets, correlating cost, iterations, and context saturation. Need fleet-wide analytics on skill adoption and workflow efficacy.

---

## User Stories

1. As an operator debugging an active agent session, I want message timestamps formatted as local `HH:mm:ss.SSS` with an active timezone indicator in the inspector header, so that I can instantly correlate agent turns with local system logs while saving horizontal space.
2. As a dashboard viewer on varying viewport widths, I want debug ID tokens constrained and truncated with ellipsis while preserving copy actions, so that the inspector header never overflows or breaks flex layout.
3. As an operator inspecting a session with tool failures, I want the Failed Tool Calls section collapsed by default with an error count badge, so that I can immediately view the message stream while having one-click access to failure details.
4. As an operator watching an ongoing agent run, I want the right-hand trace inspector to dynamically update when new turns and tool calls occur, so that I never have to refresh the browser to see live progress.
5. As an engineering manager comparing ticket performances, I want an interactive scatter plot with customizable X and Y axes where each point represents a ticket, so that I can visually analyze cost-to-iteration efficiency and spot outliers.
6. As a developer reviewing deep session traces, I want recent events displayed at the top of the inspector with quick-jump buttons to the top and bottom of the stream, so that I can immediately inspect the latest model activity without scrolling.
7. As an operator reviewing a ticket autopsy or live trace, I want to see which modular skills were loaded, by which agent, and at what turn, with direct jump links from parent tasks to child subagents.
8. As an engineering lead, I want fleet-wide skill adoption analytics and a filter in Ticket Comparisons to evaluate skill efficacy.

---

## Architectural Breakdown: Thirteen Vertical Slices (AL-44 through AL-56)

### AL-44: Global Localized Millisecond Timestamp Formatting and Timezone Header
- **Domain / Formatters**: Add `formatLocalTimeWithMs(isoStr: string): string` and `getClientTimezoneBadge(): string` in `web/src/domain/formatters.ts`.
- **Global Presentation Scope**:
  - Live Trace Inspector (`LiveTraceInspector.tsx`): Render localized `HH:mm:ss.SSS` in `.turn-timestamp` with full localized ISO datetime in `title` tooltip. Display active client timezone badge in `.inspector-title-group`.
  - Trace Waterfall (`TraceWaterfall.tsx`): Localize timeline event timestamps to `HH:mm:ss.SSS` with hover tooltip.
  - Ticket Deep Dive / Autopsy (`TicketDeepDive.tsx`): Format event timestamps and turn indicators in local time with milliseconds.
- **Date Rollover Decision**: Since this monitors live session data, operators understand natural day rollovers. The tooltip maintains the complete ISO string for date verification.
- **Testing**: Unit tests in `formatters.test.ts`, `LiveTraceInspector.test.tsx`, and `TraceWaterfall.test.tsx` verifying localized millisecond formatting, zero-padding, and tooltip preservation.

### AL-45: Debug ID Token Overflow Truncation and Responsive Layout in Message Header
- **Middle Truncation Utility**: Implement `truncateMiddle(str: string, maxLeading?: number, maxTrailing?: number): string` in `web/src/domain/formatters.ts` (e.g. `ses_0192...304a`), ensuring entropy preservation across similar prefixes.
- **Branch Context**: Accounts for identifiers introduced in branch `fix/FIX-3-centralize-live-anomalies-debug-identifiers` (merging soon).
- **Presentation**: Constrain `.debug-id-token-group` and `.debug-id-token.mono` in `web/src/index.css` with `max-width`, `min-width: 0`, `overflow: hidden`, `text-overflow: ellipsis`, and `white-space: nowrap`.
- **Integrity**: Full uncut ID string is preserved in the `title` tooltip and copied intact by the prompt copy button.
- **Testing**: Responsive styling and middle-truncation tests in `LiveTraceInspector.test.tsx`.

### AL-46: Collapsible Failed Tool Calls Section with Default-Collapsed State in Live Inspector
- **Presentation**: Refactor `section.inspector-failed-tools-section` in `LiveTraceInspector.tsx` to include an accordion toggle button (`data-testid="failed-tools-toggle-btn"`).
- **Default State**: Section is collapsed by default, displaying a prominent summary header with failure count, warning/critical severity badge, and an expand button: `Failed Tool Calls (N)`.
- **Expanded State**: Revealing the list of `FailedToolAlertCard` components upon user click.
- **Testing**: Tests in `LiveTraceInspector.test.tsx` verifying collapsed by default, prominent count badge rendering, toggle interaction, and keyboard accessibility.

### AL-47: Real-Time Live Trace Ingestion and Dynamic Inspector Refresh
- **Architecture**:
  - In `backend/internal/adapters/live/service.go`, ensure active session trace changes broadcast `LiveEventTraceUpdate` or update `LiveSessionTree` node timestamps reliably.
  - In `web/src/views/LiveObservability.tsx`, dynamically trigger `fetchLiveTrace(selectedSessionId)` when an active session's updated timestamp advances or via active polling interval (2000ms) while `selectedSessionId` is actively running.
  - Maintain user scroll position and reasoning toggle states during incremental trace re-renders.
- **Testing**: Backend test in `live_handlers_test.go` and frontend async tests in `LiveObservability.test.tsx` verifying live trace update propagation without manual reload.

### AL-48: Interactive Multi-Metric Ticket Scatter Plot in Comparison Matrix View
- **Presentation**: Build `web/src/components/TicketScatterPlot.tsx` using native SVG (zero external charting libraries).
- **Metric Controls**: Dropdown selectors for X-Axis and Y-Axis with defaults:
  - Default X-Axis: `cost_usd` (Total spend in USD)
  - Default Y-Axis: `net_lines_added` (Net lines of code delivered)
- **Outlier Ticket Exclusion Filter**: Multi-select or checklist filter listing all tickets in current comparison set, allowing operators to manually exclude outlier tickets from the scatter plot.
- **Dynamic Scale & Log Scaling**:
  - Chart automatically recalculates X and Y axes domain scales based strictly on non-excluded tickets.
  - Scale selector allowing automatic linear vs logarithmic scale based on span, or explicit user toggle.
- **Plot Interactions**:
  - Point radius and color based on ticket status (`completed` vs `in-progress`) or prefix.
  - Interactive tooltip on hover showing ticket ID, title, and formatted X/Y values.
  - On-click handler invoking `onSelectTicket(ticketId)` to navigate to deep dive autopsy.
- **Testing**: Comprehensive tests in `TicketScatterPlot.test.tsx` and integration into `TicketComparisons.test.tsx`.

### AL-49: Reverse Event Chronology, Scroll-Lock, and Quick-Scroll Navigation in Live Inspector
- **Presentation**:
  - Reverse order of message stream in `LiveTraceInspector.tsx` so the most recent turn appears at the top of `.inspector-messages-list` (newest first) by default.
  - Add explicit order toggle button (`data-testid="turn-order-toggle-btn"`) allowing operators to toggle back to chronological order (Latest First vs Oldest First).
  - Implement pause-on-scroll lock: when an operator has scrolled down into earlier turns, incoming streaming turns are paused or inserted without displacing the current scroll position.
  - Add "Scroll to Top" and "Scroll to Bottom" quick-jump buttons in the inspector controls.
- **Testing**: Tests in `LiveTraceInspector.test.tsx` asserting descending turn ordering by default, toggle back to chronological order, scroll-lock behavior, and scroll button handlers.

### AL-50: Audit and Streamline Debug Prompt Generation and Sanitization Pipeline
- **Problem Context**:
  - `web/src/utils/debugPrompt.ts` (added in `FIX-3`) spans 420+ LOC with aggressive, multi-pass regex sanitization (`stripInternalIdentifiers`, `cleanSeverityTiers`, `sanitizeTextForPrompt`) applied indiscriminately to technical payloads.
  - Applying text replacement to raw tool inputs, outputs, error logs, and message content corrupts legitimate commands (`git checkout feat/...` stripped of target branch), compiler errors, and test names.
  - Debug messages and checklists are bloated with repetitive boilerplate ("This data was extracted from the OpenCode database.") and verbose rhetorical questions.
- **Streamlining & Architecture**:
  - Pass technical payloads (`input`, `output`, `errorMessage`, `reasoningContent`, `messageContent`) verbatim into defensive Markdown code fences without destructive regex mutation.
  - Eliminate fragile grammatical repair regexes (cleaning orphaned commas, dashes, and conjunctions).
  - Streamline prompt templates across the 4 layers (Tool Execution, Execution Issue, Message Turn, Session Overview) to be concise, direct, and actionable.
  - Reduce module size to under 250 LOC (down from 420+ LOC).
- **Testing**:
  - Update `web/src/utils/debugPrompt.test.ts` to assert that raw code payloads containing ticket prefixes, branch names, and severity tags are preserved intact within code blocks, while diagnostic checklists remain sharp and concise.

### AL-51: Additive Domain Contracts, OpenCode Extractor, and Live Stream Ingestion for Skills and Tasks
- **Target**: `internal/domain`, `web/src/types.ts`, `internal/adapters/opencode`, `internal/adapters/live`, `web/src/domain/live.ts`, `backend/internal/adapters/http`, tests
- **Additive Domain Invariants**:
  - Define `SkillFact` and `TaskLinkFact` structs with `omitempty` JSON tags.
  - Add optional `Skills []SkillFact` and `TaskLinks []TaskLinkFact` to `AgenticFactLedger` without bumping `CurrentSchemaVersion` (maintains v3).
  - Add optional `skills?: SkillFact[]` and `task_links?: TaskLinkFact[]` to TypeScript `AgenticFactLedger`.
  - Extend `LivePartFact` in Go and TypeScript with optional `SkillName`, `ChildSessionID`, and `SubagentType`.
  - Maintain defensive fallbacks (`ledger.skills ?? []`) across all frontend code.
- **Dual-Plane Ingestion Logic**:
  - In `internal/adapters/opencode/extractor.go`, parse `tool == "skill"` parts: extract skill name, turn, caller role, and success state.
  - Parse `tool == "task"` parts: extract subagent type, description, parent turn, and link child session IDs.
  - In `internal/adapters/live/watcher.go`, parse skill names and child session targets from streaming tool parts during active session SQLite tailing.
  - Broadcast enriched parts over SSE event stream (`/api/live/stream`) and trace payload endpoints (`/api/live/traces/:id`).
- **Testing**: Pure domain serialization tests, OpenCode extractor tests, and live watcher SSE streaming tests with synthetic SQLite fixtures.

### AL-52: Rich Timeline Badges and Subagent Session Linking in Trace Views
- **Target**: `web/src/components/TraceWaterfall.tsx`, `web/src/components/LiveTraceInspector.tsx`, CSS, tests
- **Presentation**:
  - In `TraceWaterfall.tsx` and `LiveTraceInspector.tsx`, render skill invocations with distinct badge: `[skill: <name>]`.
  - Render subagent task delegations with badge: `[task -> <subagent_type>]`.
  - Make `[task -> ...]` badges interactive: clicking navigates directly to the child session timeline in the live tree or deep dive.
  - Prevent horizontal layout blowout using text truncation and tooltips.
- **Testing**: Component tests in `TraceWaterfall.test.tsx` and `LiveTraceInspector.test.tsx` verifying badge formatting and jump navigation.

### AL-53: Dedicated Skills and Subagents Autopsy Panel in Ticket Deep Dive
- **Target**: `web/src/views/TicketDeepDive.tsx`, `web/src/components/`, tests
- **Presentation**:
  - Add a dedicated "Skills & Subagents" section in Ticket Deep Dive below Role Effort Distribution.
  - Defensive fallback: if `skills` or `task_links` are undefined/null (legacy ledgers), display an elegant empty state without console warnings.
  - Render an interactive Skills Table listing: Skill Name, Loaded By Role, Invocations Count, Error Count, and First Turn Loaded.
  - Render a Subagent Delegation Tree showing parent-child handoffs, task prompts, duration, cost, and child session jump links.
- **Testing**: Vitest tests in `TicketDeepDive.test.tsx` verifying rendering of skills table, subagent tree, and empty states.

### AL-54: Fleet-Wide Skill Adoption Analytics and Comparison Matrix Filter
- **Target**: `internal/domain/aggregations.go`, `web/src/views/FleetOverview.tsx`, `web/src/views/TicketComparisons.tsx`, tests
- **Domain Aggregations**:
  - Add `SkillAdoptionFact` and `ComputeSkillAdoption(ledgers []AgenticFactLedger)` to `backend/internal/domain/aggregations.go`.
  - Gracefully handle legacy ledgers without skills data.
- **Presentation**:
  - In `FleetOverview.tsx`, add a "Skill Adoption" card showing top invoked skills and fleet-wide adoption rates.
  - In `TicketComparisons.tsx`, add a "Skills Used" column and a multi-select filter to slice tickets by loaded skill.
- **Testing**: Domain tests in `aggregations_test.go` and view tests in `FleetOverview.test.tsx` and `TicketComparisons.test.tsx`.

### AL-55: Formal Schema v4 Evolution, Migration Step, and Backfill Runner
- **Target**: `internal/domain/models.go`, `internal/domain/migration/runner.go`, `web/src/domain/migration.ts`, `internal/adapters/backfill`, tests
- **Hardening and Finalization**:
  - Once the data shape and UI contracts are stabilized by AL-51 through AL-54, formally bump `CurrentSchemaVersion` from 3 to 4.
  - Implement `StepV3ToV4` in Go and TypeScript to normalize legacy ledgers (initializing `skills: []` and `task_links: []`).
  - Update `agentic-lens backfill` CLI to upgrade historical workspaces to Schema v4 and retroactively extract skill/task facts from raw OpenCode session logs.
- **Testing**: Table-driven migration tests in Go and Vitest, end-to-end backfill CLI tests.

---

## Stakeholder Steering Decisions & Resolved Invariants

1. **Date Rollover (AL-44)**:
   - *Decision*: Operators understand day rollover for live data. Standard local `HH:mm:ss.SSS` is used with the full ISO string retained in hover tooltip for rollover disambiguation.
2. **ID Truncation Entropy (AL-45)**:
   - *Decision*: Use middle truncation (`ses_0192...304a`) to preserve unique trailing entropy. FIX-3 will merge soon.
3. **Critical Error Visibility (AL-46)**:
   - *Decision*: Render prominent error badge and exact failure count on the collapsed header so failures remain immediately obvious.
4. **Scroll Disorientation with Inverse Chronology (AL-49)**:
   - *Decision*: Implement a pause-on-scroll lock when scrolled down so prepending incoming events does not displace the viewport.
5. **Scatter Plot Outlier Clustering (AL-48)**:
   - *Decision*: Add a ticket selection filter to allow operators to manually remove outlier tickets. Dynamically update graph scale. Provide linear vs logarithmic scaling depending on data span.
6. **Timestamp Presentation Scope (AL-44)**:
   - *Decision*: Apply localized millisecond timestamp formatting globally across the Live Trace Inspector, Trace Waterfall, and Ticket Deep Dive.
7. **Default Scatter Metrics (AL-48)**:
   - *Decision*: Default to X = Total Cost (`cost_usd`), Y = Net Lines of Code (`net_lines_added`).
8. **Ordering Preference (AL-49)**:
   - *Decision*: Reverse chronological order by default, with an interactive toggle allowing operators to switch back to chronological order.
9. **Debug Prompt Sanitization & Messaging (AL-50)**:
   - *Decision*: Technical payloads inside code blocks must never be regex-mangled. Prompts must be concise, direct, and free of redundant boilerplate sentences. Stakeholder review of the 4 prompt layers is incorporated into AL-50.
10. **Scope of Invocations (AL-51 to AL-55)**:
    - *Decision*: Unified tracking of both modular skills and subagent task delegations.
11. **Consolidated Dual-Plane Ingestion (AL-51)**:
    - *Decision*: Additive domain models, OpenCode batch extractor, and live stream SQLite watcher are consolidated into an atomic backend slice (AL-51).
12. **UI Presentation Scope (AL-52 to AL-54)**:
    - *Decision*: Full visibility across Live Inspector & Waterfall (badges & jump links, AL-52), Ticket Deep Dive (dedicated autopsy panel, AL-53), and Fleet Overview & Comparisons (skill adoption analytics, AL-54).
13. **Additive-First Evolution & Migration-Last Strategy (AL-51, AL-55)**:
    - *Decision*: Implement domain fields and dual-plane extraction additively first under Schema v3 (`omitempty`), build and steer UI features, and formalize Schema v4 and `StepV3ToV4` backfill as the final hardening step (AL-55). This avoids premature migration churn while preserving full backward compatibility.
