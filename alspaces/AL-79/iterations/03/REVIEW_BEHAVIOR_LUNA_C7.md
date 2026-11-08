# C7 behavior re-review

**Outcome: APPROVED** for the bounded worker-disable behavior correction. This is a scoped review, not a repository or delivery approval.

## Revision and scope

Reviewed the `capture_worker.go` correction against base `e12a1d0c38cfa71efd3583daa6e6377f73de84fc`; candidate inventory hash `4bd69745495d906d01e8fe48b81e49323d2cda07cdf9397e6988911feda8833c` matched at review start and end. The correction is the two policy rechecks in `backend/cmd/agentic-lens/capture_worker.go` (candidate file SHA-256 `dfa410b400b90a5374c651c0457b8182383e9e06e6b50bb686083401deae0237`). I traced the worker admission/fence/retirement path and unchanged `automaticCaptureEnabled`, `WorkerSuccessor`, and publication request behavior.

## Findings

No blocking behavior finding in the inspected C7 scope.

The initial loop checks enablement before its fence. If disablement lands between those reads and the fence reports the owner is no longer current, retirement now rechecks policy before recording `superseded`, releasing ownership, or selecting a successor. If successor validation returns an error, the second recheck makes a newly disabled policy a normal stop; configuration/state read errors remain errors, and an enabled worker still reports takeover failure. The normal loop admission check still returns immediately on disabled policy, before progress writes, reconciliation, or acknowledgement. Existing pending generations remain available for explicit recovery.

Hypotheses investigated:

1. **Disable between admission and fence produces a false supersession/takeover failure:** corrected branch is reachable and now exits successfully when the recheck sees disabled policy. The unchanged strict race test passed 20/20 runs.
2. **The new guard hides enabled successor incompatibility or configuration failure:** enabled policy preserves the existing failed diagnostic/error path; a policy/state error on recheck is returned joined with the successor error. Source trace supports this disposition.
3. **Graceful disable alters pending acknowledgement or recovery:** the existing live-disable and publication-budget/pending-recovery controls passed; source flow still exits before consuming a generation when disabled.

## Verification and limits

- `go test -race ./cmd/agentic-lens -run '^TestCaptureContinuousNativeCallbacksDoNotStarveAllTickets$' -count=20` — passed, 20/20 (15.991 s in this review; the prior evidence records 16.346 s).
- `go test -race ./cmd/agentic-lens -run '^(TestCaptureLiveDisableStopsWritesAndManualRecoverySurvives|TestCapturePublicationBudgetExpiryRetainsPendingAndRecovers)$' -count=1` — passed (15.269 s).
- `python3 /tmp/al79-native-entrypoint-0ta97kas/verify_c7.py` — candidate hash matched before and after review.
- `git diff --check` — passed.

The supplied test evidence also reports a pre-correction 3/20 failure on the unchanged starvation race test. I did not rerun full `make`; integrated make and machine installation remain parent-owned gates. Actual user-home state was not changed, and Windows runtime behavior was outside this iteration's accepted certification scope. The review verifies the correction's ordinary policy rechecks; it does not claim an atomic synchronization boundary between policy reads and external disablement.
