# AL-84 iteration 01 behavior review

**Outcome: CHANGES_REQUESTED**
**Scope:** CLI/API/UI/capture/maintenance ledger-home resolution, moves and publication recovery.
**Candidate:** baseline `163181428ee9a5a599a2da1e91d3af5e618a9682` plus frozen 669-file manifest SHA-256 `13a044baddc800f298f54fba03ee41d4c09092d41c89228afde1a79520a821a6`. All 669 manifest file hashes matched.
**Review identity:** `gpt-6-luna`, medium effort.

## Finding

**Severity 8 — incomplete discovery can silently select or create a duplicate ticket ledger.** In [`workspace_locations.go`](../../../../../../backend/internal/adapters/workspace_locations.go), `resolveTicketLocation` accepts `len(matches) == 1` even when `DiscoverWorkspaceLocations` also returned a directory-budget exhaustion diagnostic (lines 302, 355–357). For a configured root with `max_depth: 1` and `max_directories: 1`, place one ticket ledger at the root and a second same-ID ledger in a child directory. The root is observed, the child is not visited, and implicit CLI generation, API detail, capture publication, or maintenance resolution can treat the observed ledger as unique. The same incomplete scan is ignored by `ValidateInitialLedgerDestination` (lines 408–426), allowing first publication at another explicit destination; `runGenerate` only prints discovery issues as warnings (main.go lines 158–166). This can refresh the wrong copy or create a duplicate despite the contract requiring explicit selection when IDs have multiple ledgers. Require complete discovery for uniqueness and first-publication duplicate checks, or otherwise make incomplete scope an explicit selection/error condition.

## Hypotheses investigated

- **Folder names/roadmap metadata establish identity:** refuted. Discovery only considers regular `agentic_metrics.json`; runtime resolution validates and uses recorded `ticket_id`.
- **Consumers choose different homes or first-root precedence:** mostly refuted. Repository/CLI/API detail and capture use the shared resolver; UI consumes API results. Distinct observed duplicates return ambiguous until a physical location is selected. Backfill and formatting share bounded discovery and check recorded identity, with maintenance intentionally reading identity before ownership.
- **Move/archive/pending recovery recreates the former parent or requires private state for completed reads:** no defect found in inspected paths. Publication uses held directory roots for owned writes; reads check the adjacent pending marker and return unavailable without recovering or mutating it. The supplied native receipt records archived/external move and late refresh with the former home absent.

## Verification and limits

- Passed: `go test ./internal/adapters -run 'TestWorkspaceDiscovery(DeduplicatesPhysicalScopesAndReadsAbsoluteRoots|RejectsConflictingArtifactPolicyWithoutRootPrecedence)$' -count=1`.
- Passed: `go test ./internal/adapters/http -run '^TestLedgerLocationsRemainDistinctAcrossMoveAndUnavailableRecoveryState$' -count=1`.
- Passed: `git diff --check`.
- Inspected supplied `build/al84-native-recovery/NATIVE_RECEIPT.json` and iteration test evidence; parent owns integrated `make` and cross-focus consolidation. Native receipt is not installed automatic-hook runtime coverage. Native macOS execution and actual reproduction of the budget-exhaustion duplicate scenario remain unverified.

This is a behavior-only verdict, not whole-repository approval.
