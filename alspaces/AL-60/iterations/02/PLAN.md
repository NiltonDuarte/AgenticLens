# AL-60 Iteration 02: Shared Capture Repair

## Authorization and preserved boundaries

Explicit user approval “Authorize AL-60 shared capture repair” superseded iteration 01's read-only hold. Clean checkpoint: 7b00cc1ac7fb5cd8afa2d93f7c65931c8ab2fa30. The [contract addendum](CONTRACT.md) preserves reviewed search at 7a50a5e, AL-55 telemetry, source fidelity, cancellation, ownership, pending publication and recovery. No blanket timeout increase, schema or production worker change was authorized. MGT-2 owns merging; later owners integrate after the real merge.

## Execution and steering record

- Planner authored the addendum/plan. Coder owned installer phase/diagnostic changes and narrow subprocess/transaction regressions; tester owned deterministic harness fixtures and worker success-versus-expiry coverage. All shared the checkout and preserved others' paths. Parent owned integration, review reconciliation, telemetry and publication.
- Deterministic blocked-probe red preceded production repair. Controlled manifest probe variation established only optional field differences, not the cause of the original unretained failure. [Verification](VERIFICATION.md) retains commands, outcomes and limits.
- First review: all four responsibilities under both configured profiles examined fingerprint e43f3b57f5979e59324599d75e279e2abf6d5099caccbfbe8bdcbdd33aa70770. ARCH-01, ARCH-02, DATA-01 and TEST-01 were corrected without changing accepted boundaries.
- Intermediate fingerprint 9f418c930f64269bc3da5d2f02aa18627d187729c5b71cb840fdbd6c23613ff0: both architecture profiles independently found DATA-02 in completed recovery. Three reviews ran; five queued cells were superseded before dispatch, not counted as approvals. The recovery checkpoint and deterministic retained-journal/retry guard corrected it.
- Final fingerprint 5e332cad1bfb0d1b25af2d9a8def5bfbdcd8e2b8364c3023131733041d4c8493: every responsibility/profile completed APPROVED. The integrator verified actual Windows execution via WSLInterop, separately from cross-compilation.
- Full integration exposed a vet warning about a deferred closure capturing reassigned cancellation functions. Explicit first/second defers preserve contexts, ordering and final cancellation; the extra cancellation of the already-canceled first phase is idempotent. Scoped reviews carry through this mechanical change, supported by vet/race guards. Integrated fingerprint: 3251b53bc2bfdbd9c7e6ea47d5b8e793fc22f4f9e6103ee268d9fbc622a005db.

## Approved outcome and retained review

Nineteen reviews executed across the three candidates. Behavior/integration, data/failure, architecture/maintainability and test-effectiveness each completed the final candidate under gpt-6.1-sol/high and gpt-6-luna/medium; execution identity is dispatcher-acknowledged. [Compact findings](REVIEW_SUMMARY.json) retain every distinct discovery tuple and disposition. [SUMMARY.md](../../SUMMARY.md) distinguishes discovery, subsequent verification after plan exposure, fix actors and accepted TEST-02 coverage debt. No current production leak was demonstrated by that recommendation; AL-60 maintainer owns the Unix descendant-oracle follow-up.

Full make and actual integrated Windows checks passed on 057d410253aecd734551995ff1bedb38e94a1f19. Main integration added only AL-68 workspace documentation. Completed detailed review reports/run records were compacted and removed. Required contracts, plans, ledger and compact findings remain; verification outcomes and failed-gate history are summarized in VERIFICATION.md with immutable links to archived details.

## Pre-merge artifact cleanup

User first requested artifact cleanup, then explicitly superseded raw-output retention: remove per-command evidence and indexes, keep the verification summary, and do not engage the dev workflow. This pass changes artifacts only. Compact findings, contracts, ledger and accepted-debt records stay intact. Historical failures and useful detailed output are linked to immutable existing commits from VERIFICATION.md. Verify links, preserved hashes and source equality to tested 057d410, commit/push without rewriting history, and hand MGT-2 the new head/inventory. Merge remains held for refreshed exact-head checks.
