# Work Summary: AL-9 Visual Trace Waterfall and Anomaly Inspector

## Implementation Overview by Layer

### 1. Domain Layer
- `web/src/types/telemetry.ts`: Canonical contracts for `EventFact`, `EventType`, `Anomaly`, `AnomalySeverity`, `AnomalyRuleID`, `AnomalyContext`, `BrittleToolSummary`, and `FleetAnomalySummary`.
- `web/src/domain/types.ts`: Re-exported and consolidated telemetry contracts across dashboard models with strict `readonly` modifiers, zero `any`, and full `exactOptionalPropertyTypes: true` compliance.

### 2. Adapters Layer
- `backend/internal/adapters/opencode/extractor.go`: Raised error message extraction and excerpt truncation threshold from 200 characters to 4096 characters (`errMsg[:4096]`), capturing full error payloads, stack traces, and failure diagnostics without premature truncation.
- `backend/internal/adapters/opencode/extractor_test.go`: Added boundary and adversarial table-driven tests (`TestExtractFacts_ErrorPayloadAdversarialBoundaries`) verifying exact 4096-char truncation, 4097-char off-by-one behavior, 5000-char over-limit boundaries, sub-4096 preservation, and empty error fallbacks.
- `web/src/api/client.ts`: Added `fetchWorkspaceEvents` (calling `/api/workspaces/{id}/events` with resilient fallback) and `fetchWorkspaceAnomalies` (extracting anomalies from workspace detail).
- `web/src/api/mockData.ts`: Added rich chronological event trace fixtures for ticket `AL-2` and maintained legacy fallback status for `AL-1`.
- `backend/internal/adapters/web/dist/`: Synchronized production build assets (`index.html`, JavaScript, and CSS bundles) for single-binary embedding via Go `embed.FS`.

### 3. Presentation Layer
- `web/src/components/AnomalyBadge.tsx`: Pure severity badge component (`WARNING` / `CRITICAL`), count display, accessible status/button roles, event propagation control, and tooltip support (`title="Click to open Anomaly Inspector"`).
- `web/src/components/AnomalyDrawer.tsx`: Diagnostic slide-out drawer dynamically titled (`anomaly.title` or `Tool Execution Failure` or `Execution Details`), eliminating artificial fallback strings (`Operational Anomaly Detected` and `ANOMALY_DIAGNOSTIC`). Conditionally renders the Rule Diagnosis section strictly when domain anomaly rules exist, provides expanded scrollable 380px error traceback view, and handles keyboard Escape dismissals.
- `web/src/components/TraceWaterfall.tsx`: Chronological waterfall timeline defaulted to `errors_anomalies` filter mode. Decouples timeline node selection from diagnostic drawer opening: unflagged events highlight the node without opening the drawer; flagged error events and anomaly badges open the inspector drawer. Features prominent `Inspect Failure →` action pill on tool execution errors with tooltip, inline toolbar guidance hint ("Click any error or anomaly to inspect failure details"), and dynamic CSS class differentiation.
- `web/src/components/FleetAnomalyCard.tsx`: Fleet overview anomaly and reliability card with KPI cards, rule distribution counts, brittle tools ranking table, and clean zero-anomaly state.
- `web/src/views/TicketDeepDive.tsx`: Integrated `TraceWaterfall` with decoupled selection logic (clicking normal nodes selects without opening drawer, and dismisses drawer if open; clicking errors or anomalies opens drawer), legacy v1 fallback notification banner, and connected `AnomalyDrawer`.
- `web/src/views/FleetOverview.tsx`: Integrated `FleetAnomalyCard` into the fleet reliability layout.
- `web/src/index.css`: Added styles adhering to dark design tokens. Established 1px transparent baseline border and targeted transitions (`transition: background-color 0.12s ease, border-color 0.12s ease`) on `.timeline-item`, eliminating flash animation and layout jumping on non-issue row selection. Assigned neutral cursor to benign rows and pointer cursor to issue rows (`.item-interactive`, `.item-error`, `.item-anomaly`). Styled `.btn-inspect-failure` action pill, `.waterfall-hint`, and bidirectional scrolling for `.drawer-error-content`.
- Zero emojis across all source files, CSS, docstrings, and tests.

### Verification Status
- Frontend unit and integration tests: 288/288 tests passed across 19 test files (0 skipped).
- Go backend unit and race tests: 771/771 tests passed with `-race` enabled across 10 packages (0 skipped, 0 mocks).
- Static analysis & linting: `tsc --noEmit`, `eslint`, and `go vet` passed with zero warnings.
- Production build: `npm run build` and single-binary compilation `bin/agentic-lens` succeeded cleanly.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 3
- **Active Agent Duration:** 5h 34m (excludes user wait time)
- **Context Window Health:** ELEVATED — Orchestrator peak 170.2k tokens (85% of 200k healthy ceiling, 4 compactions) | Subagents peak 184.2k (Coder)
- **Role Effort Distribution (Subagents):** Planner 13.6%, Coder 54.8%, Tester 18.3%, Reviewer 13.3%
- **Subagent Turn Velocity & Thrashing:** 55.1 avg turns/run (peak: 214 turns) | Planner 44.0, Coder 79.0, Tester 37.8, Reviewer 49.7
- **Test-to-Production Code Ratio:** 0.7:1 (+1,542 test lines / +2,243 prod lines)
- **Code Delivered:** +3,786 / -225 lines across 24 code files (excluding docs)
- **Reasoning Effort:** 232,987 reasoning tokens (53.7% of generated output)
- **Tool Health & Self-Correction:** 1,049 tool calls (31 self-corrections, 97.0% success rate)

### Anomalies & Health Flags
- **Status:** 8 execution anomalies detected (4 CRITICAL, 4 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (166.8k tokens) — Context tokens reached 166.8k in turn 56 (session: ses_eaf08e22bffexSexQrfKUuOmGi), breaching 160k critical ceiling
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.9k tokens) — Context tokens reached 160.9k in turn 71 (session: ses_eaf9229aeffe5uoQhcPFCY0dW7), breaching 160k critical ceiling
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (161.3k tokens) — Context tokens reached 161.3k in turn 120 (session: ses_eafa88c8bffemuBhYUUMhNTFJk), breaching 160k critical ceiling
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+82.1k tokens) — Context tokens surged by 82.1k tokens in turn 131 (session: ses_eafa88c8bffemuBhYUUMhNTFJk, total: 82.1k tokens, threshold: 80000)
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+46.7k tokens) — Context tokens surged by 46.7k tokens in turn 60 (session: ses_eaf08e22bffexSexQrfKUuOmGi, total: 46.7k tokens, threshold: 40000)
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+55.1k tokens) — Context tokens surged by 55.1k tokens in turn 86 (session: ses_eaf9229aeffe5uoQhcPFCY0dW7, total: 55.1k tokens, threshold: 40000)
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+52.6k tokens) — Context tokens surged by 52.6k tokens in turn 153 (session: ses_eaf9229aeffe5uoQhcPFCY0dW7, total: 52.6k tokens, threshold: 40000)
- [WARNING] **TOOL_RETRY_STREAK**: Tool retry streak: grep (3 consecutive errors) — Tool 'grep' encountered 3 consecutive execution failures (session: ses_eafa88c8bffemuBhYUUMhNTFJk, turn: 93)

### Agentic Execution Metrics

**Contributors:** nilton.guedes_duarte@digits.schwarz

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **orchestrator** | 1 | 3,647,342 | 60,674 | 15,219,976 | 80.7% | 170.2k | ELEVATED | $4.24 |
| **planner** | 3 | 720,809 | 29,197 | 5,220,519 | 87.9% | 91.8k | HEALTHY | $1.22 |
| **coder** | 5 | 2,830,918 | 80,610 | 27,954,066 | 90.8% | 184.2k | ELEVATED | $4.92 |
| **tester** | 4 | 997,216 | 17,955 | 9,879,388 | 90.8% | 143.0k | ELEVATED | $1.64 |
| **reviewer** | 3 | 713,707 | 12,774 | 6,939,625 | 90.7% | 98.1k | HEALTHY | $1.19 |
| **Total** | **16** | **8,909,992** | **201,210** | **65,213,574** | **88.0%** | **184.2k** | **ELEVATED** | **$13.20** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gemini-3.8-flash** | 16 | 8,909,992 | 201,210 | 65,213,574 | 88.0% | $13.20 |
| **Total** | **16** | **8,909,992** | **201,210** | **65,213,574** | **88.0%** | **$13.20** |
