# AL-84 data review: CHANGES_REQUESTED

Run `al84_review_data_sol`, iteration 01 independent discovery. Requested profile: `gpt-6.1-sol` / high; actual runtime model, usage and cost are not independently exposed. Reviewed baseline `163181428ee9a5a599a2da1e91d3af5e618a9682` plus the frozen tracked/untracked candidate: 669-file manifest `13a044baddc800f298f54fba03ee41d4c09092d41c89228afde1a79520a821a6`. Every listed file matched its recorded hash before and after the probes. Product and tests remained unchanged.

## Finding D-SOL-1: replacement can return another ticket's facts

**Severity: 6/10. Introduced correctness defect.** In `backend/internal/adapters/repository.go:177`, `LoadTicketWorkspace` resolves the requested identity under a read lease, releases that lease, then rereads the mutable pathname at line 181 and returns the result without checking its ticket identity. Replace the selected directory with a different valid ledger between these reads and a successful request for ticket A returns ticket B. The detailed reader checks identity again at line 158; this reader does not.

This reaches CLI ticket reads and `backend/internal/adapters/http/handlers.go:300`: the events handler labels replacement-owner coverage, anomalies and any hydrated source events with the requested ticket at lines 313–323. It therefore violates the approved ledger-identity/shared-resolution requirement even without private state or source hydration. Return the requested identity or an explicit conflict/retry; hold the resolved physical selection through the final read or validate that final snapshot before returning it. This finding establishes read misattribution, not a demonstrated overwrite of persisted facts.

**Reproduction:** copied the frozen backend to `/tmp/al84-data-sol-43w0svj8/backend`, added only disposable test files, and used Linux `renameat2(RENAME_EXCHANGE)` to swap two real directories concurrently with the unmodified reader. The configured root was explicit `selected`, depth 0/budget 1; the two supported ledgers recorded `Recorded-A` and `Recorded-B`. No mocks, source hooks or product instrumentation participated. From that copy:

```sh
go test -count=1 -run '^TestReviewLoadWorkspaceReplacement$' -v ./internal/adapters
```

Failed twice: first at iteration 4/successful read 2; retained rerun at iteration 12/successful read 1, both reporting requested `Recorded-A`, received `Recorded-B`. Probe: `internal/adapters/review_data_race_test.go`; raw retained output: `/tmp/al84-data-sol-43w0svj8/replacement-race.log`.

## Other consequential hypotheses

- **Missing private journal hides pending publication: refuted in the exercised path.** A real interrupted two-member transaction, followed by private-journal removal and closed ownership, stayed unavailable to both listing and ticket reads. Recovery explicitly failed and retained the indicator and mixed member bytes.
- **Moved recovery recreates the prior directory or loses portable history: refuted in the exercised path.** Renamed the interrupted home into an arbitrary archive folder, closed/reopened ownership and recovered with the current member paths. Exact prior session/tool facts and incomplete human evidence returned, the old home stayed absent, and repeated recovery was a no-op.
- **Source hydration becomes a persistence gate: not found in the traced paths.** Traced raw historical DTO retention and exact-owner replacement through generation, unchanged backfill ownership/replacement helpers, native-event hydration and fallback responses. Missing hydration does not delete persisted facts. Reviewed the current native receipt for mixed Codex/OpenCode owners, missing private-state reads and external-root late refresh; its synthetic OpenCode usage limitation remains explicit.

Control command `go test -count=1 -run '^TestReviewPendingAndMoveControls$' -v ./internal/adapters` passed both subcases, with the final run (including exact failed-recovery member-byte assertions) completing in 0.124s. Probe and raw output are beside the race probe as `review_data_controls_test.go` and `/tmp/al84-data-sol-43w0svj8/pending-move-controls.log`.

The investigation also traced directory leases, descriptor-relative public I/O, alias/member admission, journal/indicator ordering, maintenance readers and pending-state rejection. No sibling reports were consulted. No broad suite or native macOS/power-loss fault injection was run; integrated `make`, fixes, regression tests and consolidation remain with the parent. This verdict covers the assigned data risks only.
