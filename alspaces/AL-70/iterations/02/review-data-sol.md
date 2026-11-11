# AL-70 data integrity review

Candidate AL70-R1, DISCOVERY, base `f63e3db6b2d4baa8e22452e4ec94eff566867721`, frozen input fingerprint `f5e45930942e860629a2c981f7e6c927fd0b6db9d4508a6d867be9b90cf78d91`. All 244 manifest entries matched before and after investigation. Requirements: frozen CONTRACT, confirmed inventory verdicts through A11 and PLAN supplied under `/tmp/al70-review-r1-7c7_i1yk/requirements/`; A12 remains outside approved change scope.

Requested/dispatch-acknowledged reviewer setting: Codex `gpt-6.1-sol`, high effort. Actual model identity, effort and usage were not independently observed from source records; no developer source stores were read. This report covers data preservation, owner refresh, failure isolation and recovery only.

## Finding

**DATA-SOL-1 — An inaccessible unrelated ticket blocks valid selected migration/backfill. Severity 5/10; introduced.** Locations: `backend/internal/adapters/backfill/engine.go:84` and `backend/internal/adapters/backfill/migration_runner.go:52`.

Both discovery callbacks now return every traversal error, aborting the operation before applying `--ticket-id`. With a readable `records/GOOD-01/agentic_metrics.json` needing approved A11 cleanup and an unrelated `records/ZZ-INACCESSIBLE` directory with mode 000, both commands exit 1 with `walking workspaces directory ... ZZ-INACCESSIBLE: permission denied`. GOOD-01 remains uncleared and unavailable to ordinary reads. Removing only the unrelated directory's permission restriction makes the identical commands successfully clean GOOD-01. This also blocks all reachable valid siblings during an unfiltered batch.

This violates A03/A04's per-ledger failure isolation and the engineering requirement to distinguish fatal root/configuration failure from individual workspace failure. The baseline skipped child traversal errors; the introduced unconditional error propagation expands one workspace's failure to the entire root. Preserve actionable per-workspace diagnostics while continuing reachable valid targets, with root/configuration failures remaining fatal.

Concrete proof: binary built from the frozen checkout with `cd backend && go build -o /tmp/al70-data-sol-engine ./cmd/agentic-lens`. Isolated fixtures and full failure/control results are `/tmp/al70-data-sol-isolation-kvvj6r0o/results.json`. Commands:

```text
/tmp/al70-data-sol-engine migrate --workspaces /tmp/al70-data-sol-isolation-kvvj6r0o/records --ticket-id GOOD-01
/tmp/al70-data-sol-engine backfill --source codex --workspaces /tmp/al70-data-sol-isolation-kvvj6r0o/records --ticket-id GOOD-01
```

Each failure preserved original ledger bytes; each readable-sibling control exited 0 and removed former metadata. Initial backfill attempts without its required source flag rejected before reaching discovery and are retained separately in the raw evidence, not counted as reproduction.

## Other hypotheses and evidence

- Complete local snapshots may erase unavailable owners or legitimate repeated rows: inspected generation baseline/history DTO preservation, exact owner replacement, source collision checks and backfill reconciliation. Complete snapshots replace exact owners; partial/failed sources retain prior owners. Disabled optional categories retain recorded evidence through nil-versus-empty handling. Targeted portable refresh, copied-ledger and unavailable-child CLI tests passed.
- Required policy/Git failure may publish cached or fabricated facts: generation exits on calculation failure; backfill prepares fresh Git facts before schema/publication work. The CLI fresh-Git-failure probe passed for schemas 5 and 6 with original ledger/summary bytes retained.
- A11 ordinary reads might repair records or lose pending state: ordinary decoding rejects former metadata; explicit preparation validates contradictions, duplicate members and category presence before cleanup. Focused cleanup/pending-state, runtime rejection, representation-only formatting and unavailable-publication sibling checks passed.
- Summary failure might expose an accepted mixed artifact set: inspected effective destination resolution, overlap/alias guards, publication locks, journals, rollback and retained legacy-report member authorization. Existing publication/format failure tests and configured-summary/append-admission CLI checks passed. Capture reconcile and relocation preserve snapshots, budgets, retry state and fencing in the focused checks. Private receipts remain operational evidence rather than authority over valid shared facts.
- Unknown cost and absent denominators: inspected unchanged cost aggregation with changed context ceiling paths; no guessed ceiling replaces absent policy, and retained raw history preserves cost presence. Focused cost/ceiling tests passed. Producer model counting was not changed or reinterpreted under pending A12.

## Verification and limits

Executed current-candidate diagnostic Go tests with `-count=1`: backfill, Codex, domain, migration and adapters filtered by `Metadata|Recovered|Retain|Portable|Complete|Cleanup|Budget|FormatLedger|Publication|Unavailable|Generation|Cost|Ceiling` (all five packages passed); six named CLI tests covering portable exact refresh, copied history, unavailable child summaries, configured summary admission, fresh Git failure and optional source reads (passed); seven capture tests covering restored/missing sources, session/byte limits, collision recovery, relocation fencing/resumption and ambiguous/missing OpenCode owners (passed).

Reused `/tmp/al70-cli-final3.log`'s full CLI pass only after comparing `/tmp/al70-cli-current-fingerprint.json`: all 346 recorded inputs match. The earlier source-package checkpoint has 36 input differences, so its green log does not establish a current-candidate full source-package pass; the focused current runs above supply this review's fresh evidence. Prior tester/standalone evidence is supplemental, not a replacement for the source investigation. Full integrated `make` remains the parent's delivery gate. No real developer sources, installation/activation, product edits or sibling reviewer reports were used.

**CHANGES_REQUESTED** for DATA-SOL-1. No other blocking finding established within this investigation.
