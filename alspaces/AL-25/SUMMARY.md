# Unified Summary: AL-25 Contextual Hover Information Icons for Cards, Tables, and Graphs

## Overview
Ticket AL-25 introduces a zero-emoji, accessible contextual information hover component (`InfoTooltip` / `InfoIcon`) and integrates it across all cards, tabular headers, longitudinal progression graphs, and execution thread tables in the AgenticLens dashboard to provide instant explanations of telemetry metrics without visual clutter, boundary clipping, or horizontal overflow.

---

## Changes by Architectural Layer

### 1. Functional Core (`domain`)
- Maintained strict FCIS isolation. Zero domain logic modifications required.
- Mathematical rollups and invariants (token ceilings, cache hit rates, cost aggregations, cumulative agent compute duration) serve as the semantic source for all explanatory tooltip copy.
- Domain types (`web/src/domain/types.ts`) and pure formatters (`web/src/domain/formatters.ts`) remained untouched.

### 2. Boundary Adapters & Infrastructure (`adapters`)
- Single binary asset embedding preserved: compiled production frontend in `web/dist/` embedded cleanly into Go binary via `backend/internal/adapters/web/dist/`.
- Updated planning ledger in `workspaces/plan/TICKETS.md` moving AL-25 to In Progress / Review.

### 3. Presentation Shell (`presentation`)
- **Core Component (`web/src/components/InfoTooltip.tsx`)**:
  - Implemented `InfoTooltip` with `InfoIcon` alias.
  - Strict union literal placement: `'top' | 'bottom' | 'left' | 'right'`.
  - Accessible inline SVG circle-i icon with `aria-hidden="true"`.
  - Keyboard accessibility: focus/blur activation, Escape key dismissal and blur.
  - ARIA attributes: `role="tooltip"`, `aria-describedby` referencing generated unique ID via `useId()`, and customizable `aria-label`.
  - Event safety: `onClick` stops event propagation (`event.stopPropagation()`) to avoid triggering parent table row navigation or sorting.
  - Defensive rendering: returns `null` when content is empty or undefined.
- **Card & KPI Integrations**:
  - `web/src/components/KPICard.tsx`: Removed native HTML `title` attribute from outer container and embedded `InfoTooltip` inside `.kpi-header` beside `.kpi-title`.
  - `web/src/components/DistributionCard.tsx`: Added optional `tooltip` prop and `.distribution-header` flex row embedding `InfoTooltip`. Defensively guarded against zero total cost division.
- **Table Integrations**:
  - `web/src/components/TicketTable.tsx`: Integrated `InfoTooltip` across all 8 column headers with `placement="bottom"`.
  - `web/src/views/LandingPage.tsx`: Added `InfoTooltip` across 4 pulse cards, 4 feature navigation cards, and all 6 Quick-Launch table headers.
  - `web/src/views/FleetOverview.tsx`: Added contextual tooltips to Role Cost Distribution and Model Cost Distribution cards.
  - `web/src/views/TicketComparisons.tsx`: Added `InfoTooltip` with `placement="bottom"` to all 15 comparison matrix column headers.
  - `web/src/views/TicketDeepDive.tsx`: Added `InfoTooltip` to Subagent Turn Velocity card header, Role Resource Consumption title (`placement="left"`) + 5 column headers (`placement="bottom"`), and Model Resource Consumption title (`placement="left"`) + 5 column headers (`placement="bottom"`).
  - `web/src/components/ThreadViewer.tsx`: Added `InfoTooltip` to session threads header and all 10 column headers of `thread-models-table`.
- **Graph & Progression Visualizations**:
  - `web/src/components/TrendChart.tsx`: Added optional `tooltip?: string` prop; renders `InfoTooltip` beside chart titles in both populated (`.trend-header`) and empty-state (`.trend-card.empty`) cards inside `.trend-title-group`.
  - `web/src/components/SaturationBar.tsx`: Added optional `tooltip?: string` prop explaining context saturation against the 200,000 token ceiling inside `.saturation-title-group`.
  - `web/src/views/MetricProgression.tsx`: Attached concrete, domain-accurate tooltip descriptions to all 5 `TrendChart` instances (Spend Trajectory, Token Volume Distribution, Cache Hit Rate Trajectory, Delivered Code Line Churn, and Peak Context Saturation vs Ceiling).
- **Metric Semantic Clarifications**:
  - Clarified distinction between Active Compute Duration (cumulative agent compute execution time across all invocations and subagents, analogous to man-hours, which can exceed calendar lead time when subagents run concurrently) and Turnaround Lead Time (wall-clock elapsed time strictly from inception to resolution).
- **Theme, Styling, Responsive Scrolling & Overflow Containment (`web/src/index.css`)**:
  - High-contrast dark theme styling with CSS custom properties (`var(--bg-surface-elevated)`, `var(--border-color)`, `var(--text-primary)`, `var(--text-muted)`).
  - Absolute positioning classes (`.placement-top`, `.placement-bottom`, `.placement-left`, `.placement-right`) with pseudo-element carets (`::after`).
  - Added `.trend-title-group`, `.table-card-header`, `.turn-velocity-title-group`, `.thread-title-group`, `.th-content`, and flex alignments for `.num-col`.
  - Eliminated container clipping: changed `.table-card` and `.thread-accordion-item` to `overflow: visible;` so tooltip bubbles float freely without being truncated at card borders.
  - Added border radius management to `.thread-header-btn` (`6px`, open: `6px 6px 0 0`) and `.thread-body` (`0 0 6px 6px`).
  - Defined global `.table-responsive` class (`width: 100%; overflow-x: auto; -webkit-overflow-scrolling: touch;`) wrapping `.data-table` and `.thread-models-table`.
  - Enforced defensive sizing: `.data-table` (`min-width: 480px;`, `white-space: nowrap;` on cells) and `.thread-models-table` (`min-width: 850px;`, `white-space: nowrap;` on cells) ensuring tables establish smooth horizontal scrollbars rather than spilling out of bounds on constrained viewports.
  - Isolated coordinate spaces: card headers remain outside `.table-responsive` with unclipped `placement="left"`, while table column headers remain inside with `placement="bottom"` scrolling synchronously with table contents.
- **Automated Tests**:
  - Comprehensive unit, interaction, boundary, and zero-emoji tests in `web/src/components/Components.test.tsx` (46 tests).
  - Event propagation and table sort preservation tests in `web/src/components/TicketTable.test.tsx` (14 tests).
  - Deep-dive view tests in `web/src/views/TicketDeepDive.test.tsx` (18 tests verifying `placement-left` headers, `placement-bottom` columns, `.table-responsive` wrapping, and static CSS rules).
  - Longitudinal trend tests in `web/src/views/MetricProgression.test.tsx` (11 tests).

---

## Verification Status

- `make check`: PASS (`go vet`, TypeScript `tsc --noEmit`, ESLint 0 warnings/errors, `git diff --check`).
- `make test`: PASS (Go backend race detector tests, Vitest 15 test files / 259 tests passing with zero silent skips).
- `make build`: PASS (Single binary built at `bin/agentic-lens` with embedded web assets).
- Zero Emojis: PASS (Verified across all modified files, CSS, markup, and tests).

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 4
- **Active Agent Duration:** 3h 50m (excludes user wait time)
- **Context Window Health:** ELEVATED — Root agent peak 168.8k tokens (84% of 200k healthy ceiling, 2 compactions) | Subagents peak 175.1k (Coder)
- **Role Effort Distribution (Subagents):** Planner 14.7%, Coder 45.1%, Tester 27.0%, Reviewer 13.2%
- **Subagent Turn Velocity & Thrashing:** 43.7 avg turns/run (peak: 81 turns) | Planner 37.3, Coder 62.2, Tester 52.5, Reviewer 23.4
- **Test-to-Production Code Ratio:** 0.6:1 (+473 test lines / +728 prod lines)
- **Code Delivered:** +1,202 / -172 lines across 17 code files (excluding docs)
- **Reasoning Effort:** 259,320 reasoning tokens (56.8% of generated output)
- **Tool Health & Self-Correction:** 0 tool calls (0 self-corrections, 100.0% success rate)

### Anomalies & Health Flags
- **Status:** 15 execution anomalies detected (10 CRITICAL, 5 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (161.1k tokens) — Context tokens reached 161.1k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (164.6k tokens) — Context tokens reached 164.6k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+95.9k tokens) — Context tokens surged by 95.9k to 161.2k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (161.2k tokens) — Context tokens reached 161.2k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (168.8k tokens) — Context tokens reached 168.8k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (166.2k tokens) — Context tokens reached 166.2k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (175.1k tokens) — Context tokens reached 175.1k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (168.8k tokens) — Context tokens reached 168.8k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (166.2k tokens) — Context tokens reached 166.2k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (175.1k tokens) — Context tokens reached 175.1k, breaching the 160k critical ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+51.8k tokens) — Context tokens surged by 51.8k to 74.2k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+57.9k tokens) — Context tokens surged by 57.9k to 77.2k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+47.2k tokens) — Context tokens surged by 47.2k to 47.2k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+56.5k tokens) — Context tokens surged by 56.5k to 56.5k tokens.
- [WARNING] **TOOL_RETRY_STREAK**: Tool retry streak: mcp-gateway_mcp-exec (3 consecutive errors) — Tool 'mcp-gateway_mcp-exec' encountered 3 consecutive execution failures.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **orchestrator** | 1 | 4,819,330 | 60,928 | 16,471,228 | 77.4% | 168.8k | ELEVATED | $5.23 |
| **planner** | 4 | 875,648 | 30,444 | 5,392,808 | 86.0% | 83.8k | HEALTHY | $1.36 |
| **coder** | 5 | 2,816,988 | 55,515 | 20,858,173 | 88.1% | 175.1k | ELEVATED | $4.18 |
| **tester** | 4 | 1,336,789 | 30,085 | 15,754,485 | 92.2% | 157.2k | ELEVATED | $2.50 |
| **reviewer** | 5 | 897,592 | 19,940 | 4,580,461 | 83.6% | 84.7k | HEALTHY | $1.23 |
| **Total** | **19** | **10,746,347** | **196,912** | **63,057,155** | **85.4%** | **175.1k** | **ELEVATED** | **$14.50** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gemini-3.8-flash** | 19 | 10,746,347 | 196,912 | 63,057,155 | 85.4% | $14.50 |
| **Total** | **19** | **10,746,347** | **196,912** | **63,057,155** | **85.4%** | **$14.50** |
