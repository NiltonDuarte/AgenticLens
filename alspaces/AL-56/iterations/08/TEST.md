# Iteration 08 Test Notes

## Scope

Canonical anomalies now retain only `rule_id`, `severity`, and structured context. Display text is derived at the CLI and UI boundaries. Legacy schema-v4 title/description keys are accepted only during decoding and omitted on re-serialization.

## Focused verification

- `go test ./internal/domain ./internal/adapters ./internal/adapters/live ./internal/adapters/http` (from `backend/`): passed.
- `npm run typecheck` (from `web/`): passed.
- Focused Vitest coverage for anomaly formatting, anomaly drawer, trace waterfall, and ticket deep dive: 4 files / 148 tests passed.
- `make`: passed, including `go vet`, all race-enabled Go tests, frontend lint/typecheck, all Vitest suites (31 files / 728 tests), Vite bundle, and multi-platform Go builds.
- `git diff --check`: passed.

## Data scope

No telemetry generation, migration, or workspace backfill was run. No `alspaces/*/agentic_metrics.json` was modified.

## Independent tester pass

- Confirmed canonical domain `Anomaly` has only `rule_id`, `severity`, structured `context`, and request-local optional evidence; it has no title/description fields. DTO serialization tests verify those presentation keys are absent. Strict DTO decoding explicitly consumes legacy `title` and `description` keys, and a decode/encode regression confirms they are dropped.
- Confirmed display formatting coverage for retry streak, high failure rate, context spike, context ceiling breach, subagent rework loop, and incomplete-context fallback. Backend formatter is used by CLI and markdown summary output; UI formatting is used by the anomaly drawer and ticket deep-dive summary, while waterfall labels resolve through those surfaces. Fleet-wide rule distribution uses its static rule descriptor labels.
- Anomaly detection tests and the focused suites passed without changes to rule IDs, thresholds, severity, context values, or counts.
- Commands run:

  ```bash
  cd backend && go test ./internal/domain ./internal/adapters ./internal/adapters/live ./internal/adapters/http -count=1
  cd web && npx vitest run src/domain/anomalies.test.ts src/components/AnomalyDrawer.test.tsx src/components/TraceWaterfall.test.tsx src/views/TicketDeepDive.test.tsx --reporter=dot
  make
  git diff --check
  ```

  Backend focused suites passed. UI focused suites passed (4 files / 148 tests). Full `make` passed `go vet`, race-enabled Go suites, frontend typecheck/lint, 31 Vitest files / 728 tests, bundle build, and all four platform builds.
- Final metrics-ledger status and diff audits returned no paths. No telemetry generation, migration, or backfill was run.


## Final scoped ledger verification

After the code verification, generated telemetry for AL-56 only. The final ledger contains 26 sessions and 50 anomaly summaries; recursive checks found no anomaly `title`/`description` and no `events`, `evidence`, or `event_id` fields. It is 1,638 lines / 43,480 bytes, versus the iteration 07 baseline of 1,691 lines / 53,208 bytes (53 fewer lines / 3.1%, 9,728 fewer bytes / 18.3%) while including one additional session. Git status confirmed AL-56 was the only ticket ledger changed; no other ticket was backfilled.
