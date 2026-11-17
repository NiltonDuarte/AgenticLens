# FIX-19 verification

## Before-fix counterexample

Tester executed isolated probes against `b6ab0471224c8bcea7bc00adce867d3ba0de4a37`, the fetched current main. A disposable `git archive` copy under `/tmp/fix19-baseline-od8l79tr` used the baseline source and copied generated embed assets. Synthetic native JSONL fixtures lived in each test's temporary directory; no real source store or ticket ledger was used.

`go test -count=1 -run TestFIX19Before -v ./cmd/agentic-lens ./internal/adapters/live` passed bug-demonstration assertions:

- Actual CLI generation with identical active/archive JSONL copies produced one owner, two invocation facts and bash calls=2 from one native call. Existing exact discovery rejected those same copies.
- Actual `HydrateWorkspaceEvents` with identical active JSONL/archive gzip copies returned availability=complete, complete owners=[owner], one event and no diagnostic. It accepted an arbitrary readable copy.

`go test -count=1 -run TestFIX19Before -v ./internal/adapters/backfill` passed: actual `NewEngine().Execute` with identical active JSONL/archive gzip copies replaced accepted tools with bash calls=2 and reported DBUpdated=true, DBFound=true, RefreshIncomplete=false, Diagnostics=[].

An initial CLI probe with gzip instead reached an earlier existing source-ownership decoding error (`invalid character` for the gzip header), so it did not demonstrate doubled publication. The uncompressed archive variant reached the generation defect. This limit does not weaken the independently executed gzip hydration/backfill counterexamples.

## Candidate verification

Pending the coder's candidate and adversarial regression additions. These before-fix probes establish the defect, not successful correction or final delivery approval.
