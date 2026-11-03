# Shared capture gate diagnosis (read-only)

Status: INCOMPLETE delivery gate; repair expansion explicitly not approved. Search source remains at reviewed application commit 7a50a5e99927928995a72e13128d597c1317e9d3. Capture source/tests are unchanged from baseline 122e983f8c365a2aa248b1b6d492252555e922ae. Diagnosis by read-only explorer and parent, both gpt-6.1-sol/high; these are verification actors, not search defect discoverers.

## Observed evidence

- Initial `make` failed cmd capture automatic publication, installer shared/local native version checks and multi-root refresh; capture adapter manifest idempotency also failed. Script/static checks passed. Raw local log: `/tmp/al60-make.log`.
- Bounded `make`, two-CPU affinity with GOMAXPROCS=2 and GOFLAGS=-p=1, exit 2: LaterAgentChannelReportsBoundPublicationFailure failed with `automatic publication pending: signal: killed` and `native worker lifetime exhausted; pending generation retained`; AutomaticSharedStoreAndTicketStayInExactProjects failed worker cleanup. Capture adapter suite passed (105.100s); cmd suite failed (168.056s); other Go packages passed/cached. Raw local log: `/tmp/al60-make-bounded.log`.
- Those two failing bounded cmd tests passed without affinity in a targeted race run, GOMAXPROCS=2 and -p=1, 8.291s. This is evidence of resource sensitivity, not full gate success.
- Frontend serial suite 786/786 in 36 files passed. Typecheck/lint, canonical build and standalone embedded HTTP verification passed. No search/capture call path connection was found.
- User reports AL-61 and MGT-2 capture failures on unchanged main 122e983. These corroborate baseline scope, but are not locally reproduced baseline measurements.
- Real PATH probes during diagnosis: opencode 1.19.5 exited 0 in 0.934s; codex-cli 0.166.0 exited 0 in 0.011s. These unloaded samples do not measure loaded failure latency.

## Likely mechanisms and confidence

1. Installer deadline coupling: `backend/cmd/agentic-lens/capture.go:92,108` bounds the whole command at two seconds. `backend/internal/adapters/capture/install.go:300` derives a three-second optional real-harness probe from that same context. Required transaction native version/protocol commands subsequently use the remaining parent context (`install_transaction.go:89-95`). A slow optional probe can exhaust required validation. Execution errors, deadline cancellation and actual mismatch become the same unavailable-or-changed error. This is a source-demonstrated budget/diagnostic weakness, triggered by environmental latency; exact elapsed failures were not instrumented.
2. Manifest idempotency: live harness probe version/capability are stored on each install (`install.go:103-123,289-302`). Test native fixtures retain real harnesses on PATH; a successful then timed-out probe can change manifest bytes (`installation_unix_test.go:420-444`). No failing before/after manifest was retained, so the changed-field explanation remains a hypothesis.
3. Worker deadlines: failure-feedback fixtures set lifetime 1000ms and pass timeout 500ms (`capture_failure_feedback_test.go:37`); related isolation fixtures use lifetime 1000ms. `capture_worker.go:211-225` kills generation at min(remaining lifetime, pass timeout); lines 277-280 preserve pending work on lifetime exhaustion. The observed killed process and missing acknowledgement are consistent with deliberate budget enforcement. A targeted pass without affinity supports timing sensitivity but does not prove all worker failures environmental or correctness under sustained contention.

## Minimal repair boundary if authorized

- Capture installer/CLI deadline ownership and diagnostics: optional harness inspection must not consume required transaction validation's budget; retain native pin/protocol checks and distinguish timeout/process error from identity mismatch.
- Installer test fixtures: use controlled disposable harness executables for discovery/probe results, retaining real native/wrapper verification. Confirm the manifest field that changes before claiming that hypothesis fixed.
- Worker integration fixtures/publication adapter: distinguish deterministic timeout/recovery cases from successful real-generation cases; preserve bounded work, pending generations, ownership, atomic publication and recovery. Investigate evidence before touching production worker behavior; do not merely increase global deadlines or weaken assertions.

Next evidence: independent required-vs-optional probe timing/cancellation and failing manifest field capture in isolated fixtures; same source successful-publication case under controlled load and deliberate budget-expiry case with preserved pending state. No capture code/tests changed, no additional full suite rerun after the explicit stop. MGT-2 owns the scope decision and merge gate.
