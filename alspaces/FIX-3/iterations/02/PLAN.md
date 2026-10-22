# Implementation Plan: FIX-3 Iteration 02 - Centralize Anomaly Detection & Live Debug Identifiers

## 1. Goal and Iteration Scope

Incorporate steering feedback for ticket FIX-3 to propagate Deep Dive/Waterfall fixes into the Live Observability section:
- Centralize anomaly detection in the Go backend so both Deep Dive (retrospective event facts) and Live Observability (real-time session traces) share identical algorithms, rules, and mathematical invariants.
- Add `EventsFromLiveTrace(trace *LiveTraceView) []EventFact` in the Go domain so live traces can be cleanly projected into domain event streams for uniform anomaly evaluation.
- Ensure tool retry streaks, consecutive failure counters, and content inspection bypasses are applied identically across both sections.
- Populate `Count *int64` on `LiveAnomalyAlert` so downstream consumers receive exact calculated streak counts.
- Centralize frontend anomaly matching logic into `web/src/domain/anomalies.ts` shared by `TraceWaterfall.tsx` and `LiveTraceInspector.tsx`.
- Add concrete debug identifiers (Session ID, Message ID, Turn Debug ID, Part ID) with one-click Copy to Clipboard actions and temporary 2000ms feedback to the Live section message turns, parts, and header in `LiveTraceInspector.tsx`.
- Wire exact anomaly streak counts into `FailedToolAlertCard` in `LiveTraceInspector.tsx` instead of passing raw `trace.recent_errors.length`.

---

## 2. Grounded Reference Architecture & File Paths

### Go Functional Core (`backend/internal/domain/`)
- `backend/internal/domain/live.go`:
  - `EventsFromLiveTrace(trace *LiveTraceView) []EventFact`: Pure projection mapping live messages and parts into domain `EventFact` structs.
- `backend/internal/domain/live_anomaly.go`:
  - Extend `LiveAnomalyAlert` with `Count *int64` (`json:"count,omitempty"`).
  - Refactor `EvaluateConsecutiveToolErrors` and `EvaluateToolThrashing` to populate exact failure streak counts.
- `backend/internal/domain/anomalies.go`:
  - Shared anomaly rules (`detectToolRetryStreaks`, `detectToolThrashing`, `detectContextSaturationSpikes`).
- `backend/internal/domain/live_anomaly_test.go` and `live_test.go`:
  - Unit tests verifying `EventsFromLiveTrace`, `Count` population on `LiveAnomalyAlert`, and streak accuracy.

### Go Boundary Adapters (`backend/internal/adapters/live/`)
- `backend/internal/adapters/live/service.go`:
  - In `GetWorkspaceEvents`: maintain consecutive error state across message turns per session rather than resetting per message.
  - In `GetSessionTrace`: ensure `trace.Anomalies` are evaluated using the unified rules with exact streak counts.
- `backend/internal/adapters/live/sqlite_test.go`:
  - Unit tests verifying consecutive error tracking across turns in SQLite live sessions.

### TypeScript Domain (`web/src/domain/`)
- `web/src/domain/anomalies.ts`:
  - Pure matching module: `isToolFailureAnomalyRule`, `matchesAnomaly`, `findMatchingAnomaly`, `extractStreakCount`, `findMatchingLiveAlert`.
- `web/src/domain/anomalies.test.ts`:
  - Comprehensive unit tests covering pure anomaly predicates, session isolation, status checks, and streak extraction.
- `web/src/domain/live.ts`:
  - Update `LiveAnomalyAlert` with `readonly count?: number | null`.

### TypeScript Presentation Components (`web/src/components/`)
- `web/src/components/TraceWaterfall.tsx`:
  - Refactor to import `matchesAnomaly`, `findMatchingAnomaly`, `isToolFailureAnomalyRule` from `../domain/anomalies`.
  - Re-export for backwards compatibility with existing component test imports.
- `web/src/components/LiveTraceInspector.tsx`:
  - Header: Add Copy button beside Session ID token with 2000ms Copied feedback.
  - `LiveMessageTurnItem`: Render debug metadata with Message ID (`msg.id`), Turn token (`Turn X`), and Copy ID button with 2000ms feedback.
  - Message parts: Render Part ID (`part.id`) with Copy button in `part-header`.
  - `FailedToolAlertCard`: Match error part against `trace.anomalies` to extract exact streak count, passing to `consecutiveErrors` instead of `trace.recent_errors.length`.
- `web/src/components/LiveTraceInspector.test.tsx`:
  - Tests verifying Session ID copy, Message ID copy, and Part ID copy buttons and clipboard invocations.
  - Test verifying `FailedToolAlertCard` receives exact streak count from anomaly.

---

## 3. Phased Implementation Tasks for Coder

### Phase 1: Go Domain Centralization & Projection
- Task 1.1: In `backend/internal/domain/live.go`, implement `EventsFromLiveTrace(trace *LiveTraceView) []EventFact`.
  - Map each message to a `turn_message` event with turn number, role, timestamp, tokens, reasoning, and message ID.
  - Map each tool part to a `tool_call` event with tool name, status (error if `is_error` is true, completed otherwise), input/output snippets, and message ID.
  - Map compaction parts to `compaction` events.
- Task 1.2: In `backend/internal/domain/live_anomaly.go`, add `Count *int64` to `LiveAnomalyAlert`.
  - Refactor `EvaluateConsecutiveToolErrors` to set `Count` to the calculated streak count.
  - Refactor `EvaluateToolThrashing` to set `Count` to the failure count.
- Task 1.3: Add and execute unit tests in `live_test.go` and `live_anomaly_test.go`.

### Phase 2: Go Adapter Fixes in LiveService
- Task 2.1: In `backend/internal/adapters/live/service.go`, in `GetWorkspaceEvents`, maintain consecutive error counts across messages within each session using a per-session error accumulator map.
- Task 2.2: Verify adapter tests in `sqlite_test.go`.

### Phase 3: TypeScript Domain Anomaly Module
- Task 3.1: Create `web/src/domain/anomalies.ts` with pure anomaly matching functions.
- Task 3.2: Create `web/src/domain/anomalies.test.ts` testing all predicates.
- Task 3.3: Update `LiveAnomalyAlert` in `web/src/domain/live.ts` with `readonly count?: number | null`.
- Task 3.4: In `web/src/components/TraceWaterfall.tsx`, import matching logic from `../domain/anomalies` and re-export.

### Phase 4: Live Observability UI Debug Identifiers & Copy Buttons
- Task 4.1: In `web/src/components/LiveTraceInspector.tsx`:
  - Add Copy button to header beside Session ID.
  - In `LiveMessageTurnItem`: render Message ID and Turn index debug ID with Copy button and 2000ms feedback.
  - In message parts: render Part ID with Copy button in `part-header`.
  - In `FailedToolAlertCard`: extract streak count from `trace.anomalies` via `findMatchingLiveAlert` / `extractStreakCount` and pass to `consecutiveErrors`.
- Task 4.2: Update and expand `LiveTraceInspector.test.tsx` to assert new copy actions and streak wiring.

### Phase 5: Verification & Quality Gates
- Task 5.1: `go test -race ./...` and `go vet ./...`
- Task 5.2: `npm run typecheck`, `npm run lint`, `npm test`
- Task 5.3: `make build` single-binary verification.
- Task 5.4: Strictly zero emojis and zero any across all Go and TypeScript code.
