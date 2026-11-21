# FIX-20: Execution tasks

Use the [design](design.md) for scope and approval, and the [live and historical](specs/live-source-reporting/spec.md) and [backfill](specs/backfill-source-cancellation/spec.md) specs for observable acceptance.

## Implementation and focused verification

- [x] Planner confirms the scope, verifies native driver and file-read boundaries, and assigns implementation and test paths.
- [x] Propagate the context through live queries and cache wait, preserve revision validation, and surface an interrupted read or partial historical response. Cover Codex discovery and JSONL/gzip parsing. The live-source coder owned the live and Codex adapters. HTTP handlers already pass the request context.
- [x] Propagate cancellation through OpenCode schema discovery, v1/v2 extraction and exact snapshot cleanup. Pass the caller and workspace deadline through backfill and check cancellation before publication. Preserve ordinary partial refreshes. The backfill coder owned the OpenCode and backfill adapters and CLI entry point.
- [x] Verify already-cancelled requests, cancellation while waiting for the live cache, interruption after native reads have started, resource cleanup, a subsequent independent read, source revision changes, and useful partial historical evidence. The tester owns the regression files and [test evidence](evidence/TESTS.md).
- [x] Verify byte-for-byte preservation of the current workspace's ledger and summary after one owner completes and another is interrupted. Verify earlier committed workspaces, ordinary partial refresh, uncancelled v1/v2/Codex behavior, mixed-harness facts and Codex ambiguity rejection.

## Integration and delivery

- [x] Run affected native and public-entry-point checks. Record controlled conditions, revision, outcomes and limits without extrapolating latency.
- [x] Obtain independent review of context propagation, resource lifetime, cache and publication consistency. Resolve findings and reverify affected paths.
- [x] Integrate main and run the required build and verification.
- [x] Prepare delivery evidence and PR context. Verify the accepted ledger and record the final extraction limit.

## Workflow follow-up

Archive after verified delivery and reconcile the roadmap separately.

## Optional capture observation

- [ ] Gate the capture observation block with the shared opt-in and document it.
- [ ] Verify default omission, opt-in and replacement of the previously generated block.
- [ ] Review and verify the follow-up, then update the FIX-20 PR.
