# Behavior review — AL-64 iteration 04

- Outcome: **APPROVED** (behavior focus only)
- Reviewer: behavior reviewer, requested profile `gpt-6-luna / medium`; runtime model metadata was not independently available in this review context.
- Baseline: `d8a7c14cfd9f0ee626fedc33c70f92bd928d9192`.
- Candidate: frozen working-tree candidate identified by `/tmp/al64-04-review/fingerprint.json` SHA-256 `c766dd83ec3e9f601e6e5b302422defa2bbe6c86f5906648cae1e272a338a32d`; all 411 listed file hashes matched. Candidate diff SHA-256: `9afb94b071ce4f74e32fad4ea409f1887899d2c3b9a7f3057fa3f4695d41da82`.
- Scope: threshold selection, owner/interval attribution, exact user-facing copy, CLI emission timing and output boundaries. Product and test files remained read-only.

## Findings

None. No substantiated behavior defect in the assigned scope.

## Hypotheses investigated

1. **Rounded provisional subtotal could incorrectly trigger at 179999 ms, or known short intervals could fall through to F04.** Refuted by `SelectCheckpointFeedback`: exact interval age is derived from the explicit cutoff and compared at second precision before subtotal fallback. Complete evidence containing individually short intervals does not set `unresolved`; therefore a rounded subtotal at threshold does not emit feedback. The tester report records discriminating exact-boundary and known-short tests, including perturbations that fail if the comparison is exclusive or all rows fall back to F04.
2. **Feedback could pool owners, models, or invocation rows, or erase legitimate duplicate rows.** Refuted by row keys `(source, session, invocation row)`, iteration over each ledger owner/invocation, and stable ordering by source/session/role/model/row/interval. The selector retains row identity and does not sum subtotals. Duplicate evidence for an ambiguous row is qualified instead of being trusted as an exact interval.
3. **Unverified generic caller or retained source facts could be presented as current-caller proof.** Refuted by `main.go` passing an empty `CheckpointCaller` for generic generate, and by transient evidence being merged only from freshly extracted source events. With this supported production path, exact source-open intervals select F03; missing or incomplete interval evidence may select F04. Cached/retained session facts alone cannot certify F01/F02. The runtime report explicitly marks native caller and continuation transports unverified and does not claim native F01/F02/F05 execution.
4. **Diagnostic output could be emitted on failed/automatic generation or alter published data.** Refuted by the call site: selection and stderr writing occur after ledger save, requested summary update, publication commit and diagnostic acknowledgement, and only under `!autoFlag`. The writer error is reported as optional and does not change successful generation status. The serializer/test report and changed-path inspection show no persisted feedback field, summary/API/frontend projection, or durable dedup state. Tester reports isolated CLI checks for output separation, publication failure, disabled/custom thresholds and a failing stderr writer.
5. **Copy or actionability could claim a same-turn refresh closes the interval, instruct an unrelated owner to stop, or create a new turn automatically.** Refuted by the five renderer branches: F01 asks for natural completion and a later refresh while explicitly saying same-turn generation does not close the interval; F02 addresses the other owner without instructing forced completion; F03/F04 qualify uncertainty; F05 says the command does not end, schedule, or resume a turn. F05 is appended once only when selector evidence includes eligible F01 with continuation unverified. The plan's supplied exact-copy tests are reported as passing.

## Verification and limits

- Verified all 411 manifest content hashes: zero mismatches. Confirmed HEAD equals the frozen baseline and `git diff --check` passes.
- Read the approved iteration 04 plan, implementation, tester final report, runtime capability report, assigned behavior persona/shared review context, canonical review procedure, technical guidelines and portable-ledger regression cases.
- Reused focused CLI, domain/config/source, affected-source race and public unit evidence in `TEST_REPORT.md`; did not rerun broad suites. The parent owns final integrated `make` and standalone publication gates.
- Static call-path inspection confirmed post-publication explicit-command emission and stderr-only rendering. Fixture and CLI test evidence is credited to the tester, not claimed as independently rerun here.
- Runtime caller/current-interval transport and real completion-to-next-turn refresh remain unverified; the implementation correctly qualifies those limits rather than claiming the unavailable capability. This approval covers only the assigned behavior scope and frozen candidate, not whole-repository correctness or delivery.
