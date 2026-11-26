# GEN-34: Explain why the integrated `make` is slow and times out

## Roadmap ticket
GEN-34 is an ad hoc maintenance ticket with no roadmap entry. The user opened it because every ticket must run the integrated `make` after merging `main`, and that gate now takes long enough, and fails often enough, to hold up development.

## Why
Other ticket threads reported that `go test -race` timed out for `cmd/agentic-lens` and `internal/adapters/capture` at the default 10 minute limit. They also saw capture worker tests fail under load and then pass on retry, and FIX-21 only passed with `GOFLAGS=-timeout=90m make` on its second attempt. Nobody had measured where the time goes, so fixes were guesses.

## What changes
This change delivers an investigation report only: wall time per `make` stage, the largest consumers down to individual tests, the tests that fail because of timing, and prioritized improvement proposals that state what coverage each one keeps or loses. It does not change the Makefile, tests, production code or CI. Each accepted proposal becomes a separate ticket after the user decides.

## Capabilities
None. No observable product behavior changes, so the change skips delta specs.

## Impact
Developers and agents running `make`. Nothing in the product, tests or CI changes.
