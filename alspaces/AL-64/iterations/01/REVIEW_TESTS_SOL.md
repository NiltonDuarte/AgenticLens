# AL-64 independent test-effectiveness review

Outcome: **CHANGES_REQUESTED** for this focus. First-pass discovery, reviewer-tests, requested/inherited gpt-6.1-sol/high; model/effort attribution comes from dispatch, not separately exposed runtime metadata. Local run: review_tests_sol; candidate AL64-candidate-01. Baseline a84adbe5be46c2ef52ede0987a417562715d38e1, branch fix/AL-64-active-turn-duration. Fingerprint file SHA256: 6aab550b873d4b5de1242b889c26bb3b7c6838d4900ecae6d418b426c2a35695. Every listed application/test file matched at review start and finish. No sibling persona or conclusions were read.

The source-state versus process-liveness clarification is pending. This review assesses the candidate's documented provisional source-state semantics, without resolving that policy or authorizing release.

## Findings

### T1 — Severity 5: Invalid Codex completion silently becomes complete coverage

Classification: introduced duration-coverage defect over an existing invalid-boundary handling path; independently reproduced, with a missing consequential regression case.

Location: backend/internal/adapters/codex/rollout.go:440, especially the invalid timestamp/order branch immediately before deleting turnStarts at line 448. Relevant test: backend/internal/adapters/codex/rollout_test.go:830.

Trigger: one valid turn from second 1 through 5, followed by a turn starting at second 10 and its matching task_complete timestamped second 9; explicit cutoff second 20. The completion branch rejects the negative interval numerically, then removes the pending start without recording a limitation.

Expected: preserve the known 4 seconds and expose partial/unsupported coverage for the invalid lifecycle. Actual: 4 seconds with state complete, completed_seconds 4, no cutoff and no limitation. The invocation retains the first turn's completed_at. Downstream DurationForInvocation consequently reports no unknown coverage, and the UI/CLI can describe the incomplete accounting as completed measurements. This violates the contract's explicit invalid-lifecycle limitation requirement. It does not depend on deciding historical process liveness.

Reproduction: a disposable test appended through Go's overlay to the existing external Codex test file calls the actual ParseRolloutAt API. Command, from backend:

`go test -overlay /tmp/al64-tests-sol/invalid-overlay.json ./internal/adapters/codex -run 'TestAL64ReviewInvalidCompletionCoverage|TestRolloutActiveDurationCutoffAndTerminalBoundaries' -count=1 -v`

The existing five-case terminal/cutoff table passes; TestAL64ReviewInvalidCompletionCoverage fails, logging state complete with the known 4-second total. The overlay and probe source are outside Git. Required correction: mark invalid matched completion boundaries as limited while preserving supported known duration; add the case to a coherent existing lifecycle test. Consider malformed completion timestamps in the same branch, which also bypasses durationLimit, without inventing source duration.

### T2 — Severity 5: Both OpenCode unfinished-text paths can regress without failing their suites

Classification: required regression-evidence gap; current control implementation passes the targeted behavior. This is not a claim that the current streaming calculation is defective.

Locations: backend/internal/adapters/opencode/schema_v1_session_extraction_test.go:1717; schema_v2_session_extraction_test.go:437; backend/internal/adapters/backfill/engine_test.go:2164. The distinct production branches are schema_v1_session_extraction.go:200 (open timed text/reasoning part) and schema_v2_session_extraction.go:87 (session time_idle lifecycle state).

All newly added active family fixtures prove unfinished activity using running tools. Native evidence separately establishes supported text streaming without a running tool: v1 text time.start without time.end; v2 unfinished assistant text with session time_idle NULL. The retained deterministic fixtures do not guard those independent source-state producers.

Verification: an isolated production overlay removes only the v1 timed-text/reasoning alternative and changes v2 running initialization from activeSessions[s.id] to false, keeping running-tool recognition. The complete existing OpenCode and backfill package suites still pass:

`go test -overlay /tmp/al64-tests-sol/streaming-overlay.json ./internal/adapters/opencode ./internal/adapters/backfill -count=1`

Result: exit 0, OpenCode 1.168s, backfill 5.302s. Thus losing provisional duration for ordinary unfinished text does not fail these source/refresh suites. This is consequential because it recreates the central undercount for a natively observed activity shape.

Control and discriminating oracle: a disposable two-family test uses self-contained SQLite fixtures with the native text shapes above, explicit cutoff 10 seconds after assistant creation, and asserts 10 provisional seconds and absent completed_at. It passes both families on the frozen candidate, and fails both under the mutation with active_duration_seconds 0. Commands:

- `go test -overlay /tmp/al64-tests-sol/streaming-control-overlay.json ./internal/adapters/opencode -run TestAL64ReviewStreamingWithoutTool -v -count=1`: pass.
- `go test -overlay /tmp/al64-tests-sol/streaming-probe-mutant-overlay.json ./internal/adapters/opencode -run TestAL64ReviewStreamingWithoutTool -v -count=1`: fail in both families, as intended.

Required evidence: retain these distinct source-specific shapes in a coherent existing adapter test, using deterministic cutoff and completion/terminal expectations appropriate to the approved policy. No extra duplicate end-to-end matrix is required. The first disposable v1 probe used an incorrect positional SQL column order and failed before extraction; explicit fixture columns corrected that probe. Only the corrected passing control and discriminating failures support this finding.

## Hypotheses assessed and evidence

- **Cumulative snapshots may be added instead of replaced:** refuted for the tested family paths by inspection of the actual backfill engine matrix. It saves/reloads every stage, asserts 10 -> 20 -> 20 -> 30 -> 30, exact root identity and invocation count, completion removing provisional time/cutoff, and additive totals. Force-only rewrite is separately distinguished from RecomputeDB. These are meaningful production-boundary oracles, not copied arithmetic implementations.
- **Portable refresh may lose unavailable owners or advance their cutoffs:** the new matrix retains a same-harness unavailable provisional owner with immutable cutoff, an unrelated harness, and a completed child with parent linkage. Deep equality checks after save/reload establish their preservation. Existing unchanged recovery tests also cover missing children, malformed sources and complete owner replacement, including empty authoritative snapshots. The new fixture roots are t.TempDir paths with a synthetic FIX-DURATION ticket; no live ticket workspace or native private database is a prerequisite.
- **Rounding/component totals may drift:** domain coverage uses independent expected total 3/completed 2/provisional 1 for 1600+1600ms. ValidateDuration rejects a mutated component. The fixed-cutoff source matrix protects accumulation, while Codex's adapter table protects a future completion remaining provisional at the earlier cutoff. Invalid Codex boundaries escape these guards as T1 establishes.
- **New evidence may be stripped on persistence/migration or legacy absence may imply completion:** the real DTO serialization, runtime validator and current-schema migration decode/copy are exercised by schemas_test.go:1547. Legacy absence retains the numeric 10-second fact and yields unknown coverage; malformed sum, state, cutoff, basis and unknown member are rejected. This supports expanded readers and preservation, not old-reader forward compatibility.
- **CLI idempotence assertions may normalize away the fix:** stableDurationSnapshot copies invocation slices and normalizes only provisional evidence/total across independently clocked commands; completed rows remain exact. The active Codex result is separately checked against its stored cutoff. The engine's deterministic fixed-cutoff matrix supplies the idempotence guarantee that real-clock CLI invocations cannot establish by equality alone.
- **Persisted duration may disappear when source hydration fails:** TicketDeepDive's existing beforeEach explicitly rejects fetchWorkspaceEvents; the new component case asserts numeric value, provisional amount, cutoff and unknown coverage in the rendered KPI. The pure frontend decoder tests exercise actual decodeLedgers and malformed evidence. This establishes that component's unavailable-hydration display and decoder behavior; it does not independently establish every HTTP projection or every dashboard surface through a browser. Source trace shows the summary is explicitly carried through workflow, fleet, comparison and progression DTOs and corresponding consumers. No additional blocking display defect was substantiated in this pass.

## Verification reused and remaining limits

Reused TEST_REPORT.md's full Go race suite, frontend 834 tests, typecheck and lint on the frozen candidate; no redundant full-suite run. Independently executed the source mutation and control probes above. No product or permanent test file was modified; Go overlays and disposable fixtures remain under /tmp/al64-tests-sol. Only this assigned report was written in the checkout.

RUNTIME_EVIDENCE.md supplies independent native OpenCode source observations and actual live generate/RecomputeDB/completion/repeat numerical evidence for both pinned families. I did not repeat native provider runs and do not represent source fixtures as process-liveness evidence. Native text observations are source-shape evidence; the recorded application live numerical follow-ups use tools. T2 remains necessary deterministic evidence under the candidate's current text support.

Pending outside this scoped verdict: user source-state clarification, parent consolidation, corrected-candidate verification, required integrated make and standalone embedded-asset checks, and delivery/telemetry. The reader-first optional v6 expansion explicitly admits that old strict readers reject new fields; I do not recast that accepted deployment boundary as transparent compatibility. No release is authorized by this report.
