# fix: resolve false tool errors, retry streak misattribution, collapse waterfall by default, and add debug copy action

## Summary

Resolves four interrelated operational issues impairing developer and operator observability during agent execution runs:

1. **False Tool Call Errors**:
   - In `backend/internal/domain/live.go` (`ClassifyToolFailure`) and `backend/internal/adapters/opencode/extractor.go`, content reading/inspection tools (`read`, `grep`, `glob`, `webfetch`) with exit status `completed` were scanning document contents for syntax errors, tracebacks, and error strings.
   - Reading files documenting syntax errors (such as `TECHNICAL_GUIDELINES.md` or test specs) caused successful calls to be classified as high severity errors and triggered false consecutive failure streaks (e.g. 7 consecutive failures).
   - Completed content inspection calls now immediately return non-error with `SeverityNone`, bypassing content error scans.

2. **Tool Retry Streak Scoping & Waterfall Anomaly Matching**:
   - In `backend/internal/domain/anomalies.go` (`detectToolRetryStreaks`), consecutive error streaks are now strictly scoped per session ID, preventing errors in different sessions from aggregating into a false streak.
   - In `web/src/components/TraceWaterfall.tsx` (`matchesAnomaly`), tool failure rules (e.g. `TOOL_RETRY_STREAK`, `TOOL_HIGH_FAILURE_RATE`) now require `event.status === 'error'`, preventing successful calls from being tagged with failure anomalies. Session IDs are also strictly checked to prevent cross-session misattribution.

3. **Trace Waterfall Collapsible Section**:
   - The Execution Trace Waterfall section in `web/src/components/TraceWaterfall.tsx` is now collapsed by default, displaying a compact summary header with total event count, data source badge (`Source: Local SQLite` vs `Source: Committed Ledger`), and issues count (`X Errors, Y Anomalies` / `Clean Execution`) alongside an expand button (`data-testid="waterfall-expand-btn"`).
   - Expanding exposes the full filter toolbar and chronological timeline.

4. **Concrete Debug Identifiers with One-Click Copy Action**:
   - Added `MessageID` to `EventFact` domain models in Go and TypeScript, populated from OpenCode SQLite telemetry.
   - Provided concrete debug identifiers (Session ID, Message ID when present, and Turn Debug ID) with a Copy to Clipboard action and temporary 2000ms feedback in `AnomalyDrawer.tsx`, `TraceWaterfall.tsx`, and `FailedToolAlertCard.tsx`.

## Testing & Verification
- `go test -race -count=1 ./...` (All Go backend tests passed cleanly)
- `go vet ./...` (Zero warnings)
- `npm run typecheck` (Zero TypeScript errors)
- `npm run lint` (Zero ESLint warnings/errors)
- `npm test` (All 614 frontend tests passed across 28 test suites)
- `make build` (Production binary build and asset embedding clean)
- Zero emojis across all modified code, tests, and markdown artifacts.
