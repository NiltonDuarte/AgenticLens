# AL-56 Iteration 07 Verification

## Scope checked

- Canonical v4 persistence removes session event rows and anomaly evidence references while retaining the explicit harness source, anomaly summaries, invocation facts, and per-session tool totals.
- Event hydration selects the source from the ledger's explicit `source` field. Codex rebuilds transient events from the selected `CODEX_HOME` rollout tree; OpenCode continues to hydrate from SQLite. Anomaly associations are rebuilt for the response.
- A source-less legacy ledger returns no hydrated events and retains its stored anomaly summary.
- No implementation files or telemetry ledgers were edited by this tester. No telemetry generation, migration, or backfill command was run.

## Commands and outcomes

1. Focused backend suites:

   ```bash
   cd backend && go test ./internal/domain ./internal/adapters/codex ./internal/adapters/live ./internal/adapters/http ./internal/domain/migration
   ```

   Passed all five packages.

2. Full repository verification:

   ```bash
   make
   ```

   Passed: `go vet ./...`, `go test -race ./...`, frontend typecheck and lint, Vitest (31 files, 722 tests), frontend production build, and Linux/macOS amd64/arm64 binary builds. npm reported six existing audit advisories during dependency installation (3 moderate, 1 high, 2 critical); the command completed successfully.

3. Whitespace validation:

   ```bash
   git diff --check
   ```

   Passed.

4. Telemetry write-scope audit:

   ```bash
   git status --short -- 'alspaces/*/agentic_metrics.json'
   git diff --name-only -- 'alspaces/*/agentic_metrics.json'
   ```

   Both returned no paths. No other ticket ledger changed. No workspace-wide backfill, telemetry generation, or migration was run.

## Relevant regression coverage reviewed

- `TestPersistAnomalySummariesDropsEventsAndEvidenceButKeepsFacts` checks that canonical facts retain the Codex source, invocation rows, tool totals, and anomaly summary while event facts and evidence refs are absent.
- `TestHydrateWorkspaceEvents_CodexUsesExplicitSourceAndCODEXHome` checks local Codex rollout hydration and stable transient event IDs.
- `TestHydrateWorkspaceEvents_SourceLessLedgerStaysUnhydrated` checks that source-less ledger data is not guessed or returned as hydrated events, while stored summaries remain available.
- Existing OpenCode SQLite hydration tests continue to exercise rebuilt event details and anomaly associations; `TestHydrateWorkspaceEvents_FallbackWhenSessionNotFound` confirms empty unhydrated fallback when local rows are unavailable.
- Codex/OpenCode retention parity coverage compares complete-stream anomaly summaries and transient event associations and checks that Codex's canonical ledger omits events/evidence.

## Limitations

- Hydration depends on the user's local source files/database still being available. A copied committed ledger alone contains summaries and aggregate invocation/tool facts, not a detailed event timeline.
- Codex hydration errors or missing rollout paths degrade to an empty unhydrated response, preserving stored anomaly summaries.

## Reviewer follow-up verification

- Codex hydration now returns the empty unhydrated fallback if any root invocation has a discovery error or no rollout paths. `TestHydrateWorkspaceEvents_CodexFailsClosedWhenAnyRootIsMissing` covers two roots with one missing rollout.
- Anomaly refresh now recomputes session summaries for sessions with current source events, preserves cached session summaries only when their sessions have no current events, and recalculates ticket-level aggregates from the complete merged tool facts. `TestPersistAnomalySummariesRefreshesNewSessionsAndAggregates` checks both persistence and metrics projection with a cached old root plus a newly extracted anomalous root.
- Independently reran the exact regressions:

  ```bash
  cd backend && go test ./internal/adapters/live -run 'TestHydrateWorkspaceEvents_CodexFailsClosedWhenAnyRootIsMissing' -count=1 -v
  cd backend && go test ./internal/domain -run 'TestPersistAnomalySummariesRefreshesNewSessionsAndAggregates' -count=1 -v
  ```

  Both tests passed. The first confirms that one discoverable Codex root cannot produce a partial hydrated response when another root is missing. The second confirms the old summary survives, a retry summary is recomputed for the new event session, ticket aggregate tool failure totals reflect merged facts, and metrics projection includes all three.
- Re-ran `make` after these changes. It passed `go vet ./...`, race-enabled Go tests, TypeScript typecheck/lint, all 722 frontend tests, frontend bundle generation, and four platform builds. `git diff --check` passed.
- No metrics ledger was generated or modified, and no migration/backfill command was run.

## Round 2 reviewer follow-up

- Codex hydration now compares every invocation ID in the ledger, including children, against the thread IDs parsed from all discovered rollout files. A discovery error, absent root, absent child, or parse error returns an empty unhydrated response.
- `TestHydrateWorkspaceEvents_CodexFailsClosedWhenChildRolloutIsMissing` builds a root with two expected child invocations and only one child rollout; it verifies the available root and child do not produce a partial hydrated trace.
- Re-ran `make`; it exited successfully with Go vet, race-enabled Go tests, TypeScript typecheck/lint, 31 frontend files / 722 tests, production bundle generation, and four platform builds. `git diff --check` passed.
- Ledger scope audit remains clean: no metrics JSON changed, and no generation/backfill/migration command ran.

### Independent tester rerun

- Re-ran all three focused regressions on the current tree:

  ```bash
  cd backend && go test ./internal/adapters/live -run 'TestHydrateWorkspaceEvents_CodexFailsClosedWhenChildRolloutIsMissing|TestHydrateWorkspaceEvents_CodexFailsClosedWhenAnyRootIsMissing' -count=1 -v
  cd backend && go test ./internal/domain -run 'TestPersistAnomalySummariesRefreshesNewSessionsAndAggregates' -count=1 -v
  ```

  All passed: both missing-root and missing-child cases returned empty unhydrated responses, and cached-summary/new-session refresh retained/recomputed summaries and aggregate totals.
- Ran `make` again after the child completeness change. It passed Go vet, race-enabled Go tests, TypeScript typecheck/lint, 31 frontend test files / 722 tests, frontend bundle generation, and four platform builds. `git diff --check` passed.
- The metrics-ledger status and diff path audits returned no paths. No telemetry generation, migration, or backfill was run.
