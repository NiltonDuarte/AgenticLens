# AL-84 iteration 01 final behavior correction verification

**Outcome: APPROVED**, limited to supplied behavior corrections and consequential behavior interactions. No new findings. This does not approve other review focuses, the integrated `make`, or delivery.

**Candidate:** baseline/HEAD `163181428ee9a5a599a2da1e91d3af5e618a9682` plus the dirty candidate identified by `build/al84-review/FINAL_CORRECTED_MANIFEST.json`; 673/673 file hashes match and the pretty-serialized content-map SHA-256 is `9a78b592e2863c8ed07ef99f89f37a41d7052855d52749273c6fe4a9f40aab6f`. Requested run: AL84 iteration01 final behavior correction verification, `gpt-6-luna` / medium (runtime metadata unavailable). Product and test files remained read-only.

## Hypotheses and dispositions

- **BEH-01 (severity 6), corrected; no new discovery credit.** `cmd/agentic-lens/main.go` rejects noncanonical `--json-output` before publication, and `ValidateLedgerDestinationScope` independently enforces `agentic_metrics.json`. Focused CLI tests pass for first and existing-ledger rejection and the computed root export positive control, which preserves the canonical ledger. This keeps intentional root export behavior distinct from raw fact generation.
- **BEH-02 (severity 6), corrected; no new discovery credit.** Discovery tracks enumeration incompleteness separately from rejected ledger candidates. `ResolveTicketLocation` cannot claim implicit uniqueness on an incomplete scan, and `ValidateInitialLedgerDestination` refuses to infer absence. Explicit selection checks the selected existing home against physical containment, configured depth, regular-ledger and overlapping-artifact policy, even if budget-limited enumeration did not visit it. Partial listing and unfiltered maintenance retain discovered results with diagnostics; selected-depth completion alone is not treated as truncation.
- **BEH-03 (severity 6), corrected; verification of the earlier verification-phase finding only.** Targeted backfill and migration use `ResolveMaintenanceLocation`; explicit homes work with or without a ticket filter, while conflicting identity is rejected. Unfiltered batches retain partial results and diagnostics. Historical metadata-only identity selection stays separate from runtime fact decoding; nested migration forwards `Workspace`. A complete, valid missing-ticket scan returns a no-op, whereas incomplete or ambiguous discovery remains an error. Focused tests cover backfill, migration, nested migration and these selection boundaries.
- **DATA-01 (severity 6), consequential interaction verified; no new discovery credit.** Final single-ledger and held-directory reads revalidate decoded identity after resolution. The Linux replacement test passed 1,500 replacement reads (147 successful reads); results retained the requested identity or returned an explicit error.

## Verification and limits

Executed from `backend/`, all with `-count=1 -v`:

- `go test ./internal/adapters -run '^(TestWorkspaceIncompleteDiscoveryCannotProveUniqueOrAbsentIdentity|TestWorkspaceInspectedInvalidSiblingDoesNotConcealDuplicate|TestLoadTicketWorkspaceRejectsReplacementIdentity)$'`: passed.
- `go test ./internal/adapters/backfill -run '^(TestMaintenanceSelectionDoesNotInferUniquenessFromIncompleteDiscovery|TestBackfillCompleteDiscoveryMissingTicketIsNoOp)$'`: passed, including explicit unvisited homes with and without a filter and partial unfiltered batches.
- `go test ./cmd/agentic-lens -run '^(TestCLIGenerateRejectsNoncanonicalLedgerBeforeWrites|TestCLIComputedOutputsPreserveFactLedgerAndContainment)$'`: passed.

Reused supplied native evidence, not newly rerun: corrected lifecycle receipt preserves exact native owner objects through archive/external move and late refresh; `MAINTENANCE_AFTER_RECEIPT.json` reports seven maintenance controls with unchanged ledger-tree bytes. Reused reported full backfill ordinary/race (8.406s / 29.055s) and affected CLI (38.724s) results. Source stores, synthetic OpenCode usage counters, and native macOS/automatic-hook execution remain outside this verification. The parent owns integrated `make`, cross-focus consolidation and delivery.
