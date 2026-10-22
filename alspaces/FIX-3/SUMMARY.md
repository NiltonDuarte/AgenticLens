# Delivery Summary: FIX-3

## Overview
- **Ticket ID**: `FIX-3`
- **Ticket Title**: `Resolve False Tool Errors, Anomaly Misattributions, Waterfall Collapsibility, and Debug Identifiers`
- **Ticket Type**: `fix`
- **Branch**: `fix/FIX-3-centralize-live-anomalies-debug-identifiers`

---

## Iterations Summary

### Iteration 01
- Resolved false tool call errors on content inspection tools (`read`, `grep`, `glob`, `webfetch`) when status is completed.
- Scoped tool retry streaks per session ID in `detectToolRetryStreaks`.
- Implemented collapsible Execution Trace Waterfall collapsed by default with concise summary header.
- Added debug identifiers and clipboard copy action in `AnomalyDrawer`, `TraceWaterfall`, and `FailedToolAlertCard`.

### Iteration 02
- **Centralized Anomaly Detection Core**:
  - Implemented pure projection `EventsFromLiveTrace(trace *LiveTraceView) []EventFact` in `backend/internal/domain/live.go` mapping live traces directly to domain event facts.
  - Added `Count *int64` (`json:"count,omitempty"`) to `LiveAnomalyAlert` in `backend/internal/domain/live_anomaly.go` and populated exact failure streak counts in `EvaluateConsecutiveToolErrors` and `EvaluateToolThrashing`.
  - In `backend/internal/adapters/live/service.go` (`HydrateWorkspaceEvents`), tracked consecutive tool errors across sequential message turns per session using `sessionConsecutiveErrors` accumulator map.
- **Centralized TypeScript Domain Anomaly Module**:
  - Authored `web/src/domain/anomalies.ts` exporting pure functions `isToolFailureAnomalyRule`, `matchesAnomaly`, `findMatchingAnomaly`, `extractStreakCount`, and `findMatchingLiveAlert`.
  - Refactored `TraceWaterfall.tsx` to consume the centralized domain functions and re-export them for backwards compatibility.
- **Live Observability Debug Identifiers and Clipboard Actions**:
  - Added Copy button beside Session ID in the `LiveTraceInspector` header with 2000ms temporary feedback.
  - Added Turn token (`Turn X`), Message ID debug token, and Copy ID button with 2000ms feedback to `LiveMessageTurnItem`.
  - Added Part ID debug token and Copy button in `LiveMessagePartItem` part-header.
  - Wired `FailedToolAlertCard` consecutive errors to exact streak count from matching `LiveAnomalyAlert.count` via `findMatchingLiveAlert` and `extractStreakCount` rather than `trace.recent_errors.length`.

### Iteration 03
- **Multi-Layer Fluent AI Debug Prompt Generators (`web/src/utils/debugPrompt.ts`)**:
  - Formatted 4 dedicated Markdown investigation prompt generators:
    - Layer 1 (`generateToolExecutionPrompt`): Formats tool call, inputs, outputs, consecutive error count, and provenance details.
    - Layer 2 (`generateExecutionIssuePrompt`): Formats anomaly summary, trigger part provenance, description, and error details.
    - Layer 3 (`generateMessageTurnPrompt`): Formats turn deliberation, message body, executed tools list, and turn provenance.
    - Layer 4 (`generateSessionOverviewPrompt`): Formats subagent role, runtime model, parent session hierarchy, lifecycle status, turn/tool counts, and plain English flagged issues.
  - Enforced generic agent environment abstraction:
    - Stripped internal ticket IDs (`AL-*`, `FIX-*`, `GEN-*`, `PO-*`).
    - Stripped git branch references (`feat/*`, `fix/*`).
    - Translated internal anomaly rule codes into human-readable plain English descriptions.
    - Purged internal severity tiers (`[CRITICAL]`, `[WARNING]`, `[HIGH]`).
- **Interactive UI Integration**:
  - Unified all copy buttons under the label `"Copy Prompt"` with 2000ms `"Copied"` feedback.
  - `LiveTraceInspector.tsx`: Session header copies Layer 4; turn headers copy Layer 3; part rows copy Layer 1; failed tool alert cards copy Layer 2.
  - `FailedToolAlertCard.tsx`: Extended with contextual props and wired to copy Layer 2 prompts.
  - `AnomalyDrawer.tsx`: Header contains "Copy Prompt" button wired to copy Layer 2 prompts.
  - `TraceWaterfall.tsx`: Timeline row button copies Layer 1 for `tool_call` events and Layer 3 for message turns.

### Iteration 04
- **Strong Anomaly Rule Typing**:
  - Defined named type `LiveAnomalyRuleName string` in `backend/internal/domain/live_anomaly.go` and typed canonical rule constants and `LiveAnomalyAlert.RuleName`.
  - Exported `type AnomalyRule = AnomalyRuleID | LiveAnomalyRule` and `type ToolFailureAnomalyRule` in `web/src/domain/anomalies.ts`.
  - Refactored `isToolFailureAnomalyRule(ruleId: AnomalyRule | string): ruleId is ToolFailureAnomalyRule` with strict equality matching against the 4 canonical tool failure rules, eliminating heuristic `.startsWith('TOOL_')` and `.includes('TOOL_FAILURE')` checks.
  - Typed `LiveAnomalyAlert.rule_name` strictly as `LiveAnomalyRule`.
  - Re-exported `AnomalyRule` and `ToolFailureAnomalyRule` from `web/src/domain/types.ts` and `web/src/components/TraceWaterfall.tsx`.
- **OpenCode Database Telemetry Provenance Attribution**:
  - Updated all 4 debug prompt generators in `web/src/utils/debugPrompt.ts` (`generateToolExecutionPrompt`, `generateExecutionIssuePrompt`, `generateMessageTurnPrompt`, `generateSessionOverviewPrompt`) to explicitly declare: `This data was extracted from the OpenCode database.` in their narrative provenance paragraph.

### Iteration 05
- **Domain Enums and Primitive Obsession Remediation**:
  - **Context Window Health (`ContextHealthStatus`)**:
    - Replaced raw `string` with `domain.ContextHealthStatus` across Go backend domain structs (`GroupedMetric.ContextHealth`, `ContextWindowHealth.Status`, `ModelMetric.ContextHealth`, `RoleMetric.ContextHealth`, and `TicketComparisonRow.ContextHealthStatus`).
    - Aligned aggregations in `aggregations.go` to return `ContextHealthStatus` directly from pure helper `GetBaseHealthStatus`.
    - Typed frontend interfaces (`GroupedMetric.context_health`, `ContextWindowHealth.status`, `ModelMetric.context_health`, `RoleMetric.context_health`, `TicketComparisonRow.context_health_status`, and `TicketSelectorProps.context_health_status`) as `ContextHealthStatus`.
  - **Agent Role (`AgentRole` & `CanonicalAgentRole`)**:
    - Replaced raw `string` with `domain.AgentRole` across `InvocationFact.Role`, `EventFact.Role`, `ThreadRoleMetric.Role`, `ContextWindowHealth.SubagentsPeakRole`, and `LiveMessageFact.Role`.
    - Typed role map keys (`ContextWindowHealth.PeakByRole`, `SubagentTurnVelocity.ByRole`, and `FleetSummary.RoleCosts`) as `map[AgentRole]T`.
    - Made `OrderRoles[T ~string]` and `OrderSubagentRoles[T ~string]` generic to sort any slice of strings or `AgentRole` without conversion or allocation.
    - Exported `CanonicalAgentRole` and autocomplete union `AgentRole = CanonicalAgentRole | (string & {})` in TypeScript, updating all domain fact interfaces, anomaly contexts, and card props.
  - **Tool & Failure Severity (`FailureSeverity` & `LiveAnomalySeverity`)**:
    - Updated `LivePartFact.Severity` to `domain.FailureSeverity` and `ClassifyToolFailure` to return `(bool, FailureSeverity, *string)`.
    - Defined typed `domain.LiveAnomalySeverity` (`LiveSeverityWarning`, `LiveSeverityHigh`, `LiveSeverityCritical`) for live anomaly alerts.
    - Exported TypeScript `FailureSeverity = 'none' | 'low' | 'medium' | 'high'`, updating `LivePartFact.severity` and fixing test fixture mocks.
  - **Conventional Ticket Types (`ConventionalTicketType` & `TicketType`)**:
    - Replaced `*string` with `*domain.ConventionalTicketType` in `AgenticFactLedger`, `AgenticMetricsLedger`, `TicketComparisonRow`, and Git branch inference helpers.
    - Updated wire DTOs (`schemas.go`) with clean conversion to/from `domain.ConventionalTicketType` and generic `ptrEqual[T comparable]` migration check.
    - Updated TypeScript `StandardTicketType` and `TicketType = StandardTicketType | (string & {})`.

### Iteration 06
- **Platform Decoupling and Removal of Repository Role Assumptions**:
  - Eliminated repository-specific role constants (`RoleOrchestrator`, `RolePlanner`, `RoleCoder`, `RoleTester`, `RoleReviewer`, `RoleUnassigned`) from Go backend domain core (`backend/internal/domain/enums.go`).
  - Introduced generic platform fallback constants: `RoleRootAgent AgentRole = "RootAgent"` for nameless root sessions and `RoleUnassignedSubAgent AgentRole = "UnassignedSubAgent"` for nameless subagents and unmapped sessions.
  - Updated adapters (`backend/internal/adapters/live/service.go`, `backend/internal/adapters/opencode/extractor.go`, `backend/internal/domain/config.go`, `backend/internal/domain/aggregations.go`) to use platform fallbacks.
  - Eliminated `CanonicalAgentRole`, `CANONICAL_ROLE_ORDER`, `SUBAGENT_ROLE_ORDER`, and `isCanonicalRole` from TypeScript contracts (`web/src/types/telemetry.ts` and `web/src/domain/types.ts`), redefining `type AgentRole = string`.
- **Clean Break Renaming of Root Context Health Metrics**:
  - Renamed `ContextWindowHealth.OrchestratorPeakTokens` to `RootPeakTokens` (`json:"root_peak_tokens"`).
  - Renamed `ContextWindowHealth.OrchestratorSaturationPct` to `RootSaturationPct` (`json:"root_saturation_pct"`).
  - Renamed `TicketComparisonRow.OrchestratorSaturationPct` to `RootSaturationPct` (`json:"root_saturation_pct"`).
  - Updated frontend interfaces, UI presentation components (`SaturationBar.tsx`, `TicketComparisons.tsx`, `TicketTable.tsx`), formatters, and mock data fixtures.
- **Theme Palette Decoupling (`web/src/utils/theme.ts`)**:
  - Removed `CANONICAL_ROLE_COLORS`.
  - Implemented deterministic hash-to-palette algorithm (`hashStringToColor`) across 12 distinct accessible color tokens, ensuring zero hardcoded role names in frontend styling.
- **Generalized Debug Prompt Ticket Regex (`web/src/utils/debugPrompt.ts`)**:
  - Generalized ticket stripping regex to `[A-Z]{2,10}-\d+`, eliminating hardcoded repository prefixes (`AL`, `FIX`, `GEN`, `PO`).
- **CLI Formatter Decoupling (`backend/internal/adapters/formatters.go`)**:
  - Formatted metrics table output to report `Root agent peak %s tokens` instead of `Orchestrator peak`.

---

## Verification & Test Results
- **Go Unit & Race Tests**: `go test -race -count=1 ./...` (All tests passed cleanly).
- **Go Vet**: `go vet ./...` (Zero issues).
- **TypeScript Typecheck**: `npm run typecheck` (Zero errors).
- **ESLint**: `npm run lint` (Zero warnings).
- **Frontend Test Suite**: `npm test` (701/701 tests passing across 30 test suites).
- **Production Build**: `make build` (Generated release binaries with embedded web UI).
- **Zero Emojis**: Verified zero emojis across code, tests, and documentation.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 22
- **Active Agent Duration:** 28h 22m (excludes user wait time)
- **Context Window Health:** ELEVATED — Root agent peak 174.3k tokens (87% of 200k healthy ceiling, 21 compactions) | Subagents peak 171.2k (Coder)
- **Role Effort Distribution (Subagents):** Planner 6.8%, Coder 73.4%, Tester 11.5%, Reviewer 8.3%
- **Subagent Turn Velocity & Thrashing:** 165.5 avg turns/run (peak: 1672 turns) | Planner 56.5, Coder 434.9, Tester 84.0, Reviewer 59.3
- **Test-to-Production Code Ratio:** 2.0:1 (+3,213 test lines / +1,643 prod lines)
- **Code Delivered:** +4,857 / -774 lines across 57 code files (excluding docs)
- **Reasoning Effort:** 594,356 reasoning tokens (46.4% of generated output)
- **Tool Health & Self-Correction:** 4,724 tool calls (52 self-corrections, 98.9% success rate)

### Anomalies & Health Flags
- **Status:** 29 execution anomalies detected (21 CRITICAL, 8 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (161.0k tokens) — Context tokens reached 161.0k in turn 38 (session: ses_e9f2e1db2fferR7NvzZ0uguWkT), breaching 160k critical ceiling
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (161.4k tokens) — Context tokens reached 161.4k in turn 48 (session: ses_e9f645441ffeLoPoDrRwxIQajj), breaching 160k critical ceiling
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.2k tokens) — Context tokens reached 160.2k in turn 67 (session: ses_ea24ee73affeUPiQ7FMWOo0cS9), breaching 160k critical ceiling
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.8k tokens) — Context tokens reached 160.8k in turn 86 (session: ses_ea026f776ffeMcs4kPDJwMBVgo), breaching 160k critical ceiling
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.8k tokens) — Context tokens reached 160.8k in turn 116 (session: ses_e9f2247dbffeJHdhGhq4vE2Q5z), breaching 160k critical ceiling
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.0k tokens) — Context tokens reached 160.0k in turn 168 (session: ses_e9d912167ffeLO6s00pZOEWIE7), breaching 160k critical ceiling
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+110.1k tokens) — Context tokens surged by 110.1k tokens in turn 193 (session: ses_e9d912167ffeLO6s00pZOEWIE7, total: 110.1k tokens, threshold: 80000)
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.5k tokens) — Context tokens reached 160.5k in turn 204 (session: ses_ea224a06dffeUUVUn1Vi3HlWh6), breaching 160k critical ceiling
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+105.1k tokens) — Context tokens surged by 105.1k tokens in turn 205 (session: ses_e9f645441ffeLoPoDrRwxIQajj, total: 105.1k tokens, threshold: 80000)
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+93.6k tokens) — Context tokens surged by 93.6k tokens in turn 217 (session: ses_ea24ee73affeUPiQ7FMWOo0cS9, total: 93.6k tokens, threshold: 80000)
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+108.8k tokens) — Context tokens surged by 108.8k tokens in turn 218 (session: ses_ea224a06dffeUUVUn1Vi3HlWh6, total: 108.8k tokens, threshold: 80000)
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+114.5k tokens) — Context tokens surged by 114.5k tokens in turn 272 (session: ses_ea24ee73affeUPiQ7FMWOo0cS9, total: 114.5k tokens, threshold: 80000)
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+117.0k tokens) — Context tokens surged by 117.0k tokens in turn 380 (session: ses_e9f645441ffeLoPoDrRwxIQajj, total: 117.0k tokens, threshold: 80000)
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+84.8k tokens) — Context tokens surged by 84.8k tokens in turn 434 (session: ses_ea24ee73affeUPiQ7FMWOo0cS9, total: 84.8k tokens, threshold: 80000)
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+107.5k tokens) — Context tokens surged by 107.5k tokens in turn 474 (session: ses_ea224a06dffeUUVUn1Vi3HlWh6, total: 107.5k tokens, threshold: 80000)
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+120.2k tokens) — Context tokens surged by 120.2k tokens in turn 803 (session: ses_ea224a06dffeUUVUn1Vi3HlWh6, total: 120.2k tokens, threshold: 80000)
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+116.1k tokens) — Context tokens surged by 116.1k tokens in turn 968 (session: ses_ea224a06dffeUUVUn1Vi3HlWh6, total: 116.1k tokens, threshold: 80000)
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+103.7k tokens) — Context tokens surged by 103.7k tokens in turn 1110 (session: ses_ea224a06dffeUUVUn1Vi3HlWh6, total: 103.7k tokens, threshold: 80000)
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+93.6k tokens) — Context tokens surged by 93.6k tokens in turn 1240 (session: ses_ea224a06dffeUUVUn1Vi3HlWh6, total: 93.6k tokens, threshold: 80000)
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+110.3k tokens) — Context tokens surged by 110.3k tokens in turn 1383 (session: ses_ea224a06dffeUUVUn1Vi3HlWh6, total: 110.3k tokens, threshold: 80000)
- [CRITICAL] **TOOL_HIGH_FAILURE_RATE**: High tool failure rate: mcp-gateway_mcp-exec (100.0%) — Tool 'mcp-gateway_mcp-exec' failed 5 out of 5 calls (100.0% failure rate, threshold: >=50%)
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+42.0k tokens) — Context tokens surged by 42.0k tokens in turn 81 (session: ses_ea24ee73affeUPiQ7FMWOo0cS9, total: 42.0k tokens, threshold: 40000)
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+66.5k tokens) — Context tokens surged by 66.5k tokens in turn 95 (session: ses_ea026f776ffeMcs4kPDJwMBVgo, total: 66.5k tokens, threshold: 40000)
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+79.1k tokens) — Context tokens surged by 79.1k tokens in turn 128 (session: ses_e9f2247dbffeJHdhGhq4vE2Q5z, total: 79.1k tokens, threshold: 40000)
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+58.2k tokens) — Context tokens surged by 58.2k tokens in turn 317 (session: ses_ea24ee73affeUPiQ7FMWOo0cS9, total: 58.2k tokens, threshold: 40000)
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+58.2k tokens) — Context tokens surged by 58.2k tokens in turn 463 (session: ses_ea24ee73affeUPiQ7FMWOo0cS9, total: 58.2k tokens, threshold: 40000)
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+72.8k tokens) — Context tokens surged by 72.8k tokens in turn 601 (session: ses_ea224a06dffeUUVUn1Vi3HlWh6, total: 72.8k tokens, threshold: 40000)
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+66.9k tokens) — Context tokens surged by 66.9k tokens in turn 1459 (session: ses_ea224a06dffeUUVUn1Vi3HlWh6, total: 66.9k tokens, threshold: 40000)
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+72.3k tokens) — Context tokens surged by 72.3k tokens in turn 1579 (session: ses_ea224a06dffeUUVUn1Vi3HlWh6, total: 72.3k tokens, threshold: 40000)

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 6 | 2,014,718 | 47,071 | 18,142,843 | 90.0% | 139.9k | ELEVATED | $3.54 |
| **coder** | 7 | 20,766,440 | 373,288 | 274,234,612 | 93.0% | 171.2k | ELEVATED | $38.24 |
| **tester** | 6 | 3,415,592 | 101,053 | 35,590,910 | 91.2% | 154.4k | ELEVATED | $6.02 |
| **reviewer** | 7 | 3,002,475 | 34,794 | 21,825,067 | 87.9% | 167.0k | ELEVATED | $4.30 |
| **advisor** | 1 | 10,739,040 | 129,830 | 35,383,059 | 76.7% | 174.3k | ELEVATED | $11.54 |
| **Total** | **27** | **39,938,265** | **686,036** | **385,176,491** | **90.6%** | **174.3k** | **ELEVATED** | **$63.64** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gemini-3.8-flash** | 27 | 39,938,265 | 686,036 | 385,176,491 | 90.6% | $63.64 |
| **Total** | **27** | **39,938,265** | **686,036** | **385,176,491** | **90.6%** | **$63.64** |
