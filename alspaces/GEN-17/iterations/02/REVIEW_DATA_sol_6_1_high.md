# GEN-16 iteration 02 data and failure review

Verdict: CHANGES_REQUESTED.

Run: data_sol_6_1_high; independent discovery; assigned data integrity and failure handling. Requested and dispatch-acknowledged identity: gpt-6.1-sol, high reasoning effort, Codex provider. Observed runtime identity, usage and cost were unavailable to this reviewer and require parent source validation.

Reviewed base: 99fbcd7be0d0ed05639e4a0cb3be8744a3d5371e. Candidate fingerprint: 785b903e14bb4af8923e054f2d8d5fb69f44210150b172d431d75aa5416d6f9a, independently recomputed from `git diff HEAD --binary -- backend web`, followed by each sorted untracked backend/web path, NUL and file bytes. Requirements: CONTRACT.md with iteration 02 PLAN.md controlling the superseded child fallback. Product and tests remained read-only; no sibling persona or report was loaded.

## D-SOL-01: Reconciliation discards successfully recovered personas for persisted events

Severity: P2. Disposition: open required-contract gap in the new hydration projection, caused by interaction with an unchanged reconciliation dependency. This is not evidence of a newly introduced defect in generic reconciliation itself.

Locations: `backend/internal/adapters/live/service.go:963` resolves source evidence and assigns the recovered role; `service.go:842` then reconciles it with persisted events. `backend/internal/domain/events.go:118` enriches matching events without copying their resolved persona.

Trigger: a committed OpenCode session already contains an event, and the local source supplies a different authoritative persona for that session. The persisted role can be blank, a historical placeholder, or a named stale projection. The local event matches by stable session/event identity, as ordinary persisted source events do.

Consequence: hydration returns `is_hydrated=true` with fresh source content but retains the persisted persona on the matching event. A session with available `session.agent=planner` therefore still returns Unknown, RootAgent or coder for its previously persisted events. New unmatched events from that same source session receive planner. This violates the approved source precedence and consistent event/live projection requirements, even for a newly persisted blank role; it does not require historical ledger rewriting.

Proof: a disposable external Go overlay adds `TestReviewDataHydratedSourcePersonaForMatchedEvent` to package live_test without changing the checkout. Its isolated real SQLite fixture has session `source` with agent planner and assistant message `m`. The ledger contains `opencode:message:m` under the same session and a matching timestamp/turn. The three persisted-role variants blank, RootAgent and coder all fail the expected planner assertion after successful hydration. The control with no persisted event passes and returns planner. The source content is enriched, establishing that source access and event matching succeeded. Command from backend:

```text
go test -v -overlay /tmp/gen16-data-review-avgm9cfr/overlay.json ./internal/adapters/live -run '^TestReviewDataHydratedSourcePersonaForMatchedEvent$' -count=1
```

Observed: blank, RootAgent and coder subcases fail; unmatched-control passes. Overall exit status 1. The overlay and probe source remain outside Git in the stated temporary directory.

Required correction: apply the approved recovered persona to matched hydrated events as a read projection while retaining baseline facts on source failure, preserving unmatched occurrences and other-source data, and avoiding persistence side effects. Include a coherent existing hydration regression case with persisted matching Events, not only Invocations. The current `TestHydrateWorkspaceEvents_SessionRoleMapUsesSessionID` cases at sqlite_test.go:1277 cover source precedence only when no persisted Events reach reconciliation.

## Other consequential hypotheses investigated

- Explicit session metadata versus agreeing, conflicting, blank and user-message candidates: inspected PersonaEvidence, OpenCode v1/v2 extraction, Codex metadata resolution and their existing adapter cases. Explicit OpenCode metadata wins, conflicting assistant candidates remain unknown, and per-session event roles are patched consistently before extraction returns. Unknown roots/children retain parent linkage and invocation/token multiplicity. No additional blocking finding established.
- Cache staleness, failure and growth: persona evidence shares traceMu, source revision invalidation and trace eviction. Unchanged source reuses the cached history; only discovered sessions and one selected undiscovered lookup remain retained. Failed source reads do not return cached success. Reviewed existing in-place invalidation, bounded lookup and observer recovery cases. No additional blocking finding established.
- Unavailable optional hydration and legacy ledgers: inspected the committed baseline response, OpenCode failure exits, Codex hydration dispatch, HTTP fallback and legacy-role fixtures. Missing/closed sources retain baseline events; GET performs no write. Existing reconciliation retains unmatched occurrences and event multiplicity. D-SOL-01 concerns successful role recovery, not failure fallback.
- Aggregation omissions and topology: inspected blank-role ordering and normalized grouping through workflow, role and thread metrics, including unchanged DeriveMetricsLedger and parent-based root classification. Blank keys now remain present; child costs, turns and peaks do not depend on a synthetic persona. Raw Facts remain separate from mapped projections. Existing CLI mergeSessions replaces successful extracted session snapshots and retains unmatched sessions; that ownership path was unchanged. No additional blocking finding established.

## Verification reused and limits

Reused the tester's TEST_REPORT.md for this exact fingerprint: backend `go test -race ./...` passed all 12 packages; frontend complete suite passed 35 files and 783 tests, followed by the final selected fixture regression; final typecheck, lint and diff check passed. Inspected relevant fixture bodies rather than treating suite success as a source-precedence oracle. Executed only the isolated consequential probe above; no redundant full-suite run.

This verdict covers the assigned data/failure scope. The final integrated make gate, rebuilt embedded assets, standalone serving, source-session identity validation and publication belong to the parent. No source-session recovery claim is inferred from the synthetic fixture.
