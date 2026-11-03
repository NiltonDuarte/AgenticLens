# GEN-19 test effectiveness review

**Outcome: APPROVED (test-effectiveness scope only).** No blocking test-oracle gap found for the assigned regression hypotheses. This is not a whole-change approval.

- **Reviewed candidate:** baseline `036d808b0c77bec400f23ca77a5cf056a0b2078b`; frozen fingerprint `3a5ea9f89c48818862332b7da06019145544ec9ab2edcb830cae9a28b752765b`.
- **Scope inspected:** `scripts/release/rc-version.test.cjs`, `scripts/release/rc-version.cjs`, `scripts/tests/test_release_revision.py`, Makefile test target, and RC workflow conditions. Requirements from GEN-19 contract/plan and review brief.
- **Identity:** requested profile was gpt-6-luna/medium. Runtime-observed model/effort metadata was unavailable; I do not infer actual identity from the task label.

## Findings

None (severity n/a). Confidence: high for the listed test oracles and moderate for workflow execution, which is not locally exercised by these tests.

## Hypotheses investigated

- **Regression to the already released numeric version:** the Node tests run the installed, pinned Release Please `Manifest` implementation and assert a feature change at manifest `2.0.0` resolves to `2.1.0`. They also assert patch, breaking, commit `Release-As`, and configured `release-as` outcomes. The fixture includes a prior `v2.0.0-rc` release and stable `v2.0.0`; the merge iterator supplies conventional-commit history. Thus the main version-selection assertion would fail if the RC were treated as the numeric baseline or if this feature case returned `2.0.0`. The separate raw verification supplied with the brief reports the pinned engine's read-only calculation on actual `main`/baseline as `2.1.0`; this complements rather than replaces the fixture oracle.
- **Stale/superseded run:** controlled head sequences exercise rejection before config reads, after calculation, and at the final verification call. The tests assert explicit `skip` status. They use a fake head reader, appropriately isolating the branch-race boundary; actual GitHub head behavior is covered only by adapter response tests.
- **No candidate versus error:** a non-release-worthy commit returns an explicit skip reason, while a thrown GitHub/config API error rejects. The CLI integration check additionally verifies checkout mismatch and missing token fail with no output/environment file creation. Adapter tests establish HTTP and malformed-response errors fail closed.
- **Already-stable target:** resolver tests reject a target at or below the published stable version both during initial resolution and final verification, including a higher stable version arriving by recheck. Stable-release pagination tests exclude drafts and prereleases and retain only the configured unprefixed stable tag namespace.
- **Workflow gating:** Python tests inspect the RC workflow's named steps and assert resolver invocation/order, publish conditions, recheck-before-publish/prune, and build/tag metadata wiring. These are source-text assertions using a purpose-built step parser, not YAML/action execution. Combined with the executable resolver and adapter tests, they cover the intended wiring contract, but do not establish GitHub Actions expression semantics or hosted upload behavior. The contract already records hosted publication as unverified.

## Verification

Ran `make test-release-workflow` locally: 8 Python workflow tests passed; Node reported 13 tests passed, 0 failed, 0 skipped. Local runtime version was not recorded in this run. The brief also supplies raw evidence of the same 13 Node tests passing with zero skips on Node 20.20.2 (the CI-declared major), and 8 Python tests plus 13 Node tests passing on Node 24.22.1. No tests were added or changed in this review.

## Limits

No full integrated `make` run or hosted workflow/publication was performed as part of this test-focused review. Workflow source assertions do not simulate Actions, and fixture-backed GitHub tests do not prove remote service behavior. These limits do not leave an assigned test-oracle hypothesis unresolved; retain the hosted-boundary limitation in the overall delivery record.
