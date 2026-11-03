# AL-62: Codex skill and delegation facts

## Execution baseline

- User approved the complete AL-62 specification before implementation: "its pre-approved. full send it".
- Initial Git status: clean. Baseline: `122e983f8c365a2aa248b1b6d492252555e922ae`.
- Existing isolated branch: `AL-62/fix-codex-skill-delegation-facts-granular-ownership`. Workspace: `alspaces/AL-62/`.
- Planning, implementation, testing and required focused review pending. Telemetry deferred until final delivery.

## Source diagnosis and accepted design

The planner inspected the local AL-55 root and four child rollouts, whose metadata reports Codex 0.166.0. The source records an explicit skill-file read and native spawn calls, but the existing reader only emits generic tool events. Child histories replay parent calls; extraction must respect the existing owning-session guard. Privacy-safe identifiers and limitations are recorded in [CONTRACT.md](CONTRACT.md). Live evidence remains unchanged and is not a test fixture.

Schema 6 expands schema 5 with canonical session-owned collections and source occurrence IDs. Metric invocation rows do not establish raw-call ownership. Unresolved legacy facts remain preserved, and unknown skill execution outcomes are excluded from success rates. Fetched origin/main matches the baseline; no concurrent schema collision was observed at planning time.

## Active review and corrections

Candidate01 fingerprint: `b1e3db5e43d7e5c2b5865ba7472f3d599dd0a2984ac13c5467a6e598482376af`. All eight focus/profile discovery cells completed. Initial consolidated outcome: CHANGES_REQUESTED. Sol reviews use inherited gpt-6.1-sol/high (serving identity not independently exposed); Luna T3 dispatch acknowledges gpt-6-luna/medium (source-session identity not independently verified). Transient native capacity failures were retried without substitutions.

- Fabricated skill loading from ambiguous wrappers was independently reported by behavior, data, architecture and test reviewers (each gpt-6.1-sol/high), plus architecture (gpt-6-luna/medium). Backend coder gpt-6.1-sol/high is correcting the evidence boundary.
- Unchanged pre-upgrade capture snapshots missing newly supported facts: data reviewer gpt-6.1-sol/high; backend coder gpt-6.1-sol/high adding source-backed additive extraction upgrades while preserving freshness and writer ownership.
- Multiple-skill badge loss and offline timeline evidence gap: behavior reviewer gpt-6.1-sol/high; frontend coder gpt-6.1-sol/high correcting read-only projections.
- Missing populated auxiliary persistence assertions: test reviewer gpt-6.1-sol/high; tester gpt-6.1-sol/high extending coherent existing tests.
- Source-less explicit session provenance claim from data reviewer gpt-6-luna/medium refuted by parent: accepted ledgers enforce globally unique session IDs in migration and DTO validation. Explicit ID remains reliable evidence.
- Earlier tester gpt-6.1-sol/high growth-loss claim was refuted after independent source-checkpoint trace; fixture correction preserved AL-76 guards.

Behavior and test reviewers gpt-6-luna/medium reported no findings in their initial scopes; these do not overrule reproduced defects. Active evidence remains in iteration REVIEW.md and private temporary reports until corrected consolidated approval. Full make deferred by user until actual predecessor integration.

## Reviewed implementation checkpoint

Candidate03 fingerprint: `f300879df36eee4cbdc69b6de66999dbf469ddcd3eb6f65f8a65b71279332c8a`. All four focuses under gpt-6.1-sol/high and gpt-6-luna/medium approve the corrected scope; no open application findings. Parser, schema6, extraction-version upgrades, canonical ownership, preservation and offline timeline corrections are implemented. Candidate02 shell-comment residual was corrected with failing-before/passing-after native and wrapper cases. Verification is recorded in BACKEND_NOTES.md, FRONTEND_NOTES.md and TEST_REPORT.md in iteration01.

External gate: actual AL-60/AL-61/AL-76 predecessor merges remain pending. Fetched origin/main is still baseline at checkpoint; no upstream implementation was integrated. Per user instruction, expensive full make and final standalone verification are deferred until those merges. Preserve AL-61 timestamps, AL-76 lossless serialization/publication locks and AL-60 shared capture repairs when integrating. Manager baseline capture native/capability/worker failures remain explicit evidence. Consolidated delivery is INCOMPLETE; ticket specification/index completion, final telemetry and PR publication remain pending. Implementation will be committed as a clean checkpoint; final telemetry remains deliberately deferred because iteration01 continues.

## Authorized draft publication

The user explicitly superseded the PR publication hold and authorized a draft of the reviewed checkpoint. The implementation reviewed at `29ee4a802000c1d3302c515e001f7692d3d7fe8f` is unchanged by publication documentation. All existing review/test evidence remains preserved. Integration stays on hold until MGT-2's actual merge notification; draft publication does not establish merge readiness. Pending gates remain predecessor reconciliation, required integrated make/race/build and standalone asset verification, affected re-review, final telemetry/review retention and ticket lifecycle completion. MGT-2 owns merging.

## Draft Backend CI correction

Reproduced the four failures from [Backend job](https://github.com/NiltonDuarte/AgenticLens/actions/runs/41705995778/job/126161340901) on 22f571ae649d5dc53071fc6410f187816905aa9e. Stale root-collection and schema5 output assertions were corrected to accepted canonical ownership/current schema; explicit v4 fixtures and all five cost/evidence variants remain unchanged. Usage-only independent roots legitimately record no auxiliary activity. Native parents and source-backed positive facts remain required.

The corrected manual tamper oracle exposed a real authorization bypass, discovered/fixed by coder gpt-6.1-sol/high: canonical nested facts escaped legacy-root-only receipt checks. Candidate, committed, registered snapshot and published receipt evidence now use owner-derived projections. Strict row/provenance/multiplicity validation is unchanged. Independent tester gpt-6.1-sol/high verified a failing-before old-product overlay with current tests, then passing current CLI save/reload/rejection checks.

Focused verification: affected four tests PASS; broader CLI capture/history subset PASS; affected adapter subset PASS. Independent go test -race affected/related CLI cases PASS (18.810s), SQLite evidence ownership/multiplicity test PASS (1.167s), no skips; diff check PASS. Existing TestMain builds the CLI subprocess without race instrumentation, so only test processes are race instrumented. All eight focus/profile correction reviews APPROVED. See [correction diagnosis](iterations/01/CI_CORRECTION.md), [independent verification](iterations/01/CI_VERIFICATION.md) and active review.

Existing limit from data reviewer gpt-6.1-sol/high: interrupted publication can restore readable prior ledger bytes after its sole receipt was replaced, then generation fails closed without source/Git proof. No facts were erased in the probe, and authorization was not weakened. MGT-2 receives this shared-publication follow-up for triage. The source/schema correction identifies no contract ambiguity. Predecessor integration/full make remain held; PR103 stays draft and remote CI results are pending the corrected push. Final telemetry/review retention/lifecycle completion remain deferred with the open iteration.

## CI-PUB-01 repaired: independent durability checkpoint

The severity5 finding originally discovered by data reviewer gpt-6.1-sol/high is fixed; it remains AL-62-owned. Coder gpt-6.1-sol/high added bounded active/pending trusted receipt generations scoped to worktree, ticket, physical destination and publication generation. R0 survives rollback; R1 survives file commit before receipt finalization. Strict source/provenance/multiplicity checks remain; explicit legacy migration and independent-proof rebind cannot invent lost proof or use raw ledger/journal rows as authority. No ledger schema or generic journal change was required.

Independent tester gpt-6.1-sol/high verified both windows with actual CLI/source-backed fixtures, repeated unavailable-source generation, canonical/legacy/empty/absent states, metadata migration and lost-proof rejection, scopes, malformed metadata, relocation fences and existing source/history/checkpoint/file-journal controls. Five focused race suites pass (8.246s,3.164s,11.271s,1.646s,5.262s); changed-package go vet and diff check pass. All eight focus/model correction reviews approve with no new findings. Reviewed fingerprint: cec8ce3ace37d7fcc300ea1089663ef21153502037a8548ad1da946b81651b8f.

[Protocol](iterations/01/RECEIPT_IMPLEMENTATION.md), [focused verification](iterations/01/RECEIPT_VERIFICATION.md), [plan](iterations/01/RECEIPT_PLAN.md). Limits: reconstructed crash states, no power-loss/mid-SQLite-finalization fault injection, and CLI subprocesses built without race instrumentation. This approves only the independent repair. PR103 remains draft; predecessor integration, full make, final standalone verification/telemetry/review retention and lifecycle completion remain held or pending. Already lost R0 remains unavailable without independent trusted evidence.
