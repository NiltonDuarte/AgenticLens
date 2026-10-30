# FIX-10 Iteration 04 Review - Round 2

Verdict: **APPROVED**

## Scope and remaining findings

Independently re-reviewed the final working diff against HEAD `be1d2e4`, including all four new regression test files and all 89 changed persisted ledgers. Applied `iterations/04/PLAN.md`, `CONTRACT.md`, `TECHNICAL_GUIDELINES.md`, and the reviewer role instructions. The approved iteration plan defines the invocation-identity removal; the contract's session preservation constraints remain applicable.

No remaining actionable findings. Both round-1 findings are resolved.

## Round-1 finding P1: Resolved

`backend/cmd/agentic-lens/main.go:1320` replaces the snapshot for each freshly extracted `SessionFact.ID`. `mergeInvocations` and its identity key are gone. An advancing completion timestamp no longer retains a superseded cumulative aggregate. Sessions absent from the fresh extraction remain intact, while every row inside an authoritative snapshot survives, including indistinguishable duplicates.

`merge_internal_test.go:10` covers growth, active-to-completed changes, model changes, nil/empty authoritative snapshots, initial extraction, and absent owners. The multiplicity test at line 62 verifies three identical rows, model/role/timestamp variants, another owner's preservation, repeated-refresh idempotency, and input immutability. `main_refresh_test.go:16` exercises real Codex rollouts and OpenCode SQLite across active, completed, and growing-source stages, comparing each refresh with fresh extraction and unchanged regeneration.

An additional independent CLI reproduction used five owners per source: a root, its child, another root, an unavailable owner containing duplicate rows, and an owner from the other harness. After source growth, both Codex and OpenCode produced exactly **one root row, input 240, output 75, two turns**. The child retained both model rows and its parent identity. The other three owners were preserved exactly. Refreshed root/child snapshots equaled fresh extraction, and subsequent unchanged generation was idempotent. Reproduction ledgers: `/tmp/fix10-r04-round2-multisession-4v7i0jyx/`; results: `/tmp/fix10-r04-round2-multisession-result.json`.

## Round-1 finding P2: Resolved

`backend/internal/adapters/live/sqlite_test.go:1040` retains `root-present` and `root-missing`; the test at line 1078 retains `root-thread`, `child-one`, and `child-two`. These nonempty owners pass the empty-ID guard. The root case reaches discovery of the absent root; the child case reaches the expected-thread completeness check after discovering the available tree. Each first asserts committed-ledger fallback with no partial events, then supplies the missing rollout and asserts successful local Codex hydration with the available event. Both tests passed independently in the uncached race run.

The unrelated message IDs `m1`, the formatted concurrent-writer IDs, and `msg_1` are restored. `backend/internal/adapters/http/live_handlers_test.go` and `backend/internal/domain/live_anomaly_test.go` match HEAD byte-for-byte, including all 46 restored message IDs in the latter. A Go AST comparison of identity fields across the changed files found no remaining unintended session/message ID removals; apparent differences were the renamed hydration test, removal of the invocation-derived session wrapper, or event copying delegated to the owner-aware grouping function.

## Identity, schema, migration, and data audit

- Go domain/DTO and TypeScript invocation types omit `id`. Source searches found no invocation-ID access, generation, suffix parsing, merge key, or fabricated owner fallback. Ownership and hierarchy use `SessionFact.ID` and `ParentID`.
- Runtime validation, strict full-ledger DTO decoding, and canonical v4 migration reject invocation `id`, including empty/null values. Domain and DTO serialization omit it. The isolated legacy flat migration structure retains explicit historical owner metadata; it does not generate or serialize canonical invocation identities.
- Grouping, direct/nested Codex combination, migration invariants, model/role/thread metrics, anomaly attribution, and frontend progression preserve session ownership and legitimate repeated rows.
- Structured comparison confirms that all other JSON values are unchanged in **89 ledgers / 604 rows**: 72 production ledgers / 536 rows, 12 parity ledgers / 54 rows, and five test ledgers / 14 rows. Every nested invocation ID was removed. Another 22 rows in complete Go raw JSON ledger literals and two rows in session-array fragments also omit the key.
- No introduced domain I/O, clocks, outer-layer dependencies, forbidden untyped values, artificial sleeps, or production Node dependency were found.

## Independent verification and test performance

- `cd backend && go test -race -json -count=1 ./...`: **PASS**, 12 packages, **2,361 tests/subtests**, zero failures/skips. Wall time: **36.073 seconds**. Log: `/tmp/fix10-r04-round2-go.json`.
- `cd web && npm test -- --reporter=json --outputFile=/tmp/fix10-r04-round2-web.json`: **PASS**, **737 tests / 33 files**, zero failures/pending/todo tests. Wall time: **30.414 seconds**.
- `make check`: **PASS**, including Go vet, TypeScript typecheck, ESLint, and whitespace validation.
- Current CLI built into `/tmp` and the independent multi-session growth/fresh-extraction/idempotency checks above: **PASS** for both sources.

The slowest observed test was the existing backfill multi-machine lifecycle test at 5.45 seconds; the new CLI refresh regression took 4.76 seconds across both sources. Its disk/SQLite and subprocess work verifies the actual extraction boundary. No new artificial delay or test-performance blocker was identified. Race-run scheduling and integration timings are not pure-domain function latency measurements; this iteration has no specified total-suite performance target.

Only this review artifact was written in the repository. No application/test code, plan, contract, telemetry, summary, or PR artifact was modified. No commits, telemetry ingestion, or PR operations were performed.

Final status: **APPROVED**
