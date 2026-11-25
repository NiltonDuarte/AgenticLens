# AL-94: Technical design

## Scope and boundaries

Implement the behavior defined in the [explicit backfill spec](specs/explicit-backfill-refresh/spec.md). AL-94 covers explicit source refresh, requested summary maintenance and operator reporting. AL-95 owns opt-in preview, representative rollout measurements and any default change. Preserve current CLI exit semantics, heuristic opt-in, Force behavior, workspace resolution and portable ledger history.

## Decisions and ownership

- The **coder** owns implementation in `backend/internal/adapters/backfill/`, supporting summary or publication helpers in `backend/internal/adapters/` as needed, and backfill CLI output in `backend/cmd/agentic-lens/`. The coder also owns any backfill usage documentation needed for the new report.
- The **tester** owns focused test files for the backfill adapter, shared publication helpers and CLI. The **reviewers** independently check the implementation and its tests. The **orchestrator** owns integration and change evidence, telemetry and PR delivery.

Use the existing persisted-fact comparison after candidate validation and history preservation. Keep stable migration steps and timestamps on equivalent repeats. Plan the summary independently from the final accepted or candidate ledger. Compare the full rendered summary under the existing workspace publication lock, using the current managed-block parser. If neither file changes, skip publication. When publishing changes, preserve the existing ownership, cancellation, transaction and recovery guarantees. Count a write after successful publication.

Carry owner and collection outcomes from the exact-source readers into the report. Use the recorded owners of each selected harness as the baseline and classify successfully read owners against the final persisted facts. Keep retained collection evidence visible even when the session has some newly verified facts. Preserve failure and cancellation causes, including a store preflight failure that blocks work on both harnesses. Keep coverage, migration status and publication outcomes independent.

Reserve OpenCode reads for exact recorded owners before reading newly discovered descendants. If publication is rejected, retain verified source observations in the report and identify candidate changes as uncommitted. Store preflight can account for recorded owners by reading accepted ledgers without opening a publication transaction.

## Implementation approval

The user explicitly approved implementation with “$agentic-lens-dev aprovado” after reviewing the completed OpenSpec artifacts. Approval covers the AL-94 scope and acceptance linked above. The implementation baseline is `db4477fa`.

## Implementation approach

Reuse FIX-20's ledger equality, ordering and cancellation behavior and FIX-19's exact-owner ambiguity handling. Keep existing publication safeguards for summary-only repairs and combined ledger and summary updates. Verify native Codex and OpenCode fixtures with distinct complete, partial, all-unavailable and blocked selections. Check bytes and mtimes on unchanged repeats and summary-only repairs. Check mixed-harness preservation, real Git and migration changes, budget exhaustion, targeted retry and existing failure exits. See [FIX-20 verification](../fix-20-stop-source-reads-when-callers-cancel/delivery.md) and [AL-65 discovery](../al-65-recommend-safe-source-backed-backfill-defaults/delivery.md).

Independent review covers CLI behavior and integration, publication and recovery integrity, architecture and maintainability, and test effectiveness. Review the path from native source reads to the final report, and from summary comparison under the workspace lock through commit or recovery.

## Steering

During OpenSpec exploration the user accepted keeping `Updated` as the processed-successfully count and adding separate counts for ledger and summary writes. The proposal and spec adopt this additive reporting in place of redefining `Updated` as a write count. Preserve existing processing and exit semantics. Verify the legacy count alongside committed write counts on a changed run, an unchanged repeat, a summary-only repair and a blocked selection.

Integration with AL-98 (Claude Code sources) was requested on PR #162 without changing approved acceptance. Claude Code recovery now follows the same coverage contract as OpenCode and Codex. It reports per-owner reads, failures and retention reasons, accounts for cancellation, and reports a blocked preflight for an unusable store. A Claude Code test expectation was aligned with the existing AL-94 rule that collections count as retained only when an owner was read. FIX-21 changes automatic capture only and leaves explicit backfill unchanged, so its merge needed no backfill decision. Verification impact: the backfill package, integrated `make` and a quick independent review cover the merge.
