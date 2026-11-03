# AL-76 work summary

Initial tree clean at `122e983f8c365a2aa248b1b6d492252555e922ae`; branch `feat/AL-76-compact-ledger-format`. User pre-approved the supplied specification. [Contract](CONTRACT.md), [iteration plan and steering](iterations/01/PLAN.md).

## Delivered implementation

Adapters: lexical JSON formatting preserves original scalar spelling, member/array order, unknown fields and presence while compacting only recognized ledger records. All persisted writers use this policy. Source-independent filesystem formatting validates selection and inputs under canonical publication ownership, preserves permissions and supported aliases, rejects unsafe paths/pending journals, and publishes atomic replacements only when bytes differ.

CLI: `format --all`, `--ticket-id`, `--ticket-ids`, configured/explicit roots, dry-run and check, deterministic reports, help and operator reference. Domain calculations, frontend and release pin unchanged. No historical ticket ledgers were reformatted.

Concurrency correction: backfill/migration acquire ownership before baseline reads through publication; computed exports own canonical output without nested scanner ownership. Product revision `c32c8937d8897179a6772c6aa75300eafe94351f`; tests extended in `0dedb33`.

## Review and corrections

Behavior, data integrity and architecture reviewers, each gpt-6.1-sol/high (inherited profile; child runtime metadata unavailable), independently discovered severity 8 PUB-01: formatting could overwrite a newer migration/backfill because manual publishers bypassed ownership. Behavior and architecture independently reproduced schema downgrade; data independently traced the interleaving. Coder gpt-6.1-sol/high corrected ownership; all three Sol focuses approved the correction. Luna gpt-6-luna/medium behavior, data and architecture discovery reviews approved corrected product; dispatch settings acknowledged, independent source metadata unavailable.

Test-effectiveness reviewer gpt-6.1-sol/high discovered severity 4 missing deterministic late-publication failure coverage and severity 3 missing computed anomaly layout oracle. Tester gpt-6.1-sol/high added both regressions; Sol re-review and Luna independent test-effectiveness discovery both approved at `0dedb33`. Parent also caught slash-containing unknown-key path collision during implementation; coder escaped path segments and tester added regression.

Final consolidated delivery approval remains INCOMPLETE pending predecessor integration and required full make. Working evidence remains retained; do not compact/delete reports before that gate.

## Verification

Canonical frontend assets prepared successfully (frontend source unchanged). Focused formatter/CLI, fact/computed writer, backfill/migration and deterministic ownership race suites passed. [Testing evidence](iterations/01/TESTING.md) records commands and corrections. Latest test-only race command: `cd backend && go test -race ./internal/adapters -run 'TestFormatLedger|TestSaveMetricsLedger' -count=1` passed. Corrected publisher command: `cd backend && go test -race ./internal/adapters ./internal/adapters/backfill -run 'TestSaveMetricsLedger|TestManualWriters|TestFormatLedger|TestMigrationRunner|TestBackfillEngine' -count=1` passed. Diff whitespace checks passed.

Earlier full CLI run failed capture native-version and worker shutdown tests; isolated failed installer cases passed on rerun. MGT-2 reports matching failures on unchanged baseline and AL-60 owns the approved shared capture repair. Do not weaken tests or extend AL-76 into capture repair.

## Pending coordinated delivery

User requires waiting for MGT-2 notification that repaired predecessors AL-60 and AL-61 are ready/merged before one integration pass and expensive full make. Preserve their index edits. MGT-2 owns merging; AL-76 will publish and explicitly send PR URL, reviewed head and verification state to the manager thread. Telemetry deliberately remains deferred to final delivery, including later continuation. Lifecycle completion, telemetry, push and PR remain pending; no PR exists yet.

Built standalone source binary `/tmp/agentic-lens-AL76` successfully. Native `format --help` and `generate --help` inspected. Native commands do not accept wrapper-only `--al-offline/--al-project`; final telemetry must use supported native `generate --source codex` arguments from the project root or the proper wrapper if available. No telemetry generated early. Disposable CLI fixture verified dry-run exit 0, dirty check exit 1, successful write exit 0 and clean check exit 0 with exact large integer/exponent preservation.

All eight focus/model discovery cells are complete, with affected Sol correction re-reviews approved. Behavior/data/architecture product approvals at `c32c893` carry forward to `0dedb33`: successor changes only tests and ticket documentation, no product dependencies changed. Test-effectiveness approves final test revision directly. No known unresolved product/review finding; consolidated delivery gate remains INCOMPLETE solely for coordinated upstream integration/full make. Private active evidence: `/tmp/AL76-*-sol.md`, `*-sol-rereview.md`, `*-luna.md`, `/tmp/AL76-review/{REVIEW_RUNS,FINDINGS}.json`. Keep these until final approval and compact retention.
