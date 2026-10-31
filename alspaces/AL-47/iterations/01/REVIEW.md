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


## Startup subscription correction review (2026-10-31)

Correction verdict: APPROVED. This is an appended review of the user-preapproved bounded repair to the existing AL-47 SSE contract. The original feature approval and verification above remain historical evidence, not evidence that the startup schedule was previously safe. Baseline: `6bc1064943ae22e1b6586a6d145543c6a9cd23f2`, branch `fix/AL-47-sse-startup-subscription`. Reviewed production-plus-test diff SHA-256: `bd7cdce1e044a17425e92be6eda58a38eff3baa69f0c50b8fdbb1fd396c2637a` from the two HTTP paths, matching the tester's final diff.

### Contract alignment and ownership

Read the complete changed handler/test bodies, their existing SSE protocol/cancellation/concurrent-subscriber peers, and concrete LiveService subscription and broadcast ownership. The production change registers the non-nil subscriber and defers its cleanup before exposing success headers or performing the initial tree lookup. The handler still writes and flushes the initial session_tree before draining buffered events. Thus a real publish during that lookup is queued for this already-registered connection, rather than permanently hidden by unchanged-trace deduplication. No service/cache/frontend/schema or replay behavior was changed.

There is one deferred unsubscribe on each registered path, including cancellation while initial lookup is pending and continued streaming after an initial snapshot error. Unsupported writers return before subscribing; nil service preserves its prior headers-and-context behavior. Existing payload formatting and flush points are unchanged. Concrete broadcasting retains its bounded buffered channel and nonblocking send, so earlier registration cannot block the poller. Bounded-buffer overflow and historical replay remain outside this repair's guarantee; the original REST recovery plane remains intact.

The regression extends the existing real WAL SQLite case without removing its UPDATE/REST assertions or increasing its request timeout. Its forwarding wrapper controls the HTTP initial-tree lookup while delegating concrete subscription, source extraction and broadcast. An independent concrete subscriber proves the poller has published before the gate is released; the HTTP connection must already be registered and then receive initial tree first followed by the unchanged initial trace. The coder's temporary old-order experiment failed with `concrete service published initial trace before HTTP stream subscription`, establishing the new ownership oracle rejects the former production ordering. Reviewer reused that attributed experiment rather than claiming an independent rerun. The tester's additional lifecycle table exercises early cancellation and tree-error recovery through actual subscriber cleanup/broadcast ownership. No synthetic broadcast replacement, skip, test-only synchronization bypass, arbitrary sleep or timeout enlargement was introduced.

### Engineering and verification

No blocking findings on either contract alignment or engineering standards. The fix remains in the HTTP imperative shell and introduces no domain I/O, outer-layer dependency, any/interface{}, emojis, or unrelated refactor. Tests use isolated fixtures and do not read ticket artifacts or developer source stores. Initial snapshot error behavior is preserved rather than broadened into a new protocol. Original AL-53 failed full gate and controlled proof remain recorded in TESTING.md; later green checks do not erase that evidence or prove its exact original scheduling was instrumented.

Reused final tester evidence for this identical diff: startup plus both lifecycle cases passed `-race -count=10`; complete HTTP/live packages passed `-race -count=1`; canonical assets had been built. Reviewer ran `git diff --check` successfully and inspected relevant production/test bodies instead of repeating unchanged suites without a concrete concern. Parent still owns actual integrated full make, rebuilt embed equality, standalone version/help checks, telemetry preservation, new PR delivery and manager handoff. Merged PR #76 cannot deliver this correction. No deployment, generic late-subscriber replay, browser geometry, or exact SQLite pre/post-generation write interleaving claim is added.


## Parent startup correction integration gate (2026-10-31)

The correction was independently APPROVED and committed as `c72eb25317e8b6317448a08100364940402cd319`, based on current main `6bc1064943ae22e1b6586a6d145543c6a9cd23f2`. Fresh upstream integration was already current; no product changes followed review. Full `make` passed at this actual integrated head: Go vet and all race suites, frontend typecheck/lint, 35 files / 736 frontend tests, canonical frontend assets and four platform binaries. Log: `/tmp/al47-startup-make.log` outside Git. Existing MetricProgression act warnings remain unrelated. No tests were skipped. `diff -qr web/dist backend/internal/adapters/web/dist` passed; embedded `TestGetFS` ran in the full gate. Host version JSON and built generate help passed. A bounded standalone HTTP smoke test served byte-identical index/JS/CSS from isolated configuration with Node absent from PATH, then exited cleanly. Frontend/domain/live-service sources are unchanged by this follow-up; application/test diff is limited to the two HTTP paths. Test diff fingerprint remains `bd7cdce1e044a17425e92be6eda58a38eff3baa69f0c50b8fdbb1fd396c2637a`.

The original AL-53 gate failure at merged `62a7963bb421b3111025fc537bcd91d3d6dc5fe9` remains evidence: `/tmp/al53-integration-make.log` failed the initial readTrace before UPDATE. Retained proof source `/tmp/al53-sse-ordering-proof.go` and output `/tmp/al53-sse-ordering-proof.log` demonstrate the valid schedule; they do not establish the uninstrumented failed run's exact scheduling. The corrected controlled regression independently verifies the ordering and actual concrete service/SQLite delivery. No timeout increase, skipped test, test-only startup bypass, or generic replay guarantee was introduced.

Original feature PR #76 merged as `5da845a7c244a6b2582102c0dff0d221844d9bfc`; only the new follow-up PR will deliver this correction. The completed specification remains in plan/done; no completed ticket index entry was resurrected. Manager owns merges; this owner neither merges nor enables auto-merge. Original feature contract, iteration evidence and five ledger source identities are preserved. Final own-root Codex extraction uses conventional fix classification; the ticket's original feature evidence remains historical rather than reclassified as proof of this correction.
