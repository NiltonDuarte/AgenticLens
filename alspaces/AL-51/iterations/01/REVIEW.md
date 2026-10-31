# AL-51 Iteration 01 Review

## Scope and checks

Reviewed the iteration diff against `CONTRACT.md`, `PLAN.md`, and `TECHNICAL_GUIDELINES.md`. The implementation is additive, leaves `CurrentSchemaVersion` at 4, keeps SQLite/JSON/session correlation in adapters, and propagates the new fields through DTO persistence, live trace retrieval, and `trace_update` SSE events. TypeScript uses optional typed fields and defensive accessors; no UI scope was added. No new `any`/`interface{}` or emoji use was found in the reviewed changes.

The tester reports `make test`, `make check`, and `git diff --check` passed, including Go race tests, frontend Vitest, typecheck, and lint.

## Findings

1. **OpenCode v1 can mark failed skill invocations successful.** In `backend/internal/adapters/opencode/schema_v1_session_extraction.go`, tool failure detection compares the original status with the case-sensitive literal `"error"` and does not recognize `"failed"`, while schema v2 explicitly accepts both `error` and `failed`. `ExtractSkillTaskFacts` derives `SkillFact.Success` from the normalized event status, so a v1 skill part with `state.status: "failed"` (or differently cased `"ERROR"`) and no matching output signature is emitted as successful. Normalize and handle both failure statuses, and add a v1 regression test.

2. **The new extractor's malformed and incomplete optional-input behavior is not directly covered.** The contract and plan require missing/malformed skill/task input coverage. The added extraction tests exercise populated skill/task inputs, but do not assert that invalid JSON, a missing skill name, or missing task fields are safely skipped/handled without affecting unrelated extracted facts. Add focused cases for these inputs.

## Verdict

CHANGES_REQUESTED
