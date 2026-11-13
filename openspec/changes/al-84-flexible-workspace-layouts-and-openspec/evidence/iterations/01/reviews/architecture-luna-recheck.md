# AL-84 iteration 01 architecture correction recheck

Outcome: **APPROVED** for the assigned architecture scope.

## Reviewed candidate

Baseline `163181428ee9a5a599a2da1e91d3af5e618a9682` plus final corrected 673-file candidate, manifest embedded-content SHA-256 `9a78b592e2863c8ed07ef99f89f37a41d7052855d52749273c6fe4a9f40aab6`. Independently recomputed every listed file hash: 673 files, zero mismatches. Product and tests remained read-only. This report is the sole write. Requested run identity: AL84 iteration01 final architecture correction verification; reviewer `reviewer-architecture`, Luna/medium as assigned by the task. No claim about runtime model metadata beyond that requested identity.

## Findings

No blocking architecture findings in this scoped recheck. No new finding (severity 0).

## Correction dispositions and hypotheses

- **ARCH-01, private recovery ancestry durability — corrected.** `MkdirPublicationStateDirectories` walks to an existing ancestor, syncs it and its parent, creates each missing directory, then syncs both the new directory and its parent. `AcquirePublication` completes this before writer admission; the publication set creates journal namespace directories through the same durable helper before private journal publication. The supplied raw private-state `strace` receipt independently checks fresh-state and new-project ordering, requiring each created directory and parent sync before the pending-marker rename and public replacements. Four injected `EIO` cases fail before the marker and preserve accepted bytes. This is syscall-order/failure evidence, not a power-loss experiment.
- **BEH-02/03, incomplete discovery and maintenance resolution — corrected.** `resolveTicketLocation` preserves discovered partial results/diagnostics, but permits an implicit unique selection only when enumeration completed. Explicit selection is separately checked against configured physical scope and read directly, so a valid in-scope home omitted by the scan budget remains selectable. Backfill and migration share `ResolveMaintenanceLocation` through `discoverWorkspaceLedgers`; only a typed `not_found` with no explicit selection becomes an empty batch. Ambiguous, unavailable, invalid and incomplete states remain errors. Blank ticket metadata does not preclude explicit maintenance selection, while an explicit ticket filter still checks recorded identity.
- **Shared authority / scope carry-forward — no competing workspace resolver found in these corrected paths.** Runtime ledger projections continue to validate the supported representation; maintenance selection reads recorded ticket identity before its owning formatter/migrator validates historical schema. This keeps selection usable for metadata-only and historical-schema inputs without adding a second selection policy. Review scope carries forward the previously inspected FCIS, manager/engine, OpenSpec-roadmap separation, and selected root/ledger authority; this recheck does not re-certify unrelated behavior or data paths.

## Evidence and limits

Inspected the approved design/spec/tasks, active review and initial architecture reports, supplied findings, final manifest, correction receipts/parser, and corrected resolver/publication source. The supplied native receipts report corrected first-generation, incomplete/explicit/duplicate-selection, archive/refresh, backfill and migration cases; maintenance-after receipt reports explicit hidden and nested homes, expected rejection of incomplete implicit selection, and valid complete-empty batches. Reported backfill ordinary/race durations are 8.406s/29.055s and affected CLI duration 38.724s; production vet passed per supplied evidence. These receipts were read, not rerun. Parent integrated `make` remains pending. This is scoped architecture approval, not a whole-audit verdict or a claim of power-loss or cross-platform durability proof.
