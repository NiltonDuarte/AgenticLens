# AL-70 R2 test-effectiveness review

**Run:** `r2-tests-luna`; requested identity `codex / gpt-6-luna / medium` (runtime model metadata was not available). **Candidate:** AL70-R2, base `f63e3db6b2d4baa8e22452e4ec94eff566867721`, fingerprint `712d1b3db91b641ed2927f75081dd299a61aec84b3b907d69874c5d7614545c2`. Entry and exit manifest verification matched all 259 entries with no file mismatches.

## Verdict: APPROVED

No blocking test-effectiveness defect found in the assigned scope. Re-review verified the corrected oracles reach their claimed boundaries:

- TEST-01: `App.test.tsx` fetches a non-default literal/case-sensitive policy through `App`, then checks natural ordering; the supplied provider-removal mutant fails on that ordering, while the current focused control passes.
- TEST-02: the HTTP comparison expectation now preserves stored `bug`; alias filtering remains independently covered by domain alias cases and endpoint filter cases.
- TEST-03: live reads compare an unconditional before/after byte-hash map of two files in an owned `t.TempDir()` workspace. `TestZeroDiskPersistenceVerification` passed.
- TEST-04: Go builds SQLite, calls `LiveService` and the real HTTP handler, then byte-compares normalized JSON with the fixture. Frontend integration consumes that fixture via API validation, renders the tree, selects a node and renders the inspector; recorded-occurrence fixtures remain separate.

The real native worker publication test uses an actual fixture project, source, registry and worker subprocess for both absent and competing local budgets; it verifies publication and retained local config. It does not mock argv. The Linux permission probe drops credentials when run as root and fails if the denial control is ineffective; the targeted test passed. Structural config presence cases preserve explicit false/zero against dotted-key spoofing. Existing portable-history, source/Git ownership and publication failure tests remain distinct; no deletion is supported by counts or scores. A12 producer-count/model-label authority is not tested or inferred.

## Verification

Executed focused checks, all passing: `go test ./cmd/agentic-lens -run '^TestNativeWorkerPublicationKeepsExplicitBudgetOverrides$'`; `go test ./internal/adapters/live -run '^(TestNativeLiveHTTPWireMatchesFrontendFixture|TestLiveUnavailableDiffersFromAvailableEmptySource)$'`; `go test ./internal/adapters/http -run '^TestMultiFilterEndpoints$'`; `go test ./internal/adapters/backfill -run '^TestWorkspaceDiscoveryIsolatesUnreadableChildren$'`; `go test ./internal/adapters/live -run '^TestZeroDiskPersistenceVerification$'`; `go test ./internal/domain -run '^TestFilterLedgersByTicketTypes_AllAndAliases$'`; `go test ./internal/adapters/config -run '^TestPolicyPresenceFollowsStructuralYAMLKeys$'`; and frontend `npm test -- --run src/App.test.tsx src/views/LiveObservability.test.tsx` (215 unit prerequisites plus 28 selected integration tests passed).

Reused assigned proof logs: `/tmp/al70-r2-budget-final.log`, `/tmp/al70-r2-final-focused.log`, `/tmp/al70-r2-all-compile.log`, `/tmp/al70-data01-gscr5d4b/final-test.log`, `/tmp/al70-frontend-r2-{integration,typecheck,lint,build,app-mutant}.log`, and `/tmp/al70-r2-standalone-bud7xd_7/RESULT.json`. Full integrated `make` remains the parent’s gate; this scoped review does not certify the entire candidate.
