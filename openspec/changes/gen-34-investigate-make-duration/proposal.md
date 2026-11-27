# GEN-34: Explain why the integrated `make` is slow and speed up its slowest test fixtures

## Roadmap ticket
GEN-34 is an ad hoc maintenance ticket with no roadmap entry. The user opened it because every ticket must run the integrated `make` after merging `main`, and that gate now takes long enough, and fails often enough, to hold up development.

## Why
Other ticket threads reported that `go test -race` timed out for `cmd/agentic-lens` and `internal/adapters/capture` at the default 10 minute limit. They also saw capture worker tests fail under load and then pass on retry, and FIX-21 only passed with `GOFLAGS=-timeout=90m make` on its second attempt. Nobody had measured where the time goes, so fixes were guesses.

The [investigation report](evidence/REPORT.md) found that `capture` is CPU-bound under the race detector. A large share of that time is test setup, not the behavior under test: many tests install the 35 MB race-built test executable as a fake manager or engine, and a few tests repeat a scale case more often than their assertion needs.

## What changes
1. **Investigation report (delivered).** Wall time per `make` stage, the largest consumers down to individual tests, the tests that fail because of timing, and ranked improvement proposals with the coverage each keeps or loses.
2. **Smaller native test fixture (approved after the report).** `capture` tests that only need a native executable that answers protocol and version probes use a small, separately built fixture instead of the race-built test executable. Every assertion stays.
3. **Smaller scale cases (approved after the report).** Three scale tests run at a smaller size or iteration count, each with a stated coverage loss. The large-store case keeps its flag for full scale.

The Makefile, production code, CI and `cmd/agentic-lens` tests do not change. No test is deleted or parallelized. The report's other proposals still need their own tickets.

## Capabilities
None. Only test code and test fixtures change, so product behavior is unchanged and the change skips delta specs.

## Impact
Developers and agents running `make` or `go test` on `internal/adapters/capture` and `internal/adapters`. Product behavior, CI jobs and the Makefile stay the same.
