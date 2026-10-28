# FIX-8 Iteration 01 Plan

## Goal

Implement the contract in [../../CONTRACT.md](../../CONTRACT.md) by refreshing each Codex rollout tree once and replacing the complete matching tree session set atomically within the in-memory ledger recomputation.

## Scope and approach

1. In `backend/internal/adapters/backfill/engine.go`, change Codex recomputation to track which stored Codex sessions have already been covered by a discovered root tree. Resolve each distinct tree once; avoid repeating discovery for child sessions already represented by a successful tree refresh.
2. For a discovered tree, parse available rollout paths and combine transcripts as today, then group the combined facts into the full refreshed session slice.
3. Stage the replacement before mutating the working Codex session collection. Determine the stored entries represented by the refreshed tree using the discovered session identities/root relationship, remove those entries together, and insert all grouped refreshed sessions together. Do not assign only `freshCodex[0]`.
4. If discovery or any required parse/combine/grouping step fails, do not mutate the affected stored tree entries. Continue with independent stored Codex trees where possible.
5. Preserve all sessions outside the matched Codex tree and leave OpenCode and other-source slices untouched. Keep existing stable final sorting and recomputation result semantics, setting update/found flags only after a non-empty complete replacement is ready.

## Tests

Add focused coverage in `backend/internal/adapters/backfill/engine_test.go` (or the existing nearest backfill test file):

- **Root plus child:** Seed stored root and child sessions; make root discovery produce transcripts grouped into root and child. Assert both refreshed sessions are present exactly once, stale versions are absent, root discovery is not repeated for the child, and an unrelated source session remains identical.
- **Single session:** Seed one Codex session and return one refreshed session. Assert the result contains one refreshed copy and no duplicate ID.
- **Discovery failure:** Seed a Codex tree plus an unrelated Codex session and a non-Codex session; make discovery fail for the tree. Assert its original sessions remain unchanged, unrelated sessions survive, and result flags do not report a successful refresh for the failed tree.

Use the existing discovery injection seam if available; otherwise introduce the smallest adapter-level seam needed to make discovery outcomes deterministic in tests. Do not add mocks in domain logic.

## Verification

- Run focused backfill adapter tests covering the three cases.
- Run the repository's applicable Go checks (`go test` for the affected package, followed by the configured broader checks during implementation).
- Review the diff against the contract and confirm no `any`/`interface{}` or emoji was introduced.

## Completion boundary

This iteration covers only Codex backfill recomputation and its focused regression tests. It does not change source extraction formats, domain session grouping semantics, CLI flags, or unrelated recomputation behavior.
