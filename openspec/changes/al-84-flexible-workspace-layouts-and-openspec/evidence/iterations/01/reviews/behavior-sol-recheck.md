# AL-84 behavior correction verification

Outcome: **APPROVED**, scoped to supplied behavior corrections and affected interactions. Run `al84_behavior_final`, iteration 01 final correction verification; configured/requested profile `gpt-6.1-sol` / high. Runtime model/effort metadata is unavailable to this reviewer. No new findings.

Reviewed baseline/HEAD `163181428ee9a5a599a2da1e91d3af5e618a9682` plus final dirty candidate manifest `9a78b592e2863c8ed07ef99f89f37a41d7052855d52749273c6fe4a9f40aab6f`. Independently hashed all 673 entries: zero drift. Selected design/specs/tasks, active review and supplied findings establish this verification boundary; earlier preliminary manifests are superseded. Product/tests remained read-only.

## Hypotheses and dispositions

- **BEH-01 (6), corrected.** `backend/cmd/agentic-lens/main.go:132` rejects noncanonical raw-generation filenames before publication; `workspace_locations.go:396` also enforces the canonical destination at the shared scope boundary. First and existing-ledger rejection controls preserve ledger/summary bytes and create no pending/state artifacts. The root command still publishes a computed export through `PublishWorkspaceArtifacts`; its positive control succeeds and protects canonical facts from overwrite. Discovery credit remains behavior Sol/high.
- **BEH-02 (6), corrected.** `workspace_locations.go:66` tracks enumeration incompleteness separately from rejected individual ledgers. Implicit resolution cannot establish uniqueness during an incomplete scan, and initial publication refuses incomplete absence checks. Explicit existing-home selection independently validates physical containment, depth, regular ledger and overlapping artifact policy, so budget-omitted homes remain available. Listing and unfiltered maintenance retain partial results with diagnostics; reaching selected depth remains normal. Checked both runtime/maintenance resolver variants and complete/invalid-sibling controls. Discovery credit remains behavior Luna/medium.
- **BEH-03 (6), corrected.** `backfill/discovery.go:15` routes targeted backfill/migration through `ResolveMaintenanceLocation`, preserving partial unfiltered batches. Metadata-only identity parsing admits historical schemas without introducing a runtime compatibility reader. Explicit homes work with or without a ticket filter; conflicting filters fail. `backfill/engine.go:169` forwards `Workspace` into nested migration. Complete valid absence returns a zero-scanned no-op before state/source initialization. Final owned reads recheck supplied ticket identity in `backfill/publication.go:17`. This finding originated during the earlier verification phase with behavior Sol/high; this report verifies its correction and claims no new discovery.
- **DATA-01 (6), affected interaction verified.** Both single-ledger and held-directory detail readers revalidate the final decoded identity in `repository.go:130` and `:168`. Real directory exchanges produced only the requested identity or explicit errors: 1,500 reads, including 180 successful reads, followed by successful stable recovery. Discovery credit remains data Sol/high.

## Verification and limits

Executed from `backend/`, all with `-count=1 -v`:

- `go test ./internal/adapters -run '^(TestWorkspaceIncompleteDiscoveryCannotProveUniqueOrAbsentIdentity|TestWorkspaceInspectedInvalidSiblingDoesNotConcealDuplicate|TestLoadTicketWorkspaceRejectsReplacementIdentity)$'`: passed, 5.007s.
- `go test ./internal/adapters/backfill -run '^(TestMaintenanceSelectionDoesNotInferUniquenessFromIncompleteDiscovery|TestBackfillCompleteDiscoveryMissingTicketIsNoOp)$'`: passed, 0.154s; includes legacy-schema, blank-ticket, nested-selection and unchanged-tree/private-state controls.
- `go test ./cmd/agentic-lens -run '^(TestCLIGenerateRejectsNoncanonicalLedgerBeforeWrites|TestCLIComputedOutputsPreserveFactLedgerAndContainment)$'`: passed, 4.283s.
- `git diff --check`: passed.

Reused inspected native logs/receipts in `build/al84-native-recovery/`: corrected lifecycle receipt (`56580636…`) retains both native owner objects through archive/external move, and final maintenance receipt uses current engine SHA-256 `bc37483f466739e3f579062e3f32bc4054cb6cb5f93402ef909ae8df4e8c5f41` (independently checked). Seven maintenance controls preserve the ledger tree exactly. Reused full backfill ordinary/race and affected CLI results: 8.406s / 29.055s / 38.724s. These are supplied executions, not newly repeated native lifecycle runs.

Parent owns integrated `make`, other focus verdicts, consolidation and publication. No broad audit, installed automatic-hook execution or native macOS execution was repeated. Durability and test-oracle findings remain with their assigned reviewers; this approval does not certify those boundaries or authorize merge.
