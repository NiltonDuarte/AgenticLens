# AL-64 candidate 03 bounded test-effectiveness recheck

- Persona: test effectiveness reviewer.
- Requested identity: gpt-6-luna / medium. Runtime identity is not independently exposed in this report channel; attribution here records the requested identity, not inferred runtime metadata.
- Baseline: `a84adbe5be46c2ef52ede0987a417562715d38e1`.
- Candidate: `fix/AL-64-active-turn-duration`; frozen corrected app/test manifest has 46 paths, all SHA256 entries verified with no missing or mismatched path. Combined manifest fingerprint: `894207b62c28077b7fdf23b72c42aec720a3b1e2b2bb66ae0724295030810cdb`.
- Scope: DUR01-DUR06 correction regression oracles and their production paths. This is a bounded recheck, not whole-change approval.

## Corrections and test-oracle dispositions

| Correction | Test effectiveness disposition | Evidence / production boundary |
|---|---|---|
| DUR01 invalid Codex completion | **Verified.** Reversed and unparsable completion cases preserve the earlier 4-second known subtotal and require partial evidence, a limitation, and unknown aggregate coverage. | `TestRolloutActiveDurationCutoffAndTerminalBoundaries` parses actual JSONL via `ParseRolloutAt`; malformed completion cannot falsely certify the known interval complete. Pre-fix overlay made both cases fail; corrected Codex race suite passed. |
| DUR02 v1 native activity start vs generic progress | **Verified.** The progress/update timestamp is advanced past the fixed cutoff while native tool `state.time.start` remains before it; expected active duration still ends at the supplied cutoff. | `TestSchemaV1ActiveDurationSnapshotAndCompletion` drives the real SQLite v1 extraction path. Assertion expects 10/20 seconds through explicit cutoff. The pre-fix overlay failed. |
| DUR03 newer message supersession boundary | **Verified.** A newer user message created after each tested cutoff is present, but must not suppress the earlier active interval at those cutoffs. | Same v1 table case `future_newer_message`; actual extraction is invoked at advancing and repeated fixed cutoffs. This distinguishes legitimate at-or-before-cutoff supersession from future records. Pre-fix overlay failed. |
| DUR04 fallback operation cutoff | **Verified.** Explicit fallback retains the unavailable exact owner and discovered owner while the active fallback invocation is exactly 3 seconds with the fixed operation cutoff, provisional state and that cutoff persisted. | `TestBackfillEngine_ExplicitFallbackResolvesWhenAllowed` reaches the `BackfillEngine.Execute` fallback path and asserts resulting ledger facts. It uses a running-tool source fixture; pre-fix overlay failed when the operation cutoff was not propagated. |
| DUR05 no-tool native streaming | **Verified.** v1 text and reasoning open part timestamps and v2 idle-null unfinished text each require advancing and repeatable provisional snapshots, no fabricated completion, completion replacement and stable completed repeat. | `TestSchemaV1ActiveDurationSnapshotAndCompletion` and `TestSchemaV2ActiveDurationSnapshotAndCompletion` invoke real family extractors. Disposable path-removal mutant made each no-tool case fail; corrected controls pass. These are deterministic source fixtures, complemented by (not substituted for) the native OpenCode v1/v2 observations in `RUNTIME_EVIDENCE.md`. |
| DUR06 invalid/missing Codex start | **Verified; discovery credit remains with architecture reviewer gpt-6.1-sol/high.** Both cases retain known completed subtotal and completion timestamp, mark partial/unknown, and avoid invented provisional cutoff. | The Codex lifecycle table exercises malformed and absent start timestamps followed by a valid earlier turn and later completion. Removing the limitation branch in a pre-fix overlay made both assertions fail; corrected full Codex race suite passed. |

## Hypotheses and side effects checked

- A v1 progress update cannot move the native activity start or make a future row end current activity: the regression cases separate source start, generic update and explicit cutoff.
- A user message supersedes only when its creation is at/before cutoff; a later message does not erase a valid earlier snapshot.
- Invalid or missing Codex boundaries preserve independently measured completed intervals and expose the omitted lifecycle uncertainty. The new cases also check no provisional cutoff is emitted and no completion boundary is fabricated for the unfinished interval.
- Exact-owner replacement and fallback retention are observable in existing fallback assertions: missing exact owner survives alongside fallback enrichment. Fixed cutoff is checked on the produced fallback invocation.
- Completion replaces provisional values rather than adding to them in both OpenCode family tests; repeated completed extraction remains stable.
- The no-tool controls reach streaming evidence directly: v1 timed text/reasoning and v2 session idle-null. Removing each corresponding path produces zero-duration assertion failures.

No new concrete defect in this bounded test-effectiveness scope was confirmed.

## Verification

Executed as independent read-only manifest verification: all 46 manifest file hashes and the combined fingerprint matched.

Reused, revision-bound evidence in `TEST_REPORT.md` and `RECHECK_TESTS_SOL_03.md`:

- DUR01-04 before/after overlay evidence: pre-fix cases red, corrected controls green.
- DUR05 no-tool path-removal mutants red, corrected controls green.
- DUR06 pre-fix start-limitation overlay red; corrected full Codex race suite green.
- Corrected affected race suites: `go test -race ./internal/adapters/codex ./internal/adapters/opencode ./internal/adapters/backfill -count=1` passed at candidate fingerprint `e5c640fe…8709306b` for DUR01-05; corrected Codex race suite passed after DUR06 at candidate fingerprint `894207b6…0810cdb`.
- Earlier full Go race, frontend 834-test, typecheck and lint results are reusable for unchanged scopes as the correction is limited to the reviewed source/test paths. Integrated `make` remains parent-owned and pending.

The overlay failures are expected regression-discrimination results, not failures of the corrected candidate. I did not rerun broad suites or alter application/test files.

## Limits and outcome

- Native OpenCode generation, active refresh, completion refresh and unchanged repeat are verified in the runtime evidence. Native Codex copied-source CLI generation/refresh remains unverified because project-ownership validation correctly rejected the disposable project; no synthetic identity bypass was used.
- Persisted open source intervals remain **provisional with process-liveness uncertainty**. The user policy decision between this source-state treatment and additional live-process proof is still pending; this recheck does not resolve it.
- Reader-first v6 compatibility limit remains as documented: old strict readers reject emitted evidence, so consumers must upgrade before writer rollout. No release is authorized by this report.
- **Scoped outcome: APPROVED** for test effectiveness of candidate 03 corrections DUR01-DUR06, conditional on the unresolved policy and parent-owned integrated verification. This is not final delivery approval.
