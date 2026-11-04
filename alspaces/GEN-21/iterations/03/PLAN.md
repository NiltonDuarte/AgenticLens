# GEN-21 iteration 03: Console-only completed migration audit

## Steering and authorization

Source: on 2026-11-04 the user asked, "can you remove it? I prefer it to be printed out in the console, so user can revert the changes if they want to", after the explanation that completed schema-6 migration metadata remained solely for its audit. This directly approves removal of completed metadata and console audit output. No renewed approval is required for this bounded implementation. Baseline: 8b47f0a; preserve unrelated checkout changes.

Changed acceptance: completed migrations with no pending steps omit the migration member from persisted ledgers, including the already migrated repository history. Their audit remains available in explicit migrate/backfill command results and console output. Preserve pending and partial migration state unchanged; no schema bump or source refresh is required for metadata cleanup.

Superseded decision: iterations 01/02 retained completed ownership/loss audits in ledger migration objects. Their original verification and hashes remain historical evidence, while iteration 03 records metadata-only cleanup hashes. Preserved guarantees: all recorded session facts, source ownership, ordering, multiplicity, invocation amounts, partial-state resumability, runtime validation, atomic publication and exact explicit backup behavior.

## Grounded implementation boundaries

The domain step backend/internal/domain/migration/step_v5_v6.go, StepV5ToV6LeanOccurrences.Apply, creates ownership/loss accounting in MigrationInfo. Keep this pure transient result so the shell can report every count, including nonzero losses. Domain audit assertions in backend/internal/domain/migration/lean_v6_test.go remain valuable. Do not remove audit production merely to remove its persisted representation.

The persistence boundary backend/internal/adapters/backfill/migration_runner.go, MigrationRunner.migrateWorkspace, currently copies Migration.Notes to TicketSchemaMigrationResult.DiffSummary and then serializes the same metadata. It also returns up-to-date before processing valid current-schema ledgers with completed migration objects. Add explicit normalization of such objects, report the retained audit, then omit the object before persistence. Validate the raw current-schema input before any rewrite; invalid, unknown or unsupported members must not disappear through permissive decoding. Metadata-free current files remain byte-identical and up-to-date. Cleanup must not change updated_at or other ledger values solely to erase migration metadata. Do not globally change read/DTO serialization, since ordinary reads must never migrate or normalize historical files silently.

The corresponding boundary backend/internal/adapters/backfill/engine.go, processWorkspace migration evaluation near hasOwnershipAudit, already processes cleanupCompletedMigration. Remove the ownership-audit exception that keeps completed metadata, while retaining the existing audit in TicketMigrationResult.Notes and console output. Completion cleanup requires both status completed and zero pending steps. Contradictory completed metadata containing pending steps, partial status and pending status must never be silently cleared. Preserve existing source-availability semantics; this change does not authorize clearing source_backfill or refreshing unavailable owners.

CLI presentation lives in backend/cmd/agentic-lens/main.go, runMigrate and runBackfill, with typed result structures in backend/internal/adapters/backfill/report.go if needed. Emit audit details for both newly migrated files and cleanup of already completed metadata, with the actual ledger path so nested workspaces sharing the same ticket ID are unambiguous. Preserve completed-step and recorded audit information in console/report output before erasing its persisted representation. Distinguish dry-run plans, successful writes and errors truthfully.

Retain the existing explicit --backup option; do not introduce a new automatic-backup policy. Successful --backup output identifies the actual .bak path and explains that restoring that file reverts the ledger. Without --backup, console guidance identifies Git/original backups as recovery and recommends --backup before destructive legacy migration. A dry run creates neither ledger changes nor backups. The existing .bak behavior is unchanged; printing an audit is not itself a reconstruction mechanism for discarded legacy fields. Do not falsely claim a backup was created on a failed or dry-run operation.

## Writable ownership

One coder owns production changes to backend/internal/adapters/backfill/migration_runner.go, engine.go and report.go; backend/cmd/agentic-lens/main.go; and canonical documentation docs/OWNED_OCCURRENCES.md plus .agents/skills/agentic-lens-operator/references/cli.md. Change additional production paths only if required by the same boundary, and report them to the parent. No frontend or configuration changes are planned.

A tester owns focused test updates in backend/internal/adapters/backfill/migration_runner_test.go, engine_test.go and backend/cmd/agentic-lens/lean_migration_test.go, with coherent adjacent existing tests only where needed. The coder may make a minimal existing-test expectation change to keep its focused checks runnable, then hand ownership explicitly to the tester. The parent owns GEN-21 artifacts, real-ledger cleanup, review coordination, integration, telemetry, commits and PR delivery. Agents share the checkout and must preserve each other's changes.

## Acceptance and valuable regression coverage

1. A real v5-to-v6 fixture with ownerless rows/retired payloads reports the nonzero audit to console/result, keeps owned facts/costs/proven child IDs and writes schema 6 without a completed migration member. Extend TestCLIMigrationUsesDirectV5ToFinalV6AndRejectsUnreleasedSchemas; replace its persisted-audit oracle with absent metadata plus console counts, without weakening its existing rejection, ownership or backup checks.
2. Explicit migrate of valid schema 6 with completed/no-pending metadata reports that saved audit and removes only migration. Cover a nested ledger path, exact byte backup, dry-run no write, and a second execution remaining byte-identical with no repeated migration. Extend the migration runner's existing backup/dry-run/current-ledger coverage rather than adding generic fixture-only tests.
3. Backfill after successful completion removes completed ownership audits and retains them in the report. Update the multi-machine completion assertion around engine_test.go:560 to check absence plus result audit. Retain partial and missing-source assertions in the same scenario so audit cleanup cannot erase a resumable blocker.
4. Completed status with nonempty pending steps is not a cleanup candidate; partial/pending metadata remains preserved or fails explicitly under current runtime admission. Malformed or unsupported current-schema input is not rewritten and creates no false successful backup. Extend coherent existing negative cases.
5. Verify all recorded data except the completed migration member remains equal across real-ledger cleanup. Do not write tests that read live alspaces history; use temporary isolated fixtures for automated tests. Full make is required after integrated application changes.

Applicable regression boundaries: portable ledgers without original sources; pending migration handoff between machines; explicit loss accounting with rollback backups; configurable roots; actual current-schema wire validation; unsupported/malformed sibling isolation; idempotent publication. No original source stores, capture-state repair or new authenticity policy is required.

## Documentation alignment

Update the owning schema contract docs/OWNED_OCCURRENCES.md to state that completed migration audit is console/report output and absent from canonical persisted history; partial/pending metadata remains resumable. Update operator migrate/backfill CLI guidance with this behavior and explicit backup/restore limitations. Search active contributor guidance for claims that completed audits must persist, but preserve historical ticket records. TECHNICAL_GUIDELINES.md's loss-accounting requirement remains satisfied through console/report audit and should link to the canonical owner rather than duplicate its details. The operator rule forbidding manual removal of partial metadata remains valid.

## Review roster and verification ownership

Freeze the implemented revision/diff and dispatch all four focused personas in both configured profiles, in capacity-bounded waves. Each reviewer gets the contract, this plan, baseline, exact reviewed revision, its own persona and read-only product paths. Reports are working artifacts outside Git under /tmp/agenticlens-GEN-21-review03/.

- reviewer-behavior, gpt-6.1-sol/high: behavior-sol.md; CLI audit visibility, repeated cleanup and truthful status/rollback output.
- reviewer-behavior, gpt-6-luna/medium: behavior-luna.md; same independent behavior scope.
- reviewer-data, gpt-6.1-sol/high: data-sol.md; pending-state preservation, raw validation, unsupported input and unchanged recorded facts.
- reviewer-data, gpt-6-luna/medium: data-luna.md; same independent data/failure scope.
- reviewer-architecture, gpt-6.1-sol/high: architecture-sol.md; FCIS, explicit persistence boundaries and documentation consistency.
- reviewer-architecture, gpt-6-luna/medium: architecture-luna.md; same independent architecture scope.
- reviewer-tests, gpt-6.1-sol/high: tests-sol.md; changed oracles detect console-only audit without weakening preservation/rollback tests.
- reviewer-tests, gpt-6-luna/medium: tests-luna.md; same independent test-effectiveness scope.

The parent owns consolidated findings, corrections, revision-aware rechecks and approval. Retain final compact REVIEW_SUMMARY.json with numeric severities, brief findings, all discovering persona/model/effort tuples and dispositions; link it from SUMMARY.md, then remove resolved detailed reports under the canonical lifecycle. No incomplete review passes delivery.

## Operational cleanup and delivery

After implementation and focused checks, build the changed CLI and use its supported migrate command for cleanup, first with dry run and then explicit --backup. Before invocation, save the current 119 ledger originals outside Git; preserve printed per-ledger audit/rollback output locally and a concise structured cleanup manifest in iteration 03. Restore only the completed migration member's absence, not source facts: compare each result semantically to the original with that one member removed. Verify partially migrated files, if any, remain intact. Move verified .bak files outside Git after recording their locations; committed prior revision remains durable rollback history. Re-run migration to verify no further writes and scan the final history.

Commit task-owned implementation, integrate current origin/main and run make. Update summary/PR around the final delivered behavior and operational cleanup. Generate final telemetry only at the last step; if the previously recorded local capture-coordinate incompatibility remains, report the failed extraction without fabricated ledger or registry reset. Check PR #110's state before delivery and publish any undelivered work through the correct open/new PR, link it immediately and report the final URL.
