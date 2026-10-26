# AL-56 Iteration 04 Plan: Exact Iteration and Tool Aggregation

## Objective

Resolve the two `CHANGES_REQUESTED` findings from `iterations/04/REVIEW.md`: preserve exact zero for SDLC iterations and make tool totals additive across distinct sessions while replacing a session's facts on repeat generation. Requiring session identity for every schema-v4 tool fact is an intentional contract tightening; backward compatibility for unscoped tool rows is dropped. Re-extraction can recover those facts only while their source logs remain available. Verify the implementation against the active Codex harness and preserve OpenCode behavior.

## Implementation Scope

- In `runGenerate`, always assign `codeFacts.SDLCIterations` to the sum of `UserIterations` across root invocations, including when that sum is zero. Remove the stale cached-code preservation and forced minimum-one fallback for this derived field. Keep a clear migration diagnostic for ledgers before the iteration-attribution schema; do not use that diagnostic as a reason to invent a nonzero total.
- Make `ToolFact.session_id` mandatory and nonempty in the v4 domain model, DTO, and TypeScript type. Validate it at the ledger boundary and reject missing or whitespace-only values. Have OpenCode and Codex extraction emit tool counts per session and tool name. Merge by `(session_id, tool_name)`: replace matching local session snapshots with their fresh authoritative values, retain cached sessions absent from local extraction, and add newly observed sessions. Derive displayed totals by summing the merged session facts.
- Treat the serialized `tools` collection as atomic session-scoped facts, not dashboard rows. Update domain aggregation and every presentation consumer to group by tool name and sum calls/errors before rendering per-tool summaries, anomaly calculations, or tool counts. In particular, a displayed number of tools must count unique names. Preserve the current one-row-per-tool user-facing view.
- **Schema compatibility decision:** keep the ledger at schema v4, but require session provenance for every tool row. Do not support unscoped legacy tool rows in v4. Extend the existing v3-to-v4 migration to fail clearly when a source ledger has any nonempty tool facts without `session_id`; tell the user to regenerate/re-extract the ledger from source logs when available. Do not invent session IDs, silently discard tool counts, or partially rewrite the ledger. Ledgers with no tool rows can migrate normally. Apply the same validation to already-v4 ledgers: if any tool row lacks a nonempty session ID, generation, load/backfill, and summary update must fail with actionable re-extraction guidance. Backfill/migration must be idempotent for valid ledgers, and failed migration must leave the source ledger unchanged. Document that recovery is unavailable when source logs are gone, and update setup/research/CLI documentation to explain this contract tightening and recovery path.
- Keep event merge behavior and invocation merging unchanged unless implementation reveals a direct dependency needed for these two fixes. Do not expand into general schema or unrelated workspace/setup changes.
- Keep all fixtures and tests repository-independent, with temporary project directories and stable synthetic session IDs.

## Acceptance Tests

1. Replace the current assertion that schema-v4 zero iterations default to one. Assert zero root iteration facts produce `code.sdlc_iterations == 0`, even if cached `code.sdlc_iterations` is positive.
2. Add a mixed-root test where cached facts contain positive `user_iterations` but freshly authoritative matching root facts report zero. Assert the output uses the exact merged root invocation sum, zero, and does not retain stale cached totals.
3. Keep/add a positive sum test across two distinct root sessions and a child invocation. Assert only root invocation counters are summed and zero remains a valid result.
4. Test tool accumulation with Device A and Device B facts for the same tool name but different session IDs. Assert calls and errors sum across sessions.
5. Test tool refresh idempotency: re-run with updated facts for an existing session ID and assert that session is replaced, not added; repeat the same run and assert totals remain unchanged.
6. Test domain/DTO validation rejects a tool row with missing, null, empty, or whitespace-only `session_id`, while a ledger with an empty tools collection remains valid. Test schema-v4 serialization/deserialization preserves required session provenance.
7. Test OpenCode and Codex extractors emit nonempty session-scoped tool facts, including root and child sessions, and test aggregation keys facts by both session ID and tool name.
8. Add domain and UI/presentation tests showing multiple session rows for the same tool aggregate to one displayed tool row with summed calls/errors; verify tool count is the number of unique names and anomaly/brittle-tool summaries do not duplicate the same tool.
9. Test v3-to-v4 migration/backfill fails clearly and atomically when nonempty legacy tool aggregates lack session IDs, with actionable source-log re-extraction guidance and no lost/rewritten counts. Test v3-to-v4 succeeds when there are no tool rows and remains idempotent. Test already-v4 unscoped rows fail validation/backfill with the same guidance. Test valid v4 session-scoped tool rows round-trip and backfill idempotently. Migration diagnostics must not coerce generated zero totals.

## Codex and Regression Verification

1. Run focused tests for Codex rollout parsing/discovery, OpenCode extraction, tool fact DTOs/merging, invocation aggregation, and schema migration.
2. Run the backend race suite, frontend tests/checks, and `make`; record any unavailable targets explicitly.
3. Build the CLI and run `generate --source codex` against the active Codex harness and a real local rollout. Verify discovery resolves the current root thread, output remains schema v4, root iterations equal the sum of root invocation facts, and newly extracted tool facts are session-scoped.
4. Repeat generation against a stable completed rollout. Compare invocation, iteration, and tool facts to confirm no duplicates, stale iteration fallback, or tool-count inflation. An actively changing rollout is a smoke check only, not idempotency evidence.
5. Review generated output for transcript, prompt, command/tool payload, or private-path leakage. Keep real-session output only in the tracked AL-56 workspace.
6. Run OpenCode generation without `--source` and verify the default remains OpenCode and existing extraction behavior passes.

## Completion Criteria

- The two P1 findings in `iterations/04/REVIEW.md` are resolved by focused regression tests.
- Iteration aggregation equals the exact root `UserIterations` sum, including zero; no stale or default fallback remains.
- Tool totals sum distinct sessions and refresh matching sessions idempotently; unscoped nonempty tool rows are rejected with re-extraction guidance rather than preserved or dropped.
- Ledger schema remains v4 and every valid v4 tool row has nonempty session provenance. Migration/backfill rejects old unscoped tool aggregates without changing them and gives an actionable source-log re-extraction path; ledgers with no tool facts migrate normally.
- Existing per-tool dashboard, summary, count, and anomaly behavior remains aggregated by unique tool name rather than exposing per-session rows as duplicate tools.
- Codex real-rollout checks pass; repeated-run claims use a stable completed rollout.
- OpenCode regression and required repository checks pass.
- Findings and verdict are recorded in `iterations/04/REVIEW.md` before updating the ticket summary/PR deliverables.

## Deliverables

- Updated implementation and regression tests for the two findings.
- Revised findings and verdict in `iterations/04/REVIEW.md` after re-review.
- Updated `SUMMARY.md` and `PULL_REQUEST.md` only after the fixes and verification are complete.
- Telemetry refresh as the final workflow step for this iteration, per the repository instructions.
