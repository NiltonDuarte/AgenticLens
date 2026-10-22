# fix: centralize anomaly detection, live debug identifiers, and AI debug prompts (FIX-3)

## Summary

Extends and unifies anomaly detection, metric consistency, and debug actions across both retrospective Deep Dive analysis and real-time Live Observability:

1. **Centralized Go Domain Anomaly Detection & Projection**:
   - Implemented `EventsFromLiveTrace(trace *LiveTraceView) []EventFact` in `backend/internal/domain/live.go` providing pure projection from live trace views into domain event streams.
   - Added `Count *int64` (`json:"count,omitempty"`) to `LiveAnomalyAlert` in `backend/internal/domain/live_anomaly.go`.
   - Wired exact calculated streak counts into `EvaluateConsecutiveToolErrors` and `EvaluateToolThrashing`.
   - In `backend/internal/adapters/live/service.go` (`HydrateWorkspaceEvents`), accumulated consecutive tool errors across sequential message turns within each session via `sessionConsecutiveErrors`.

2. **Shared TypeScript Anomaly Domain Module**:
   - Centralized anomaly matching predicates and helpers into `web/src/domain/anomalies.ts`: `isToolFailureAnomalyRule`, `matchesAnomaly`, `findMatchingAnomaly`, `extractStreakCount`, and `findMatchingLiveAlert`.
   - Refactored `TraceWaterfall.tsx` to consume these domain helpers and re-export them for backwards compatibility.

3. **Multi-Layer Fluent AI Debug Prompts & Generic Agent Abstraction (`web/src/utils/debugPrompt.ts`)**:
   - Developed 4-layer Markdown debug prompt generators tailored for external AI reasoning without internal project leakages:
     - **Layer 1 (Tool Call / Part Execution)**: Heading `# Tool Execution Failure: <tool>`, narrative provenance, fenced inputs and outputs, and targeted diagnostic checks.
     - **Layer 2 (Execution Issue / Anomaly Card)**: Heading `# Agent Execution Issue: <Sanitized Summary>`, subagent role/model/part provenance, error details block, and loop-recovery diagnostic checks.
     - **Layer 3 (Message Turn Review)**: Heading `# Turn <N> Review (<role> / <model>)`, model deliberation blockquotes, message text body, executed tools listing, and intent vs outcome checks.
     - **Layer 4 (Session Overview)**: Heading `# Session Overview: <session_id>`, subagent role/model/parent narrative, turn/tool counts, plain English flagged issues, and session-level checks.
   - Applied strict sanitization filters:
     - Stripped internal ticket identifiers (`AL-*`, `FIX-*`, `GEN-*`, `PO-*`).
     - Removed git branch patterns (`feat/*`, `fix/*`).
     - Translated internal anomaly rule codes into human-readable plain English descriptions.
     - Purged internal severity tags (`[CRITICAL]`, `[WARNING]`, `[HIGH]`).

4. **Interactive UI Integration with Temporary Feedback**:
   - Unified all copy actions to `"Copy Prompt"` with a deterministic 2000ms `"Copied"` feedback state.
   - Wired `LiveTraceInspector.tsx` (session header -> Layer 4; turn headers -> Layer 3; part rows -> Layer 1; failed tool cards -> Layer 2).
   - Wired `FailedToolAlertCard.tsx` (header copy button -> Layer 2 prompt with contextual error details).
   - Wired `AnomalyDrawer.tsx` (header copy button -> Layer 2 prompt with trigger event details while retaining discrete ID tokens).
   - Wired `TraceWaterfall.tsx` (timeline row button -> Layer 1 for tool calls, Layer 3 for message turns).

5. **Strong Anomaly Rule Typing and OpenCode DB Provenance**:
   - Introduced named type `LiveAnomalyRuleName` in Go domain and typed canonical constants and `LiveAnomalyAlert.RuleName`.
   - Exported `AnomalyRule = AnomalyRuleID | LiveAnomalyRule` and `ToolFailureAnomalyRule` in `web/src/domain/anomalies.ts`.
   - Replaced content heuristic checks (`startsWith('TOOL_')`, `includes('TOOL_FAILURE')`) in `isToolFailureAnomalyRule` with strict equality matching against canonical tool failure rules.
   - Typed `LiveAnomalyAlert.rule_name` strictly as `LiveAnomalyRule`.
   - Added explicit OpenCode database provenance statement (`This data was extracted from the OpenCode database.`) across all 4 prompt generator layers.

6. **Primitive Obsession Remediation for Declared Domain Enums**:
   - Replaced raw `string` with `domain.ContextHealthStatus` across Go backend domain structs (`GroupedMetric`, `ContextWindowHealth`, `ModelMetric`, `RoleMetric`, `TicketComparisonRow`) and TypeScript interfaces.
   - Replaced raw `string` with `domain.AgentRole` across fact records (`InvocationFact.Role`, `EventFact.Role`, `ThreadRoleMetric.Role`, `ContextWindowHealth.SubagentsPeakRole`, `LiveMessageFact.Role`) and role maps. Made `OrderRoles[T ~string]` generic.
   - Exported `CanonicalAgentRole` and autocomplete union `AgentRole = CanonicalAgentRole | (string & {})` in TypeScript, updating all domain fact interfaces.
   - Updated `LivePartFact.Severity` to `FailureSeverity`, `ClassifyToolFailure` to return `(bool, FailureSeverity, *string)`, and typed `LiveAnomalySeverity`. Exported `FailureSeverity` in TypeScript.
   - Replaced raw `*string` with `*domain.ConventionalTicketType` in `AgenticFactLedger`, `AgenticMetricsLedger`, and `TicketComparisonRow`. Updated TypeScript `StandardTicketType` and `TicketType = StandardTicketType | (string & {})`.

7. **Platform Decoupling and Removal of Repository Role Assumptions**:
   - Eliminated repository-specific role constants (`RoleOrchestrator`, `RolePlanner`, `RoleCoder`, `RoleTester`, `RoleReviewer`, `RoleUnassigned`) from Go backend domain core (`backend/internal/domain/enums.go`).
   - Declared generic platform fallback constants: `RoleRootAgent AgentRole = "RootAgent"` for nameless root sessions and `RoleUnassignedSubAgent AgentRole = "UnassignedSubAgent"` for nameless subagents and unmapped sessions.
   - Updated live adapters (`service.go`, `extractor.go`, `aggregations.go`, `config.go`) to use platform fallbacks.
   - Eliminated `CanonicalAgentRole`, `CANONICAL_ROLE_ORDER`, `SUBAGENT_ROLE_ORDER`, and `isCanonicalRole` from TypeScript contracts (`web/src/types/telemetry.ts` and `web/src/domain/types.ts`), redefining `type AgentRole = string`.
   - Clean break renamed root context metrics from `orchestrator_peak_tokens` / `orchestrator_saturation_pct` to `root_peak_tokens` / `root_saturation_pct` across Go structs, JSON tags, TypeScript contracts, and UI components.
   - Decoupled theme role colors in `web/src/utils/theme.ts` via deterministic hash-to-palette algorithm (`hashStringToColor`), eliminating hardcoded agent role styling.
   - Generalized ticket stripping regex in `web/src/utils/debugPrompt.ts` to `[A-Z]{2,10}-\d+`, eliminating hardcoded repository ticket prefixes (`AL`, `FIX`, `GEN`, `PO`).
   - Decoupled CLI formatter output string to report `Root agent peak %s tokens`.

## Testing & Verification
- `go test -race -count=1 ./...` (All Go backend tests passed cleanly)
- `go vet ./...` (Zero warnings)
- `npm run typecheck` (Zero TypeScript errors)
- `npm run lint` (Zero ESLint warnings/errors)
- `npm test` (All 728 frontend tests passed across 30 test suites)
- `make build` (Production binary build and asset embedding clean)
- Zero emojis across all modified code, tests, and markdown artifacts.
