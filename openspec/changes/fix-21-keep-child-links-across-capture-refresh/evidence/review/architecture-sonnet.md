# FIX-21 architecture review (arch-sonnet)

Run: FIX-21 discovery run arch-sonnet, phase first pass, focus reviewer-architecture.
Requested identity: Claude Sonnet 5.5 / high (runtime reports claude-sonnet-5-5; effort not independently observable).
Reviewed revision: base 133100bb .. frozen candidate 0f73f759 (branch FIX-21/fix-capture-refresh-child-links). Product code read-only.

## Verdict: APPROVED (scoped to architecture and maintainability)

No blocker. Five optional cleanups (A1-A5, severity 2-4), none a correctness defect.

## Hypotheses investigated

| # | Hypothesis | Disposition |
|---|---|---|
| H1 | Link retention is not pure/immutable or uses any/interface{} | Refuted. `domain/child_links.go` has no imports, no I/O, deterministic (slice-ordered fixpoint, no map iteration affects output). Inputs are copied (`append([]SessionFact(nil), ...)`, per-owner delegation copy). Typed string enums for evidence and outcome. Diff grep for `any`/`interface{}` in backend adds: none (only the word "any" in comments; `panic` only in test-support). |
| H2 | I/O or policy sits in the wrong layer | Refuted with one suggestion (A4). Domain holds the rule; adapter applies it at the two annotation points; the CLI helper owns the generation-only application, stderr message and diagnostic re-record. main.go diff is 4 lines at one site. |
| H3 | The capture-only discovery API leaks into manual/backfill/live | Refuted. Remaining production callers of the unchanged lookups: `live/service.go:1060` (`DiscoverExactRolloutPathContext`), `main.go:683` and `backfill/engine.go:40` (`DiscoverRolloutPaths`). `DiscoverExactCaptureRolloutPath`, `HeaderCache`, `CreationWindow` are used only from `adapters/capture`. `CreationWindow{}` and a nil `*HeaderCache` are zero-effect, so opt-in is by construction. |
| H4 | Budget split/deferral model is incoherent | Refuted; coherent but dense. Invariants hold: identification clamps to `min(bytes, split.identification)`, extraction to `min(bytes, split.extraction)` until `releaseSplit`, all charges funnel through `charge`, so totals cannot exceed `max_source_bytes`; `SourceAttempts` excludes header reads (deferral refunds `sessions++`, `attempts--`). Deferred retry is bounded: `deferred` can only be set while the split is unreleased, i.e. during rotating rounds, and the final refresh runs with the split released. See A5 for naming/comment precision. |
| H5 | `codexstoretest` is not test-only | Refuted. `go list -f '{{.Imports}}' ./...` shows no non-test importer; only `_test.go` files import it (capture, codex, cmd). It imports stdlib only. Caveat A3: nothing mechanical enforces it. |
| H6 | main.go footprint collides with PR #162 (AL-94) | Refuted. `gh pr diff 162` touches main.go only in `runBackfill` (hunks at ~1315, 1342, 1377). FIX-21's hunk is at ~598 in `runGeneration`. PR #162 does not touch `capture.go` or `capture_worker.go`. |
| H7 | Dispatch does not match comments | Mostly refuted; see A5 (non-causal `deferred` comment), A1 (error wrapping divergence not stated). |
| H8 | Behavior drift between old and new header parsing | Refuted for identity/parent: `decodeSessionMetadata` already folds `source.subagent.thread_spawn.parent_thread_id` into `ParentID`, so `RolloutHeader.ParentID` matches the old `Transcript.ParentThread`. |

## Findings (all introduced by this change, all optional)

### A1. Exact lookup and rollout listing duplicated rather than parameterized (sev 4, suggestion)
- Location: `codex/discovery.go` `DiscoverExactCaptureRolloutPath` and `listFilteredRolloutPaths` versus `discoverExactRolloutPath` and `listRolloutPaths` (about 35 and 55 duplicated lines).
- Expectation: one selection core (ambiguity rule, incomplete-discovery rule, error mapping) with a capture option for the name filter and header cache, since the design requires the capture path to differ only by filtering.
- Divergences already present: (a) an empty store is wrapped as `ErrOwnerUnavailable` only on the capture path, the manual path returns an untyped error; capture-side isolation (`unassignedRelocationIsolation`) depends on that typed wrapping; (b) the manual walker claims each path while walking, the capture walker claims after filtering (intended, but undocumented at the call site); (c) the oversized-first-record fallback exists only in the capture copy via `oversizedThreadID`.
- Trigger: any later change to ambiguity or error semantics in one copy.
- Consequence: silent drift between manual and capture exact-owner resolution; the two paths can disagree on the same store.
- Evidence: side-by-side read of discovery.go lines ~500-550 and ~600-670. Not exercised as a defect.
- Cheap mitigation if not refactored: a shared test pinning that both paths return the same ambiguity and unavailable results for one store.

### A2. Header/ancestry decoding duplicated beside `readSessionContext` (sev 3, suggestion)
- Location: `codex/discovery_budget.go` `headerFromMetadata` / `RolloutHeader.SessionContext` versus `codex/context.go` `readSessionContext` (same nested `thread_spawn.parent_thread_id` struct and the same "both spellings must agree" condition, same error texts).
- Consequence: two copies of the ancestry-conflict rule and two session_meta readers (`ReadSessionContextWithinBudget` still serves `capture/project.go:317` and `codex_context.go:36`). A new ancestry spelling must be added twice.
- Suggestion: have `readSessionContext` decode through `RolloutHeader`/`SessionContext()`, or extract the conflict check to one function.

### A3. Test-support package placement, and production code now used only by tests (sev 2, suggestion)
- `codexstoretest` is a non-`_test` package importing `testing` and `time/tzdata`; test-only status rests on a doc comment and a one-off `go list` (TESTS.md). A future production import would compile and embed tzdata. Suggest a cheap guard (a test that fails if `go list -deps` of the production main packages includes it) or accept as a documented convention.
- `codex.DiscoverExactRolloutPathWithBudget` has no production caller after this change (only `discovery_budget_test.go`, `discovery_window_test.go`, `reconcile_large_store_test.go`). Decide: delete with its tests, or keep deliberately as the manual reference used by A1's equivalence test. `ChildCreationTolerance` and `MaxLocalDayOffset` are exported without any consumer outside `discovery.go`.

### A4. Retention coverage straddles adapter and CLI via mutable result state (sev 3, suggestion)
- Location: `capture/reconcile.go` (`countedChildLinks`, `reconciliation`, `AddRetainedChildLinks` pointer receiver, `ReconciliationDiagnostic` value receiver), `cmd/agentic-lens/capture.go` `retainCaptureChildLinks`.
- Observations: the reconciliation diagnostic is persisted twice for one pass (ReconcileAt's defer writes it "completed", then the CLI helper re-records it "partial" when it adds links), leaving a window where readers see the pre-retention state; the CLI helper must know adapter persistence details (it reopens the store with its own 2 s timeout). The harness constant `SourceCodex` and the baseline-as-accepted policy are repeated at three call sites (`finishCaptureAnnotations`, `regroupArchiveFamily`, CLI), as intended by D3 but with no shared constructor.
- Suggestion: move diagnostic re-recording behind one `capture` entry point (for example a `Store` method taking the result) so the CLI helper stays pure orchestration plus the stderr line. main.go would not change. Acceptable as is: the window is narrow and the design accepted the two-site model.
- Minor: in `regroupArchiveFamily` the function has a local `previous` (with synthetic ParentID-only entries) and passes `group.previous` to retention; the choice is defensible (unobserved evidence still allows retention) but easy to misread. A one-line comment would help.

### A5. Naming and comment precision (sev 2, suggestion)
- `reconcile.go` refreshTicketOwners: comment says an owner "refused only because extraction is holding identification's reserved share" is deferred, but the test is `budget error && extractionReserved()`, which is true whenever the reservation is currently active, regardless of whether the reservation caused this refusal (also true for a discovery session-limit refusal). Harmless (bounded single retry in the released final refresh) but the comment overstates the condition.
- `readMetadata` returns a `codex.RolloutHeader`, not metadata/transcript; `readHeader`/`readMetadata` are two near-synonyms on `workBudget`/package scope with different charge targets. `IsolationReason`/`IsolatedSource` live in `source_database.go` while `ReconcileResult` that carries them is in `reconcile.go`.
- `ChildLinkDecision.Evidence` is the empty string (not a `ChildParentEvidence` value) for `refused_ambiguous`; the typed enum has no "not evaluated" member. Domain-only; no consumer depends on it.
- Isolation is encoded persistently as `CapturePending` plus prose in `Detail`, and distinguished only by the request-local `IsolatedSources` list. This matches the design (request-local by decision) but means any other consumer of `Registrations` still sees these rows as ordinary pending.

## Cross-focus observation (not verified, for data/behavior reviewers)
A retained link whose child rollout is permanently unobservable (deleted) is re-retained on every pass and counted in `PendingChildren`, so that ticket's reconciliation diagnostic stays "partial" and the CLI message repeats. `PendingChildren` is consumed only for diagnostics and the summary line (`reconcile.go:102,383,639`, `capture.go:816`), not for worker retry or success, so I found no loop. Accepted facts are preserved (matches AGENTS.md rule 6). I did not test convergence on such a store.

## Verification performed
- Read design.md, full diff of production files, `domain/child_links.go` in full, unchanged consumers (`discoverExactRolloutPath`, `listRolloutPaths`, `readSessionContext`, `decodeSessionMetadata`, worker unassigned pass, main.go call site).
- `go vet ./internal/domain/ ./internal/adapters/codex/...` clean.
- `go list -f '{{.ImportPath}} {{.Imports}}' ./...`: `codexstoretest` has no non-test importer.
- `gh pr diff 162`: no main.go overlap with the FIX-21 hunk; no capture.go/capture_worker.go overlap.
- Grepped production callers of the manual, live and backfill lookups and the new capture-only API (list in H3).
- Reused (not rerun) TESTS.md results for domain/codex/capture/cmd suites, race and mutations M1-M6. No whole-package suites run.

## Gaps
- Did not execute mutation or large-store probes myself; relied on TESTS.md.
- Did not assess byte-accounting correctness of `readRolloutHeader`'s `counterReader` against gzip rollouts (data/behavior focus).
- A1 divergences are read from source, not demonstrated with a differing-output probe.
