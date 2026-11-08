# AL-63 planning outcome

Refined the existing capability-audit specification so an operator can compare OpenCode v1, OpenCode v2 and Codex without mistaking declared support or fixture coverage for native runtime verification. All eight capability areas remain in scope. Evidence must identify versions, operations, source identities, results and limits, with historical evidence reused only within its stated applicability.

The specification now explicitly covers partial/unavailable hydration, complete-empty traces, unavailable same-harness owners, smaller authoritative snapshots and aggregate/idempotence behavior under the existing portable-ledger contract. FIX-18 retains live-support corrections and REV-7 retains read-contract investigation. No capability parity, new runtime support or changed backfill default is promised.

This is planning delivery under the supplied AL-63 identifier. No application, tests, runtime scripts, developer contract or new tickets changed. The audit itself remains Ready and unexecuted; AL-65 remains blocked. No native compatibility, successful backfill or original-ledger preservation probe is claimed. [Iteration plan](iterations/01/PLAN.md) records scope and verification boundaries; acceptance lives only in the [ticket specification](../plan/to-do/AL-63%20verify%20source%20capabilities%20across%20harnesses.md).

Verification is pending delivery checks. Initial checkout was clean at eb7d169c; fetched origin/main matched it. Author inspection confirmed source enums identify OpenCode/Codex, historical hydration dispatches per source, and current live SQLite queries reference v1 tables. These source observations motivate separate matrix columns; they do not certify runtime behavior.
