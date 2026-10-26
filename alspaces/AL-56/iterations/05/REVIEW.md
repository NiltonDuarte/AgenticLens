# AL-56 Iteration 05 Review: Session-Owned Schema v4

## Scope

Reviewed the approved iteration 05 plan and implementation diff against the ticket contract and `TECHNICAL_GUIDELINES.md`. The review covered the session-owned v4 model and DTO, migrations and incremental generation paths, Codex/OpenCode assembly, domain and UI consumers, privacy, tracked ledger conversions, and embedded frontend assets. The user explicitly approved removing legacy tool totals without reliable session attribution.

## Prior Findings Resolved

- Interim and canonical-v4 events are normalized through the persistence allowlist. The persisted event DTO has no `error_message` field; strict JSON decoding rejects that field. The migration runner handles sanitization before its up-to-date shortcut.
- The incremental loader and migration engine reject mixed session-owned and root-level legacy collections before writing output.
- Session construction combines duplicate same-name tool facts; persisted DTO validation rejects duplicate tool names, invalid counts, duplicate session IDs, malformed required collections, padded IDs, and dangling session parents.
- The migration runner validates canonical v4 ledgers before reporting them up to date and validates migrated DTOs before writing. Invalid input remains byte-for-byte unchanged.
- Removed unscoped tool rows are reported in CLI migration output and ledger notes. Migration status is `completed` when no steps are pending and `partial` only when pending steps remain.

## Data and Privacy Audit

- During review, independently parsed the initial 84-file migration sweep: all files were schema v4 with session-owned facts; 545 sessions/invocations were preserved, and 843 unscoped tool rows were removed across 66 ledgers with notes. After the user clarified that other ticket sources live on a different machine, all non-AL-56 workspace ledgers were restored. These sweep totals are historical review evidence, not the final branch contents.
- The final branch keeps only the AL-56 workspace ledger migrated; synthetic `testdata/` fixtures remain schema-v4-shaped for tests. No duplicate per-session tool names or persisted private event payloads were found in the audited data.
- No persisted session event contains prompt, reasoning, tool-input, tool-output, or error-message payloads. The 12 ineligible events removed from `testdata/workspaces/AL-2` follow the existing event allowlist.
- The JSON changes are limited to the approved session-owned relocation, schema-v4 `user_iterations`, documented migration notes, and the AL-2 privacy pruning.

## Verification Reviewed

- The latest tester report records `make` passing after all fixes: Go vet and race tests, frontend typecheck/lint, all 719 frontend tests, production frontend build, and multi-platform CLI builds. `git diff --check` passed.
- Focused migration-runner and DTO tests cover current-v4 fast-path validation, sanitization, strict `error_message` rejection, mixed-shape atomic rejection, duplicate tool names, invalid tool counts, missing collections, padded IDs, dangling parents, completed/partial migration status, and unchanged input on validation errors.
- Live Codex smoke generation was deferred to the final telemetry step per the ticket workflow; adapter extraction and session-refresh behavior are covered by fixtures and regression tests.
- Dependency installation reported six existing npm audit advisories (three moderate, one high, two critical); checks and builds pass.

## Verdict

`APPROVED`

The session-owned v4 contract is consistently implemented across migration, persistence, extraction, aggregation, and UI consumption. Migration loss is explicit, validated input is not silently discarded, and the final repository checks pass.

## Post-review workspace scope correction

After this review, the user requested that other ticket ledgers remain unchanged until backfill runs on the machine containing their source data. Those tracked files were restored; no workspace-wide backfill has been run. The code review and test verdict remain applicable to the implementation and canonical fixtures.
