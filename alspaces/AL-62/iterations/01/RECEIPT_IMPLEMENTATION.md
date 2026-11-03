# CI-PUB-01 receipt durability implementation

Baseline: b4e3c6a6bd4738536514f4d1502b53644fb31847. Approved bounded repair; discovery credit remains reviewer-data, gpt-6.1-sol/high, severity5. No predecessor integration, full make or merge readiness is claimed.

## Protocol

New capture-local metadata table `capture_auxiliary_publications` is keyed by verified worktree, ticket and canonical physical ledger destination. Its explicit version1 record retains at most active R0 plus pending R1. Each immutable generation contains a random publication ID, exact ledger SHA-256 digest and trusted exact auxiliary snapshot, including empty snapshots. Generation IDs/digests are publication metadata, never source facts. Existing fact-ledger schema and file journal are unchanged. Payloads are bounded; malformed/unknown versions/fields/identities fail closed.

| Durable file state after existing validated recovery | Receipt action |
| --- | --- |
| Restored L0 | Select active R0; discard unfinished pending R1 |
| Committed L1, finalization interrupted | Promote matching pending R1 atomically |
| First publication restored absence | Clear pending only when no active proof exists; retain a bound empty metadata state, not an invented old receipt |
| No matching trusted generation | No receipt evidence; explicit independent-proof recovery or failure |

CLI holds its existing publication lock. It first performs existing journal validation/recovery, then selects metadata. Independently authorized old unbound evidence is explicitly migrated before the next file journal. After writing trusted normalized facts and requested summary, it stages R1 before file commit, then finalizes after commit. R0 survives the rollback window; R1 survives the opposite window. Selection/migration/rebinding/finalization reject any remaining journal marker. All operations are idempotent; equal artifact digests cannot replace a trusted snapshot with different facts.

## Boundary APIs and compatibility

`SelectPublicationAuxiliary` returns only matching receipt evidence and a bound flag. `MigratePublicationAuxiliary` establishes an association now, after independent baseline proof; it never claims historical destination binding. `StagePublicationAuxiliary` and `FinalizePublicationAuxiliary` coordinate exact trusted snapshots with existing file publication. Accepted registered snapshots remain separate from receipt selection through `RegisteredAuxiliaryEvidence`; legacy `AuxiliaryEvidence` reads remain available to unmigrated callers.

Valid bound metadata with an unmatched digest returns `ErrAuxiliaryArtifactMismatch`, never legacy fallback. The CLI excludes both receipt kinds and can use only independently verified registered/pinned-Git evidence (or genuinely empty auxiliary facts with existing owner checks). Only after that validation may named `RebindPublicationAuxiliary` replace an idle destination association. This preserves the existing alternate-registry newer-canonical case without adding metric authority. Pending generations and markers prohibit rebinding. Missing old proof cannot be reconstructed from ledger/journal bytes; unavailable independent proof remains a recovery error with stored artifacts preserved.

Pure normalization of already trusted factLedger determines staged auxiliary rows; raw written bytes provide only the artifact digest. Exact provenance/max-per-snapshot multiplicity validation is unchanged. The new operational table participates in existing relocation write fences and destination conflict checks; legacy protocol1 import and unsupported shared-protocol2 relocation behavior remain unchanged.

## Evidence

- Existing CLI capture/history/authority/generation checks: `go test ./cmd/agentic-lens -run '^TestCapture(ManualAuxiliary|Committed|Merged|Generation|AuxiliaryOnly|CLIGenerate)' -count=1` PASS23.355s; `/tmp/al62-receipt-cli.log`.
- Existing adapter receipt/reconcile/relocation checks passed13.444s before the final named-rebind/close-error refinement; `/tmp/al62-receipt-adapter.log`. Tester reruns final affected cases independently.
- Existing legacy v4 and hostile-Git fixtures explicitly model pre-binding metadata rather than bypassing production bound-digest guards. The newer-canonical fixture stages/commits/finalizes a trusted generation. Their original cost/history/multiplicity/source authority assertions remain.
- Product candidate frozen for independent tester receipt-window, migration, absence, malformed metadata, scoping and relocation regressions plus focused race checks. Tester owns the two new regression files and reports results separately.

Limits: interruption states are deterministic crash-state reconstruction, not power-loss certification. TestMain subprocess CLI builds are not automatically race-instrumented. Parent owns final review/consolidation, commits, telemetry and draft PR publication; full gate/integration holds remain in effect.
