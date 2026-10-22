# feat(live): implement live anomaly detection engine and real-time session health rules (AL-19)

## Summary of Changes
This pull request implements the real-time live anomaly detection engine and health scoring rules for AgenticLens as specified in AL-19, including post-recovery verification and type parity hardening:

- **Functional Core Live Anomaly Engine** (`backend/internal/domain/live_anomaly.go`):
  - Pure deterministic evaluators with value semantics and zero external I/O or network dependencies.
  - `ConsecutiveToolErrors`: Triggers `HIGH` on >= 3 failures, `CRITICAL` on >= 5 failures; resets on success.
  - `ContextCeilingApproach`: Tracks context token usage against model context ceiling; triggers `WARNING` at >= 80% (160k tokens) and `CRITICAL` at >= 100% (200k tokens).
  - `TokenVelocityRunaway`: Detects token velocity surges (> 30k tokens or > 40k delta for `WARNING`, > 60k tokens or > 80k delta for `CRITICAL`).
  - `ToolThrashing`: Detects repeated tool failures without forward progress (>= 3 for `WARNING`, >= 5 for `CRITICAL`).
  - `SortLiveAnomalies`: Deterministic multi-key alert sorting (`CRITICAL` > `HIGH` > `WARNING`, rule name ascending, timestamp descending, session ID ascending, trigger event ascending).
  - Extended domain models in `live.go` (`LiveMessageFact` token breakdown, `LiveSessionNode.Anomalies`, `LiveTraceView.Anomalies`, `LiveSessionTree.ActiveAnomalies`).

- **Live Engine Token Ingestion & SSE Alerting** (`backend/internal/adapters/live/service.go`):
  - Extracted token usage from SQLite message facts (`tokens.input`, `tokens.output`, `tokens.cache.read`, `tokens.reasoning`).
  - Evaluated trace anomalies and attached them to live trace and session tree nodes.
  - Broadcast `event: anomaly_alert` over Server-Sent Events with thread-safe deduplication.

- **HTTP Handlers & DTO Serialization** (`backend/internal/adapters/http/`):
  - Enforced response serializers so empty slices serialize as `[]` rather than `null`.
  - Refactored test assertions in `live_handlers_test.go` to eliminate untyped `interface{}` using concrete structs and `map[string]json.RawMessage`.
  - Added strict session ID path traversal and hostile character validation.

- **Frontend TypeScript Parity & Test Suite** (`web/src/domain/`):
  - Reconciled model parity in `web/src/domain/live.ts` (`parent_id?: string | null` on `LiveSessionNode`, `output_snippet?: string` on `LivePartFact`, `reasoning_content?: string | null` on `LiveMessageFact`).
  - Authored comprehensive unit test suite in `web/src/domain/live.test.ts` (22 tests) covering all anomaly severity tiers, token math, hierarchy nesting, and boundary cases.

Pointers:
- Contract: `workspaces/AL-19/CONTRACT.md`
- Summary: `workspaces/AL-19/SUMMARY.md`

## Verification
- `make check`: `go vet ./...`, `tsc --noEmit`, ESLint, and `git diff --check` passed cleanly.
- `make test`: All 11 Go backend packages passed with race detector enabled (`go test -race ./...`); all 22 frontend Vitest test suites passed (434 tests). Zero `t.Skip` instances.
- `make build`: Bundled web distribution embedded into Go binary; compiled multi-platform binaries for all supported platforms.
- Zero emojis enforced across all files.
