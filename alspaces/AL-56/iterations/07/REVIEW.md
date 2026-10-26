# AL-56 Iteration 07 Review

## Verdict: Changes requested

The ledger serialization change follows the approved direction: `PersistAnomalySummaries` removes event rows and evidence refs while retaining session invocation/tool facts, and DTO conversion omits those transient fields. Event endpoint fallback is empty and keeps stored summaries. The source selector is persisted, Codex and OpenCode hydration are dispatched explicitly, and event/anomaly associations are request-local. No `agentic_metrics.json` changes appear in the worktree; no generation or backfill was run. `git diff --check` passes. The parent reports full `make` passed; this review did not rerun it.

## Findings

1. **Codex hydration can report a partial trace as fully hydrated.** In `backend/internal/adapters/live/service.go:957-969`, discovery errors for any root invocation are skipped. If another root yields rollout paths, the endpoint returns `IsHydrated: true` with only that root's events. The plan requires a missing source tree to fail closed to an empty unhydrated response. Return the fallback if any expected root cannot be discovered, or define an explicit completeness rule and report partial hydration distinctly. Add a fixture with two root invocation IDs where only one exists under `CODEX_HOME`.

2. **Generating after loading a compact ledger can preserve old summaries and omit new ones.** `mergeSessions` replaces only sessions found in the currently extracted source; other cached session records (including their anomaly summaries) remain. `PersistAnomalySummaries` and `DeriveMetricsLedger` both use `hasPersistedAnomalySummaries` to skip detection whenever any persisted summary exists. In the Codex path, extraction selects the supplied nonce/session root, not all cached root IDs. Thus a cached summary for an older root can prevent detection of anomalies in the newly extracted full event stream; the new session's summary list is empty and the stored/output metrics miss its anomalies. Recompute/merge summaries intentionally from complete current source facts and preserved cached summaries for source sessions that were not re-extracted, or otherwise clear stale cached summaries before projection. Add a test with one cached summarized root plus a newly extracted root that has a detectable anomaly.

## Verified

- Canonical DTO construction no longer serializes `sessions[].events` or anomaly evidence; the domain canonicalizer also strips both while retaining per-session tool and invocation facts.
- Legacy event/evidence fields remain readable for normalization, and `LoadTicketWorkspace` calls the canonicalizer before returning facts.
- The endpoint's source-less fallback contains zero events and retains `AnomalyFacts`; unknown source values also remain unhydrated rather than guessing a harness.
- Codex hydration selects rollout trees by root invocation ID and parses discovered paths; OpenCode uses its existing SQLite route. The API response carries explicit local Codex/SQLite source identifiers.
- A source-backed Codex hydration fixture exists, and source-less ledger fallback is covered. I did not find a Codex endpoint fixture covering child rollout hydration or missing one among multiple roots; the missing-root case is also the fail-closed issue above.
- The test report for iteration 07 is not present yet. The workspace status and diff name queries show no ticket ledger changes; no backfill or generation command was run.

## Review basis

Reviewed `iterations/07/PLAN.md`, the working diff in CLI generation, DTOs, repository normalization, migration/backfill, HTTP/live hydration, Codex discovery/parser use, UI labeling/anomaly association, and `TECHNICAL_GUIDELINES.md`. No implementation or test files were changed during this review.

## Round 2 implementation response

- Addressed the missing-child partial hydration finding by validating that every invocation ID, including child IDs, appears among parsed transcript thread IDs. Any absent/unreadable expected thread returns the empty unhydrated fallback.
- Added `TestHydrateWorkspaceEvents_CodexFailsClosedWhenChildRolloutIsMissing`, with one root, two expected children, and one missing child rollout.
- Re-ran `make` successfully after the change. Updated detailed results are in `iterations/07/TEST.md`. No ledger generation, migration, or backfill was run.

## Round 2 Verdict: Changes requested

Both round 1 findings are addressed:

- Codex root discovery now returns the empty unhydrated fallback if any root invocation cannot be discovered or yields no rollout paths. The two-root/one-missing regression verifies that partial root discovery is not presented as a complete trace.
- `RefreshAnomalySummaries` recomputes anomalies for sessions with current events and ticket aggregates from merged tool facts, while retaining cached per-session summaries for sessions with no current events. The new regression exercises a cached root summary plus a newly extracted root with a retry anomaly and verifies aggregate refresh.

One source completeness edge remains: **a missing child rollout can still be presented as fully hydrated.** `hydrateCodexEvents` only verifies that each root has a non-empty `DiscoverRolloutPaths` result. Discovery includes children it can read, but does not require every child invocation ID recorded under that root in the ledger to appear in the discovered tree. If the root rollout exists and one child rollout has been removed or is unreadable, discovery can return only the root (or the other children), and the endpoint will return `IsHydrated: true` with an incomplete ticket trace. Before accepting hydration, compare discovered/parsed thread IDs with the expected root and child invocation IDs from that root's ledger subtree; if any expected child is absent, return the empty unhydrated fallback. Add a root-plus-two-children fixture with one child rollout missing.

The TEST report records `make` passed and confirms no ledger changes or generation/backfill. `git diff --check` passes; this review did not rerun tests or mutate implementation/telemetry.

## Round 3 Verdict: Approved

The child completeness issue is resolved. `hydrateCodexEvents` now collects every invocation ID in the ledger, parses all discovered root trees, and checks that each expected ID appears as a parsed transcript thread ID (falling back to the transcript session ID). A missing root, missing child, discovery failure, parse failure, or empty tree returns the empty unhydrated response. The new regression includes a root, two expected children, and one missing child rollout; it verifies the present root/child do not produce a partial hydrated response.

The cached-summary refresh regression remains in place and checks both persistence and metrics projection for old cached sessions, newly extracted event sessions, and updated ticket aggregates. Canonical output continues to remove events and evidence refs; the source selector and aggregate/session facts remain serialized. The TEST report records the repeat full `make` pass and clean ledger scope. I found no remaining material issue in the iteration 07 diff. This review did not run generation, backfill, migration, or tests, and did not change ledgers or implementation files.
