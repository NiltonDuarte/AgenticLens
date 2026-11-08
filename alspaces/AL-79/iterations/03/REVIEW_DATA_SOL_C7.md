# AL-79 iteration03 C7 data/failure re-review

**APPROVED, scoped. Findings: 0 blocking, 0 additional defects.** The supplied severity-4 disable race is corrected; discovery credit remains with the integration orchestrator and test engineer.

Run: `i03_data_sol_c7`, correction re-review, assigned profile gpt-6.1-sol/high (runtime identity and usage not independently exposed). Base/HEAD: `e12a1d0c38cfa71efd3583daa6e6377f73de84fc`, plus the uncommitted C7 correction. Frozen 71-path candidate SHA-256: `4bd69745495d906d01e8fe48b81e49323d2cda07cdf9397e6988911feda8833c`. Scope is the 15 added lines in `backend/cmd/agentic-lens/capture_worker.go`, assessed against the contract and current iteration03 plan.

The investigated failure hypotheses were:

- **False takeover after disablement:** the loop can read enabled and then receive a false fence after policy disablement. Retirement now rechecks valid policy before the superseded diagnostic and successor selection. A second recheck handles disablement during failed successor validation. Both stop normally without claiming a completed pass.
- **Lost generation or retained ownership:** neither added exit calls `AcknowledgePublication`, mutates registrations, nor publishes artifacts. The first exit retains `held=true`, so the existing defer closes the worker lock. The later exit occurs after successful lock release and `held=false`. The sole acknowledgement remains after successful reconciliation/publication and uses the generation observed before that pass; later requests stay pending. Existing idle-shutdown and lifetime paths keep their ordering.
- **Hidden operational errors:** `ResolveCapture` validates configuration before applying disablement. The first guard returns policy/state errors; the second joins policy errors with the successor error. If capture remains enabled, failed successor validation still records failure and returns the error. Fence errors for source/state changes still return directly. Valid disablement intentionally stops takeover while preserving pending state.

Inspected unchanged dependencies include the fence, automatic-enabled policy reader, successor validation, publication request/acknowledgement and lock release. C6 data scope carries unchanged: installer/migration discovery and recovery authority, leases, manager installation/cache recovery, executable retention, source/binding and portable-ledger persistence have no C7 delta. The guards introduce no receipt requirement or history replacement.

Verification: candidate verifier matched at review start and end; `git diff --check` passed. Reused tester evidence in `TEST_EVIDENCE.md`, matching worker SHA-256 `dfa410b400b90a5374c651c0457b8182383e9e06e6b50bb686083401deae0237`: unchanged strict starvation/exit test failed 3/20 before and passed 20/20 with `-race` after (16.346s); four-mode live-disable plus budget-expiry/pending-recovery controls passed (15.327s). Inspected their real worker, lock, ledger-byte and acknowledgement assertions; independently confirmed the test file hash is unchanged. No duplicate full suite was run.

Limits: concurrent timing is supported by repeated tester runs and source ordering, not exhaustive scheduling. Parent full `make` and machine installation remain pending. Actual home was untouched; Windows runtime is excluded. This verdict does not certify delivery or authorize merge.
