# AL-53 Iteration 01 Review

## Verdict

APPROVED. No actionable correctness or contract-alignment defects found in the reviewed frontend implementation and regression tests. Final delivery still requires the parent's integrated `make`, embedded-asset comparison, and standalone binary verification.

## Review boundary and identity

Reviewed against `CONTRACT.md`, `PLAN.md`, `.agents/agent/reviewer.md`, `TECHNICAL_GUIDELINES.md`, and the applicable regression checklist. Baseline commit: `85a2af94b5036b0918aaf9a7a89aed3a990f2650`.

Read complete changed production/test bodies and relevant unchanged data accessors, loading/enrichment flow, thread behavior, TypeScript configuration, and CSS containment rules. Reviewed paths: `SkillsSubagentsPanel.tsx`, `ThreadViewer.tsx`, `domain/formatters.ts`, `index.css`, `views/TicketDeepDive.tsx`, and `views/TicketDeepDive.test.tsx` under `web/src/`.

The reviewed frontend snapshot SHA-256 is `6f7731e119b383e930f46ccc561f40215a4d8e52c75fd635f915eacaca52445a`. Its input is each lexically sorted repository-relative path, a NUL byte, its complete file bytes, and a NUL byte. This includes the new untracked panel file, unlike a tracked-only Git diff hash.

## Axis 1: Specification and contract alignment

- The dedicated collapsible panel renders skills and delegations independently from persisted `ledger.facts` via existing defensive accessors. Both absent and empty optional supported-v4 collections reach the specified empty message. No legacy schema acceptance or backend changes were introduced.
- Skill names, invocation turns/timestamps, roles, recorded boolean status, and verbatim error text remain available. Error-like text does not overwrite successful status. Repeated facts remain separate rows and unknown roles preserve their source labels.
- Delegations show source description, parent turn, optional duration/cost, and child identity. Missing measurements explicitly display `Unavailable`; measured zero remains zero. No aggregate-derived cost or fabricated session destination is introduced.
- Exact available child IDs receive native keyboard actions. Navigation expands and focuses the matching thread while retaining other expanded threads. A fresh request object supports repeated jumps to the same child after manual collapse. Unavailable children retain their full IDs and explanatory text without a broken action. Ticket reload clears the request; optional event enrichment does not replace panel facts.
- Semantic tables, captions, scoped headers, labelled regions, expanded/control associations, focusable rows, and native buttons provide keyboard access. Existing first-thread expansion and manual toggling remain intact. New wrapping/overflow rules preserve source text and use existing table containment without changing tooltip containers.

## Axis 2: Engineering standards

The change stays within presentation and pure formatter responsibilities. No Go, persistence, extraction, configuration, manifests, or network API changes. TypeScript strict, noImplicitAny, and exactOptionalPropertyTypes remain enabled. Inspected changed TypeScript files contain no any types; no emojis were introduced. Test fixtures use local synthetic variants rather than live ticket artifacts or developer session stores.

The tester extends the existing view suite with five runtime cases, including two parameterized optional-collection cases. Assertions exercise the rendered production path, full source text, repeated facts, optional metrics, exact sibling-sensitive child matching, repeat keyboard jumps, collapse semantics, and pending/failed optional enrichment. No test removals or silent skips were introduced. These are behavior guards rather than fixture literal or accessor-only tests.

## Verification evidence and limits

Reused reported results for the unchanged reviewed snapshot: coder typecheck/lint and 131 existing view/formatter tests; tester `npm test -- src/views/TicketDeepDive.test.tsx` with 62 passing tests, plus `npm run typecheck` and `npm run lint`. The tester additionally reports full `npm test` passed 34 files / 717 tests with no skips in 32.68s. After the final test-only Tab-focus assertion, all five AL-53 cases passed in a filtered rerun (57 unrelated view cases explicitly excluded), followed by passing typecheck/lint/diff checks. Details are recorded in `TESTING.md`; parent owns the required integrated `make` after implementation commit/upstream synchronization. Reviewer independently ran `git diff --check` successfully and inspected production/test bodies rather than rerunning suites without a concrete concern.

Browser geometry, narrow-viewport dimensions, actual scroll distance, and visual contrast were not measured. DOM interaction tests establish keyboard expansion and focus; CSS inspection establishes intended containment and wrapping, not browser geometry. Fresh embedded asset and standalone host binary evidence must be added by the parent before handoff. Approval does not claim those pending delivery boundaries passed.

## Parent delivery verification

Integrated `make` passed on implementation commit `f58657c` after an already-current origin/main merge: Go vet/race suites, frontend typecheck/lint, all 717 tests in 34 files, canonical frontend assets, and four standalone platform builds. The embedded-web-asset Go test passed. Fresh frontend and embed input trees matched exactly, and a temporary standalone Linux amd64 server served the same index/JS/CSS bytes, including new panel strings. Server was stopped. Generated assets remain ignored. No application/test changes followed review; later edits are ticket lifecycle, evidence, and telemetry only. Browser geometry remains unmeasured.

## Routine main integration review: MGT-1 F1 and F5 (2026-10-31)

APPROVED. Independently reviewed the resolved merge of actual `origin/main` at `6bc1064943ae22e1b6586a6d145543c6a9cd23f2` into AL-53 head `57058521c498b9e376bbab5ec2355adec63d7729`, against the iteration 01 integration addendum and management findings F1/F5. This is a same-iteration integration correction, with no new feature acceptance boundary.

Reviewed the complete combined navigation/view test changes and compared both starting heads. The six-path frontend snapshot, using the path/NUL/content/NUL SHA-256 method defined above, is `de2cb6d976b8cf09b8dc3344a88b4653df32609ca7872b1df48bd7e46e0de5f0`. Merge conflict list is empty; independent working-tree and staged `git diff --check` both pass.

F1 is resolved in the reviewed working result. `ThreadViewer.tsx` matches actual main exactly and has one `focusSession: { id }` contract. Both the waterfall's `onSelectSession` and the panel's `onJumpToSession` call one view handler that creates a fresh focus request for each activation. The effect validates exact membership, expands the requested ID without closing other threads, focuses its exact header ref and dispatches scroll. Full source IDs remain available in the panel text and both controls' accessible names. Native Enter/Space actions work repeatedly after manual close. Existing trace badge propagation guards preserve parent event selection semantics.

The new view lifecycle case coherently extends the upstream committed-badge test rather than dropping it. Its original collapsed-child, expansion, focus, visible body and absent-dialog assertions remain, strengthened by a long source ID, prefix-sharing sibling, error-task parent, both controls in the same mounted view, repeated keyboard activation and four scroll calls. Ticket switching with an active child request and parent failure inspector proves focus/expansion/selection reset; late old-ticket enrichment cannot restore the old parent trace or navigation badge. The original five AL-53 cases and their meaningful assertions remain unchanged. No duplicated isolated navigation suite was introduced.

Upstream live refresh, reading-anchor behavior and fleet skill features are preserved: focused Git comparisons against actual main show no changes to ThreadViewer, LiveTraceInspector and its tests, LiveObservability and its tests, useLiveTrace, liveTraceRefresh, FleetOverview, or TicketComparisons. The overall production diff against main is confined to AL-53 presentation additions. No speculative AL-43 feature or AL-55 implementation was incorporated.

F5 is resolved in planning metadata: the index removes completed AL-53 and moves AL-55 to Ready with no blockers; its specification agrees on Ready/None and preserves the remaining replanning/implementation scope. Completed upstream ticket removals and done moves remain present in the integrated result.

Verification reused for this stable snapshot: tester reports `npm test -- src/views/TicketDeepDive.test.tsx src/components/TraceWaterfall.test.tsx` passed 146 tests across both files, no skips, 20.60 seconds (view: 63 tests), with passing typecheck/lint/diff checks; details are in `TESTING.md`. No duplicate suite run was warranted by the read-only audit. Parent still owns merge commit, required full `make`, fresh embedded-asset comparison, standalone binary checks and final source-backed telemetry. This approval establishes the reviewed integration correction; browser geometry and physical scroll distance remain unmeasured.

## Parent integration gate failure

Full make on merge commit `62a7963bb421b3111025fc537bcd91d3d6dc5fe9` failed in inherited HTTP/SSE test `TestLiveStreamAndTraceRecoverInPlaceSQLiteUpdate` awaiting its initial trace (line 805). All other race packages and preceding static/build checks passed. Focused repeats passing do not clear this failure. Tester deterministically confirmed an unchanged main subscription startup gap; see TESTING.md. Combined-navigation review remains approved, but PR readiness is HELD pending manager/AL-47 routing and a fresh complete integration gate. No Go fixes or synchronization bypass are included in AL-53.

## Concrete blocker reproduction and ownership hold

Tester implementation is retained outside Git at `/tmp/al53-sse-ordering-proof.go`; diagnostic output is `/tmp/al53-sse-ordering-proof.log`. From backend, the diagnostic used a disposable `.al53-sse-ordering-proof/main.go` copy and `go run -race ./.al53-sse-ordering-proof`, then removed that directory. It uses the unchanged concrete HTTP handler, LiveService and real SQLite; pauses the handler initial GetSessionTree until an independent monitor subscriber observes the poller's first trace, then allows HTTP subscription. The stream reports one session_tree and zero trace_update before deadline. This controlled interleaving reproduces the gap without changing production source or the intended mutation fixture. See TESTING.md for precise setup. A successful diagnostic exit proves the defect, not readiness. Raw rollouts/transcripts remain outside Git.

MGT-1 confirmed receipt and routed the production fix/regression to the existing AL-47 owner in a new follow-up PR. AL-53 stops backend investigation and awaits an explicit verified fix-merge instruction before another fetch/integration/full-make attempt. No fixture-only bypass or repeated full-gate retry on unchanged defective main.

Independent remaining checks on the unchanged integrated product snapshot passed: `make test-frontend build` (35 files / 741 frontend tests, four platform binaries); embed input matches frontend build exactly; temporary standalone Linux amd64 server serves fresh index/JS/CSS byte-for-byte with the new panel bundled. Server stopped. These checks do not clear the failed full gate. Browser geometry remains unmeasured.
