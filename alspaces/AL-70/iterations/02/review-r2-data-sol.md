# AL-70 data re-review

**APPROVED within the assigned data scope.** No new blocking finding. This verifies the supplied corrections; it does not grant discovery credit for CFG-01, CAP-02 or CTX-01 or certify the parent-owned integrated delivery gate.

Run `r2-data-sol`, RE_REVIEW, candidate AL70-R2; base/HEAD `f63e3db6b2d4baa8e22452e4ec94eff566867721`, fingerprint `712d1b3db91b641ed2927f75081dd299a61aec84b3b907d69874c5d7614545c2`. All 259 manifest entries matched on entry and exit. Frozen CONTRACT, confirmed inventory through A11 and correction PLAN govern. A12 remains unanswered; no producer-count, model-label or schema requirement was added. Requested/dispatch-acknowledged setting: Codex `gpt-6.1-sol`, high effort; runtime identity and usage were not independently observed from source metadata.

## Findings and correction evidence

- **DATA-01, severity 5/10, introduced, fixed.** `backfill/discovery.go` now propagates root failures and records descendant traversal failures as incomplete coverage while continuing reachable ledgers. Both migration and backfill retain diagnostics when the target list is empty; the CLI no longer returns before printing those diagnostics. Fresh real permission-denied tests cover selected readable targets, unfiltered siblings, inaccessible selected targets and fatal inaccessible roots. A newly built CLI also passed four isolated probes: both commands cleaned GOOD-01 despite its inaccessible sibling, while selecting only the inaccessible workspace preserved GOOD-01 bytes and printed the inaccessible path plus incomplete-coverage warning. Raw commands/output/hashes: `/tmp/al70-r2-data-sol-l6cbdign/results.json`. This directly exercises the earlier R1 failure/control boundary.
- **CTX-01, severity 4/10, introduced, fixed.** `backfill/publication.go:backfillPublicationPaths` validates required context ceilings for summary publication. Both production callers invoke it before publication ownership, baseline reading, journal checks, backup creation or writes. Fresh migration/backfill tests remove warning and critical settings separately, preserving original ledger/human notes and creating neither backup nor journal; ledger-only controls still succeed without unused context policy. This guards existing A11 metadata cleanup as well as the general publication path.
- **CAP-02, severity 5/10, introduced, fixed.** Worker child arguments carry admitted positive session/byte budgets. Generate applies explicit precedence, passes that effective policy into reconciliation, and `SetSourceBudgets` applies it to registry source discovery without configuration writes. Fresh actual native-worker publication tests succeeded with missing and competing local budgets, checked the recorded source value and publication acknowledgement, and preserved local configuration. Explicit zero/negative overrides failed before output.
- **CFG-01, severity 6/10, introduced, fixed.** Structural node paths escape literal dots/backslashes, preventing root dotted keys from authorizing absent nested policy. Fresh tests cover missing structural booleans/timeout, explicit false/zero controls and local capture enablement. Typed decoding and presence checks now refer to the same structural authority.

## Preservation boundaries and verification

Reinspected affected generation callers, exact-owner reconciliation, source-budget application and publication ordering. Unavailable/failed reads retain snapshots; successful complete reads alone enter refreshed-owner replacement. Budget propagation adds no receipt requirement or source authority. Recovery/member authorization, current-ledger preparation, former-metadata contradiction checks, owner replacement, capture relocation and domain aggregation are unchanged from R1 by manifest comparison (unchanged baseline files remain outside the changed-input manifest). Their R1 inspected A03–A11 preservation, formatting, exact interrupted-write recovery and multiplicity boundaries carry forward on that basis; changed callers were inspected above.

Fresh commands, all with `-count=1`, from `backend/`:

```text
go test -count=1 ./internal/adapters/backfill -run '^(TestWorkspaceDiscoveryIsolatesUnreadableChildren|TestWorkspaceDiscoveryKeepsRootFailuresFatal|TestMetadataSummaryRequiresContextPolicyBeforePublication)$'
go test -count=1 ./internal/adapters/config -run '^TestPolicyPresenceFollowsStructuralYAMLKeys$'
go test -count=1 ./internal/adapters/live ./internal/adapters/http -run '^(TestLiveUnavailableDiffersFromAvailableEmptySource|TestHydrationAvailabilityPreservesLeanFactsAndLedgerBytes|TestLeanFactsSurviveSourceTraceAvailabilityWithoutGETWrites)$'
go test -count=1 ./cmd/agentic-lens -run '^(TestNativeWorkerPublicationKeepsExplicitBudgetOverrides|TestGenerateRejectsInvalidExplicitCaptureBudgets)$'
go build -o /tmp/al70-r2-data-sol-engine ./cmd/agentic-lens
```

All passed: backfill 0.141s, config 0.047s, live 0.392s, HTTP 0.072s, CLI 4.108s; build succeeded. The live tests distinguish unavailable from available-empty source data and preserve persisted facts/bytes across absent, failed and partial hydration. Reused full backfill proof `/tmp/al70-data01-gscr5d4b/final-test.log` (PASS 4.012s): all 246 recorded dependency hashes still match. Parent-provided focused/compile/frontend logs and `/tmp/al70-r2-standalone-bud7xd_7/RESULT.json` are supplemental evidence; the latter reports unchanged bytes for 12 disposable source-free ledgers.

No failed verification or material gap remains in this assignment. Full integrated `make`, external runtime delivery and consolidation remain parent-owned. No developer stores, installation/activation, sibling R2 conclusions, product/test edits, telemetry extraction or publication were used.
