# AL-51 Work Summary

## Domain

- Added optional `SkillFact` and `TaskLinkFact` collections to the Go and TypeScript ledger contracts, plus optional live part attributes. Kept the current schema version at 4 and added no migration.
- Preserved legacy ledger decoding and serialization through the adapter DTO and repository paths.

## Adapters

- Extracted skill invocations and task delegations from OpenCode v1 and v2 session events, including invocation outcomes, ordering, and explicit child-session association where recorded.
- Propagated historical facts into generated and backfilled ledgers.
- Added skill/task attributes to live trace parts and `trace_update` SSE events. Child-session resolution is best-effort, so unavailable optional enrichment does not fail trace retrieval.
- Malformed or incomplete optional tool inputs are ignored without interrupting ordinary extraction or traces.

## Presentation contracts

- Added optional TypeScript ledger and live trace fields. Legacy payloads continue to be accepted; no skill/task UI was introduced.

## Verification

- Iteration 01 initially identified a v1 status-normalization defect; corrected in iteration 02.
- Iteration 02 identified optional child lookup failure propagation; corrected in iteration 03.
- `make test` passed: Go race tests and 822 frontend tests.
- `make check` passed: Go vet, TypeScript typecheck, and ESLint.
- Focused Go package tests and `git diff --check` passed.
- Final reviewer verdict: APPROVED (`iterations/03/REVIEW.md`).
