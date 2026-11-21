# FIX-20: Technical design

## Scope and boundaries

Propagate cancellation through the live OpenCode observer, historical OpenCode and Codex hydration, and explicit source-backed backfill. Verify the shared exact OpenCode reader on both schema families. Preserve current source selection, Codex ambiguity rejection, portable ledger facts and the existing policy for ordinary partial refreshes. Broader source lifetime, ordering, timeout policy and cache redesign remain outside this change.

See [live and historical acceptance](specs/live-source-reporting/spec.md) and [backfill acceptance](specs/backfill-source-cancellation/spec.md).

## Decisions and ownership

- The implementation owner carries the caller context through native query, discovery and read loops. Apply the existing per-workspace deadline as a child of the backfill caller context. Stop before opening more sources or retrying extraction once the relevant context is cancelled. Preserve existing Codex discovery and decompressed-byte accounting.
- Live cache admission must retain revision checks around native reads and leave accepted snapshots usable until revision invalidation. Make waiting for the cache cancellable, and stage lookup eviction and snapshot insertion so a cancelled request cannot change the accepted cache. Keep ordinary child lookup diagnostics and usable partial data.
- Keep the exact OpenCode schema probe, budget checks and extraction on one dedicated connection within one read snapshot. Cleanup must work after cancellation. Stage complete-owner evidence until the read and cleanup succeed. Historical hydration can return usable gathered evidence with incomplete coverage and should avoid permissive rereads after cancellation.
- The tester owns self-contained regression fixtures and tests. Reviewers cover cancellation through public entry points, resource lifetime, cache revision races and publication consistency. The orchestrator owns integration and delivery evidence.

## Implementation approval

The user approved implementation with “$agentic-lens-dev aprovado, continue” after receiving the artifact reading order. The approval covers the linked proposal and specs, including preservation of the current workspace when cancellation precedes publication. Planning and implementation start from `83f6f6c1`.

## Implementation approach

Verify cancellation before work, while waiting for the cache, and during actual SQLite and Codex reads. Synchronize tests with the read boundary and use a bounded deadline as a failure guard. Record driver and fixture conditions, verify that the next independent read succeeds, and check revision invalidation separately. At the backfill boundary, exercise a complete owner followed by an interrupted owner, compare original ledger and summary bytes, and verify earlier committed workspaces. Retain healthy v1, v2 and Codex controls, including mixed-owner history and ordinary partial refresh.

Independent review covers public behavior, ledger and cache integrity, adapter and resource ownership, and the strength of the cancellation fixtures. The pinned SQLite driver can interrupt initial query execution, but one review probe demonstrated that a later streaming step can still hold the trace lock after cancellation. Keep expensive native source evaluation within the driver's interrupt window and verify a controlled later-row case. Checks at row boundaries and bounded rollout reads limit further work after cancellation. Ordinary synchronous filesystem calls may complete before control returns.

## Steering

In exploration, the user selected option 1 for backfill cancellation. Before publication starts, discard prepared updates for the interrupted workspace, even if some owners finished. Preserve its existing ledger and summary and keep earlier committed workspaces. This settles the earlier choice between publishing completed owners and preserving the entire interrupted workspace. Ordinary source failures still follow existing partial-refresh policy. Verification now requires a byte-for-byte comparison of both artifacts after an owner has finished and the next owner is interrupted.

The user expanded FIX-20 with “implemente no mesmo FIX-20” to make capture observation opt-in. The next successful automatic publication replaces the generated metrics block, removing any previous observation when the option is false. The ledger, standard metrics, authored content and capture diagnostics are preserved.

The user then required configuration writing to emit the default explicitly and older shared YAML that omits the setting to fail validation. This supersedes the earlier omission fallback. Verify that setup writes `false` by default, preserves an explicitly supplied `true`, and that the loader reports an omitted key even when summary publication is disabled. The shared change is confined to configuration loading and serialization and their publication consumers. Carry forward the cancellation, cache and ledger review results because those paths are unchanged.

The user requested stable ordering for every array in the raw ledger after partial publication moved 35 of 38 sessions in a Git diff. Order sessions by ID after refreshed and retained records are combined. For reserialized facts, order unordered collections by stable keys, keep chronological collections in timestamp order with source order for ties and unknown timestamps, and preserve the declared migration step sequence. Preserve the nested arrays of untouched historical records in their recorded order. Generation, source backfill and schema migration use the same policy for freshly serialized facts. Metadata-only cleanup and the whitespace formatter retain the input order. The first rewrite can normalize the session list and the arrays of refreshed facts. The approval to implement is the explicit request to keep all ledger arrays stable.

The persistence review traced two existing backfill transformations that defeated that contract. Converting an unavailable session through the domain representation lost its stored anomaly evidence. Rebuilding migration status alphabetized the surviving steps. Backfill must retain the supported stored session representation when its source is unavailable and preserve the relative order of surviving migration steps. Ticket-level anomaly summaries and their references form one recorded snapshot that remains authoritative during incomplete source coverage. A complete analysis replaces that snapshot together. Verification exercises these boundaries through backfill and the domain anomaly tests.
