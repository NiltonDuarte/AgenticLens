# FIX-21 test effectiveness review (tests-opus)

- **Run:** FIX-21 discovery, focus reviewer-tests, run ID tests-opus, first pass.
- **Requested profile:** Claude Opus 5.5, medium effort. The runtime reports model `claude-opus-5-5`. Usage and cost are not available to this reviewer.
- **Reviewed revision:** base `133100bb` to head `0f73f759` (`git diff 133100bb 0f73f759`). The planning artifacts are those committed at `030bf0ec`, plus the change's current spec, design and tasks.
- **Verdict (test-effectiveness scope): CHANGES_REQUESTED.** The core oracles are strong. However, one claimed invocation point of the retention rule has no protecting test, and one CLI scenario does not establish what its name claims. A spec'd fallback is also unprotected at the capture boundary. None of these is a product defect proven by this review.

## Summary

Most new tests reach the production boundary and use exact oracles:
- The domain rule is table-tested with checks for input immutability, preserved order, unchanged facts beyond the links, and idempotence.
- Capture partial refresh is driven through `Store.Reconcile` with a real undersized budget. The test proves which owners were read, and that registry links, refreshed links, published links and unread-child snapshots are kept.
- The CLI cross-machine case runs the real binary against an empty registry and a committed ledger. It checks that other owners and other harnesses are preserved, and that ledger bytes stay identical on a repeated run.
- The budget tests use exact byte oracles derived from the fixture's known sizes.
- The archived recovery test covers recovered, missing, ambiguous, foreign-project, conflicting-parent, outside-store and non-not-found cases.

Fixtures are self-contained under `t.TempDir()`. There are no domain mocks, and no skips in the new or changed tests.

Three gaps let plausible regressions escape. Two smaller oracles could also be tighter.

## Hypotheses investigated

| # | Hypothesis | Disposition |
|---|---|---|
| H1 | Removing retention from `regroupArchiveFamily`, the second capture invocation point, is detected | **Not detected** (F1) |
| H2 | The CLI case "parent read with some children" reads some children with the parent and freshly resolves their links | **Not established** (F2). The probe shows that all 4 links are retained and none resolves freshly |
| H3 | A wrong session_meta-timestamp window for a non-v7 parent, which would prune real children, is detected at the capture boundary | **Not detected** (F3) |
| H4 | The partial refresh test reaches a real partial read | Confirmed. It asserts the parent and children 1 and 2 were read, children 3 and 4 were not, and the parent's input changed to 41 |
| H5 | The idempotence oracles are meaningful | Confirmed at domain level, at capture level (snapshots identical, parent re-read) and at CLI level (byte-identical ledger) |
| H6 | Contradiction clearing is asserted through capture | Confirmed for both the task-name result and the explicit session result. The link is cleared, the other links are kept, and a repeat does not restore it |
| H7 | The header cache is shared across consumers in one pass | Confirmed indirectly. The pass-1 exact total charges each family header and the non-date file once, so a second charge would fail it |
| H8 | The rationale for each changed oracle holds | Confirmed (see "Changed oracles") |
| H9 | Reservation, deferral, refund and deferred-first ordering are covered | Confirmed. Exact totals, both key orders, and the M5/M6 evidence in TESTS.md |
| H10 | Worker isolation is protected | Confirmed, with a timing-based oracle (see the observations) |
| H11 | `codexstoretest` is a sound, deterministic generator | Largely confirmed (see "Generator") |

## Findings

### F1. The retention call in `regroupArchiveFamily` has no protecting test

- **Location:** `backend/internal/adapters/capture/archive_projection.go:107-110`. The tests that should protect it are `reconcile_child_links_test.go` and `reconcile_archived_rollout_test.go`.
- **Classification:** introduced test gap.
- **Expectation:** the design names two capture invocation points, `finishCaptureAnnotations` and `regroupArchiveFamily`. Each needs a test that fails when its retention is removed.
- **Trigger:** an owner in a family group is superseded between annotation and the final registry read, so `reprojectArchiveFamilies` regroups the family during a partial pass.
- **Consequence if regressed:** the regrouped request-local registrations replace the refreshed parent's delegations with only the links the fresh read resolved. Accepted links would then disappear from `Sessions` and `RefreshedSessions`, which are published, in exactly the rare path the design chose to cover. The retained pairs would also no longer feed `PendingChildren`.
- **Evidence:** a disposable mutation in a scratch copy deleted the three added lines. Then `go test -count=1 -run 'Archive|Archived|ChildLink|Contradicting|Nested' ./internal/adapters/capture/` returned **ok**. TESTS.md M1 mutated only the shared domain rule, so it cannot attribute detection to this call site.
- **Severity:** 5.
- **Requested action for the tester:** add a capture case in the style of the superseded-after-annotations case. A partially read family has accepted links, and the regroup must keep them and count them.

### F2. The CLI subcase "parent read with some children" does not establish its scenario, and its oracle is loose

- **Location:** `backend/cmd/agentic-lens/capture_child_links_test.go`, subtest `parent read with some children`. The weak assertion is `latest.PendingChildren < 2`, so any figure of 2 or more passes. The test also asserts "Capture kept 4" in both subcases.
- **Classification:** introduced test weakness.
- **Expectation:** the subcase should prove that some children are read with the parent and resolve freshly, and that only the unread links are retained. The design deduplication in `AddRetainedChildLinks` would then be exercised at the CLI level.
- **Observed:** the focused run logs "unresolved child evidence 4" for both subcases. A capture-level probe used the same family, a 160 KiB budget and queued starts for children 0 and 1. In pass 0 it refreshed the parent and both children, but the parent's fresh links were all empty. Only pass 1 resolved two links. In the CLI's single pass, therefore, all 4 links are retained: two are confirmed and two are unobserved. No link is freshly resolved.
- **Consequence:** two things go untested at the CLI level.
  - A pass in which the fresh read resolves some links, so retention runs alongside fresh links.
  - The CLI deduplication of retained pairs against registry pending work.

  A regression that over- or under-counts in that mix would pass. The subtest name also misdescribes what is proven.
- **Side observation, outside test focus, for the parent and behavior reviewers:** children discovered from queued `SubagentStart` hints are extracted in the same pass as the parent, but are not combined with it, so their links resolve only on the next pass. This may be pre-existing. It explains why a single `--capture-auto` pass reports confirmed children as unresolved evidence.
- **Evidence:** `go test -count=1 -run 'TestCaptureAutoKeepsCommittedChildLinksWithoutCaptureState' -v ./cmd/agentic-lens/` passed, and both subcases logged 4. The disposable capture probe `TestZZProbeQueuedChildrenSamePass` ran in a scratch copy and logged pass 0 `links=["" "" "" ""]` and pass 1 `links=[C1 C2 "" ""]`.
- **Severity:** 4.
- **Requested action:** assert the exact figure. Then either make the subcase reach a fresh partial resolution (for example by settling the children in one prior pass on machine B), or rename it to what it proves.

### F3. The parent session_meta-timestamp window fallback is unprotected at the capture boundary

- **Location:** `backend/internal/adapters/capture/children.go:221-233` (`childCreationWindow`). Only the unit test `TestCreationWindowParentInstant` covers the rule.
- **Classification:** introduced test gap.
- **Expectation:** the spec requirement "The parent's instant SHALL come from its own UUIDv7, or else from its session_meta timestamp" needs one capture-level case: a non-v7 parent with conventionally named v7 children.
- **Trigger:** a regression in how capture reads or applies the parent header, such as the wrong field, a timezone error or an off-by-a-day shift.
- **Consequence:** real children of a non-v7 parent would be pruned silently. This is the unsafe direction of the pre-filter. Practical reach is lower because Codex parents normally carry UUIDv7 identities.
- **Evidence:** a disposable mutation in a scratch copy shifted `header.CreatedAt` by +48 h before `NewCreationWindow`. The whole `./internal/adapters/capture/` package then failed only `TestCompatibleGlobalWrapperUpdatePreservesOtherPinnedProject`. That failure is an artifact of the copy: `scripts/tests/fixtures` is missing outside the repository. Every FIX-21 and existing capture test passed.
- **Severity:** 3.

### Suggestions (severity 1 to 2)

- **S1.** `TestRetainAcceptedChildLinksIsIdempotentAfterSiblingRelease` asserts only that the result is stable, not its value. I traced it: the expected result is `[C1 C2]`, with outcomes retained and kept_pending_confirmation. A regression that is stable but wrong, such as `["" C2]`, would pass. Assert the links. Severity 2.
- **S2.** "Manual generate does not carry links forward" is protected only because the call sits inside `if *autoFlag` in `main.go`. No test runs a manual generate over a ledger that has a link whose child is absent, to show that the link is cleared rather than kept. The spec's only listed scenario (full manual linking) is covered. Severity 2.
- **S3.** No automated guard stops a production package from importing `codexstoretest`. TESTS.md records a one-off `go list -deps` check. The package imports `testing` and `time/tzdata`, so an accidental production import would embed both. Consider a `go list`-based test or lint rule. Severity 1.
- **S4.** In the large-store fixture, the non-v7, mismatch and gzip irregular names sit in pruned old day directories. Only unit tests cover "a non-v7 or malformed name outside a pruned directory is read in a child sweep" (`TestCreationWindowFileRule`); no capture sweep does. The capture sweep oracle covers only the non-date directory. Severity 2.

## Changed oracles in existing tests

- **`TestReconcileCodexRecoversNestedSpawnWithoutChildCallbacks`: the change is justified.** The UUIDv7 prefix `019c0000-0000-7…` encodes 2026-01-27T15:00:18.048Z, so the old `sessions/2026/01/01` directory contradicted the IDs. A real store never produces that, and the spec permits pruning it. The intent and assertions are unchanged, and nesting and contradiction coverage are kept. No weakening.
- **`TestReconcileUnassignedRootsDoNotConsumeTicketWorkOrPublishIdentities`: the change is justified, and stronger.** The old oracle `SourceBytes == allowance` used `allowance = discovery + extraction`. The new tight case computes the same sum independently and asserts `SourceBytes == work` and attempts == 1. The roomy case (twice the allowance) proves the figure is the work done, not the cap. The fixture asserts that each file is below 4 KiB, which keeps the "whole file per header" premise honest. The attempts, starvation, disclosure and registry-count assertions are all kept.

## Generator (`codexstoretest`)

- **Determinism.** It is seeded and deterministic. UUIDv7 creation is encoded correctly. `TaskFamilyUUID` puts milliseconds in the first 48 bits; I checked its format arithmetic.
- **Real JSONL where it matters.** Family rollouts are dense, real JSONL. Sparse holes are used only for unrelated files that capture never reads past the header.
- **Self-test.** It checks recorded sizes against stat, the exact header size, that each name carries its own UUID, the instant ordering relative to the root, the machine-local day directory, each irregular note exactly once, and the family count. A generator regression therefore fails at the source.
- **Time zones.** `time/tzdata` makes CET and CEST portable across hosts.
- **Minor.** `IncidentScaleSpec` panics rather than failing a test when the zone cannot load, which is unreachable with tzdata embedded.
- **Verdict.** It is a sound generator for these oracles.

## Runtime

- **New tests are fast.** The focused CLI run here took 9.3 s for the two new CLI tests and one other.
- **Package growth reported in TESTS.md:** capture from 83 to 117 s, cmd from 316 to 387 s, and capture race from 379 to 410 s.
- **Large-store cost.** The large-store default (divisor 4) is the main cost. It is bounded by the flag, and divisor 1 is opt-in.
- **The cmd increase looks larger than the new CLI tests explain.** They took about 5 s in my run, so the reported +71 s is likely host-load variance. Re-measure before treating it as cost.
- **Conclusion.** Proportionate.

## Observations (no finding)

- **The worker isolation test depends on timing.** It requires `elapsed <= 8 s` against a 10 s lifetime. It fails safe, since a regression runs until the lifetime ends. On a heavily loaded host, a legitimate idle exit could approach the margin. The `lifetime exhausted` and error checks are the primary oracle.
- **The large-store tests use package-internal hooks.** These are `workBudget` fields, `refreshTicketOwners` and `recoverChildren`. This couples them to the implementation, but it is the only way to assert a specific phase order. The end-to-end `ReconcileAt` totals complement them.

## Verification performed

All runs were from `backend/`. Product code was not modified. The mutations and probes ran in a disposable scratch copy, which has since been deleted. `git status` shows only the preexisting `tasks.md` modification and the review directory.

- `go test -count=1 -run 'TestCaptureAutoKeepsCommittedChildLinksWithoutCaptureState|Isolat' -v ./cmd/agentic-lens/`: PASS in 9.3 s. Both subcases logged unresolved child evidence 4.
- Probe `TestZZProbeQueuedChildrenSamePass` (scratch only): result recorded under F2.
- Mutation: `regroupArchiveFamily` retention removed, then the focused archive and child-link run. **Survived** (F1).
- Mutation: the parent-header window shifted by +48 h, then the whole capture package. **Survived** (F3). The single failure was a scratch-copy environment failure that is unrelated to the mutation.
- Reused from TESTS.md, without rerunning the package suites: M1 to M6, the full package results, the race results and the large-store divisor 1 result.

## Gaps in this review

- I did not rerun the full package or race suites, per the brief.
- I did not independently re-derive every byte total in the large-store test. I checked the oracle formulas against the cost rule, and the M3 and M5 evidence.
- I did not probe the behavior of the CLI `--capture-auto` path when RestoredSessions are present, which is the second `RetainAcceptedChildLinks` call in `retainCaptureChildLinks`. No test I found exercises a restore together with retention.
