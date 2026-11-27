# GEN-34: Technical design

## Scope and boundaries
GEN-34 delivers [the investigation report](evidence/REPORT.md) and, after the steering below, two test-only speedups in `internal/adapters/capture` and `internal/adapters`. Production code, the Makefile, CI, `cmd/agentic-lens`, other tickets' changes and ledgers stay untouched. No test is deleted or parallelized.

## Decisions and ownership
- **Native fixture.** A tiny `package main` under `backend/internal/adapters/capture/testdata/nativefixture/` is built once per test process, without `-race`, through a `sync.Once` helper. This follows the existing `installationManagerExecutable` helper and `cmd/agentic-lens` `TestMain`. The fixture prints the same probe answers the test executable gives today, including the `bad-manager` and `old-interface` variants selected by `AL79_CAPTURE_NATIVE_FIXTURE`. It receives the capture protocol from `HandlerProtocolVersion` through `-ldflags -X` and exits non-zero on any other command. The fixture is a real native executable for the host, so production validation still runs in full: format and architecture, the executable bit, protocol and version probes, the self-update checksum and journal, protected-manager ownership and checksum, engine installation and executable retention.
- **Converted call sites.** Converted: `compatibleCaptureTestExecutable` (it copies the binary through self-update, journals it and hashes it, and reaches 24 tests), and the native symlinks in `installFixture` and the relocation test, which use the binary only as a probe responder. Unchanged, because they must run test code from the test executable: `install_native_validation_test.go`, the `opencode` probe in `install_deadline_unix_test.go`, and the subprocess re-runs in `diagnostics_test.go` and `store_test.go`. The test executable's protocol-answering `init` and `TestMain` branch are removed, so no test can fall back to the 35 MB binary without anyone noticing.
- **Scale cases.** `DiagnosticHeadLookup` counts {32, 128}; 400 workspace-replacement reads instead of 1500; large-store default divisor 8 (the flag keeps 1 to 16).
- **Owners.** The coder owns the `_test.go` files named in [tasks.md](tasks.md) and the new fixture source. The tester owns new verification evidence under `evidence/`. The orchestrator owns `delivery.md`, telemetry, commits and PR #171. The planner owns this design, the proposal and the tasks.

## Implementation approval
The original ticket request approved the investigation, the report, telemetry, a commit, a push and an unmerged PR. The test changes are approved by the user's steering below. The Makefile, production code and CI remain excluded.

## Implementation approach
Steps are in [tasks.md](tasks.md). Run affected packages with and without `-race` and compare the timings with the baseline in `evidence/`: an isolated `capture` race run took 8:45 wall time, and per-test numbers are in `go-slowest-tests.tsv`. Then run the integrated `make` after merging `main`. Focused test and behavior reviews come before a consolidated review.

## Steering
- **Source and intent:** after reading the report, the user asked to clean up slow tests and then approved: "vamos implementar essa melhoria nesse mesmo ticket. Vamos implementar 1 e 2" ("let's implement this improvement in this same ticket; let's implement 1 and 2"). Items 1 and 2 are the orchestrator's conversation items, the native fixture and the scale cases, not the report's proposals 1 and 2.
- **Baseline:** commit `ab42d501`, with the report measured on `be630e36`.
- **Superseded decision:** "this change delivers an investigation report only, and each accepted proposal becomes a separate ticket". Test changes now land here. The other proposals still need their own tickets.
- **Preserved guarantees:** every converted assertion keeps its oracle, and production validation is unchanged. Race detection still covers all production code running in the test process; only the fixture, which contains no production code, is not race-built.
- **Coverage lost:** the 256-owner case, the race detection probability per run of the replacement test (at 1 % per read: 98 % instead of nearly 100 %), and default-scale large-store coverage below divisor 8. Copy and journal paths now handle about 2.4 MB instead of 35 MB.
- **Verification impact:** the report's measurements stay valid for `be630e36`. New results apply to the candidate revision.
- **Delivery status:** approved. Not yet implemented.
