# AL-47 iteration 01 review

Status: APPROVED

## Reviewed boundary and evidence

Reviewed the working-tree implementation against baseline `1d20117711aab96a230679d3cdb9406853fe3f04`, the approved CONTRACT.md and PLAN.md, TECHNICAL_GUIDELINES.md, and the applicable polling/asynchronous, UI-state, optional-data, and embedding regression checklist sections. Read the changed production functions, new hook/helpers, changed test bodies, and relevant existing consumers/helpers. No application or test edits were made by reviewer.

The tracked production diff at review has SHA-256 `5abd9d61dc619e5ffce340c3957296b9db517264c2cc1896b83939ec5fc475c8` for the five tracked production paths (Go domain/live service/SQLite and view/inspector); the new `web/src/hooks/useLiveTrace.ts` and `web/src/domain/liveTraceRefresh.ts` were also reviewed directly. This identifies the pre-integration implementation, not a committed or upstream-integrated head.

## Axis 1: specification and contract

No blocking findings.

- Existing SSE protocol is preserved. Selected running/active traces gain a 2000 ms REST recovery plane in connected and fallback modes. Selection, manual requests, interval/revision triggers, and resume use one owner with one in-flight request and one pending follow-up.
- Selection, pause, close, and unmount invalidate old owners. SSE acceptance advances a push version so a preceding REST success/error cannot overwrite it. Successful tree traffic cannot erase the separate trace error. Last valid same-session messages remain mounted while refreshing or failing.
- Inactive/missing selections stop intervals. Active-to-inactive revision changes can request one reconciliation. Explicit paused manual refresh is retained without restarting automatic work.
- Stable message/session keys preserve local reasoning/failure/order state. Reading-anchor adjustment extends to existing-message growth and preserves existing ordering/navigation controls. Loading does not replace accepted inspector content.
- SQLite `data_version` stays tied to a retained observer connection; replacement increments its identity. Serialized cache access checks generation before and after reconstruction. A write during reconstruction prevents cache installation and invalidates the next lookup. Canceled/failed observations do not report cached success. Discovered entries plus one on-demand lookup bound snapshot retention; broadcaster snapshots are removed when sessions leave the tree.
- On unchanged input, tree status is still derived from the current clock, but trace reconstruction and trace JSON serialization are avoided. Meaningful trace comparison remains complete. A database write can invalidate all visible traces, so this is source-generation reuse rather than per-row incremental work; the contract explicitly permits that cost.
- Reviewed actual cached-slice ownership: production consumers read message/part slices. HTTP DTO normalization replaces top-level nil slice fields on its value copy and does not mutate cached nested facts. No current shared-cache mutation defect was found.

## Axis 2: engineering standards

Go comparison is pure and imports no I/O; SQLite observation/synchronization remains in adapters. Network lifecycle is confined to the frontend hook/view and pure status/comparison helpers remain in frontend domain. No API/schema/status/ordering expansion or unrelated application cleanup was introduced. The new TypeScript paths contain no `any`; changed implementation introduces no `interface{}` or emojis. Revision errors are contextualized and returned. Existing optional child-link lookup tolerance is preserved as required by the grounded plan; this review does not claim to repair that pre-existing partial-data boundary.

Added tests exercise production behavior with independent expected values. Real WAL fixtures verify in-place changes without timestamp/count changes, legacy optional columns, cache reuse/eviction, canceled observer recovery, and SQLite-to-SSE-to-REST delivery at a non-default poll cadence. Controlled frontend promises/timers cover stale responses, coalescing, errors/recovery, pause/resume, completion/stall/missing/close cleanup, fallback updates, DOM identity, reasoning state, and scroll arithmetic. No tests were removed or silently skipped.

## Verification reuse and remaining limits

Reused tester evidence in TESTING.md: full `go test -race ./...` passed; full frontend suite passed 34 files/720 tests before two final parameter cases, followed by final focused 49-test pass; final TypeScript typecheck/lint passed. Reviewer ran `git diff --check` successfully and inspected source/test bodies rather than duplicating unchanged suites. The new TypeScript no-any search returned no matches.

Parent still must commit/integrate upstream and run full `make` with the final two frontend cases, then verify rebuilt embedded assets and standalone execution. Approval does not establish those pending gates. Real-browser geometry and deterministic injection of a write exactly between revision observations remain unverified; the pre/post-generation branch was inspected, and race-suite success alone is not claimed as a proof of that interleaving. No measured latency or memory benchmark is claimed.

Verdict: APPROVED for integration and the remaining parent-owned delivery checks.


## Parent integration verification

Reviewed implementation committed as `f531208`; upstream merged cleanly at `6968687122d6146010e4255ae46f3ea0f1b6936b`. Upstream modified planning artifacts only, with no application/test changes. Full post-integration `make` passed, including Go vet/race tests, TypeScript typecheck/lint, all 722 frontend tests, canonical frontend embed build, and four platform binaries. Existing MetricProgression act warnings remain; no tests skipped. Embedded `TestGetFS` ran successfully. Generated web/dist and backend/internal/adapters/web/dist match byte for byte. A bounded standalone linux/amd64 HTTP smoke test served identical index plus JS/CSS from an isolated workspace with no Node in PATH and exited cleanly. Built version JSON and generate help passed. Source and test code were unchanged after this verification. Prior browser-geometry and precise concurrent-injection limits remain.
