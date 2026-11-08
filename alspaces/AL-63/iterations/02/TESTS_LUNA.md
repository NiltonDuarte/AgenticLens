# Test evidence review: AL-63 iteration 02

**Scope/revision:** Test-oracle strength for the capability audit at `f4152225af06e2594477574c0e9f175a81218053`. Requested profile was Codex / gpt-6-luna / medium; observed runtime settings are unavailable, so this is acknowledged/requested attribution only. I did not inspect sibling reports. Application/test sources are frozen and were inspected only. No tests were run; parent reports `make` passed on unchanged app inputs at `0a8cfcee`.

## Evidence inventory

| Family / capability | Existing meaningful evidence and oracle | Limitation for AL-63 |
|---|---|---|
| Codex ingestion, facts, errors | `backend/internal/adapters/codex/rollout_test.go`: malformed-record continuation (204), native command/output/error mapping (157, 456–527), skill ownership/multiplicity and delegation identity (556–750), malformed skill/response incompleteness (750–785), nullable missing fields (808). `discovery_test.go` covers nonce matching and child discovery (16, 121). Assertions reach parsers over constructed rollout JSON/provenance fixtures, not native user-store sessions. | Supports parser behavior under fixtures; no runtime observation that a particular installed Codex store/version yields these records. Native root/child/session preservation and execution errors remain dependent on disposable audit probes. |
| Codex hydration/recovery | `backend/internal/adapters/live/sqlite_test.go`: explicit CODEX_HOME/source (959), multi-invocation (987), fail closed on missing root/child (1208,1247), source-less unavailability and preservation of lean facts/ledger bytes (1295,1308). | SQLite/rollout stores are test-built. No combined oracle for both missing and malformed root/child with previous tree plus skill/task snapshots retained, mixed-age diagnostics, exact-source owner preservation, growth refresh and unchanged repeat. |
| OpenCode schema v1 ingestion/facts/errors | `backend/internal/adapters/opencode/schema_v1_session_extraction_test.go`: extraction/skills/delegations (17,213), malformed optional inputs preserving other facts (323), tool execution errors and boundary payloads (621–869), strict recovery rejecting skipped records (1613), snapshot/completion (1720). | SQLite rows/JSON are constructed fixtures. These establish v1 adapter behavior against fixture schema; not a live/native OpenCode installation observation. |
| OpenCode schema v2 ingestion/facts/errors | `schema_v2_session_extraction_test.go`: normalized extraction and explicit child session (31,92), malformed/incomplete optional facts retaining other facts (163), malformed JSON/unknown variants fail with context (213), unknown child handling (378), active duration snapshot (437). | Constructed SQLite/JSON fixtures. No broad native version/schema parity claim follows. |
| OpenCode hydration/live | `live/sqlite_test.go`: malformed parts fallback (447), multi-fetch and child resolution (538), skill/task child attributes (614), malformed optional input (684), lookup failure survival (755), successful hydration/no writes (800), snapshot reuse/invalidation and legacy revision recovery (2286,2354). `http/server_test.go` 2586+ covers mocked hydrator response and fallback boundaries; `live_handlers_test.go` 805+ exercises snapshot ordering and REST recovery. | Strong local adapter/service/API boundaries but mostly synthetic SQLite, hooks, or mock hydrator; not native source observations. Does not establish all requested retention combinations across schema v1 and v2. |
| Live observation and execution errors | OpenCode v1 extraction tool error cases above; `backend/internal/adapters/live/sqlite_test.go` 1532,1744,1967,2067 cover live service token/anomaly, corrupt payloads, cross-turn errors and streak isolation. Codex rollout output-status tests cover parsed execution evidence. | Tests establish local parsing/derived diagnostics, not a running agent's complete live event stream or all harness execution-error reporting. No three-family parity oracle. |
| Shared domain/persistence | `backend/internal/domain/events_test.go` includes incomplete-source summary retention/clearing (792–821); `backend/internal/domain/generation_test.go` checks no fabricated root/child observation for partial generation (60–68); Codex retention parity test (codex/retention_test.go:16) compares compact summaries generated from Codex transcript events and OpenCode SQLite extraction. | Domain tests use in-memory values; parity compares normalized synthetic evidence and cannot prove source authenticity/equivalent native coverage. |

## Requested scenario coverage gaps

The existing suite has meaningful component-level coverage for missing Codex root/child, malformed Codex/OpenCode records, optional malformed skill/task payloads, unavailable OpenCode child lookup, snapshots, error parsing and some incomplete-source retention. It does not provide one durable end-to-end oracle spanning both OpenCode schema families for missing/malformed source records and children while preserving prior trees and skill/task snapshots, reporting mixed-age/incomplete diagnostics, then refreshing a growing source and remaining stable on repeat. AL-63 acceptance appropriately calls for disposable evidence probes; absence of a durable test does not establish absent runtime behavior.

Native versus constructed evidence must remain explicit: Codex rollout JSON, OpenCode SQLite schemas/rows, malformed records, and child lookup failures in these tests are constructed fixtures. No native source-session data was opened or altered in this review. Live adapter tests create private SQLite databases; HTTP tests include mocks. These prove code paths and oracle behavior under fixtures, not native harness capability or cross-harness parity.

## Commands and outcome

- `git status --short` — ticket plan modified and iteration 02 untracked before report; preserved.
- `git rev-parse HEAD` — `f4152225af06e2594477574c0e9f175a81218053`.
- `rg --files .agents | rg 'review-context|canonical'` — located `.agents/agent/references/review-context.md`; read shared context and canonical review skill.
- `rg -n '^func Test' ...` — inventoried named tests cited above; no execution implied.
- Test execution: none, per bounded scope and parent instruction not to rerun suites.

## Candidate findings and verdict

No confirmed product/test defect in this evidence-inventory assignment (no severity candidates). Material evidence limits above remain for the parent matrix: test evidence cannot itself verify native runtime capability, exact stored-session identity, source growth/repeat behavior, or full two-schema preservation/recovery combinations. These are audit probes in AL-63 acceptance, not proof of missing behavior.

**APPROVED (scoped):** existing tests provide substantial fixture-backed component evidence and are accurately usable when labeled as such. This approval does not certify native capability, parity, AL-63 completion, or unexecuted recovery scenarios.
