# AL-60 iteration 03 verification

Reviewed/tested source: `7c97677dacbdb71466bb0d041704c2cc2b2619e2`, based on current main `6244585dc5cc41ef012cd3e1fb3b964ccd910a95`. Only four Go test files change. Production, CI and the previously reviewed search behavior remain unchanged.

## Failure and deterministic reproduction

At AL-61 head `64d1e3d5ad36b1db1e764c48ddc57298f60a4d3f`, Backend job 126168359349 failed only the shared-store isolation case: TempDir RemoveAll reported directory-not-empty in 0.53s; other Go packages passed. Its backend/CI matched main byte-for-byte.

The coder held real configuration admission and launched a native worker. Generation equaled ACK and the lease was free, so the old cleanup oracle passed while that process remained alive. Removing only worker.lock, releasing admission and waiting for the native recreated worker.lock through AcquireWorker, even with capture disabled. The isolated pre-fix probe intentionally failed on this late write. Whole-registry deletion was rejected by existing validation and was not evidence of recreation. This establishes a reachable late writer, not the exact original CI file or scheduling.

## Correction and checks

Unix FIFO EOF and Windows exclusive worker.log access observe inherited native/generator/successor handles through registry close. Successors acquire their handles before predecessor exit. Cleanup disables future capture passes and waits for real completion. Automatic publication, exact project/session/token/multiplicity, pending/ACK and exclusive ownership checks remain.

Data/failure reviewer gpt-6-luna/medium discovered DATA-03 (severity 5): reporting the five-second timeout still permitted unsafe removal. The correction retains that diagnostic and error, then drains completion without a local deadline. Persistent observer failure reports and retains fixtures until the existing suite timeout. No global timeout or product policy changed.

From backend/, against the corrected source:

- `go test -race ./cmd/agentic-lens -run '^(TestCaptureWorkerCompletionIncludesDelayedNativeStartup|TestCaptureWorkerCompletionSurvivesFixtureEarlyReturn|TestCaptureAutomaticSharedStoreAndTicketStayInExactProjects)$' -count=10`: PASS, 15.978s.
- Windows whole CLI test package cross-compilation: PASS.
- `git diff --check`: PASS.

The early-return case deterministically cancels the first wait, observes its background drain still pending with fixture present while native admission is held, releases admission, and requires the preserved cancellation plus native-created lease before fixture removal. Earlier pre-review-correction race runs passed: three cases count5 (8.481s), original isolation count20 (23.914s). These do not replace the corrected-source result.

Full integrated `make`: PASS, 217.7s, after synchronizing current main. Environment bounded to four allowed CPUs, GOMAXPROCS=2 and GOFLAGS=-p=1; no product/test deadline increased. Go race suite passed (CLI 107.194s; unchanged packages reused valid cache), all 786 frontend tests passed, workflow/install/distribution checks passed, frontend embedded assets rebuilt and five platform binaries built. Final scoped reviews approved; the refuted fallback claim was independently tested with the early-return race case (1.946s). Final publication commits contain only ticket summaries/telemetry. Source equality against tested 7c97677 passed for every path outside the AL-60 workspace; production/CI match main and earlier contracts/compact findings/verification plus AL-55 telemetry remain byte-identical. Final hosted checks remain the merge owner's gate. Coder, tester and integrator use gpt-6.1-sol/high; discovery credit is separate from verification.

## Limits

No skips. Race detection instruments tests, not the normally built native CLI. Actual Fatal/Goexit and permanently broken-observer paths were source-inspected, not executed. The unchanged exact Windows completion helper passed actual-platform standalone direct-native and successor-handle probes (0.11s); full Windows capture integration was unavailable because Windows Go is not installed. A permanently stuck writer intentionally reaches the existing suite timeout rather than unsafe teardown. Earlier installer descendant-oracle TEST-02 debt remains accepted and owned by the AL-60 maintainer. No raw output or archive links are tracked.
