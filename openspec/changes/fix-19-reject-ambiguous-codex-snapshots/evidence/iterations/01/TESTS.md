# FIX-19 verification

## Before-fix counterexample

Tester executed isolated probes against `b6ab0471224c8bcea7bc00adce867d3ba0de4a37`, the fetched current main. A disposable `git archive` copy under `/tmp/fix19-baseline-od8l79tr` used the baseline source and copied generated embed assets. Synthetic native JSONL fixtures lived in each test's temporary directory; no real source store or ticket ledger was used.

`go test -count=1 -run TestFIX19Before -v ./cmd/agentic-lens ./internal/adapters/live` passed bug-demonstration assertions:

- Actual CLI generation with identical active/archive JSONL copies produced one owner, two invocation facts and bash calls=2 from one native call. Existing exact discovery rejected those same copies.
- Actual `HydrateWorkspaceEvents` with identical active JSONL/archive gzip copies returned availability=complete, complete owners=[owner], one event and no diagnostic. It accepted an arbitrary readable copy.

`go test -count=1 -run TestFIX19Before -v ./internal/adapters/backfill` passed: actual `NewEngine().Execute` with identical active JSONL/archive gzip copies replaced accepted tools with bash calls=2 and reported DBUpdated=true, DBFound=true, RefreshIncomplete=false, Diagnostics=[].

An initial CLI probe with gzip instead reached an earlier existing source-ownership decoding error (`invalid character` for the gzip header), so it did not demonstrate doubled publication. The uncompressed archive variant reached the generation defect. This limit does not weaken the independently executed gzip hydration/backfill counterexamples.

## Candidate verification

Candidate product/test tree was frozen and committed as `5fa131140a643c76a7a23c6c87e3cda4d24dac99`: baseline plus scoped codex discovery/combination and live hydration changes and their added tests. The focused/race executions use that unchanged product/test tree; only documentation changed during execution. Tests use temporary fixture roots and actual CLI, backfill engine and hydration service calls with no domain mocks.

- `go test -count=1 -run 'TestFIX19|TestCombinedCodexSessions|TestCodex.*(Multiplicity|OwnedOccurrences|LeanOccurrences)|TestManualGenerateLowerAndEmpty' ./internal/adapters/codex ./internal/adapters/backfill ./internal/adapters/live ./cmd/agentic-lens`: codex/live/CLI passed. Initial backfill fixture failed before consumption because it omitted schema; fixture corrected to explicit current schema, then `go test -count=1 -run TestFIX19 ./internal/adapters/backfill` passed.
- Copied all four durable `fix19_*_test.go` files into the disposable baseline and executed `go test -count=1 -run 'TestFIX19(Discovery|Generate|Backfill|Hydration)' ./internal/adapters/codex ./internal/adapters/backfill ./internal/adapters/live ./cmd/agentic-lens`: every package failed at the intended ambiguity/coverage oracle. CLI JSONL variants accepted duplicated publication; gzip variants failed for an earlier generic error instead of ambiguity. Backfill lacked ambiguity diagnostics, discovery returned duplicate paths, and hydration exposed/certified duplicate child events.

The added tests cover duplicate root publication, child relationships (same and conflicting parent), nonce and explicit selection, JSONL/gzip copies, divergent values, warning-bearing copies, failed-refresh byte preservation, unchanged repeats, genuine repeated skills/delegations, exact-discovery rejection, healthy hydration parent with ambiguous child, accepted ambiguous/unavailable same-harness and other-harness owners alongside successful independent refresh.

Existing `TestCombinedCodexSessionsPreserveRepeatedRowsAndNestedOwners` provides distinct repeated-invocation and disjoint nested-combination oracles. Existing manual lower/empty refresh tests exercise portable ownership without receipts; existing mixed-source hydration verifies optional enrichment retains usable other-harness rows. `go test -race -count=1 ./internal/adapters/codex ./internal/adapters/backfill ./internal/adapters/live ./cmd/agentic-lens` passed on the initial candidate (codex 1.257s, backfill 26.837s, live 6.379s, CLI 310.685s). This evidence does not cover the later automatic-source correction.

## Review correction: automatic registered extraction

Architecture review found automatic extraction bypassed discovery when it used a registered source path. The initial test inventory missed this consuming path. Tester added `TestFIX19AutomaticGenerateRejectsRegisteredOwnerCopies`: a real bound owner first publishes accepted input=17, then active input changes to 29 and an identical JSONL, divergent gzip (input=99), or warning-bearing JSONL archive copy appears. Repeated automatic generation must report explicit ambiguity/incomplete extraction and preserve accepted owner session facts. Automatic partial publication for unaffected owners remains allowed; whole-ledger metadata byte equality and a nonzero exit are not required.

`go test -count=1 -run TestFIX19Automatic -v ./cmd/agentic-lens` failed all three variants on initial candidate `5fa13114`: every automatic command succeeded and published despite exact discovery independently rejecting the fixture. This proves the added test detects the confirmed review finding. The initial test asserted a nonzero exit and whole-ledger equality; parent clarified that owner-level preservation and explicit incomplete diagnostics are the accepted oracle. The missing diagnostic still violates that oracle. The first capture guard correction was tested with `go test -race -count=1 -run TestFIX19 ./cmd/agentic-lens ./internal/adapters/codex ./internal/adapters/backfill ./internal/adapters/live`: codex/backfill/live passed, but all three automatic variants still exited zero with only a successful publication message and no ambiguity diagnostic. Coder is correcting downstream failure propagation. Corrected-product verification remains pending; final integrated make/review remain the parent’s delivery gates.
