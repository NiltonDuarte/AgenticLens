# Data review — GEN-16 iteration 02

- **Scope/revision:** Data and failure handling for the requested candidate based on `99fbcd7be0d0ed05639e4a0cb3be8744a3d5371e`; inspected the backend diff and untracked `backend/internal/domain/persona.go`, plus relevant unchanged extraction, persistence, live SQLite, aggregation, and ledger-read paths. Product/test fingerprint supplied by the coordinator: `785b903e14bb4af8923e054f2d8d5fb69f44210150b172d431d75aa5416d6f9a`.
- **Run identity:** requested profile `data_luna_6_medium` (`gpt-6-luna`, medium). That identity is acknowledged from the assignment; source-observed runtime metadata is unavailable here, so I do not assert it as measured runtime identity.
- **Outcome: CHANGES_REQUESTED** for the bounded-polling contract below. This finding concerns a live lookup behavior that the new persona fallback joins; it is not a defect in persona precedence itself.

## Findings

### D1 — Any database write invalidates every cached live trace

- **Classification:** Pre-existing invalidation behavior, newly relevant to the persona lookup requirement in this iteration.
- **Trigger/path:** An external writer changes any row in the OpenCode SQLite database. `SQLiteHook.SourceRevision` uses connection-level `PRAGMA data_version` (`backend/internal/adapters/live/sqlite.go:143`); `sessionTraceSnapshot` clears both `traceCache` and the new `tracePersonas` map whenever that single database-wide revision changes (`backend/internal/adapters/live/service.go:395-408`). `GetSessionTree` then asks for a trace for every active record (`service.go:324-344`), and each miss calls `FetchSessionTrace`, reading all messages and parts for that session (`sqlite.go:265+`). The persona evidence is also recomputed across all messages by `assistantPersonaEvidence` (`service.go:591-602`).
- **Consequence:** With one actively updating session, every poll can reread full histories for all active sessions, including unchanged sessions. Work scales with the sum of histories of active sessions per poll, rather than only changed histories. The new persona cache shares this global invalidation and does not bound that work per session independently.
- **Evidence/status:** Static source trace; no polling workload probe was run. This behavior predates the persona patch insofar as the global trace cache also used the database-wide revision, but the accepted iteration contract explicitly requires avoiding repeated full-history scans on every poll. A per-session change signal/revision or equivalent bounded incremental lookup is needed to satisfy that requirement. Do not treat the existing snapshot reuse test as disproving this: unchanged-database reuse is a different case from one-session-changes/other-sessions-unchanged.

## Hypotheses investigated

- **Metadata precedence and raw identity:** No defect found. Codex `agent_role` wins over path-derived persona; `agent_path` is now decoded and used for roots as well as children (`codex/rollout.go:107-140,349-362`). OpenCode v1/v2 collect assistant-agent candidates, reject disagreements through `PersonaEvidence`, then apply the resolved session role to invocations and events (`opencode/schema_v1_session_extraction.go:209-212,334-341,613-654`; v2 `:153-156,290-295,315-333`; `domain/persona.go:12-34`). User transport role is not consulted as persona evidence.
- **Unavailable/conflicting enrichment and legacy reads:** No data-loss defect found in inspected paths. `HydrateWorkspaceEvents` starts with committed ledger events as its fallback and returns that baseline when source sessions cannot be verified, message/part fetch fails, or the source is empty (`live/service.go:829-857,917-957`). Local persona projection uses explicit session metadata first, uses assistant evidence only when unambiguous, and preserves a known stored role when no usable evidence exists; historical placeholder values remain readable (`service.go:930-969`). Hydration is a response projection; inspected code does not persist it.
- **Identity, ownership, totals:** No persona-driven identity or ownership mutation found. Source session IDs and parent links are retained by extraction; aggregation continues to determine root status from `ParentID` while grouping blank roles as a role bucket (`domain/aggregations.go:179-198,423-460`). The role update is applied to invocation facts after candidate resolution, so token/cost/invocation counts are not dropped by blank roles.
- **Polling/cache reuse:** Unchanged source revisions reuse cached traces, which refutes an unconditional scan on every idle poll. It does not refute D1 when any database write advances the shared revision.

## Verification and limits

- Reused the supplied `iterations/02/TEST_REPORT.md`: it records passing affected adapter/domain checks, Go race tests across 12 packages, 783 frontend tests, typecheck and lint. I did not rerun suites, consistent with the assignment.
- No isolated polling probe was run; D1 follows from the revision and cache control flow. Source-session recovery was tested only with isolated fixtures per the supplied report, not against a live source database.
- This is a scoped data review, not a whole-change approval. Other review focuses and parent-owned integration/build gates remain outside this report.
