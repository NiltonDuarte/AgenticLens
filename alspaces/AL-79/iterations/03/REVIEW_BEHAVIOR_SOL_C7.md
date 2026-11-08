# AL-79 iteration 03 C7 behavior review

**APPROVED — 0 findings, 0 blocking defects.** Scoped correction review; not whole-ticket delivery approval.

Run: `i03_behavior_sol_c7`, correction phase; requested profile `gpt-6.1-sol/high` (runtime model/usage metadata unavailable). Base: `e12a1d0c38cfa71efd3583daa6e6377f73de84fc`; uncommitted candidate C7 inventory SHA-256: `4bd69745495d906d01e8fe48b81e49323d2cda07cdf9397e6988911feda8833c`. Only production delta: 15 added lines in `backend/cmd/agentic-lens/capture_worker.go`. Prior C6 scopes carry by unchanged source; sibling C7 reports were not consulted.

The supported operation is disabling an active automatic publisher while callbacks continue. I traced loop admission through `captureWorkerFence.current`, retirement, `automaticCaptureEnabled`, successor validation and worker cleanup. Investigated hypotheses:

- A disable between admission and fencing still returns a successor error: refuted by the retirement recheck at lines 55–58. Valid disable returns nil before supersession diagnostics or handoff; the deferred lock release still runs.
- A later disable during successor validation still fails the owner: refuted by lines 86–92. A validated disabled result returns nil. Earlier retirement diagnostics may already exist, but no failed-worker diagnostic is added for this normal stop.
- The new checks hide malformed policy/state or enabled takeover errors: configuration/operational read errors propagate; successor and policy errors are joined, and still-enabled successor failures retain their detailed diagnostic plus the deferred worker-failure diagnostic. Existing fence state/source errors return before retirement.
- Disable acknowledges unpublished work or redirects ownership: neither added branch updates requests, registrations, ledgers or ACKs. The existing successful-pass ACK remains tied to its captured generation. Stopping with the lease already released does not acknowledge pending work.

Inspected unchanged test oracles: `startCorrectionWorker` requires nil process exit and released ownership; the continuous-callback test requires both exact session/token ledgers before its deadline. The live-disable matrix establishes a nonidle process, checks stable output after YAML/disable/uninstall, retains dry-run publication and manual recovery. Budget expiry requires unchanged prior bytes and unchanged ACK before recovery.

Reused tester evidence in `TEST_EVIDENCE.md`: identical strict starvation test under `-race -count=20` failed 3/20 before and passed 20/20 after (16.346s); live-disable and budget/pending-recovery controls passed together (15.327s). Source/test hashes match that evidence. Independently executed the supplied candidate verifier at review start/end: MATCH; inspected source and exact diff without product edits.

Limits: scheduling repetitions are not exhaustive interleaving proof; the later successor-error guard has static path evidence rather than a forced-interleaving test. No remaining material gap for this bounded review. Renewed integrated `make`, machine installation and non-Windows runtime delivery remain the parent's work; actual home and trust were untouched. Windows runtime is excluded from this review.
