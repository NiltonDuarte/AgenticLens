# AL-79 active consolidated review

Verdict: CHANGES_REQUESTED. All twelve C2 focus/model cells completed; confirmed blockers and a material transaction-evidence gap remain. No consolidated approval or publication authorization is issued.

## Frozen revisions

C1 source SHA2565489a593e9b782aafa6c257efc7c2ae98fcdf2d1b7793759e6a08a8fc34d9a78. C2 source SHA25684aa66679b8b2b643ccdfd3c17148b77a369ea383a546eee372397778218a7de, base4df45967c801824500ac2ea39a8f52fa9cbdcf66. C1 had two disclosed test-only fixture drifts carried into C2; C2 source stayed frozen throughout discovery. Immutable raw diffs and detailed run/finding records remain outside Git under /tmp/al79-review-coordinator-b67f2d.

## Review roster

| Focus | Model/effort | Phase/candidate | Outcome | Report |
| --- | --- | --- | --- | --- |
| behavior | gpt-6.1-sol/high | discovery/AL79-C2 | CHANGES_REQUESTED | [REVIEW_BEHAVIOR_SOL.md](REVIEW_BEHAVIOR_SOL.md) |
| data | gpt-6.1-sol/high | discovery/AL79-C2 | CHANGES_REQUESTED | [REVIEW_DATA_SOL.md](REVIEW_DATA_SOL.md) |
| architecture | gpt-6.1-sol/high | discovery/AL79-C1 | CHANGES_REQUESTED | [REVIEW_ARCHITECTURE_SOL.md](REVIEW_ARCHITECTURE_SOL.md) |
| tests | gpt-6.1-sol/high | discovery/AL79-C2 | INCOMPLETE | [REVIEW_TESTS_SOL.md](REVIEW_TESTS_SOL.md) |
| security | gpt-6.1-sol/high | discovery/AL79-C2 | CHANGES_REQUESTED | [REVIEW_SECURITY_SOL.md](REVIEW_SECURITY_SOL.md) |
| build | gpt-6.1-sol/high | discovery/AL79-C2 | CHANGES_REQUESTED | [REVIEW_BUILD_SOL.md](REVIEW_BUILD_SOL.md) |
| behavior | gpt-6-luna/medium | discovery/AL79-C2 | INCOMPLETE | [REVIEW_BEHAVIOR_LUNA.md](REVIEW_BEHAVIOR_LUNA.md) |
| data | gpt-6-luna/medium | discovery/AL79-C2 | CHANGES_REQUESTED | [REVIEW_DATA_LUNA.md](REVIEW_DATA_LUNA.md) |
| architecture | gpt-6-luna/medium | discovery/AL79-C2 | CHANGES_REQUESTED | [REVIEW_ARCHITECTURE_LUNA.md](REVIEW_ARCHITECTURE_LUNA.md) |
| tests | gpt-6-luna/medium | discovery/AL79-C2 | APPROVED | [REVIEW_TESTS_LUNA.md](REVIEW_TESTS_LUNA.md) |
| security | gpt-6-luna/medium | discovery/AL79-C2 | APPROVED | [REVIEW_SECURITY_LUNA.md](REVIEW_SECURITY_LUNA.md) |
| build | gpt-6-luna/medium | discovery/AL79-C2 | INCOMPLETE | [REVIEW_BUILD_LUNA.md](REVIEW_BUILD_LUNA.md) |
| architecture | gpt-6.1-sol/high | re_review/AL79-C2 | CHANGES_REQUESTED | [REVIEW_ARCHITECTURE_SOL_C2.md](REVIEW_ARCHITECTURE_SOL_C2.md) |

All requested model/effort settings were dispatch acknowledged. T3 completed-run metadata exposes Luna model identity but not source effort/usage/cost; no unsupported source-observed tuple is claimed.

## Canonical findings

| Finding | Severity | Issue | Disposition |
| --- | --- | --- | --- |
| ARCH-01 | 7 | Competing dispatcher ownership breaks A-to-B installation and later uninstall. | fixed_pending_final_review; confirmed |
| ARCH-02 | 6 | Machine native dispatcher rejects supported legacy manual pins. | fixed_pending_final_review; confirmed |
| ARCH-03 | 6 | Native dispatch changes nested manual caller working directory. | fixed_pending_final_review; confirmed |
| BEH-01 | 6 | User installation silently permits duplicate active legacy project callbacks. | open; confirmed |
| BEH-02 | 7 | Legacy disable permits a live worker to keep publishing fresh source changes. | open; confirmed |
| DATA-01 | 6 | Automatic generation inherits shared capture limits instead of selected-local defaults. | open; confirmed |
| DATA-02 | 6 | Root-journal recovery rewrites home hooks before acquiring the home lock. | open; confirmed |
| DATA-03 | 6 | An explicitly bound child prevents parent reconciliation and publication. | open; confirmed |
| ARCH-04 | 6 | Terminating native dispatch leaves the selected manual CLI process alive. | open; confirmed |
| ARCH-05 | 6 | PowerShell native routing bypasses existing lossless transport and timeout. | open; confirmed |
| ARCH-06 | 5 | Windows missing-launcher repair stores inconsistent executable ownership hashes. | open; confirmed |
| TEST-01 | 4 | Native user installation and migration interruption/recovery lack a direct acceptance oracle. | open; confirmed |
| SEC-01 | 5 | Native Windows binding feedback uses a non-executable POSIX-quoted command. | open; confirmed |
| BUILD-01 | 7 | User installation accepts an incompatible native executable and breaks the shared dispatcher. | open; confirmed |
| ARCH-07 | 0 | Alleged callback process-cwd routing loss conflates process cwd with declared native session cwd. | resolved; refuted |
| DATA-04 | 4 | Retaining mutable development output can accept incomplete bytes changed after selection. | open; confirmed |

Original C1 ARCH-01/02/03 corrections pass actual installer/native counterexamples in coder, parent and architecture re-review. New C2 findings remain open. Parent independently reproduced the behavior/data overlays, incompatible-binary installation and retained-copy mutation. Signal/process-ownership discovery is shared by architecture reviewer and build/distribution reviewer plus the independent parent, all gpt-6.1-sol/high; supplied issue verification by parent is not discovery. DATA-04 was reported by data reviewer gpt-6-luna/medium and independently confirmed by parent. Other Sol findings retain their actual focus/model/effort in the named reports and detailed observations.

ARCH-07 is refuted: the callback JSON cwd is harness-declared session context, distinct from process cwd. PO20 section3 explicitly requires both process-cwd independence and rejection of contradictory authoritative context. The existing unrelated-process-cwd regression passes; forcing source rollout cwd over a conflicting declared native cwd is not the approved requirement. Reported severity8 is retained in working observations; reconciled severity0 reflects no demonstrated defect. No source change is requested for this claim.

TEST-01 remains an evidence gap despite the other test reviewer scoped approval: legacy journal tests and normal-path user install tests do not establish interrupted new-root-journal migration recovery, and a dropped-migration overlay survived those selected tests. Approval votes do not refute a demonstrated gap. Likewise no-findings specialist passes do not refute specific static Windows command/transport counterexamples.

## Verification and limits

Genuine normal-trust Codex0.167 Linux native root/child callback/binding/source/automatic publication and separate worktree preservation passed on its stated revision. Source-matched facts and exact build/source identifiers are in NATIVE_EVIDENCE.md. Native Windows remains unexecuted; cross-compilation/static evidence does not establish runtime behavior. Existing CI has a real Windows gate on pull requests. Required main integration/make/race and final revision-aware consolidation remain pending.

## Correction release

The parent explicitly releases C2 source/test freeze now that every assigned C2 first pass has completed. C3 is a correction tranche in iteration01, preserving the approved contract, not new product scope. Implementation engineer owns production/runtime/setup; Go tester owns Go tests; distribution tester owns scripts/tests. Each preserves others' changes. The prepared PLAN.md tranche and this canonical list govern corrections; no orchestrator application edits or new user approval are required for these review fixes.


## C3 first-wave result and C4 correction release

Frozen C3 source hashbaf424ef84db886d4d8d1de9f5c271c3d4955c554714d400df4397a5d5f42e0c passed targeted package and POSIX checks. Architecture Sol6.1/high approved its scoped re-review. Behavior Sol6.1/high confirmed residual BEH-01 (explicit migration leaving an unknown executable legacy callback); data Sol6.1/high confirmed residual DATA-02 (resource-set change before leases) and new DATA-05 (external-parent fresh publication falsely complete/clean). Sources remained frozen. Reports REVIEW_BEHAVIOR_SOL_C3.md, REVIEW_DATA_SOL_C3.md, REVIEW_ARCHITECTURE_SOL_C3.md retain precise evidence and controls.

C3 is not a deliverable candidate. Nine queued C3 cells were never dispatched because the first wave found blockers; their obligations move to corrected C4 with both profiles and every selected focus preserved. The parent explicitly releases source/test freeze solely for these three narrow corrections, then will freeze a new candidate for complete coverage. No acceptance change or user gate is introduced. DATA-05 must preserve correct child facts, per-owner analysis and known cached summaries while avoiding fabricated fresh root/group completeness and unqualified clean wording. Integration/make/real Windows stage remain pending; no publication approval is claimed.


## C4 frozen correction verification

Candidate AL79-C4 source SHA256 9d1ca49a319643feeb13774a28942037cfff9d175884d53a8e8495660a90e265 against baseline4df45967c801824500ac2ea39a8f52fa9cbdcf66 is frozen. C4 corrects residual BEH-01, DATA-02 and DATA-05 without changing approved boundaries. All confirmed findings, including tester-discovered INST-01, are fixed pending current re-review; ARCH-07 remains refuted. Both configured profiles and six focuses remain required. C4 targeted Go packages and the affected POSIX migration oracle pass; full integration/make and native Windows hosting stage remain pending. No consolidated approval is issued.

C4 Sol behavior, data/failure and architecture correction reviews are scoped APPROVED with unchanged source fingerprint. Independent probes passed the residual migration, live disable, separate-child publication and lease-race cases; disposable mutations detected the repaired recovery and partial-analysis failures. Tests/security/build correction reviews continue. These approvals do not satisfy integration or the remaining profile coverage.

All six C4 Sol6.1/high focuses are scoped APPROVED. Security verified corrected platform binding/transport source and isolated admission/context cases; build reviewer independently passed actual POSIX signal-ownership and incompatible-binary refusal. Test reviewer verified actual transaction and summary oracles, including consequential mutation detection. Both-profile completion, integratedmake and native Windows hosting stage remain pending.

C4 Luna behavior/data/architecture/test-effectiveness are scoped APPROVED. Architecture explicitly withdrew the earlier A1 routing objection based on the declared-context/process-cwd distinction. Current source remains frozen. Final security/build Luna cells are running; fullmake/integration and nativeWindows hosting-stage limits remain.

## C4 focused review readiness and remaining platform gate

All twelve C4 focus/profile cells completed with no new substantiated defect. Eleven are scoped APPROVED; build/distribution gpt-6-luna/medium is INCOMPLETE solely for native Windows activation, repair and PowerShell execution. The parent retains that material platform gap; there is no final consolidated approval. Existing source corrections and actual Linux acceptance support proceeding with the reversible local implementation commit, upstream integration and required fullmake verification. Native Windows must be resolved before completion and AL-80 unblocking. Detailed evidence is retained while this gap remains.

## Integrated gate and C5 test-only correction

Committed C4 implementation7a45ad23 and merged documentation-only main731de946 at fb0d3dccc5e2da3f1004a973cf5ca31acf3995e3 without conflicts. No runtime/test source changed through integration. The first fullmake failed three stale-policy assertions: TEST-02 CLI external-parent rejection and TEST-03 shared-only scan/migration machine-state prerequisites. The test engineer traced both against approved acceptance, transferred malformed-value/nonmutation checks and added positive controls; focused race checks pass. Production is unchanged. Parent diagnostically stopped the already-failing run's still-busy capture package with SIGQUIT at337s; this is not a natural capture assertion failure or completed capture result. The frozen C5 three-test-file correction receives narrow test-effectiveness review under both profiles. Other C4 production reviews carry with unchanged-code rationale; nativeWindows review remains incomplete.

C5 narrow test-effectiveness review is scoped APPROVED under both gpt-6.1-sol/high and gpt-6-luna/medium. Both independently verified assertion transfers; targeted race evidence is revision matched. Production C4 reviews carry unchanged. Fullmake rerun remains active; nativeWindows platform gap remains material. No final consolidated approval or completed-review cleanup.

The measured single capture race fixture passed32.81s (33.86s wall), with25,751,408-byte executable,34,350,514-byte pending journal and999,640KiB maxRSS. This explains substantial runtime without establishing a hang or full-suite timeout. The full rerun retains the default10m Go timeout and all assertions. Raw /tmp/al79-tester-integrated/capture-race-narrow.log; parent owns full-run outcome. Native Windows requires PR-triggered CI; an intermediate draft verification PR will preserve INCOMPLETE status and AL-80 blocking until that material stage passes. It is not the final delivery/approval checkpoint.

The draft verification PR is scheduled concurrently with the still-active C5 fullmake rerun to obtain otherwise unavailable hosted Windows evidence. Source is frozen at the reviewed C5 content; this checkpoint commits artifacts only. Neither an incomplete review nor a pending make is promoted to approval. Final readiness, completed-review cleanup, lifecycle promotion and final telemetry remain gated on actual results.

C5 fullmake ended exit2 after the full capture package completed442.180s. All earlier failing packages pass; the only failure is TEST-04, a remaining adapter assertion expecting an explicit child ticket to conflict with its parent. This is superseded by DATA-03 acceptance, not a confirmed runtime defect. The test owner will retain the original successful-read/post-validation-failure diagnostics oracle using genuinely inherited conflicting ownership and explicitly verify independent child success. The actual immutable-journal checkpoint helper also gets one-read path caching without removing durable-prefix assertions; measured repeated decoding accounts for substantial verification cost. Both test files receive narrow review.

Draft PR130 has no CI runs because GitHub reports a merge conflict. New main76fa93b5 changes backlog documentation only, including AL-79 specification rename and index rewrite. Parent will merge after test freeze/commit, preserve upstream documentation and AL-79 In Progress/AL-80 Blocked status, then push to enable actual Windows CI. Accepted scope and all runtime source remain unchanged.

## C6 integrated gate and actual Windows result

Fullmake PASS exit0 on reviewed C6 runtime/test source8b15b2ee after integration of main76fa93b5, including complete capture lock matrix, Go race, frontend integration678tests and canonical multi-platformbuild. Raw /tmp/al79-integrated-make-c6.log. Both C6 test-review profiles APPROVED. NativeWindows job128826559416 on the samehead failed; watchdog and installer-inspection process tests passed, distribution step failed and plugin transport was skipped. ActualCI exposed PowerShell logicalcwd/Win32cwd mismatch and successful native diagnostic suppression; C7 productowner corrects only the two wrapper copies and setupdownload.ps1. Distribution owner separately repairs fixture lifecycle, repeated-install and supportedbootstrap-encoding errors while preserving actualruntime oracles. No Go/TS runtime change or scope expansion is authorized. Final approval staysINCOMPLETE until affected review and actualWindows execution pass.

## Explicit user acceptance delta

User2026-11-08: "Remove windows test CI, I will remove support from windows later". Iteration02 records approved removal of the Windows CI job/dependency and associated delivery blocker. Runtime/support and release targets remain. Windows incomplete evidence is retained as partial/accepted limit, not passed or hidden. Further Windows execution stopped; existing completed C7 PowerShell fixes and coherent fixture corrections remain for narrow review, unfinished loaded-upgrade extension is withdrawn. All other contract guarantees and fullmake evidence remain applicable.
