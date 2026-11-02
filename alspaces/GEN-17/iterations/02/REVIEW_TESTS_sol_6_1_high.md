# GEN-16 iteration 02 test-effectiveness review

Outcome: CHANGES_REQUESTED, scoped to test effectiveness and the demonstrated production-boundary failure below.

Run: tests_sol_6_1_high; independent discovery. Requested/configured profile: gpt-6.1-sol, high reasoning effort. Source-observed runtime model/effort: unavailable. Reviewer responsibility: test effectiveness. No sibling personas, reports or conclusions were loaded.

Reviewed baseline: 99fbcd7be0d0ed05639e4a0cb3be8744a3d5371e. Reviewed candidate fingerprint: 785b903e14bb4af8923e054f2d8d5fb69f44210150b172d431d75aa5416d6f9a, independently recomputed from `git diff HEAD --binary -- backend web`, followed by sorted untracked backend/web path bytes, NUL, and file contents. Scope includes the untracked domain/persona.go. Requirements: CONTRACT.md plus controlling iterations/02/PLAN.md; iteration 02 supersedes the historical child fallback.

## Finding TS1: Hydration persona tests omit persisted events and miss stale-role reconciliation

Priority: P2. Classification: introduced required-contract omission in the new persona projection and its regression coverage; the merge helper causing the failure is an unchanged dependency. Discovering actor: test-effectiveness reviewer, gpt-6.1-sol/high. Verification actor: same reviewer, using an isolated Go overlay.

The new cases in `backend/internal/adapters/live/sqlite_test.go:1277` claim legacy-label recovery, explicit source precedence and conflict-to-unknown behavior, but the ledger at line 1242 has invocation facts only. With no `SessionFact.Events`, every hydrated event takes the append path. The assertions never exercise reconciliation against an existing event, a normal supported persisted-ledger boundary.

Trigger: a stored event has the same source session/event identity as a successfully hydrated message, and the stored role differs from current source resolution. `HydrateWorkspaceEvents` resolves the correct new role at `backend/internal/adapters/live/service.go:963`, then calls `ReconcileEvents` at line 842. Both stable-ID and unique-occurrence matching call `enrichEvent` (`backend/internal/domain/events.go:75`, `:106`). That helper at line 118 copies payload details and duration but preserves `recorded.Role` unconditionally.

Observed consequence: the API returns `IsHydrated=true`, but the existing event retains RootAgent, UnassignedSubAgent or the old named persona while a newly appended event has the resolved persona. The source-metadata precedence and conflict-to-unknown contract therefore depends on whether an event was already persisted. No disk rewrite is required to correct this read projection; unavailable-source baseline preservation remains required.

Disposable proof: copied sqlite_test.go outside the checkout, inserted the following baseline event immediately after the existing subtest assigns its stored invocation role, and used Go's `-overlay` option. Product and shared test sources were untouched.

```go
ledger.Facts.Sessions[0].Events = []domain.SessionEventFact{{
    EventID: "opencode:message:m_custom",
    Timestamp: time.UnixMilli(now).UTC().Format("2006-01-02T15:04:05.000000+00:00"),
    Turn: 1, Role: domain.AgentRole(tc.stored),
    EventType: domain.EventTypeTurnMessage, Status: "completed",
}}
```

Executed from backend:

- Control: `go test -count=1 ./internal/adapters/live -run '^TestHydrateWorkspaceEvents_SessionRoleMapUsesSessionID$' -v` passed, including all four new subtests.
- Probe: `go test -count=1 -overlay /tmp/gen16-test-review-iglv_t5j/overlay.json ./internal/adapters/live -run '^TestHydrateWorkspaceEvents_SessionRoleMapUsesSessionID$' -v` failed all four new subtests: RootAgent versus reviewer; UnassignedSubAgent versus coder; RootAgent versus Build; planner versus empty role.

Required correction: ensure the accepted source persona reaches matched hydrated events, and extend the real hydration fixture with persisted matching events so the above failure is caught. Preserve unmatched baseline rows, occurrence counts, unavailable-source fallback and input immutability. Cover the shared unique-occurrence path if the correction changes matching/projection semantics; no broad test matrix is requested.

## Other hypotheses assessed

- Extraction fabricates root/child placeholders or loses source persona: inspected Codex rollout parsing and both OpenCode schema-family producers with their real SQLite/rollout fixtures. The added invocation/event assertions cover blank children, explicit precedence, user transport exclusion and raw persona spelling. Existing v1 conflict coverage repeats the first candidate after disagreement, which detects accidental conflict reset in PersonaEvidence. No separate blocking finding.
- Unknown-role grouping loses totals or confuses ancestry: inspected ordering, role and thread grouping, workflow aggregation and existing/new assertions. The unknown-root/child case assigns root cost/turns distinct from child metrics, so it would expose role-based topology leakage. Mixed whitespace/blank buckets assert token, cost and invocation preservation. No separate blocking finding.
- Persona fallback reconstructs unchanged histories on every poll: traced tracePersonas ownership through source revision, trace cache and eviction. Existing pointer-reuse and in-place-update test reaches the cached production trace and checks persona invalidation. The new persona scan is performed inside the cached trace construction. No added unconditional unchanged-poll scan found.
- Unknown presentation/filtering hides blank events: inspected the changed renderers, MultiSelect integration and TraceWaterfall test. Its Unknown selection asserts both blank/whitespace rows remain and the named row disappears, then recovers No filter. Live tree and TicketDeepDive assertions reach presentation with blank personas and unavailable optional hydration. No separate blocking finding.
- Legacy reads fail when optional source data is missing: inspected existing fallback flow and the extended missing-source equality assertions for both historical labels. Those assertions preserve the full recorded events, rather than only checking hydration status.

## Verification and limits

Reused TEST_REPORT.md evidence for the frozen candidate: Go race suite across 12 packages, frontend 783 tests, final selected frontend fixture rerun, typecheck/lint and diff check. These results are execution evidence, not substitutes for this source/oracle review. Executed only the focused control and overlay probe above; no suites were redundantly rerun. No tests were deleted in the reviewed diff.

Final integrated make, canonical frontend build, embedded asset comparison and standalone binary serving remain parent-owned delivery gates. No browser geometry, real user source-session recovery, exhaustive mutation analysis or whole-repository correctness claim is made. The overlay is disposable local evidence; the snippet above records its complete fixture change. Only this assigned evidence file was written in the repository; no product implementation, commits, telemetry or publication performed.
