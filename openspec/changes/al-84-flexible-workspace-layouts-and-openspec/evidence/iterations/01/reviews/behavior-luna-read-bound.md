# AL-84 behavior read-binding delta: APPROVED

Run: `AL84 narrow same-ID readbinding behavior delta verification`, iteration 01; requested reviewer profile `gpt-6-luna` / medium. Baseline/HEAD: `163181428ee9a5a599a2da1e91d3af5e618a9682`. Candidate: baseline plus `build/al84-review/READ_BOUND_MANIFEST.json`, content hash `41a2ee0b246538aba898442b36c22015511557d4fca7e231504f6f9a3d068720` (674 listed paths; every listed checkout file hash matched). Scope was the unstaged read-binding correction in `workspace_locations.go`, `repository.go`, `publication_set.go`, and its two focused tests. No product or test files were changed during this review.

## Finding

No blocking behavior finding in the reviewed delta. The earlier correlated same-ID source gap is closed for the covered read paths. During resolution, runtime reads now record the selected directory's physical identity and retain a separate descriptor to that directory. Before a later ledger or artifact read, `openResolvedWorkspaceDirectory` opens the selected home and checks that it is still the resolved physical directory. A replacement between resolution and final admission returns an explicit unavailable/retry error. Once admitted, ledger and configured artifacts are read through the same held directory. The independent descriptor does not hold the publication flock: `TestResolvedReadReferenceAllowsWriterAdmission` confirms writer acquisition succeeds while the read reference remains open. References and read leases are closed on success and error paths.

The read does not attach identity metadata to the ledger or require private operational state. Resolution still validates supported ledger contents; `readDirectoryLedger` still applies the adjacent pending-publication indicator check before and after reading. This preserves supported portable facts on machines without private state while treating an interrupted set as incomplete. Explicit same-ID selection remains required by the workspace-locations spec. If B already occupies the selected path before a fresh resolution, resolving and reading B is valid; if the path changes after A was resolved, the operation retries instead of returning B's same-ID facts. Detail reads preserve ledger/artifact pairing from one admitted directory.

## Hypotheses and evidence

- **Resolved A replaced by same-ID B before final open:** source guard compares the final held directory with the identity observed during resolution; the deterministic test asserts retry/unavailable, with no ledger or artifact payload returned. A disposable copy with only that physical-directory guard disabled fails this test (expected mutant failure in `SAME_ID_TEST_RECEIPT.json`).
- **Concurrent same-ID detail reads mix copies:** real Linux directory exchange test accepts either copy when resolution can select either, but requires ledger title and artifact marker to agree. The receipt records 56 successful reads of 400 attempts plus stable controls; no title-count-only inference is used.
- **Resolution reference blocks writers or creates local state:** source uses a lockless independent directory descriptor; the writer-admission test passes, and the deterministic resolution test checks private home and local config remain absent.
- **Existing behavior approvals BEH-01/02/03:** carried forward from `behavior-luna-recheck.md` because this delta changes only runtime read binding and does not modify canonical output admission, incomplete discovery/selection policy, or maintenance selection. Their earlier evidence and limits remain as recorded there; this report does not expand their coverage.

## Verification and limits

Executed from `backend/` against the candidate:

- `go test ./internal/adapters -run '^(TestResolvedWorkspaceReadsRejectSameIDDirectoryReplacement|TestResolvedReadReferenceAllowsWriterAdmission)$' -count=1 -v` — passed.
- `go test -race ./internal/adapters -run '^(TestResolvedWorkspaceReadsRejectSameIDDirectoryReplacement|TestResolvedReadReferenceAllowsWriterAdmission)$' -count=1 -v` — passed.

Also reviewed the corresponding ordinary/race adapters, capture, and HTTP receipts plus the focused CLI-selection receipt in `build/al84-correction-tests/SAME_ID_TEST_RECEIPT.json`. The test and source manifests in that receipt match the supplied 362-file source fingerprint. This is a focused Linux read-binding review, not integrated `make`, native macOS/Windows race execution, or approval of other review focuses. Parent-owned integrated verification remains pending.
