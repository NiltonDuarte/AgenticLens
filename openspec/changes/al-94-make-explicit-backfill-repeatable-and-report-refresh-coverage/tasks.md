# AL-94: Execution tasks

Start execution after explicit scope approval is recorded in the [design](design.md). Verify behavior against the [spec](specs/explicit-backfill-refresh/spec.md).

## Implementation and focused verification

- [x] Coder: Reuse the ledger comparison and make summary replacement conditional under the publication lock. Preserve pending migration state on equivalent repeats, genuine Git and migration changes, and cancellation and rollback behavior.
- [x] Coder: Carry exact-owner and collection outcomes through source recovery, and report them separately from migration and committed writes. Preserve `Updated`, flags and exits. Report unavailable-store preflight and budget-limited retries, and document how to interpret the counts.
- [x] Tester: Exercise exact-source Codex and OpenCode backfill on self-contained native fixtures. Cover complete and still-partial repeats, a changed parent with a missing child, unrelated harness history, verified-empty collections, all unavailable owners, and unsupported or ambiguous source evidence as applicable.
- [x] Tester: Assert bytes, mtimes and reported writes for identical summaries, missing or stale summaries with authored text preserved, and calls without summary updates. Exercise DB-only and combined DB/Git calls, a resolved migration step, ordinary selection and Force.
- [x] Tester: Exercise the CLI for a missing selected store, including `--source both`, and for batch budget exhaustion followed by an exact-ticket retry. Reconcile per-ticket and aggregate counts and confirm existing exit and preservation behavior.

## Integration and delivery

- [x] Orchestrator: Integrate with current main and record the candidate revision, commands, results and limits for affected checks and the full `make` gate.
- [x] Reviewers: Independently check publication and recovery, coverage accuracy, CLI compatibility and regression tests. Orchestrator records findings and their dispositions under the change's evidence directory.
- [x] Orchestrator: Record verified outcomes and PR context in delivery, then extract and verify final source-backed telemetry before the artifact commit and PR handoff.

## Workflow follow-up

Archive after verified delivery. Reconcile roadmap completion separately.
