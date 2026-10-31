# AL-55 iteration 01: source-backed backfill reconciliation

Status: Planning only; awaiting PO-17 reconciliation and explicit user scope approval. Contract: `alspaces/AL-55/CONTRACT.md`. Grounded commit: `bdef2d0ef98e37fe259921bc30ba2daefbfda717`. No coder, tester or product implementation has been dispatched.

## Authority and steering record

Source: initial manager/user task, 2026-10-31. Changed acceptance: treat the original AL-55 proposal as partially superseded; inspect actual v4 implementation and propose only justified missing behavior. Preserved guarantees: optional AL-51 facts, source-backed recovery, exact identity, unrelated harness preservation, AL-43 integration fields. Superseded decision: a second v4 bump, replacement v3-to-v4 step and client migration engine. Verification impact: existing schema migration tests are evidence owners, not a reason to recreate their coverage. This is the initial developer proposal, not an accepted implementation scope.

## Grounding completed

- Read configuration, planner persona, canonical technical guidelines, regression checklist and AL-55 specification. Workspace root resolves to `alspaces`.
- Traced raw OpenCode events through `opencode.ExtractSkillTaskFacts`, ledger DTO/domain optional arrays, `BackfillEngine.processSingleWorkspace`, persistence and frontend read helpers.
- Read migration chain/validator, v3-to-v4 step, source-backfill pending state and runtime eligibility validation. Identified existing atomic write/dry-run/standalone backup behavior.
- Inspected backfill source-selection, exact lookup, fallback, refresh branches and existing test bodies/coverage owners. Findings are static source evidence; no failure reproduction or runtime extraction is claimed.
- Inspected baseline CLI flag declarations and help implementation in `backend/cmd/agentic-lens/main.go`. `bin/agentic-lens` is absent in this checkout. Parent separately executed help successfully using the existing host binary `/home/niltonduarte/.t3/worktrees/AgenticLens/t3-19f29886/bin/agentic-lens-linux-amd64`, reporting version `v2.0.0-rc.93+4770c64-9-g7b8833d-dirty`, commit `7b8833d`, built `2026-10-31T13:44:29Z`. This is an older binary, not a fresh build of the grounded baseline. Its backfill help requires source selection, describes `--recompute-db` as OpenCode invocation/tool recomputation, describes `--force` as forcing migration/recomputation of completed ledgers, exposes opt-in heuristic fallbacks, and exposes no backup option. Its migrate help offers target version 4, backup and dry-run. Its generate help supports source Codex, explicit session ID, literal nonce, ticket type and title. This confirms the available host command interface only; it does not establish current-baseline runtime behavior.

## Proposed sequence after approval

1. Manager reconciles the five contract decisions with PO-17 and obtains explicit user approval. Record exact accepted choices and their source before freezing the contract. Do not infer approval from ticket status or unanswered messages.
2. Coder implements only approved backfill changes. Primary likely files: `backend/internal/adapters/backfill/engine.go`, with `report.go` or domain migration evaluation only where required to express selected coverage and diagnostics. Optional CLI backup work is excluded unless separately accepted. Avoid changes to schema version, schema v3-to-v4 migration and frontend.
3. Tester extends coherent existing backfill fixture suites for missing behaviors listed below. Any source-reader issue discovered while establishing coverage must be returned as a contract boundary concern; do not conceal failed extraction as empty success.
4. Reviewer audits the integrated diff against the frozen contract, regression checklist and existing AL-43 content fields. Record tested diff, actual commands and skipped boundaries in REVIEW.md; reuse valid unchanged-code results.
5. Parent integrates and runs the required full `make` gate when implementation is authorized and complete, then handles telemetry, reviewed task-owned staging, commits, PR delivery and manager handoff. No merge or auto-merge. Shared specification/index lifecycle remains serialized by manager.

## Proportional verification design

| Plausible regression | Existing owner and smallest meaningful extension |
| --- | --- |
| Explicit source recompute silently skips ordinary canonical v4 | Extend `TestBackfillEngine_ExplicitRecomputeDBDoesNotSkipCompletedLedger` in `backfill/engine_test.go` with migration absent and available-source fixture. Assert actual updated source facts, not only report counters. Include non-recompute unchanged behavior in the same table where coherent. |
| Partial OpenCode root refresh erases unavailable-root skills/task rows | Extend multi-root backfill fixture coverage with root plus child, second unavailable root, populated optional arrays and another harness. Assert exact preserved array values/multiplicity and parent/source/session identities, partial outcome, and no false source-backfill completion. |
| Complete recovery duplicates cumulative or repeated rows | In the same skill/task recovery fixture, make all required roots available, verify independently expected skill/task rows including repeated legitimate calls, then repeat unchanged and compare semantic facts. Grow source once and verify replacement rather than accumulation. |
| Codex one-tree success clears mixed-source/global incomplete recovery | Extend Codex multi-root/discovery/parse failure coverage in `backfill/engine_test.go`: one available tree, one unavailable stored owner, pending source-backfill. Assert success is local, failure remains explicit, pending requirement stays, and existing other-harness arrays/facts survive. |
| Missing stored child is silently considered complete | Extend the selected tree fixture to include a stored child absent from source discovery. Assert child remains and complete recovery is not claimed. The fixture must reach discovery/extraction rather than failing earlier DTO validation. |
| Safe write or dry-run guarantee regresses under incomplete recovery | Extend `TestBackfillEngine_DryRunPreservesFiles` with the approved recovery scenario if needed. Assert ledger and summary bytes unchanged in dry-run; malformed preparation/validation leaves original ledger intact. Reuse existing atomic/migration-runner tests instead of duplicating backup tests when backup is unchanged. |

Existing producer coverage to retain: `schema_v1_session_extraction_test.go` and `schema_v2_session_extraction_test.go` test skill status, malformed inputs, task metadata and child-link verification. DTO preservation is covered by `TestAgenticFactLedgerDTOPreservesSkillAndTaskFacts` in `adapters/schemas_test.go`; frontend optional read helpers are covered in `web/src/domain/events.test.ts`. Do not add synthetic JSON roundtrip/cast tests as a substitute for backfill production-path assertions.

Existing migration coverage to retain: `TestEngineMigrate`, `TestStepV3ToV4UserIterations`, ownership regression tests, runner batch-v4/idempotency/backup tests, `TestBackfillEngine_MigrationThenSourceBackfillMakesLedgerReady` and `TestBackfillEngine_SourceBackfillFailuresRemainPendingAndRuntimeIneligible`. No duplicate Go/TypeScript migration engine is planned.

All fixtures must use temporary configurable roots and isolated SQLite/rollout records, never live ticket workspaces or source databases. Sanitized or synthetic test inputs prove deterministic behavior, not recovery of real historical sessions. No arbitrary test count target.

## Verification status and remaining gate

Baseline verification is source inspection only. No application changes, no tests run, no full `make` pass, no historical ledger migrated. The baseline checkout has no built CLI; parent verified help on the separately identified older host binary above. After approval, install/build through canonical Make prerequisites, verify freshly built baseline command help, run focused affected backfill/domain checks, and run `make` after integration. No frontend change is currently proposed; preserve the shipped embedded asset tree through the canonical build.

Selected regression checklist boundaries: extraction/refresh/migration ownership and multiplicity; unavailable optional enrichment with persisted baseline; explicit partial recovery; configurable roots and isolated fixtures; failed preparation before persistence; preservation of unrelated harnesses; complete final evidence. Final telemetry belongs to the parent at iteration handoff, with actual Codex root/child identities and literal nonce, separate from manager and PO roots.
