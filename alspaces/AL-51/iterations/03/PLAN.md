# AL-51 Iteration 03 Plan

## Goal

Preserve live trace availability when optional task child-session correlation cannot be resolved.

## Review-driven work

1. **Make child-session correlation best-effort**
   - In `backend/internal/adapters/live/service.go`, adjust `LiveService.GetSessionTrace` so an error from `SQLiteHook.ResolveChildSessionIDs` does not fail trace retrieval.
   - Continue returning the successfully fetched and parsed messages and parts. Leave `child_session_id` absent for task parts when child lookup is unavailable; retain IDs when lookup succeeds.
   - Keep errors from required trace reads and parsing on their existing error paths. Do not hide unrelated failures or change the SQLite child lookup contract.

2. **Prove fallback behavior**
   - Add a focused test in `backend/internal/adapters/live/sqlite_test.go` that makes optional child lookup fail after trace records are available.
   - Assert `GetSessionTrace` returns no error, preserves the task part and its other parsed attributes, and omits `child_session_id`.
   - Keep the existing successful parent-child correlation assertion to guard the normal path.

## Acceptance checks

- A child-session lookup failure does not make `GetSessionTrace` fail or discard otherwise available trace data.
- Successful lookups continue to populate only explicitly linked child session IDs.
- Required trace query errors retain their existing behavior.
- Current schema version remains 4; no migration, domain contract, or unrelated stream behavior changes.
- Focused live adapter tests and the relevant project checks pass; record outcomes in the iteration review and ticket summary.

## Boundaries

- This is a narrow resilience correction within the approved AL-51 contract.
- Do not infer child IDs, change persistence, add UI presentation, or expand the live trace data shape.
