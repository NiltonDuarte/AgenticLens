# feat(live): implement live anomaly detection engine and real-time session health rules (AL-19)

## Summary of Changes
This pull request implements the real-time live anomaly detection engine and health scoring rules for AgenticLens as specified in AL-19:
- **Functional Core Live Anomaly Engine**: Deterministic, pure evaluation functions with zero external I/O or network dependencies (`backend/internal/domain/live_anomaly.go`):
  - `ConsecutiveToolErrors`: Triggers `HIGH` on >= 3 failures, `CRITICAL` on >= 5 failures, resets on success.
  - `ContextCeilingApproach`: Tracks context token usage against model context ceiling; triggers `WARNING` at >= 80% and `CRITICAL` at >= 100%.
  - `TokenVelocityRunaway`: Detects token velocity surges (>= 30k WARNING, >= 50k CRITICAL in a single turn).
  - `ToolThrashing`: Detects repeated tool failures and thrashing without progress (>= 5 WARNING, >= 8 CRITICAL).
  - Multi-key deterministic alert sorting (`CRITICAL` > `HIGH` > `WARNING`).
- **Live Engine Token Ingestion & SSE Alerting**:
  - Extracted token usage from SQLite message facts in `backend/internal/adapters/live/service.go`.
  - Evaluated trace anomalies and attached them to live trace and session tree nodes.
  - Broadcast `event: anomaly_alert` over Server-Sent Events with deduplication state.
- **HTTP Handlers & DTO Serialization**:
  - Response writers ensure empty slices serialize as `[]` rather than `null`.
  - Added comprehensive HTTP and SSE test coverage in `backend/internal/adapters/http/live_handlers_test.go`.
- **Frontend TypeScript Mirroring**:
  - Added strict, strongly-typed domain interfaces in `web/src/domain/live.ts` and exported from `web/src/domain/types.ts`.

## Verification
- `make check`: `go vet ./...` and `git diff --check` passed cleanly.
- `make test-backend`: Complete backend test suite passed with race detector enabled (`go test -race ./...`).
- `make build`: Binaries compiled cleanly for all supported targets.
- Zero emojis enforced across all files.
