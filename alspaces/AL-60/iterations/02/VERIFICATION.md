# AL-60 verification summary

Tested revision: `057d410253aecd734551995ff1bedb38e94a1f19`. Cleanup preserves application, test and CI source exactly. Final review approval, discovery credit and dispositions remain in [SUMMARY.md](../../SUMMARY.md) and [compact findings](REVIEW_SUMMARY.json).

## Passing checks

- Full `make` passed with four-CPU affinity, `GOMAXPROCS=2` and `GOFLAGS=-p=1`: vet, typecheck/lint, script checks, every Go package with race detection, 786 frontend tests in 36 files, canonical frontend build/sync and five platform binaries. [Archived output](https://github.com/NiltonDuarte/AgenticLens/blob/eb3d515d85493372aab099d7cda82390e3c63fb3/alspaces/AL-60/iterations/02/evidence/MAKE_VERIFIED.txt).
- Standalone host HTTP index/assets matched the generated embed tree byte-for-byte and contained search/no-match copy. No Node runtime served the assets.
- Actual Windows PE tests via WSLInterop passed launcher-exit/cancel ownership and both sources' canceled/deadline recovery cases. Cross-compilation alone was not treated as runtime evidence. [Archived Windows output](https://github.com/NiltonDuarte/AgenticLens/blob/eb3d515d85493372aab099d7cda82390e3c63fb3/alspaces/AL-60/iterations/02/evidence/WINDOWS_INTEGRATED.txt).
- Focused commands used `go test -race ./internal/adapters/capture` and `go test -race ./cmd/agentic-lens` with named installer/worker regressions. Deterministic reds preceded fixes for phase-budget consumption, inherited pipes, SIGTERM, final-write/recovery cancellation and hostile fixture inheritance. Pin/wrapper, pending/ACK, isolation and ownership assertions remained intact. [Archived deterministic outputs](https://github.com/NiltonDuarte/AgenticLens/tree/eb3d515d85493372aab099d7cda82390e3c63fb3/alspaces/AL-60/iterations/02/evidence).

## Historical failed gates

- Initial full make failed unchanged shared capture publication/native-version/manifest checks. [Initial output](https://github.com/NiltonDuarte/AgenticLens/blob/eb3d515d85493372aab099d7cda82390e3c63fb3/alspaces/AL-60/iterations/01/evidence/MAKE_INITIAL.txt).
- Bounded retry failed worker lifetime/cleanup cases; the failing pair passed targeted without affinity, supporting resource sensitivity without proving all failures environmental. [Bounded output](https://github.com/NiltonDuarte/AgenticLens/blob/eb3d515d85493372aab099d7cda82390e3c63fb3/alspaces/AL-60/iterations/01/evidence/MAKE_BOUNDED.txt).
- Integrated vet rejected a closure capturing reassigned phase cancellation. Explicit first/second defers preserved behavior; vet and cancellation guards passed, followed by the successful full make above. [Vet failure](https://github.com/NiltonDuarte/AgenticLens/blob/eb3d515d85493372aab099d7cda82390e3c63fb3/alspaces/AL-60/iterations/02/evidence/MAKE_INTEGRATED_VET_FAILURE.txt).

## Limits and retention

TEST-02 remains accepted nonblocking coverage debt: Unix pipe timing lacks a direct descendant-death/reachability oracle; owner AL-60 maintainer. No current leak was demonstrated. Controlled manifest variation established optional-field changes, not the field behind the historical unretained failure. Windows race instrumentation, external real-harness callback compatibility and browser geometry were not exercised locally.

This summary replaces tracked raw outputs and evidence indexes at the user's request. Detailed output remains available through immutable links to the existing committed versions; no unresolved or accepted-debt record was removed. Contracts, telemetry and compact findings are intact. Artifact-only link/hash/source checks carry prior application verification forward; refreshed exact-head CI and merging remain with MGT-2.
