# GEN-19 test effectiveness review

Verdict: APPROVED for the assigned test-effectiveness scope, with one nonblocking improvement. This is not whole-change approval or publication authorization.

- Run: review-tests-sol, independent first pass.
- Baseline: 036d808b0c77bec400f23ca77a5cf056a0b2078b.
- Reviewed candidate fingerprint: 3a5ea9f89c48818862332b7da06019145544ec9ab2edcb830cae9a28b752765b. Independently recomputed every listed file SHA-256 and confirmed all match CANDIDATE.json.
- Requested identity: gpt-6.1-sol / high. Harness: Codex. Exact runtime model and reasoning metadata are not independently exposed; requested identity is not source-observed identity.
- Scope: regression oracle effectiveness, real engine fixture reachability, stale/error/no-candidate/stable-target paths, workflow execution claims and proportionality. No sibling review conclusions were read. No product or test files were changed.

## Findings

No blocking correctness or contract finding was established.

### T1: Successful CLI metadata handoff lacks an executable regression oracle

Severity: 3/10. Classification: nonblocking test improvement. Confidence: high.

Locations: scripts/tests/test_release_revision.py:143; scripts/release/rc-version.test.cjs:128; scripts/release/rc-version.cjs:146.

The Python CLI cases exercise only checkout mismatch and missing-token rejection. The Node cases separately exercise candidateMetadata and the resolver, while the workflow test checks literal wiring. No existing case invokes a ready CLI result and asserts the actual GITHUB_ENV metadata and GITHUB_OUTPUT permission. A regression that computes the correct target but omits the environment-file write would allow the CLI to succeed without VERSION/RC_TARGET being propagated to build and final eligibility steps.

Evidence: in an isolated temporary copy, replacing the production GITHUB_ENV append at rc-version.cjs:150 with `void metadata;` left both the full Node release suite and the Python release revision suite green. The unmodified implementation currently performs the write correctly, so this is a demonstrated oracle limit, not a discovered shipping defect. A coherent successful CLI fixture asserting file contents, and optionally the existing build command consuming them, would close this gap. This need not become a hosted publication test.

## Hypotheses and evidence

- Wrong next version: the production resolveCandidate calls the pinned Release Please 17.3.0 Manifest implementation. Only GitHub reads are replaced in the fixture. Real conventional-commit parsing and the Go release strategy calculate independent expected values: 2.0.0 plus feat -> 2.1.0, fix -> 2.0.1, breaking -> 3.0.0, commit-body Release-As -> 2.3.0, configured release-as -> 4.0.0. Configuration and manifest reads are asserted to use main. These are behavioral oracles, not copied increment logic.
- Prior RC contamination: each engine fixture yields v2.0.0-rc.113+1234567 before v2.0.0. The real engine matches the manifest's stable version to the stable release, then considers commits after that release. The expected versions therefore exercise the relevant prior-RC path. This is a synthetic API fixture, not evidence of real GitHub pagination inside Release Please.
- Stale runs: fixtures reach a stale initial head, advancement after engine calculation, and advancement at the resolver's eligibility check. Final verifyCandidate also has a stale-head assertion. An isolated mutation disabling all three head-check locations was rejected by the Node suite.
- Already stable targets: both equal-to-target and greater-than-target published versions are rejected by verifyCandidate, and initial resolution is also checked against a published target. An isolated mutation disabling the published-version comparison was rejected.
- No candidate versus errors: a real engine calculation for a chore-only commit skips with an explicit reason; a configuration-read failure rejects instead. Empty/missing root configuration and invalid/missing manifest versions reject. A configured release-as lower than the manifest version rejects. Mutating the no-candidate skip into an exception was rejected.
- Original regression: mutating the selected target to the last released manifest version was rejected by the Node suite. In the current implementation the independent monotonic-version guard rejects this before ready metadata can be produced; feature/patch/breaking assertions also explicitly require the intended next values.
- Transport and stable-list filtering: the real githubContext adapter executes against intercepted fetch calls. A full first page of prereleases forces a second page; the expected stable list excludes drafts, prereleases and other tag namespaces. HTTP failures and malformed top-level API responses reject. These are adapter-level tests; engine history traversal remains behind its GitHub fixture boundary.
- Metadata shape: the exact RC version, run number, SHA, date and RC_TARGET are asserted through the production candidateMetadata helper, with invalid versions and run numbers rejected. Successful CLI file serialization is verified by the live probe below; its lack of a durable regression oracle remains the T1 limit.
- Workflow wiring: the Python test checks literal command order, candidate/eligibility guards, build variable arguments and published tag source. It does not parse GitHub expressions or execute an Actions runner. Source inspection confirms the current candidate and eligibility step IDs and output names agree. Existing stable-workflow shell tests still execute the real shell blocks against disposable Git repositories.

## Executed and reused verification

Executed in temporary copies outside the shared source tree, with a symlink to the already installed pinned dependencies:

- Control: `node --test <copy>/scripts/release/rc-version.test.cjs` and `python3 <copy>/scripts/tests/test_release_revision.py` both exited 0 on Node v24.22.1.
- Repeated those two commands for five independently copied mutations: reuse last released target; remove stale guards; disable published-stable comparison; error instead of no-candidate skip; omit ready environment write. The first four exited 1 for Node and 0 for Python; the last exited 0 for both. All temporary copies were removed automatically. No network publication or production mutation was performed.
- Reused supplied tester evidence: make test-release-workflow passed with 8 Python tests and 13 Node tests, zero skips on Node v24.22.1. Parent additionally supplied all 13 Node tests passing with zero skips on v20.20.2, matching CI's major version.
- Reused supplied read-only live calculation: the pinned engine returned 2.1.0 against main/036d808. This complements, but does not replace, the synthetic engine tests.
- Independently executed the actual production `node scripts/release/rc-version.cjs resolve` against the checked-out 036d808 and live main, with GITHUB_RUN_NUMBER=114 and temporary GITHUB_ENV/GITHUB_OUTPUT files. Exit 0; environment file contained VERSION=v2.1.0-rc.114+036d808, GIT_COMMIT=036d808, BUILD_DATE=2026-11-03T12:00:20Z and RC_TARGET=2.1.0; output file contained should_publish=true. Authentication was passed only through the child-process environment and was not printed. The temporary files were removed. This is current CLI execution evidence, not a durable regression oracle or hosted publication.
- Inspected existing Makefile release-build, release-package, release-verify and linker flag consumers. The added tests are proportionate: one real-engine table with shared fixtures, focused guard cases and a small extension of the existing workflow suite. No removal or count-based reduction is recommended.

## Limits

No hosted Actions execution, upload, download or publication was exercised. The full integrated make gate was pending in the supplied brief and is owned by the parent; this scoped verdict does not certify it. The successful resolver CLI metadata write was independently executed against live read-only GitHub data, and build consumers were inspected; there is no end-to-end executable test connecting those stages. Ambiguous/missing library candidate output and all unsupported dependency shapes are guarded in source but not explicitly reached by the current real-engine fixtures. No claim of comprehensive mutation coverage is made. GitHub reads and publication remain non-atomic as documented in the frozen contract.
