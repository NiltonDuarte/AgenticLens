# GEN-20: Unique worktree directory identifiers

## Request and status

Replace sequential AgenticLens state worktree directory names with an existing worktree identifier when available, otherwise a short UUID. The user explicitly approved the contract and plan on 2026-11-03; implementation is in progress.

## Baseline

- Initial revision: dc44619ae15586fe1194f872743033f523635c92.
- Initial Git working tree: clean; original branch GEN-20/unique-worktree-identifiers.
- Ticket branch: chore/GEN-20-unique-worktree-identifiers.
- Configured workspace root: alspaces. No existing GEN-20 workspace or ticket index entry was found.
- Conventional classification: chore, as required for generic tickets by repository configuration.

## Work and verification

- Loaded configuration and development workflow; assigned a planner to ground the contract and iteration 01 plan in existing capture storage and relocation behavior.
- Prepared [CONTRACT.md](CONTRACT.md) and [iteration 01 plan](iterations/01/PLAN.md): persist qualified Git worktree identifiers or a 16-hex UUID fallback; preserve existing numeric directories and internal SQL IDs; update preview, installation and relocation consistently with a safe registry protocol transition.
- Existing allocation already uses immediate SQLite transactions and a unique worktree directory constraint. No committed-allocation race has been demonstrated.
- No application changes or test execution yet. No delivery or review approval claimed.
- Telemetry ingestion is deferred until the final delivery step, after approved implementation and verification.

## Implementation progress

The following records iteration 01; its backward-compatibility decisions are superseded by the user's iteration 02 clarification below.

- The coder implemented registry protocol 3 with a persisted token mapping while retaining handler and manifest protocols. Existing numeric paths are retained during the transactional protocol-2 upgrade.
- New synchronized first-open tests exposed SQLite contention during the initial WAL transition. The coder added bounded retries for writable constructors; broader verification is pending.
- During implementation inspection, the coordinator (gpt-6.1-sol, high reasoning effort) noted that validating the root through an arbitrary child named probe could let an unrelated reserved entry block all allocations. The coder (gpt-6.1-sol, high reasoning effort) also identified and corrected this during self-review, and passed focused coordination tests after the correction.

## Iteration 02 steering

The user clarified: "we dont need this compatibility. there is not old binary, the version that introduced the database was not released yet". This explicitly removes backward-compatibility requirements for the unreleased database, including old-binary protocol fencing and legacy numeric-state preservation. Unique stable directory allocation, concurrency isolation, safe paths, provisional previews and current-schema installation/relocation guarantees remain required. The planner is preparing the delta; compatibility-only implementation and tests will be removed. No automatic deletion of local state is authorized.

Iteration 01 coder verification before this steering passed: capture adapter tests; targeted CLI tests; focused coordination tests with race detection and three repetitions; and the full affected capture/CLI race suite. These results apply to the earlier diff; affected verification must be repeated after simplification.

Iteration 02 removes the registry protocol bump, upgrade/backfill, numeric fallback and compatibility-only migration test. Registry protocol 2 includes the token table directly. Unsupported local layouts fail explicitly without deleting or rewriting their state. The coder passed revised focused capture and CLI race tests covering allocation, alias validation, unsupported layouts, dry-run/manifest behavior, missing-state inspection and relocation. Independent test assessment and the integrated make gate remain pending.
