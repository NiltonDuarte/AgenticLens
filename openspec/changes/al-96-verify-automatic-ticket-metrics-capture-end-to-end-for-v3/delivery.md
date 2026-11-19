# fix(AL-96): capture session metrics before ticket binding

## Outcome

The installed worker captures native session facts before a ticket or its destination exists. It writes a Git-visible snapshot in the configured `agenticlens.pending_dir` and keeps the same capture identity after ticket binding. Each completed native interval settles in the pending snapshot while it waits for a destination.

Once the bound agent indicates an eligible directory, the worker publishes `agentic_metrics.json` and any configured summary. It removes the pending snapshot after the ledger is durable. An existing canonical ledger takes precedence. The worker retains facts when publication is blocked and follows a canonical ledger moved within configured roots.

The [design](design.md#capture-before-ticket-binding) records the revised scope. A separate roadmap ticket [AL-97](../../../alspaces/plan/to-do/AL-97%20warn%20about%20staged%20pending%20capture%20snapshots.md) covers the advisory Git hook.

## Revised verification

The [native gate](evidence/native-gate.md#integrated-pending-capture-package) passed on the checksummed Linux amd64 package `v3.0.0-rc.0+al96.8588f84d` for Codex 0.170.0 and OpenCode 1.19.5. Both clients captured four exact native owners, excluded an unrelated session, and left the Git index intact. Unbound completion settled in 0.987 and 1.078 seconds respectively. The first ledger and summary matched the native reconciliation cutoff, and the later turn preserved the recorded owners and facts.

Targeted Go race checks passed for native binding and destination correction, portable history, retained anomaly evidence, first-publication interruption, cleanup retry, staged facts and co-located directories. Setup tests passed for preserving valid policy and rejecting a pending path that overlaps effective private state before writing configuration. Nine oracle checks passed, along with strict OpenSpec validation.

The full integrated `GOMAXPROCS=6 VITEST_MAX_FORKS=2 VITEST_MIN_FORKS=1 make -j2` passed at `6c178e54` over `origin/main` at `a5c8a25e`. It covered race-enabled Go integration, 1,050 frontend tests, static checks, installation, distribution and multi-platform builds. Earlier integration runs surfaced frontend timeouts and two outdated capture fixture assumptions. The final run bounded concurrency and used the corrected explicit-type and asynchronous reconcile fixtures. Product and native runner evidence comes from `8588f84d`.

Independent behavior, architecture, data integrity and test reviews used gpt-6.2-sol/high and gpt-6-luna/medium. Behavior, data and test rechecks covered `8588f84d`. The architecture approvals at `b353bc76` carry forward because the final directory sync uses the existing retained lease, with its ordering checked by data review at `8588f84d`. Both test profiles approved the fixture changes at `6c178e54`. The [review summary](evidence/REVIEW_SUMMARY.json) preserves the findings and fixes in both delivery scopes.

## Earlier operator-preparation delivery

`capture prepare-publication` lets the operator select an empty destination for the first publication. The installed lifecycle binds the native session and the worker creates the ledger and any configured summary. The destination selection closes after publication or after an existing ledger is observed. Refresh follows the current ledger across configured roots. A pending local transaction recovers under publication ownership before a fresh write can proceed.

Codex capture checks native parent metadata to recover nested sessions without their own callbacks. Workers respond to eligible lifecycle requests during the periodic wait, retain normal retries for persistent failures, and finish a progressing child scan across bounded worker lifetimes.

The product candidate is `02ff2608` over `origin/main` at `a5c8a25e`. The runner at `32fd1347` passed the [ordered native gate](evidence/native-gate.md) on the checksummed `v3.0.0-rc.0+al96.02ff2608` package. Both clients published four exact owners and reported zero provisional seconds in their first completion-aware commit.

The full integrated `make -j2` passed on `32fd1347` with the final Go stress fixture diff (file blob `6e938471b937518584548b706ff4d9547b7ce4a2`). It covered Go race and 1,050 frontend tests, static checks, installation, distribution and native builds. An earlier integrated attempt hit the synthetic test's observer limit. The revised fixture retains bounded directory visits and actual worker handoffs while allowing publisher startup under race-suite load. It passed twice under concurrent Make load and in the full integrated suite.

Independent behavior, architecture, data and test reviews used gpt-6.2-sol/high and gpt-6-luna/medium. Follow-up checks cleared the affected boundaries. The [review summary](evidence/REVIEW_SUMMARY.json) preserves the confirmed issues and their dispositions. The focused Go races also covered protected Codex, concurrent prepared destinations, two pending tickets, interrupted renewals and first commits, imported history, relocation, deletion, worker timing, registry contention classification and bounded Codex discovery.

## Limits

Packaged native checks target Linux amd64, Codex 0.170.0 and OpenCode 1.19.5. OpenCode 2.2.0 standalone has prior manual extraction evidence. Locally packaged candidates establish the installed path on the tested build. Stable release publication is outside AL-96 delivery.

Capture status for the development worktree reported zero registrations and zero requested publication generations. The shared project selects the installed engine `v3.0.0-rc.156+9308ac7`. Its AL-96 ledger and contributor coverage remain unavailable. The packaged gate verifies independent native fixture ledgers through automatic capture.

## Pull request

Capture native session metrics before ticket binding with one portable Git-visible snapshot per root. After verified binding, the agent supplies the first ledger directory through the installed integration. The worker validates destination ownership, commits the ledger and cleans up the pending snapshot once publication is durable. Existing ledgers retain their canonical location and history.

The packaged native gate begins before binding and checks provisional time while work is open. It verifies completion in the pending snapshot, delayed binding, an unavailable destination and its retry, then automatic publication and a later turn. The worker captures source-proven descendants and preserves their facts across these stages.

The checksummed Linux amd64 candidate `v3.0.0-rc.0+al96.8588f84d` passed with Codex 0.170.0 and OpenCode 1.19.5. Both clients settled unbound completion within 1.1 seconds, published the exact four-owner tree at the first commit and preserved the staged Git entries. The subsequent turn retained previously recorded facts.

The full integrated `make -j2` passed at `6c178e54` with bounded Go and Vitest concurrency. Strict OpenSpec validation and nine native oracle checks passed.

## Pull request record

[AgenticLens PR 156](https://github.com/NiltonDuarte/AgenticLens/pull/156)
