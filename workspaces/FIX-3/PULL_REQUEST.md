# fix: centralize anomaly detection and add live debug identifiers (FIX-3)

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

3. **Live Section Debug Identifiers with One-Click Copy**:
   - In `LiveTraceInspector.tsx`:
     - Added Copy button beside Session ID in the inspector header with 2000ms temporary "Copied" feedback.
     - Added Turn token (`Turn X`), Message ID debug token, and Copy ID button with 2000ms feedback to `LiveMessageTurnItem`.
     - Added Part ID debug token and Copy button in `LiveMessagePartItem` part headers.
     - Wired `FailedToolAlertCard` consecutive errors to exact streak count from matching anomaly rules via `findMatchingLiveAlert` and `extractStreakCount` instead of raw `trace.recent_errors.length`.

## Testing & Verification
- `go test -race -count=1 ./...` (All Go backend tests passed cleanly)
- `go vet ./...` (Zero warnings)
- `npm run typecheck` (Zero TypeScript errors)
- `npm run lint` (Zero ESLint warnings/errors)
- `npm test` (All 648 frontend tests passed across 29 test suites)
- `make build` (Production binary build and asset embedding clean)
- Zero emojis across all modified code, tests, and markdown artifacts.
