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

## Testing & Verification
- `go test -race -count=1 ./...` (All Go backend tests passed cleanly)
- `go vet ./...` (Zero warnings)
- `npm run typecheck` (Zero TypeScript errors)
- `npm run lint` (Zero ESLint warnings/errors)
- `npm test` (All 725 frontend tests passed across 30 test suites)
- `make build` (Production binary build and asset embedding clean)
- Zero emojis across all modified code, tests, and markdown artifacts.
