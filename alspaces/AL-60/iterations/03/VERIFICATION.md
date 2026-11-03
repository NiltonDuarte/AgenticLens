# AL-60 iteration 03 verification

Independent tester: baseline `e19a1983c600bd66d9caf6d52bbd87b756f8da67` plus the frozen changes to `capture_isolation_test.go`, the Unix/Windows completion helpers, and tester-owned `capture_worker_completion_failure_test.go`. No runtime source was changed. Commands ran from `backend/` on Linux against this uncommitted candidate.

The early-return guard launches the real native behind production configuration admission. Generation equals ACK and the worker lease is free, yet the canceled completion wait rejects success while the native remains blocked. The nested fixture returns without releasing admission in its body. Its cleanup deliberately cancels a bounded wait and observes the subsequent background drain against the real completion observer. While admission remains held, the drain has not returned and the fixture directory exists. Cleanup then releases admission, requires the initial cancellation to remain visible in the returned error, and verifies the native recreated the removed lease file before Go removes the fixture directories. The outer test verifies directory removal. Deferred canonical cleanup protects the probe's own failure paths. No intentional test failure or skip is needed.

This covers fixture early return and failed-wait recovery. Actual Fatal/Goexit and a permanently failing observer are source-inspected paths; neither is deliberately induced in a subprocess. The latter retains fixture ownership until the existing suite timeout rather than authorizing removal without completion proof.

The original automatic trigger and exact per-project session identity, input-token count and session/invocation multiplicity assertions remain intact. Existing generation/ACK and exclusive worker-lease checks are preserved. The deterministic delayed-start case demonstrates why those checks alone cannot establish native completion; repeated success is additional evidence.

Executed checks before the diagnostic-timeout review correction:

- `go test -race ./cmd/agentic-lens -run '^(TestCaptureWorkerCompletionIncludesDelayedNativeStartup|TestCaptureWorkerCompletionSurvivesFixtureEarlyReturn|TestCaptureAutomaticSharedStoreAndTicketStayInExactProjects)$' -count=5`: passed, 8.481 seconds.
- `go test -race ./cmd/agentic-lens -run '^TestCaptureAutomaticSharedStoreAndTicketStayInExactProjects$' -count=20`: passed, 23.914 seconds.
- `git diff --check`: passed.

After correction, against the frozen source including the extended canceled-wait guard:

- `go test -race ./cmd/agentic-lens -run '^(TestCaptureWorkerCompletionIncludesDelayedNativeStartup|TestCaptureWorkerCompletionSurvivesFixtureEarlyReturn|TestCaptureAutomaticSharedStoreAndTicketStayInExactProjects)$' -count=10`: passed, 15.978 seconds.
- `git diff --check`: passed.

No skips occurred. The CLI is built normally by TestMain; race detection instruments the test/helper process, not the native CLI subprocess. The parent owns integrated make and revision-aware review after this correction.

## Reproduced writer and correction

Before correction, the coder held real config admission and launched a native worker. Generation equaled ACK and worker.lock was momentarily acquired/released, yet the native remained live. Removing only worker.lock and releasing admission caused the native to recreate it through AcquireWorker's file creation, even with capture disabled. The isolated probe intentionally failed on that late write. Deleting the whole registry instead was rejected by existing validation and was not treated as proof of registry recreation. The CI directory error alone does not identify its exact original file; this reproduces a reachable writer at the same teardown boundary.

The correction is test-only. Unix continuously drains a FIFO worker.log and waits for EOF after final callbacks. Windows waits for exclusive access after all inherited native/generator/successor log handles close. Every launch owns its log handle before Start returns; successors open their handles before predecessor exit. This includes registry close after lease release, unlike the original lock snapshot. Deferred cleanup disables new passes/handoffs and waits for descriptor completion, including early cleanup returns. All isolation, pending/ACK and lease assertions remain.

Data/failure reviewer gpt-6-luna/medium identified that reporting the original five-second wait error still allowed fixture removal beneath a live writer (severity 5). The correction preserves that diagnostic deadline and initial error, then waits without a local deadline for actual completion. A final observer error is reported and parks cleanup until the existing suite timeout. The tester's controlled cancellation reaches the fallback deterministically without waiting five seconds or deliberately failing the suite. No framework or global timeout changed.

Coder focused Linux cases passed, including race count5 (9.207s). The Windows CLI test package cross-compiled. The exact Windows helper was copied into a private standalone file-list probe (without CLI TestMain, which requires an unavailable Windows Go installation) and executed on actual Windows via WSLInterop: direct native and successor-held handles rejected premature canceled completion, then completed after controlled TCP release; combined PASS0.11s. This is helper-level Windows runtime evidence, not a full Windows capture integration run. Coder/tester actors are gpt-6.1-sol/high; supplied reproduction/verification is distinct from independent review discovery. No raw logs, temporary executables or archive links are tracked.
